package tools

import (
	"os"
	"path/filepath"
	"testing"
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
