# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=${VERSION:-dev} -X main.commit=${COMMIT:-unknown} -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o terminal-velocity \
    cmd/server/main.go

# Build genmap tool
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o genmap \
    cmd/genmap/main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 terminalvelocity && \
    adduser -D -u 1000 -G terminalvelocity terminalvelocity

# Set working directory
WORKDIR /app

# Copy binaries from builder
COPY --from=builder /build/terminal-velocity /app/
COPY --from=builder /build/genmap /app/

# Copy configuration files
COPY configs/config.example.yaml /app/configs/config.yaml

# Create directories
RUN mkdir -p /app/logs /app/data && \
    chown -R terminalvelocity:terminalvelocity /app

# Switch to non-root user
USER terminalvelocity

# Expose SSH port
EXPOSE 2222

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD nc -z localhost 2222 || exit 1

# Run the application
CMD ["/app/terminal-velocity"]
