# syntax=docker/dockerfile:1

# ============================================================================
# Build stage
# ============================================================================
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build arguments for version info
ARG VERSION=dev
ARG COMMIT=unknown

# Build arguments for target platform (automatically set by Docker Buildx)
ARG TARGETOS=linux
ARG TARGETARCH=amd64

# Build the binary with optimizations
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags="-w -s -X main.Version=${VERSION} -X main.Commit=${COMMIT}" \
    -o /snipo \
    ./cmd/server

# ============================================================================
# Final stage - minimal runtime image
# ============================================================================
FROM alpine:3.20

# Install ca-certificates for HTTPS and create non-root user
RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -u 1000 snipo \
    && mkdir -p /data /tmp \
    && chown -R snipo:snipo /data /tmp

# Copy the binary with proper permissions
COPY --from=builder --chown=root:root --chmod=755 /snipo /snipo

# Create data directory structure
WORKDIR /data

# Add security labels
LABEL org.opencontainers.image.source="https://github.com/MohamedElashri/snipo" \
      org.opencontainers.image.description="Self-hosted snippet manager" \
      org.opencontainers.image.licenses="GPL-3.0" \
      org.opencontainers.image.vendor="Mohamed Elashri"

# Expose the default port
EXPOSE 8080

# Health check using the built-in health command
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD ["/snipo", "health"]

# Run as non-root user 
USER snipo

# Default environment variables
ENV SNIPO_HOST=0.0.0.0 \
    SNIPO_PORT=8080 \
    SNIPO_DB_PATH=/data/snipo.db \
    SNIPO_LOG_FORMAT=json

# Run the server
ENTRYPOINT ["/snipo"]
CMD ["serve"]
