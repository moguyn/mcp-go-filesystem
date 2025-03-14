package tools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestValidatePath(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "tools-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a subdirectory
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Test cases
	testCases := []struct {
		name          string
		path          string
		allowedDirs   []string
		shouldSucceed bool
	}{
		{
			name:          "Allowed directory",
			path:          tempDir,
			allowedDirs:   []string{tempDir},
			shouldSucceed: true,
		},
		{
			name:          "Subdirectory of allowed directory",
			path:          subDir,
			allowedDirs:   []string{tempDir},
			shouldSucceed: true,
		},
		{
			name:          "File in allowed directory",
			path:          filepath.Join(tempDir, "file.txt"),
			allowedDirs:   []string{tempDir},
			shouldSucceed: true,
		},
		{
			name:          "Parent of allowed directory",
			path:          filepath.Dir(tempDir),
			allowedDirs:   []string{tempDir},
			shouldSucceed: false,
		},
		{
			name:          "Unrelated directory",
			path:          "/tmp",
			allowedDirs:   []string{tempDir},
			shouldSucceed: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validPath, err := ValidatePath(tc.path, tc.allowedDirs)
			if tc.shouldSucceed {
				if err != nil {
					t.Errorf("ValidatePath(%q, %v) returned error: %v", tc.path, tc.allowedDirs, err)
				}
				if validPath == "" {
					t.Errorf("ValidatePath(%q, %v) returned empty path", tc.path, tc.allowedDirs)
				}
			} else {
				if err == nil {
					t.Errorf("ValidatePath(%q, %v) did not return error", tc.path, tc.allowedDirs)
				}
			}
		})
	}
}

func TestHandleReadFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "tools-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file
	testContent := "Hello, world!"
	testFilePath := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFilePath, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a request
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"path": testFilePath,
	}

	// Call the handler
	result, err := handleReadFile(request, []string{tempDir})
	if err != nil {
		t.Fatalf("handleReadFile returned an error: %v", err)
	}

	// Check the result
	if result.IsError {
		t.Errorf("handleReadFile returned an error result")
	}

	if len(result.Content) != 1 {
		t.Errorf("handleReadFile returned unexpected content length: got %d, want 1", len(result.Content))
		return
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Errorf("handleReadFile returned non-text content: %T", result.Content[0])
		return
	}

	if textContent.Text != testContent {
		t.Errorf("handleReadFile returned wrong content: got %q, want %q", textContent.Text, testContent)
	}
}

func TestHandleWriteFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "tools-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file path
	testContent := "Hello, world!"
	testFilePath := filepath.Join(tempDir, "test.txt")

	// Create a request
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"path":    testFilePath,
		"content": testContent,
	}

	// Call the handler
	result, err := handleWriteFile(request, []string{tempDir})
	if err != nil {
		t.Fatalf("handleWriteFile returned an error: %v", err)
	}

	// Check the result
	if result.IsError {
		t.Errorf("handleWriteFile returned an error result")
	}

	// Read the file to verify content
	content, err := os.ReadFile(testFilePath)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("handleWriteFile wrote wrong content: got %q, want %q", string(content), testContent)
	}
}

func TestHandleListAllowedDirectories(t *testing.T) {
	// Create test directories
	dirs := []string{"/tmp/dir1", "/tmp/dir2"}

	// Create a request
	request := mcp.CallToolRequest{}

	// Call the handler
	result, err := handleListAllowedDirectories(request, dirs)
	if err != nil {
		t.Fatalf("handleListAllowedDirectories returned an error: %v", err)
	}

	// Check the result
	if result.IsError {
		t.Errorf("handleListAllowedDirectories returned an error result")
	}

	if len(result.Content) != 1 {
		t.Errorf("handleListAllowedDirectories returned unexpected content length: got %d, want 1", len(result.Content))
		return
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Errorf("handleListAllowedDirectories returned non-text content: %T", result.Content[0])
		return
	}

	// Verify that both directories are listed in the result
	for _, dir := range dirs {
		if !contains(textContent.Text, dir) {
			t.Errorf("handleListAllowedDirectories did not include directory %s in result", dir)
		}
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
