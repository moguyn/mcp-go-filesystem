package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func TestRegisterTools(t *testing.T) {
	s := server.NewMCPServer("test-server", "1.0.0")
	allowedDirs := []string{"/test/dir1", "/test/dir2"}

	// Test registration
	RegisterTools(s, allowedDirs)

	// Verify that all tools are registered
	expectedTools := []string{
		"read_file",
		"read_multiple_files",
		"write_file",
		"create_directory",
		"list_directory",
		"directory_tree",
		"move_file",
		"search_files",
		"get_file_info",
		"list_allowed_directories",
		"edit_file",
	}

	// Create a tools/list request
	request := mcp.JSONRPCRequest{
		JSONRPC: mcp.JSONRPC_VERSION,
		ID:      1,
		Request: mcp.Request{
			Method: "tools/list",
		},
	}

	// Marshal the request to JSON
	requestJSON, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Send the request to the server
	response := s.HandleMessage(context.Background(), requestJSON)
	if response == nil {
		t.Fatal("Expected response from HandleMessage but got nil")
	}

	// Check if the response is a JSONRPCResponse
	jsonResponse, ok := response.(mcp.JSONRPCResponse)
	if !ok {
		t.Fatal("Expected JSONRPCResponse but got something else")
	}

	// Check if the result is a ListToolsResult
	result, ok := jsonResponse.Result.(mcp.ListToolsResult)
	if !ok {
		t.Fatal("Expected ListToolsResult but got something else")
	}

	// Create a map of registered tools for easy lookup
	registeredTools := make(map[string]bool)
	for _, tool := range result.Tools {
		registeredTools[tool.Name] = true
	}

	// Check that all expected tools are registered
	for _, toolName := range expectedTools {
		if !registeredTools[toolName] {
			t.Errorf("Tool %q was not registered", toolName)
		}
	}
}
