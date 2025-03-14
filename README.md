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
- Docker (optional, for containerization)

## Building

```bash
# Using Go directly
go build -o mcp-server-filesystem ./cmd/server
```

## Usage

### Direct Execution

```bash
# Allow access to specific directories
./mcp-server-filesystem /path/to/dir1 /path/to/dir2
```

### Docker

```bash
# Mount directories to /projects
docker run -i --rm \
  --mount type=bind,src=/Users/username/Desktop,dst=/projects/Desktop \
  --mount type=bind,src=/path/to/other/allowed/dir,dst=/projects/other/allowed/dir,ro \
  mcp/filesystem-go /projects
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