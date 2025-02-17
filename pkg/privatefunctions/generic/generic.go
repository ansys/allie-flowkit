package generic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/ansys/allie-sharedtypes/pkg/logging"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// CreatePayloadAndSendHttpRequest creates a JSON payload from a request object and sends an HTTP POST request to the specified URL.
// The response body is decoded into the responsePtr object.
//
// Parameters:
//   - url: the URL to send the HTTP POST request to.
//   - requestType: the type of HTTP request to send (POST, GET, PUT, DELETE).
//   - requestObject: the request object to create the JSON payload from.
//   - responsePtr: a pointer to the response object to decode the JSON response body into.
//
// Returns:
//   - an error if there was an issue creating the JSON payload, sending the HTTP POST request, or decoding the JSON response body.
//   - the status code of the HTTP response.
func CreatePayloadAndSendHttpRequest(url string, requestType string, requestObject interface{}, responsePtr interface{}) (funcError error, statusCode int) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic in CreatePayloadAndSendHttpRequest: %v", r)
			funcError = r.(error)
			return
		}
	}()
	// Check if the request type is valid (POST, GET, PUT, DELETE)
	if requestType != "POST" && requestType != "GET" && requestType != "PUT" && requestType != "DELETE" {
		return fmt.Errorf("invalid request type: %s", requestType), 0
	}

	// Define the JSON payload.
	jsonPayload, err := json.Marshal(requestObject)
	if err != nil {
		return fmt.Errorf("error marshalling JSON payload: %v", err), 0
	}

	// Create a new HTTP POST request.
	req, err := http.NewRequest(requestType, url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err, 0
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	// Send the HTTP POST request.
	resp, err := client.Do(req)
	if err != nil {
		return err, resp.StatusCode
	}
	defer resp.Body.Close()

	// Decode the JSON response body into the 'data' struct.
	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()
	if err := decoder.Decode(responsePtr); err != nil {
		return err, 0
	}

	// Check the response status code.
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(resp.Status), resp.StatusCode
	}

	return nil, 0
}

// ExtractStringFieldFromStruct extracts a string field from a struct.
//
// Parameters:
//   - data: the struct to extract the string field from.
//   - fieldName: the name of the field to extract.
//
// Returns:
//   - the string field value.
//   - an error if the field is not found or is not a string.
func ExtractStringFieldFromStruct(data interface{}, fieldName string) (string, error) {
	v := reflect.ValueOf(data)

	// Dereference pointer if needed
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Ensure it's a struct or map[string]interface{}
	if v.Kind() == reflect.Struct {
		// Get field by name
		field := v.FieldByName(fieldName)
		if !field.IsValid() {
			return "", fmt.Errorf("field '%s' not found", fieldName)
		}

		// Ensure field is a string
		if field.Kind() != reflect.String {
			return "", fmt.Errorf("field '%s' is not a string", fieldName)
		}

		return field.String(), nil
	} else {
		// If it's a map extract the field
		field := v.MapIndex(reflect.ValueOf(fieldName))
		if !field.IsValid() {
			return "", fmt.Errorf("field '%s' not found", fieldName)
		}

		fieldValue := field.Interface()

		// Check if the field is of type string
		strVal, ok := fieldValue.(string)
		if !ok {
			return "", fmt.Errorf("field '%s' is not a string", fieldName)
		}

		return strVal, nil
	}
}

// SnakeToCamel converts a snake_case string to camelCase or PascalCase based on upperFirst flag
//
// Parameters:
//   - s: the snake_case string to convert.
//   - upperFirst: a flag to determine if the first letter should be capitalized.
//
// Returns:
//   - the camelCase or PascalCase string.
func SnakeToCamel(s string, upperFirst bool) string {
	parts := strings.Split(s, "_")
	if len(parts) == 0 {
		return s
	}

	// Use proper Unicode-aware title casing
	titleCaser := cases.Title(language.English)

	// Process the first part based on upperFirst flag
	var result string
	if upperFirst {
		result = titleCaser.String(parts[0]) // PascalCase: Capitalize first letter
	} else {
		result = strings.ToLower(parts[0]) // camelCase: Keep lowercase for first word
	}

	// Capitalize the first letter of subsequent parts
	for _, part := range parts[1:] {
		if len(part) > 0 {
			result += titleCaser.String(part) // Capitalize each word properly
		}
	}

	return result
}
