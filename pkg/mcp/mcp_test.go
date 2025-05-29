// This test file validates the MCP functions, only for testing purposes
// It requires a dummy WebSocket MCP server running at ws://localhost:9090

package externalfunctions

import (
	"encoding/json"
	"fmt"
	"testing"

	"golang.org/x/net/context"
	"nhooyr.io/websocket/wsjson"
	"nhooyr.io/websocket"
)

// Helper for connecting to dummy server and sending a payload
func sendToDummyServer(t *testing.T, payload map[string]interface{}) map[string]interface{} {
	ctx := context.Background()
	c, _, err := websocket.Dial(ctx, "ws://localhost:9090", nil)
	if err != nil {
		t.Fatalf("WebSocket dial error: %v", err)
	}
	defer c.Close(websocket.StatusNormalClosure, "closing")

	if err := wsjson.Write(ctx, c, payload); err != nil {
		t.Fatalf("Failed to write JSON to WebSocket: %v", err)
	}

	var result map[string]interface{}
	if err := wsjson.Read(ctx, c, &result); err != nil {
		t.Fatalf("Failed to read JSON from WebSocket: %v", err)
	}

	return result
}

func TestListMCPItems(t *testing.T) {
	result, err := ListMCPItems()
	if err != nil {
		t.Fatalf("ListMCPItems failed: %v", err)
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println("ListMCPItems:", string(out))
}

func TestExecuteMCPTool(t *testing.T) {
	result, err := ExecuteMCPTool("AssignStringToString", map[string]interface{}{
		"arg1": "TestInput",
	})
	if err != nil {
		t.Fatalf("ExecuteMCPTool failed: %v", err)
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println("ExecuteMCPTool:", string(out))
}

func TestGetMCPResource(t *testing.T) {
	result, err := GetMCPResource("DefaultConfig")
	if err != nil {
		t.Fatalf("GetMCPResource failed: %v", err)
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println("GetMCPResource:", string(out))
}

func TestGetMCPPrompt(t *testing.T) {
	result, err := GetMCPPrompt("SystemPrompt")
	if err != nil {
		t.Fatalf("GetMCPPrompt failed: %v", err)
	}
	fmt.Println("GetMCPPrompt:", result)
}
