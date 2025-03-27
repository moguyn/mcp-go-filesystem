package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/moguyn/mcp-go-filesystem/internal/errors"
	"github.com/moguyn/mcp-go-filesystem/internal/tools"
)

// ServerMode defines the mode in which the server operates
type ServerMode string

const (
	// StdioMode indicates the server should run using standard input/output
	StdioMode ServerMode = "stdio"
	// SSEMode indicates the server should run as an HTTP server with SSE
	SSEMode ServerMode = "sse"

	// Environment variable names
	envServerMode = "MCP_SERVER_MODE"
	envListenAddr = "MCP_LISTEN_ADDR"
)

// Config holds the configuration for the filesystem server
type Config struct {
	Version     string
	AllowedDirs []string
	ServerMode  ServerMode
	ListenAddr  string
	LogLevel    string
}

// DefaultConfig returns a default configuration
func DefaultConfig(version string) *Config {
	return &Config{
		Version:     version,
		ServerMode:  StdioMode,
		ListenAddr:  "0.0.0.0:38085",
		AllowedDirs: make([]string, 0),
		LogLevel:    "INFO",
	}
}

// ParseCommandLineArgs parses command line arguments and returns a server configuration
func ParseCommandLineArgs(version string, args []string) (*Config, error) {
	// Check for help flag
	if len(args) < 2 {
		return nil, errors.NewFileSystemError("parse_args", "", errors.ErrInvalidArgument)
	}

	if args[1] == "--help" || args[1] == "-h" {
		return nil, fmt.Errorf("help requested")
	}

	// Initialize default configuration
	config := DefaultConfig(version)

	// Check environment variables first
	if mode := os.Getenv(envServerMode); mode != "" {
		switch strings.ToLower(mode) {
		case "stdio":
			config.ServerMode = StdioMode
		case "sse":
			config.ServerMode = SSEMode
		default:
			return nil, errors.NewFileSystemError("parse_args", "", fmt.Errorf("invalid server mode in environment: %s", mode))
		}
	}

	if addr := os.Getenv(envListenAddr); addr != "" {
		config.ListenAddr = addr
	}

	// Parse command line options (these will override environment variables)
	for i := 1; i < len(args); i++ {
		arg := args[i]

		// Check for options
		if strings.HasPrefix(arg, "--mode=") {
			mode := strings.TrimPrefix(arg, "--mode=")
			switch mode {
			case "stdio":
				config.ServerMode = StdioMode
			case "sse":
				config.ServerMode = SSEMode
			default:
				return nil, errors.NewFileSystemError("parse_args", "", fmt.Errorf("invalid server mode: %s", mode))
			}
			continue
		}

		if strings.HasPrefix(arg, "--listen=") {
			config.ListenAddr = strings.TrimPrefix(arg, "--listen=")
			continue
		}

		if strings.HasPrefix(arg, "--log-level=") {
			config.LogLevel = strings.ToUpper(strings.TrimPrefix(arg, "--log-level="))
			continue
		}

		// If not an option, treat as directory
		dir, err := validateDirectory(arg)
		if err != nil {
			return nil, err
		}

		config.AllowedDirs = append(config.AllowedDirs, dir)
	}

	// Ensure we have at least one allowed directory
	if len(config.AllowedDirs) == 0 {
		return nil, errors.NewFileSystemError("parse_args", "", errors.ErrInvalidArgument)
	}

	return config, nil
}

// validateDirectory validates that a directory exists and is accessible
func validateDirectory(path string) (string, error) {
	// Normalize and resolve path
	expandedPath := tools.ExpandHome(path)
	absPath, err := filepath.Abs(expandedPath)
	if err != nil {
		return "", errors.NewFileSystemError("validate_directory", path, err)
	}
	normalizedPath := filepath.Clean(absPath)

	// Validate directory exists and is accessible
	info, err := os.Stat(normalizedPath)
	if err != nil {
		return "", errors.NewFileSystemError("validate_directory", path, err)
	}
	if !info.IsDir() {
		return "", errors.NewFileSystemError("validate_directory", path, fmt.Errorf("not a directory"))
	}

	return normalizedPath, nil
}

// PrintUsage prints usage information
func PrintUsage(version string) {
	fmt.Fprintf(os.Stderr, "MCP Filesystem Server v%s\n\n", version)
	fmt.Fprintln(os.Stderr, "Usage: mcp-server-filesystem [options] <allowed-directory> [additional-directories...]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Options:")
	fmt.Fprintln(os.Stderr, "  --help, -h           Show this help message")
	fmt.Fprintln(os.Stderr, "  --mode=<mode>        Server mode: 'stdio' (default) or 'sse'")
	fmt.Fprintln(os.Stderr, "  --listen=<address>   HTTP listen address for SSE mode (default: 0.0.0.0:38085)")
	fmt.Fprintln(os.Stderr, "  --log-level=<level>  Log level: DEBUG, INFO, WARN, ERROR, FATAL (default: INFO)")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Environment Variables:")
	fmt.Fprintln(os.Stderr, "  MCP_SERVER_MODE      Server mode (overridden by --mode)")
	fmt.Fprintln(os.Stderr, "  MCP_LISTEN_ADDR      HTTP listen address (overridden by --listen)")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "The server will only allow operations within the specified directories.")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Examples:")
	fmt.Fprintln(os.Stderr, "  mcp-server-filesystem /path/to/dir1 /path/to/dir2")
	fmt.Fprintln(os.Stderr, "  mcp-server-filesystem --mode=sse --listen=0.0.0.0:38085 --log-level=DEBUG /path/to/dir")
	fmt.Fprintln(os.Stderr, "  MCP_SERVER_MODE=sse MCP_LISTEN_ADDR=0.0.0.0:38086 mcp-server-filesystem /path/to/dir")
}
