package generic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ansys/allie-sharedtypes/pkg/logging"
)

func CreatePayloadAndSendHttpRequest(url string, requestObject interface{}, responsePtr interface{}) (funcError error, statusCode int) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic in CreatePayloadAndSendHttpRequest: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Define the JSON payload.
	jsonPayload, err := json.Marshal(requestObject)
	if err != nil {
		return fmt.Errorf("error marshalling JSON payload: %v", err), 0
	}

	// Create a new HTTP POST request.
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
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
	if err := decoder.Decode(responsePtr); err != nil {
		return err, 0
	}

	// Check the response status code.
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(resp.Status), resp.StatusCode
	}

	return nil, 0
}
