package server

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// handleListTools handles the mcp.list_tools request
func (s *Server) handleListTools(id string) error {
	tools := []Tool{
		{
			Name: "read_file",
			Description: "Read the complete contents of a file from the file system. " +
				"Handles various text encodings and provides detailed error messages " +
				"if the file cannot be read. Use this tool when you need to examine " +
				"the contents of a single file. Only works within allowed directories.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the file to read",
					},
				},
				"required": []string{"path"},
			},
		},
		{
			Name: "read_multiple_files",
			Description: "Read the contents of multiple files simultaneously. This is more " +
				"efficient than reading files one by one when you need to analyze " +
				"or compare multiple files. Each file's content is returned with its " +
				"path as a reference. Failed reads for individual files won't stop " +
				"the entire operation. Only works within allowed directories.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"paths": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Paths to the files to read",
					},
				},
				"required": []string{"paths"},
			},
		},
		{
			Name: "write_file",
			Description: "Create a new file or completely overwrite an existing file with new content. " +
				"Use with caution as it will overwrite existing files without warning. " +
				"Handles text content with proper encoding. Only works within allowed directories.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the file to write",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Content to write to the file",
					},
				},
				"required": []string{"path", "content"},
			},
		},
		{
			Name: "edit_file",
			Description: "Make line-based edits to a text file. Each edit replaces exact line sequences " +
				"with new content. Returns a git-style diff showing the changes made. " +
				"Only works within allowed directories.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the file to edit",
					},
					"edits": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"oldText": map[string]interface{}{
									"type":        "string",
									"description": "Text to search for - must match exactly",
								},
								"newText": map[string]interface{}{
									"type":        "string",
									"description": "Text to replace with",
								},
							},
							"required": []string{"oldText", "newText"},
						},
						"description": "List of edit operations to perform",
					},
					"dryRun": map[string]interface{}{
						"type":        "boolean",
						"description": "Preview changes using git-style diff format",
						"default":     false,
					},
				},
				"required": []string{"path", "edits"},
			},
		},
		{
			Name: "create_directory",
			Description: "Create a new directory or ensure a directory exists. Can create multiple " +
				"nested directories in one operation. If the directory already exists, " +
				"this operation will succeed silently. Perfect for setting up directory " +
				"structures for projects or ensuring required paths exist. Only works within allowed directories.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the directory to create",
					},
				},
				"required": []string{"path"},
			},
		},
		{
			Name: "list_directory",
			Description: "Get a detailed listing of all files and directories in a specified path. " +
				"Results clearly distinguish between files and directories with [FILE] and [DIR] " +
				"prefixes. This tool is essential for understanding directory structure and " +
				"finding specific files within a directory. Only works within allowed directories.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the directory to list",
					},
				},
				"required": []string{"path"},
			},
		},
		{
			Name: "directory_tree",
			Description: "Get a recursive tree view of files and directories as a JSON structure. " +
				"Each entry includes 'name', 'type' (file/directory), and 'children' for directories. " +
				"Files have no children array, while directories always have a children array (which may be empty). " +
				"The output is formatted with 2-space indentation for readability. Only works within allowed directories.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the directory to get tree for",
					},
				},
				"required": []string{"path"},
			},
		},
		{
			Name: "move_file",
			Description: "Move or rename files and directories. Can move files between directories " +
				"and rename them in a single operation. If the destination exists, the " +
				"operation will fail. Works across different directories and can be used " +
				"for simple renaming within the same directory. Both source and destination must be within allowed directories.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"source": map[string]interface{}{
						"type":        "string",
						"description": "Path to the source file or directory",
					},
					"destination": map[string]interface{}{
						"type":        "string",
						"description": "Path to the destination file or directory",
					},
				},
				"required": []string{"source", "destination"},
			},
		},
		{
			Name: "search_files",
			Description: "Recursively search for files and directories matching a pattern. " +
				"Searches through all subdirectories from the starting path. The search " +
				"is case-insensitive and matches partial names. Returns full paths to all " +
				"matching items. Great for finding files when you don't know their exact location. " +
				"Only searches within allowed directories.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the directory to search in",
					},
					"pattern": map[string]interface{}{
						"type":        "string",
						"description": "Pattern to search for",
					},
					"excludePatterns": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Patterns to exclude from search",
						"default":     []string{},
					},
				},
				"required": []string{"path", "pattern"},
			},
		},
		{
			Name: "get_file_info",
			Description: "Retrieve detailed metadata about a file or directory. Returns comprehensive " +
				"information including size, creation time, last modified time, permissions, " +
				"and type. This tool is perfect for understanding file characteristics " +
				"without reading the actual content. Only works within allowed directories.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the file or directory to get info for",
					},
				},
				"required": []string{"path"},
			},
		},
		{
			Name: "list_allowed_directories",
			Description: "Returns the list of directories that this server is allowed to access. " +
				"Use this to understand which directories are available before trying to access files.",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
				"required":   []string{},
			},
		},
	}

	return s.sendResponse(id, ListToolsResponse{Tools: tools})
}

// handleCallTool handles the mcp.call_tool request
func (s *Server) handleCallTool(id string, params map[string]interface{}) error {
	toolName, ok := params["name"].(string)
	if !ok {
		return s.sendErrorResponse("missing or invalid tool name")
	}

	args, ok := params["arguments"].(map[string]interface{})
	if !ok {
		args = map[string]interface{}{}
	}

	var response ToolResponse
	var err error

	switch toolName {
	case "read_file":
		response, err = s.handleReadFile(args)
	case "read_multiple_files":
		response, err = s.handleReadMultipleFiles(args)
	case "write_file":
		response, err = s.handleWriteFile(args)
	case "edit_file":
		response, err = s.handleEditFile(args)
	case "create_directory":
		response, err = s.handleCreateDirectory(args)
	case "list_directory":
		response, err = s.handleListDirectory(args)
	case "directory_tree":
		response, err = s.handleDirectoryTree(args)
	case "move_file":
		response, err = s.handleMoveFile(args)
	case "search_files":
		response, err = s.handleSearchFiles(args)
	case "get_file_info":
		response, err = s.handleGetFileInfo(args)
	case "list_allowed_directories":
		response, err = s.handleListAllowedDirectories(args)
	default:
		return s.sendErrorResponse(fmt.Sprintf("unknown tool: %s", toolName))
	}

	if err != nil {
		response = ToolResponse{
			Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
			IsError: true,
		}
	}

	return s.sendResponse(id, response)
}

// handleReadFile handles the read_file tool
func (s *Server) handleReadFile(args map[string]interface{}) (ToolResponse, error) {
	path, ok := args["path"].(string)
	if !ok {
		return ToolResponse{}, fmt.Errorf("missing or invalid path argument")
	}

	validPath, err := s.ValidatePath(path)
	if err != nil {
		return ToolResponse{}, err
	}

	content, err := os.ReadFile(validPath)
	if err != nil {
		return ToolResponse{}, fmt.Errorf("error reading file: %w", err)
	}

	return ToolResponse{
		Content: []ContentItem{{Type: "text", Text: string(content)}},
	}, nil
}

// handleReadMultipleFiles handles the read_multiple_files tool
func (s *Server) handleReadMultipleFiles(args map[string]interface{}) (ToolResponse, error) {
	pathsInterface, ok := args["paths"].([]interface{})
	if !ok {
		return ToolResponse{}, fmt.Errorf("missing or invalid paths argument")
	}

	paths := make([]string, len(pathsInterface))
	for i, p := range pathsInterface {
		path, ok := p.(string)
		if !ok {
			return ToolResponse{}, fmt.Errorf("invalid path at index %d", i)
		}
		paths[i] = path
	}

	var results []string
	for _, path := range paths {
		result := path + ":\n"
		validPath, err := s.ValidatePath(path)
		if err != nil {
			results = append(results, result+"Error - "+err.Error())
			continue
		}

		content, err := os.ReadFile(validPath)
		if err != nil {
			results = append(results, result+"Error - "+err.Error())
			continue
		}

		results = append(results, result+string(content)+"\n")
	}

	return ToolResponse{
		Content: []ContentItem{{Type: "text", Text: strings.Join(results, "\n---\n")}},
	}, nil
}

// handleWriteFile handles the write_file tool
func (s *Server) handleWriteFile(args map[string]interface{}) (ToolResponse, error) {
	path, ok := args["path"].(string)
	if !ok {
		return ToolResponse{}, fmt.Errorf("missing or invalid path argument")
	}

	content, ok := args["content"].(string)
	if !ok {
		return ToolResponse{}, fmt.Errorf("missing or invalid content argument")
	}

	validPath, err := s.ValidatePath(path)
	if err != nil {
		return ToolResponse{}, err
	}

	// Ensure parent directory exists
	dir := filepath.Dir(validPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return ToolResponse{}, fmt.Errorf("error creating parent directory: %w", err)
	}

	if err := os.WriteFile(validPath, []byte(content), 0644); err != nil {
		return ToolResponse{}, fmt.Errorf("error writing file: %w", err)
	}

	return ToolResponse{
		Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Successfully wrote to %s", path)}},
	}, nil
}

// handleCreateDirectory handles the create_directory tool
func (s *Server) handleCreateDirectory(args map[string]interface{}) (ToolResponse, error) {
	path, ok := args["path"].(string)
	if !ok {
		return ToolResponse{}, fmt.Errorf("missing or invalid path argument")
	}

	validPath, err := s.ValidatePath(path)
	if err != nil {
		return ToolResponse{}, err
	}

	if err := os.MkdirAll(validPath, 0755); err != nil {
		return ToolResponse{}, fmt.Errorf("error creating directory: %w", err)
	}

	return ToolResponse{
		Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Successfully created directory %s", path)}},
	}, nil
}

// handleListDirectory handles the list_directory tool
func (s *Server) handleListDirectory(args map[string]interface{}) (ToolResponse, error) {
	path, ok := args["path"].(string)
	if !ok {
		return ToolResponse{}, fmt.Errorf("missing or invalid path argument")
	}

	validPath, err := s.ValidatePath(path)
	if err != nil {
		return ToolResponse{}, err
	}

	entries, err := os.ReadDir(validPath)
	if err != nil {
		return ToolResponse{}, fmt.Errorf("error reading directory: %w", err)
	}

	var formatted []string
	for _, entry := range entries {
		prefix := "[FILE]"
		if entry.IsDir() {
			prefix = "[DIR]"
		}
		formatted = append(formatted, fmt.Sprintf("%s %s", prefix, entry.Name()))
	}

	return ToolResponse{
		Content: []ContentItem{{Type: "text", Text: strings.Join(formatted, "\n")}},
	}, nil
}

// handleDirectoryTree handles the directory_tree tool
func (s *Server) handleDirectoryTree(args map[string]interface{}) (ToolResponse, error) {
	path, ok := args["path"].(string)
	if !ok {
		return ToolResponse{}, fmt.Errorf("missing or invalid path argument")
	}

	validPath, err := s.ValidatePath(path)
	if err != nil {
		return ToolResponse{}, err
	}

	tree, err := s.buildDirectoryTree(validPath)
	if err != nil {
		return ToolResponse{}, err
	}

	treeJSON, err := json.MarshalIndent(tree, "", "  ")
	if err != nil {
		return ToolResponse{}, fmt.Errorf("error marshaling tree to JSON: %w", err)
	}

	return ToolResponse{
		Content: []ContentItem{{Type: "text", Text: string(treeJSON)}},
	}, nil
}

// buildDirectoryTree builds a tree representation of a directory
func (s *Server) buildDirectoryTree(path string) ([]TreeEntry, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("error reading directory: %w", err)
	}

	var result []TreeEntry
	for _, entry := range entries {
		entryType := "file"
		if entry.IsDir() {
			entryType = "directory"
		}

		treeEntry := TreeEntry{
			Name: entry.Name(),
			Type: entryType,
		}

		if entry.IsDir() {
			subPath := filepath.Join(path, entry.Name())
			children, err := s.buildDirectoryTree(subPath)
			if err != nil {
				// Skip directories we can't read
				continue
			}
			treeEntry.Children = children
		}

		result = append(result, treeEntry)
	}

	return result, nil
}

// handleMoveFile handles the move_file tool
func (s *Server) handleMoveFile(args map[string]interface{}) (ToolResponse, error) {
	source, ok := args["source"].(string)
	if !ok {
		return ToolResponse{}, fmt.Errorf("missing or invalid source argument")
	}

	destination, ok := args["destination"].(string)
	if !ok {
		return ToolResponse{}, fmt.Errorf("missing or invalid destination argument")
	}

	validSourcePath, err := s.ValidatePath(source)
	if err != nil {
		return ToolResponse{}, err
	}

	validDestPath, err := s.ValidatePath(destination)
	if err != nil {
		return ToolResponse{}, err
	}

	// Ensure parent directory of destination exists
	destDir := filepath.Dir(validDestPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return ToolResponse{}, fmt.Errorf("error creating destination parent directory: %w", err)
	}

	if err := os.Rename(validSourcePath, validDestPath); err != nil {
		return ToolResponse{}, fmt.Errorf("error moving file: %w", err)
	}

	return ToolResponse{
		Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Successfully moved %s to %s", source, destination)}},
	}, nil
}

// handleSearchFiles handles the search_files tool
func (s *Server) handleSearchFiles(args map[string]interface{}) (ToolResponse, error) {
	path, ok := args["path"].(string)
	if !ok {
		return ToolResponse{}, fmt.Errorf("missing or invalid path argument")
	}

	pattern, ok := args["pattern"].(string)
	if !ok {
		return ToolResponse{}, fmt.Errorf("missing or invalid pattern argument")
	}

	var excludePatterns []string
	if excludePatternsInterface, ok := args["excludePatterns"].([]interface{}); ok {
		for _, p := range excludePatternsInterface {
			if patternStr, ok := p.(string); ok {
				excludePatterns = append(excludePatterns, patternStr)
			}
		}
	}

	validPath, err := s.ValidatePath(path)
	if err != nil {
		return ToolResponse{}, err
	}

	results, err := s.searchFiles(validPath, pattern, excludePatterns)
	if err != nil {
		return ToolResponse{}, err
	}

	var responseText string
	if len(results) > 0 {
		responseText = strings.Join(results, "\n")
	} else {
		responseText = "No matches found"
	}

	return ToolResponse{
		Content: []ContentItem{{Type: "text", Text: responseText}},
	}, nil
}

// searchFiles searches for files matching a pattern
func (s *Server) searchFiles(rootPath, pattern string, excludePatterns []string) ([]string, error) {
	pattern = strings.ToLower(pattern)
	var results []string

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Check if path is excluded
		relativePath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return nil // Skip if we can't get relative path
		}

		for _, excludePattern := range excludePatterns {
			matched, err := filepath.Match(excludePattern, relativePath)
			if err == nil && matched {
				if d.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
		}

		// Check if name matches pattern
		if strings.Contains(strings.ToLower(d.Name()), pattern) {
			results = append(results, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error searching files: %w", err)
	}

	return results, nil
}

// handleGetFileInfo handles the get_file_info tool
func (s *Server) handleGetFileInfo(args map[string]interface{}) (ToolResponse, error) {
	path, ok := args["path"].(string)
	if !ok {
		return ToolResponse{}, fmt.Errorf("missing or invalid path argument")
	}

	validPath, err := s.ValidatePath(path)
	if err != nil {
		return ToolResponse{}, err
	}

	info, err := os.Stat(validPath)
	if err != nil {
		return ToolResponse{}, fmt.Errorf("error getting file info: %w", err)
	}

	fileInfo := FileInfo{
		Size:        info.Size(),
		Created:     info.ModTime().Format(time.RFC3339),
		Modified:    info.ModTime().Format(time.RFC3339),
		Accessed:    info.ModTime().Format(time.RFC3339), // Go doesn't provide access time
		IsDirectory: info.IsDir(),
		IsFile:      !info.IsDir(),
		Permissions: strconv.FormatUint(uint64(info.Mode().Perm()), 8),
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("size: %d", fileInfo.Size))
	lines = append(lines, fmt.Sprintf("created: %s", fileInfo.Created))
	lines = append(lines, fmt.Sprintf("modified: %s", fileInfo.Modified))
	lines = append(lines, fmt.Sprintf("accessed: %s", fileInfo.Accessed))
	lines = append(lines, fmt.Sprintf("isDirectory: %t", fileInfo.IsDirectory))
	lines = append(lines, fmt.Sprintf("isFile: %t", fileInfo.IsFile))
	lines = append(lines, fmt.Sprintf("permissions: %s", fileInfo.Permissions))

	return ToolResponse{
		Content: []ContentItem{{Type: "text", Text: strings.Join(lines, "\n")}},
	}, nil
}

// handleListAllowedDirectories handles the list_allowed_directories tool
func (s *Server) handleListAllowedDirectories(args map[string]interface{}) (ToolResponse, error) {
	return ToolResponse{
		Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Allowed directories:\n%s", strings.Join(s.allowedDirectories, "\n"))}},
	}, nil
}

// handleEditFile handles the edit_file tool
func (s *Server) handleEditFile(args map[string]interface{}) (ToolResponse, error) {
	path, ok := args["path"].(string)
	if !ok {
		return ToolResponse{}, fmt.Errorf("missing or invalid path argument")
	}

	editsInterface, ok := args["edits"].([]interface{})
	if !ok {
		return ToolResponse{}, fmt.Errorf("missing or invalid edits argument")
	}

	dryRun := false
	if dryRunVal, ok := args["dryRun"].(bool); ok {
		dryRun = dryRunVal
	}

	validPath, err := s.ValidatePath(path)
	if err != nil {
		return ToolResponse{}, err
	}

	// Read file content
	content, err := os.ReadFile(validPath)
	if err != nil {
		return ToolResponse{}, fmt.Errorf("error reading file: %w", err)
	}

	// Parse edits
	edits := make([]EditOperation, 0, len(editsInterface))
	for i, editInterface := range editsInterface {
		editMap, ok := editInterface.(map[string]interface{})
		if !ok {
			return ToolResponse{}, fmt.Errorf("invalid edit at index %d", i)
		}

		oldText, ok := editMap["oldText"].(string)
		if !ok {
			return ToolResponse{}, fmt.Errorf("missing or invalid oldText in edit at index %d", i)
		}

		newText, ok := editMap["newText"].(string)
		if !ok {
			return ToolResponse{}, fmt.Errorf("missing or invalid newText in edit at index %d", i)
		}

		edits = append(edits, EditOperation{
			OldText: oldText,
			NewText: newText,
		})
	}

	// Apply edits
	modifiedContent := string(content)
	for _, edit := range edits {
		if !strings.Contains(modifiedContent, edit.OldText) {
			// Try line-by-line matching with flexibility for whitespace
			oldLines := strings.Split(edit.OldText, "\n")
			contentLines := strings.Split(modifiedContent, "\n")
			matchFound := false

			for i := 0; i <= len(contentLines)-len(oldLines); i++ {
				potentialMatch := contentLines[i : i+len(oldLines)]
				isMatch := true

				for j, oldLine := range oldLines {
					contentLine := potentialMatch[j]
					if strings.TrimSpace(oldLine) != strings.TrimSpace(contentLine) {
						isMatch = false
						break
					}
				}

				if isMatch {
					// Preserve original indentation of first line
					originalIndent := ""
					if match := regexp.MustCompile(`^\s*`).FindString(contentLines[i]); match != "" {
						originalIndent = match
					}

					newLines := strings.Split(edit.NewText, "\n")
					for j, line := range newLines {
						if j == 0 {
							newLines[j] = originalIndent + strings.TrimLeft(line, " \t")
						} else {
							// For subsequent lines, try to preserve relative indentation
							oldIndent := ""
							if j < len(oldLines) {
								oldIndent = regexp.MustCompile(`^\s*`).FindString(oldLines[j])
							}
							newIndent := regexp.MustCompile(`^\s*`).FindString(line)

							relativeIndent := ""
							if len(newIndent) > len(oldIndent) {
								relativeIndent = strings.Repeat(" ", len(newIndent)-len(oldIndent))
							}
							newLines[j] = originalIndent + relativeIndent + strings.TrimLeft(line, " \t")
						}
					}

					// Replace the lines
					contentLines = append(
						contentLines[:i],
						append(
							newLines,
							contentLines[i+len(oldLines):]...,
						)...,
					)
					modifiedContent = strings.Join(contentLines, "\n")
					matchFound = true
					break
				}
			}

			if !matchFound {
				return ToolResponse{}, fmt.Errorf("could not find exact match for edit:\n%s", edit.OldText)
			}
		} else {
			modifiedContent = strings.Replace(modifiedContent, edit.OldText, edit.NewText, 1)
		}
	}

	// Create unified diff
	diff := createUnifiedDiff(string(content), modifiedContent, path)

	if !dryRun {
		if err := os.WriteFile(validPath, []byte(modifiedContent), 0644); err != nil {
			return ToolResponse{}, fmt.Errorf("error writing file: %w", err)
		}
	}

	return ToolResponse{
		Content: []ContentItem{{Type: "text", Text: diff}},
	}, nil
}
