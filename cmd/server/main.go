package main

import (
	"fmt"
	"os"

	"github.com/moguyn/mcp-go-filesystem/internal/server"
)

// Version information set by build flags
var (
	Version = "0.0.1"
)

func main() {
	// Print version information
	fmt.Fprintf(os.Stderr, "MCP Filesystem Server v%s\n", Version)

	// Parse command line arguments
	config, err := server.ParseCommandLineArgs(Version, os.Args)
	if err != nil {
		if err.Error() == "help requested" {
			server.PrintUsage(Version)
			os.Exit(0)
		} else if err.Error() == "insufficient arguments" {
			server.PrintUsage(Version)
			os.Exit(1)
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			server.PrintUsage(Version)
			os.Exit(1)
		}
	}

	// Create and initialize the server
	s := server.NewServer(config.Version, config.AllowedDirs, config.ServerMode, config.ListenAddr)

	// Start the server
	if err := s.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error running server: %v\n", err)
		os.Exit(1)
	}
}
