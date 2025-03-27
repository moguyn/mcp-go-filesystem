FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod vendor
RUN go mod tidy

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

EXPOSE 38085

ENV MCP_SERVER_MODE=sse
ENV MCP_LISTEN_ADDR=0.0.0.0:38085

# Set the entrypoint
ENTRYPOINT ["/app/mcp-server-filesystem", "/projects"]
