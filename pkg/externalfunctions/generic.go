// Copyright (C) 2025 ANSYS, Inc. and/or its affiliates.
// SPDX-License-Identifier: MIT
//
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package externalfunctions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/ansys/aali-sharedtypes/pkg/sharedtypes"
)

// SendAPICall sends an API call to the specified URL with the specified headers and query parameters.
//
// Tags:
//   - @displayName: REST Call
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
// Tags:
//   - @displayName: Assign String to String
//
// Parameters:
//   - inputString: the input string
//
// Returns:
//   - outputString: the output string
func AssignStringToString(inputString string) (outputString string) {
	return inputString
}

// PrintFeedback prints the feedback to the console in JSON format
//
// Tags:
//   - @displayName: Print Feedback
//
// Parameters:
//   - feedback: the feedback to print
func PrintFeedback(feedback sharedtypes.Feedback) {
	// create json string from feedback struct
	jsonString, err := json.Marshal(feedback)
	if err != nil {
		panic(fmt.Sprintf("Error marshalling feedback to JSON: %v", err))
	}
	// print json string to console
	fmt.Println(string(jsonString))
}
