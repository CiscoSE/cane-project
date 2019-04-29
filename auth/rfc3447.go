package auth

import (
	"cane-project/account"
	"cane-project/model"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

// var host *url.URL

// var privateKey *rsa.PrivateKey
// var publicKey string

// IntersightResponse Type
type IntersightResponse map[string][]map[string]interface{}

var digestAlgorithm = "rsa-sha256"

var httpVerbs = map[string]bool{
	"GET":    true,
	"POST":   true,
	"PATCH":  true,
	"DELETE": true,
}

// Loads the RSA private key from a file and assigns it to
// the privateKey variable
func loadPrivateKey(keyData string) *rsa.PrivateKey {
	block, _ := pem.Decode([]byte(keyData))
	key, _ := x509.ParsePKCS1PrivateKey(block.Bytes)

	return key
}

// Loads the RSA private key from a file and assigns it to
// the privateKey variable
// func getPrivateKey() {
// 	b, err := ioutil.ReadFile("keys/private_key.pem")
// 	if err != nil {
// 		fmt.Print(err)
// 	}

// 	block, _ := pem.Decode(b)
// 	key, _ := x509.ParsePKCS1PrivateKey(block.Bytes)

// 	privateKey = key
// }

// Loads the Intersight KeyID key from a file and assigns it
// to the publicKey variable
// func getPublicKey() {
// 	b, err := ioutil.ReadFile("keys/public_key.txt")
// 	if err != nil {
// 		fmt.Print(err)
// 	}

// 	publicKey = string(b)
// }

// Generates a SHA256 digest from a string
func getSHA256Digest(data string) []byte {
	digest := sha256.Sum256([]byte(data))

	return digest[:]
}

// Prepares an Intersight header string by combining a map of
// the headers and the request target
func prepStringToSign(reqTarget string, hdrs map[string]interface{}) string {
	var keys []string

	ss := "(request-target): " + strings.ToLower(reqTarget) + "\n"

	length := len(hdrs)

	for k := range hdrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	count := 0
	for _, key := range keys {
		ss = ss + strings.ToLower(key) + ": " + hdrs[key].(string)
		if count < length-1 {
			ss = ss + "\n"
		}
		count++
	}

	return ss
}

// Generates a base64 encoded RSA SHA256 signed string y using
// the assigned privateKey to sign the request
func getRSASigSHA256b64Encode(data string, privateKey *rsa.PrivateKey) string {
	digest := getSHA256Digest(data)

	signature, err := rsa.SignPKCS1v15(nil, privateKey, crypto.SHA256, digest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from signing: %s\n", err)
		return ""
	}

	return base64.StdEncoding.EncodeToString(signature)
}

// Prepares the Intersight authorization string by properly
// formatting each of the required components and returning
// it as a string
func getAuthHeader(hdrs map[string]interface{}, signedMsg string, publicKey string) string {
	var keys []string

	authStr := "Signature"

	authStr = authStr + " " + "keyId=\"" + strings.TrimRight(publicKey, "\n") + "\"," + "algorithm=\"" + digestAlgorithm + "\"," + "headers=\"(request-target)"

	for k := range hdrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		authStr = authStr + " " + strings.ToLower(key)
	}

	authStr = authStr + "\""

	authStr = authStr + "," + "signature=\"" + signedMsg + "\""

	return authStr
}

// Returns the current time in GMT format
func getGMTDate() string {
	t := time.Now().UTC()

	return t.Format(http.TimeFormat)
}

// RFC3447Auth Function
// Invokes the Intersight API. Given an HTTP method,
// target resource path, and query parameters for filtering,
// a body (if necessary), MOID, name (if MOID is not known),
// and an HTTP proxy (if required), will return an HTTP
// *Response with the outcome of the API call
// func RFC3447Auth(httpMethod string, resourcePath string, queryParams url.Values, body map[string]interface{}, moid string, name string) *http.Request {
func RFC3447Auth(api model.API, method string, queryParams url.Values, body string) (*http.Request, error) {
	var bodyString string
	var queryPath string

	resourcePath := api.Path

	device, deviceErr := account.GetDeviceFromDB(api.DeviceAccount)

	if deviceErr != nil {
		log.Print(deviceErr)
		fmt.Println("Errored when creating the HTTP request!")
		return nil, deviceErr
	}

	host, err := url.Parse(device.BaseURL)
	if err != nil {
		panic("Cannot parse *host*!")
	}

	targetHost := host.Hostname()
	targetPath := host.RequestURI()
	// method := strings.ToUpper(httpMethod)
	targetMethod := strings.ToUpper(method)

	// Verify an accepted HTTP verb was chosen
	if !httpVerbs[targetMethod] {
		fmt.Println("Please select a valid HTTP verb (GET/POST/PATCH/DELETE)")
	}

	// Verify the resource path isn't empy & is a valid String type
	if len(api.Path) == 0 {
		fmt.Println("The *resourcePath* value is required")
	}

	// Verify the MOID is of proper length if it is set
	// if len(moid) != 0 && len([]byte(moid)) != 24 {
	// 	fmt.Println("Invalid *moid* value!")
	// }

	// Encode Query Params and append to resourcePath
	if len(queryParams) != 0 && targetMethod == "GET" {
		encodedQuery := queryParams.Encode()
		queryPath = strings.Replace(encodedQuery, "+", "%20", -1)
		resourcePath += "?" + queryPath
	}

	// Convert BODY to a String
	if targetMethod == "POST" || targetMethod == "PATCH" {
		if len(api.Body) > 0 {
			bodyBytes, _ := json.Marshal(api.Body)
			bodyString = string(bodyBytes)
		} else {
			fmt.Println("The *body* value must be set with \"POST/PATCH!\"")
			panic("The *body* value must be set with \"POST/PATCH!\"")
		}
	}

	// Find MOID by NAME if MOID is not set
	// if method == "PATCH" || method == "DELETE" {
	// 	if len(moid) == 0 {
	// 		fmt.Println("Must set either *moid* with \"PATCH/DELETE!\"")
	// 	}
	// }

	// Append MOID to URL
	// if len(moid) != 0 && method != "POST" {
	// 	resourcePath += "/" + moid
	// }

	// Concatenate URLs for headers
	targetURL := host.String() + resourcePath
	requestTarget := targetMethod + " " + targetPath + resourcePath

	// Get the date in GMT format
	currDate := getGMTDate()

	// Load public/private keys
	privateKey := loadPrivateKey(device.AuthObj["privateKey"].(string))
	publicKey := device.AuthObj["publicKey"].(string)

	// Generate Body Digest
	bodyDigest := getSHA256Digest(bodyString)
	b64BodyDigest := base64.StdEncoding.EncodeToString(bodyDigest)

	// Generate the authorization header
	authHeader := map[string]interface{}{
		"Date":   currDate,
		"Host":   targetHost,
		"Digest": "SHA-256=" + b64BodyDigest,
	}

	// Generate authorization string
	stringToSign := prepStringToSign(requestTarget, authHeader)
	b64SignedMsg := getRSASigSHA256b64Encode(stringToSign, privateKey)
	headerAuth := getAuthHeader(authHeader, b64SignedMsg, publicKey)

	// Generate the HTTP requests header
	requestHeader := map[string]string{
		"Accept":        "application/json",
		"Host":          targetHost,
		"Date":          currDate,
		"Digest":        "SHA-256=" + b64BodyDigest,
		"Authorization": headerAuth,
	}

	// Create HTTP request
	req, err := http.NewRequest(targetMethod, targetURL, strings.NewReader(bodyString))

	if err != nil {
		log.Print(err)
		fmt.Println("Errored when creating the HTTP request!")
		return nil, nil
	}

	// Append headers to HTTP request
	for key, value := range requestHeader {
		req.Header.Add(key, value)
	}

	// Add query params and call HTTP request
	req.URL.RawQuery = queryPath

	fmt.Println("RFC3447 REQUEST:")
	fmt.Println(req)

	return req, nil
}
