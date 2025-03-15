package server

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/moguyn/mcp-go-filesystem/internal/tools"
)

// ServerConfig holds the configuration for the filesystem server
type ServerConfig struct {
	Version     string
	AllowedDirs []string
	ServerMode  ServerMode
	ListenAddr  string
}

// ParseCommandLineArgs parses command line arguments and returns a server configuration
func ParseCommandLineArgs(version string, args []string) (*ServerConfig, error) {
	// Check for help flag
	if len(args) < 2 {
		return nil, fmt.Errorf("insufficient arguments")
	}

	if args[1] == "--help" || args[1] == "-h" {
		return nil, fmt.Errorf("help requested")
	}

	// Initialize default configuration
	config := &ServerConfig{
		Version:     version,
		ServerMode:  StdioMode,
		ListenAddr:  "0.0.0.0:38085",
		AllowedDirs: make([]string, 0),
	}

	// Parse command line options
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
				return nil, fmt.Errorf("invalid server mode: %s", mode)
			}
			continue
		}

		if strings.HasPrefix(arg, "--listen=") {
			config.ListenAddr = strings.TrimPrefix(arg, "--listen=")
			continue
		}

		// If not an option, treat as directory
		// Normalize and resolve path
		expandedPath := tools.ExpandHome(arg)
		absPath, err := filepath.Abs(expandedPath)
		if err != nil {
			return nil, fmt.Errorf("error resolving path %s: %v", arg, err)
		}
		normalizedPath := filepath.Clean(absPath)

		// Validate directory exists and is accessible
		info, err := os.Stat(normalizedPath)
		if err != nil {
			return nil, fmt.Errorf("error accessing directory %s: %v", arg, err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("%s is not a directory", arg)
		}

		config.AllowedDirs = append(config.AllowedDirs, normalizedPath)
	}

	// Ensure we have at least one allowed directory
	if len(config.AllowedDirs) == 0 {
		return nil, fmt.Errorf("no allowed directories specified")
	}

	return config, nil
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
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "The server will only allow operations within the specified directories.")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Examples:")
	fmt.Fprintln(os.Stderr, "  mcp-server-filesystem /path/to/dir1 /path/to/dir2")
	fmt.Fprintln(os.Stderr, "  mcp-server-filesystem --mode=sse --listen=0.0.0.0:38085 /path/to/dir")
}
