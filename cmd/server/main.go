package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/moguyn/mcp-go-filesystem/internal/config"
	"github.com/moguyn/mcp-go-filesystem/internal/logging"
	"github.com/moguyn/mcp-go-filesystem/internal/server"
)

// Version information set by build flags
var (
	Version = "0.0.1"
)

func main() {
	// Create a logger
	logger := logging.DefaultLogger("main")

	// Print version information
	logger.Info("MCP Filesystem Server v%s", Version)

	// Parse command line arguments
	cfg, err := config.ParseCommandLineArgs(Version, os.Args)
	if err != nil {
		if err.Error() == "help requested" {
			config.PrintUsage(Version)
			os.Exit(0)
		} else {
			logger.Error("Error: %v", err)
			config.PrintUsage(Version)
			os.Exit(1)
		}
	}

	// Create and initialize the server
	s := server.NewServer(cfg)

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := s.Start(); err != nil {
			errChan <- err
		}
	}()

	// Wait for either an error or a signal
	select {
	case err := <-errChan:
		logger.Fatal("Fatal error running server: %v", err)
	case sig := <-sigChan:
		logger.Info("Received signal: %v", sig)
		s.Stop()
	}
}
