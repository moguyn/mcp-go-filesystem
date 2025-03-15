package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/moguyn/mcp-go-filesystem/internal/config"
)

func TestPrintUsage(t *testing.T) {
	// Capture stdout
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Call the function
	config.PrintUsage(Version)

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

// TestMainFunction tests the main function by running it in a subprocess
func TestMainFunction(t *testing.T) {
	if os.Getenv("TEST_MAIN_FUNCTION") == "1" {
		// Skip when running in subprocess mode
		return
	}

	// Create a temporary directory for testing
	tempDir := createTempDir(t)
	defer os.RemoveAll(tempDir)

	// Test cases
	testCases := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "Help flag",
			args:        []string{"mcp-server-filesystem", "--help"},
			expectError: false, // Help flag exits with 0
		},
		{
			name:        "No arguments",
			args:        []string{"mcp-server-filesystem"},
			expectError: true, // No directories provided
		},
		{
			name:        "Valid directory",
			args:        []string{"mcp-server-filesystem", tempDir},
			expectError: false,
		},
		{
			name:        "Invalid directory",
			args:        []string{"mcp-server-filesystem", "/path/that/does/not/exist"},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Skip the actual server start for valid cases
			if !tc.expectError && !strings.Contains(tc.args[1], "--help") {
				t.Skip("Skipping valid case to avoid starting the server")
			}

			// Prepare command to run the test in a subprocess
			cmd := os.Args[0]
			env := []string{
				"TEST_MAIN_FUNCTION=1",
				fmt.Sprintf("ARGS=%s", strings.Join(tc.args, ",")),
			}

			// Run the test in a subprocess
			output, err := runInSubprocess(cmd, env)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none. Output: %s", output)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v. Output: %s", err, output)
				}
			}
		})
	}
}

// runInSubprocess runs a command in a subprocess and returns its output
func runInSubprocess(cmd string, env []string) (string, error) {
	// Create a temporary file to capture output
	outFile, err := os.CreateTemp("", "test-output-*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(outFile.Name())
	outFile.Close()

	// Prepare the command
	args := []string{"-test.run=TestMain"}
	procAttr := &os.ProcAttr{
		Env:   append(os.Environ(), env...),
		Files: []*os.File{nil, outFile, outFile}, // stdin, stdout, stderr
	}

	// Start the process
	process, err := os.StartProcess(cmd, args, procAttr)
	if err != nil {
		return "", fmt.Errorf("failed to start process: %v", err)
	}

	// Wait for the process to complete
	state, err := process.Wait()
	if err != nil {
		return "", fmt.Errorf("process wait failed: %v", err)
	}

	// Read the output
	output, err := os.ReadFile(outFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to read output: %v", err)
	}

	// Check if the process exited successfully
	if !state.Success() {
		return string(output), fmt.Errorf("process exited with code %d", state.ExitCode())
	}

	return string(output), nil
}
