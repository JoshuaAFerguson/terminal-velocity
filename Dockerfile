# Build stage
#
# Multi-arch build: GOARCH is derived from TARGETARCH so `docker buildx`
# and plain `docker build` on arm64 hosts (Orange Pi 5, Raspberry Pi 4+)
# produce the right binary. Falls back to amd64 when TARGETARCH is unset
# (e.g. local builds without buildx).
FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder

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

# Build args. buildx sets TARGETOS/TARGETARCH automatically; plain
# `docker build` leaves them empty, in which case Go's native defaults
# (from `go env GOOS GOARCH`) apply — which is what we want on an
# arm64 host. Don't default them to amd64: that produced an `exec
# format error` on Orange Pi 5 deployments when TARGETARCH wasn't
# populated by buildx.
ARG TARGETOS
ARG TARGETARCH

# Build the application. Bash-style ${VAR:+value} substitution only
# emits the GOOS=/GOARCH= prefix when the arg is non-empty, so Go
# falls back to its native target on a plain `docker build` while
# still respecting buildx's cross-compile args.
RUN CGO_ENABLED=0 \
    ${TARGETOS:+GOOS=${TARGETOS}} \
    ${TARGETARCH:+GOARCH=${TARGETARCH}} \
    go build \
    -ldflags="-w -s -X main.version=${VERSION:-dev} -X main.commit=${COMMIT:-unknown} -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o terminal-velocity \
    cmd/server/main.go

# Build genmap tool
RUN CGO_ENABLED=0 \
    ${TARGETOS:+GOOS=${TARGETOS}} \
    ${TARGETARCH:+GOARCH=${TARGETARCH}} \
    go build \
    -ldflags="-w -s" \
    -o genmap \
    cmd/genmap/main.go

# Final stage
FROM --platform=$TARGETPLATFORM alpine:latest

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
