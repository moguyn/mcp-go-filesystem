package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

// FileReader defines operations for reading files
type FileReader interface {
	ReadFile(path string) (string, error)
	ReadMultipleFiles(paths []string) ([]FileContent, error)
}

// FileWriter defines operations for writing files
type FileWriter interface {
	WriteFile(path, content string, append bool) error
	EditFile(path, content string, startLine, endLine int) error
}

// DirectoryManager defines operations for directory management
type DirectoryManager interface {
	CreateDirectory(path string) error
	ListDirectory(path string) ([]FileInfo, error)
	DeleteDirectory(path string, recursive bool) error
}

// FileManager defines operations for file management
type FileManager interface {
	DeleteFile(path string) error
	MoveFile(sourcePath, destinationPath string) error
	CopyFile(sourcePath, destinationPath string) error
}

// SearchProvider defines operations for searching files
type SearchProvider interface {
	SearchFiles(query string, path string, recursive bool) ([]SearchResult, error)
}

// PathValidator defines operations for validating paths
type PathValidator interface {
	ValidatePath(requestedPath string) (string, error)
}

// AllowedDirectoriesProvider defines operations for listing allowed directories
type AllowedDirectoriesProvider interface {
	ListAllowedDirectories() []string
}

// FileContent represents the content of a file with its path
type FileContent struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Error   string `json:"error,omitempty"`
}

// FileInfo represents information about a file or directory
type FileInfo struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	IsDir     bool   `json:"is_dir"`
	ModTime   string `json:"mod_time"`
	Extension string `json:"extension,omitempty"`
}

// SearchResult represents a search result
type SearchResult struct {
	Path    string `json:"path"`
	Line    int    `json:"line,omitempty"`
	Content string `json:"content,omitempty"`
}

// ToolHandler defines the function signature for handling tool requests
type ToolHandler func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

// TreeEntry represents an entry in a directory tree
type TreeEntry struct {
	Name     string      `json:"name"`
	Type     string      `json:"type"` // "file" or "directory"
	Children []TreeEntry `json:"children,omitempty"`
}
