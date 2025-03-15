package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

// handleReadMultipleFiles handles the read_multiple_files tool
func handleReadMultipleFiles(request mcp.CallToolRequest, allowedDirectories []string) (*mcp.CallToolResult, error) {
	pathsJSON, ok := request.Params.Arguments["paths"].(string)
	if !ok {
		return nil, fmt.Errorf("paths must be a JSON string array")
	}

	// Parse JSON array
	var paths []string
	if err := json.Unmarshal([]byte(pathsJSON), &paths); err != nil {
		return nil, fmt.Errorf("invalid JSON array: %v", err)
	}

	// Read each file
	contents := make([]interface{}, 0, len(paths))
	hasError := false
	for _, path := range paths {
		// Validate path
		validPath, err := ValidatePath(path, allowedDirectories)
		if err != nil {
			hasError = true
			contents = append(contents, &mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Error with %s: Invalid path: %v", path, err),
			})
			continue
		}

		// Read file
		// #nosec G304 -- validPath has been validated by ValidatePath
		content, err := os.ReadFile(validPath)
		if err != nil {
			hasError = true
			contents = append(contents, &mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Error reading %s: %v", path, err),
			})
			continue
		}

		contents = append(contents, &mcp.TextContent{
			Type: "text",
			Text: string(content),
		})
	}

	result := &mcp.CallToolResult{
		Content: contents,
		IsError: hasError,
	}
	return result, nil
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

// handleEditFile handles the edit_file tool
func handleEditFile(request mcp.CallToolRequest, allowedDirectories []string) (*mcp.CallToolResult, error) {
	path, ok := request.Params.Arguments["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path must be a string")
	}

	oldText, ok := request.Params.Arguments["old_text"].(string)
	if !ok {
		return nil, fmt.Errorf("old_text must be a string")
	}

	newText, ok := request.Params.Arguments["new_text"].(string)
	if !ok {
		return nil, fmt.Errorf("new_text must be a string")
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

	// Replace text
	newContent := strings.Replace(string(content), oldText, newText, -1)
	if newContent == string(content) {
		return nil, fmt.Errorf("text '%s' not found in file", oldText)
	}

	// Write file
	if err := os.WriteFile(validPath, []byte(newContent), 0600); err != nil {
		return nil, fmt.Errorf("error writing file: %v", err)
	}

	return mcp.NewToolResultText(fmt.Sprintf("Edited file %s: replaced '%s' with '%s'", path, oldText, newText)), nil
}
