package tools

import (
	"os/user"
	"path/filepath"
	"testing"
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
