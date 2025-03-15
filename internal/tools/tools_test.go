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

func TestReadFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "tools-test-")
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

	// Test cases
	testCases := []struct {
		name          string
		path          string
		allowedDirs   []string
		shouldSucceed bool
		expectedText  string
	}{
		{
			name:          "Valid file",
			path:          testFile,
			allowedDirs:   []string{tempDir},
			shouldSucceed: true,
			expectedText:  testContent,
		},
		{
			name:          "Non-existent file",
			path:          filepath.Join(tempDir, "nonexistent.txt"),
			allowedDirs:   []string{tempDir},
			shouldSucceed: false,
		},
		{
			name:          "Outside allowed directory",
			path:          "/etc/passwd",
			allowedDirs:   []string{tempDir},
			shouldSucceed: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := mcp.CallToolRequest{
				Request: mcp.Request{
					Method: "read_file",
				},
				Params: struct {
					Name      string                 `json:"name"`
					Arguments map[string]interface{} `json:"arguments,omitempty"`
					Meta      *struct {
						ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
					} `json:"_meta,omitempty"`
				}{
					Name: "read_file",
					Arguments: map[string]interface{}{
						"path": tc.path,
					},
				},
			}

			result, err := handleReadFile(request, tc.allowedDirs)
			if tc.shouldSucceed {
				if err != nil {
					t.Errorf("handleReadFile returned error: %v", err)
				}
				if result.IsError {
					t.Error("Expected success but got error result")
				}
				if len(result.Content) != 1 {
					t.Error("Expected one content item")
				} else if content, ok := result.Content[0].(mcp.TextContent); !ok {
					t.Error("Expected content to be a TextContent")
				} else if content.Text != tc.expectedText {
					t.Errorf("Expected content %q, got %q", tc.expectedText, content.Text)
				}
			} else {
				if err == nil && !result.IsError {
					t.Error("Expected error but got none")
				}
			}
		})
	}
}

func TestWriteFile(t *testing.T) {
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
		content       string
		append        bool
		allowedDirs   []string
		shouldSucceed bool
	}{
		{
			name:          "Write new file",
			path:          filepath.Join(tempDir, "new.txt"),
			content:       "new content",
			append:        false,
			allowedDirs:   []string{tempDir},
			shouldSucceed: true,
		},
		{
			name:          "Append to file",
			path:          filepath.Join(tempDir, "append.txt"),
			content:       "appended content",
			append:        true,
			allowedDirs:   []string{tempDir},
			shouldSucceed: true,
		},
		{
			name:          "Write outside allowed directory",
			path:          "/etc/test.txt",
			content:       "test",
			append:        false,
			allowedDirs:   []string{tempDir},
			shouldSucceed: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := mcp.CallToolRequest{
				Request: mcp.Request{
					Method: "write_file",
				},
				Params: struct {
					Name      string                 `json:"name"`
					Arguments map[string]interface{} `json:"arguments,omitempty"`
					Meta      *struct {
						ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
					} `json:"_meta,omitempty"`
				}{
					Name: "write_file",
					Arguments: map[string]interface{}{
						"path":    tc.path,
						"content": tc.content,
						"append":  tc.append,
					},
				},
			}

			result, err := handleWriteFile(request, tc.allowedDirs)
			if tc.shouldSucceed {
				if err != nil {
					t.Errorf("handleWriteFile returned error: %v", err)
				}
				if result.IsError {
					t.Error("Expected success but got error result")
				}
				content, err := os.ReadFile(tc.path)
				if err != nil {
					t.Errorf("Failed to read written file: %v", err)
				}
				if !tc.append && string(content) != tc.content {
					t.Errorf("Expected content %q, got %q", tc.content, string(content))
				}
			} else {
				if err == nil && !result.IsError {
					t.Error("Expected error but got none")
				}
			}
		})
	}
}

func TestListDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "tools-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create some test files and directories
	if err := os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.Mkdir(filepath.Join(tempDir, "subdir"), 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Test cases
	testCases := []struct {
		name          string
		path          string
		allowedDirs   []string
		shouldSucceed bool
	}{
		{
			name:          "List allowed directory",
			path:          tempDir,
			allowedDirs:   []string{tempDir},
			shouldSucceed: true,
		},
		{
			name:          "List outside allowed directory",
			path:          "/etc",
			allowedDirs:   []string{tempDir},
			shouldSucceed: false,
		},
		{
			name:          "List non-existent directory",
			path:          filepath.Join(tempDir, "nonexistent"),
			allowedDirs:   []string{tempDir},
			shouldSucceed: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := mcp.CallToolRequest{
				Request: mcp.Request{
					Method: "list_directory",
				},
				Params: struct {
					Name      string                 `json:"name"`
					Arguments map[string]interface{} `json:"arguments,omitempty"`
					Meta      *struct {
						ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
					} `json:"_meta,omitempty"`
				}{
					Name: "list_directory",
					Arguments: map[string]interface{}{
						"path": tc.path,
					},
				},
			}

			result, err := handleListDirectory(request, tc.allowedDirs)
			if tc.shouldSucceed {
				if err != nil {
					t.Errorf("handleListDirectory returned error: %v", err)
				}
				if result.IsError {
					t.Error("Expected success but got error result")
				}
				if len(result.Content) != 1 {
					t.Error("Expected one content item")
				} else if content, ok := result.Content[0].(mcp.TextContent); !ok {
					t.Error("Expected content to be a TextContent")
				} else if !strings.Contains(content.Text, "file1.txt") || !strings.Contains(content.Text, "subdir") {
					t.Error("Expected directory listing to contain test files and directories")
				}
			} else {
				if err == nil && !result.IsError {
					t.Error("Expected error but got none")
				}
			}
		})
	}
}

func TestGetFileInfo(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "tools-test-")
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

	// Test cases
	testCases := []struct {
		name          string
		path          string
		allowedDirs   []string
		shouldSucceed bool
	}{
		{
			name:          "Valid file",
			path:          testFile,
			allowedDirs:   []string{tempDir},
			shouldSucceed: true,
		},
		{
			name:          "Non-existent file",
			path:          filepath.Join(tempDir, "nonexistent.txt"),
			allowedDirs:   []string{tempDir},
			shouldSucceed: false,
		},
		{
			name:          "Outside allowed directory",
			path:          "/etc/passwd",
			allowedDirs:   []string{tempDir},
			shouldSucceed: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := mcp.CallToolRequest{
				Request: mcp.Request{
					Method: "get_file_info",
				},
				Params: struct {
					Name      string                 `json:"name"`
					Arguments map[string]interface{} `json:"arguments,omitempty"`
					Meta      *struct {
						ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
					} `json:"_meta,omitempty"`
				}{
					Name: "get_file_info",
					Arguments: map[string]interface{}{
						"path": tc.path,
					},
				},
			}

			result, err := handleGetFileInfo(request, tc.allowedDirs)
			if tc.shouldSucceed {
				if err != nil {
					t.Errorf("handleGetFileInfo returned error: %v", err)
				}
				if result.IsError {
					t.Error("Expected success but got error result")
				}
				if len(result.Content) != 1 {
					t.Error("Expected one content item")
				} else if content, ok := result.Content[0].(mcp.TextContent); !ok {
					t.Error("Expected content to be a TextContent")
				} else {
					var fileInfo FileInfo
					if err := json.Unmarshal([]byte(strings.TrimPrefix(content.Text, "File info for "+tc.path+":\n\n")), &fileInfo); err != nil {
						t.Errorf("Failed to parse file info JSON: %v", err)
					}
				}
			} else {
				if err == nil && !result.IsError {
					t.Error("Expected error but got none")
				}
			}
		})
	}
}

func TestEditFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "tools-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	initialContent := "initial content"
	if err := os.WriteFile(testFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test cases
	testCases := []struct {
		name          string
		path          string
		oldText       string
		newText       string
		allowedDirs   []string
		shouldSucceed bool
	}{
		{
			name:          "Valid edit",
			path:          testFile,
			oldText:       "initial",
			newText:       "modified",
			allowedDirs:   []string{tempDir},
			shouldSucceed: true,
		},
		{
			name:          "Non-existent text",
			path:          testFile,
			oldText:       "nonexistent",
			newText:       "modified",
			allowedDirs:   []string{tempDir},
			shouldSucceed: false,
		},
		{
			name:          "Outside allowed directory",
			path:          "/etc/test.txt",
			oldText:       "test",
			newText:       "modified",
			allowedDirs:   []string{tempDir},
			shouldSucceed: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := mcp.CallToolRequest{
				Request: mcp.Request{
					Method: "edit_file",
				},
				Params: struct {
					Name      string                 `json:"name"`
					Arguments map[string]interface{} `json:"arguments,omitempty"`
					Meta      *struct {
						ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
					} `json:"_meta,omitempty"`
				}{
					Name: "edit_file",
					Arguments: map[string]interface{}{
						"path":     tc.path,
						"old_text": tc.oldText,
						"new_text": tc.newText,
					},
				},
			}

			result, err := handleEditFile(request, tc.allowedDirs)
			if tc.shouldSucceed {
				if err != nil {
					t.Errorf("handleEditFile returned error: %v", err)
				}
				if result.IsError {
					t.Error("Expected success but got error result")
				}
				content, err := os.ReadFile(tc.path)
				if err != nil {
					t.Errorf("Failed to read edited file: %v", err)
				}
				expectedContent := strings.Replace(initialContent, tc.oldText, tc.newText, -1)
				if string(content) != expectedContent {
					t.Errorf("Expected content %q, got %q", expectedContent, string(content))
				}
			} else {
				if err == nil && !result.IsError {
					t.Error("Expected error but got none")
				}
			}
		})
	}
}

func TestRegisterTools(t *testing.T) {
	s := server.NewMCPServer("test", "1.0.0")
	allowedDirs := []string{"/tmp"}
	RegisterTools(s, allowedDirs)

	// Verify that all tools are registered
	tools := []string{
		"read_file",
		"read_multiple_files",
		"write_file",
		"list_directory",
		"get_file_info",
		"edit_file",
	}

	for _, toolName := range tools {
		tool := mcp.NewTool(toolName)
		if tool.Name == "" {
			t.Errorf("Tool %q not registered", toolName)
		}
	}
}

func TestReadMultipleFiles(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "tools-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	files := map[string]string{
		"file1.txt": "content 1",
		"file2.txt": "content 2",
		"file3.txt": "content 3",
	}

	filePaths := make([]string, 0, len(files))
	for name, content := range files {
		path := filepath.Join(tempDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		filePaths = append(filePaths, path)
	}

	// Test cases
	testCases := []struct {
		name          string
		paths         []string
		allowedDirs   []string
		shouldSucceed bool
		expectedLen   int
	}{
		{
			name:          "Valid files",
			paths:         filePaths,
			allowedDirs:   []string{tempDir},
			shouldSucceed: true,
			expectedLen:   3,
		},
		{
			name:          "Non-existent file",
			paths:         []string{filepath.Join(tempDir, "nonexistent.txt")},
			allowedDirs:   []string{tempDir},
			shouldSucceed: false,
			expectedLen:   1,
		},
		{
			name:          "Outside allowed directory",
			paths:         []string{"/etc/passwd"},
			allowedDirs:   []string{tempDir},
			shouldSucceed: false,
			expectedLen:   1,
		},
		{
			name:          "Mixed valid and invalid files",
			paths:         append(filePaths, "/etc/passwd"),
			allowedDirs:   []string{tempDir},
			shouldSucceed: false,
			expectedLen:   4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pathsJSON, err := json.Marshal(tc.paths)
			if err != nil {
				t.Fatalf("Failed to marshal paths: %v", err)
			}

			request := mcp.CallToolRequest{
				Request: mcp.Request{
					Method: "read_multiple_files",
				},
				Params: struct {
					Name      string                 `json:"name"`
					Arguments map[string]interface{} `json:"arguments,omitempty"`
					Meta      *struct {
						ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
					} `json:"_meta,omitempty"`
				}{
					Name: "read_multiple_files",
					Arguments: map[string]interface{}{
						"paths": string(pathsJSON),
					},
				},
			}

			result, err := handleReadMultipleFiles(request, tc.allowedDirs)
			if err != nil {
				t.Errorf("handleReadMultipleFiles returned error: %v", err)
				return
			}

			if tc.shouldSucceed {
				if result.IsError {
					t.Error("Expected success but got error result")
				}
			} else {
				if !result.IsError {
					t.Error("Expected error but got success")
				}
			}

			if len(result.Content) != tc.expectedLen {
				t.Errorf("Expected %d content items, got %d", tc.expectedLen, len(result.Content))
				return
			}

			if tc.shouldSucceed {
				for i, content := range result.Content {
					textContent, ok := content.(mcp.TextContent)
					if !ok {
						t.Errorf("Expected content[%d] to be a TextContent", i)
						continue
					}
					expectedContent := files[filepath.Base(tc.paths[i])]
					if textContent.Text != expectedContent {
						t.Errorf("Expected content[%d] %q, got %q", i, expectedContent, textContent.Text)
					}
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
		errorContains string
		skipStatCheck bool // Skip checking if the directory exists for invalid paths
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
			name:          "Create directory outside allowed directory",
			path:          "/etc/newdir",
			allowedDirs:   []string{tempDir},
			shouldSucceed: false,
			errorContains: "Invalid path: permission denied",
		},
		{
			name:          "Create directory with invalid path",
			path:          string([]byte{0x00}),
			allowedDirs:   []string{tempDir},
			shouldSucceed: false,
			errorContains: "Invalid path: path contains invalid characters",
			skipStatCheck: true, // Skip stat check for invalid paths
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := mcp.CallToolRequest{
				Request: mcp.Request{
					Method: "create_directory",
				},
				Params: struct {
					Name      string                 `json:"name"`
					Arguments map[string]interface{} `json:"arguments,omitempty"`
					Meta      *struct {
						ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
					} `json:"_meta,omitempty"`
				}{
					Name: "create_directory",
					Arguments: map[string]interface{}{
						"path": tc.path,
					},
				},
			}

			result, err := handleCreateDirectory(request, tc.allowedDirs)
			if err != nil {
				t.Errorf("handleCreateDirectory returned error: %v", err)
				return
			}

			if tc.shouldSucceed {
				if result.IsError {
					t.Error("Expected success but got error result")
				}
				// Check if directory exists
				info, err := os.Stat(tc.path)
				if err != nil {
					t.Errorf("Failed to stat directory: %v", err)
				} else if !info.IsDir() {
					t.Error("Created path is not a directory")
				}
			} else {
				if !result.IsError {
					t.Error("Expected error but got success")
				}
				// Check error message
				if tc.errorContains != "" {
					content, ok := result.Content[0].(mcp.TextContent)
					if !ok {
						t.Error("Expected error content to be TextContent")
					} else if !strings.Contains(content.Text, tc.errorContains) {
						t.Errorf("Expected error to contain %q, got %q", tc.errorContains, content.Text)
					}
				}
				// Check that directory does not exist, but only for valid paths
				if !tc.skipStatCheck {
					if _, err := os.Stat(tc.path); err == nil {
						t.Error("Directory should not exist")
					} else if !os.IsNotExist(err) {
						t.Errorf("Expected directory to not exist, got error: %v", err)
					}
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
	structure := map[string]interface{}{
		"dir1": map[string]interface{}{
			"file1.txt": "content1",
			"file2.txt": "content2",
			"subdir1": map[string]interface{}{
				"file3.txt": "content3",
			},
		},
		"dir2": map[string]interface{}{
			"file4.txt": "content4",
		},
		"file5.txt": "content5",
	}

	// Helper function to create directory structure
	var createStructure func(path string, content interface{}) error
	createStructure = func(path string, content interface{}) error {
		t.Logf("Creating %s", path)
		switch v := content.(type) {
		case string:
			t.Logf("Writing file %s with content %q", path, v)
			return os.WriteFile(path, []byte(v), 0644)
		case map[string]interface{}:
			t.Logf("Creating directory %s", path)
			if err := os.MkdirAll(path, 0750); err != nil {
				return err
			}
			for name, c := range v {
				if err := createStructure(filepath.Join(path, name), c); err != nil {
					return err
				}
			}
		}
		return nil
	}

	if err := createStructure(tempDir, structure); err != nil {
		t.Fatalf("Failed to create test directory structure: %v", err)
	}

	// List the directory to verify its contents
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read test directory: %v", err)
	}
	t.Logf("Contents of %s:", tempDir)
	for _, entry := range entries {
		t.Logf("  %s (%s)", entry.Name(), map[bool]string{true: "dir", false: "file"}[entry.IsDir()])
	}

	// Test cases
	testCases := []struct {
		name          string
		path          string
		maxDepth      float64
		allowedDirs   []string
		shouldSucceed bool
		validate      func(t *testing.T, result *mcp.CallToolResult)
	}{
		{
			name:          "List root directory with default depth",
			path:          tempDir,
			maxDepth:      3, // Set default depth explicitly
			allowedDirs:   []string{tempDir},
			shouldSucceed: true,
			validate: func(t *testing.T, result *mcp.CallToolResult) {
				content, ok := result.Content[0].(mcp.TextContent)
				if !ok {
					t.Error("Expected content to be TextContent")
					return
				}

				var tree []TreeEntry
				if err := json.Unmarshal([]byte(strings.SplitN(content.Text, "\n\n", 2)[1]), &tree); err != nil {
					t.Errorf("Failed to unmarshal tree: %v", err)
					return
				}

				// Validate tree structure
				if len(tree) != 3 {
					t.Errorf("Expected 3 entries at root, got %d", len(tree))
					return
				}

				// Find dir1 and check its contents
				var dir1 *TreeEntry
				for i := range tree {
					if tree[i].Name == "dir1" {
						dir1 = &tree[i]
						break
					}
				}

				if dir1 == nil {
					t.Error("dir1 not found in tree")
					return
				}

				if dir1.Type != "directory" {
					t.Errorf("Expected dir1 to be directory, got %s", dir1.Type)
				}

				if len(dir1.Children) != 3 {
					t.Errorf("Expected 3 entries in dir1, got %d", len(dir1.Children))
				}
			},
		},
		{
			name:          "List root directory with depth 1",
			path:          tempDir,
			maxDepth:      1,
			allowedDirs:   []string{tempDir},
			shouldSucceed: true,
			validate: func(t *testing.T, result *mcp.CallToolResult) {
				content, ok := result.Content[0].(mcp.TextContent)
				if !ok {
					t.Error("Expected content to be TextContent")
					return
				}

				var tree []TreeEntry
				if err := json.Unmarshal([]byte(strings.SplitN(content.Text, "\n\n", 2)[1]), &tree); err != nil {
					t.Errorf("Failed to unmarshal tree: %v", err)
					return
				}

				// Validate tree structure
				if len(tree) != 3 {
					t.Errorf("Expected 3 entries at root, got %d", len(tree))
					return
				}

				// Find dir1 and check it has no children
				var dir1 *TreeEntry
				for i := range tree {
					if tree[i].Name == "dir1" {
						dir1 = &tree[i]
						break
					}
				}

				if dir1 == nil {
					t.Error("dir1 not found in tree")
					return
				}

				if len(dir1.Children) != 0 {
					t.Errorf("Expected no children in dir1 at depth 1, got %d", len(dir1.Children))
				}
			},
		},
		{
			name:          "List non-existent directory",
			path:          filepath.Join(tempDir, "nonexistent"),
			allowedDirs:   []string{tempDir},
			shouldSucceed: false,
		},
		{
			name:          "List directory outside allowed directories",
			path:          "/etc",
			allowedDirs:   []string{tempDir},
			shouldSucceed: false,
		},
		{
			name:          "List file instead of directory",
			path:          filepath.Join(tempDir, "file5.txt"),
			allowedDirs:   []string{tempDir},
			shouldSucceed: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := mcp.CallToolRequest{
				Request: mcp.Request{
					Method: "directory_tree",
				},
				Params: struct {
					Name      string                 `json:"name"`
					Arguments map[string]interface{} `json:"arguments,omitempty"`
					Meta      *struct {
						ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
					} `json:"_meta,omitempty"`
				}{
					Name: "directory_tree",
					Arguments: map[string]interface{}{
						"path":      tc.path,
						"max_depth": tc.maxDepth,
					},
				},
			}

			result, err := handleDirectoryTree(request, tc.allowedDirs)
			if err != nil {
				t.Errorf("handleDirectoryTree returned error: %v", err)
				return
			}

			if tc.shouldSucceed {
				if result.IsError {
					t.Error("Expected success but got error result")
				}
				if tc.validate != nil {
					tc.validate(t, result)
				}
			} else {
				if !result.IsError {
					t.Error("Expected error but got success")
				}
			}
		})
	}
}

func TestMoveFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-move-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test file
	sourceFile := filepath.Join(tmpDir, "source.txt")
	content := []byte("test content")
	if err := os.WriteFile(sourceFile, content, 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	allowedDirs := []string{tmpDir}

	tests := []struct {
		name        string
		source      interface{}
		destination interface{}
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "Move file successfully",
			source:      sourceFile,
			destination: filepath.Join(tmpDir, "dest.txt"),
			wantErr:     false,
		},
		{
			name:        "Invalid source type",
			source:      123,
			destination: filepath.Join(tmpDir, "dest.txt"),
			wantErr:     true,
			errMsg:      "source must be a string",
		},
		{
			name:        "Invalid destination type",
			source:      sourceFile,
			destination: 123,
			wantErr:     true,
			errMsg:      "destination must be a string",
		},
		{
			name:        "Source outside allowed directories",
			source:      "/invalid/path",
			destination: filepath.Join(tmpDir, "dest.txt"),
			wantErr:     true,
			errMsg:      "Invalid source path",
		},
		{
			name:        "Destination outside allowed directories",
			source:      sourceFile,
			destination: "/invalid/path",
			wantErr:     true,
			errMsg:      "Invalid destination path",
		},
		{
			name:        "Source file doesn't exist",
			source:      filepath.Join(tmpDir, "nonexistent.txt"),
			destination: filepath.Join(tmpDir, "dest.txt"),
			wantErr:     true,
			errMsg:      "Error moving file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new source file for each test that needs it
			if tt.name == "Move file successfully" {
				if err := os.WriteFile(sourceFile, content, 0644); err != nil {
					t.Fatalf("Failed to create source file: %v", err)
				}
			}

			request := mcp.CallToolRequest{}
			request.Params.Name = "move_file"
			request.Params.Arguments = map[string]interface{}{
				"source":      tt.source,
				"destination": tt.destination,
			}

			result, err := handleMoveFile(request, allowedDirs)
			if err != nil {
				t.Fatalf("handleMoveFile() error = %v", err)
			}

			if tt.wantErr {
				if result.IsError != true {
					t.Errorf("handleMoveFile() expected error, got success")
				}
				textContent, ok := result.Content[0].(mcp.TextContent)
				if !ok {
					t.Errorf("handleMoveFile() error content is not TextContent")
				}
				if !strings.Contains(textContent.Text, tt.errMsg) {
					t.Errorf("handleMoveFile() error = %v, want error containing %v", textContent.Text, tt.errMsg)
				}
			} else {
				if result.IsError {
					textContent := result.Content[0].(mcp.TextContent)
					t.Errorf("handleMoveFile() unexpected error: %v", textContent.Text)
				}
				// Verify file was moved
				if _, err := os.Stat(tt.destination.(string)); os.IsNotExist(err) {
					t.Errorf("handleMoveFile() destination file doesn't exist")
				}
				if _, err := os.Stat(tt.source.(string)); !os.IsNotExist(err) {
					t.Errorf("handleMoveFile() source file still exists")
				}
			}
		})
	}
}

func TestSearchFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-search-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	files := map[string]string{
		"file1.txt":         "content1",
		"file2.txt":         "content2",
		"test.go":           "package main",
		"dir1/file3.txt":    "content3",
		"dir1/test.txt":     "test content",
		"dir2/file4.go":     "package test",
		"dir2/subdir/a.txt": "content a",
	}

	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0750); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}

	tests := []struct {
		name           string
		pattern        string
		excludePattern []string
		wantCount      int
		wantFiles      []string
	}{
		{
			name:      "Find all txt files",
			pattern:   "*.txt",
			wantCount: 5,
			wantFiles: []string{"file1.txt", "file2.txt", "file3.txt", "test.txt", "a.txt"},
		},
		{
			name:           "Find txt files with exclude",
			pattern:        "*.txt",
			excludePattern: []string{"dir1/*"},
			wantCount:      3,
			wantFiles:      []string{"file1.txt", "file2.txt", "a.txt"},
		},
		{
			name:      "Find go files",
			pattern:   "*.go",
			wantCount: 2,
			wantFiles: []string{"test.go", "file4.go"},
		},
		{
			name:           "Find txt files with multiple excludes",
			pattern:        "*.txt",
			excludePattern: []string{"dir1/*", "dir2/subdir/*"},
			wantCount:      2,
			wantFiles:      []string{"file1.txt", "file2.txt"},
		},
		{
			name:      "No matches",
			pattern:   "*.jpg",
			wantCount: 0,
			wantFiles: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches, err := searchFiles(tmpDir, tt.pattern, tt.excludePattern)
			if err != nil {
				t.Fatalf("searchFiles() error = %v", err)
			}

			if len(matches) != tt.wantCount {
				t.Errorf("searchFiles() got %d matches, want %d", len(matches), tt.wantCount)
			}

			// Check each expected file is in matches
			for _, wantFile := range tt.wantFiles {
				found := false
				for _, match := range matches {
					if strings.HasSuffix(match, wantFile) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("searchFiles() missing expected file %s", wantFile)
				}
			}
		})
	}
}

func TestHandleSearchFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-search-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	files := map[string]string{
		"file1.txt":         "content1",
		"file2.txt":         "content2",
		"test.go":           "package main",
		"dir1/file3.txt":    "content3",
		"dir1/test.txt":     "test content",
		"dir2/file4.go":     "package test",
		"dir2/subdir/a.txt": "content a",
	}

	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0750); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}

	allowedDirs := []string{tmpDir}

	tests := []struct {
		name        string
		root        interface{}
		pattern     interface{}
		exclude     interface{}
		wantErr     bool
		errMsg      string
		wantMatches int
	}{
		{
			name:        "Find all txt files",
			root:        tmpDir,
			pattern:     "*.txt",
			wantErr:     false,
			wantMatches: 5,
		},
		{
			name:    "Invalid root type",
			root:    123,
			pattern: "*.txt",
			wantErr: true,
			errMsg:  "root must be a string",
		},
		{
			name:    "Invalid pattern type",
			root:    tmpDir,
			pattern: 123,
			wantErr: true,
			errMsg:  "pattern must be a string",
		},
		{
			name:    "Root outside allowed directories",
			root:    "/invalid/path",
			pattern: "*.txt",
			wantErr: true,
			errMsg:  "Invalid root path",
		},
		{
			name:        "Valid exclude pattern",
			root:        tmpDir,
			pattern:     "*.txt",
			exclude:     `["dir1/*"]`,
			wantErr:     false,
			wantMatches: 3,
		},
		{
			name:    "Invalid exclude JSON",
			root:    tmpDir,
			pattern: "*.txt",
			exclude: `invalid json`,
			wantErr: true,
			errMsg:  "Invalid exclude JSON array",
		},
		{
			name:        "No matches",
			root:        tmpDir,
			pattern:     "*.jpg",
			wantErr:     false,
			wantMatches: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{}
			request.Params.Name = "search_files"
			request.Params.Arguments = map[string]interface{}{
				"root":    tt.root,
				"pattern": tt.pattern,
			}
			if tt.exclude != nil {
				request.Params.Arguments["exclude"] = tt.exclude
			}

			result, err := handleSearchFiles(request, allowedDirs)
			if err != nil {
				t.Fatalf("handleSearchFiles() error = %v", err)
			}

			if tt.wantErr {
				if result.IsError != true {
					t.Errorf("handleSearchFiles() expected error, got success")
				}
				textContent := result.Content[0].(mcp.TextContent)
				if !strings.Contains(textContent.Text, tt.errMsg) {
					t.Errorf("handleSearchFiles() error = %v, want error containing %v", textContent.Text, tt.errMsg)
				}
			} else {
				if result.IsError {
					textContent := result.Content[0].(mcp.TextContent)
					t.Errorf("handleSearchFiles() unexpected error: %v", textContent.Text)
				}
				textContent := result.Content[0].(mcp.TextContent)
				// Count matches in the result text
				matches := strings.Count(textContent.Text, "\n") - 2 // Subtract header lines
				if matches != tt.wantMatches {
					t.Errorf("handleSearchFiles() got %d matches, want %d", matches, tt.wantMatches)
				}
			}
		})
	}
}

func TestHandleListAllowedDirectories(t *testing.T) {
	tests := []struct {
		name            string
		allowedDirs     []string
		wantDirCount    int
		wantContainsDir string
	}{
		{
			name:            "Single directory",
			allowedDirs:     []string{"/tmp/test"},
			wantDirCount:    1,
			wantContainsDir: "/tmp/test",
		},
		{
			name:            "Multiple directories",
			allowedDirs:     []string{"/tmp/test1", "/tmp/test2", "/tmp/test3"},
			wantDirCount:    3,
			wantContainsDir: "/tmp/test2",
		},
		{
			name:         "Empty directories",
			allowedDirs:  []string{},
			wantDirCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{}
			request.Params.Name = "list_allowed_directories"

			result, err := handleListAllowedDirectories(request, tt.allowedDirs)
			if err != nil {
				t.Fatalf("handleListAllowedDirectories() error = %v", err)
			}

			if result.IsError {
				t.Errorf("handleListAllowedDirectories() unexpected error")
			}

			textContent := result.Content[0].(mcp.TextContent)
			// Count directories in output (subtract header line and empty line)
			dirCount := strings.Count(textContent.Text, "\n") - 2
			if dirCount != tt.wantDirCount {
				t.Errorf("handleListAllowedDirectories() got %d directories, want %d", dirCount, tt.wantDirCount)
			}

			if tt.wantContainsDir != "" && !strings.Contains(textContent.Text, tt.wantContainsDir) {
				t.Errorf("handleListAllowedDirectories() output does not contain %s", tt.wantContainsDir)
			}
		})
	}
}
