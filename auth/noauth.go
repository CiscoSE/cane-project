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

// NoAuth Function
func NoAuth(api model.API, method string, queryParams url.Values, body string) (*http.Request, error) {
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
		return nil, err
	}

	fmt.Println("REQ: ", req)

	return req, nil
}
