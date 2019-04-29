package auth

import (
	"cane-project/account"
	"cane-project/database"
	"cane-project/model"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// SessionAuth Function
func SessionAuth(api model.API, method string, queryParams url.Values) (*http.Request, error) {
	device, deviceErr := account.GetDeviceFromDB(api.DeviceAccount)
	var queryPath string

	if deviceErr != nil {
		log.Print(deviceErr)
		fmt.Println("Errored when creating the HTTP request!")
		return nil, deviceErr
	}

	host, err := url.Parse(device.BaseURL)
	if err != nil {
		panic("Cannot parse *host*!")
	}

	fmt.Println("SCHEME: ", host.Scheme)
	fmt.Println("HOSTNAME: ", host.Hostname())
	fmt.Println("ENDPOINT: ", api.Path)

	targetMethod := strings.ToUpper(method)

	fmt.Println("METHOD: ", targetMethod)

	// Encode Query Params and append to resourcePath
	if len(queryParams) != 0 && targetMethod == "GET" {
		encodedQuery := queryParams.Encode()
		queryPath = "?" + strings.Replace(encodedQuery, "+", "%20", -1)
	}

	targetURL := host.String() + api.Path + queryPath

	fmt.Println("TARGETURL: ", targetURL)

	// Create HTTP request
	req, err := http.NewRequest(targetMethod, targetURL, strings.NewReader(api.Body))

	if err != nil {
		log.Print(err)
		fmt.Println("Errored when creating the HTTP request!")
	}

	fmt.Println("REQ: ", req)

	// filter := primitive.M{
	// 	"_id": primitive.ObjectID(device.AuthObj),
	// }

	// foundVal, foundErr := database.FindOne("auth", device.AuthType, filter)

	// if foundErr != nil {
	// 	fmt.Println(foundErr)
	// 	return nil, foundErr
	// }

	// apiHeader := foundVal["header"].(string)
	// apiKey := foundVal["key"].(string)

	apiHeader := device.AuthObj["header"].(string)
	apiKey := device.AuthObj["key"].(string)

	fmt.Println("APIHEADER: ", apiHeader)
	fmt.Println("APIKEY: ", apiKey)

	// Append headers to HTTP request
	req.Header.Add(apiHeader, apiKey)

	return req, nil
}

// SessionTimer Function
func SessionTimer(accountName string, cookieTime time.Duration, cookieToken string) {
	session := map[string]interface{}{
		"deviceName":   accountName,
		"cookieExpire": time.Now().Add(cookieTime * time.Second),
		"token":        cookieToken,
	}

	saveID, _ := database.Save("session", "sessions", session)

	fmt.Println(saveID)
}
