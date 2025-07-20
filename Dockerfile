# Multi-stage Docker build for mdriver
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    gcc \
    musl-dev \
    fuse-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-s -w -X main.version=$(git describe --tags --always --dirty) -X main.commit=$(git rev-parse HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o mdriver .

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    fuse \
    && rm -rf /var/cache/apk/*

# Create non-root user
RUN addgroup -g 1001 -S mdriver && \
    adduser -S -D -H -u 1001 -h /tmp -s /sbin/nologin -G mdriver -g mdriver mdriver

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/mdriver /usr/local/bin/mdriver
COPY --from=builder /app/config.yaml /app/config.yaml

# Set proper permissions
RUN chmod +x /usr/local/bin/mdriver && \
    chown mdriver:mdriver /usr/local/bin/mdriver /app/config.yaml

# Create directories for mounting and caching
RUN mkdir -p /mnt/tusd /app/cache /app/sync && \
    chown mdriver:mdriver /mnt/tusd /app/cache /app/sync

USER mdriver

EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD mdriver -version || exit 1

ENTRYPOINT ["/usr/local/bin/mdriver"]
CMD ["-config", "/app/config.yaml"]