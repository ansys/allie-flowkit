// These functions connect to a locally running mock MCP server via WebSocket (ws://localhost:9090)
// and simulate the core behavior of MCP components: tool listing, execution, resource access, and prompt retrieval

package externalfunctions

import (
	"fmt"

	"github.com/gorilla/websocket"
)

// ListMCPItems calls the dummy MCP server to get available tools, resources, and prompts
func ListMCPItems() (map[string][]string, error) {
	c, _, err := websocket.DefaultDialer.Dial("ws://localhost:9090", nil)
	if err != nil {
		return nil, fmt.Errorf("WebSocket dial failed: %w", err)
	}
	defer c.Close()

	request := map[string]interface{}{
		"intent": "list",
	}

	if err := c.WriteJSON(request); err != nil {
		return nil, fmt.Errorf("failed to send list request: %w", err)
	}

	var response map[string][]string
	if err := c.ReadJSON(&response); err != nil {
		return nil, fmt.Errorf("failed to decode MCP items: %w", err)
	}

	return response, nil
}

// ExecuteMCPTool simulates executing a tool with arguments via the dummy MCP server
func ExecuteMCPTool(toolName string, args map[string]interface{}) (map[string]interface{}, error) {
	c, _, err := websocket.DefaultDialer.Dial("ws://localhost:9090", nil)
	if err != nil {
		return nil, fmt.Errorf("WebSocket dial failed: %w", err)
	}
	defer c.Close()

	request := map[string]interface{}{
		"intent": "execute",
		"tool":   toolName,
		"args":   args,
	}

	if err := c.WriteJSON(request); err != nil {
		return nil, fmt.Errorf("failed to send execute request: %w", err)
	}

	var response map[string]interface{}
	if err := c.ReadJSON(&response); err != nil {
		return nil, fmt.Errorf("failed to decode execute response: %w", err)
	}

	return response, nil
}

// GetMCPResource fetches mock data for a resource from the dummy MCP server
func GetMCPResource(name string) (map[string]interface{}, error) {
	c, _, err := websocket.DefaultDialer.Dial("ws://localhost:9090", nil)
	if err != nil {
		return nil, fmt.Errorf("WebSocket dial failed: %w", err)
	}
	defer c.Close()

	request := map[string]interface{}{
		"intent": "get_resource",
		"name":   name,
	}

	if err := c.WriteJSON(request); err != nil {
		return nil, fmt.Errorf("failed to send get_resource request: %w", err)
	}

	var response map[string]interface{}
	if err := c.ReadJSON(&response); err != nil {
		return nil, fmt.Errorf("failed to decode resource %q: %w", name, err)
	}

	return response, nil
}

// GetMCPPrompt fetches a prompt string from the dummy MCP server
func GetMCPPrompt(name string) (string, error) {
	c, _, err := websocket.DefaultDialer.Dial("ws://localhost:9090", nil)
	if err != nil {
		return "", fmt.Errorf("WebSocket dial failed: %w", err)
	}
	defer c.Close()

	request := map[string]interface{}{
		"intent": "get_prompt",
		"name":   name,
	}

	if err := c.WriteJSON(request); err != nil {
		return "", fmt.Errorf("failed to send get_prompt request: %w", err)
	}

	var response struct {
		Prompt string `json:"prompt"`
	}
	if err := c.ReadJSON(&response); err != nil {
		return "", fmt.Errorf("failed to decode prompt response: %w", err)
	}

	return response.Prompt, nil
}
