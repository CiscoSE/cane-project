package workflow

import (
	"cane-project/api"
	"cane-project/database"
	"cane-project/model"
	"cane-project/util"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	//"github.com/mongodb/mongo-go-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/fatih/structs"
	"github.com/tidwall/sjson"

	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/gjson"

	"github.com/go-chi/chi"
	//"github.com/mongodb/mongo-go-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Workflow Alias
type Workflow model.Workflow

// JSONBody Alias
type JSONBody map[string]interface{}

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

	util.RespondwithString(w, http.StatusCreated, "")
}

// UpdateWorkflow Function
func UpdateWorkflow(w http.ResponseWriter, r *http.Request) {
	var jBody JSONBody

	filter := primitive.M{
		"name": chi.URLParam(r, "workflowname"),
	}

	_, findErr := database.FindOne("workflows", "workflow", filter)

	if findErr != nil {
		util.RespondWithError(w, http.StatusBadRequest, "workflow not found")
		return
	}

	decodeErr := json.NewDecoder(r.Body).Decode(&jBody)

	if decodeErr != nil {
		fmt.Println(decodeErr)
		util.RespondWithError(w, http.StatusBadRequest, "error decoding json body")
		return
	}

	// _, replaceErr := database.ReplaceOne("workflows", "workflow", filter, primitive.M(jBody))

	// if replaceErr != nil {
	// 	fmt.Println(replaceErr)
	// 	util.RespondWithError(w, http.StatusBadRequest, "error updating workflow in database")
	// 	return
	// }

	deleteErr := database.Delete("workflows", "workflow", filter)

	if deleteErr != nil {
		fmt.Println(deleteErr)
		util.RespondWithError(w, http.StatusBadRequest, "error deleting workflow in database")
		return
	}

	_, replaceErr := database.Save("workflows", "workflow", primitive.M(jBody))

	if replaceErr != nil {
		fmt.Println(replaceErr)
		util.RespondWithError(w, http.StatusBadRequest, "error re-creating workflow in database")
		return
	}

	util.RespondwithString(w, http.StatusOK, "")
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

// ToWorkflow Function
func (j JSONBody) ToWorkflow() (Workflow, error) {
	var w Workflow

	mapErr := mapstructure.Decode(j, &w)

	if mapErr != nil {
		return w, errors.New("error unmarshaling user details")
	}

	return w, nil
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
	// var setData gjson.Result
	var stepAPI model.API
	var stepAPIErr error

	apiResults := make(map[string]model.StepResult)
	varPool := make(map[string]map[string]string)
	zeroMap := make(map[string]interface{})

	bodyBuilder := ""
	stepHeader := make(map[string]string)
	var stepQuery url.Values

	fmt.Println("Beginning Step Loop...")

	fmt.Println("Step Zero Body:")
	fmt.Println(stepZero)
	fmt.Println("Step Zero Body Length:")
	fmt.Println(len(stepZero))

	fmt.Println("Unmarshal Step Zero BODY to MAP")

	if len(stepZero) != 0 {
		decoder := json.NewDecoder(strings.NewReader(stepZero))
		decoder.UseNumber()
		zeroErr := decoder.Decode(&zeroMap)

		if zeroErr != nil {
			fmt.Println("Error unmarshalling Zero BODY to MAP!")
			fmt.Println(zeroErr)
			return
		}
	} else {
		fmt.Println("Empty StepZero BODY!")
	}

	fmt.Println("Zero MAP:")
	fmt.Println(zeroMap)

	// Add Variables from ZeroMap to VarPool
	fmt.Println("Adding Variables from ZeroMap...")

	for key, val := range zeroMap {
		fmt.Print("Mapping Zero Variable: ")
		fmt.Println(val)

		switch val.(type) {
		case int, int32, int64:
			fmt.Println("INT")
			intVal := strconv.Itoa(val.(int))
			fmt.Println("Mapping: " + intVal + " as (int)")
			varPool[key] = map[string]string{intVal: "int"}
		case json.Number:
			fmt.Println("JSON NUMBER")
			jnumVal := fmt.Sprintf("%v", val)
			fmt.Println("Mapping: " + jnumVal + " as (int)")
			varPool[key] = map[string]string{jnumVal: "int"}
		case float32, float64:
			fmt.Println("FLOAT")
			floatVal := fmt.Sprintf("%f", val)
			fmt.Println("Mapping: " + floatVal + " as (float)")
			varPool[key] = map[string]string{floatVal: "float"}
		case bool:
			fmt.Println("BOOL")
			fmt.Println("Mapping: " + strconv.FormatBool(val.(bool)) + " as (bool)")
			varPool[key] = map[string]string{strconv.FormatBool(val.(bool)): "bool"}
		case string:
			fmt.Println("STRING")
			fmt.Println("Mapping: " + val.(string) + " as (string)")
			varPool[key] = map[string]string{val.(string): "string"}
		default:
			fmt.Println("UNKNOWN")
			fmt.Println("Unknown: " + val.(string) + " type (" + reflect.TypeOf(val).String() + ")")
		}
	}

	fmt.Println("VarPool MAP:")
	fmt.Println(varPool)

	// For each step in "STEPS"
	for i := 0; i < len(targetWorkflow.Steps); i++ {
		var step model.StepResult

		stepMethod := targetWorkflow.Steps[i].Verb
		fmt.Println("Step Method: " + stepMethod)

		fmt.Println("Setting API Status to 1...")

		step.Status = 1
		apiResults[strconv.Itoa(i+1)] = step
		workflowClaim.WorkflowResults = apiResults
		workflowClaim.CurrentStatus = 1
		workflowClaim.Save()

		fmt.Println("Loading Step API...")

		stepAPI, stepAPIErr = api.GetAPIFromDB(targetWorkflow.Steps[i].DeviceAccount, targetWorkflow.Steps[i].APICall)

		if stepAPIErr != nil {
			fmt.Println("Error getting API from DB...")
			fmt.Println(stepAPIErr)
			step.Error = stepAPIErr.Error()
			step.Status = -1
			apiResults[strconv.Itoa(i+1)] = step
			workflowClaim.WorkflowResults = apiResults
			workflowClaim.CurrentStatus = -1
			workflowClaim.Save()
			return
		}

		step.API = structs.Map(stepAPI)
		delete(step.API, "_id")
		step.Account = stepAPI.DeviceAccount

		// Build BODY string from Map
		fmt.Println("Building BODY from Map...")

		bodyBuilder = ""
		// stepBody = make(map[string]string)

		for bodyCount := 0; bodyCount < len(targetWorkflow.Steps[i].Body); bodyCount++ {
			for key, val := range targetWorkflow.Steps[i].Body[bodyCount] {
				if len(util.GetVariables(val)) > 0 {
					fmt.Print("Found variable(s): ")
					fmt.Println(util.GetVariables(val))

					for _, variable := range util.GetVariables(val) {
						rawVar := variable
						rawVar = strings.Replace(rawVar, "{{", "", 1)
						rawVar = strings.Replace(rawVar, "}}", "", 1)

						if poolVal, ok := varPool[rawVar]; ok {
							for replaceVar := range poolVal {
								val = strings.Replace(val, variable, replaceVar, 1)
							}
						} else {
							fmt.Println("ERROR MISSING BODY VARIABLE!!!")
							// Fail Workflow Here
						}
					}
				}

				var newVal interface{}

				if tempVal, err := strconv.ParseInt(val, 10, 64); err == nil {
					fmt.Println("Updated BODY VAL is an INT!")
					newVal = tempVal
				} else if tempVal, err := strconv.ParseFloat(val, 64); err == nil {
					fmt.Println("Updated BODY VAL is a FLOAT!")
					newVal = tempVal
				} else if strings.ToLower(val) == "false" || strings.ToLower(val) == "true" {
					fmt.Println("Updated BODY VAL is a BOOL!")
					newVal, _ = strconv.ParseBool(val)
				} else {
					fmt.Println("Updated BODY VAL is a STRING!")
					newVal = val
				}

				bodyBuilder, _ = sjson.Set(bodyBuilder, key, newVal)
			}
		}

		fmt.Println("BodyBuider:")
		fmt.Println(bodyBuilder)
		fmt.Println("BodyBuider Length:")
		fmt.Println(len(bodyBuilder))

		fmt.Println("Parsing BODY to Map...")

		var stepBody = make(map[string]interface{})

		if len(bodyBuilder) != 0 {
			decoder := json.NewDecoder(strings.NewReader(bodyBuilder))
			decoder.UseNumber()
			bodyErr := decoder.Decode(&stepBody)

			if bodyErr != nil {
				fmt.Println("Error parsing BODY to Map!")
				return
			}
		} else {
			fmt.Println("Empty BodyBuilder, setting StepBody to {}")
		}

		fmt.Println("BODY:")
		fmt.Println(stepBody)

		// Build HEADER from Map
		fmt.Println("Building HEADER from Map...")

		stepHeader = make(map[string]string)

		for headerCount := 0; headerCount < len(targetWorkflow.Steps[i].Headers); headerCount++ {
			for key, val := range targetWorkflow.Steps[i].Headers[headerCount] {
				if len(util.GetVariables(val)) > 0 {
					fmt.Print("Found variable(s): ")
					fmt.Println(util.GetVariables(val))

					for _, variable := range util.GetVariables(val) {
						rawVar := variable
						rawVar = strings.Replace(rawVar, "{{", "", 1)
						rawVar = strings.Replace(rawVar, "}}", "", 1)

						if poolVal, ok := varPool[rawVar]; ok {
							for replaceVar := range poolVal {
								val = strings.Replace(val, variable, replaceVar, 1)
							}
						} else {
							fmt.Println("ERROR MISSING HEADER VARIABLE!!!")
							// Fail Workflow Here
						}
					}
				}

				stepHeader[key] = val
			}
		}

		fmt.Println("HEADER:")
		fmt.Println(stepHeader)

		step.ReqHeaders = stepHeader

		// Build QUERY from Map
		fmt.Println("Building QUERY from Map...")

		stepQuery = make(map[string][]string)

		for queryCount := 0; queryCount < len(targetWorkflow.Steps[i].Query); queryCount++ {
			for key, val := range targetWorkflow.Steps[i].Query[queryCount] {
				if len(util.GetVariables(val)) > 0 {
					fmt.Print("Found variable(s): ")
					fmt.Println(util.GetVariables(val))

					for _, variable := range util.GetVariables(val) {
						rawVar := variable
						rawVar = strings.Replace(rawVar, "{{", "", 1)
						rawVar = strings.Replace(rawVar, "}}", "", 1)

						if poolVal, ok := varPool[rawVar]; ok {
							for replaceVar := range poolVal {
								val = strings.Replace(val, variable, replaceVar, 1)
							}
						} else {
							fmt.Println("ERROR MISSING QUERY VARIABLE!!!")
							// Fail Workflow Here
						}
					}
				}

				stepQuery.Add(key, val)
			}
		}

		fmt.Println("QUERY:")
		fmt.Println(stepQuery)

		step.ReqQuery = stepQuery

		fmt.Println("Updated Body: ", stepBody) // This doesn't look right...
		step.ReqBody = stepBody                 // Why is this stepBody vs. stepAPI.Body???

		varMatch := regexp.MustCompile(`([{]{2}[a-zA-Z]*[}]{2}){1}`)
		searchPath := varMatch.FindString(step.API["path"].(string))

		for searchPath != "" {
			fmt.Println("SearchPath: " + searchPath)

			val := searchPath
			val = strings.Replace(val, "{{", "", 1)
			val = strings.Replace(val, "}}", "", 1)

			fmt.Println("Variable to Replace: " + val)
			fmt.Println("Current Variable Pool:")
			fmt.Println(varPool)

			if poolVal, ok := varPool[val]; ok {
				for replaceVar := range poolVal {
					fmt.Println("Replace Variable Value: " + replaceVar)
					step.API["path"] = strings.Replace(step.API["path"].(string), searchPath, replaceVar, 1)
				}
			} else {
				fmt.Println("Replace Variable Not Found!")
				step.API["path"] = strings.Replace(step.API["path"].(string), searchPath, "<error>", 1)
			}

			searchPath = varMatch.FindString(step.API["path"].(string))
		}

		fmt.Println("Updated API Path:")
		fmt.Println(step.API["path"])

		var targetAPI model.API
		var bodyString string
		mapstructure.Decode(step.API, &targetAPI)

		if len(step.ReqBody) > 0 {
			bodyBytes, _ := json.Marshal(step.ReqBody)
			bodyString = string(bodyBytes)
		} else {
			bodyString = ""
		}

		apiResp, apiErr := api.CallAPI(targetAPI, stepMethod, stepQuery, stepHeader, bodyString)

		if apiErr != nil {
			fmt.Println(apiErr)
			step.Error = apiErr.Error()
			step.Status = -1
			apiResults[strconv.Itoa(i+1)] = step
			workflowClaim.WorkflowResults = apiResults
			workflowClaim.CurrentStatus = -1
			workflowClaim.Save()
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
			return
		}

		fmt.Println("API Response Body:")
		fmt.Println(string(respBody))

		step.ResBody = string(respBody)

		fmt.Println("Response Status Code:")
		fmt.Println(apiResp.StatusCode)

		step.ResStatus = apiResp.StatusCode

		if apiResp.StatusCode > 299 {
			fmt.Println("Error! Status code > 299!")
			step.Error = "Error calling API"
			step.Status = -1
			apiResults[strconv.Itoa(i+1)] = step
			workflowClaim.WorkflowResults = apiResults
			workflowClaim.CurrentStatus = -1
			workflowClaim.Save()
			return
		}

		if !gjson.Valid(step.ResBody) {
			fmt.Println("GJSON Reports ResBody is invalid JSON!")
		}

		// Extract VARIABLES from Response BODY
		fmt.Println("Extracting VARIABLES from Reponse Body...")

		for varCount := 0; varCount < len(targetWorkflow.Steps[i].Variables); varCount++ {
			for key, val := range targetWorkflow.Steps[i].Variables[varCount] {
				fmt.Println("Extracting Variable: " + val)

				varValue := gjson.Get(step.ResBody, val)

				if varValue.Exists() {
					fmt.Println("(" + val + ") Found! Value: " + varValue.String())
					fmt.Println("GJSON Type:" + string(varValue.Type.String()))

					switch varKind := reflect.TypeOf(varValue.Value()).Kind(); varKind {
					case reflect.Int:
						fmt.Println("Value: ", val)
						fmt.Println("Kind: ", varKind)
						varPool[key] = map[string]string{varValue.String(): "int"}
					case reflect.Float64:
						fmt.Println("Value: ", val)
						fmt.Println("Kind: ", varKind)
						if strings.ContainsAny(varValue.String(), ".") {
							varPool[key] = map[string]string{varValue.String(): "float"}
						} else {
							fmt.Println("No decimal, storing as INT...")
							varPool[key] = map[string]string{varValue.String(): "int"}
						}
					case reflect.String:
						fmt.Println("Value: ", val)
						fmt.Println("Kind: ", varKind)
						varPool[key] = map[string]string{varValue.String(): "string"}
					default:
						fmt.Println("Value: ", val)
						fmt.Println("Unidentified Kind: ", varKind)
					}
				} else {
					fmt.Println("(" + val + ") not found in Response Body!")
				}
			}
		}

		// Complete Workflow Step
		fmt.Println("Workflow Step (" + strconv.Itoa(i) + ") successfully completed!")

		step.Status = 2
		apiResults[strconv.Itoa(i+1)] = step
		workflowClaim.WorkflowResults = apiResults
		// workflowClaim.CurrentStatus = 2
		workflowClaim.Save()
	}

	// Complete Workflow
	fmt.Println("Workflow Execution successfully completed!")

	workflowClaim.CurrentStatus = 2
	workflowClaim.Save()
}
