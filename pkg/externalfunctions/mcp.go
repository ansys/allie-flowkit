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
	"context"
	"fmt"

	"nhooyr.io/websocket"
)

// ListAll retrieves all tools, resources, and prompts from the MCP server.
//
// Tags:
//   - @displayName: List MCP Items
//
// Parameters:
//   - serverURL: the WebSocket URL of the MCP server
//
// Returns:
//   - result: a map with lists of tool/resource/prompt names categorized by type
//   - error: any error that occurred during the process
func ListAll(serverURL string) (map[string][]string, error) {
	ctx := context.Background()

	conn, err := connectToMCP(ctx, serverURL)
	if err != nil {
		return nil, err
	}
	defer conn.Close(websocket.StatusNormalClosure, "closing after list request")

	request := map[string]interface{}{"intent": "list"}
	response, err := sendMCPRequest(ctx, conn, request)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]string)
	for k, v := range response {
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

// ExecuteTool executes a specific tool via the MCP server with provided arguments.
//
// Tags:
//   - @displayName: Execute MCP Tool
//
// Parameters:
//   - serverURL: the WebSocket URL of the MCP server
//   - toolName: the name of the tool to execute
//   - args: a map of arguments to pass to the tool
//
// Returns:
//   - result: the response from the tool execution
//   - error: any error that occurred during execution
func ExecuteTool(serverURL, toolName string, args map[string]interface{}) (map[string]interface{}, error) {
	ctx := context.Background()

	conn, err := connectToMCP(ctx, serverURL)
	if err != nil {
		return nil, err
	}
	defer conn.Close(websocket.StatusNormalClosure, "closing after execute request")

	request := map[string]interface{}{
		"intent": "execute",
		"tool":   toolName,
		"args":   args,
	}
	return sendMCPRequest(ctx, conn, request)
}

// GetResource retrieves a named resource from the MCP server.
//
// Tags:
//   - @displayName: Get MCP Resource
//
// Parameters:
//   - serverURL: the WebSocket URL of the MCP server
//   - resourceName: the name of the resource to retrieve
//
// Returns:
//   - result: the retrieved resource as a map
//   - error: any error that occurred during the request
func GetResource(serverURL, resourceName string) (map[string]interface{}, error) {
	ctx := context.Background()

	conn, err := connectToMCP(ctx, serverURL)
	if err != nil {
		return nil, err
	}
	defer conn.Close(websocket.StatusNormalClosure, "closing after resource fetch")

	request := map[string]interface{}{
		"intent": "get_resource",
		"name":   resourceName,
	}
	return sendMCPRequest(ctx, conn, request)
}

// GetSystemPrompt retrieves a system prompt by name from the MCP server.
//
// Tags:
//   - @displayName: Get MCP Prompt
//
// Parameters:
//   - serverURL: the WebSocket URL of the MCP server
//   - promptName: the name of the system prompt to retrieve
//
// Returns:
//   - promptStr: the text of the retrieved prompt
//   - error: any error that occurred during the request
func GetSystemPrompt(serverURL, promptName string) (string, error) {
	ctx := context.Background()

	conn, err := connectToMCP(ctx, serverURL)
	if err != nil {
		return "", err
	}
	defer conn.Close(websocket.StatusNormalClosure, "closing after prompt fetch")

	request := map[string]interface{}{
		"intent": "get_prompt",
		"name":   promptName,
	}
	response, err := sendMCPRequest(ctx, conn, request)
	if err != nil {
		return "", err
	}

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
