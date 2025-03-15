package tools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func TestReadFile(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "test content"
	err = os.WriteFile(testFile, []byte(testContent), 0644)
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
			name:          "Read existing file",
			path:          testFile,
			expectedError: false,
		},
		{
			name:          "Read non-existent file",
			path:          filepath.Join(tmpDir, "nonexistent.txt"),
			expectedError: true,
		},
		{
			name:          "Read file outside allowed directories",
			path:          "/etc/passwd",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{
				Request: mcp.Request{
					Method: "tools/read_file",
				},
			}
			request.Params.Name = "read_file"
			request.Params.Arguments = map[string]interface{}{
				"path": tt.path,
			}

			result, err := handleReadFile(request, []string{tmpDir})
			if tt.expectedError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if result == nil {
				t.Fatal("Expected result but got nil")
			}
			if result.IsError {
				t.Errorf("Expected success but got error: %v", result.Content)
			}
			if len(result.Content) == 0 {
				t.Fatal("Expected content but got none")
			}

			if tt.path == testFile {
				textContent, ok := result.Content[0].(*mcp.TextContent)
				if !ok {
					// Try to convert from value to pointer
					if tc, ok := result.Content[0].(mcp.TextContent); ok {
						textContent = &tc
					} else {
						t.Fatal("Expected TextContent but got different type")
					}
				}
				if textContent.Text != testContent {
					t.Errorf("Expected content %q but got %q", testContent, textContent.Text)
				}
			}
		})
	}
}

func TestWriteFile(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize server with allowed directories
	s := server.NewMCPServer("test-server", "1.0.0")
	RegisterTools(s, []string{tmpDir})

	tests := []struct {
		name          string
		path          string
		content       string
		append        bool
		expectedError bool
	}{
		{
			name:          "Write new file",
			path:          filepath.Join(tmpDir, "new.txt"),
			content:       "new content",
			append:        false,
			expectedError: false,
		},
		{
			name:          "Overwrite existing file",
			path:          filepath.Join(tmpDir, "existing.txt"),
			content:       "overwritten content",
			append:        false,
			expectedError: false,
		},
		{
			name:          "Append to existing file",
			path:          filepath.Join(tmpDir, "append.txt"),
			content:       " appended content",
			append:        true,
			expectedError: false,
		},
		{
			name:          "Write file outside allowed directories",
			path:          "/etc/test.txt",
			content:       "test content",
			append:        false,
			expectedError: true,
		},
	}

	// Create existing files for testing
	existingFile := filepath.Join(tmpDir, "existing.txt")
	err = os.WriteFile(existingFile, []byte("original content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	appendFile := filepath.Join(tmpDir, "append.txt")
	err = os.WriteFile(appendFile, []byte("original content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{
				Request: mcp.Request{
					Method: "tools/write_file",
				},
			}
			request.Params.Name = "write_file"
			request.Params.Arguments = map[string]interface{}{
				"path":    tt.path,
				"content": tt.content,
				"append":  tt.append,
			}

			result, err := handleWriteFile(request, []string{tmpDir})
			if tt.expectedError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if result == nil {
				t.Fatal("Expected result but got nil")
			}
			if result.IsError {
				t.Errorf("Expected success but got error: %v", result.Content)
			}

			// Verify file contents
			content, err := os.ReadFile(tt.path)
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}

			if tt.append {
				if string(content) != "original content"+tt.content {
					t.Errorf("Expected content %q but got %q", "original content"+tt.content, string(content))
				}
			} else {
				if string(content) != tt.content {
					t.Errorf("Expected content %q but got %q", tt.content, string(content))
				}
			}
		})
	}
}
