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

	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

// APIKeyAuth Function
func APIKeyAuth(api model.API) (*http.Request, error) {
	device, deviceErr := account.GetDeviceFromDB(api.DeviceAccount)

	if deviceErr != nil {
		log.Print(deviceErr)
		fmt.Println("Errored when creating the HTTP request!")
		return nil, deviceErr
	}

	host, err := url.Parse(device.URL)
	if err != nil {
		panic("Cannot parse *host*!")
	}

	fmt.Println("SCHEME: ", host.Scheme)
	fmt.Println("HOSTNAME: ", host.Hostname())
	fmt.Println("ENDPOINT: ", api.URL)

	targetMethod := strings.ToUpper(api.Method)

	fmt.Println("METHOD: ", targetMethod)

	targetURL := host.String() + api.URL

	fmt.Println("TARGETURL: ", targetURL)

	// Create HTTP request
	req, err := http.NewRequest(targetMethod, targetURL, strings.NewReader(api.Body))

	if err != nil {
		log.Print(err)
		fmt.Println("Errored when creating the HTTP request!")
	}

	fmt.Println("REQ: ", req)

	filter := primitive.M{
		"_id": primitive.ObjectID(device.AuthObj),
	}

	foundVal, foundErr := database.FindOne("auth", device.AuthType, filter)

	if foundErr != nil {
		fmt.Println(foundErr)
		return nil, foundErr
	}

	apiHeader := foundVal["header"].(string)
	apiKey := foundVal["key"].(string)

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

	fmt.Println("FORMED HEADER: ", req.Header)

	// For testing, fix this later to support XML & JSON
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	return req, nil
}
