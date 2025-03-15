package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/moguyn/mcp-go-filesystem/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	// Test data
	version := "1.0.0"
	allowedDirs := []string{"/test/dir1", "/test/dir2"}
	mode := config.StdioMode
	httpListenAddr := "localhost:8080"

	// Create a config
	cfg := &config.Config{
		Version:     version,
		AllowedDirs: allowedDirs,
		ServerMode:  mode,
		ListenAddr:  httpListenAddr,
		LogLevel:    "INFO",
	}

	// Create a new server
	s := NewServer(cfg)

	// Verify the server was created correctly
	assert.NotNil(t, s)
	assert.NotNil(t, s.mcpServer)
	assert.Equal(t, version, s.version)
	assert.Equal(t, allowedDirs, s.allowedDirs)
	assert.Equal(t, mode, s.mode)
	assert.Equal(t, httpListenAddr, s.httpListenAddr)
}

func TestInitialize(t *testing.T) {
	// Create a test config
	cfg := &config.Config{
		Version:     "1.0.0",
		AllowedDirs: []string{"/test/dir"},
		ServerMode:  config.StdioMode,
		ListenAddr:  "localhost:8080",
		LogLevel:    "INFO",
	}

	// Create a test server
	s := NewServer(cfg)

	// Initialize should not panic
	assert.NotPanics(t, func() {
		s.initialize()
	})
}

func TestServerModes(t *testing.T) {
	// Test the server mode constants
	assert.Equal(t, "stdio", string(config.StdioMode))
	assert.Equal(t, "sse", string(config.SSEMode))
}

func TestStartInvalidMode(t *testing.T) {
	// Create a server with an invalid mode
	invalidMode := config.ServerMode("invalid")
	cfg := &config.Config{
		Version:     "1.0.0",
		AllowedDirs: []string{"/test/dir"},
		ServerMode:  invalidMode,
		ListenAddr:  "localhost:8080",
		LogLevel:    "INFO",
	}
	s := NewServer(cfg)

	// Start should return an error for invalid mode
	err := s.Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported server mode")
}

// TestServerModeString tests the string representation of ServerMode
func TestServerModeString(t *testing.T) {
	assert.Equal(t, "stdio", string(config.StdioMode))
	assert.Equal(t, "sse", string(config.SSEMode))

	// Test custom mode string
	customMode := config.ServerMode("custom")
	assert.Equal(t, "custom", string(customMode))
}

// TestStartModes tests the different server modes without actually starting servers
func TestStartModes(t *testing.T) {
	// Test StdioMode - we can't actually test this fully without mocking os.Stdin/os.Stdout
	// but we can at least verify the code path is taken
	t.Run("StdioMode", func(t *testing.T) {
		// Create a server with stdio mode
		cfg := &config.Config{
			Version:     "1.0.0",
			AllowedDirs: []string{"/test/dir"},
			ServerMode:  config.StdioMode,
			ListenAddr:  "localhost:8080",
			LogLevel:    "INFO",
		}
		s := NewServer(cfg)

		// We can't fully test this without mocking os.Stdin/os.Stdout
		// Just verify the server is configured correctly
		assert.Equal(t, config.StdioMode, s.mode)
	})

	// Test SSEMode - we can verify the code path but not actually start the server
	t.Run("SSEMode", func(t *testing.T) {
		// Create a server with SSE mode
		cfg := &config.Config{
			Version:     "1.0.0",
			AllowedDirs: []string{"/test/dir"},
			ServerMode:  config.SSEMode,
			ListenAddr:  "localhost:0", // Use port 0 to get a random available port
			LogLevel:    "INFO",
		}
		s := NewServer(cfg)

		// Verify the server is configured correctly
		assert.Equal(t, config.SSEMode, s.mode)
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
		mode        config.ServerMode
		listenAddr  string
	}{
		{
			name:        "Single directory with stdio mode",
			allowedDirs: []string{tempDir},
			mode:        config.StdioMode,
			listenAddr:  "0.0.0.0:38085",
		},
		{
			name:        "Single directory with SSE mode",
			allowedDirs: []string{tempDir},
			mode:        config.SSEMode,
			listenAddr:  "0.0.0.0:38085",
		},
		{
			name:        "Multiple directories",
			allowedDirs: []string{tempDir, tempDir},
			mode:        config.StdioMode,
			listenAddr:  "0.0.0.0:38085",
		},
		{
			name:        "Custom listen address",
			allowedDirs: []string{tempDir},
			mode:        config.SSEMode,
			listenAddr:  "127.0.0.1:38086",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create server with test configuration
			cfg := &config.Config{
				Version:     "test-version",
				AllowedDirs: tc.allowedDirs,
				ServerMode:  tc.mode,
				ListenAddr:  tc.listenAddr,
				LogLevel:    "INFO",
			}
			server := NewServer(cfg)

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
	_, err := config.ParseCommandLineArgs("test-version", []string{"program", "--help"})
	if err == nil || err.Error() != "help requested" {
		t.Errorf("Expected 'help requested' error for --help flag, got: %v", err)
	}

	_, err = config.ParseCommandLineArgs("test-version", []string{"program", "-h"})
	if err == nil || err.Error() != "help requested" {
		t.Errorf("Expected 'help requested' error for -h flag, got: %v", err)
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
	}{
		{
			name:      "Valid directory",
			args:      []string{"program", tempDir},
			shouldErr: false,
		},
		{
			name:      "Valid directory with mode",
			args:      []string{"program", "--mode=stdio", tempDir},
			shouldErr: false,
		},
		{
			name:      "Valid directory with SSE mode and listen address",
			args:      []string{"program", "--mode=sse", "--listen=127.0.0.1:38086", tempDir},
			shouldErr: false,
		},
		{
			name:      "Invalid mode",
			args:      []string{"program", "--mode=invalid", tempDir},
			shouldErr: true,
		},
		{
			name:      "Non-existent directory",
			args:      []string{"program", nonExistentDir},
			shouldErr: true,
		},
		{
			name:      "File instead of directory",
			args:      []string{"program", filePath},
			shouldErr: true,
		},
		{
			name:      "No directories",
			args:      []string{"program", "--mode=stdio"},
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := config.ParseCommandLineArgs("test-version", tc.args)

			if tc.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cfg)

				// Verify the configuration
				assert.Equal(t, "test-version", cfg.Version)

				// Check if allowed directories are set
				assert.Greater(t, len(cfg.AllowedDirs), 0)

				// Check mode if specified
				for _, arg := range tc.args {
					if strings.HasPrefix(arg, "--mode=") {
						mode := strings.TrimPrefix(arg, "--mode=")
						switch mode {
						case "stdio":
							assert.Equal(t, config.StdioMode, cfg.ServerMode)
						case "sse":
							assert.Equal(t, config.SSEMode, cfg.ServerMode)
						}
					}

					if strings.HasPrefix(arg, "--listen=") {
						listen := strings.TrimPrefix(arg, "--listen=")
						assert.Equal(t, listen, cfg.ListenAddr)
					}
				}
			}
		})
	}
}

func TestStop(t *testing.T) {
	// Create a test config
	cfg := &config.Config{
		Version:     "1.0.0",
		AllowedDirs: []string{"/test/dir"},
		ServerMode:  config.StdioMode,
		ListenAddr:  "localhost:8080",
		LogLevel:    "INFO",
	}

	// Create a new server
	s := NewServer(cfg)

	// Ensure context is not canceled before Stop
	select {
	case <-s.ctx.Done():
		t.Error("Context should not be canceled before Stop is called")
	default:
		// This is the expected path
	}

	// Call Stop
	s.Stop()

	// Verify context is canceled after Stop
	select {
	case <-s.ctx.Done():
		// This is the expected path
	default:
		t.Error("Context should be canceled after Stop is called")
	}
}

func TestStartSSEServer(t *testing.T) {
	// Create a test config
	cfg := &config.Config{
		Version:     "1.0.0",
		AllowedDirs: []string{"/test/dir"},
		ServerMode:  config.SSEMode,
		ListenAddr:  "localhost:8080",
		LogLevel:    "INFO",
	}

	// Create a new server
	s := NewServer(cfg)

	// Mock the startSSEServer function to avoid actually starting a server
	startCalled := false
	origStartSSEServer := startSSEServer
	defer func() { startSSEServer = origStartSSEServer }()

	startSSEServer = func(s *Server) error {
		startCalled = true
		// Verify the server is properly configured
		assert.Equal(t, "localhost:8080", s.httpListenAddr)
		return nil
	}

	// Call Start
	err := s.Start()

	// Verify startSSEServer was called
	assert.True(t, startCalled)
	assert.Nil(t, err)
}
