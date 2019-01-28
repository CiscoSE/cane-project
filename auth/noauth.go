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
func NoAuth(api model.API) (*http.Request, error) {
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
		return nil, err
	}

	fmt.Println("REQ: ", req)

	return req, nil
}
