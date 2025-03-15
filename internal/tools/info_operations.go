package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// handleGetFileInfo handles the get_file_info tool
func handleGetFileInfo(request mcp.CallToolRequest, allowedDirectories []string) (*mcp.CallToolResult, error) {
	path, ok := request.Params.Arguments["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path must be a string")
	}

	// Validate path
	validPath, err := ValidatePath(path, allowedDirectories)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %v", err)
	}

	// Get file info
	info, err := os.Stat(validPath)
	if err != nil {
		return nil, fmt.Errorf("error getting file info: %v", err)
	}

	// Create file info
	fileInfo := FileInfo{
		Name:    filepath.Base(validPath),
		Path:    path,
		Size:    info.Size(),
		IsDir:   info.IsDir(),
		ModTime: info.ModTime().Format(time.RFC3339),
	}

	// Add extension if it's a file
	if !info.IsDir() {
		fileInfo.Extension = filepath.Ext(validPath)
	}

	// Format results
	fileInfoJSON, err := json.MarshalIndent(fileInfo, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error formatting file info: %v", err)
	}

	return mcp.NewToolResultText(fmt.Sprintf("File info for %s:\n\n%s", path, string(fileInfoJSON))), nil
}

// handleListAllowedDirectories handles the list_allowed_directories tool
func handleListAllowedDirectories(request mcp.CallToolRequest, allowedDirectories []string) (*mcp.CallToolResult, error) {
	var resultBuilder strings.Builder
	resultBuilder.WriteString("Allowed directories:\n\n")
	for _, dir := range allowedDirectories {
		resultBuilder.WriteString(fmt.Sprintf("%s\n", dir))
	}
	return mcp.NewToolResultText(resultBuilder.String()), nil
}
