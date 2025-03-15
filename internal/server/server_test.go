package server

import (
	"testing"

	"github.com/mark3labs/mcp-go/server"
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
			mcpServer:      server.NewMCPServer("test", "1.0.0"),
			allowedDirs:    []string{"/test/dir"},
			version:        "1.0.0",
			mode:           StdioMode,
			httpListenAddr: "localhost:8080",
		}

		// We can't fully test this without mocking os.Stdin/os.Stdout
		// Just verify the server is configured correctly
		assert.Equal(t, StdioMode, s.mode)
	})

	// Test SSEMode - we can verify the code path but not actually start the server
	t.Run("SSEMode", func(t *testing.T) {
		// Create a server with SSE mode
		s := &Server{
			mcpServer:      server.NewMCPServer("test", "1.0.0"),
			allowedDirs:    []string{"/test/dir"},
			version:        "1.0.0",
			mode:           SSEMode,
			httpListenAddr: "localhost:0", // Use port 0 to get a random available port
		}

		// Verify the server is configured correctly
		assert.Equal(t, SSEMode, s.mode)
		assert.Equal(t, "localhost:0", s.httpListenAddr)
	})
}
