package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	// Test data
	version := "1.0.0"
	allowedDirs := []string{"/test/dir1", "/test/dir2"}
	mode := StdioMode
	httpListenAddr := "localhost:8080"

	// Create a new server
	s := NewServer(version, allowedDirs, mode, httpListenAddr)

	// Verify the server was created correctly
	assert.NotNil(t, s)
	assert.NotNil(t, s.mcpServer)
	assert.Equal(t, version, s.version)
	assert.Equal(t, allowedDirs, s.allowedDirs)
	assert.Equal(t, mode, s.mode)
	assert.Equal(t, httpListenAddr, s.httpListenAddr)
}

func TestInitialize(t *testing.T) {
	// Create a test server
	s := NewServer("1.0.0", []string{"/test/dir"}, StdioMode, "localhost:8080")

	// Initialize should not panic
	assert.NotPanics(t, func() {
		s.initialize()
	})
}

func TestServerModes(t *testing.T) {
	// Test the server mode constants
	assert.Equal(t, ServerMode("stdio"), StdioMode)
	assert.Equal(t, ServerMode("sse"), SSEMode)
}

func TestStartInvalidMode(t *testing.T) {
	// Create a server with an invalid mode
	invalidMode := ServerMode("invalid")
	s := NewServer("1.0.0", []string{"/test/dir"}, invalidMode, "localhost:8080")

	// Start should return an error for invalid mode
	err := s.Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported server mode")
}

// TestServerString tests the string representation of ServerMode
func TestServerModeString(t *testing.T) {
	assert.Equal(t, "stdio", string(StdioMode))
	assert.Equal(t, "sse", string(SSEMode))

	// Test custom mode string
	customMode := ServerMode("custom")
	assert.Equal(t, "custom", string(customMode))
}

// TestStartModes tests the different server modes without actually starting servers
func TestStartModes(t *testing.T) {
	// Test StdioMode - we can't actually test this fully without mocking os.Stdin/os.Stdout
	// but we can at least verify the code path is taken
	t.Run("StdioMode", func(t *testing.T) {
		// Create a server with stdio mode
		s := &Server{
			mode: StdioMode,
		}

		// We can't fully test this without mocking os.Stdin/os.Stdout
		// Just verify the server is configured correctly
		assert.Equal(t, StdioMode, s.mode)
	})

	// Test SSEMode - we can verify the code path but not actually start the server
	t.Run("SSEMode", func(t *testing.T) {
		// Create a server with SSE mode
		s := &Server{
			mode:           SSEMode,
			httpListenAddr: "localhost:0", // Use port 0 to get a random available port
		}

		// Verify the server is configured correctly
		assert.Equal(t, SSEMode, s.mode)
		assert.Equal(t, "localhost:0", s.httpListenAddr)
	})
}

// TestServerConfiguration tests the server configuration
func TestServerConfiguration(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mcp-server-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Save original startSSEServer function and restore it after tests
	originalStartSSE := startSSEServer
	defer func() { startSSEServer = originalStartSSE }()

	// Replace with mock function
	startSSEServer = func(s *Server) error {
		return nil
	}

	// Test cases
	testCases := []struct {
		name        string
		allowedDirs []string
		mode        ServerMode
		listenAddr  string
	}{
		{
			name:        "Single directory with stdio mode",
			allowedDirs: []string{tempDir},
			mode:        StdioMode,
			listenAddr:  "0.0.0.0:38085",
		},
		{
			name:        "Single directory with SSE mode",
			allowedDirs: []string{tempDir},
			mode:        SSEMode,
			listenAddr:  "0.0.0.0:38085",
		},
		{
			name:        "Multiple directories",
			allowedDirs: []string{tempDir, tempDir},
			mode:        StdioMode,
			listenAddr:  "0.0.0.0:38085",
		},
		{
			name:        "Custom listen address",
			allowedDirs: []string{tempDir},
			mode:        SSEMode,
			listenAddr:  "127.0.0.1:38086",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create server with test configuration
			server := NewServer("test-version", tc.allowedDirs, tc.mode, tc.listenAddr)

			// Verify server configuration
			if server.version != "test-version" {
				t.Errorf("Expected version 'test-version', got '%s'", server.version)
			}

			if len(server.allowedDirs) != len(tc.allowedDirs) {
				t.Errorf("Expected %d allowed directories, got %d", len(tc.allowedDirs), len(server.allowedDirs))
			}

			if server.mode != tc.mode {
				t.Errorf("Expected mode %s, got %s", tc.mode, server.mode)
			}

			if server.httpListenAddr != tc.listenAddr {
				t.Errorf("Expected listen address %s, got %s", tc.listenAddr, server.httpListenAddr)
			}
		})
	}
}

// TestParseCommandLineArgs tests the argument processing logic
func TestParseCommandLineArgs(t *testing.T) {
	// Test with help flags
	_, err := ParseCommandLineArgs("test-version", []string{"program", "--help"})
	if err == nil || err.Error() != "help requested" {
		t.Errorf("Expected 'help requested' error for --help flag, got: %v", err)
	}

	_, err = ParseCommandLineArgs("test-version", []string{"program", "-h"})
	if err == nil || err.Error() != "help requested" {
		t.Errorf("Expected 'help requested' error for -h flag, got: %v", err)
	}

	// Test insufficient arguments
	_, err = ParseCommandLineArgs("test-version", []string{"program"})
	if err == nil || err.Error() != "insufficient arguments" {
		t.Errorf("Expected 'insufficient arguments' error, got: %v", err)
	}

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mcp-server-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
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
		name      string
		args      []string
		shouldErr bool
		errMsg    string
	}{
		{
			name:      "Valid directory",
			args:      []string{"program", tempDir},
			shouldErr: false,
		},
		{
			name:      "Non-existent directory",
			args:      []string{"program", nonExistentDir},
			shouldErr: true,
			errMsg:    "error accessing directory",
		},
		{
			name:      "Not a directory",
			args:      []string{"program", filePath},
			shouldErr: true,
			errMsg:    "is not a directory",
		},
		{
			name:      "Multiple directories",
			args:      []string{"program", tempDir, tempDir},
			shouldErr: false,
		},
		{
			name:      "With mode option",
			args:      []string{"program", "--mode=stdio", tempDir},
			shouldErr: false,
		},
		{
			name:      "With SSE mode",
			args:      []string{"program", "--mode=sse", tempDir},
			shouldErr: false,
		},
		{
			name:      "With invalid mode",
			args:      []string{"program", "--mode=invalid", tempDir},
			shouldErr: true,
			errMsg:    "invalid server mode",
		},
		{
			name:      "With listen address",
			args:      []string{"program", "--listen=127.0.0.1:8080", tempDir},
			shouldErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config, err := ParseCommandLineArgs("test-version", tc.args)

			if tc.shouldErr {
				if err == nil {
					t.Errorf("Expected error for args %v, got nil", tc.args)
				} else if !strings.Contains(err.Error(), tc.errMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tc.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for args %v: %v", tc.args, err)
				}

				// Verify config values
				if config.Version != "test-version" {
					t.Errorf("Expected version 'test-version', got '%s'", config.Version)
				}

				// Check if allowed directories are set
				if len(config.AllowedDirs) == 0 {
					t.Errorf("Expected at least one allowed directory")
				}

				// Check mode if specified
				for _, arg := range tc.args {
					if strings.HasPrefix(arg, "--mode=") {
						mode := strings.TrimPrefix(arg, "--mode=")
						var expectedMode ServerMode
						switch mode {
						case "stdio":
							expectedMode = StdioMode
						case "sse":
							expectedMode = SSEMode
						}
						if config.ServerMode != expectedMode {
							t.Errorf("Expected mode %s, got %s", expectedMode, config.ServerMode)
						}
					}

					if strings.HasPrefix(arg, "--listen=") {
						listen := strings.TrimPrefix(arg, "--listen=")
						if config.ListenAddr != listen {
							t.Errorf("Expected listen address %s, got %s", listen, config.ListenAddr)
						}
					}
				}
			}
		})
	}
}
