package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

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
