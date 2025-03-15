package tools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
)

func TestGetFileInfo(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create a test directory
	testDir := filepath.Join(tmpDir, "testdir")
	err = os.Mkdir(testDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Initialize server with allowed directories
	s := server.NewMCPServer("test-server", "1.0.0")
	RegisterTools(s, []string{tmpDir})

	tests := []struct {
		name          string
		path          string
		expectedError bool
	}{
		{
			name:          "Get file info",
			path:          testFile,
			expectedError: false,
		},
		{
			name:          "Get directory info",
			path:          testDir,
			expectedError: false,
		},
		{
			name:          "Path outside allowed directories",
			path:          "/etc/passwd",
			expectedError: true,
		},
		{
			name:          "Non-existent path",
			path:          filepath.Join(tmpDir, "nonexistent"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{}
			request.Params.Name = "get_file_info"
			request.Params.Arguments = map[string]interface{}{
				"path": tt.path,
			}

			result, err := handleGetFileInfo(request, []string{tmpDir})
			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.False(t, result.IsError)
			assert.NotEmpty(t, result.Content)
		})
	}
}

func TestListAllowedDirectories(t *testing.T) {
	// Create temporary directories for testing
	tmpDir1, err := os.MkdirTemp("", "test1-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir1)

	tmpDir2, err := os.MkdirTemp("", "test2-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir2)

	// Initialize server with allowed directories
	s := server.NewMCPServer("test-server", "1.0.0")
	allowedDirs := []string{tmpDir1, tmpDir2}
	RegisterTools(s, allowedDirs)

	request := mcp.CallToolRequest{}
	request.Params.Name = "list_allowed_directories"
	request.Params.Arguments = map[string]interface{}{}

	result, err := handleListAllowedDirectories(request, allowedDirs)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.IsError)
	assert.NotEmpty(t, result.Content)
}
