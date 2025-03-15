package tools

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
)

// handleReadFile handles the read_file tool
func handleReadFile(request mcp.CallToolRequest, allowedDirectories []string) (*mcp.CallToolResult, error) {
	path, ok := request.Params.Arguments["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path must be a string")
	}

	// Validate path
	validPath, err := ValidatePath(path, allowedDirectories)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %v", err)
	}

	// Read file
	// #nosec G304 -- validPath has been validated by ValidatePath
	content, err := os.ReadFile(validPath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	return mcp.NewToolResultText(string(content)), nil
}

// handleWriteFile handles the write_file tool
func handleWriteFile(request mcp.CallToolRequest, allowedDirectories []string) (*mcp.CallToolResult, error) {
	path, ok := request.Params.Arguments["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path must be a string")
	}

	content, ok := request.Params.Arguments["content"].(string)
	if !ok {
		return nil, fmt.Errorf("content must be a string")
	}

	append, _ := request.Params.Arguments["append"].(bool)

	// Validate path
	validPath, err := ValidatePath(path, allowedDirectories)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %v", err)
	}

	// Create parent directories if they don't exist
	dir := filepath.Dir(validPath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return nil, fmt.Errorf("error creating directories: %v", err)
	}

	// Write file
	var flag int
	if append {
		flag = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	} else {
		flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	}

	// #nosec G304 -- validPath has been validated by ValidatePath
	file, err := os.OpenFile(validPath, flag, 0600)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return nil, fmt.Errorf("error writing to file: %v", err)
	}

	action := "Written"
	if append {
		action = "Appended"
	}

	return mcp.NewToolResultText(fmt.Sprintf("%s %d bytes to %s", action, len(content), path)), nil
}
