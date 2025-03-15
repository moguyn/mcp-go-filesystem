package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
)

func TestNewServiceProvider(t *testing.T) {
	allowedDirs := []string{"/tmp", "/var"}
	provider := NewServiceProvider(allowedDirs)

	assert.NotNil(t, provider)
	assert.NotNil(t, provider.fileService)
	assert.NotNil(t, provider.fileWriter)
	assert.NotNil(t, provider.fileManager)
	assert.NotNil(t, provider.directoryService)
	assert.NotNil(t, provider.searchService)
	assert.NotNil(t, provider.logger)
}

func TestRegisterTools(t *testing.T) {
	// Create a test server
	s := server.NewMCPServer("test-server", "1.0.0")

	// Register tools
	allowedDirs := []string{"/tmp", "/var"}
	RegisterTools(s, allowedDirs)

	// Verify tools were registered
	// This is a basic test since we can't easily check the internal state of the server
	assert.NotNil(t, s)
}

func setupTestEnvironment(t *testing.T) (string, *ServiceProvider, func()) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-handlers-*")
	if err != nil {
		t.Fatal(err)
	}

	// Create test files and directories
	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	testDir := filepath.Join(tmpDir, "testdir")
	err = os.Mkdir(testDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create a service provider
	provider := NewServiceProvider([]string{tmpDir})

	// Return cleanup function
	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, provider, cleanup
}

func TestHandleReadFile(t *testing.T) {
	tmpDir, provider, cleanup := setupTestEnvironment(t)
	defer cleanup()

	testFile := filepath.Join(tmpDir, "test.txt")

	// Create test request
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"path": testFile,
	}

	// Call handler
	result, err := provider.handleReadFile(context.Background(), request)

	// Verify result
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Extract text content from the result
	textContent, ok := result.Content[0].(mcp.TextContent)
	assert.True(t, ok)
	assert.Equal(t, "test content", textContent.Text)
}

func TestHandleReadMultipleFiles(t *testing.T) {
	tmpDir, provider, cleanup := setupTestEnvironment(t)
	defer cleanup()

	testFile := filepath.Join(tmpDir, "test.txt")

	// Create test request
	request := mcp.CallToolRequest{}
	pathsJSON, _ := json.Marshal([]string{testFile})
	request.Params.Arguments = map[string]interface{}{
		"paths": string(pathsJSON),
	}

	// Call handler
	result, err := provider.handleReadMultipleFiles(context.Background(), request)

	// Verify result
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Extract text content from the result
	textContent, ok := result.Content[0].(mcp.TextContent)
	assert.True(t, ok)

	var results []FileContent
	err = json.Unmarshal([]byte(textContent.Text), &results)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, testFile, results[0].Path)
	assert.Equal(t, "test content", results[0].Content)
}

func TestHandleWriteFile(t *testing.T) {
	tmpDir, provider, cleanup := setupTestEnvironment(t)
	defer cleanup()

	newFile := filepath.Join(tmpDir, "new.txt")

	// Create test request
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"path":    newFile,
		"content": "new content",
		"append":  false,
	}

	// Call handler
	result, err := provider.handleWriteFile(context.Background(), request)

	// Verify result
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Check file was created
	content, err := os.ReadFile(newFile)
	assert.NoError(t, err)
	assert.Equal(t, "new content", string(content))
}

func TestHandleEditFile(t *testing.T) {
	tmpDir, provider, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create a file with multiple lines
	multilineFile := filepath.Join(tmpDir, "multiline.txt")
	multilineContent := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5"
	err := os.WriteFile(multilineFile, []byte(multilineContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create test request
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"path":       multilineFile,
		"content":    "New Line 2\nNew Line 3",
		"start_line": float64(2),
		"end_line":   float64(3),
	}

	// Call handler
	result, err := provider.handleEditFile(context.Background(), request)

	// Verify result
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Check file was edited
	content, err := os.ReadFile(multilineFile)
	assert.NoError(t, err)
	assert.Equal(t, "Line 1\nNew Line 2\nNew Line 3\nLine 4\nLine 5", string(content))
}

func TestHandleListDirectory(t *testing.T) {
	tmpDir, provider, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create test request
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"path": tmpDir,
	}

	// Call handler
	result, err := provider.handleListDirectory(context.Background(), request)

	// Verify result
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Extract text content from the result
	textContent, ok := result.Content[0].(mcp.TextContent)
	assert.True(t, ok)

	var entries []FileInfo
	err = json.Unmarshal([]byte(textContent.Text), &entries)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(entries)) // test.txt and testdir
}

func TestHandleCreateDirectory(t *testing.T) {
	tmpDir, provider, cleanup := setupTestEnvironment(t)
	defer cleanup()

	newDir := filepath.Join(tmpDir, "newdir")

	// Create test request
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"path": newDir,
	}

	// Call handler
	result, err := provider.handleCreateDirectory(context.Background(), request)

	// Verify result
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Check directory was created
	info, err := os.Stat(newDir)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestHandleDeleteDirectory(t *testing.T) {
	tmpDir, provider, cleanup := setupTestEnvironment(t)
	defer cleanup()

	dirToDelete := filepath.Join(tmpDir, "testdir")

	// Create test request
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"path":      dirToDelete,
		"recursive": true,
	}

	// Call handler
	result, err := provider.handleDeleteDirectory(context.Background(), request)

	// Verify result
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Check directory was deleted
	_, err = os.Stat(dirToDelete)
	assert.True(t, os.IsNotExist(err))
}

func TestHandleDeleteFile(t *testing.T) {
	tmpDir, provider, cleanup := setupTestEnvironment(t)
	defer cleanup()

	fileToDelete := filepath.Join(tmpDir, "test.txt")

	// Create test request
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"path": fileToDelete,
	}

	// Call handler
	result, err := provider.handleDeleteFile(context.Background(), request)

	// Verify result
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Check file was deleted
	_, err = os.Stat(fileToDelete)
	assert.True(t, os.IsNotExist(err))
}

func TestHandleMoveFile(t *testing.T) {
	tmpDir, provider, cleanup := setupTestEnvironment(t)
	defer cleanup()

	sourceFile := filepath.Join(tmpDir, "test.txt")
	destFile := filepath.Join(tmpDir, "moved.txt")

	// Create test request
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"source_path":      sourceFile,
		"destination_path": destFile,
	}

	// Call handler
	result, err := provider.handleMoveFile(context.Background(), request)

	// Verify result
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Check file was moved
	_, err = os.Stat(sourceFile)
	assert.True(t, os.IsNotExist(err))

	_, err = os.Stat(destFile)
	assert.NoError(t, err)
}

func TestHandleCopyFile(t *testing.T) {
	tmpDir, provider, cleanup := setupTestEnvironment(t)
	defer cleanup()

	sourceFile := filepath.Join(tmpDir, "test.txt")
	destFile := filepath.Join(tmpDir, "copied.txt")

	// Create test request
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"source_path":      sourceFile,
		"destination_path": destFile,
	}

	// Call handler
	result, err := provider.handleCopyFile(context.Background(), request)

	// Verify result
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Check file was copied
	_, err = os.Stat(sourceFile)
	assert.NoError(t, err)

	_, err = os.Stat(destFile)
	assert.NoError(t, err)
}

func TestHandleSearchFiles(t *testing.T) {
	tmpDir, provider, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create test request
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"query":     "test",
		"path":      tmpDir,
		"recursive": true,
	}

	// Call handler
	result, err := provider.handleSearchFiles(context.Background(), request)

	// Verify result
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Extract text content from the result
	textContent, ok := result.Content[0].(mcp.TextContent)
	assert.True(t, ok)

	var searchResults []SearchResult
	err = json.Unmarshal([]byte(textContent.Text), &searchResults)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(searchResults)) // Should find "test content" in test.txt
}
