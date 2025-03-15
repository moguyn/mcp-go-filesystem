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

// handleCreateDirectory handles the create_directory tool
func handleCreateDirectory(request mcp.CallToolRequest, allowedDirectories []string) (*mcp.CallToolResult, error) {
	path, ok := request.Params.Arguments["path"].(string)
	if !ok {
		return mcp.NewToolResultError("path must be a string"), nil
	}

	// Validate path
	validPath, err := ValidatePath(path, allowedDirectories)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid path: %v", err)), nil
	}

	// Create directory
	if err := os.MkdirAll(validPath, 0750); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error creating directory: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Created directory %s", path)), nil
}

// handleListDirectory handles the list_directory tool
func handleListDirectory(request mcp.CallToolRequest, allowedDirectories []string) (*mcp.CallToolResult, error) {
	path, ok := request.Params.Arguments["path"].(string)
	if !ok {
		return mcp.NewToolResultError("path must be a string"), nil
	}

	// Validate path
	validPath, err := ValidatePath(path, allowedDirectories)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid path: %v", err)), nil
	}

	// Read directory
	entries, err := os.ReadDir(validPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error reading directory: %v", err)), nil
	}

	// Format results
	var resultBuilder strings.Builder
	resultBuilder.WriteString(fmt.Sprintf("Contents of %s:\n\n", path))

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		entryType := "file"
		if entry.IsDir() {
			entryType = "directory"
		}

		resultBuilder.WriteString(fmt.Sprintf("%s (%s, %d bytes, modified %s)\n",
			entry.Name(),
			entryType,
			info.Size(),
			info.ModTime().Format(time.RFC3339),
		))
	}

	return mcp.NewToolResultText(resultBuilder.String()), nil
}

// handleDirectoryTree handles the directory_tree tool
func handleDirectoryTree(request mcp.CallToolRequest, allowedDirectories []string) (*mcp.CallToolResult, error) {
	path, ok := request.Params.Arguments["path"].(string)
	if !ok {
		return mcp.NewToolResultError("path must be a string"), nil
	}

	maxDepth := 3 // Default max depth
	if maxDepthRaw, ok := request.Params.Arguments["max_depth"].(float64); ok {
		maxDepth = int(maxDepthRaw)
	}

	// Validate path
	validPath, err := ValidatePath(path, allowedDirectories)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid path: %v", err)), nil
	}

	// Check if path is a directory
	info, err := os.Stat(validPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error accessing path: %v", err)), nil
	}
	if !info.IsDir() {
		return mcp.NewToolResultError("Path is not a directory"), nil
	}

	// Build directory tree
	tree, err := buildDirectoryTree(validPath, maxDepth)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error building directory tree: %v", err)), nil
	}

	// Format results
	treeJSON, err := json.MarshalIndent(tree, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error formatting directory tree: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Directory tree for %s:\n\n%s", path, string(treeJSON))), nil
}

// buildDirectoryTree builds a directory tree
func buildDirectoryTree(rootPath string, maxDepth int) ([]TreeEntry, error) {
	if maxDepth <= 0 {
		return []TreeEntry{}, nil
	}

	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	result := make([]TreeEntry, 0, len(entries))
	for _, entry := range entries {
		entryType := "file"
		var children []TreeEntry

		if entry.IsDir() {
			entryType = "directory"
			if maxDepth > 1 {
				children, err = buildDirectoryTree(filepath.Join(rootPath, entry.Name()), maxDepth-1)
				if err != nil {
					// Log the error but continue with other entries
					children = []TreeEntry{}
				}
			} else {
				children = []TreeEntry{}
			}
		}

		result = append(result, TreeEntry{
			Name:     entry.Name(),
			Type:     entryType,
			Children: children,
		})
	}

	return result, nil
}
