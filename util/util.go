package util

import (
	"bytes"
	"cane-project/model"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	//"github.com/mongodb/mongo-go-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ProxyURL Variable
var ProxyURL string

// IgnoreSSL Global Setting
var IgnoreSSL bool

// DebugEnable Global Setting
var DebugEnable bool

// XMLNode Struct
type XMLNode struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:"-"`
	Content []byte     `xml:",innerxml"`
	Nodes   []XMLNode  `xml:",any"`
}

// JSONNode Struct
type JSONNode struct {
	Node map[string]interface{}
}

func init() {
	IgnoreSSL = true

	httpProxy := os.Getenv("CANE_PROXY")

	if len(httpProxy) > 0 {
		fmt.Println("Setting Proxy:", httpProxy)
		ProxyURL = httpProxy
	}
}

// EncodeBase64 Function
func EncodeBase64(encode string) string {
	return base64.StdEncoding.EncodeToString([]byte(encode))
}

// RespondwithJSON write json response format
func RespondwithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// RespondwithXML write xml response format
func RespondwithXML(w http.ResponseWriter, code int, payload interface{}) {
	response := []byte(payload.(model.XMLNode).Marshal())
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(code)
	w.Write(response)
}

// RespondwithString write string response format
func RespondwithString(w http.ResponseWriter, code int, payload interface{}) {
	response := []byte(payload.(string))
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(code)
	w.Write(response)
}

// RespondWithError return error message
func RespondWithError(w http.ResponseWriter, code int, msg string) {
	RespondwithJSON(w, code, map[string]string{"message": msg})
}

// JSONPrettyPrint Function
// Formats JSON into a much more easily readable format
// by properly indenting with two spaces where needed
// and returns it as a string
func JSONPrettyPrint(input string) string {
	var output bytes.Buffer
	err := json.Indent(&output, []byte(input), "", "  ")
	if err != nil {
		return input
	}
	return output.String()
}

// UnmarshalJSON Function
func UnmarshalJSON(data []byte, target interface{}) error {
	err := json.Unmarshal(data, &target)

	if err != nil {
		return err
	}

	return nil
}

// StringInSlice Function
func StringInSlice(a []string, b string) bool {
	for _, i := range a {
		if b == i {
			return true
		}
	}
	return false
}

// StructToMap Function
func StructToMap(iface interface{}) map[string]interface{} {
	newMap := map[string]interface{}{}

	iVal := reflect.ValueOf(iface).Elem()
	typ := iVal.Type()

	for i := 0; i < iVal.NumField(); i++ {
		f := iVal.Field(i)
		var v string
		switch f.Interface().(type) {
		case int, int8, int16, int32, int64:
			v = strconv.FormatInt(f.Int(), 10)
		case uint, uint8, uint16, uint32, uint64:
			v = strconv.FormatUint(f.Uint(), 10)
		case float32:
			v = strconv.FormatFloat(f.Float(), 'f', 4, 32)
		case float64:
			v = strconv.FormatFloat(f.Float(), 'f', 4, 64)
		case []byte:
			v = string(f.Bytes())
		case string:
			v = f.String()
		case primitive.ObjectID:
			v = f.Interface().(primitive.ObjectID).Hex()
		}

		newMap[typ.Field(i).Name] = v
	}

	return newMap
}

// IsVar Function
func IsVar(data string) bool {
	if strings.HasPrefix(data, "{{") && strings.HasSuffix(data, "}}") {
		return true
	}

	return false
}

// GetVariables Function
func GetVariables(data string) []string {
	regex := regexp.MustCompile(`{{[a-zA-Z]+}}`)
	matches := regex.FindAllString(data, -1)

	return matches
}
