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

	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

// SessionAuth Function
func SessionAuth(api model.API) (*http.Request, error) {
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
