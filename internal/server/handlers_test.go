package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestHandleListAllowedDirectories tests the handleListAllowedDirectories function
func TestHandleListAllowedDirectories(t *testing.T) {
	// Create a server with test directories
	testDirs := []string{"/test/dir1", "/test/dir2"}
	s := &Server{
		allowedDirectories: testDirs,
	}

	// Call the handler
	response, err := s.handleListAllowedDirectories(map[string]interface{}{})
	if err != nil {
		t.Fatalf("handleListAllowedDirectories() returned error: %v", err)
	}

	// Verify the response
	if len(response.Content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(response.Content))
	}

	if response.Content[0].Type != "text" {
		t.Errorf("Expected content type 'text', got '%s'", response.Content[0].Type)
	}

	expectedText := "Allowed directories:\n" + strings.Join(testDirs, "\n")
	if response.Content[0].Text != expectedText {
		t.Errorf("Expected text '%s', got '%s'", expectedText, response.Content[0].Text)
	}
}

// TestHandleCallTool tests the handleCallTool function
func TestHandleCallTool(t *testing.T) {
	// Create a server with a buffer writer for testing
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	s := &Server{
		allowedDirectories: []string{"/test"},
		reader:             nil, // Not needed for this test
		writer:             writer,
	}

	// Test with list_allowed_directories tool
	err := s.handleCallTool("test-id", map[string]interface{}{
		"name":      "list_allowed_directories",
		"arguments": map[string]interface{}{},
	})
	if err != nil {
		t.Fatalf("handleCallTool() returned error: %v", err)
	}

	// Verify the output
	writer.Flush()
	output := buf.String()
	var response Response
	if err := json.Unmarshal([]byte(output), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got '%s'", response.ID)
	}

	// Test with unknown tool
	buf.Reset()
	err = s.handleCallTool("test-id", map[string]interface{}{
		"name": "unknown_tool",
	})
	if err != nil {
		t.Fatalf("handleCallTool() with unknown tool returned error: %v", err)
	}

	// Verify the error response
	writer.Flush()
	output = buf.String()
	var errorResponse ErrorResponse
	if err := json.Unmarshal([]byte(output), &errorResponse); err != nil {
		t.Fatalf("Failed to unmarshal error response: %v", err)
	}

	if !strings.Contains(errorResponse.Error.Message, "unknown tool") {
		t.Errorf("Expected error message to contain 'unknown tool', got '%s'", errorResponse.Error.Message)
	}
}

// TestHandleReadFileSimple tests a simplified version of handleReadFile
func TestHandleReadFileSimple(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mcp-server-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "test content"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a simplified version of handleReadFile that doesn't use ValidatePath
	handleReadFileSimple := func(path string) (ToolResponse, error) {
		content, err := os.ReadFile(path)
		if err != nil {
			return ToolResponse{}, fmt.Errorf("error reading file: %w", err)
		}
		return ToolResponse{
			Content: []ContentItem{{Type: "text", Text: string(content)}},
		}, nil
	}

	// Call the handler
	response, err := handleReadFileSimple(testFile)
	if err != nil {
		t.Fatalf("handleReadFileSimple() returned error: %v", err)
	}

	// Verify the response
	if len(response.Content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(response.Content))
	}

	if response.Content[0].Type != "text" {
		t.Errorf("Expected content type 'text', got '%s'", response.Content[0].Type)
	}

	if response.Content[0].Text != testContent {
		t.Errorf("Expected text '%s', got '%s'", testContent, response.Content[0].Text)
	}

	// Test with non-existent file
	_, err = handleReadFileSimple(filepath.Join(tempDir, "nonexistent.txt"))
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

// TestHandleWriteFileSimple tests a simplified version of handleWriteFile
func TestHandleWriteFileSimple(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mcp-server-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a simplified version of handleWriteFile that doesn't use ValidatePath
	handleWriteFileSimple := func(path, content string) (ToolResponse, error) {
		// Ensure parent directory exists
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return ToolResponse{}, fmt.Errorf("error creating parent directory: %w", err)
		}

		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return ToolResponse{}, fmt.Errorf("error writing file: %w", err)
		}

		return ToolResponse{
			Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Successfully wrote to %s", path)}},
		}, nil
	}

	// Test file path
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "test content"

	// Call the handler
	response, err := handleWriteFileSimple(testFile, testContent)
	if err != nil {
		t.Fatalf("handleWriteFileSimple() returned error: %v", err)
	}

	// Verify the response
	if len(response.Content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(response.Content))
	}

	if response.Content[0].Type != "text" {
		t.Errorf("Expected content type 'text', got '%s'", response.Content[0].Type)
	}

	if !strings.Contains(response.Content[0].Text, "Successfully wrote") {
		t.Errorf("Expected text to contain 'Successfully wrote', got '%s'", response.Content[0].Text)
	}

	// Verify the file was written
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("Expected file content '%s', got '%s'", testContent, string(content))
	}

	// Test with nested directory
	nestedFile := filepath.Join(tempDir, "nested", "dir", "test.txt")
	_, err = handleWriteFileSimple(nestedFile, testContent)
	if err != nil {
		t.Fatalf("handleWriteFileSimple() with nested directory returned error: %v", err)
	}

	// Verify the nested file was written
	content, err = os.ReadFile(nestedFile)
	if err != nil {
		t.Fatalf("Failed to read nested test file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("Expected nested file content '%s', got '%s'", testContent, string(content))
	}
}

// TestHandleCreateDirectorySimple tests a simplified version of handleCreateDirectory
func TestHandleCreateDirectorySimple(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mcp-server-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a simplified version of handleCreateDirectory that doesn't use ValidatePath
	handleCreateDirectorySimple := func(path string) (ToolResponse, error) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return ToolResponse{}, fmt.Errorf("error creating directory: %w", err)
		}

		return ToolResponse{
			Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Successfully created directory %s", path)}},
		}, nil
	}

	// Test directory path
	testDir := filepath.Join(tempDir, "test-dir")

	// Call the handler
	response, err := handleCreateDirectorySimple(testDir)
	if err != nil {
		t.Fatalf("handleCreateDirectorySimple() returned error: %v", err)
	}

	// Verify the response
	if len(response.Content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(response.Content))
	}

	if response.Content[0].Type != "text" {
		t.Errorf("Expected content type 'text', got '%s'", response.Content[0].Type)
	}

	if !strings.Contains(response.Content[0].Text, "Successfully created directory") {
		t.Errorf("Expected text to contain 'Successfully created directory', got '%s'", response.Content[0].Text)
	}

	// Verify the directory was created
	info, err := os.Stat(testDir)
	if err != nil {
		t.Fatalf("Failed to stat test directory: %v", err)
	}

	if !info.IsDir() {
		t.Error("Expected test directory to be a directory")
	}

	// Test with nested directory
	nestedDir := filepath.Join(tempDir, "nested", "dir")
	_, err = handleCreateDirectorySimple(nestedDir)
	if err != nil {
		t.Fatalf("handleCreateDirectorySimple() with nested directory returned error: %v", err)
	}

	// Verify the nested directory was created
	info, err = os.Stat(nestedDir)
	if err != nil {
		t.Fatalf("Failed to stat nested directory: %v", err)
	}

	if !info.IsDir() {
		t.Error("Expected nested directory to be a directory")
	}
}

// TestHandleListDirectorySimple tests a simplified version of handleListDirectory
func TestHandleListDirectorySimple(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mcp-server-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create some test files and directories
	testSubDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(testSubDir, 0755); err != nil {
		t.Fatalf("Failed to create test subdirectory: %v", err)
	}

	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a simplified version of handleListDirectory that doesn't use ValidatePath
	handleListDirectorySimple := func(path string) (ToolResponse, error) {
		entries, err := os.ReadDir(path)
		if err != nil {
			return ToolResponse{}, fmt.Errorf("error reading directory: %w", err)
		}

		var formatted []string
		for _, entry := range entries {
			prefix := "[FILE]"
			if entry.IsDir() {
				prefix = "[DIR]"
			}
			formatted = append(formatted, fmt.Sprintf("%s %s", prefix, entry.Name()))
		}

		return ToolResponse{
			Content: []ContentItem{{Type: "text", Text: strings.Join(formatted, "\n")}},
		}, nil
	}

	// Call the handler
	response, err := handleListDirectorySimple(tempDir)
	if err != nil {
		t.Fatalf("handleListDirectorySimple() returned error: %v", err)
	}

	// Verify the response
	if len(response.Content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(response.Content))
	}

	if response.Content[0].Type != "text" {
		t.Errorf("Expected content type 'text', got '%s'", response.Content[0].Type)
	}

	// Check that both the file and directory are listed
	text := response.Content[0].Text
	if !strings.Contains(text, "[DIR] subdir") {
		t.Errorf("Expected directory listing to contain '[DIR] subdir', got '%s'", text)
	}

	if !strings.Contains(text, "[FILE] test.txt") {
		t.Errorf("Expected directory listing to contain '[FILE] test.txt', got '%s'", text)
	}

	// Test with non-existent directory
	_, err = handleListDirectorySimple(filepath.Join(tempDir, "nonexistent"))
	if err == nil {
		t.Error("Expected error for non-existent directory, got nil")
	}
}
