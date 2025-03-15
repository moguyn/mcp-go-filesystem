package server

import (
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/server"
	"github.com/moguyn/mcp-go-filesystem/internal/tools"
)

// ServerMode defines the mode in which the server operates
type ServerMode string

const (
	// StdioMode indicates the server should run using standard input/output
	StdioMode ServerMode = "stdio"
	// SSEMode indicates the server should run as an HTTP server with SSE
	SSEMode ServerMode = "sse"
)

// Server represents the filesystem server
type Server struct {
	mcpServer      *server.MCPServer
	allowedDirs    []string
	version        string
	mode           ServerMode
	httpListenAddr string
}

// NewServer creates a new filesystem server
func NewServer(version string, allowedDirs []string, mode ServerMode, httpListenAddr string) *Server {
	// Create a new MCP server
	mcpServer := server.NewMCPServer(
		"mcp-go-filesystem",
		version,
	)

	return &Server{
		mcpServer:      mcpServer,
		allowedDirs:    allowedDirs,
		version:        version,
		mode:           mode,
		httpListenAddr: httpListenAddr,
	}
}

// Initialize sets up the server by registering all tools
func (s *Server) Initialize() {
	// Register all filesystem tools
	tools.RegisterTools(s.mcpServer, s.allowedDirs)
}

// Start starts the server in the configured mode
func (s *Server) Start() error {
	fmt.Fprintln(os.Stderr, "Allowed directories:", s.allowedDirs)

	switch s.mode {
	case StdioMode:
		fmt.Fprintln(os.Stderr, "Running in stdio mode")
		return server.ServeStdio(s.mcpServer)
	case SSEMode:
		fmt.Fprintf(os.Stderr, "Running in SSE mode on %s\n", s.httpListenAddr)
		return startSSEServer(s)
	default:
		return fmt.Errorf("unsupported server mode: %s", s.mode)
	}
}

// startSSEServer starts the server in SSE mode
func (s *Server) startSSEServer() error {
	// Create an SSE server
	baseURL := "http://" + s.httpListenAddr
	sseServer := server.NewSSEServer(s.mcpServer, baseURL)

	// Start the SSE server
	return sseServer.Start(s.httpListenAddr)
}

// Default implementation of startSSEServer
var startSSEServer = func(s *Server) error {
	return s.startSSEServer()
}
