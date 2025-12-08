# syntax=docker/dockerfile:1

# ============================================================================
# Build stage
# ============================================================================
FROM golang:1.23-alpine AS builder

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

# Build the binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=${VERSION} -X main.Commit=${COMMIT}" \
    -o /snipo \
    ./cmd/server

# ============================================================================
# Final stage - minimal runtime image
# ============================================================================
FROM scratch

# Import CA certificates and timezone data from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /snipo /snipo

# Create data directory structure
WORKDIR /data

# Expose the default port
EXPOSE 8080

# Health check using the built-in health command
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD ["/snipo", "health"]

# Run as non-root user 
USER 1000:1000

# Default environment variables
ENV SNIPO_HOST=0.0.0.0 \
    SNIPO_PORT=8080 \
    SNIPO_DB_PATH=/data/snipo.db \
    SNIPO_LOG_FORMAT=json

# Run the server
ENTRYPOINT ["/snipo"]
CMD ["serve"]
