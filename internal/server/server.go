package server

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/server"
	"github.com/moguyn/mcp-go-filesystem/internal/config"
	"github.com/moguyn/mcp-go-filesystem/internal/logging"
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
	mode           config.ServerMode
	httpListenAddr string
	logger         *logging.Logger
	ctx            context.Context
	cancel         context.CancelFunc
}

// NewServer creates a new filesystem server
func NewServer(cfg *config.Config) *Server {
	// Create a new MCP server
	mcpServer := server.NewMCPServer(
		"mcp-go-filesystem",
		cfg.Version,
	)

	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Create a logger
	logger := logging.DefaultLogger("server")
	if cfg.LogLevel != "" {
		var level logging.LogLevel
		switch cfg.LogLevel {
		case "DEBUG":
			level = logging.DEBUG
		case "INFO":
			level = logging.INFO
		case "WARN":
			level = logging.WARN
		case "ERROR":
			level = logging.ERROR
		case "FATAL":
			level = logging.FATAL
		default:
			level = logging.INFO
		}
		logger.SetLevel(level)
	}

	return &Server{
		mcpServer:      mcpServer,
		allowedDirs:    cfg.AllowedDirs,
		version:        cfg.Version,
		mode:           cfg.ServerMode,
		httpListenAddr: cfg.ListenAddr,
		logger:         logger,
		ctx:            ctx,
		cancel:         cancel,
	}
}

// initialize sets up the server by registering all tools
func (s *Server) initialize() {
	// Register all filesystem tools
	tools.RegisterTools(s.mcpServer, s.allowedDirs)
}

// Start starts the server in the configured mode
func (s *Server) Start() error {
	// Initialize the server before starting
	s.initialize()

	s.logger.Info("Allowed directories: %v", s.allowedDirs)

	switch s.mode {
	case config.StdioMode:
		s.logger.Info("Running in stdio mode")
		return server.ServeStdio(s.mcpServer)
	case config.SSEMode:
		s.logger.Info("Running in SSE mode on %s", s.httpListenAddr)
		return startSSEServer(s)
	default:
		return fmt.Errorf("unsupported server mode: %s", s.mode)
	}
}

// Stop stops the server
func (s *Server) Stop() {
	s.logger.Info("Stopping server")
	s.cancel()
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
