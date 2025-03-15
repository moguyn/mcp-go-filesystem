package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFileService(t *testing.T) {
	// Test creating a new file service
	allowedDirs := []string{"/tmp", "/var"}
	service := NewFileService(allowedDirs)

	assert.NotNil(t, service)
	assert.Equal(t, allowedDirs, service.allowedDirs)
	assert.NotNil(t, service.logger)
	assert.NotNil(t, service.validator)
}

func TestFileService_ReadFile(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-file-service-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	files := map[string]string{
		"file1.txt": "content1",
		"file2.md":  "content2",
		"empty.txt": "",
	}

	for name, content := range files {
		err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Initialize service with allowed directories
	service := NewFileService([]string{tmpDir})

	// Test cases
	tests := []struct {
		name        string
		path        string
		expectError bool
		expected    string
	}{
		{
			name:        "Read existing file",
			path:        filepath.Join(tmpDir, "file1.txt"),
			expectError: false,
			expected:    "content1",
		},
		{
			name:        "Read empty file",
			path:        filepath.Join(tmpDir, "empty.txt"),
			expectError: false,
			expected:    "",
		},
		{
			name:        "Read non-existent file",
			path:        filepath.Join(tmpDir, "non-existent.txt"),
			expectError: true,
			expected:    "",
		},
		{
			name:        "Read file outside allowed path",
			path:        filepath.Join(os.TempDir(), "outside.txt"),
			expectError: true,
			expected:    "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			content, err := service.ReadFile(tc.path)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, content)
			}
		})
	}
}

func TestFileService_ReadMultipleFiles(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-file-service-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	files := map[string]string{
		"file1.txt": "content1",
		"file2.md":  "content2",
	}

	for name, content := range files {
		err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Initialize service with allowed directories
	service := NewFileService([]string{tmpDir})

	// Test cases
	tests := []struct {
		name          string
		paths         []string
		expectedCount int
		errorCount    int
	}{
		{
			name: "Read multiple existing files",
			paths: []string{
				filepath.Join(tmpDir, "file1.txt"),
				filepath.Join(tmpDir, "file2.md"),
			},
			expectedCount: 2,
			errorCount:    0,
		},
		{
			name: "Read mix of existing and non-existing files",
			paths: []string{
				filepath.Join(tmpDir, "file1.txt"),
				filepath.Join(tmpDir, "non-existent.txt"),
			},
			expectedCount: 2,
			errorCount:    1,
		},
		{
			name: "Read files outside allowed path",
			paths: []string{
				filepath.Join(tmpDir, "file1.txt"),
				filepath.Join(os.TempDir(), "outside.txt"),
			},
			expectedCount: 2,
			errorCount:    1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			results, err := service.ReadMultipleFiles(tc.paths)

			assert.NoError(t, err) // ReadMultipleFiles should not return an error
			assert.Equal(t, tc.expectedCount, len(results))

			errorCount := 0
			for _, result := range results {
				if result.Error != "" {
					errorCount++
				} else {
					// For successful reads, verify content
					for _, path := range tc.paths {
						if path == result.Path {
							filename := filepath.Base(path)
							expectedContent, exists := files[filename]
							if exists {
								assert.Equal(t, expectedContent, result.Content)
							}
						}
					}
				}
			}
			assert.Equal(t, tc.errorCount, errorCount)
		})
	}
}

func TestFileService_WriteFile(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-file-service-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create an existing file
	existingFile := filepath.Join(tmpDir, "existing.txt")
	err = os.WriteFile(existingFile, []byte("initial content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Initialize service with allowed directories
	service := NewFileService([]string{tmpDir})

	// Test cases
	tests := []struct {
		name        string
		path        string
		content     string
		append      bool
		expectError bool
		expected    string
	}{
		{
			name:        "Write to new file",
			path:        filepath.Join(tmpDir, "new.txt"),
			content:     "new content",
			append:      false,
			expectError: false,
			expected:    "new content",
		},
		{
			name:        "Overwrite existing file",
			path:        existingFile,
			content:     "overwritten content",
			append:      false,
			expectError: false,
			expected:    "overwritten content",
		},
		{
			name:        "Append to existing file",
			path:        existingFile,
			content:     " appended content",
			append:      true,
			expectError: false,
			expected:    "overwritten content appended content",
		},
		{
			name:        "Write to nested directory",
			path:        filepath.Join(tmpDir, "nested", "file.txt"),
			content:     "nested content",
			append:      false,
			expectError: false,
			expected:    "nested content",
		},
		{
			name:        "Write outside allowed path",
			path:        filepath.Join(os.TempDir(), "outside.txt"),
			content:     "outside content",
			append:      false,
			expectError: true,
			expected:    "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := service.WriteFile(tc.path, tc.content, tc.append)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify file content
				content, readErr := os.ReadFile(tc.path)
				assert.NoError(t, readErr)
				assert.Equal(t, tc.expected, string(content))
			}
		})
	}
}

func TestFileService_EditFile(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-file-service-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test file with multiple lines
	multilineContent := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5"
	multilineFile := filepath.Join(tmpDir, "multiline.txt")
	err = os.WriteFile(multilineFile, []byte(multilineContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Initialize service with allowed directories
	service := NewFileService([]string{tmpDir})

	// Test cases
	tests := []struct {
		name        string
		path        string
		content     string
		startLine   int
		endLine     int
		expectError bool
		expected    string
	}{
		{
			name:        "Edit middle lines",
			path:        multilineFile,
			content:     "New Line 2\nNew Line 3",
			startLine:   2,
			endLine:     3,
			expectError: false,
			expected:    "Line 1\nNew Line 2\nNew Line 3\nLine 4\nLine 5",
		},
		{
			name:        "Edit first line",
			path:        multilineFile,
			content:     "New Line 1",
			startLine:   1,
			endLine:     1,
			expectError: false,
			expected:    "New Line 1\nNew Line 2\nNew Line 3\nLine 4\nLine 5",
		},
		{
			name:        "Edit last line",
			path:        multilineFile,
			content:     "New Line 5",
			startLine:   5,
			endLine:     5,
			expectError: false,
			expected:    "New Line 1\nNew Line 2\nNew Line 3\nLine 4\nNew Line 5",
		},
		{
			name:        "Edit with invalid line range",
			path:        multilineFile,
			content:     "Invalid",
			startLine:   6,
			endLine:     7,
			expectError: true,
			expected:    "New Line 1\nNew Line 2\nNew Line 3\nLine 4\nNew Line 5",
		},
		{
			name:        "Edit non-existent file",
			path:        filepath.Join(tmpDir, "non-existent.txt"),
			content:     "Content",
			startLine:   1,
			endLine:     1,
			expectError: true,
			expected:    "",
		},
		{
			name:        "Edit outside allowed path",
			path:        filepath.Join(os.TempDir(), "outside.txt"),
			content:     "Outside content",
			startLine:   1,
			endLine:     1,
			expectError: true,
			expected:    "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := service.EditFile(tc.path, tc.content, tc.startLine, tc.endLine)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify file content
				content, readErr := os.ReadFile(tc.path)
				assert.NoError(t, readErr)
				assert.Equal(t, tc.expected, string(content))
			}
		})
	}
}

func TestFileService_DeleteFile(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-file-service-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	existingFile := filepath.Join(tmpDir, "existing.txt")
	err = os.WriteFile(existingFile, []byte("content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Initialize service with allowed directories
	service := NewFileService([]string{tmpDir})

	// Test cases
	tests := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name:        "Delete existing file",
			path:        existingFile,
			expectError: false,
		},
		{
			name:        "Delete non-existent file",
			path:        filepath.Join(tmpDir, "non-existent.txt"),
			expectError: true,
		},
		{
			name:        "Delete outside allowed path",
			path:        filepath.Join(os.TempDir(), "outside.txt"),
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Skip if file already deleted in previous test
			if _, err := os.Stat(tc.path); os.IsNotExist(err) && !tc.expectError {
				t.Skip("File already deleted")
			}

			err := service.DeleteFile(tc.path)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify file was deleted
				_, statErr := os.Stat(tc.path)
				assert.True(t, os.IsNotExist(statErr))
			}
		})
	}
}

func TestFileService_MoveFile(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-file-service-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create subdirectories
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test files
	sourceFile := filepath.Join(tmpDir, "source.txt")
	err = os.WriteFile(sourceFile, []byte("source content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Initialize service with allowed directories
	service := NewFileService([]string{tmpDir})

	// Test cases
	tests := []struct {
		name        string
		sourcePath  string
		destPath    string
		expectError bool
	}{
		{
			name:        "Move file to new location",
			sourcePath:  sourceFile,
			destPath:    filepath.Join(tmpDir, "moved.txt"),
			expectError: false,
		},
		{
			name:        "Move file to subdirectory",
			sourcePath:  filepath.Join(tmpDir, "moved.txt"),
			destPath:    filepath.Join(subDir, "moved.txt"),
			expectError: false,
		},
		{
			name:        "Move non-existent file",
			sourcePath:  filepath.Join(tmpDir, "non-existent.txt"),
			destPath:    filepath.Join(tmpDir, "target.txt"),
			expectError: true,
		},
		{
			name:        "Move outside allowed path",
			sourcePath:  filepath.Join(subDir, "moved.txt"),
			destPath:    filepath.Join(os.TempDir(), "outside.txt"),
			expectError: true,
		},
		{
			name:        "Source outside allowed path",
			sourcePath:  filepath.Join(os.TempDir(), "outside.txt"),
			destPath:    filepath.Join(tmpDir, "inside.txt"),
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Skip if source file doesn't exist and we don't expect an error
			if _, err := os.Stat(tc.sourcePath); os.IsNotExist(err) && !tc.expectError {
				t.Skip("Source file doesn't exist")
			}

			err := service.MoveFile(tc.sourcePath, tc.destPath)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify source file no longer exists
				_, sourceErr := os.Stat(tc.sourcePath)
				assert.True(t, os.IsNotExist(sourceErr))

				// Verify destination file exists
				destInfo, destErr := os.Stat(tc.destPath)
				assert.NoError(t, destErr)
				assert.False(t, destInfo.IsDir())

				// Verify content
				content, readErr := os.ReadFile(tc.destPath)
				assert.NoError(t, readErr)
				assert.Equal(t, "source content", string(content))
			}
		})
	}
}

func TestFileService_CopyFile(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-file-service-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create subdirectories
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test files
	sourceFile := filepath.Join(tmpDir, "source.txt")
	err = os.WriteFile(sourceFile, []byte("source content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Initialize service with allowed directories
	service := NewFileService([]string{tmpDir})

	// Test cases
	tests := []struct {
		name        string
		sourcePath  string
		destPath    string
		expectError bool
	}{
		{
			name:        "Copy file to new location",
			sourcePath:  sourceFile,
			destPath:    filepath.Join(tmpDir, "copied.txt"),
			expectError: false,
		},
		{
			name:        "Copy file to subdirectory",
			sourcePath:  sourceFile,
			destPath:    filepath.Join(subDir, "copied.txt"),
			expectError: false,
		},
		{
			name:        "Copy non-existent file",
			sourcePath:  filepath.Join(tmpDir, "non-existent.txt"),
			destPath:    filepath.Join(tmpDir, "target.txt"),
			expectError: true,
		},
		{
			name:        "Copy outside allowed path",
			sourcePath:  sourceFile,
			destPath:    filepath.Join(os.TempDir(), "outside.txt"),
			expectError: true,
		},
		{
			name:        "Source outside allowed path",
			sourcePath:  filepath.Join(os.TempDir(), "outside.txt"),
			destPath:    filepath.Join(tmpDir, "inside.txt"),
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := service.CopyFile(tc.sourcePath, tc.destPath)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify source file still exists
				_, sourceErr := os.Stat(tc.sourcePath)
				assert.NoError(t, sourceErr)

				// Verify destination file exists
				destInfo, destErr := os.Stat(tc.destPath)
				assert.NoError(t, destErr)
				assert.False(t, destInfo.IsDir())

				// Verify content
				content, readErr := os.ReadFile(tc.destPath)
				assert.NoError(t, readErr)
				assert.Equal(t, "source content", string(content))
			}
		})
	}
}

func TestFileService_ValidatePath(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-file-service-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize service with allowed directories
	service := NewFileService([]string{tmpDir})

	// Test cases
	tests := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name:        "Valid path inside allowed directory",
			path:        filepath.Join(tmpDir, "file.txt"),
			expectError: false,
		},
		{
			name:        "Valid nested path",
			path:        filepath.Join(tmpDir, "subdir", "file.txt"),
			expectError: false,
		},
		{
			name:        "Path outside allowed directories",
			path:        filepath.Join(os.TempDir(), "outside.txt"),
			expectError: true,
		},
		{
			name:        "Path with directory traversal",
			path:        filepath.Join(tmpDir, "..", "traversal.txt"),
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			validPath, err := service.validator.ValidatePath(tc.path)

			if tc.expectError {
				assert.Error(t, err)
				assert.Empty(t, validPath)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, validPath)
				// Ensure the path is absolute
				assert.True(t, filepath.IsAbs(validPath))
				// Ensure the path is within the allowed directory
				assert.True(t, strings.HasPrefix(validPath, tmpDir))
			}
		})
	}
}
