# MCP Filesystem Server

A secure filesystem server for the Model Control Protocol (MCP) that provides controlled access to the local filesystem.

## Features

- Secure access to specified directories only
- Support for both stdio and SSE (Server-Sent Events) modes
- File operations: read, write, edit, move
- Directory operations: create, list, tree view
- Search operations: find files by pattern
- Information operations: get file metadata

## Installation

```bash
# Build from source
make build

# Install binary
make install
```

## Usage

```
MCP Filesystem Server

Usage: mcp-server-filesystem [options] <allowed-directory> [additional-directories...]

Options:
  --help, -h           Show this help message
  --mode=<mode>        Server mode: 'stdio' (default) or 'sse'
  --listen=<address>   HTTP listen address for SSE mode (default: 127.0.0.1:8080)

Examples:
  mcp-server-filesystem /path/to/dir1 /path/to/dir2
  mcp-server-filesystem --mode=sse --listen=0.0.0.0:8080 /path/to/dir
```

### Server Modes

- **Stdio Mode**: Communicates through standard input/output for integration with applications that manage I/O streams.
- **SSE Mode**: Runs as an HTTP server with Server-Sent Events support for real-time communication over HTTP.

## Development

```bash
# Install dependencies
make deps

# Run tests
make test

# Run linting
make lint

# Run all checks
make check
```

## Security

The server only allows operations within the directories specified on the command line. Any attempt to access files outside these directories will be rejected.

See [SECURITY.md](SECURITY.md) for our security policy and reporting vulnerabilities.

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

[MIT License](LICENSE) 