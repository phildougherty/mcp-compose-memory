# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev ca-certificates

# Set work directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags="-s -w" \
    -o mcp-compose-memory \
    .

# Final stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata postgresql-client

# Set timezone
ENV TZ=America/New_York

# Set work directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/mcp-compose-memory .

# Make binary executable
RUN chmod +x ./mcp-compose-memory

# Create data directory with proper permissions  
RUN mkdir -p /data && chmod 755 /data

# Default environment
ENV DATABASE_URL=postgresql://postgres:password@postgres:5432/memory_graph?sslmode=disable

# Expose the port
EXPOSE 3001

# Add health check
HEALTHCHECK --interval=30s --timeout=5s --retries=3 --start-period=15s \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3001/health || exit 1

# Run the server
CMD ["./mcp-compose-memory", "--host", "0.0.0.0", "--port", "3001"]
