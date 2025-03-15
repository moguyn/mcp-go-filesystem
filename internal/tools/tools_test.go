package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// Mock types for testing
type mockToolParameters struct {
	Path    string `json:"path"`
	Content string `json:"content,omitempty"`
	Append  bool   `json:"append,omitempty"`
}

func newMockRequest(params mockToolParameters) mcp.CallToolRequest {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = make(map[string]interface{})

	// Convert mockToolParameters to map[string]interface{}
	if params.Path != "" {
		req.Params.Arguments["path"] = params.Path
	}
	if params.Content != "" {
		req.Params.Arguments["content"] = params.Content
	}
	if params.Append {
		req.Params.Arguments["append"] = params.Append
	}

	return req
}

// Helper function to convert MCP result to testable format
func extractResult(result interface{}) (string, string, interface{}, error) {
	if mcpResult, ok := result.(*mcp.CallToolResult); ok {
		var message, text string
		var data interface{}

		if mcpResult.IsError {
			for _, content := range mcpResult.Content {
				if textContent, ok := content.(mcp.TextContent); ok {
					if textContent.Type == "text" {
						message = textContent.Text
						break
					}
				}
			}
		} else {
			for _, content := range mcpResult.Content {
				if textContent, ok := content.(mcp.TextContent); ok {
					if textContent.Type == "text" {
						text = textContent.Text
						break
					}
				}
			}
		}

		return message, text, data, nil
	}
	return "", "", nil, fmt.Errorf("unexpected result type: %T", result)
}

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

func TestExpandHome(t *testing.T) {
	usr, err := user.Current()
	if err != nil {
		t.Fatalf("Failed to get current user: %v", err)
	}

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "No tilde",
			path:     "/absolute/path",
			expected: "/absolute/path",
		},
		{
			name:     "Just tilde",
			path:     "~",
			expected: usr.HomeDir,
		},
		{
			name:     "Tilde with path",
			path:     "~/documents",
			expected: filepath.Join(usr.HomeDir, "documents"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ExpandHome(tc.path)
			if result != tc.expected {
				t.Errorf("ExpandHome(%q) = %q, want %q", tc.path, result, tc.expected)
			}
		})
	}
}

func TestFileInfo(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "fileinfo-test-")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write some content
	content := []byte("test content")
	if _, err := tempFile.Write(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	// Get file info
	info, err := os.Stat(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to get file info: %v", err)
	}

	fileInfo := FileInfo{
		Size:        info.Size(),
		Created:     time.Now().Format(time.RFC3339),
		Modified:    info.ModTime().Format(time.RFC3339),
		Accessed:    time.Now().Format(time.RFC3339),
		IsDirectory: info.IsDir(),
		IsFile:      !info.IsDir(),
		Permissions: info.Mode().String(),
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(fileInfo)
	if err != nil {
		t.Fatalf("Failed to marshal FileInfo: %v", err)
	}

	var unmarshaledInfo FileInfo
	if err := json.Unmarshal(jsonData, &unmarshaledInfo); err != nil {
		t.Fatalf("Failed to unmarshal FileInfo: %v", err)
	}

	// Verify fields
	if unmarshaledInfo.Size != fileInfo.Size {
		t.Errorf("Size mismatch: got %d, want %d", unmarshaledInfo.Size, fileInfo.Size)
	}
	if unmarshaledInfo.IsDirectory != fileInfo.IsDirectory {
		t.Errorf("IsDirectory mismatch: got %v, want %v", unmarshaledInfo.IsDirectory, fileInfo.IsDirectory)
	}
	if unmarshaledInfo.IsFile != fileInfo.IsFile {
		t.Errorf("IsFile mismatch: got %v, want %v", unmarshaledInfo.IsFile, fileInfo.IsFile)
	}
}

func TestTreeEntry(t *testing.T) {
	// Create test directory structure
	tempDir, err := os.MkdirTemp("", "tree-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a subdirectory and some files
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	files := []string{"file1.txt", "file2.txt"}
	for _, file := range files {
		path := filepath.Join(tempDir, file)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Create tree structure
	tree := TreeEntry{
		Name: filepath.Base(tempDir),
		Type: "directory",
		Children: []TreeEntry{
			{
				Name: "subdir",
				Type: "directory",
			},
			{
				Name: "file1.txt",
				Type: "file",
			},
			{
				Name: "file2.txt",
				Type: "file",
			},
		},
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(tree)
	if err != nil {
		t.Fatalf("Failed to marshal TreeEntry: %v", err)
	}

	var unmarshaledTree TreeEntry
	if err := json.Unmarshal(jsonData, &unmarshaledTree); err != nil {
		t.Fatalf("Failed to unmarshal TreeEntry: %v", err)
	}

	// Verify structure
	if unmarshaledTree.Name != tree.Name {
		t.Errorf("Name mismatch: got %s, want %s", unmarshaledTree.Name, tree.Name)
	}
	if unmarshaledTree.Type != tree.Type {
		t.Errorf("Type mismatch: got %s, want %s", unmarshaledTree.Type, tree.Type)
	}
	if len(unmarshaledTree.Children) != len(tree.Children) {
		t.Errorf("Children length mismatch: got %d, want %d", len(unmarshaledTree.Children), len(tree.Children))
	}
}

func TestBuildDirectoryTree(t *testing.T) {
	// Create test directory structure
	tempDir, err := os.MkdirTemp("", "dirtree-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create nested structure
	dirs := []string{
		filepath.Join(tempDir, "dir1"),
		filepath.Join(tempDir, "dir1", "subdir1"),
		filepath.Join(tempDir, "dir2"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
	}

	files := []string{
		filepath.Join(tempDir, "file1.txt"),
		filepath.Join(tempDir, "dir1", "file2.txt"),
		filepath.Join(tempDir, "dir1", "subdir1", "file3.txt"),
	}

	for _, file := range files {
		if err := os.WriteFile(file, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}

	// Test with different max depths
	tests := []struct {
		name     string
		maxDepth int
		wantLen  int // Expected number of top-level entries
	}{
		{
			name:     "Depth 1",
			maxDepth: 1,
			wantLen:  3, // dir1, dir2, file1.txt
		},
		{
			name:     "Depth 2",
			maxDepth: 2,
			wantLen:  3,
		},
		{
			name:     "Depth 3",
			maxDepth: 3,
			wantLen:  3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := buildDirectoryTree(tempDir, tc.maxDepth)
			if err != nil {
				t.Fatalf("buildDirectoryTree failed: %v", err)
			}

			if len(tree) != tc.wantLen {
				t.Errorf("got %d top-level entries, want %d", len(tree), tc.wantLen)
			}

			// Verify structure based on depth
			for _, entry := range tree {
				verifyTreeDepth(t, entry, tc.maxDepth, 1)
			}
		})
	}
}

func verifyTreeDepth(t *testing.T, entry TreeEntry, maxDepth, currentDepth int) {
	t.Helper()

	if entry.Type != "file" && entry.Type != "directory" {
		t.Errorf("Invalid entry type: %s", entry.Type)
	}

	if currentDepth < maxDepth && entry.Type == "directory" {
		if len(entry.Children) == 0 && entry.Name != "dir2" { // dir2 is empty
			t.Errorf("Directory %s at depth %d has no children", entry.Name, currentDepth)
		}

		for _, child := range entry.Children {
			verifyTreeDepth(t, child, maxDepth, currentDepth+1)
		}
	}
}

func TestHandleReadFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "handler-test-")
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

	tests := []struct {
		name          string
		path          string
		allowedDirs   []string
		expectError   bool
		expectedError string
	}{
		{
			name:        "Valid file in allowed directory",
			path:        testFile,
			allowedDirs: []string{tempDir},
			expectError: false,
		},
		{
			name:          "File outside allowed directory",
			path:          "/etc/passwd",
			allowedDirs:   []string{tempDir},
			expectError:   true,
			expectedError: "Invalid path: permission denied",
		},
		{
			name:          "Non-existent file",
			path:          filepath.Join(tempDir, "nonexistent.txt"),
			allowedDirs:   []string{tempDir},
			expectError:   true,
			expectedError: "Error reading file:",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := newMockRequest(mockToolParameters{
				Path: tc.path,
			})

			result, err := handleReadFile(request, tc.allowedDirs)
			if err != nil {
				t.Fatalf("handleReadFile returned unexpected error: %v", err)
			}

			message, text, _, err := extractResult(result)
			if err != nil {
				t.Fatalf("Failed to extract result: %v", err)
			}

			if tc.expectError {
				if message == "" {
					t.Error("Expected an error message but got none")
				} else if !strings.Contains(message, tc.expectedError) {
					t.Errorf("Expected error containing %q, got %q", tc.expectedError, message)
				}
			} else {
				if message != "" {
					t.Errorf("Unexpected error message: %v", message)
				}
				if text != testContent {
					t.Errorf("Expected content %q, got %q", testContent, text)
				}
			}
		})
	}
}

func TestHandleWriteFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "handler-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "test content"

	tests := []struct {
		name          string
		path          string
		content       string
		append        bool
		allowedDirs   []string
		expectError   bool
		expectedError string
	}{
		{
			name:        "Write new file",
			path:        testFile,
			content:     testContent,
			append:      false,
			allowedDirs: []string{tempDir},
			expectError: false,
		},
		{
			name:        "Append to existing file",
			path:        testFile,
			content:     " additional content",
			append:      true,
			allowedDirs: []string{tempDir},
			expectError: false,
		},
		{
			name:          "Write outside allowed directory",
			path:          "/etc/test.txt",
			content:       testContent,
			append:        false,
			allowedDirs:   []string{tempDir},
			expectError:   true,
			expectedError: "Invalid path: permission denied",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := newMockRequest(mockToolParameters{
				Path:    tc.path,
				Content: tc.content,
				Append:  tc.append,
			})

			result, err := handleWriteFile(request, tc.allowedDirs)
			if err != nil {
				t.Fatalf("handleWriteFile returned unexpected error: %v", err)
			}

			message, _, _, err := extractResult(result)
			if err != nil {
				t.Fatalf("Failed to extract result: %v", err)
			}

			if tc.expectError {
				if message == "" {
					t.Error("Expected an error message but got none")
				} else if !strings.Contains(message, tc.expectedError) {
					t.Errorf("Expected error containing %q, got %q", tc.expectedError, message)
				}
			} else {
				if message != "" {
					t.Errorf("Unexpected error message: %v", message)
				}

				// Verify file contents
				content, err := os.ReadFile(tc.path)
				if err != nil {
					t.Fatalf("Failed to read test file: %v", err)
				}

				expectedContent := tc.content
				if tc.append {
					expectedContent = testContent + tc.content
				}

				if string(content) != expectedContent {
					t.Errorf("Expected content %q, got %q", expectedContent, string(content))
				}
			}
		})
	}
}

func TestHandleCreateDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "handler-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name          string
		path          string
		allowedDirs   []string
		expectError   bool
		expectedError string
	}{
		{
			name:        "Create new directory",
			path:        filepath.Join(tempDir, "newdir"),
			allowedDirs: []string{tempDir},
			expectError: false,
		},
		{
			name:        "Create nested directory",
			path:        filepath.Join(tempDir, "parent", "child"),
			allowedDirs: []string{tempDir},
			expectError: false,
		},
		{
			name:          "Create outside allowed directory",
			path:          "/etc/newdir",
			allowedDirs:   []string{tempDir},
			expectError:   true,
			expectedError: "Invalid path: permission denied",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := newMockRequest(mockToolParameters{
				Path: tc.path,
			})

			result, err := handleCreateDirectory(request, tc.allowedDirs)
			if err != nil {
				t.Fatalf("handleCreateDirectory returned unexpected error: %v", err)
			}

			message, _, _, err := extractResult(result)
			if err != nil {
				t.Fatalf("Failed to extract result: %v", err)
			}

			if tc.expectError {
				if message == "" {
					t.Error("Expected an error message but got none")
				} else if !strings.Contains(message, tc.expectedError) {
					t.Errorf("Expected error containing %q, got %q", tc.expectedError, message)
				}
			} else {
				if message != "" {
					t.Errorf("Unexpected error message: %v", message)
				}

				// Verify directory exists
				info, err := os.Stat(tc.path)
				if err != nil {
					t.Errorf("Failed to stat created directory: %v", err)
				} else if !info.IsDir() {
					t.Error("Created path is not a directory")
				}
			}
		})
	}
}

func TestHandleListDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "list-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files and directories
	testFiles := []string{"file1.txt", "file2.txt"}
	testDirs := []string{"dir1", "dir2", "dir1/subdir"}

	for _, dir := range testDirs {
		if err := os.MkdirAll(filepath.Join(tempDir, dir), 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}
	}

	for _, file := range testFiles {
		if err := os.WriteFile(filepath.Join(tempDir, file), []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	tests := []struct {
		name          string
		path          string
		allowedDirs   []string
		expectError   bool
		expectedError string
		expectedCount int
	}{
		{
			name:          "List valid directory",
			path:          tempDir,
			allowedDirs:   []string{tempDir},
			expectError:   false,
			expectedCount: 4, // 2 files + 2 directories
		},
		{
			name:          "List subdirectory",
			path:          filepath.Join(tempDir, "dir1"),
			allowedDirs:   []string{tempDir},
			expectError:   false,
			expectedCount: 1, // 1 subdirectory
		},
		{
			name:          "List outside allowed directory",
			path:          "/etc",
			allowedDirs:   []string{tempDir},
			expectError:   true,
			expectedError: "Invalid path: permission denied",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := newMockRequest(mockToolParameters{
				Path: tc.path,
			})

			result, err := handleListDirectory(request, tc.allowedDirs)
			if err != nil {
				t.Fatalf("handleListDirectory returned unexpected error: %v", err)
			}

			message, text, _, err := extractResult(result)
			if err != nil {
				t.Fatalf("Failed to extract result: %v", err)
			}

			if tc.expectError {
				if message == "" {
					t.Error("Expected an error message but got none")
				} else if !strings.Contains(message, tc.expectedError) {
					t.Errorf("Expected error containing %q, got %q", tc.expectedError, message)
				}
			} else {
				if message != "" {
					t.Errorf("Unexpected error message: %v", message)
				}

				var entries []FileInfo
				if err := json.Unmarshal([]byte(text), &entries); err != nil {
					t.Fatalf("Failed to unmarshal result: %v", err)
				}

				if len(entries) != tc.expectedCount {
					t.Errorf("Expected %d entries, got %d", tc.expectedCount, len(entries))
				}
			}
		})
	}
}

func TestHandleSearchFiles(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "search-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files with different extensions
	files := map[string]string{
		"test1.txt":        "content1",
		"test2.txt":        "content2",
		"dir1/test3.txt":   "content3",
		"dir1/test.go":     "package main",
		"dir2/test.py":     "print('hello')",
		"dir2/subdir/file": "data",
	}

	for path, content := range files {
		fullPath := filepath.Join(tempDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}

	tests := []struct {
		name          string
		root          string
		pattern       string
		exclude       []string
		allowedDirs   []string
		expectError   bool
		expectedError string
		expectedCount int
	}{
		{
			name:          "Search txt files",
			root:          tempDir,
			pattern:       "*.txt",
			allowedDirs:   []string{tempDir},
			expectError:   false,
			expectedCount: 3,
		},
		{
			name:          "Search with exclude",
			root:          tempDir,
			pattern:       "*.*",
			exclude:       []string{"*.txt"},
			allowedDirs:   []string{tempDir},
			expectError:   false,
			expectedCount: 2, // .go and .py files
		},
		{
			name:          "Search outside allowed directory",
			root:          "/etc",
			pattern:       "*",
			allowedDirs:   []string{tempDir},
			expectError:   true,
			expectedError: "permission denied",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Validate path
			validPath, err := ValidatePath(tc.root, tc.allowedDirs)
			if err != nil {
				if !tc.expectError {
					t.Errorf("Unexpected error: %v", err)
				} else if !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("Expected error containing %q, got %v", tc.expectedError, err)
				}
				return
			}

			// Search files
			var matches []string
			err = filepath.Walk(validPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					return nil
				}
				match, err := filepath.Match(tc.pattern, filepath.Base(path))
				if err != nil {
					return err
				}
				if match {
					// Check excludes
					excluded := false
					for _, excl := range tc.exclude {
						match, err := filepath.Match(excl, filepath.Base(path))
						if err != nil {
							return err
						}
						if match {
							excluded = true
							break
						}
					}
					if !excluded {
						matches = append(matches, path)
					}
				}
				return nil
			})

			if err != nil {
				if !tc.expectError {
					t.Fatalf("Failed to search files: %v", err)
				}
				return
			}

			if tc.expectError {
				t.Error("Expected an error but got none")
			} else if len(matches) != tc.expectedCount {
				t.Errorf("Expected %d matches, got %d: %v", tc.expectedCount, len(matches), matches)
			}
		})
	}
}

func TestHandleGetFileInfo(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "fileinfo-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files and directories
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "test content"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	testDir := filepath.Join(tempDir, "testdir")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	tests := []struct {
		name          string
		path          string
		allowedDirs   []string
		expectError   bool
		expectedError string
		isDir         bool
	}{
		{
			name:        "Get file info",
			path:        testFile,
			allowedDirs: []string{tempDir},
			expectError: false,
			isDir:       false,
		},
		{
			name:        "Get directory info",
			path:        testDir,
			allowedDirs: []string{tempDir},
			expectError: false,
			isDir:       true,
		},
		{
			name:          "Get info outside allowed directory",
			path:          "/etc/passwd",
			allowedDirs:   []string{tempDir},
			expectError:   true,
			expectedError: "permission denied",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Validate path
			validPath, err := ValidatePath(tc.path, tc.allowedDirs)
			if err != nil {
				if !tc.expectError {
					t.Errorf("Unexpected error: %v", err)
				} else if !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("Expected error containing %q, got %v", tc.expectedError, err)
				}
				return
			}

			// Get file info
			info, err := os.Stat(validPath)
			if err != nil {
				if !tc.expectError {
					t.Fatalf("Failed to get file info: %v", err)
				}
				return
			}

			if tc.expectError {
				t.Error("Expected an error but got none")
			} else {
				if info.IsDir() != tc.isDir {
					t.Errorf("Expected IsDir() = %v, got %v", tc.isDir, info.IsDir())
				}
				if !tc.isDir && info.Size() != int64(len(testContent)) {
					t.Errorf("Expected size %d, got %d", len(testContent), info.Size())
				}
			}
		})
	}
}

func TestHandleMoveFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "move-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files and directories
	testFile := filepath.Join(tempDir, "source.txt")
	testContent := "test content"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name          string
		source        string
		dest          string
		allowedDirs   []string
		expectError   bool
		expectedError string
	}{
		{
			name:        "Move file in same directory",
			source:      testFile,
			dest:        filepath.Join(tempDir, "dest.txt"),
			allowedDirs: []string{tempDir},
			expectError: false,
		},
		{
			name:        "Move file to subdirectory",
			source:      testFile,
			dest:        filepath.Join(tempDir, "subdir", "dest.txt"),
			allowedDirs: []string{tempDir},
			expectError: false,
		},
		{
			name:          "Move outside allowed directory",
			source:        testFile,
			dest:          "/etc/test.txt",
			allowedDirs:   []string{tempDir},
			expectError:   true,
			expectedError: "permission denied",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create source file if it doesn't exist
			if _, err := os.Stat(tc.source); os.IsNotExist(err) {
				if err := os.WriteFile(tc.source, []byte(testContent), 0644); err != nil {
					t.Fatalf("Failed to create source file: %v", err)
				}
			}

			// Validate source path
			validSource, err := ValidatePath(tc.source, tc.allowedDirs)
			if err != nil {
				if !tc.expectError {
					t.Errorf("Unexpected source validation error: %v", err)
				}
				return
			}

			// Validate destination path
			validDest, err := ValidatePath(tc.dest, tc.allowedDirs)
			if err != nil {
				if !tc.expectError {
					t.Errorf("Unexpected destination validation error: %v", err)
				} else if !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("Expected error containing %q, got %v", tc.expectedError, err)
				}
				return
			}

			// Create parent directory for destination if needed
			if err := os.MkdirAll(filepath.Dir(validDest), 0755); err != nil {
				t.Fatalf("Failed to create destination directory: %v", err)
			}

			// Move file
			if err := os.Rename(validSource, validDest); err != nil {
				if !tc.expectError {
					t.Fatalf("Failed to move file: %v", err)
				}
				return
			}

			if tc.expectError {
				t.Error("Expected an error but got none")
			} else {
				// Verify file was moved
				if _, err := os.Stat(validSource); !os.IsNotExist(err) {
					t.Error("Source file still exists")
				}
				content, err := os.ReadFile(validDest)
				if err != nil {
					t.Fatalf("Failed to read destination file: %v", err)
				}
				if string(content) != testContent {
					t.Errorf("Expected content %q, got %q", testContent, string(content))
				}
			}
		})
	}
}

func TestHandleReadMultipleFiles(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "readmulti-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	files := map[string]string{
		"file1.txt":     "content1",
		"file2.txt":     "content2",
		"dir/file3.txt": "content3",
	}

	for path, content := range files {
		fullPath := filepath.Join(tempDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}

	tests := []struct {
		name          string
		paths         []string
		allowedDirs   []string
		expectError   bool
		expectedError string
		expectedCount int
	}{
		{
			name: "Read multiple files",
			paths: []string{
				filepath.Join(tempDir, "file1.txt"),
				filepath.Join(tempDir, "file2.txt"),
			},
			allowedDirs:   []string{tempDir},
			expectError:   false,
			expectedCount: 2,
		},
		{
			name: "Read nested file",
			paths: []string{
				filepath.Join(tempDir, "dir", "file3.txt"),
			},
			allowedDirs:   []string{tempDir},
			expectError:   false,
			expectedCount: 1,
		},
		{
			name:          "Read outside allowed directory",
			paths:         []string{"/etc/passwd"},
			allowedDirs:   []string{tempDir},
			expectError:   true,
			expectedError: "permission denied",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var contents []string
			for _, path := range tc.paths {
				// Validate path
				validPath, err := ValidatePath(path, tc.allowedDirs)
				if err != nil {
					if !tc.expectError {
						t.Errorf("Unexpected error: %v", err)
					} else if !strings.Contains(err.Error(), tc.expectedError) {
						t.Errorf("Expected error containing %q, got %v", tc.expectedError, err)
					}
					return
				}

				// Read file
				content, err := os.ReadFile(validPath)
				if err != nil {
					if !tc.expectError {
						t.Fatalf("Failed to read file: %v", err)
					}
					return
				}
				contents = append(contents, string(content))
			}

			if tc.expectError {
				t.Error("Expected an error but got none")
			} else if len(contents) != tc.expectedCount {
				t.Errorf("Expected %d files, got %d", tc.expectedCount, len(contents))
			}
		})
	}
}

func TestHandleEditFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "edit-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test file
	testFile := filepath.Join(tempDir, "test.txt")
	originalContent := "Hello, World!\nThis is a test file.\n"
	if err := os.WriteFile(testFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name            string
		path            string
		oldText         string
		newText         string
		allowedDirs     []string
		expectError     bool
		expectedError   string
		expectedContent string
	}{
		{
			name:            "Replace text",
			path:            testFile,
			oldText:         "World",
			newText:         "Universe",
			allowedDirs:     []string{tempDir},
			expectError:     false,
			expectedContent: "Hello, Universe!\nThis is a test file.\n",
		},
		{
			name:            "Replace multiple occurrences",
			path:            testFile,
			oldText:         "is",
			newText:         "was",
			allowedDirs:     []string{tempDir},
			expectError:     false,
			expectedContent: "Hello, World!\nThwas was a test file.\n",
		},
		{
			name:          "Edit outside allowed directory",
			path:          "/etc/passwd",
			oldText:       "test",
			newText:       "replaced",
			allowedDirs:   []string{tempDir},
			expectError:   true,
			expectedError: "permission denied",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Reset test file content
			if err := os.WriteFile(testFile, []byte(originalContent), 0644); err != nil {
				t.Fatalf("Failed to reset test file: %v", err)
			}

			// Validate path
			validPath, err := ValidatePath(tc.path, tc.allowedDirs)
			if err != nil {
				if !tc.expectError {
					t.Errorf("Unexpected error: %v", err)
				} else if !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("Expected error containing %q, got %v", tc.expectedError, err)
				}
				return
			}

			// Read file
			content, err := os.ReadFile(validPath)
			if err != nil {
				if !tc.expectError {
					t.Fatalf("Failed to read file: %v", err)
				}
				return
			}

			// Replace text
			newContent := strings.ReplaceAll(string(content), tc.oldText, tc.newText)

			// Write back
			if err := os.WriteFile(validPath, []byte(newContent), 0644); err != nil {
				if !tc.expectError {
					t.Fatalf("Failed to write file: %v", err)
				}
				return
			}

			if tc.expectError {
				t.Error("Expected an error but got none")
			} else {
				// Verify content
				content, err := os.ReadFile(validPath)
				if err != nil {
					t.Fatalf("Failed to read file after edit: %v", err)
				}
				if string(content) != tc.expectedContent {
					t.Errorf("Expected content %q, got %q", tc.expectedContent, string(content))
				}
			}
		})
	}
}

func TestHandleListAllowedDirectories(t *testing.T) {
	// Create temporary directories for testing
	dirs := make([]string, 3)
	for i := range dirs {
		dir, err := os.MkdirTemp("", "allowed-test-")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(dir)
		dirs[i] = dir
	}

	tests := []struct {
		name        string
		allowedDirs []string
		expectCount int
	}{
		{
			name:        "List single directory",
			allowedDirs: []string{dirs[0]},
			expectCount: 1,
		},
		{
			name:        "List multiple directories",
			allowedDirs: dirs,
			expectCount: 3,
		},
		{
			name:        "List empty directories",
			allowedDirs: []string{},
			expectCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.allowedDirs) != tc.expectCount {
				t.Errorf("Expected %d directories, got %d", tc.expectCount, len(tc.allowedDirs))
			}

			for _, dir := range tc.allowedDirs {
				info, err := os.Stat(dir)
				if err != nil {
					t.Errorf("Failed to stat directory %s: %v", dir, err)
				} else if !info.IsDir() {
					t.Errorf("Path %s is not a directory", dir)
				}
			}
		})
	}
}
