# syntax=docker/dockerfile:1

# ============================================================================
# Build stage
# ============================================================================
FROM golang:1.24-alpine AS builder

# Build arguments for target platform (automatically set by Docker Buildx)
# MUST be declared immediately after FROM to get the correct values
ARG TARGETOS
ARG TARGETARCH

# Build arguments for version info
ARG VERSION=dev
ARG COMMIT=unknown

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata file

WORKDIR /src

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Verify platform information (helps with debugging multi-arch builds)
RUN echo "Building for TARGETOS=${TARGETOS} TARGETARCH=${TARGETARCH}"

# Build the binary with optimizations
# Force fresh build by using unique build ID
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags="-w -s -X main.Version=${VERSION} -X main.Commit=${COMMIT}" \
    -o /snipo \
    ./cmd/server

# Verify the built binary architecture
RUN file /snipo && ls -lh /snipo

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
