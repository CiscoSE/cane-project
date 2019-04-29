package model

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
)

// JSONNode Type
type JSONNode map[string]interface{}

// XMLAttrs Type
type XMLAttrs map[string][]map[string]interface{}

// IsJSON Function
func IsJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

// JSONfromBytes Function
func JSONfromBytes(input []byte) (JSONNode, error) {
	var j JSONNode

	jsonErr := json.Unmarshal(input, &j)

	if jsonErr != nil {
		return j, jsonErr
	}

	return j, nil
}

// JSONVars Function for JSONNode
func (j JSONNode) JSONVars() {
	for key, val := range j {
		switch valType := reflect.ValueOf(val).Kind(); valType {
		case reflect.Map:
			tempKey := JSONNode(j[key].(map[string]interface{}))
			tempKey.JSONVars()
			j[key] = tempKey
		default:
			j[key] = "{{var_" + key + "}}"
		}
	}
}

// Marshal Function for JSONNode
func (j JSONNode) Marshal(args ...int) string {
	prefix := ""
	indent := "    "

	if len(args) == 1 {
		prefix = strings.Repeat(" ", args[0])
	} else if len(args) == 2 {
		prefix = strings.Repeat(" ", args[0])
		indent = strings.Repeat(" ", args[1])
	}

	jsonBytes, jsonErr := json.MarshalIndent(j, prefix, indent)
	jsonString := string(jsonBytes)

	if jsonErr != nil {
		panic(jsonErr)
	}

	return jsonString
}

// StripJSON Function
func (j JSONNode) StripJSON() interface{} {
	// Wrap the original in a reflect.Value
	original := reflect.ValueOf(j)

	copy := reflect.New(original.Type()).Elem()
	StripJSONRecursive(copy, original)

	// Remove the reflection wrapper
	return copy.Interface()
}

// ToXML Function
func (j JSONNode) ToXML() string {
	var newNode JSONNode
	var returnString string

	for k, v := range j {
		switch valType := reflect.TypeOf(v).Kind(); valType {
		case reflect.Map:
			mapstructure.Decode(v, &newNode)

			if k == "cdata" {
				returnString += "<![CDATA["
				returnString += newNode.ToXML()
				returnString += "]]>"
			} else {
				if attrs, ok := v.(map[string]interface{})["attr"]; ok {
					returnString += "<" + k

					for attrKey, attrVal := range attrs.(map[string]interface{}) {
						returnString += " " + attrKey + "=\""
						returnString += attrVal.(string) + "\""
					}

					if data, ok := v.(map[string]interface{})["data"]; ok {
						returnString += ">"
						returnString += data.(string)
						returnString += "</" + k + ">"
					} else {
						returnString += " />"
					}
				} else {
					returnString += "<" + k + ">"
					returnString += newNode.ToXML()
					returnString += "</" + k + ">"
				}
			}
		case reflect.String:
			if k == "data" {
				returnString += v.(string)
			} else {
				returnString += "<" + k + ">"
				returnString += v.(string)
				returnString += "</" + k + ">"
			}

			return returnString
		default:
			fmt.Println("Cannot match type for:", v)
		}
	}

	return returnString
}

// StripJSONRecursive Function
func StripJSONRecursive(copy, original reflect.Value) {
	switch original.Kind() {
	// The first cases handle nested structures and translate them recursively

	// If it is a pointer we need to unwrap and call once again
	case reflect.Ptr:
		// To get the actual value of the original we have to call Elem()
		// At the same time this unwraps the pointer so we don't end up in
		// an infinite recursion
		originalValue := original.Elem()
		// Check if the pointer is nil
		if !originalValue.IsValid() {
			return
		}
		// Allocate a new object and set the pointer to it
		copy.Set(reflect.New(originalValue.Type()))
		// Unwrap the newly created pointer
		StripJSONRecursive(copy.Elem(), originalValue)

	// If it is an interface (which is very similar to a pointer), do basically the
	// same as for the pointer. Though a pointer is not the same as an interface so
	// note that we have to call Elem() after creating a new object because otherwise
	// we would end up with an actual pointer
	case reflect.Interface:
		// Get rid of the wrapping interface
		originalValue := original.Elem()
		// Create a new object. Now new gives us a pointer, but we want the value it
		// points to, so we have to call Elem() to unwrap it
		copyValue := reflect.New(originalValue.Type()).Elem()
		StripJSONRecursive(copyValue, originalValue)
		copy.Set(copyValue)

	// If it is a struct we translate each field
	case reflect.Struct:
		for i := 0; i < original.NumField(); i++ {
			StripJSONRecursive(copy.Field(i), original.Field(i))
		}

	// If it is a slice we create a new slice and translate each element
	case reflect.Slice:
		// newSlice := reflect.MakeSlice(original.Type(), original.Len(), original.Cap())
		switch sliceKind := original.Index(0).Kind(); sliceKind {
		case reflect.String:
			fmt.Println("Slice String")
			// copy.Set(reflect.MakeSlice(original.Type(), 0, original.Cap()))
			copy.Set(reflect.MakeSlice(original.Type(), 0, 0))
		case reflect.Int:
			// copy.Set(reflect.MakeSlice(original.Type(), 0, original.Cap()))
			copy.Set(reflect.MakeSlice(original.Type(), 0, 0))
		default:
			copy.Set(reflect.MakeSlice(original.Type(), original.Len(), original.Cap()))
			for i := 0; i < original.Len(); i++ {
				StripJSONRecursive(copy.Index(i), original.Index(i))
			}
		}

	// If it is a map we create a new map and translate each value
	case reflect.Map:
		copy.Set(reflect.MakeMap(original.Type()))
		for _, key := range original.MapKeys() {
			originalValue := original.MapIndex(key)
			// New gives us a pointer, but again we want the value
			copyValue := reflect.New(originalValue.Type()).Elem()
			StripJSONRecursive(copyValue, originalValue)
			copy.SetMapIndex(key, copyValue)
		}

	// Otherwise we cannot traverse anywhere so this finishes the the recursion

	// If it is a string translate it (yay finally we're doing what we came for)
	case reflect.String:
		copy.SetString("")

	// And everything else will simply be taken from the original
	default:
		copy.Set(original)
	}

}
