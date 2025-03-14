package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// Server represents the MCP filesystem server
type Server struct {
	allowedDirectories []string
	reader             *bufio.Reader
	writer             *bufio.Writer
}

// NewServer creates a new MCP filesystem server
func NewServer(allowedDirectories []string) *Server {
	return &Server{
		allowedDirectories: allowedDirectories,
		reader:             bufio.NewReader(os.Stdin),
		writer:             bufio.NewWriter(os.Stdout),
	}
}

// Run starts the server and processes incoming requests
func (s *Server) Run() error {
	for {
		// Read a line from stdin
		line, err := s.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("error reading from stdin: %w", err)
		}

		// Parse the request
		var request Request
		if err := json.Unmarshal([]byte(line), &request); err != nil {
			if err := s.sendErrorResponse(fmt.Sprintf("error parsing request: %v", err)); err != nil {
				return fmt.Errorf("error sending error response: %w", err)
			}
			continue
		}

		// Process the request
		switch request.Method {
		case "mcp.list_tools":
			if err := s.handleListTools(request.ID); err != nil {
				return fmt.Errorf("error handling list_tools: %w", err)
			}
		case "mcp.call_tool":
			if err := s.handleCallTool(request.ID, request.Params); err != nil {
				return fmt.Errorf("error handling call_tool: %w", err)
			}
		default:
			if err := s.sendErrorResponse(fmt.Sprintf("unknown method: %s", request.Method)); err != nil {
				return fmt.Errorf("error sending error response: %w", err)
			}
		}
	}
}

// ExpandHome expands the tilde in a file path to the user's home directory
func ExpandHome(path string) string {
	if !strings.HasPrefix(path, "~/") && path != "~" {
		return path
	}

	currentUser, err := user.Current()
	if err != nil {
		return path
	}

	if path == "~" {
		return currentUser.HomeDir
	}
	return filepath.Join(currentUser.HomeDir, path[2:])
}

// ValidatePath checks if a path is within the allowed directories
func (s *Server) ValidatePath(requestedPath string) (string, error) {
	expandedPath := ExpandHome(requestedPath)
	var absolute string
	if filepath.IsAbs(expandedPath) {
		absolute = filepath.Clean(expandedPath)
	} else {
		var err error
		absolute, err = filepath.Abs(expandedPath)
		if err != nil {
			return "", fmt.Errorf("error resolving absolute path: %w", err)
		}
	}

	normalizedRequested := filepath.Clean(absolute)

	// Check if path is within allowed directories
	isAllowed := false
	for _, dir := range s.allowedDirectories {
		if strings.HasPrefix(normalizedRequested, dir) {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return "", fmt.Errorf("access denied - path outside allowed directories: %s not in %s", absolute, strings.Join(s.allowedDirectories, ", "))
	}

	// Handle symlinks by checking their real path
	realPath, err := filepath.EvalSymlinks(absolute)
	if err == nil {
		normalizedReal := filepath.Clean(realPath)
		isRealPathAllowed := false
		for _, dir := range s.allowedDirectories {
			if strings.HasPrefix(normalizedReal, dir) {
				isRealPathAllowed = true
				break
			}
		}
		if !isRealPathAllowed {
			return "", fmt.Errorf("access denied - symlink target outside allowed directories")
		}
		return realPath, nil
	}

	// For new files that don't exist yet, verify parent directory
	parentDir := filepath.Dir(absolute)
	realParentPath, err := filepath.EvalSymlinks(parentDir)
	if err != nil {
		return "", fmt.Errorf("parent directory does not exist: %s", parentDir)
	}

	normalizedParent := filepath.Clean(realParentPath)
	isParentAllowed := false
	for _, dir := range s.allowedDirectories {
		if strings.HasPrefix(normalizedParent, dir) {
			isParentAllowed = true
			break
		}
	}
	if !isParentAllowed {
		return "", fmt.Errorf("access denied - parent directory outside allowed directories")
	}

	return absolute, nil
}

// sendResponse sends a JSON response to stdout
func (s *Server) sendResponse(id string, result interface{}) error {
	response := Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	return s.sendJSON(response)
}

// sendErrorResponse sends an error response
func (s *Server) sendErrorResponse(message string) error {
	errorResponse := ErrorResponse{
		JSONRPC: "2.0",
		Error: Error{
			Message: message,
		},
	}
	return s.sendJSON(errorResponse)
}

// sendJSON marshals and sends a JSON object to stdout
func (s *Server) sendJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}

	if _, err := s.writer.Write(data); err != nil {
		return fmt.Errorf("error writing response: %w", err)
	}
	if err := s.writer.WriteByte('\n'); err != nil {
		return fmt.Errorf("error writing newline: %w", err)
	}
	return s.writer.Flush()
}
