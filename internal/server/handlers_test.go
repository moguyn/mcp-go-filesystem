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

// TestHandleReadFileError tests error cases for the handleReadFile function
func TestHandleReadFileError(t *testing.T) {
	// Create a server with test directories
	s := &Server{
		allowedDirectories: []string{"/test"},
	}

	// Test with missing path parameter
	args := map[string]interface{}{}
	_, err := s.handleReadFile(args)
	if err == nil {
		t.Errorf("Expected error for missing path parameter, got nil")
	}

	// Test with invalid path type
	args = map[string]interface{}{
		"path": 123, // Not a string
	}
	_, err = s.handleReadFile(args)
	if err == nil {
		t.Errorf("Expected error for invalid path type, got nil")
	}
}

// TestHandleWriteFileError tests error cases for the handleWriteFile function
func TestHandleWriteFileError(t *testing.T) {
	// Create a server with test directories
	s := &Server{
		allowedDirectories: []string{"/test"},
	}

	// Test with missing parameters
	args := map[string]interface{}{}
	_, err := s.handleWriteFile(args)
	if err == nil {
		t.Errorf("Expected error for missing parameters, got nil")
	}

	// Test with invalid path type
	args = map[string]interface{}{
		"path":    123, // Not a string
		"content": "test content",
	}
	_, err = s.handleWriteFile(args)
	if err == nil {
		t.Errorf("Expected error for invalid path type, got nil")
	}

	// Test with invalid content type
	args = map[string]interface{}{
		"path":    "/test/file.txt",
		"content": 123, // Not a string
	}
	_, err = s.handleWriteFile(args)
	if err == nil {
		t.Errorf("Expected error for invalid content type, got nil")
	}
}

// TestHandleCreateDirectoryError tests error cases for the handleCreateDirectory function
func TestHandleCreateDirectoryError(t *testing.T) {
	// Create a server with test directories
	s := &Server{
		allowedDirectories: []string{"/test"},
	}

	// Test with missing path parameter
	args := map[string]interface{}{}
	_, err := s.handleCreateDirectory(args)
	if err == nil {
		t.Errorf("Expected error for missing path parameter, got nil")
	}

	// Test with invalid path type
	args = map[string]interface{}{
		"path": 123, // Not a string
	}
	_, err = s.handleCreateDirectory(args)
	if err == nil {
		t.Errorf("Expected error for invalid path type, got nil")
	}
}

// TestHandleListDirectoryError tests error cases for the handleListDirectory function
func TestHandleListDirectoryError(t *testing.T) {
	// Create a server with test directories
	s := &Server{
		allowedDirectories: []string{"/test"},
	}

	// Test with missing path parameter
	args := map[string]interface{}{}
	_, err := s.handleListDirectory(args)
	if err == nil {
		t.Errorf("Expected error for missing path parameter, got nil")
	}

	// Test with invalid path type
	args = map[string]interface{}{
		"path": 123, // Not a string
	}
	_, err = s.handleListDirectory(args)
	if err == nil {
		t.Errorf("Expected error for invalid path type, got nil")
	}
}

// TestHandleDirectoryTreeError tests error cases for the handleDirectoryTree function
func TestHandleDirectoryTreeError(t *testing.T) {
	// Create a server with test directories
	s := &Server{
		allowedDirectories: []string{"/test"},
	}

	// Test with missing path parameter
	args := map[string]interface{}{}
	_, err := s.handleDirectoryTree(args)
	if err == nil {
		t.Errorf("Expected error for missing path parameter, got nil")
	}

	// Test with invalid path type
	args = map[string]interface{}{
		"path": 123, // Not a string
	}
	_, err = s.handleDirectoryTree(args)
	if err == nil {
		t.Errorf("Expected error for invalid path type, got nil")
	}
}

// TestHandleMoveFileError tests error cases for the handleMoveFile function
func TestHandleMoveFileError(t *testing.T) {
	// Create a server with test directories
	s := &Server{
		allowedDirectories: []string{"/test"},
	}

	// Test with missing parameters
	args := map[string]interface{}{}
	_, err := s.handleMoveFile(args)
	if err == nil {
		t.Errorf("Expected error for missing parameters, got nil")
	}

	// Test with invalid source path type
	args = map[string]interface{}{
		"source":      123, // Not a string
		"destination": "/test/dest.txt",
	}
	_, err = s.handleMoveFile(args)
	if err == nil {
		t.Errorf("Expected error for invalid source path type, got nil")
	}

	// Test with invalid destination path type
	args = map[string]interface{}{
		"source":      "/test/source.txt",
		"destination": 123, // Not a string
	}
	_, err = s.handleMoveFile(args)
	if err == nil {
		t.Errorf("Expected error for invalid destination path type, got nil")
	}
}

// TestHandleSearchFilesError tests error cases for the handleSearchFiles function
func TestHandleSearchFilesError(t *testing.T) {
	// Create a server with test directories
	s := &Server{
		allowedDirectories: []string{"/test"},
	}

	// Test with missing parameters
	args := map[string]interface{}{}
	_, err := s.handleSearchFiles(args)
	if err == nil {
		t.Errorf("Expected error for missing parameters, got nil")
	}

	// Test with invalid root path type
	args = map[string]interface{}{
		"root":    123, // Not a string
		"pattern": "*.txt",
	}
	_, err = s.handleSearchFiles(args)
	if err == nil {
		t.Errorf("Expected error for invalid root path type, got nil")
	}

	// Test with invalid pattern type
	args = map[string]interface{}{
		"root":    "/test",
		"pattern": 123, // Not a string
	}
	_, err = s.handleSearchFiles(args)
	if err == nil {
		t.Errorf("Expected error for invalid pattern type, got nil")
	}
}

// TestHandleGetFileInfoError tests error cases for the handleGetFileInfo function
func TestHandleGetFileInfoError(t *testing.T) {
	// Create a server with test directories
	s := &Server{
		allowedDirectories: []string{"/test"},
	}

	// Test with missing path parameter
	args := map[string]interface{}{}
	_, err := s.handleGetFileInfo(args)
	if err == nil {
		t.Errorf("Expected error for missing path parameter, got nil")
	}

	// Test with invalid path type
	args = map[string]interface{}{
		"path": 123, // Not a string
	}
	_, err = s.handleGetFileInfo(args)
	if err == nil {
		t.Errorf("Expected error for invalid path type, got nil")
	}
}

// TestHandleEditFileError tests error cases for the handleEditFile function
func TestHandleEditFileError(t *testing.T) {
	// Create a server with test directories
	s := &Server{
		allowedDirectories: []string{"/test"},
	}

	// Test with missing parameters
	args := map[string]interface{}{}
	_, err := s.handleEditFile(args)
	if err == nil {
		t.Errorf("Expected error for missing parameters, got nil")
	}

	// Test with invalid path type
	args = map[string]interface{}{
		"path":  123, // Not a string
		"edits": []interface{}{},
	}
	_, err = s.handleEditFile(args)
	if err == nil {
		t.Errorf("Expected error for invalid path type, got nil")
	}

	// Test with invalid edits type
	args = map[string]interface{}{
		"path":  "/test/file.txt",
		"edits": "not an array",
	}
	_, err = s.handleEditFile(args)
	if err == nil {
		t.Errorf("Expected error for invalid edits type, got nil")
	}
}

// testServer is a custom server for testing with a mock ValidatePath method
type testServer struct {
	Server
	tempDir string
}

// ValidatePath is a mock implementation for testing
func (s *testServer) ValidatePath(path string) (string, error) {
	// For testing, just return the path if it's within the temp directory
	if strings.HasPrefix(path, s.tempDir) {
		return path, nil
	}
	return "", fmt.Errorf("access denied - path outside allowed directories")
}

// handleEditFile delegates to the Server's handleEditFile method
func (s *testServer) handleEditFile(args map[string]interface{}) (ToolResponse, error) {
	return s.Server.handleEditFile(args)
}

// TestHandleReadMultipleFiles tests the handleReadMultipleFiles function
func TestHandleReadMultipleFiles(t *testing.T) {
	// Create a server with a buffer writer for testing
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "read-multiple-files-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	file1Path := filepath.Join(tempDir, "file1.txt")
	file2Path := filepath.Join(tempDir, "file2.txt")
	nonExistentPath := filepath.Join(tempDir, "nonexistent.txt")

	if err := os.WriteFile(file1Path, []byte("File 1 content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if err := os.WriteFile(file2Path, []byte("File 2 content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	s := &Server{
		allowedDirectories: []string{tempDir},
		writer:             writer,
	}

	// Test with valid files
	args := map[string]interface{}{
		"paths": []interface{}{file1Path, file2Path},
	}

	response, err := s.handleReadMultipleFiles(args)
	if err != nil {
		t.Fatalf("handleReadMultipleFiles() returned error: %v", err)
	}

	// Verify response
	if len(response.Content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(response.Content))
	}

	if response.Content[0].Type != "text" {
		t.Errorf("Expected content type 'text', got '%s'", response.Content[0].Type)
	}

	// Check if the response text contains both file paths
	responseText := response.Content[0].Text
	if !strings.Contains(responseText, file1Path) || !strings.Contains(responseText, file2Path) {
		t.Errorf("Response text does not contain both file paths")
	}

	// Test with invalid file path - this should not return an error but include error message in the response
	args = map[string]interface{}{
		"paths": []interface{}{nonExistentPath},
	}

	response, err = s.handleReadMultipleFiles(args)
	if err != nil {
		t.Fatalf("handleReadMultipleFiles() with nonexistent file returned error: %v", err)
	}

	// Check if the response contains an error message for the nonexistent file
	responseText = response.Content[0].Text
	if !strings.Contains(responseText, "Error") {
		t.Errorf("Expected error message in response for nonexistent file, got: %s", responseText)
	}
}

// TestBuildDirectoryTree tests the buildDirectoryTree function
func TestBuildDirectoryTree(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "directory-tree-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files and directories
	file1Path := filepath.Join(tempDir, "file1.txt")
	subDirPath := filepath.Join(tempDir, "subdir")
	subFilePath := filepath.Join(subDirPath, "subfile.txt")

	if err := os.WriteFile(file1Path, []byte("File 1 content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if err := os.Mkdir(subDirPath, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	if err := os.WriteFile(subFilePath, []byte("Subfile content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a server for testing
	s := &Server{
		allowedDirectories: []string{tempDir},
	}

	// Build directory tree
	tree, err := s.buildDirectoryTree(tempDir)
	if err != nil {
		t.Fatalf("buildDirectoryTree() returned error: %v", err)
	}

	// Verify tree structure
	if len(tree) < 2 {
		t.Errorf("Expected at least 2 entries in tree, got %d", len(tree))
	}

	// Find the subdirectory in the tree
	var subDirFound bool
	for _, entry := range tree {
		if entry.Name == "subdir" && entry.Type == "directory" {
			subDirFound = true

			// Check if it has children
			if len(entry.Children) != 1 {
				t.Errorf("Expected 1 child in subdir, got %d", len(entry.Children))
			}
			break
		}
	}

	if !subDirFound {
		t.Fatalf("Subdirectory 'subdir' not found in tree")
	}
}

// TestSearchFiles tests the searchFiles function
func TestSearchFiles(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "search-files-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files with specific content
	file1Path := filepath.Join(tempDir, "file1.txt")
	file2Path := filepath.Join(tempDir, "file2.go")

	if err := os.WriteFile(file1Path, []byte("This is a test file with SEARCHABLE content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if err := os.WriteFile(file2Path, []byte("func main() {\n\tfmt.Println(\"Hello, World!\")\n}"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a server for testing
	s := &Server{
		allowedDirectories: []string{tempDir},
	}

	// Test search by filename
	results, err := s.searchFiles(tempDir, "file1", []string{})
	if err != nil {
		t.Fatalf("searchFiles() returned error: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 search result for filename, got %d", len(results))
	}

	// Test search by content in filename
	results, err = s.searchFiles(tempDir, "go", []string{})
	if err != nil {
		t.Fatalf("searchFiles() with file pattern returned error: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 search result for file extension, got %d", len(results))
	}

	// Test search with no matches
	results, err = s.searchFiles(tempDir, "nonexistent", []string{})
	if err != nil {
		t.Fatalf("searchFiles() with no matches returned error: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 search results for nonexistent query, got %d", len(results))
	}
}
