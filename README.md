# MCP Filesystem Server

A secure filesystem server for the Model Control Protocol (MCP) that provides controlled access to the local filesystem.

## Features

- Secure access to specified directories only
- Support for both stdio and SSE (Server-Sent Events) modes
- File operations: read, write, edit, move
- Directory operations: create, list, tree view
- Search operations: find files by pattern
- Information operations: get file metadata

## Usage

```
MCP Filesystem Server

Usage: mcp-server-filesystem [options] <allowed-directory> [additional-directories...]

Options:
  --help, -h           Show this help message
  --mode=<mode>        Server mode: 'stdio' (default) or 'sse'
  --listen=<address>   HTTP listen address for SSE mode (default: 127.0.0.1:8080)

The server will only allow operations within the specified directories.

Examples:
  mcp-server-filesystem /path/to/dir1 /path/to/dir2
  mcp-server-filesystem --mode=sse --listen=0.0.0.0:8080 /path/to/dir
```

### Stdio Mode

In stdio mode, the server communicates through standard input and output, making it suitable for integration with other applications that can manage I/O streams.

### SSE Mode

In SSE mode, the server runs as an HTTP server with Server-Sent Events support, allowing real-time communication over HTTP. This is useful for web-based clients or when you need to access the server over a network.

## Security

The server only allows operations within the directories specified on the command line. Any attempt to access files outside these directories will be rejected.

## Building

```bash
go build -o mcp-server-filesystem ./cmd/server
```

## License

[MIT License](LICENSE) 