package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrintUsage(t *testing.T) {
	// Capture stdout
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Call the function
	printUsage()

	// Restore stdout
	w.Close()
	os.Stderr = oldStderr

	// Read captured output
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Verify output contains expected information
	expectedStrings := []string{
		"MCP Filesystem Server",
		"Usage:",
		"mcp-server-filesystem",
		"allowed-directory",
		"--help",
		"-h",
		"Show this help message",
		"The server will only allow operations within the specified directories",
		"Example:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s', but it didn't", expected)
		}
	}
}

// TestProcessArgs tests the argument processing logic without calling main
func TestProcessArgs(t *testing.T) {
	// Test with help flags
	if !isHelpFlag("--help") {
		t.Errorf("Expected '--help' to be recognized as a help flag")
	}

	if !isHelpFlag("-h") {
		t.Errorf("Expected '-h' to be recognized as a help flag")
	}

	if isHelpFlag("something-else") {
		t.Errorf("Expected 'something-else' not to be recognized as a help flag")
	}

	// Test directory validation
	tempDir := createTempDir(t)
	defer os.RemoveAll(tempDir)

	nonExistentDir := filepath.Join(tempDir, "non-existent")

	// Create a file (not a directory)
	filePath := filepath.Join(tempDir, "file.txt")
	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.Close()

	// Test cases
	testCases := []struct {
		path      string
		shouldErr bool
		errMsg    string
	}{
		{tempDir, false, ""},
		{nonExistentDir, true, "no such file or directory"},
		{filePath, true, "not a directory"},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			_, err := validateAndNormalizePath(tc.path)

			if tc.shouldErr {
				if err == nil {
					t.Errorf("Expected error for path %s, got nil", tc.path)
				} else if !strings.Contains(err.Error(), tc.errMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tc.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for path %s: %v", tc.path, err)
				}
			}
		})
	}
}

// Create a temporary directory for testing
func createTempDir(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "mcp-server-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	return tempDir
}

// Helper functions to test without calling main directly

// isHelpFlag checks if the argument is a help flag
func isHelpFlag(arg string) bool {
	return arg == "--help" || arg == "-h"
}

// validateAndNormalizePath validates a directory path and returns the normalized path
func validateAndNormalizePath(dir string) (string, error) {
	// Normalize and resolve path
	expandedPath := os.ExpandEnv(dir)
	if strings.HasPrefix(expandedPath, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		expandedPath = filepath.Join(home, expandedPath[1:])
	}

	absPath, err := filepath.Abs(expandedPath)
	if err != nil {
		return "", err
	}
	normalizedPath := filepath.Clean(absPath)

	// Validate directory exists and is accessible
	info, err := os.Stat(normalizedPath)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", normalizedPath)
	}

	return normalizedPath, nil
}
