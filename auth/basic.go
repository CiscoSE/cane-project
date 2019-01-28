package auth

import (
	"cane-project/account"
	"cane-project/database"
	"cane-project/model"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

// BasicAuth Function
func BasicAuth(api model.API) (*http.Request, error) {
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

	targetMethod := strings.ToUpper(api.Method)

	targetURL := host.String() + api.URL

	// Create HTTP request
	req, err := http.NewRequest(targetMethod, targetURL, strings.NewReader(api.Body))

	if err != nil {
		log.Print(err)
		fmt.Println("Errored when creating the HTTP request!")
	}

	filter := primitive.M{
		"_id": primitive.ObjectID(device.AuthObj),
	}

	foundVal, foundErr := database.FindOne("auth", device.AuthType, filter)

	if foundErr != nil {
		fmt.Println(foundErr)
		return nil, foundErr
	}

	userPass := []byte(foundVal["username"].(string) + ":" + foundVal["password"].(string))
	authKey := "Basic " + base64.StdEncoding.EncodeToString(userPass)

	// Append headers to HTTP request
	req.Header.Add("Authorization", authKey)

	return req, nil
}
