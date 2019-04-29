package account

import (
	"cane-project/database"
	"cane-project/model"
	"cane-project/util"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/fatih/structs"

	//"github.com/mongodb/mongo-go-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mitchellh/mapstructure"

	"github.com/go-chi/chi"
	//"github.com/mongodb/mongo-go-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateDevice Function
func CreateDevice(w http.ResponseWriter, r *http.Request) {
	var jBody JSONBody

	decodeErr := json.NewDecoder(r.Body).Decode(&jBody)

	if decodeErr != nil {
		fmt.Println(decodeErr)
		util.RespondWithError(w, http.StatusBadRequest, "error decoding json body")
		return
	}

	if len(jBody) != 5 {
		util.RespondWithError(w, http.StatusBadRequest, "invalid number of keys in device body")
		return
	}

	device, deviceErr := jBody.ToDevice()

	if deviceErr != nil {
		fmt.Println(deviceErr)
		util.RespondWithError(w, http.StatusBadRequest, deviceErr.Error())
		return
	}

	deviceValidErr := device.Valid()

	if deviceValidErr != nil {
		util.RespondWithError(w, http.StatusBadRequest, deviceValidErr.Error())
		return
	}

	if DeviceExists(device.Name) {
		util.RespondWithError(w, http.StatusBadRequest, "existing device account")
		return
	}

	_, deviceSaveErr := database.Save("accounts", "devices", device)

	if deviceSaveErr != nil {
		util.RespondWithError(w, http.StatusBadRequest, "error saving device to database")
		return
	}

	util.RespondwithString(w, http.StatusCreated, "")
}

// UpdateDevice Function
func UpdateDevice(w http.ResponseWriter, r *http.Request) {
	var jBody JSONBody
	var value interface{}
	var ok bool

	decodeErr := json.NewDecoder(r.Body).Decode(&jBody)

	if decodeErr != nil {
		fmt.Println(decodeErr)
		util.RespondWithError(w, http.StatusBadRequest, "error decoding json body")
		return
	}

	device, deviceErr := jBody.ToDevice()

	if deviceErr != nil {
		fmt.Println(deviceErr)
		util.RespondWithError(w, http.StatusBadRequest, "error converting json to device")
		return
	}

	deviceFilter := primitive.M{
		"name": chi.URLParam(r, "devicename"),
	}

	loadDevice, loadDeviceErr := database.FindOne("accounts", "devices", deviceFilter)

	if loadDeviceErr != nil {
		util.RespondWithError(w, http.StatusBadRequest, "device not found")
		return
	}

	currDevice, currDeviceErr := JSONBody(loadDevice).ToDevice()

	if currDeviceErr != nil {
		util.RespondWithError(w, http.StatusBadRequest, "error unmarshaling device details")
		return
	}

	value, ok = jBody["name"]
	if ok {
		util.RespondWithError(w, http.StatusBadRequest, "cannot modify device name")
		return
	}

	if device.Name != "" {
		currDevice.Name = device.Name
	}

	if device.BaseURL != "" {
		currDevice.BaseURL = device.BaseURL
	}

	if device.AuthType != "" {
		currDevice.AuthType = device.AuthType
	}

	value, ok = jBody["requireProxy"]
	if ok {
		currDevice.RequireProxy = value.(bool)
	}

	if device.AuthObj != nil {
		currDevice.AuthObj = device.AuthObj
	}

	validErr := currDevice.Valid()

	if validErr != nil {
		util.RespondWithError(w, http.StatusBadRequest, validErr.Error())
		return
	}

	_, updatedErr := database.FindAndReplace("accounts", "devices", deviceFilter, structs.Map(currDevice))

	if updatedErr != nil {
		util.RespondWithError(w, http.StatusBadRequest, "error saving updated device to database")
		return
	}

	util.RespondwithString(w, http.StatusOK, "")
}

// DeleteDevice Function
func DeleteDevice(w http.ResponseWriter, r *http.Request) {
	deviceName := chi.URLParam(r, "devicename")

	deviceFilter := primitive.M{
		"name": deviceName,
	}

	_, findErr := GetDeviceFromDB(deviceName)

	if findErr != nil {
		util.RespondWithError(w, http.StatusBadRequest, "device not found")
		return
	}

	// Add in capability for query parameter force=true to remove device & all dependents
	_, depsErr := database.FindOne("apis", deviceName, deviceFilter)

	if depsErr == nil {
		util.RespondWithError(w, http.StatusBadRequest, "cannot delete device while dependent apis exist")
		return
	}

	deleteDeviceErr := database.Delete("accounts", "devices", deviceFilter)

	if deleteDeviceErr != nil {
		util.RespondWithError(w, http.StatusBadRequest, "device not found")
		return
	}

	util.RespondwithString(w, http.StatusOK, "")
}

// GetDevice Function
func GetDevice(w http.ResponseWriter, r *http.Request) {
	// var authType string

	filter := primitive.M{
		"name": chi.URLParam(r, "devicename"),
	}

	findDeviceVal, findDeviceErr := database.FindOne("accounts", "devices", filter)

	if findDeviceErr != nil {
		util.RespondWithError(w, http.StatusBadRequest, "device not found")
		return
	}

	delete(findDeviceVal, "_id")

	util.RespondwithJSON(w, http.StatusOK, findDeviceVal)
}

// GetDevices Function
func GetDevices(w http.ResponseWriter, r *http.Request) {
	var opts options.FindOptions
	var deviceList []string

	projection := primitive.M{
		"_id":  0,
		"name": 1,
	}

	opts.SetProjection(projection)

	findVals, findErr := database.FindAll("accounts", "devices", primitive.M{}, opts)

	if findErr != nil {
		util.RespondWithError(w, http.StatusBadRequest, "no devices found")
		return
	}

	for _, device := range findVals {
		deviceList = append(deviceList, device["name"].(string))
	}

	util.RespondwithJSON(w, http.StatusOK, map[string][]string{"devices": deviceList})
}

// GetDeviceID Function
func GetDeviceID(deviceName string) (primitive.ObjectID, error) {
	var deviceID primitive.ObjectID

	filter := primitive.M{
		"name": deviceName,
	}

	findVal, findErr := database.FindOne("accounts", "devices", filter)

	if findErr != nil {
		return deviceID, errors.New("device not found")
	}

	deviceID = findVal["_id"].(primitive.ObjectID)

	return deviceID, nil
}

// GetDeviceFromDB Function
func GetDeviceFromDB(deviceName string) (model.DeviceAccount, error) {
	var device model.DeviceAccount

	filter := primitive.M{
		"name": deviceName,
	}

	findVal, findErr := database.FindOne("accounts", "devices", filter)

	if findErr != nil {
		return device, errors.New("device not found")
	}

	mapErr := mapstructure.Decode(findVal, &device)

	if mapErr != nil {
		return device, errors.New("error unmarshaling device details")
	}

	return device, nil
}

// DeviceExists Function
func DeviceExists(deviceName string) bool {
	filter := primitive.M{
		"name": deviceName,
	}

	_, findErr := database.FindOne("accounts", "devices", filter)

	if findErr == nil {
		return true
	}

	return false
}

// ToDevice Function
func (j JSONBody) ToDevice() (Device, error) {
	var d Device

	mapErr := mapstructure.Decode(j, &d)

	if mapErr != nil {
		return d, errors.New("error unmarshaling device details")
	}

	return d, nil
}

// Valid Function for model.DeviceAccount
func (d *Device) Valid() error {
	if len(d.Name) == 0 {
		return errors.New("name cannot be empty")
	}

	if len(d.BaseURL) == 0 {
		return errors.New("url cannot be empty")
	}

	if len(d.AuthType) == 0 {
		return errors.New("authType cannot be empty")
	} else if !ValidAuthType(d.AuthType) {
		return errors.New("invalid authType")
	}

	authErr := AuthValid(d.AuthType, d.AuthObj)

	if authErr != nil {
		return authErr
	}

	return nil
}

// AuthValid Function
func AuthValid(authType string, authObj map[string]interface{}) error {
	switch authType {
	case "none":
		if len(authObj) != 0 {
			return errors.New("authbody provided with authtype of none")
		}
	case "basic":
		var auth BasicAuth

		if len(authObj) != 2 {
			return errors.New("invalid number of keys in auth body")
		}

		mapstructure.Decode(authObj, &auth)

		if len(auth.UserName) == 0 {
			return errors.New("username cannot be empty")
		}

		if len(auth.Password) == 0 {
			return errors.New("password cannot be empty")
		}
	case "session":
		var auth SessionAuth

		if len(authObj) != 5 {
			return errors.New("invalid number of keys in auth body")
		}

		mapstructure.Decode(authObj, &auth)

		if len(auth.UserName) == 0 {
			return errors.New("username cannot be empty")
		}

		if len(auth.Password) == 0 {
			return errors.New("password cannot be empty")
		}

		if len(auth.AuthBody) == 0 {
			return errors.New("authbody cannot be empty")
		}

		if len(auth.AuthBodyMap) == 0 {
			return errors.New("authbodymap cannot be empty")
		}

		if auth.CookieLifetime == 0 {
			return errors.New("invalid cookieLifetime value")
		}
	case "apikey":
		var auth APIKeyAuth

		if len(authObj) != 2 {
			return errors.New("invalid number of keys in auth body")
		}

		mapstructure.Decode(authObj, &auth)

		// if len(auth.Header) == 0 {
		// 	return errors.New("header cannot be empty")
		// }

		if len(auth.Key) == 0 {
			return errors.New("key cannot be empty")
		}
	case "rfc3447":
		var auth RFC3447Auth

		if len(authObj) != 2 {
			return errors.New("invalid number of keys in auth body")
		}

		mapstructure.Decode(authObj, &auth)

		if len(auth.PublicKey) == 0 {
			return errors.New("publickey cannot be empty")
		}

		if len(auth.PrivateKey) == 0 {
			return errors.New("privatekey cannot be empty")
		}
	default:
		return errors.New("invalid auth type")
	}

	return nil
}

// ValidAuthType Function
func ValidAuthType(auth string) bool {
	for _, authType := range AuthTypes {
		if auth == authType {
			return true
		}
	}
	return false
}
