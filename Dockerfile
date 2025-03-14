FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o mcp-server-filesystem ./cmd/server

# Create a minimal image
FROM alpine:latest

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/mcp-server-filesystem /app/mcp-server-filesystem

# Create a directory for projects
RUN mkdir -p /projects

# Set the entrypoint
ENTRYPOINT ["/app/mcp-server-filesystem"]

# Default command is to serve the /projects directory
CMD ["/projects"] 