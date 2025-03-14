.PHONY: build clean test lint fmt vet run docker-build docker-run help

# Binary name
BINARY_NAME=mcp-server-filesystem
# Docker image name
DOCKER_IMAGE=mcp/filesystem-go
# Main package path
MAIN_PACKAGE=./cmd/server
# Build directory
BUILD_DIR=./build
# Version from git tag or commit hash
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
# Build date
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildDate=$(BUILD_DATE)"

# Default target
.DEFAULT_GOAL := help

# Help target
help: ## Display this help
	@echo "MCP Filesystem Server - Makefile commands:"
	@echo
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Build the application
build: ## Build the application
	go build $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PACKAGE)

# Build for multiple platforms
build-all: ## Build for multiple platforms (linux, darwin, windows)
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)

# Clean build artifacts
clean: ## Remove build artifacts
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)

# Run the application
run: build ## Run the application (specify DIR=/path/to/dir to set allowed directory)
	./$(BINARY_NAME) $(DIR)

# Run tests
test: ## Run tests
	go test -v ./...

# Run tests with coverage
test-coverage: ## Run tests with coverage
	mkdir -p $(BUILD_DIR)
	go test -coverprofile=$(BUILD_DIR)/coverage.out ./...
	go tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html
	@echo "Coverage report generated at $(BUILD_DIR)/coverage.html"

# Format code
fmt: ## Format code
	go fmt ./...

# Vet code
vet: ## Vet code
	go vet ./...

# Lint code
lint: ## Lint code (requires golangci-lint)
	@which golangci-lint > /dev/null || (echo "golangci-lint is required. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run ./...

# Check code (format, vet, lint)
check: fmt vet lint ## Run all code quality checks

# Install dependencies
deps: ## Install dependencies
	go mod download
	go mod tidy

# Build Docker image
docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE) .

# Run Docker container
docker-run: ## Run Docker container (specify DIR=/path/to/dir to mount directory)
	docker run -i --rm \
		--mount type=bind,src=$(DIR),dst=/projects/dir \
		$(DOCKER_IMAGE) /projects

# Generate documentation
docs: ## Generate documentation
	@mkdir -p $(BUILD_DIR)/docs
	@echo "Generating documentation..."
	@go doc -all ./... > $(BUILD_DIR)/docs/api.txt
	@echo "Documentation generated at $(BUILD_DIR)/docs/api.txt"

# Install the binary
install: build ## Install the binary
	go install $(LDFLAGS) $(MAIN_PACKAGE)

# Update dependencies
update-deps: ## Update dependencies
	go get -u ./...
	go mod tidy 