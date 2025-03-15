# Contributing to MCP Filesystem Server

Thank you for your interest in contributing to MCP Filesystem Server! This document provides guidelines and instructions for contributing.

## Code of Conduct

Please be respectful and considerate of others when contributing to this project.

## How to Contribute

1. Fork the repository
2. Create a new branch for your feature or bugfix
3. Make your changes
4. Run tests and linting
5. Submit a pull request

## Development Setup

1. Clone the repository
2. Install Go (version 1.18 or later)
3. Install dependencies: `make deps`

## Development Workflow

- Build: `make build`
- Test: `make test`
- Lint: `make lint`
- Format code: `make fmt`
- Run all checks: `make check`

## Pull Request Guidelines

- Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification for commit messages
- Include tests for new features
- Update documentation as needed
- Ensure all tests pass and linting checks succeed

## Code Style

This project follows standard Go code style guidelines. Run `make fmt` to format your code before submitting.

## License

By contributing to MCP Filesystem Server, you agree that your contributions will be licensed under the project's [MIT License](LICENSE). 