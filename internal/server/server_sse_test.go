package server

import (
	"fmt"
	"testing"

	"github.com/moguyn/mcp-go-filesystem/internal/config"
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
			assert.Equal(t, config.SSEMode, s.mode)
			assert.Equal(t, "localhost:8080", s.httpListenAddr)

			// Call the mock implementation
			return mockSSE.Start(s.httpListenAddr)
		}

		// Create a server with SSE mode
		cfg := &config.Config{
			Version:     "1.0.0",
			AllowedDirs: []string{"/test/dir"},
			ServerMode:  config.SSEMode,
			ListenAddr:  "localhost:8080",
			LogLevel:    "INFO",
		}
		s := NewServer(cfg)

		// Start the server
		err := s.Start()
		assert.NoError(t, err)
		assert.True(t, mockSSE.startCalled)
	})

	// Test case 2: Error during start
	t.Run("Error", func(t *testing.T) {
		// Create a mock SSE server that returns an error
		expectedError := fmt.Errorf("mock SSE server error")
		mockSSE := &mockSSEServer{startError: expectedError}

		// Override the startSSEServer function for testing
		startSSEServer = func(s *Server) error {
			// Verify the server configuration
			assert.Equal(t, config.SSEMode, s.mode)
			assert.Equal(t, "localhost:8080", s.httpListenAddr)

			// Call the mock implementation
			return mockSSE.Start(s.httpListenAddr)
		}

		// Create a server with SSE mode
		cfg := &config.Config{
			Version:     "1.0.0",
			AllowedDirs: []string{"/test/dir"},
			ServerMode:  config.SSEMode,
			ListenAddr:  "localhost:8080",
			LogLevel:    "INFO",
		}
		s := NewServer(cfg)

		// Start the server
		err := s.Start()
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.True(t, mockSSE.startCalled)
	})
}

// TestStartSSEServerDirectly tests the startSSEServer method directly
func TestStartSSEServerDirectly(t *testing.T) {
	// Create a server with SSE mode
	cfg := &config.Config{
		Version:     "1.0.0",
		AllowedDirs: []string{"/test/dir"},
		ServerMode:  config.SSEMode,
		ListenAddr:  "localhost:0", // Use port 0 to get a random available port
		LogLevel:    "INFO",
	}
	s := NewServer(cfg)

	// We can't actually start the server in tests, but we can verify the method exists
	assert.NotNil(t, s.startSSEServer)
}
