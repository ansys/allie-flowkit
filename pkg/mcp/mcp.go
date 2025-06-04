// mcp.go
// Package mcp provides an interface for connecting to WebSocket-based MCP servers.
package mcp

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

// ConnectToMCP connects to the WebSocket MCP server at the specified URL
// Returns a pointer to the WebSocket connection
func ConnectToMCP(serverURL string) (*websocket.Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to WebSocket server: %w", err)
	}
	return conn, nil
}

// sendRequest sends a JSON request to the provided WebSocket connection and waits for a response
// It returns the parsed response as a map
func sendRequest(conn *websocket.Conn, request interface{}) (map[string]interface{}, error) {
	// Marshal the request to JSON
	requestData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send the JSON request
	if err := conn.WriteMessage(websocket.TextMessage, requestData); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Wait for the response
	_, responseData, err := conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Unmarshal the response
	var response map[string]interface{}
	if err := json.Unmarshal(responseData, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return response, nil
}

// ListAll connects to MCP and retrieves a list of tools/resources/prompts
func ListAll(serverURL string) (map[string][]string, error) {
	conn, err := ConnectToMCP(serverURL)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	request := map[string]interface{}{
		"intent": "list",
	}

	response, err := sendRequest(conn, request)
	if err != nil {
		return nil, err
	}

	// Convert map[string]interface{} to map[string][]string
	result := make(map[string][]string)
	for k, v := range response {
		// Try to convert each value to []string
		if arr, ok := v.([]interface{}); ok {
			strArr := make([]string, len(arr))
			for i, item := range arr {
				if s, ok := item.(string); ok {
					strArr[i] = s
				} else {
					return nil, fmt.Errorf("expected string in array for key %s", k)
				}
			}
			result[k] = strArr
		} else {
			return nil, fmt.Errorf("expected array for key %s", k)
		}
	}

	return result, nil
}

// ExecuteTool connects to MCP and executes a tool with the given name and arguments
func ExecuteTool(serverURL, toolName string, args map[string]interface{}) (map[string]interface{}, error) {
	conn, err := ConnectToMCP(serverURL)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	request := map[string]interface{}{
		"intent": "execute",
		"tool":   toolName,
		"args":   args,
	}

	response, err := sendRequest(conn, request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// GetResource connects to MCP and retrieves a resource by name
func GetResource(serverURL, resourceName string) (map[string]interface{}, error) {
	conn, err := ConnectToMCP(serverURL)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	request := map[string]interface{}{
		"intent": "get_resource",
		"name":   resourceName,
	}

	response, err := sendRequest(conn, request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// GetSystemPrompt connects to MCP and retrieves a prompt by name
func GetSystemPrompt(serverURL, promptName string) (string, error) {
	conn, err := ConnectToMCP(serverURL)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	request := map[string]interface{}{
		"intent": "get_prompt",
		"name":   promptName,
	}

	response, err := sendRequest(conn, request)
	if err != nil {
		return "", err
	}

	// Assuming the response contains a field called "prompt"
	prompt, exists := response["prompt"]
	if !exists {
		return "", fmt.Errorf("prompt not found in response")
	}

	promptStr, ok := prompt.(string)
	if !ok {
		return "", fmt.Errorf("prompt is not a string")
	}

	return promptStr, nil
}
