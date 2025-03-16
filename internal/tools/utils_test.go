package tools

import (
	"os/user"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpandHome(t *testing.T) {
	usr, err := user.Current()
	if err != nil {
		t.Fatalf("Failed to get current user: %v", err)
	}

	tests := []struct {
		name     string
		path     string
		wantPath string
	}{
		{
			name:     "No tilde",
			path:     "/absolute/path",
			wantPath: "/absolute/path",
		},
		{
			name:     "Just tilde",
			path:     "~",
			wantPath: usr.HomeDir,
		},
		{
			name:     "Tilde with path",
			path:     "~/Documents",
			wantPath: filepath.Join(usr.HomeDir, "Documents"),
		},
		{
			name:     "Tilde with nested path",
			path:     "~/Documents/folder/file.txt",
			wantPath: filepath.Join(usr.HomeDir, "Documents/folder/file.txt"),
		},
		{
			name:     "Relative path",
			path:     "relative/path",
			wantPath: "relative/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExpandHome(tt.path)
			if got != tt.wantPath {
				t.Errorf("ExpandHome() = %v, want %v", got, tt.wantPath)
			}
		})
	}
}

func TestValidatePath(t *testing.T) {
	// Define test allowed directories
	allowedDirs := []string{"/tmp", "/var", "/home/user"}

	tests := []struct {
		name        string
		path        string
		expectError bool
		errorCheck  func(t *testing.T, err error)
	}{
		{
			name:        "Valid path inside allowed directory",
			path:        "/tmp/file.txt",
			expectError: false,
		},
		{
			name:        "Path outside allowed directories",
			path:        "/etc/hosts",
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				// Check that the error message contains all allowed directories
				errMsg := err.Error()
				assert.Contains(t, errMsg, "path not allowed: /etc/hosts")
				for _, dir := range allowedDirs {
					assert.Contains(t, errMsg, dir)
				}
			},
		},
		{
			name:        "Path with invalid characters",
			path:        string([]byte{0}) + "/tmp/file.txt",
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid characters")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			validPath, err := ValidatePath(tc.path, allowedDirs)

			if tc.expectError {
				assert.Error(t, err)
				assert.Empty(t, validPath)
				if tc.errorCheck != nil {
					tc.errorCheck(t, err)
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, validPath)
			}
		})
	}
}
