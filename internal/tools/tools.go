package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// FileInfo represents metadata about a file or directory
type FileInfo struct {
	Size        int64  `json:"size"`
	Created     string `json:"created"`
	Modified    string `json:"modified"`
	Accessed    string `json:"accessed"`
	IsDirectory bool   `json:"isDirectory"`
	IsFile      bool   `json:"isFile"`
	Permissions string `json:"permissions"`
}

// TreeEntry represents an entry in a directory tree
type TreeEntry struct {
	Name     string      `json:"name"`
	Type     string      `json:"type"` // "file" or "directory"
	Children []TreeEntry `json:"children,omitempty"`
}

// RegisterTools registers all filesystem tools with the MCP server
func RegisterTools(s *server.MCPServer, allowedDirectories []string) {
	// Register read_file tool
	readFileTool := mcp.NewTool("read_file",
		mcp.WithDescription("Read the complete contents of a file from the file system. "+
			"Handles various text encodings and provides detailed error messages "+
			"if the file cannot be read. Use this tool when you need to examine "+
			"the contents of a single file. Only works within allowed directories."),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Path to the file to read"),
		),
	)
	s.AddTool(readFileTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleReadFile(request, allowedDirectories)
	})

	// Register read_multiple_files tool
	readMultipleFilesTool := mcp.NewTool("read_multiple_files",
		mcp.WithDescription("Read the contents of multiple files simultaneously. This is more "+
			"efficient than reading files one by one when you need to analyze "+
			"or compare multiple files. Each file's content is returned with its "+
			"path as a reference. Failed reads for individual files won't stop "+
			"the entire operation. Only works within allowed directories."),
		mcp.WithString("paths",
			mcp.Required(),
			mcp.Description("JSON array of paths to the files to read"),
		),
	)
	s.AddTool(readMultipleFilesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleReadMultipleFiles(request, allowedDirectories)
	})

	// Register write_file tool
	writeFileTool := mcp.NewTool("write_file",
		mcp.WithDescription("Write content to a file on the file system. Creates the file if it "+
			"doesn't exist, or overwrites it if it does. Can optionally append to "+
			"existing files instead of overwriting. Only works within allowed directories."),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Path to the file to write"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("Content to write to the file"),
		),
		mcp.WithBoolean("append",
			mcp.Description("Whether to append to the file instead of overwriting it"),
		),
	)
	s.AddTool(writeFileTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleWriteFile(request, allowedDirectories)
	})

	// Register create_directory tool
	createDirectoryTool := mcp.NewTool("create_directory",
		mcp.WithDescription("Create a new directory on the file system. Creates parent directories "+
			"as needed. Only works within allowed directories."),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Path to the directory to create"),
		),
	)
	s.AddTool(createDirectoryTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleCreateDirectory(request, allowedDirectories)
	})

	// Register list_directory tool
	listDirectoryTool := mcp.NewTool("list_directory",
		mcp.WithDescription("List the contents of a directory on the file system. Returns a list of "+
			"files and directories with their metadata. Only works within allowed directories."),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Path to the directory to list"),
		),
	)
	s.AddTool(listDirectoryTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleListDirectory(request, allowedDirectories)
	})

	// Register directory_tree tool
	directoryTreeTool := mcp.NewTool("directory_tree",
		mcp.WithDescription("Get a hierarchical tree representation of a directory and its contents. "+
			"Useful for understanding the structure of a project or codebase. "+
			"Only works within allowed directories."),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Path to the root directory"),
		),
		mcp.WithNumber("max_depth",
			mcp.Description("Maximum depth to traverse (default: 3)"),
		),
	)
	s.AddTool(directoryTreeTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleDirectoryTree(request, allowedDirectories)
	})

	// Register move_file tool
	moveFileTool := mcp.NewTool("move_file",
		mcp.WithDescription("Move or rename a file or directory on the file system. "+
			"Only works within allowed directories."),
		mcp.WithString("source",
			mcp.Required(),
			mcp.Description("Path to the source file or directory"),
		),
		mcp.WithString("destination",
			mcp.Required(),
			mcp.Description("Path to the destination file or directory"),
		),
	)
	s.AddTool(moveFileTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleMoveFile(request, allowedDirectories)
	})

	// Register search_files tool
	searchFilesTool := mcp.NewTool("search_files",
		mcp.WithDescription("Search for files matching a pattern within a directory. "+
			"Supports glob patterns and can exclude files matching certain patterns. "+
			"Only works within allowed directories."),
		mcp.WithString("root",
			mcp.Required(),
			mcp.Description("Root directory to search in"),
		),
		mcp.WithString("pattern",
			mcp.Required(),
			mcp.Description("Glob pattern to match files against"),
		),
		mcp.WithString("exclude",
			mcp.Description("JSON array of patterns to exclude from the search"),
		),
	)
	s.AddTool(searchFilesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleSearchFiles(request, allowedDirectories)
	})

	// Register get_file_info tool
	getFileInfoTool := mcp.NewTool("get_file_info",
		mcp.WithDescription("Get detailed metadata about a file or directory, including size, "+
			"creation time, modification time, and permissions. "+
			"Only works within allowed directories."),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Path to the file or directory"),
		),
	)
	s.AddTool(getFileInfoTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleGetFileInfo(request, allowedDirectories)
	})

	// Register list_allowed_directories tool
	listAllowedDirectoriesTool := mcp.NewTool("list_allowed_directories",
		mcp.WithDescription("List all directories that the server is allowed to access."),
		mcp.WithString("properties", mcp.Description("Empty properties object")),
	)
	s.AddTool(listAllowedDirectoriesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleListAllowedDirectories(request, allowedDirectories)
	})

	// Register edit_file tool
	editFileTool := mcp.NewTool("edit_file",
		mcp.WithDescription("Edit a file by replacing text. Allows for precise edits without "+
			"having to rewrite the entire file. Only works within allowed directories."),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Path to the file to edit"),
		),
		mcp.WithString("old_text",
			mcp.Required(),
			mcp.Description("Text to replace"),
		),
		mcp.WithString("new_text",
			mcp.Required(),
			mcp.Description("Text to replace with"),
		),
	)
	s.AddTool(editFileTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleEditFile(request, allowedDirectories)
	})
}

// handleReadFile handles the read_file tool
func handleReadFile(request mcp.CallToolRequest, allowedDirectories []string) (*mcp.CallToolResult, error) {
	path, ok := request.Params.Arguments["path"].(string)
	if !ok {
		return mcp.NewToolResultError("path must be a string"), nil
	}

	// Validate path
	validPath, err := ValidatePath(path, allowedDirectories)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid path: %v", err)), nil
	}

	// Read file
	// #nosec G304 -- validPath has been validated by ValidatePath
	content, err := os.ReadFile(validPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error reading file: %v", err)), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

// handleReadMultipleFiles handles the read_multiple_files tool
func handleReadMultipleFiles(request mcp.CallToolRequest, allowedDirectories []string) (*mcp.CallToolResult, error) {
	pathsJSON, ok := request.Params.Arguments["paths"].(string)
	if !ok {
		return mcp.NewToolResultError("paths must be a JSON string array"), nil
	}

	// Parse JSON array
	var paths []string
	if err := json.Unmarshal([]byte(pathsJSON), &paths); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid JSON array: %v", err)), nil
	}

	// Read each file
	var resultBuilder strings.Builder
	for _, path := range paths {
		// Validate path
		validPath, err := ValidatePath(path, allowedDirectories)
		if err != nil {
			resultBuilder.WriteString(fmt.Sprintf("Error with %s: Invalid path: %v\n\n", path, err))
			continue
		}

		// Read file
		// #nosec G304 -- validPath has been validated by ValidatePath
		content, err := os.ReadFile(validPath)
		if err != nil {
			resultBuilder.WriteString(fmt.Sprintf("Error reading %s: %v\n\n", path, err))
			continue
		}

		resultBuilder.WriteString(fmt.Sprintf("=== %s ===\n%s\n\n", path, string(content)))
	}

	return mcp.NewToolResultText(resultBuilder.String()), nil
}

// handleWriteFile handles the write_file tool
func handleWriteFile(request mcp.CallToolRequest, allowedDirectories []string) (*mcp.CallToolResult, error) {
	path, ok := request.Params.Arguments["path"].(string)
	if !ok {
		return mcp.NewToolResultError("path must be a string"), nil
	}

	content, ok := request.Params.Arguments["content"].(string)
	if !ok {
		return mcp.NewToolResultError("content must be a string"), nil
	}

	append, _ := request.Params.Arguments["append"].(bool)

	// Validate path
	validPath, err := ValidatePath(path, allowedDirectories)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid path: %v", err)), nil
	}

	// Create parent directories if they don't exist
	dir := filepath.Dir(validPath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error creating directories: %v", err)), nil
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
		return mcp.NewToolResultError(fmt.Sprintf("Error opening file: %v", err)), nil
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error writing to file: %v", err)), nil
	}

	action := "Written"
	if append {
		action = "Appended"
	}

	return mcp.NewToolResultText(fmt.Sprintf("%s %d bytes to %s", action, len(content), path)), nil
}

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

	// Build FileInfo array
	fileInfos := make([]FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		fileInfo := FileInfo{
			Size:        info.Size(),
			Created:     time.Now().Format(time.RFC3339), // os.FileInfo doesn't provide creation time
			Modified:    info.ModTime().Format(time.RFC3339),
			Accessed:    time.Now().Format(time.RFC3339), // os.FileInfo doesn't provide access time
			IsDirectory: entry.IsDir(),
			IsFile:      !entry.IsDir(),
			Permissions: info.Mode().String(),
		}
		fileInfos = append(fileInfos, fileInfo)
	}

	// Convert to JSON
	jsonData, err := json.Marshal(fileInfos)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error formatting directory listing: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
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
		return nil, nil
	}

	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, err
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
					// Skip this directory if there's an error
					continue
				}
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

// handleMoveFile handles the move_file tool
func handleMoveFile(request mcp.CallToolRequest, allowedDirectories []string) (*mcp.CallToolResult, error) {
	source, ok := request.Params.Arguments["source"].(string)
	if !ok {
		return mcp.NewToolResultError("source must be a string"), nil
	}

	destination, ok := request.Params.Arguments["destination"].(string)
	if !ok {
		return mcp.NewToolResultError("destination must be a string"), nil
	}

	// Validate paths
	validSource, err := ValidatePath(source, allowedDirectories)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid source path: %v", err)), nil
	}

	validDestination, err := ValidatePath(destination, allowedDirectories)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid destination path: %v", err)), nil
	}

	// Create parent directories if they don't exist
	destDir := filepath.Dir(validDestination)
	if err := os.MkdirAll(destDir, 0750); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error creating destination directories: %v", err)), nil
	}

	// Move file
	if err := os.Rename(validSource, validDestination); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error moving file: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Moved %s to %s", source, destination)), nil
}

// handleSearchFiles handles the search_files tool
func handleSearchFiles(request mcp.CallToolRequest, allowedDirectories []string) (*mcp.CallToolResult, error) {
	root, ok := request.Params.Arguments["root"].(string)
	if !ok {
		return mcp.NewToolResultError("root must be a string"), nil
	}

	pattern, ok := request.Params.Arguments["pattern"].(string)
	if !ok {
		return mcp.NewToolResultError("pattern must be a string"), nil
	}

	// Parse exclude patterns
	var excludePatterns []string
	if excludeJSON, ok := request.Params.Arguments["exclude"].(string); ok && excludeJSON != "" {
		if err := json.Unmarshal([]byte(excludeJSON), &excludePatterns); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid exclude JSON array: %v", err)), nil
		}
	}

	// Validate path
	validRoot, err := ValidatePath(root, allowedDirectories)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid root path: %v", err)), nil
	}

	// Search files
	matches, err := searchFiles(validRoot, pattern, excludePatterns)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error searching files: %v", err)), nil
	}

	// Format results
	var resultBuilder strings.Builder
	resultBuilder.WriteString(fmt.Sprintf("Found %d matches for pattern '%s' in %s:\n\n", len(matches), pattern, root))
	for _, match := range matches {
		resultBuilder.WriteString(fmt.Sprintf("%s\n", match))
	}

	return mcp.NewToolResultText(resultBuilder.String()), nil
}

// searchFiles searches for files matching a pattern
func searchFiles(rootPath, pattern string, excludePatterns []string) ([]string, error) {
	var matches []string

	// Compile exclude patterns
	var excludeRegexps []*regexp.Regexp
	for _, excludePattern := range excludePatterns {
		// Convert glob pattern to regexp
		regexpPattern := "^" + strings.ReplaceAll(strings.ReplaceAll(excludePattern, ".", "\\."), "*", ".*") + "$"
		re, err := regexp.Compile(regexpPattern)
		if err != nil {
			continue
		}
		excludeRegexps = append(excludeRegexps, re)
	}

	// Walk directory
	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Check if file matches pattern
		match, err := filepath.Match(pattern, filepath.Base(path))
		if err != nil {
			return nil // Skip invalid patterns
		}

		if match {
			// Check if file is excluded
			relPath, err := filepath.Rel(rootPath, path)
			if err != nil {
				return nil
			}

			excluded := false
			for _, re := range excludeRegexps {
				if re.MatchString(relPath) || re.MatchString(filepath.Base(path)) {
					excluded = true
					break
				}
			}

			if !excluded {
				matches = append(matches, path)
			}
		}

		return nil
	})

	return matches, err
}

// handleGetFileInfo handles the get_file_info tool
func handleGetFileInfo(request mcp.CallToolRequest, allowedDirectories []string) (*mcp.CallToolResult, error) {
	path, ok := request.Params.Arguments["path"].(string)
	if !ok {
		return mcp.NewToolResultError("path must be a string"), nil
	}

	// Validate path
	validPath, err := ValidatePath(path, allowedDirectories)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid path: %v", err)), nil
	}

	// Get file info
	info, err := os.Stat(validPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error getting file info: %v", err)), nil
	}

	// Format permissions
	perms := info.Mode().String()

	// Get file times - use modification time as fallback for created/accessed
	created := info.ModTime().Format(time.RFC3339)
	accessed := info.ModTime().Format(time.RFC3339)

	// Create file info
	fileInfo := FileInfo{
		Size:        info.Size(),
		Created:     created,
		Modified:    info.ModTime().Format(time.RFC3339),
		Accessed:    accessed,
		IsDirectory: info.IsDir(),
		IsFile:      !info.IsDir(),
		Permissions: perms,
	}

	// Format results
	fileInfoJSON, err := json.MarshalIndent(fileInfo, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error formatting file info: %v", err)), nil
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

// handleEditFile handles the edit_file tool
func handleEditFile(request mcp.CallToolRequest, allowedDirectories []string) (*mcp.CallToolResult, error) {
	path, ok := request.Params.Arguments["path"].(string)
	if !ok {
		return mcp.NewToolResultError("path must be a string"), nil
	}

	oldText, ok := request.Params.Arguments["old_text"].(string)
	if !ok {
		return mcp.NewToolResultError("old_text must be a string"), nil
	}

	newText, ok := request.Params.Arguments["new_text"].(string)
	if !ok {
		return mcp.NewToolResultError("new_text must be a string"), nil
	}

	// Validate path
	validPath, err := ValidatePath(path, allowedDirectories)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid path: %v", err)), nil
	}

	// Read file
	// #nosec G304 -- validPath has been validated by ValidatePath
	content, err := os.ReadFile(validPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error reading file: %v", err)), nil
	}

	// Replace text
	newContent := strings.Replace(string(content), oldText, newText, -1)
	if newContent == string(content) {
		return mcp.NewToolResultError(fmt.Sprintf("Text '%s' not found in file", oldText)), nil
	}

	// Write file
	if err := os.WriteFile(validPath, []byte(newContent), 0600); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error writing file: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Edited file %s: replaced '%s' with '%s'", path, oldText, newText)), nil
}
