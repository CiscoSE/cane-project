package routing

import (
	"cane-project/account"
	"cane-project/api"
	"cane-project/database"
	"cane-project/jwt"
	"cane-project/model"
	"cane-project/util"
	"cane-project/workflow"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	//"github.com/mongodb/mongo-go-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/go-chi/jwtauth"
	"github.com/mitchellh/mapstructure"

	//"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/tidwall/gjson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Router Variable
var Router *chi.Mux

func init() {
	Router = chi.NewRouter()
}

// Routers Function
func Routers() {
	var opts options.FindOptions
	var iterVals []model.RouteValue
	Router = chi.NewRouter()

	filter := primitive.M{}
	foundVals, _ := database.FindAll("routing", "routes", filter, opts)
	mapstructure.Decode(foundVals, &iterVals)

	fmt.Println("Updating routes...")

	cors := cors.New(cors.Options{
		// AllowedOrigins: []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})

	Router.Use(cors.Handler)

	// Public Default Routes
	Router.Post("/login", account.Login)
	// Router.Get("/testPath/*", TestPath)
	Router.Post("/testJSON", JSONTest)
	Router.Post("/testXML", XMLTest)
	Router.Post("/testGJSON", TestGJSON)
	Router.Get("/testQuery", TestQuery)

	// Private Default Routes
	Router.Group(func(r chi.Router) {
		/* Allow Cross-Origin Requests */
		r.Use(cors.Handler)

		/* JWT Token Security */
		r.Use(jwtauth.Verifier(jwt.TokenAuth))
		r.Use(jwtauth.Authenticator)

		/* Final Routes */
		/* /user */
		r.Get("/user", account.GetUsers)
		r.Get("/user/{username}", account.GetUser)
		r.Post("/user", account.CreateUser)
		r.Patch("/user/{username}", account.UpdateUser)
		r.Delete("/user/{username}", account.DeleteUser)

		/* /device */
		r.Get("/device", account.GetDevices)
		r.Get("/device/{devicename}", account.GetDevice)
		r.Post("/device", account.CreateDevice)
		r.Patch("/device/{devicename}", account.UpdateDevice)
		r.Delete("/device/{devicename}", account.DeleteDevice)

		/* /api */
		r.Get("/api/{devicename}", api.GetAPIs)
		r.Get("/api/{devicename}/{apiname}", api.GetAPI)
		r.Post("/api", api.CreateAPI)
		r.Patch("/api/{devicename}/{apiname}", api.UpdateAPI)
		r.Delete("/api/{devicename}/{apiname}", api.DeleteAPI)

		/* /workflow */
		r.Get("/workflow", workflow.GetWorkflows)
		r.Get("/workflow/{workflowname}", workflow.GetWorkflow)
		r.Post("/workflow", workflow.CreateWorkflow)
		r.Patch("/workflow/{workflowname}", workflow.UpdateWorkflow)
		r.Delete("/workflow/{workflowname}", workflow.DeleteWorkflow)
		r.Post("/workflow/{workflowname}", workflow.CallWorkflow)

		/* /claim */
		r.Get("/claim", workflow.GetClaims)
		r.Get("/claim/{claimcode}", workflow.GetClaim)

		/* Dynamic Paths for Pass-Through & Workflow */
		r.HandleFunc("/{devicename}/*", PassThroughAPI)
		// r.Get("/testPath/*", TestPath)

		/* Old Routes (Testing) */
		// r.Post("/addRoute", AddRoutes)
		r.Post("/parseVars", ParseVars)
		r.Post("/test/{devicename}/{apiname}", TestCallAPI)
		// r.Post("/validateToken", account.ValidateUserToken)
		// r.Patch("/updateToken/{user}", account.RefreshToken)
	})

	// Dynamic Routes
	for i := range iterVals {
		routeVal := iterVals[i]

		if routeVal.Enable {
			newRoute := "/v" + strconv.Itoa(routeVal.Version) + "/" + routeVal.Category + "/" + routeVal.Route
			newMessage := routeVal.Message

			if routeVal.Verb == "get" {
				Router.Get(newRoute, func(w http.ResponseWriter, r *http.Request) {
					util.RespondwithJSON(w, http.StatusCreated, newMessage)
				})
			} else if routeVal.Verb == "post" {
				Router.Post(newRoute, func(w http.ResponseWriter, r *http.Request) {
					postJSON := make(map[string]interface{})
					err := json.NewDecoder(r.Body).Decode(&postJSON)

					fmt.Println(postJSON)

					if err != nil {
						panic(err)
					}

					util.RespondwithJSON(w, http.StatusCreated, postJSON)
				})
			}
		}
	}
}

// ParseVars function
func ParseVars(w http.ResponseWriter, r *http.Request) {
	bodyBytes, bodyErr := ioutil.ReadAll(r.Body)

	// fmt.Println(string(bodyReader))

	if bodyErr != nil {
		fmt.Println(bodyErr)
		util.RespondWithError(w, http.StatusBadRequest, "invalid data")
		return
	}

	if model.IsJSON(string(bodyBytes)) {
		var input map[string]interface{}

		json.Unmarshal(bodyBytes, &input)

		fixed := model.JSONNode(input).StripJSON()

		util.RespondwithJSON(w, http.StatusCreated, fixed)
	}

	if model.IsXML(string(bodyBytes)) {
		x, xmlErr := model.XMLfromBytes(bodyBytes)

		if xmlErr != nil {
			fmt.Println(xmlErr)
			util.RespondWithError(w, http.StatusBadRequest, "invalid xml")
		}

		mapString := x.XMLtoJSON()

		fixed := model.JSONNode(mapString).StripJSON()

		jBytes, _ := json.MarshalIndent(fixed, "", "  ")

		// jString := string(jBytes)

		j, jsonErr := model.JSONfromBytes(jBytes)

		if jsonErr != nil {
			fmt.Println(jsonErr)
			util.RespondWithError(w, http.StatusBadRequest, "invalid json")
		}

		fixedXML := j.ToXML()

		util.RespondwithJSON(w, http.StatusCreated, fixedXML)
	}
}

// PassThroughAPI function
func PassThroughAPI(w http.ResponseWriter, r *http.Request) {
	var newAPI model.API

	uri := chi.URLParam(r, "*")
	device := chi.URLParam(r, "devicename")
	queryParams := r.URL.Query()
	routeContext := chi.RouteContext(r.Context())
	routePattern := routeContext.RoutePattern()
	routeMethod := routeContext.RouteMethod

	account := strings.Replace(routePattern, "/", "", -1)
	account = strings.Replace(account, "*", "", -1)

	bodyBytes, bodyErr := ioutil.ReadAll(r.Body)
	bodyString := string(bodyBytes)

	if bodyErr != nil {
		util.RespondWithError(w, http.StatusBadRequest, bodyErr.Error())
		return
	}

	newAPI.DeviceAccount = device
	// newAPI.Method = routeMethod

	fmt.Println("Pass Through Method: " + routeMethod)

	if len(uri) > 0 {
		newAPI.Path += "/"
	}

	newAPI.Path += uri
	newAPI.Body = bodyString

	if model.IsJSON(bodyString) {
		newAPI.Type = "json"
	} else if model.IsXML(bodyString) {
		newAPI.Type = "xml"
	} else {
		newAPI.Type = "error"
	}

	resp, err := api.CallAPI(newAPI, routeMethod, queryParams, nil, bodyString)

	if err != nil {
		util.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	bodyMap := map[string]interface{}{}
	bodyArray := []map[string]interface{}{}

	mapErr := json.Unmarshal(respBody, &bodyMap)
	arrayErr := json.Unmarshal(respBody, &bodyArray)

	if mapErr == nil {
		util.RespondwithJSON(w, resp.StatusCode, bodyMap)
		return
	}

	if arrayErr == nil {
		util.RespondwithJSON(w, resp.StatusCode, bodyArray)
		return
	}

	util.RespondWithError(w, http.StatusBadRequest, "error unmarshaling response body")
	return
}

// AddRoutes function
func AddRoutes(w http.ResponseWriter, r *http.Request) {
	var target model.RouteValue

	json.NewDecoder(r.Body).Decode(&target)

	if !(ValidateRoute(target)) {
		util.RespondWithError(w, http.StatusBadRequest, "invalid route")
		return
	}

	fmt.Println("Adding routes to database...")

	postID, postErr := database.Save("routing", "routes", target)

	if postErr != nil {
		fmt.Println(postErr)
		util.RespondWithError(w, http.StatusBadRequest, "failed saving route")
		return
	}

	Routers()

	util.RespondwithJSON(w, http.StatusCreated, postID)
}

// ValidateRoute Function
func ValidateRoute(route model.RouteValue) bool {
	verbs := []string{"get", "post", "patch", "delete"}
	categories := []string{"network", "compute", "storage", "security", "virtualization", "cloud"}

	if !(util.StringInSlice(verbs, route.Verb)) {
		return false
	}

	if !(util.StringInSlice(categories, route.Category)) {
		return false
	}

	return true
}

// TestCallAPI Function
func TestCallAPI(w http.ResponseWriter, r *http.Request) {
	var apiBody map[string]interface{}
	var apiResponse map[string]interface{}
	var callAPI model.API

	deviceName := chi.URLParam(r, "devicename")
	apiName := chi.URLParam(r, "apiname")

	bodyBytes, bodyErr := ioutil.ReadAll(r.Body)

	if bodyErr != nil {
		util.RespondWithError(w, http.StatusBadRequest, bodyErr.Error())
		return
	}

	bodyString := string(bodyBytes)

	if len(bodyString) > 0 {
		decodeBodyErr := json.Unmarshal(bodyBytes, &apiBody)

		if decodeBodyErr != nil {
			fmt.Println(decodeBodyErr)
			util.RespondWithError(w, http.StatusBadRequest, "error decoding json body")
			return
		}
	}

	targetFilter := primitive.M{
		"name": apiName,
	}

	targetAPI, targetErr := database.FindOne("apis", deviceName, targetFilter)

	if targetErr != nil {
		fmt.Println(targetErr)
		util.RespondWithError(w, http.StatusBadRequest, "no such api")
		return
	}

	mapstructure.Decode(targetAPI, &callAPI)

	if len(bodyString) > 0 {
		callAPI.Body = bodyString
	}

	resp, respErr := api.CallAPI(callAPI, "", nil, nil, bodyString)

	if respErr != nil {
		fmt.Println(respErr)
		util.RespondWithError(w, http.StatusBadRequest, "error calling api")
		return
	}

	defer resp.Body.Close()
	respBytes, _ := ioutil.ReadAll(resp.Body)

	fmt.Println(string(respBytes))

	decodeRespErr := json.Unmarshal(respBytes, &apiResponse)

	if decodeRespErr != nil {
		fmt.Println(decodeRespErr)
		util.RespondWithError(w, http.StatusBadRequest, "error decoding json response")
		return
	}

	util.RespondwithJSON(w, http.StatusOK, apiResponse)
}

// APLTest Function
// func APLTest(w http.ResponseWriter, r *http.Request) {
// 	model.APLtoJSON()
// }

// JSONTest Function
func JSONTest(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}

	bodyBytes, bodyErr := ioutil.ReadAll(r.Body)

	if bodyErr != nil {
		fmt.Println(bodyErr)
		util.RespondWithError(w, http.StatusBadRequest, "error ready body")
	}

	json.Unmarshal(bodyBytes, &input)

	fixed := model.JSONNode(input).StripJSON()

	util.RespondwithJSON(w, http.StatusCreated, fixed)
}

// XMLTest Function
func XMLTest(w http.ResponseWriter, r *http.Request) {
	bodyReader, _ := ioutil.ReadAll(r.Body)

	x, xmlErr := model.XMLfromBytes(bodyReader)

	if xmlErr != nil {
		fmt.Println(xmlErr)
		util.RespondWithError(w, http.StatusBadRequest, "invalid xml")
	}

	mapString := x.XMLtoJSON()

	fmt.Println(x.XMLtoJSON())
	fmt.Println("-------------------------------")

	jBytes, _ := json.MarshalIndent(mapString, "", "  ")

	jString := string(jBytes)

	fmt.Println(jString)
	fmt.Println("-------------------------------")

	j, jsonErr := model.JSONfromBytes(jBytes)

	if jsonErr != nil {
		fmt.Println(jsonErr)
		util.RespondWithError(w, http.StatusBadRequest, "invalid json")
	}

	fmt.Println(j.ToXML())

	util.RespondwithJSON(w, http.StatusCreated, mapString)
}

// TestGJSON Function
func TestGJSON(w http.ResponseWriter, r *http.Request) {
	// var output map[string]interface{}

	// json := `{"name":{"first":"Janet","last":"Prichard"},"age":47}`

	bodyBytes, _ := ioutil.ReadAll(r.Body)
	bodyString := string(bodyBytes)

	// grab := "cuicoperationRequest.payload.cdata.ucsUuidPool.accountName"
	grab := "details.@location"

	value := gjson.Get(bodyString, grab)

	gJSON := value.String()

	// json.Unmarshal([]byte(gJSON), &output)

	util.RespondwithJSON(w, http.StatusCreated, map[string]interface{}{"value": gJSON})
}

// ClaimTest Function
func ClaimTest(w http.ResponseWriter, r *http.Request) {
	var testResult model.StepResult

	stepResults := make(map[string]model.StepResult)

	fmt.Println("Generating new claim...")
	claim := workflow.GenerateClaim()

	fmt.Println("Saving new claim...")
	claim.Save()

	fmt.Println("Loading fake step data...")

	testResult.Account = "testaccount"
	// testResult.API = "testcall"
	testResult.Error = ""
	// testResult.ReqBody = "{req_body}"
	testResult.ResBody = "{res_body}"
	testResult.Status = 2

	fmt.Println("Assigning fake step data to fake results...")
	stepResults["1"] = testResult

	fmt.Println("Assigning fake results to claim...")
	claim.WorkflowResults = stepResults

	fmt.Println("Saving updated claim...")
	claim.Save()

	util.RespondwithJSON(w, http.StatusCreated, map[string]interface{}{"claim": claim})
}

// TestQuery Function
func TestQuery(w http.ResponseWriter, r *http.Request) {
	query := url.Values{}

	query = r.URL.Query()

	fmt.Println(query)
}
