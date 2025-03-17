package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/moguyn/mcp-go-filesystem/internal/errors"
	"github.com/moguyn/mcp-go-filesystem/internal/logging"
)

// ServiceProvider provides access to all services
type ServiceProvider struct {
	fileService      FileReader
	fileWriter       FileWriter
	fileManager      FileManager
	directoryService DirectoryManager
	searchService    SearchProvider
	logger           *logging.Logger
	allowedDirs      []string
}

// NewServiceProvider creates a new ServiceProvider
func NewServiceProvider(allowedDirectories []string) *ServiceProvider {
	fileService := NewFileService(allowedDirectories)
	directoryService := NewDirectoryService(allowedDirectories)
	searchService := NewSearchService(allowedDirectories)

	return &ServiceProvider{
		fileService:      fileService,
		fileWriter:       fileService,
		fileManager:      fileService,
		directoryService: directoryService,
		searchService:    searchService,
		logger:           logging.DefaultLogger("service_provider"),
		allowedDirs:      allowedDirectories,
	}
}

// ListAllowedDirectories returns the list of allowed directories
func (p *ServiceProvider) ListAllowedDirectories() []string {
	return p.allowedDirs
}

// RegisterTools registers all filesystem tools with the MCP server
func RegisterTools(s *server.MCPServer, allowedDirectories []string) {
	// Create service provider
	provider := NewServiceProvider(allowedDirectories)

	// Register read_file tool
	readFileTool := mcp.NewTool("read_file",
		mcp.WithDescription(`description: Read the complete contents of a file from the file system. This tool safely reads files only within allowed directories and handles various encodings. Returns the full text content of the specified file.
demo_commands: [{"path": "/path/to/file.txt"}, {"path": "./relative/path/document.md"}]`),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Path to the file to read"),
		),
	)
	s.AddTool(readFileTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return provider.handleReadFile(ctx, request)
	})

	// Register read_multiple_files tool
	readMultipleFilesTool := mcp.NewTool("read_multiple_files",
		mcp.WithDescription(`description: Read multiple files in a single operation. This is more efficient than making separate read requests when analyzing related files. Provide a JSON array of file paths, and receive a JSON object mapping each path to its content.
demo_commands: [{"paths": "[\"./config.json\", \"./settings.yaml\", \"./data/sample.txt\"]"}]`),
		mcp.WithString("paths",
			mcp.Required(),
			mcp.Description("JSON array of paths to the files to read"),
		),
	)
	s.AddTool(readMultipleFilesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return provider.handleReadMultipleFiles(ctx, request)
	})

	// Register write_file tool
	writeFileTool := mcp.NewTool("write_file",
		mcp.WithDescription(`description: Write content to a file, creating it if it doesn't exist or overwriting/appending if it does. Use the append flag to add content to the end of an existing file rather than replacing its contents.
demo_commands: [{"path": "./new_file.txt", "content": "Hello, world!"}, {"path": "./logs.txt", "content": "New log entry", "append": true}]`),
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
		return provider.handleWriteFile(ctx, request)
	})

	// Register edit_file tool
	editFileTool := mcp.NewTool("edit_file",
		mcp.WithDescription(`description: Edit a specific portion of a file by replacing lines between start_line and end_line with new content. This is useful for making precise changes without rewriting the entire file. Line numbers are 1-indexed.
demo_commands: [{"path": "./config.json", "content": "  \"debug\": true,", "start_line": 5, "end_line": 5}, {"path": "./src/main.go", "content": "// TODO: Implement error handling", "start_line": 42, "end_line": 45}]`),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Path to the file to edit"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("New content to replace the specified lines"),
		),
		mcp.WithNumber("start_line",
			mcp.Required(),
			mcp.Description("Line number to start editing from (1-indexed)"),
		),
		mcp.WithNumber("end_line",
			mcp.Required(),
			mcp.Description("Line number to end editing at (1-indexed, inclusive)"),
		),
	)
	s.AddTool(editFileTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return provider.handleEditFile(ctx, request)
	})

	// Register list_directory tool
	listDirectoryTool := mcp.NewTool("list_directory",
		mcp.WithDescription(`description: List all files and subdirectories in a specified directory, including metadata like file size, modification time, and file type. Returns a JSON array of entry objects.
demo_commands: [{"path": "."}, {"path": "./src"}, {"path": "/allowed/directory/path"}]`),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Path to the directory to list"),
		),
	)
	s.AddTool(listDirectoryTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return provider.handleListDirectory(ctx, request)
	})

	// Register create_directory tool
	createDirectoryTool := mcp.NewTool("create_directory",
		mcp.WithDescription(`description: Create a new directory at the specified path. Automatically creates any necessary parent directories that don't exist (similar to mkdir -p). Only works within allowed directories.
demo_commands: [{"path": "./new_directory"}, {"path": "./parent/child/grandchild"}]`),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Path to the directory to create"),
		),
	)
	s.AddTool(createDirectoryTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return provider.handleCreateDirectory(ctx, request)
	})

	// Register delete_directory tool
	deleteDirectoryTool := mcp.NewTool("delete_directory",
		mcp.WithDescription(`description: Delete a directory at the specified path. By default, only empty directories can be deleted. Set recursive to true to delete all contents within the directory as well.
demo_commands: [{"path": "./empty_dir"}, {"path": "./project_backup", "recursive": true}]`),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Path to the directory to delete"),
		),
		mcp.WithBoolean("recursive",
			mcp.Description("Whether to delete non-empty directories recursively"),
		),
	)
	s.AddTool(deleteDirectoryTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return provider.handleDeleteDirectory(ctx, request)
	})

	// Register delete_file tool
	deleteFileTool := mcp.NewTool("delete_file",
		mcp.WithDescription(`description: Delete a file at the specified path. This permanently removes the file from the filesystem. Only works within allowed directories.
demo_commands: [{"path": "./temp.txt"}, {"path": "./logs/old_log.txt"}]`),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Path to the file to delete"),
		),
	)
	s.AddTool(deleteFileTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return provider.handleDeleteFile(ctx, request)
	})

	// Register move_file tool
	moveFileTool := mcp.NewTool("move_file",
		mcp.WithDescription(`description: Move or rename a file from source_path to destination_path. This is equivalent to both moving a file to a different directory and renaming it in the same directory. Both paths must be within allowed directories.
demo_commands: [{"source_path": "./old_name.txt", "destination_path": "./new_name.txt"}, {"source_path": "./file.txt", "destination_path": "./subfolder/file.txt"}]`),
		mcp.WithString("source_path",
			mcp.Required(),
			mcp.Description("Path to the file to move"),
		),
		mcp.WithString("destination_path",
			mcp.Required(),
			mcp.Description("Path to move the file to"),
		),
	)
	s.AddTool(moveFileTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return provider.handleMoveFile(ctx, request)
	})

	// Register copy_file tool
	copyFileTool := mcp.NewTool("copy_file",
		mcp.WithDescription(`description: Copy a file from source_path to destination_path while keeping the original file intact. This creates a duplicate of the file at the new location. Both paths must be within allowed directories.
demo_commands: [{"source_path": "./template.html", "destination_path": "./pages/new_page.html"}, {"source_path": "./config.json", "destination_path": "./config_backup.json"}]`),
		mcp.WithString("source_path",
			mcp.Required(),
			mcp.Description("Path to the file to copy"),
		),
		mcp.WithString("destination_path",
			mcp.Required(),
			mcp.Description("Path to copy the file to"),
		),
	)
	s.AddTool(copyFileTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return provider.handleCopyFile(ctx, request)
	})

	// Register search_files tool
	searchFilesTool := mcp.NewTool("search_files",
		mcp.WithDescription(`description: Search for text content within files in a directory. Returns matching files with line numbers and surrounding context for each match. Set recursive to true to search in all subdirectories recursively.
demo_commands: [{"query": "function main", "path": "./src", "recursive": true}, {"query": "TODO", "path": "./", "recursive": false}]`),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Text to search for"),
		),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Path to the directory to search in"),
		),
		mcp.WithBoolean("recursive",
			mcp.Description("Whether to search recursively in subdirectories"),
		),
	)
	s.AddTool(searchFilesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return provider.handleSearchFiles(ctx, request)
	})

	// Register list_allowed_directories tool
	listAllowedDirectoriesTool := mcp.NewTool("list_allowed_directories",
		mcp.WithDescription(`description: List all directories that are allowed to be accessed by the filesystem tools. This helps you understand which paths you can work with using the other tools. The response is a JSON array of directory paths.
demo_commands: [{"hi": ""}]`),
		mcp.WithString("hi",
			mcp.Description("no effect"),
		),
	)
	s.AddTool(listAllowedDirectoriesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return provider.handleListAllowedDirectories(ctx, request)
	})
}

// Handler methods for ServiceProvider

func (p *ServiceProvider) handleReadFile(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, ok := request.Params.Arguments["path"].(string)
	if !ok {
		return nil, errors.NewFileSystemError("read_file", "", errors.ErrInvalidArgument)
	}

	content, err := p.fileService.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(content), nil
}

func (p *ServiceProvider) handleReadMultipleFiles(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathsJSON, ok := request.Params.Arguments["paths"].(string)
	if !ok {
		return nil, errors.NewFileSystemError("read_multiple_files", "", errors.ErrInvalidArgument)
	}

	// Parse JSON array
	var paths []string
	if err := json.Unmarshal([]byte(pathsJSON), &paths); err != nil {
		return nil, errors.NewFileSystemError("read_multiple_files", "", err)
	}

	results, err := p.fileService.ReadMultipleFiles(paths)
	if err != nil {
		return nil, err
	}

	// Convert results to JSON
	resultJSON, err := json.Marshal(results)
	if err != nil {
		return nil, errors.NewFileSystemError("read_multiple_files", "", err)
	}

	return mcp.NewToolResultText(string(resultJSON)), nil
}

func (p *ServiceProvider) handleWriteFile(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, ok := request.Params.Arguments["path"].(string)
	if !ok {
		return nil, errors.NewFileSystemError("write_file", "", errors.ErrInvalidArgument)
	}

	content, ok := request.Params.Arguments["content"].(string)
	if !ok {
		return nil, errors.NewFileSystemError("write_file", "", errors.ErrInvalidArgument)
	}

	appendFlag := false
	if appendArg, ok := request.Params.Arguments["append"].(bool); ok {
		appendFlag = appendArg
	}

	if err := p.fileWriter.WriteFile(path, content, appendFlag); err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(fmt.Sprintf("File written successfully: %s", path)), nil
}

func (p *ServiceProvider) handleEditFile(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, ok := request.Params.Arguments["path"].(string)
	if !ok {
		return nil, errors.NewFileSystemError("edit_file", "", errors.ErrInvalidArgument)
	}

	content, ok := request.Params.Arguments["content"].(string)
	if !ok {
		return nil, errors.NewFileSystemError("edit_file", "", errors.ErrInvalidArgument)
	}

	startLine, ok := request.Params.Arguments["start_line"].(float64)
	if !ok {
		return nil, errors.NewFileSystemError("edit_file", "", errors.ErrInvalidArgument)
	}

	endLine, ok := request.Params.Arguments["end_line"].(float64)
	if !ok {
		return nil, errors.NewFileSystemError("edit_file", "", errors.ErrInvalidArgument)
	}

	if err := p.fileWriter.EditFile(path, content, int(startLine), int(endLine)); err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(fmt.Sprintf("File edited successfully: %s (lines %d-%d)", path, int(startLine), int(endLine))), nil
}

func (p *ServiceProvider) handleListDirectory(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, ok := request.Params.Arguments["path"].(string)
	if !ok {
		return nil, errors.NewFileSystemError("list_directory", "", errors.ErrInvalidArgument)
	}

	entries, err := p.directoryService.ListDirectory(path)
	if err != nil {
		return nil, err
	}

	// Convert entries to JSON
	entriesJSON, err := json.Marshal(entries)
	if err != nil {
		return nil, errors.NewFileSystemError("list_directory", "", err)
	}

	return mcp.NewToolResultText(string(entriesJSON)), nil
}

func (p *ServiceProvider) handleCreateDirectory(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, ok := request.Params.Arguments["path"].(string)
	if !ok {
		return nil, errors.NewFileSystemError("create_directory", "", errors.ErrInvalidArgument)
	}

	if err := p.directoryService.CreateDirectory(path); err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(fmt.Sprintf("Directory created successfully: %s", path)), nil
}

func (p *ServiceProvider) handleDeleteDirectory(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, ok := request.Params.Arguments["path"].(string)
	if !ok {
		return nil, errors.NewFileSystemError("delete_directory", "", errors.ErrInvalidArgument)
	}

	recursive := false
	if recursiveArg, ok := request.Params.Arguments["recursive"].(bool); ok {
		recursive = recursiveArg
	}

	if err := p.directoryService.DeleteDirectory(path, recursive); err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(fmt.Sprintf("Directory deleted successfully: %s", path)), nil
}

func (p *ServiceProvider) handleDeleteFile(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, ok := request.Params.Arguments["path"].(string)
	if !ok {
		return nil, errors.NewFileSystemError("delete_file", "", errors.ErrInvalidArgument)
	}

	if err := p.fileManager.DeleteFile(path); err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(fmt.Sprintf("File deleted successfully: %s", path)), nil
}

func (p *ServiceProvider) handleMoveFile(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sourcePath, ok := request.Params.Arguments["source_path"].(string)
	if !ok {
		return nil, errors.NewFileSystemError("move_file", "", errors.ErrInvalidArgument)
	}

	destinationPath, ok := request.Params.Arguments["destination_path"].(string)
	if !ok {
		return nil, errors.NewFileSystemError("move_file", "", errors.ErrInvalidArgument)
	}

	if err := p.fileManager.MoveFile(sourcePath, destinationPath); err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(fmt.Sprintf("File moved successfully from %s to %s", sourcePath, destinationPath)), nil
}

func (p *ServiceProvider) handleCopyFile(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sourcePath, ok := request.Params.Arguments["source_path"].(string)
	if !ok {
		return nil, errors.NewFileSystemError("copy_file", "", errors.ErrInvalidArgument)
	}

	destinationPath, ok := request.Params.Arguments["destination_path"].(string)
	if !ok {
		return nil, errors.NewFileSystemError("copy_file", "", errors.ErrInvalidArgument)
	}

	if err := p.fileManager.CopyFile(sourcePath, destinationPath); err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(fmt.Sprintf("File copied successfully from %s to %s", sourcePath, destinationPath)), nil
}

func (p *ServiceProvider) handleSearchFiles(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, ok := request.Params.Arguments["query"].(string)
	if !ok {
		return nil, errors.NewFileSystemError("search_files", "", errors.ErrInvalidArgument)
	}

	path, ok := request.Params.Arguments["path"].(string)
	if !ok {
		return nil, errors.NewFileSystemError("search_files", "", errors.ErrInvalidArgument)
	}

	recursive := false
	if recursiveArg, ok := request.Params.Arguments["recursive"].(bool); ok {
		recursive = recursiveArg
	}

	results, err := p.searchService.SearchFiles(query, path, recursive)
	if err != nil {
		return nil, err
	}

	// Convert results to JSON
	resultsJSON, err := json.Marshal(results)
	if err != nil {
		return nil, errors.NewFileSystemError("search_files", "", err)
	}

	return mcp.NewToolResultText(string(resultsJSON)), nil
}

func (p *ServiceProvider) handleListAllowedDirectories(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	directories := p.ListAllowedDirectories()

	// Convert directories to JSON
	directoriesJSON, err := json.Marshal(directories)
	if err != nil {
		return nil, errors.NewFileSystemError("list_allowed_directories", "", err)
	}

	return mcp.NewToolResultText(string(directoriesJSON)), nil
}
