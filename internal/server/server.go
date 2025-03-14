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
	// Print a debug message to stderr
	fmt.Fprintf(os.Stderr, "MCP Filesystem Server starting\n")

	for {
		// Read a line from stdin
		line, err := s.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("error reading from stdin: %w", err)
		}

		// Debug print the received line
		fmt.Fprintf(os.Stderr, "Received: %s", line)

		// First, try to parse as a generic JSON-RPC message to get the ID and method
		var genericRequest struct {
			JSONRPC string          `json:"jsonrpc"`
			ID      interface{}     `json:"id"`
			Method  string          `json:"method"`
			Params  json.RawMessage `json:"params"`
		}

		if err := json.Unmarshal([]byte(line), &genericRequest); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing generic request: %v\n", err)
			// We can't send a proper error response without an ID
			// Send a response with a null ID
			errorResp := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      nil,
				"error": map[string]interface{}{
					"code":    -32700, // Parse error
					"message": fmt.Sprintf("Parse error: %v", err),
				},
			}
			if err := s.sendJSON(errorResp); err != nil {
				return fmt.Errorf("error sending error response: %w", err)
			}
			continue
		}

		// Process based on method
		switch genericRequest.Method {
		case "initialize":
			fmt.Fprintf(os.Stderr, "Handling initialize request\n")

			// Parse the initialize params
			var params InitializeParams
			if err := json.Unmarshal(genericRequest.Params, &params); err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing initialize params: %v\n", err)
				errorResp := map[string]interface{}{
					"jsonrpc": "2.0",
					"id":      genericRequest.ID,
					"error": map[string]interface{}{
						"code":    -32602, // Invalid params
						"message": fmt.Sprintf("Invalid params: %v", err),
					},
				}
				if err := s.sendJSON(errorResp); err != nil {
					return fmt.Errorf("error sending error response: %w", err)
				}
				continue
			}

			// Send initialize response
			result := InitializeResult{
				ServerInfo: Implementation{
					Name:    "mcp-go-filesystem",
					Version: "1.0.0",
				},
				ProtocolVersion: "2024-11-05",
				Capabilities:    map[string]interface{}{},
			}

			response := Response{
				JSONRPC: "2.0",
				ID:      genericRequest.ID,
				Result:  result,
			}

			if err := s.sendJSON(response); err != nil {
				return fmt.Errorf("error sending initialize response: %w", err)
			}

		case "mcp.list_tools":
			if err := s.handleListTools(genericRequest.ID); err != nil {
				return fmt.Errorf("error handling list_tools: %w", err)
			}

		case "mcp.call_tool":
			// Parse the call_tool params
			var params map[string]interface{}
			if err := json.Unmarshal(genericRequest.Params, &params); err != nil {
				errorResp := map[string]interface{}{
					"jsonrpc": "2.0",
					"id":      genericRequest.ID,
					"error": map[string]interface{}{
						"code":    -32602, // Invalid params
						"message": fmt.Sprintf("Invalid params: %v", err),
					},
				}
				if err := s.sendJSON(errorResp); err != nil {
					return fmt.Errorf("error sending error response: %w", err)
				}
				continue
			}

			if err := s.handleCallTool(genericRequest.ID, params); err != nil {
				return fmt.Errorf("error handling call_tool: %w", err)
			}

		default:
			errorResp := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      genericRequest.ID,
				"error": map[string]interface{}{
					"code":    -32601, // Method not found
					"message": fmt.Sprintf("Method not found: %s", genericRequest.Method),
				},
			}
			if err := s.sendJSON(errorResp); err != nil {
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
func (s *Server) sendResponse(id interface{}, result interface{}) error {
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  result,
	}
	return s.sendJSON(response)
}

// sendErrorResponse sends an error response
func (s *Server) sendErrorResponse(message string) error {
	errorResponse := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      nil,
		"error": map[string]interface{}{
			"code":    -32000,
			"message": message,
		},
	}
	return s.sendJSON(errorResponse)
}

// sendErrorResponseWithID sends an error response with the specified request ID
func (s *Server) sendErrorResponseWithID(id interface{}, message string) error {
	errorResponse := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"error": map[string]interface{}{
			"code":    -32000,
			"message": message,
		},
	}
	return s.sendJSON(errorResponse)
}

// sendJSON marshals and sends a JSON object to stdout
func (s *Server) sendJSON(v interface{}) error {
	// Debug print to stderr
	fmt.Fprintf(os.Stderr, "Sending response: %+v\n", v)

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
