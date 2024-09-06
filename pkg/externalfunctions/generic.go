package externalfunctions

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// SendAPICall sends an API call to the specified URL with the specified headers and query parameters.
//
// Parameters:
//   - requestType: the type of the request (GET, POST, PUT, PATCH, DELETE)
//   - urlString: the URL to send the request to
//   - headers: the headers to include in the request
//   - query: the query parameters to include in the request
//   - jsonBody: the body of the request as a JSON string
//
// Returns:
//   - success: a boolean indicating whether the request was successful
//   - returnJsonBody: the JSON body of the response as a string
func SendRestAPICall(requestType string, endpoint string, header map[string]string, query map[string]string, jsonBody string) (success bool, returnJsonBody string) {
	// verify correct request type
	if requestType != "GET" && requestType != "POST" && requestType != "PUT" && requestType != "PATCH" && requestType != "DELETE" {
		panic(fmt.Sprintf("Invalid request type: %v", requestType))
	}

	// Parse the URL and add query parameters
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		panic(fmt.Sprintf("Error parsing URL: %v", err))
	}

	q := parsedURL.Query()
	for key, value := range query {
		q.Add(key, value)
	}
	parsedURL.RawQuery = q.Encode()

	// Create the HTTP request
	var req *http.Request
	if jsonBody != "" {
		req, err = http.NewRequest(requestType, parsedURL.String(), bytes.NewBuffer([]byte(jsonBody)))
	} else {
		req, err = http.NewRequest(requestType, parsedURL.String(), nil)
	}
	if err != nil {
		panic(fmt.Sprintf("Error creating request: %v", err))
	}

	// Add headers
	for key, value := range header {
		req.Header.Add(key, value)
	}

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(fmt.Sprintf("Error executing request: %v", err))
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("Error reading response body: %v", err))
	}

	// Check if the response code is successful (2xx)
	success = resp.StatusCode >= 200 && resp.StatusCode < 300

	return success, string(body)
}

// AssignStringToString assigns a string to another string
//
// Parameters:
//   - inputString: the input string
//
// Returns:
//   - outputString: the output string
func AssignStringToString(inputString string) (outputString string) {
	return inputString
}
