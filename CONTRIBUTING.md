# Contributing to MCP Filesystem Server

Thank you for your interest in contributing to MCP Filesystem Server! This document provides guidelines and instructions for contributing.

## Code of Conduct

Please be respectful and considerate of others when contributing to this project. We aim to foster an inclusive and welcoming community for all contributors.

## How to Contribute

1. Fork the repository
2. Create a new branch for your feature or bugfix using a descriptive name (e.g., `feature/add-new-operation` or `fix/directory-listing-bug`)
3. Make your changes
4. Run tests and linting
5. Submit a pull request

## Development Setup

1. Clone the repository
2. Install Go (version 1.18 or later)
3. Install dependencies: `make deps`
4. Install development tools: `make install-tools` (installs golangci-lint and other required tools)

## Development Workflow

- Build: `make build`
- Test: `make test`
- Lint: `make lint`
- Format code: `make fmt`
- Run all checks: `make check`
- Run with specific directory: `make run DIR=/path/to/dir`

## Pull Request Guidelines

- Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification for commit messages
  - Use types like `feat:`, `fix:`, `docs:`, `style:`, `refactor:`, `perf:`, `test:`, `chore:`
  - Example: `feat: add support for symbolic links`
- Include tests for new features
- Update documentation as needed
- Ensure all tests pass and linting checks succeed
- Keep pull requests focused on a single topic

## Code Style

This project follows standard Go code style guidelines. Run `make fmt` to format your code before submitting. Additionally:

- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use meaningful variable and function names
- Add comments for exported functions, types, and packages

## License

By contributing to MCP Filesystem Server, you agree that your contributions will be licensed under the project's [MIT License](LICENSE). 