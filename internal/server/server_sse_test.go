package server

import (
	"fmt"
	"testing"

	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
)

// mockSSEServer is a mock implementation of the SSE server for testing
type mockSSEServer struct {
	startCalled bool
	startError  error
}

func (m *mockSSEServer) Start(addr string) error {
	m.startCalled = true
	return m.startError
}

// TestStartSSEServerWithMock tests the startSSEServer method using a mock
func TestStartSSEServerWithMock(t *testing.T) {
	// Create a test server
	originalStartSSEServer := startSSEServer
	defer func() { startSSEServer = originalStartSSEServer }()

	// Test case 1: Successful start
	t.Run("Success", func(t *testing.T) {
		// Create a mock SSE server that returns no error
		mockSSE := &mockSSEServer{startError: nil}

		// Override the startSSEServer function for testing
		startSSEServer = func(s *Server) error {
			// Verify the server configuration
			assert.Equal(t, SSEMode, s.mode)
			assert.Equal(t, "localhost:8080", s.httpListenAddr)

			// Call the mock implementation
			return mockSSE.Start(s.httpListenAddr)
		}

		// Create a server with SSE mode
		s := &Server{
			mcpServer:      server.NewMCPServer("test", "1.0.0"),
			allowedDirs:    []string{"/test/dir"},
			version:        "1.0.0",
			mode:           SSEMode,
			httpListenAddr: "localhost:8080",
		}

		// Start the server
		err := s.Start()

		// Verify no error was returned
		assert.NoError(t, err)
		assert.True(t, mockSSE.startCalled)
	})

	// Test case 2: Error during start
	t.Run("Error", func(t *testing.T) {
		// Create a mock SSE server that returns an error
		expectedError := fmt.Errorf("mock SSE server start error")
		mockSSE := &mockSSEServer{startError: expectedError}

		// Override the startSSEServer function for testing
		startSSEServer = func(s *Server) error {
			// Call the mock implementation
			return mockSSE.Start(s.httpListenAddr)
		}

		// Create a server with SSE mode
		s := &Server{
			mcpServer:      server.NewMCPServer("test", "1.0.0"),
			allowedDirs:    []string{"/test/dir"},
			version:        "1.0.0",
			mode:           SSEMode,
			httpListenAddr: "localhost:8080",
		}

		// Start the server
		err := s.Start()

		// Verify the expected error was returned
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.True(t, mockSSE.startCalled)
	})
}

// TestStartSSEServerDirectly tests the startSSEServer method directly
// This test will not actually start a server, but will verify the function
// doesn't panic and returns a non-nil error when given an invalid address
func TestStartSSEServerDirectly(t *testing.T) {
	// Create a server with an invalid address to force an error
	s := &Server{
		mcpServer:      server.NewMCPServer("test", "1.0.0"),
		allowedDirs:    []string{"/test/dir"},
		version:        "1.0.0",
		mode:           SSEMode,
		httpListenAddr: "invalid-address::", // Invalid address format
	}

	// Call startSSEServer directly
	err := s.startSSEServer()

	// We expect an error due to the invalid address
	assert.Error(t, err)
}
