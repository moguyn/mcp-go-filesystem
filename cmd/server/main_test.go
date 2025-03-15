package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/moguyn/mcp-go-filesystem/internal/server"
)

func TestPrintUsage(t *testing.T) {
	// Capture stdout
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Call the function
	server.PrintUsage(Version)

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
		"Examples:",
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

// TestMainFunction tests the main function with various arguments
func TestMainFunction(t *testing.T) {
	if os.Getenv("TEST_MAIN_FUNCTION") == "1" {
		// This will be executed when the test spawns a subprocess
		// We don't actually call main() here to avoid exiting the test process
		return
	}

	// Create a temporary directory for testing
	tempDir := createTempDir(t)
	defer os.RemoveAll(tempDir)

	// Test cases for different command line arguments
	testCases := []struct {
		name          string
		args          []string
		expectedError bool
		expectedText  string
	}{
		{
			name:          "Help flag",
			args:          []string{"--help"},
			expectedError: false,
			expectedText:  "Usage:",
		},
		{
			name:          "Short help flag",
			args:          []string{"-h"},
			expectedError: false,
			expectedText:  "Usage:",
		},
		{
			name:          "No arguments",
			args:          []string{},
			expectedError: true,
			expectedText:  "Usage:",
		},
		{
			name:          "Non-existent directory",
			args:          []string{filepath.Join(tempDir, "non-existent")},
			expectedError: true,
			expectedText:  "error accessing directory",
		},
		{
			name:          "Valid directory",
			args:          []string{tempDir},
			expectedError: false,
			expectedText:  "MCP Filesystem Server",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Prepare command to run the test in a subprocess
			cmd := buildTestCommand(t, tc.args)

			// Run the command and capture output
			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			// Check if error matches expectation
			if tc.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v, output: %s", err, outputStr)
				}
			}

			// Check if output contains expected text
			if !strings.Contains(outputStr, tc.expectedText) {
				t.Errorf("Expected output to contain '%s', but got: %s", tc.expectedText, outputStr)
			}
		})
	}
}

// buildTestCommand creates a command to test main function in a subprocess
func buildTestCommand(t *testing.T, args []string) *exec.Cmd {
	executable, err := os.Executable()
	if err != nil {
		t.Fatalf("Could not get test executable: %v", err)
	}

	// Prepare command with TEST_MAIN_FUNCTION=1 environment
	cmd := exec.Command(executable, "-test.run=TestMainFunction")
	cmd.Env = append(os.Environ(), "TEST_MAIN_FUNCTION=1")

	// Add the program name as first argument (simulating os.Args[0])
	allArgs := append([]string{"mcp-server-filesystem"}, args...)
	cmd.Env = append(cmd.Env, fmt.Sprintf("ARGS=%s", strings.Join(allArgs, ",")))

	return cmd
}

// TestMain is used to set up the test environment
func TestMain(m *testing.M) {
	// Check if we're in the subprocess mode
	if os.Getenv("TEST_MAIN_FUNCTION") == "1" {
		// Parse ARGS environment variable
		argsStr := os.Getenv("ARGS")
		if argsStr != "" {
			args := strings.Split(argsStr, ",")
			// Replace os.Args with our test args
			os.Args = args
			// Call main() which will exit the process
			main()
			return
		}
	}

	// Normal test execution
	os.Exit(m.Run())
}
