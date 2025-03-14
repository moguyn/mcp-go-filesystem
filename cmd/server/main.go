package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/modelcontextprotocol/server-filesystem/internal/server"
)

// Version information set by build flags
var (
	Version   = "dev"
	BuildDate = "unknown"
)

// printUsage prints usage information
func printUsage() {
	fmt.Fprintf(os.Stderr, "MCP Filesystem Server v%s (built %s)\n\n", Version, BuildDate)
	fmt.Fprintln(os.Stderr, "Usage: mcp-server-filesystem <allowed-directory> [additional-directories...]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Options:")
	fmt.Fprintln(os.Stderr, "  --help, -h    Show this help message")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "The server will only allow operations within the specified directories.")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Example:")
	fmt.Fprintln(os.Stderr, "  mcp-server-filesystem /path/to/dir1 /path/to/dir2")
}

func main() {
	// Print version information
	fmt.Fprintf(os.Stderr, "MCP Filesystem Server v%s (built %s)\n", Version, BuildDate)

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

	// Get allowed directories from command line arguments
	allowedDirs := make([]string, 0, len(os.Args)-1)
	for _, dir := range os.Args[1:] {
		// Normalize and resolve path
		expandedPath := server.ExpandHome(dir)
		absPath, err := filepath.Abs(expandedPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving path %s: %v\n", dir, err)
			os.Exit(1)
		}
		normalizedPath := filepath.Clean(absPath)

		// Validate directory exists and is accessible
		info, err := os.Stat(normalizedPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error accessing directory %s: %v\n", dir, err)
			os.Exit(1)
		}
		if !info.IsDir() {
			fmt.Fprintf(os.Stderr, "Error: %s is not a directory\n", dir)
			os.Exit(1)
		}

		allowedDirs = append(allowedDirs, normalizedPath)
	}

	// Start the MCP server
	fmt.Fprintln(os.Stderr, "Secure MCP Filesystem Server running on stdio")
	fmt.Fprintln(os.Stderr, "Allowed directories:", allowedDirs)

	// Create and run the server
	s := server.NewServer(allowedDirs)
	if err := s.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error running server: %v\n", err)
		os.Exit(1)
	}
}
