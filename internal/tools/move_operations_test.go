package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func TestMoveFile(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create source file
	sourceFile := filepath.Join(tmpDir, "source.txt")
	sourceContent := "test content"
	err = os.WriteFile(sourceFile, []byte(sourceContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create source directory
	sourceDir := filepath.Join(tmpDir, "sourcedir")
	err = os.Mkdir(sourceDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Initialize server with allowed directories
	s := server.NewMCPServer("test-server", "1.0.0")
	RegisterTools(s, []string{tmpDir})

	testCases := []struct {
		name          string
		source        string
		destination   string
		allowedDirs   []string
		shouldSucceed bool
		expectedError string
	}{
		{
			name:          "Move file",
			source:        sourceFile,
			destination:   filepath.Join(tmpDir, "destination.txt"),
			allowedDirs:   []string{tmpDir},
			shouldSucceed: true,
		},
		{
			name:          "Move directory",
			source:        sourceDir,
			destination:   filepath.Join(tmpDir, "destdir"),
			allowedDirs:   []string{tmpDir},
			shouldSucceed: true,
		},
		{
			name:          "Move non-existent file",
			source:        filepath.Join(tmpDir, "nonexistent.txt"),
			destination:   filepath.Join(tmpDir, "dest.txt"),
			allowedDirs:   []string{tmpDir},
			shouldSucceed: false,
			expectedError: "no such file or directory",
		},
		{
			name:          "Move file outside allowed directories",
			source:        sourceFile,
			destination:   "/etc/test.txt",
			allowedDirs:   []string{tmpDir},
			shouldSucceed: false,
			expectedError: "invalid destination path",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset source file before each test
			if tc.source == sourceFile {
				err := os.WriteFile(sourceFile, []byte(sourceContent), 0644)
				if err != nil {
					t.Fatalf("Failed to reset source file: %v", err)
				}
			}

			request := mcp.CallToolRequest{
				Request: mcp.Request{
					Method: "tools/move_file",
				},
			}
			request.Params.Name = "move_file"
			request.Params.Arguments = map[string]interface{}{
				"source":      tc.source,
				"destination": tc.destination,
			}

			result, err := handleMoveFile(request, tc.allowedDirs)
			if tc.shouldSucceed {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Fatal("Expected result but got nil")
				}
				if result.IsError {
					t.Errorf("Expected success but got error: %v", result.Content)
				}

				// Verify source no longer exists
				if _, err := os.Stat(tc.source); !os.IsNotExist(err) {
					t.Error("Source still exists after move")
				}

				// Verify destination exists
				if _, err := os.Stat(tc.destination); os.IsNotExist(err) {
					t.Error("Destination does not exist after move")
				}

				if tc.source == sourceFile {
					// Verify file contents
					content, err := os.ReadFile(tc.destination)
					if err != nil {
						t.Fatalf("Failed to read destination file: %v", err)
					}
					if string(content) != sourceContent {
						t.Errorf("Expected content %q but got %q", sourceContent, string(content))
					}
				}
			} else {
				if err == nil {
					t.Error("Expected error but got none")
				} else if !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("Expected error containing %q but got %q", tc.expectedError, err.Error())
				}
			}
		})
	}
}
