package tools

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// ExpandHome expands the tilde (~) in a path to the user's home directory
func ExpandHome(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	usr, err := user.Current()
	if err != nil {
		return path
	}

	if path == "~" {
		return usr.HomeDir
	}

	return filepath.Join(usr.HomeDir, path[2:])
}

// ValidatePath checks if a path is within the allowed directories
func ValidatePath(requestedPath string, allowedDirectories []string) (string, error) {
	// Normalize the path
	expandedPath := ExpandHome(requestedPath)
	absPath, err := filepath.Abs(expandedPath)
	if err != nil {
		return "", err
	}
	normalizedPath := filepath.Clean(absPath)

	// Check if the path is within any of the allowed directories
	for _, allowedDir := range allowedDirectories {
		// Normalize the allowed directory
		allowedDirAbs, err := filepath.Abs(allowedDir)
		if err != nil {
			continue
		}
		allowedDirNormalized := filepath.Clean(allowedDirAbs)

		// Check if the path is the allowed directory or a subdirectory
		if normalizedPath == allowedDirNormalized || strings.HasPrefix(normalizedPath, allowedDirNormalized+string(filepath.Separator)) {
			return normalizedPath, nil
		}
	}

	return "", os.ErrPermission
}
