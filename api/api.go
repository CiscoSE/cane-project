package api

import (
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
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo/options"
)

// APITypes Variable
var APITypes []string

func init() {
	APITypes = []string{
		"xml",
		"json",
	}
}

// CreateAPI Function
func CreateAPI(w http.ResponseWriter, r *http.Request) {
	var api model.API

	json.NewDecoder(r.Body).Decode(&api)

	api.Body = strings.Replace(api.Body, "\n", "", -1)
	api.Body = strings.Replace(api.Body, "\\", "", -1)

	accountFilter := primitive.M{
		"name": api.DeviceAccount,
	}

	_, accountErr := database.FindOne("accounts", "devices", accountFilter)

	if accountErr != nil {
		fmt.Println(accountErr)
		util.RespondWithError(w, http.StatusBadRequest, "no such account")
		return
	}

	existFilter := primitive.M{
		"name": api.Name,
	}

	_, existErr := database.FindOne("apis", api.DeviceAccount, existFilter)

	if existErr == nil {
		fmt.Println(existErr)
		util.RespondWithError(w, http.StatusBadRequest, "api already exists")
		return
	}

	_, saveErr := database.Save("apis", api.DeviceAccount, api)

	if saveErr != nil {
		fmt.Println(saveErr)
		util.RespondWithError(w, http.StatusBadRequest, "error saving api")
		return
	}

	// api.ID = saveID.(primitive.ObjectID)

	// foundVal, _ := database.FindOne("apis", api.DeviceAccount, existFilter)

	util.RespondwithString(w, http.StatusCreated, "")
}

// UpdateAPI Function
func UpdateAPI(w http.ResponseWriter, r *http.Request) {
	var originalAPI model.API
	var patchAPI map[string]interface{}

	apiAccount := chi.URLParam(r, "devicename")
	apiName := chi.URLParam(r, "apiname")

	json.NewDecoder(r.Body).Decode(&patchAPI)

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

	mapstructure.Decode(loadVal, &originalAPI)

	originalAPI.Method = patchAPI["method"].(string)
	originalAPI.URL = patchAPI["url"].(string)
	originalAPI.Body = patchAPI["body"].(string)
	originalAPI.Type = patchAPI["type"].(string)

	updatedAPI := structs.Map(originalAPI)

	delete(updatedAPI, "ID")

	_, saveErr := database.FindAndReplace("apis", apiAccount, loadFilter, updatedAPI)

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
func CallAPI(targetAPI model.API) (*http.Response, error) {
	transport := &http.Transport{}
	client := &http.Client{}

	var targetDevice model.DeviceAccount
	var req *http.Request
	var reqErr error

	proxyURL, err := url.Parse(util.ProxyURL)
	if err != nil {
		fmt.Println("Invalid proxy URL format: ", util.ProxyURL)
	}

	// if util.IgnoreSSL {
	// 	transport = &http.Transport{
	// 		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	// 	}
	// }

	// Add proxy settings to the HTTP Transport object
	// if len(proxyURL.String()) > 0 {
	// 	transport.Proxy = http.ProxyURL(proxyURL)
	// }

	// client = &http.Client{
	// 	Transport: transport,
	// 	Timeout:   30 * time.Second,
	// }

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
		req, reqErr = auth.NoAuth(targetAPI)
	case "basic":
		req, reqErr = auth.BasicAuth(targetAPI)
	case "apikey":
		req, reqErr = auth.APIKeyAuth(targetAPI)
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

	client = &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
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
