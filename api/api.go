package api

import (
	"cane-project/account"
	"cane-project/auth"
	"cane-project/database"
	"cane-project/model"
	"cane-project/util"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/fatih/structs"
	"github.com/go-chi/chi"
	"github.com/mitchellh/mapstructure"

	//"github.com/mongodb/mongo-go-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/bson/primitive"
	//"github.com/mongodb/mongo-go-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// API Alias
type API model.API

// JSONBody Alias
type JSONBody map[string]interface{}

// APITypes Slice
var APITypes = []string{
	"XML",
	"JSON",
}

// APIMethods Slice
var APIMethods = []string{
	"POST",
	"GET",
	"PATCH",
	"DELETE",
}

// CreateAPI Function
func CreateAPI(w http.ResponseWriter, r *http.Request) {
	var api API

	decodeErr := json.NewDecoder(r.Body).Decode(&api)

	if decodeErr != nil {
		fmt.Println(decodeErr)
		util.RespondWithError(w, http.StatusBadRequest, "error decoding json body")
		return
	}

	// Cleanup Body & Method fields
	api.Body = strings.Replace(api.Body, "\n", "", -1)
	api.Body = strings.Replace(api.Body, "\\", "", -1)
	// api.Method = strings.ToUpper(api.Method)

	validErr := api.Valid()

	if validErr != nil {
		fmt.Println(validErr)
		util.RespondWithError(w, http.StatusBadRequest, validErr.Error())
		return
	}

	existFilter := primitive.M{
		"name": api.Name,
	}

	_, existErr := database.FindOne("apis", api.DeviceAccount, existFilter)

	if existErr == nil {
		util.RespondWithError(w, http.StatusBadRequest, "api already exists")
		return
	}

	_, saveErr := database.Save("apis", api.DeviceAccount, api)

	if saveErr != nil {
		util.RespondWithError(w, http.StatusBadRequest, "error saving api to database")
		return
	}

	util.RespondwithString(w, http.StatusCreated, "")
}

// UpdateAPI Function
func UpdateAPI(w http.ResponseWriter, r *http.Request) {
	var currAPI API
	var jBody JSONBody

	apiAccount := chi.URLParam(r, "devicename")
	apiName := chi.URLParam(r, "apiname")

	decodeErr := json.NewDecoder(r.Body).Decode(&jBody)

	if decodeErr != nil {
		fmt.Println(decodeErr)
		util.RespondWithError(w, http.StatusBadRequest, "error decoding json body")
		return
	}

	api, apiErr := jBody.ToAPI()

	if apiErr != nil {
		fmt.Println(apiErr)
		util.RespondWithError(w, http.StatusBadRequest, apiErr.Error())
		return
	}

	// Cleanup Body & Method fields
	api.Body = strings.Replace(api.Body, "\n", "", -1)
	api.Body = strings.Replace(api.Body, "\\", "", -1)
	// api.Method = strings.ToUpper(api.Method)

	accountFilter := primitive.M{
		"name": apiAccount,
	}

	_, accountErr := database.FindOne("accounts", "devices", accountFilter)

	if accountErr != nil {
		fmt.Println(accountErr)
		util.RespondWithError(w, http.StatusBadRequest, "no such account")
		return
	}

	loadFilter := primitive.M{
		"name": apiName,
	}

	loadVal, loadErr := database.FindOne("apis", apiAccount, loadFilter)

	if loadErr != nil {
		fmt.Println(loadErr)
		util.RespondWithError(w, http.StatusBadRequest, "no such api")
		return
	}

	mapstructure.Decode(loadVal, &currAPI)

	// if api.Method != "" {
	// 	currAPI.Method = api.Method
	// }

	if api.Path != "" {
		currAPI.Path = api.Path
	}

	if api.Body != "" {
		currAPI.Body = api.Body
	}

	if api.Type != "" {
		currAPI.Type = api.Type
	}

	// updatedAPI := structs.Map(originalAPI)

	// delete(updatedAPI, "ID")

	validErr := currAPI.Valid()

	if validErr != nil {
		fmt.Println(validErr)
		util.RespondWithError(w, http.StatusBadRequest, validErr.Error())
		return
	}

	_, saveErr := database.FindAndReplace("apis", apiAccount, loadFilter, structs.Map(currAPI))

	if saveErr != nil {
		fmt.Println(saveErr)
		util.RespondWithError(w, http.StatusBadRequest, "error saving api")
		return
	}

	util.RespondwithString(w, http.StatusOK, "")
}

// DeleteAPI Function
func DeleteAPI(w http.ResponseWriter, r *http.Request) {
	apiAccount := chi.URLParam(r, "devicename")
	apiName := chi.URLParam(r, "apiname")

	filter := primitive.M{
		"name": apiName,
	}

	_, findErr := database.FindOne("apis", apiAccount, filter)

	if findErr != nil {
		fmt.Println(findErr)
		util.RespondWithError(w, http.StatusBadRequest, "api not found")
		return
	}

	deleteErr := database.Delete("apis", apiAccount, filter)

	if deleteErr != nil {
		fmt.Println(deleteErr)
		util.RespondWithError(w, http.StatusBadRequest, "api not found")
		return
	}

	// Add a check if this is the last API, if so, drop the collection

	util.RespondwithString(w, http.StatusOK, "")
}

// LoadAPI Function
func LoadAPI(w http.ResponseWriter, r *http.Request) {
	apiAccount := chi.URLParam(r, "account")
	apiName := chi.URLParam(r, "name")

	getAPI, getErr := GetAPIFromDB(apiAccount, apiName)

	if getErr != nil {
		fmt.Println(getErr)
		util.RespondWithError(w, http.StatusBadRequest, "api not found")
		return
	}

	util.RespondwithJSON(w, http.StatusOK, getAPI)
}

// GetAPIFromDB Function
func GetAPIFromDB(apiAccount string, apiName string) (model.API, error) {
	var api model.API

	filter := primitive.M{
		"name": apiName,
	}

	foundVal, foundErr := database.FindOne("apis", apiAccount, filter)

	if foundErr != nil {
		fmt.Println(foundErr)
		return api, foundErr
	}

	mapErr := mapstructure.Decode(foundVal, &api)

	if mapErr != nil {
		fmt.Println(mapErr)
		return api, mapErr
	}

	return api, nil
}

// CallAPI Function
func CallAPI(targetAPI model.API, stepMethod string, queryParams url.Values, headerVals map[string]string, stepBody string) (*http.Response, error) {
	transport := &http.Transport{}
	client := &http.Client{}

	var targetDevice model.DeviceAccount
	var req *http.Request
	var reqErr error

	proxyURL, err := url.Parse(util.ProxyURL)
	if err != nil {
		fmt.Println("Invalid proxy URL format: ", util.ProxyURL)
	}

	deviceFilter := primitive.M{
		"name": targetAPI.DeviceAccount,
	}

	deviceResult, deviceDBErr := database.FindOne("accounts", "devices", deviceFilter)

	if deviceDBErr != nil {
		fmt.Println(deviceDBErr)
		fmt.Println("Error loading device for CallAPI")
		return nil, errors.New("error accessing database")
	}

	deviceDecodeErr := mapstructure.Decode(deviceResult, &targetDevice)

	if deviceDecodeErr != nil {
		fmt.Println(deviceDecodeErr)
		fmt.Println("Error decoding device for CallAPI")
		return nil, errors.New("invalid device account")
	}

	switch targetDevice.AuthType {
	case "none":
		req, reqErr = auth.NoAuth(targetAPI, stepMethod, queryParams, stepBody)
	case "basic":
		req, reqErr = auth.BasicAuth(targetAPI, stepMethod, queryParams, stepBody)
	case "apikey":
		req, reqErr = auth.APIKeyAuth(targetAPI, stepMethod, queryParams, stepBody)
	case "rfc3447":
		req, reqErr = auth.RFC3447Auth(targetAPI, stepMethod, queryParams, stepBody)
	default:
		fmt.Println("Invalid AuthType!")
		return nil, errors.New("invalid authtype")
	}

	if reqErr != nil {
		fmt.Println("Error getting request for CallAPI")
		fmt.Println(reqErr)
		return nil, errors.New("error creating api request")
	}

	// Ignore self-sgned SSL certificates if set globally
	if util.IgnoreSSL {
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	// Add proxy settings to the HTTP Transport object
	if targetDevice.RequireProxy {
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	// Append headers to HTTP request
	for key, value := range headerVals {
		req.Header.Add(key, value)
	}

	fmt.Println("Request Path: " + req.URL.String())

	client = &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second,
	}

	resp, respErr := client.Do(req)

	if respErr != nil {
		fmt.Println("Errored when sending request to the server!")
		fmt.Println(respErr)
		return nil, errors.New("error calling api")
	}

	return resp, nil
}

// GetAPI Function
func GetAPI(w http.ResponseWriter, r *http.Request) {
	var returnAPI model.API

	device := chi.URLParam(r, "devicename")
	api := chi.URLParam(r, "apiname")

	filter := primitive.M{
		"name": api,
	}

	findVal, findErr := database.FindOne("apis", device, filter)

	if findErr != nil {
		fmt.Println(findErr)
		util.RespondWithError(w, http.StatusBadRequest, "api not found")
		return
	}

	mapstructure.Decode(findVal, &returnAPI)

	util.RespondwithJSON(w, http.StatusOK, returnAPI)
}

// GetAPIs Function
func GetAPIs(w http.ResponseWriter, r *http.Request) {
	var opts options.FindOptions
	device := chi.URLParam(r, "devicename")

	var apis []string

	findVal, findErr := database.FindAll("apis", device, primitive.M{}, opts)

	if findErr != nil {
		fmt.Println(findErr)
		util.RespondWithError(w, http.StatusBadRequest, "device not found")
		return
	}

	for key := range findVal {
		apis = append(apis, findVal[key]["name"].(string))
	}

	if apis == nil {
		fmt.Println("invalid device or no apis")
		util.RespondWithError(w, http.StatusNoContent, "invalid account or no apis")
		return
	}

	util.RespondwithJSON(w, http.StatusOK, map[string][]string{"apis": apis})
}

// ToAPI Function
func (j JSONBody) ToAPI() (API, error) {
	var a API

	mapErr := mapstructure.Decode(j, &a)

	if mapErr != nil {
		return a, errors.New("error unmarshaling api details")
	}

	return a, nil
}

// Valid Dunction
func (a API) Valid() error {

	if a.Name == "" {
		return errors.New("name cannot be empty")
	}

	if a.DeviceAccount == "" {
		return errors.New("deviceAccount cannot be empty")
	}

	_, deviceErr := account.GetDeviceFromDB(a.DeviceAccount)

	if deviceErr != nil {
		return errors.New("target deviceAccount does not exist")
	}

	// if !ValidMethod(a.Method) {
	// 	return errors.New("invalid method")
	// }

	if a.Path == "" {
		return errors.New("url cannot be empty")
	}

	// if a.Body == "" && (a.Method == "POST" || a.Method == "PATCH") {
	// 	return errors.New("body cannot be empty with POST & PATCH")
	// }

	if !ValidType(a.Type) {
		return errors.New("type cannot be empty")
	}

	return nil
}

// ValidMethod Function
func ValidMethod(method string) bool {
	for _, apiMethod := range APIMethods {
		if method == apiMethod {
			return true
		}
	}

	return false
}

// ValidType Function
func ValidType(aType string) bool {
	for _, apiType := range APITypes {
		if aType == apiType {
			return true
		}
	}

	return false
}
