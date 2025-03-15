package tools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDirectoryService(t *testing.T) {
	// Test creating a new directory service
	allowedDirs := []string{"/tmp", "/var"}
	service := NewDirectoryService(allowedDirs)

	assert.NotNil(t, service)
	assert.Equal(t, allowedDirs, service.allowedDirs)
	assert.NotNil(t, service.logger)
	assert.NotNil(t, service.validator)
}

func TestDirectoryService_CreateDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-dir-service-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize service with allowed directories
	service := NewDirectoryService([]string{tmpDir})

	// Test cases
	tests := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name:        "Create valid directory",
			path:        filepath.Join(tmpDir, "new-dir"),
			expectError: false,
		},
		{
			name:        "Create nested directory",
			path:        filepath.Join(tmpDir, "parent", "child"),
			expectError: false,
		},
		{
			name:        "Create directory outside allowed path",
			path:        filepath.Join(os.TempDir(), "outside-dir"),
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := service.CreateDirectory(tc.path)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify directory was created
				info, statErr := os.Stat(tc.path)
				assert.NoError(t, statErr)
				assert.True(t, info.IsDir())
			}
		})
	}
}

func TestDirectoryService_ListDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-dir-service-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files and directories
	files := map[string]string{
		"file1.txt": "content1",
		"file2.md":  "content2",
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

	// Initialize service with allowed directories
	service := NewDirectoryService([]string{tmpDir})

	// Test cases
	tests := []struct {
		name          string
		path          string
		expectError   bool
		expectedCount int
	}{
		{
			name:          "List valid directory",
			path:          tmpDir,
			expectError:   false,
			expectedCount: len(files) + len(dirs),
		},
		{
			name:          "List non-existent directory",
			path:          filepath.Join(tmpDir, "non-existent"),
			expectError:   true,
			expectedCount: 0,
		},
		{
			name:          "List outside allowed path",
			path:          os.TempDir(),
			expectError:   true,
			expectedCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := service.ListDirectory(tc.path)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCount, len(result))

				// Verify file entries
				for _, fileInfo := range result {
					if fileInfo.IsDir {
						assert.Contains(t, dirs, fileInfo.Name)
						assert.Empty(t, fileInfo.Extension)
					} else {
						assert.Contains(t, files, fileInfo.Name)
						assert.Equal(t, filepath.Ext(fileInfo.Name), fileInfo.Extension)
						assert.NotEmpty(t, fileInfo.ModTime)
					}
				}
			}
		})
	}
}

func TestDirectoryService_DeleteDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-dir-service-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test directories
	emptyDir := filepath.Join(tmpDir, "empty-dir")
	nonEmptyDir := filepath.Join(tmpDir, "non-empty-dir")
	nestedFile := filepath.Join(nonEmptyDir, "file.txt")

	if err := os.Mkdir(emptyDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.Mkdir(nonEmptyDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(nestedFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Initialize service with allowed directories
	service := NewDirectoryService([]string{tmpDir})

	// Test cases
	tests := []struct {
		name        string
		path        string
		recursive   bool
		expectError bool
	}{
		{
			name:        "Delete empty directory",
			path:        emptyDir,
			recursive:   false,
			expectError: false,
		},
		{
			name:        "Delete non-empty directory without recursive",
			path:        nonEmptyDir,
			recursive:   false,
			expectError: true,
		},
		{
			name:        "Delete non-empty directory with recursive",
			path:        nonEmptyDir,
			recursive:   true,
			expectError: false,
		},
		{
			name:        "Delete non-existent directory",
			path:        filepath.Join(tmpDir, "non-existent"),
			recursive:   false,
			expectError: true,
		},
		{
			name:        "Delete outside allowed path",
			path:        os.TempDir(),
			recursive:   false,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Skip if directory already deleted in previous test
			if _, err := os.Stat(tc.path); os.IsNotExist(err) && !tc.expectError {
				t.Skip("Directory already deleted")
			}

			err := service.DeleteDirectory(tc.path, tc.recursive)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify directory was deleted
				_, statErr := os.Stat(tc.path)
				assert.True(t, os.IsNotExist(statErr))
			}
		})
	}
}
