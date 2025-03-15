package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/moguyn/mcp-go-filesystem/internal/server"
	"github.com/moguyn/mcp-go-filesystem/internal/tools"
)

// Version information set by build flags
var (
	Version = "0.0.1"
)

// printUsage prints usage information
func printUsage() {
	fmt.Fprintf(os.Stderr, "MCP Filesystem Server v%s\n\n", Version)
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

func main() {
	// Print version information
	fmt.Fprintf(os.Stderr, "MCP Filesystem Server v%s\n", Version)

	// Command line argument parsing
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Check for help flag
	if os.Args[1] == "--help" || os.Args[1] == "-h" {
		printUsage()
		os.Exit(0)
	}

	// Parse command line options
	serverMode := server.StdioMode
	listenAddr := "0.0.0.0:38085"

	// Get allowed directories from command line arguments
	allowedDirs := make([]string, 0)

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]

		// Check for options
		if strings.HasPrefix(arg, "--mode=") {
			mode := strings.TrimPrefix(arg, "--mode=")
			switch mode {
			case "stdio":
				serverMode = server.StdioMode
			case "sse":
				serverMode = server.SSEMode
			default:
				fmt.Fprintf(os.Stderr, "Error: Invalid server mode: %s\n", mode)
				printUsage()
				os.Exit(1)
			}
			continue
		}

		if strings.HasPrefix(arg, "--listen=") {
			listenAddr = strings.TrimPrefix(arg, "--listen=")
			continue
		}

		// If not an option, treat as directory
		// Normalize and resolve path
		expandedPath := tools.ExpandHome(arg)
		absPath, err := filepath.Abs(expandedPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving path %s: %v\n", arg, err)
			os.Exit(1)
		}
		normalizedPath := filepath.Clean(absPath)

		// Validate directory exists and is accessible
		info, err := os.Stat(normalizedPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error accessing directory %s: %v\n", arg, err)
			os.Exit(1)
		}
		if !info.IsDir() {
			fmt.Fprintf(os.Stderr, "Error: %s is not a directory\n", arg)
			os.Exit(1)
		}

		allowedDirs = append(allowedDirs, normalizedPath)
	}

	// Ensure we have at least one allowed directory
	if len(allowedDirs) == 0 {
		fmt.Fprintln(os.Stderr, "Error: No allowed directories specified")
		printUsage()
		os.Exit(1)
	}

	// Create and initialize the server
	s := server.NewServer(Version, allowedDirs, serverMode, listenAddr)
	s.Initialize()

	// Start the server
	if err := s.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error running server: %v\n", err)
		os.Exit(1)
	}
}
