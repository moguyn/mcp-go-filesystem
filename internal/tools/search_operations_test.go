package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func TestSearchFiles(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	files := []string{
		"file1.txt",
		"file2.txt",
		"dir1/file3.txt",
		"dir1/dir2/file4.txt",
	}

	for _, file := range files {
		path := filepath.Join(tmpDir, file)
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", file, err)
		}
	}

	// Initialize server with allowed directories
	s := server.NewMCPServer("test-server", "1.0.0")
	RegisterTools(s, []string{tmpDir})

	testCases := []struct {
		name          string
		root          string
		pattern       string
		exclude       []string
		allowedDirs   []string
		shouldSucceed bool
		expectedFiles int
		expectedError string
	}{
		{
			name:          "Search all txt files",
			root:          tmpDir,
			pattern:       "*.txt",
			allowedDirs:   []string{tmpDir},
			shouldSucceed: true,
			expectedFiles: 4,
		},
		{
			name:          "Search with exclude pattern",
			root:          tmpDir,
			pattern:       "*.txt",
			exclude:       []string{"dir1/*"},
			allowedDirs:   []string{tmpDir},
			shouldSucceed: true,
			expectedFiles: 2,
		},
		{
			name:          "Search outside allowed directories",
			root:          "/etc",
			pattern:       "*",
			allowedDirs:   []string{tmpDir},
			shouldSucceed: false,
			expectedError: "invalid root path",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := mcp.CallToolRequest{
				Request: mcp.Request{
					Method: "tools/search_files",
				},
			}
			request.Params.Name = "search_files"
			request.Params.Arguments = map[string]interface{}{
				"root":    tc.root,
				"pattern": tc.pattern,
			}

			if len(tc.exclude) > 0 {
				excludeJSON, err := json.Marshal(tc.exclude)
				if err != nil {
					t.Fatalf("Failed to marshal exclude patterns: %v", err)
				}
				request.Params.Arguments["exclude"] = string(excludeJSON)
			}

			result, err := handleSearchFiles(request, tc.allowedDirs)
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

				textContent, ok := result.Content[0].(*mcp.TextContent)
				if !ok {
					t.Fatal("Expected TextContent but got different type")
				}
				content := textContent.Text

				// Count matches by splitting lines and filtering out empty lines
				lines := strings.Split(content, "\n")
				var matches int
				for _, line := range lines {
					if strings.TrimSpace(line) != "" && strings.HasSuffix(line, ".txt") {
						matches++
					}
				}

				if matches != tc.expectedFiles {
					t.Errorf("Expected %d matches but got %d", tc.expectedFiles, matches)
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
