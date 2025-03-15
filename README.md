# MCP Filesystem Server

[![Go CI](https://github.com/moguyn/mcp-go-filesystem/actions/workflows/ci.yml/badge.svg)](https://github.com/moguyn/mcp-go-filesystem/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/moguyn/mcp-go-filesystem/branch/main/graph/badge.svg)](https://codecov.io/gh/moguyn/mcp-go-filesystem)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A secure filesystem server for the Model Control Protocol (MCP) that provides controlled access to the local filesystem.

## Overview

MCP Filesystem Server enables secure, controlled access to specified directories on the local filesystem. It's designed to be used as part of the Model Control Protocol ecosystem, providing file system operations for AI models and other applications.

## Features

- **Security**: Access limited to explicitly allowed directories
- **Multiple Modes**: Support for both stdio and SSE (Server-Sent Events) modes
- **File Operations**: Read, write, edit, move, and delete files
- **Directory Operations**: Create, list, and navigate directory structures
- **Search Capabilities**: Find files by pattern or content
- **Metadata Access**: Get detailed file and directory information

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/your-org/mcp-server-filesystem.git
cd mcp-server-filesystem

# Build
make build

# Install binary
make install
```

### Using Docker

```bash
# Build Docker image
make docker-build

# Run with Docker
make docker-run DIR=/path/to/dir
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

# Run tests with coverage
make test-coverage

# Run linting
make lint

# Format code
make fmt

# Run all checks
make check
```

## Security

The server implements strict security measures:

- Only allows operations within explicitly specified directories
- Rejects any attempt to access files outside allowed directories
- Validates all paths to prevent directory traversal attacks

See [SECURITY.md](SECURITY.md) for our security policy and vulnerability reporting process.

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on how to contribute to this project.

## License

This project is licensed under the [MIT License](LICENSE). 