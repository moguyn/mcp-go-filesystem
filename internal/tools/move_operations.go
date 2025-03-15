package tools

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
)

// handleMoveFile handles the move_file tool
func handleMoveFile(request mcp.CallToolRequest, allowedDirectories []string) (*mcp.CallToolResult, error) {
	source, ok := request.Params.Arguments["source"].(string)
	if !ok {
		return nil, fmt.Errorf("source must be a string")
	}

	destination, ok := request.Params.Arguments["destination"].(string)
	if !ok {
		return nil, fmt.Errorf("destination must be a string")
	}

	// Validate paths
	validSource, err := ValidatePath(source, allowedDirectories)
	if err != nil {
		return nil, fmt.Errorf("invalid source path: %v", err)
	}

	validDestination, err := ValidatePath(destination, allowedDirectories)
	if err != nil {
		return nil, fmt.Errorf("invalid destination path: %v", err)
	}

	// Create parent directories if they don't exist
	destDir := filepath.Dir(validDestination)
	if err := os.MkdirAll(destDir, 0750); err != nil {
		return nil, fmt.Errorf("error creating destination directories: %v", err)
	}

	// Move file
	if err := os.Rename(validSource, validDestination); err != nil {
		return nil, fmt.Errorf("error moving file: %v", err)
	}

	return mcp.NewToolResultText(fmt.Sprintf("Moved %s to %s", source, destination)), nil
}
