# Go Filesystem MCP Server

Go implementation of the Model Context Protocol (MCP) server for filesystem operations.

## Features

- Read/write files
- Create/list/delete directories
- Move files/directories
- Search files
- Get file metadata

**Note**: The server will only allow operations within directories specified via command-line arguments.

## Development

### Prerequisites

- Go 1.24 or later
- Make (optional, for using the Makefile)
- Docker (optional, for containerization)

### Makefile Commands

The project includes a Makefile with common commands:

```bash
# Display all available commands
make help

# Build the application
make build

# Build for multiple platforms (linux, darwin, windows)
make build-all

# Run tests
make test

# Run tests with coverage report
make test-coverage

# Format code
make fmt

# Run code vetting
make vet

# Run linting (requires golangci-lint)
make lint

# Run all code quality checks
make check

# Clean build artifacts
make clean

# Build Docker image
make docker-build

# Run in Docker container
make docker-run DIR=/path/to/dir

# Install dependencies
make deps

# Update dependencies
make update-deps

# Generate documentation
make docs

# Install the binary
make install
```

## Building

```bash
# Using Go directly
go build -o mcp-server-filesystem ./cmd/server

# Using Make
make build

# Build Docker image
make docker-build
```

## Usage

### Direct Execution

```bash
# Allow access to specific directories
./mcp-server-filesystem /path/to/dir1 /path/to/dir2

# Using Make
make run DIR=/path/to/dir
```

### Docker

```bash
# Mount directories to /projects
docker run -i --rm \
  --mount type=bind,src=/Users/username/Desktop,dst=/projects/Desktop \
  --mount type=bind,src=/path/to/other/allowed/dir,dst=/projects/other/allowed/dir,ro \
  mcp/filesystem-go /projects

# Using Make
make docker-run DIR=/path/to/dir
```

## Integration with Claude Desktop

Add this to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "--mount", "type=bind,src=/Users/username/Desktop,dst=/projects/Desktop",
        "--mount", "type=bind,src=/path/to/other/allowed/dir,dst=/projects/other/allowed/dir,ro",
        "mcp/filesystem-go",
        "/projects"
      ]
    }
  }
}
```

## Continuous Integration

This project uses GitHub Actions for continuous integration. The CI pipeline:

1. Builds the project on multiple Go versions
2. Runs all tests with race detection
3. Performs linting with golangci-lint
4. Builds binaries for multiple platforms (Linux, macOS, Windows)

## API

The server implements the Model Context Protocol (MCP) with the following tools:

- **read_file**: Read contents of a file
- **read_multiple_files**: Read multiple files simultaneously
- **write_file**: Create or overwrite a file
- **edit_file**: Make selective edits to a file
- **create_directory**: Create a directory
- **list_directory**: List directory contents
- **directory_tree**: Get recursive directory structure
- **move_file**: Move or rename files/directories
- **search_files**: Search for files matching a pattern
- **get_file_info**: Get file metadata
- **list_allowed_directories**: List allowed directories

## License

MIT License 