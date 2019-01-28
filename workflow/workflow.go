package workflow

import (
	"cane-project/api"
	"cane-project/database"
	"cane-project/model"
	"cane-project/util"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/mongodb/mongo-go-driver/mongo/options"

	"github.com/tidwall/sjson"

	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/gjson"

	"github.com/go-chi/chi"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

// CreateWorkflow Function
func CreateWorkflow(w http.ResponseWriter, r *http.Request) {
	var target model.Workflow

	jsonErr := json.NewDecoder(r.Body).Decode(&target)

	if jsonErr != nil {
		fmt.Println(jsonErr)
		util.RespondWithError(w, http.StatusBadRequest, "error decoding json")
		return
	}

	filter := primitive.M{
		"name": target.Name,
	}

	_, findErr := database.FindOne("workflows", "workflow", filter)

	if findErr == nil {
		fmt.Println(findErr)
		util.RespondWithError(w, http.StatusBadRequest, "existing workflow")
		return
	}

	_, saveErr := database.Save("workflows", "workflow", target)

	if saveErr != nil {
		fmt.Println(saveErr)
		util.RespondWithError(w, http.StatusBadRequest, "error saving workflow")
		return
	}

	// target.ID = deviceID.(primitive.ObjectID)

	// foundVal, _ := database.FindOne("workflows", "workflow", filter)

	util.RespondwithString(w, http.StatusCreated, "")
}

// DeleteWorkflow Function
func DeleteWorkflow(w http.ResponseWriter, r *http.Request) {
	filter := primitive.M{
		"name": chi.URLParam(r, "workflowname"),
	}

	deleteErr := database.Delete("workflows", "workflow", filter)

	if deleteErr != nil {
		fmt.Println(deleteErr)
		util.RespondWithError(w, http.StatusBadRequest, "workflow not found")
		return
	}

	util.RespondwithString(w, http.StatusOK, "")
}

// GetWorkflow Function
func GetWorkflow(w http.ResponseWriter, r *http.Request) {
	filter := primitive.M{
		"name": chi.URLParam(r, "workflowname"),
	}

	foundVal, foundErr := database.FindOne("workflows", "workflow", filter)

	if foundErr != nil {
		fmt.Println(foundErr)
		util.RespondWithError(w, http.StatusBadRequest, "workflow not found")
		return
	}

	util.RespondwithJSON(w, http.StatusOK, foundVal)
}

// GetWorkflows Function
func GetWorkflows(w http.ResponseWriter, r *http.Request) {
	var opts options.FindOptions
	var workflows []string

	foundVal, foundErr := database.FindAll("workflows", "workflow", primitive.M{}, opts)

	if foundErr != nil {
		fmt.Println(foundErr)
		util.RespondWithError(w, http.StatusBadRequest, "no workflows found")
		return
	}

	if len(foundVal) == 0 {
		fmt.Println(foundVal)
		util.RespondWithError(w, http.StatusBadRequest, "empty workflows list")
		return
	}

	for key := range foundVal {
		workflows = append(workflows, foundVal[key]["name"].(string))
	}

	util.RespondwithJSON(w, http.StatusOK, map[string][]string{"workflows": workflows})
}

// CallWorkflow Function
func CallWorkflow(w http.ResponseWriter, r *http.Request) {
	var targetWorkflow model.Workflow

	bodyBytes, bodyErr := ioutil.ReadAll(r.Body)
	stepZero := string(bodyBytes)

	if bodyErr != nil {
		fmt.Println(bodyErr)
		util.RespondWithError(w, http.StatusBadRequest, "error reading body")
		return
	}

	filter := primitive.M{
		"name": chi.URLParam(r, "workflowname"),
	}

	foundVal, foundErr := database.FindOne("workflows", "workflow", filter)

	if foundErr != nil {
		fmt.Println(foundErr)
		util.RespondWithError(w, http.StatusBadRequest, "workflow not found")
		return
	}

	mapErr := mapstructure.Decode(foundVal, &targetWorkflow)

	if mapErr != nil {
		fmt.Println(mapErr)
		util.RespondWithError(w, http.StatusBadRequest, "error parsing workflow")
		return
	}

	fmt.Println("Generating Claim Code...")
	workflowClaim := GenerateClaim()
	workflowClaim.Save()
	util.RespondwithJSON(w, http.StatusCreated, map[string]interface{}{"claimCode": workflowClaim.ClaimCode})

	go ExecuteWorkflow(stepZero, targetWorkflow, workflowClaim)
}

// ExecuteWorkflow Function
func ExecuteWorkflow(stepZero string, targetWorkflow model.Workflow, workflowClaim Claim) {
	var setData gjson.Result
	var stepAPI model.API
	var stepAPIErr error

	// apiResults := make(map[string]interface{})
	apiResults := make(map[string]model.StepResult)

	fmt.Println("Beginning Step Loop...")

	fmt.Println("Step Zero Body:")
	fmt.Println(stepZero)

	// For each step in "STEPS"
	for i := 0; i < len(targetWorkflow.Steps); i++ {
		var step model.StepResult

		fmt.Println("Setting API Status to 1...")

		step.Status = 1
		apiResults[strconv.Itoa(i+1)] = step
		workflowClaim.WorkflowResults = apiResults
		workflowClaim.CurrentStatus = 1
		workflowClaim.Save()

		fmt.Println("Loading Step API...")

		stepAPI, stepAPIErr = api.GetAPIFromDB(targetWorkflow.Steps[i].DeviceAccount, targetWorkflow.Steps[i].APICall)

		if stepAPIErr != nil {
			fmt.Println(stepAPIErr)
			step.Error = stepAPIErr.Error()
			step.Status = -1
			apiResults[strconv.Itoa(i+1)] = step
			workflowClaim.WorkflowResults = apiResults
			workflowClaim.CurrentStatus = -1
			workflowClaim.Save()
			// util.RespondWithError(w, http.StatusBadRequest, "error loading target API")
			return
		}

		step.APICall = stepAPI.Name
		step.APIAccount = stepAPI.DeviceAccount

		fmt.Println("Beginning VarMap Loop...")

		// For each Variable Map in "VARMAP"
		for j := 0; j < len(targetWorkflow.Steps[i].VarMap); j++ {
			for key, val := range targetWorkflow.Steps[i].VarMap[j] {
				left := strings.Index(key, "{")
				right := strings.Index(key, "}")

				stepFrom := key[(left + 1):right]
				fromMap := key[(right + 1):]

				fmt.Println("From Step: ", stepFrom)
				fmt.Println("From Map: ", fromMap)
				fmt.Println("APIResults:")
				fmt.Println(apiResults)

				if stepFrom == "0" {
					setData = gjson.Get(stepZero, fromMap)
				} else if stepFrom == "s" {
					var gString gjson.Result
					gString.Str = fromMap
					gString.Type = gjson.String
					setData = gString
				} else if stepFrom == "n" {
					var gString gjson.Result
					gString.Num, _ = strconv.ParseFloat(fromMap, 64)
					gString.Type = gjson.Number
					setData = gString
				} else {
					fmt.Println("Res Body: ", apiResults[stepFrom].ResBody)
					setData = gjson.Get(apiResults[stepFrom].ResBody, fromMap)
				}

				var typedData interface{}

				stepAPI.Body = strings.Replace(stepAPI.Body, "\n", "", -1)
				stepAPI.Body = strings.Replace(stepAPI.Body, "\t", "", -1)
				stepAPI.Body = strings.Replace(stepAPI.Body, "\r", "", -1)
				stepAPI.Body = strings.Replace(stepAPI.Body, "\\", "", -1)

				fmt.Println("STRIPPED API BODY:")
				fmt.Println(stepAPI.Body)

				fmt.Println("Determining TypeData...")
				fmt.Println("GJSON Results: ", gjson.Get(stepAPI.Body, val))

				if gjson.Get(stepAPI.Body, val).Exists() {
					switch dataKind := reflect.TypeOf(gjson.Get(stepAPI.Body, val).Value()).Kind(); dataKind {
					case reflect.Int:
						fmt.Println("Value: ", val)
						fmt.Println("Kind: ", dataKind)
						typedData = setData.Int()
					case reflect.Float64:
						fmt.Println("Value: ", val)
						fmt.Println("Kind: ", dataKind)
						typedData = setData.Float()
					case reflect.String:
						fmt.Println("Value: ", val)
						fmt.Println("Kind: ", dataKind)
						typedData = setData.String()
					default:
						fmt.Println("Value: ", val)
						fmt.Println("Unidentified Kind: ", dataKind)
					}
				} else {
					fmt.Println("Mapping Error!")
					fmt.Println("Step Body:")
					fmt.Println(stepAPI.Body)
					output := fmt.Sprintf("Map Value [%s]", val)
					fmt.Println(output)
					step.Error = "Invalid mapping data, target value does not exist"
					step.Status = -1
					apiResults[strconv.Itoa(i+1)] = step
					workflowClaim.WorkflowResults = apiResults
					workflowClaim.CurrentStatus = -1
					workflowClaim.Save()
					// util.RespondWithError(w, http.StatusBadRequest, "Invalid mapping data")
					return
				}

				fmt.Println("Setting StepAPI Body...")

				var sjsonSetErr error
				stepAPI.Body, sjsonSetErr = sjson.Set(stepAPI.Body, val, typedData)

				if sjsonSetErr != nil {
					fmt.Println(sjsonSetErr)
					step.Error = sjsonSetErr.Error()
					step.Status = -1
					apiResults[strconv.Itoa(i+1)] = step
					workflowClaim.WorkflowResults = apiResults
					workflowClaim.CurrentStatus = -1
					workflowClaim.Save()
					// util.RespondWithError(w, http.StatusBadRequest, "error mapping api variables")
					return
				}
			}
		}

		fmt.Println("Updated Body: ", stepAPI.Body)
		step.ReqBody = stepAPI.Body

		apiResp, apiErr := api.CallAPI(stepAPI)

		if apiErr != nil {
			fmt.Println(apiErr)
			step.Error = apiErr.Error()
			step.Status = -1
			apiResults[strconv.Itoa(i+1)] = step
			workflowClaim.WorkflowResults = apiResults
			workflowClaim.CurrentStatus = -1
			workflowClaim.Save()
			// util.RespondWithError(w, http.StatusBadRequest, "error executing API")
			return
		}

		fmt.Println("API Response:")
		fmt.Println(apiResp)

		defer apiResp.Body.Close()

		respBody, respErr := ioutil.ReadAll(apiResp.Body)

		if respErr != nil {
			fmt.Println(respErr)
			step.Error = respErr.Error()
			step.Status = -1
			apiResults[strconv.Itoa(i+1)] = step
			workflowClaim.WorkflowResults = apiResults
			workflowClaim.CurrentStatus = -1
			workflowClaim.Save()
			// util.RespondWithError(w, http.StatusBadRequest, "error reading response body")
			return
		}

		fmt.Println("API Response Body:")
		fmt.Println(string(respBody))

		step.ResBody = string(respBody)

		// bodyObject := make(map[string]interface{})
		// marshalErr := json.Unmarshal(respBody, &bodyObject)

		// if marshalErr != nil {
		// 	fmt.Println(marshalErr)
		// 	step.Status = -1
		// 	step.Error = marshalErr.Error()
		// 	apiResults[strconv.Itoa(i+1)] = step
		// 	workflowClaim.WorkflowResults = apiResults
		// 	workflowClaim.CurrentStatus = -1
		// 	workflowClaim.Save()
		// 	// util.RespondWithError(w, http.StatusBadRequest, "error parsing response body")
		// 	return
		// }

		step.Status = 2
		apiResults[strconv.Itoa(i+1)] = step
		workflowClaim.WorkflowResults = apiResults
		workflowClaim.CurrentStatus = 2
		workflowClaim.Save()
	}

	// util.RespondwithJSON(w, http.StatusOK, apiResults)
}
