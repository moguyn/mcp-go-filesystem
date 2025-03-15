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

func TestListDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files and directories
	files := map[string]string{
		"file1.txt": "content1",
		"file2.txt": "content2",
	}
	dirs := []string{"dir1", "dir2"}

	for name, content := range files {
		err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	for _, name := range dirs {
		err := os.Mkdir(filepath.Join(tmpDir, name), 0755)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Initialize server with allowed directories
	s := server.NewMCPServer("test-server", "1.0.0")
	RegisterTools(s, []string{tmpDir})

	testCases := []struct {
		name          string
		path          string
		allowedDirs   []string
		shouldSucceed bool
		expectedError string
	}{
		{
			name:          "List directory contents",
			path:          tmpDir,
			allowedDirs:   []string{tmpDir},
			shouldSucceed: true,
		},
		{
			name:          "List non-existent directory",
			path:          filepath.Join(tmpDir, "nonexistent"),
			allowedDirs:   []string{tmpDir},
			shouldSucceed: false,
			expectedError: "Error reading directory",
		},
		{
			name:          "List directory outside allowed directories",
			path:          "/etc",
			allowedDirs:   []string{tmpDir},
			shouldSucceed: false,
			expectedError: "Invalid path",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := mcp.CallToolRequest{
				Request: mcp.Request{
					Method: "tools/list_directory",
				},
			}
			request.Params.Name = "list_directory"
			request.Params.Arguments = map[string]interface{}{
				"path": tc.path,
			}

			result, err := handleListDirectory(request, tc.allowedDirs)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tc.shouldSucceed {
				if result.IsError {
					t.Errorf("Expected success but got error: %v", result.Content)
				}

				if len(result.Content) == 0 {
					t.Fatal("Expected content but got none")
				}

				textContent, ok := result.Content[0].(*mcp.TextContent)
				if !ok {
					// Try to convert from value to pointer
					if tc, ok := result.Content[0].(mcp.TextContent); ok {
						textContent = &tc
					} else {
						t.Fatal("Expected TextContent but got different type")
					}
				}

				content := textContent.Text

				// Check that all files and directories are listed
				for name := range files {
					if !strings.Contains(content, name) {
						t.Errorf("Expected output to contain file %q", name)
					}
				}
				for _, name := range dirs {
					if !strings.Contains(content, name) {
						t.Errorf("Expected output to contain directory %q", name)
					}
				}
			} else {
				if !result.IsError {
					t.Error("Expected error but got success")
				}
				if len(result.Content) == 0 {
					t.Fatal("Expected error content but got none")
				}
				textContent, ok := result.Content[0].(*mcp.TextContent)
				if !ok {
					// Try to convert from value to pointer
					if tc, ok := result.Content[0].(mcp.TextContent); ok {
						textContent = &tc
					} else {
						t.Fatal("Expected TextContent but got different type")
					}
				}
				if !strings.Contains(textContent.Text, tc.expectedError) {
					t.Errorf("Expected error containing %q but got %q", tc.expectedError, textContent.Text)
				}
			}
		})
	}
}

func TestCreateDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "tools-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test cases
	testCases := []struct {
		name          string
		path          string
		allowedDirs   []string
		shouldSucceed bool
		expectedError string
	}{
		{
			name:          "Create new directory",
			path:          filepath.Join(tempDir, "newdir"),
			allowedDirs:   []string{tempDir},
			shouldSucceed: true,
		},
		{
			name:          "Create nested directory",
			path:          filepath.Join(tempDir, "parent", "child"),
			allowedDirs:   []string{tempDir},
			shouldSucceed: true,
		},
		{
			name:          "Create directory outside allowed directories",
			path:          filepath.Join(tempDir, "outside"),
			allowedDirs:   []string{"/other/dir"},
			shouldSucceed: false,
			expectedError: "Invalid path",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := mcp.CallToolRequest{
				Request: mcp.Request{
					Method: "tools/create_directory",
				},
			}
			request.Params.Name = "create_directory"
			request.Params.Arguments = map[string]interface{}{
				"path": tc.path,
			}

			result, err := handleCreateDirectory(request, tc.allowedDirs)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tc.shouldSucceed {
				if result.IsError {
					t.Errorf("Expected success but got error: %v", result.Content)
				}

				// Verify directory exists
				info, err := os.Stat(tc.path)
				if err != nil {
					t.Errorf("Failed to stat created directory: %v", err)
				} else if !info.IsDir() {
					t.Error("Created path is not a directory")
				}
			} else {
				if !result.IsError {
					t.Error("Expected error but got success")
				}
				if len(result.Content) == 0 {
					t.Fatal("Expected error content but got none")
				}
				textContent, ok := result.Content[0].(*mcp.TextContent)
				if !ok {
					// Try to convert from value to pointer
					if tc, ok := result.Content[0].(mcp.TextContent); ok {
						textContent = &tc
					} else {
						t.Fatal("Expected TextContent but got different type")
					}
				}
				if !strings.Contains(textContent.Text, tc.expectedError) {
					t.Errorf("Expected error containing %q but got %q", tc.expectedError, textContent.Text)
				}
			}
		})
	}
}

func TestDirectoryTree(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "tools-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test directory structure
	dirs := []string{
		"dir1",
		"dir1/subdir1",
		"dir1/subdir2",
		"dir2",
		"dir2/subdir1",
	}

	files := []string{
		"dir1/file1.txt",
		"dir1/subdir1/file2.txt",
		"dir2/file3.txt",
	}

	for _, dir := range dirs {
		err := os.MkdirAll(filepath.Join(tempDir, dir), 0755)
		if err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	for _, file := range files {
		err := os.WriteFile(filepath.Join(tempDir, file), []byte("test"), 0644)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", file, err)
		}
	}

	// Test cases
	testCases := []struct {
		name          string
		path          string
		maxDepth      float64
		allowedDirs   []string
		shouldSucceed bool
		expectedError string
	}{
		{
			name:          "Get directory tree",
			path:          tempDir,
			maxDepth:      3,
			allowedDirs:   []string{tempDir},
			shouldSucceed: true,
		},
		{
			name:          "Get directory tree with depth 1",
			path:          tempDir,
			maxDepth:      1,
			allowedDirs:   []string{tempDir},
			shouldSucceed: true,
		},
		{
			name:          "Get tree for non-existent directory",
			path:          filepath.Join(tempDir, "nonexistent"),
			maxDepth:      3,
			allowedDirs:   []string{tempDir},
			shouldSucceed: false,
			expectedError: "Error accessing path",
		},
		{
			name:          "Get tree for directory outside allowed directories",
			path:          "/etc",
			maxDepth:      3,
			allowedDirs:   []string{tempDir},
			shouldSucceed: false,
			expectedError: "Invalid path",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := mcp.CallToolRequest{
				Request: mcp.Request{
					Method: "tools/directory_tree",
				},
			}
			request.Params.Name = "directory_tree"
			request.Params.Arguments = map[string]interface{}{
				"path":      tc.path,
				"max_depth": tc.maxDepth,
			}

			result, err := handleDirectoryTree(request, tc.allowedDirs)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tc.shouldSucceed {
				if result.IsError {
					t.Errorf("Expected success but got error: %v", result.Content)
				}

				if len(result.Content) == 0 {
					t.Fatal("Expected content but got none")
				}

				textContent, ok := result.Content[0].(*mcp.TextContent)
				if !ok {
					// Try to convert from value to pointer
					if tc, ok := result.Content[0].(mcp.TextContent); ok {
						textContent = &tc
					} else {
						t.Fatal("Expected TextContent but got different type")
					}
				}

				// Parse the JSON response
				var tree []TreeEntry
				if err := json.Unmarshal([]byte(strings.Split(textContent.Text, "\n\n")[1]), &tree); err != nil {
					t.Fatalf("Failed to parse tree JSON: %v", err)
				}

				// Verify tree structure
				verifyTree(t, tree, tempDir, tc.maxDepth)
			} else {
				if !result.IsError {
					t.Error("Expected error but got success")
				}
				if len(result.Content) == 0 {
					t.Fatal("Expected error content but got none")
				}
				textContent, ok := result.Content[0].(*mcp.TextContent)
				if !ok {
					// Try to convert from value to pointer
					if tc, ok := result.Content[0].(mcp.TextContent); ok {
						textContent = &tc
					} else {
						t.Fatal("Expected TextContent but got different type")
					}
				}
				if !strings.Contains(textContent.Text, tc.expectedError) {
					t.Errorf("Expected error containing %q but got %q", tc.expectedError, textContent.Text)
				}
			}
		})
	}
}

// verifyTree is a helper function to verify the directory tree structure
func verifyTree(t *testing.T, tree []TreeEntry, rootPath string, maxDepth float64) {
	for _, entry := range tree {
		path := filepath.Join(rootPath, entry.Name)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("Failed to stat entry %s: %v", entry.Name, err)
			continue
		}

		if info.IsDir() != (entry.Type == "directory") {
			t.Errorf("Entry %s: expected type %v but got %s", entry.Name, info.IsDir(), entry.Type)
		}

		if entry.Type == "directory" && maxDepth != 1 {
			verifyTree(t, entry.Children, path, maxDepth-1)
		}
	}
}
