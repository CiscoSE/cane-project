package auth

import (
	"cane-project/account"
	"cane-project/model"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// APIKeyAuth Function
func APIKeyAuth(api model.API, method string, queryParams url.Values, body string) (*http.Request, error) {
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

	targetMethod := strings.ToUpper(method)

	// Encode Query Params and append to resourcePath
	if len(queryParams) != 0 && targetMethod == "GET" {
		encodedQuery := queryParams.Encode()
		queryPath = "?" + strings.Replace(encodedQuery, "+", "%20", -1)
	}

	targetURL := host.String() + api.Path + queryPath

	// Create HTTP request
	req, err := http.NewRequest(targetMethod, targetURL, strings.NewReader(body))

	if err != nil {
		log.Print(err)
		fmt.Println("Errored when creating the HTTP request!")
	}

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

	// Append headers to HTTP request
	if apiHeader != "" {
		header := fmt.Sprintf("ADDING HEADER--> %s: %s", apiHeader, apiKey)
		fmt.Println(header)
		req.Header.Add(apiHeader, apiKey)
	} else {
		bearerToken := "Bearer " + apiKey
		header := fmt.Sprintf("ADDING HEADER--> Authorization: Bearer %s", apiKey)
		fmt.Println(header)
		req.Header.Add("Authorization", bearerToken)
	}

	// For testing, fix this later to support XML & JSON
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	return req, nil
}
