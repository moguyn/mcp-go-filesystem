package server

// Request represents an incoming JSON-RPC request
type Request struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      string                 `json:"id"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
}

// Response represents a successful JSON-RPC response
type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      string      `json:"id"`
	Result  interface{} `json:"result"`
}

// ErrorResponse represents an error JSON-RPC response
type ErrorResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      string `json:"id,omitempty"`
	Error   Error  `json:"error"`
}

// Error represents a JSON-RPC error
type Error struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message"`
}

// Tool represents a tool that can be called by the client
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

// ListToolsResponse represents the response to a list_tools request
type ListToolsResponse struct {
	Tools []Tool `json:"tools"`
}

// ContentItem represents a content item in a tool response
type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ToolResponse represents the response from a tool call
type ToolResponse struct {
	Content []ContentItem `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

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

// EditOperation represents a text replacement operation
type EditOperation struct {
	OldText string `json:"oldText"`
	NewText string `json:"newText"`
}
