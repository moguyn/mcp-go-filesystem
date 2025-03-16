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
		return provider.handleReadFile(ctx, request)
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
		return provider.handleReadMultipleFiles(ctx, request)
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
		return provider.handleWriteFile(ctx, request)
	})

	// Register edit_file tool
	editFileTool := mcp.NewTool("edit_file",
		mcp.WithDescription("Edit a specific portion of a file by replacing lines between "+
			"start_line and end_line with the provided content. This is useful "+
			"for making targeted changes to files without rewriting the entire "+
			"file. Only works within allowed directories."),
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
		mcp.WithDescription("List the contents of a directory, showing files and subdirectories "+
			"with their metadata. This is useful for exploring the file system "+
			"and understanding the structure of directories. Only works within "+
			"allowed directories."),
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
		mcp.WithDescription("Create a new directory at the specified path. Creates parent "+
			"directories as needed if they don't exist. Only works within "+
			"allowed directories."),
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
		mcp.WithDescription("Delete a directory at the specified path. Can optionally delete "+
			"non-empty directories recursively. Only works within allowed directories."),
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
		mcp.WithDescription("Delete a file at the specified path. Only works within allowed directories."),
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
		mcp.WithDescription("Move a file from one location to another. This can also be used "+
			"to rename files. Only works within allowed directories."),
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
		mcp.WithDescription("Copy a file from one location to another. Only works within allowed directories."),
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
		mcp.WithDescription("Search for files containing the specified query string. "+
			"Returns matching files with line numbers and context. "+
			"Only works within allowed directories."),
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
		mcp.WithDescription("List all directories that are allowed to be accessed by the filesystem tools. "+
			"This provides transparency about which directories can be manipulated using the filesystem tools."),
	)
	s.AddTool(listAllowedDirectoriesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return provider.handleListAllowedDirectories(ctx, request)
	})
}

// Handler methods for ServiceProvider

func (p *ServiceProvider) handleReadFile(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
