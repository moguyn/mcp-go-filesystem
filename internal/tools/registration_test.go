package tools

import (
	"testing"

	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
)

func TestNewServiceProvider(t *testing.T) {
	allowedDirs := []string{"/test/dir1", "/test/dir2"}
	provider := NewServiceProvider(allowedDirs)

	assert.NotNil(t, provider)
	assert.NotNil(t, provider.fileService)
	assert.NotNil(t, provider.fileWriter)
	assert.NotNil(t, provider.fileManager)
	assert.NotNil(t, provider.directoryService)
	assert.NotNil(t, provider.searchService)
	assert.NotNil(t, provider.logger)
}

func TestRegisterTools(t *testing.T) {
	// Create a mock server
	mcpServer := server.NewMCPServer("test", "1.0.0")

	// Register tools
	allowedDirs := []string{"/test/dir1", "/test/dir2"}

	// This should not panic
	RegisterTools(mcpServer, allowedDirs)

	// We can't easily verify the registered tools with the real MCPServer
	// So we'll just make sure it doesn't panic
	assert.NotNil(t, mcpServer)
}
