# Development Guide

This document covers building, testing, and contributing to Snipo.

## Prerequisites

- **Go 1.23+**
- **Make** (optional, for convenience commands)
- **Docker** (optional, for containerized builds)

## Building

### From Source

```bash
# Build binary
make build
# Output: ./bin/snipo

# Or directly with Go
go build -o bin/snipo ./cmd/server
```

### Docker Image

```bash
# Build local image
make docker

# Or with docker build
docker build -t snipo:local .
```

## Running

### Development Mode

```bash
# Run with hot reload (requires air: go install github.com/air-verse/air@latest)
make dev

# Or run directly
export SNIPO_MASTER_PASSWORD="dev-password"
export SNIPO_SESSION_SECRET="dev-secret-at-least-32-characters-long" ## Generate it with: "openssl rand -hex 32
go run ./cmd/server serve
```

### With Docker Compose

```bash
# Copy example environment
cp .env.example .env
# Edit .env with your settings

docker compose up -d
```

## Testing

```bash
# Run all tests
make test

# Run with coverage
make coverage ## we have poor coverage right now, you are welcome to improve it

# Run specific package tests
go test -v ./internal/api/handlers/...

# Run with race detection
go test -race ./...
```

## Linting

```bash
# Run linter (requires golangci-lint) - all contributions must pass this
make lint

# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

## Configuration Reference

All configuration is via environment variables. See [`.env.example`](../.env.example) for defaults.

### Core Settings

| Variable | Default | Description |
|----------|---------|-------------|
| `SNIPO_HOST` | `0.0.0.0` | Server bind address |
| `SNIPO_PORT` | `8080` | Server port |
| `SNIPO_DB_PATH` | `./data/snipo.db` | SQLite database path |
| `SNIPO_MASTER_PASSWORD` | **required** | Login password |
| `SNIPO_SESSION_SECRET` | **required** | Session signing key (32+ chars) |
| `SNIPO_SESSION_DURATION` | `168h` | Session lifetime |
| `SNIPO_TRUST_PROXY` | `false` | Trust X-Forwarded-For headers |

### Rate Limiting

| Variable | Default | Description |
|----------|---------|-------------|
| `SNIPO_RATE_LIMIT` | `100` | Requests per window |
| `SNIPO_RATE_WINDOW` | `1m` | Rate limit window duration |

### S3 Backup

| Variable | Default | Description |
|----------|---------|-------------|
| `SNIPO_S3_ENABLED` | `false` | Enable S3 backup |
| `SNIPO_S3_ENDPOINT` | `s3.amazonaws.com` | S3 endpoint URL |
| `SNIPO_S3_ACCESS_KEY` | - | Access key ID |
| `SNIPO_S3_SECRET_KEY` | - | Secret access key |
| `SNIPO_S3_BUCKET` | `snipo-backups` | Bucket name |
| `SNIPO_S3_REGION` | `us-east-1` | AWS region |
| `SNIPO_S3_SSL` | `true` | Use HTTPS |

### Logging

| Variable | Default | Description |
|----------|---------|-------------|
| `SNIPO_LOG_LEVEL` | `info` | Log level: debug, info, warn, error |
| `SNIPO_LOG_FORMAT` | `json` | Log format: json, text |

## Database

Snipo uses SQLite with automatic migrations. The database file is created at `SNIPO_DB_PATH` on first run.

### Migrations

Migrations are embedded in the binary and run automatically on startup. Migration files are in `migrations/`.

### Manual Database Access

```bash
sqlite3 ./data/snipo.db
```

## API Development

The API follows `RESTful` conventions. See [`docs/openapi.yaml`](openapi.yaml) for the complete specification.

### Authentication

API requests require one of:
- Session cookie (from web login)
- Bearer token: `Authorization: Bearer <token>`
- API key header: `X-API-Key: <key>`

Create API tokens via Settings → API Tokens in the web UI.

### Example Requests

```bash
# Create snippet
curl -X POST http://localhost:8080/api/v1/snippets \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Example",
    "files": [{"filename": "main.go", "content": "package main", "language": "go"}]
  }'

# Search snippets
curl "http://localhost:8080/api/v1/snippets/search?q=example" \
  -H "Authorization: Bearer TOKEN"

# Export backup
curl -o backup.json "http://localhost:8080/api/v1/backup/export" \
  -H "Authorization: Bearer TOKEN"
```

## Releasing

Releases are automated via GitHub Actions when a version tag is pushed.

### Creating a Release

```bash
git checkout main
git pull origin main
git tag v1.0.0
git push origin v1.0.0
```

### Version Format

Follow [Semantic Versioning](https://semver.org/):
- **Major** (`v2.0.0`): Breaking changes (If needed)
- **Minor** (`v1.1.0`): New features, backward compatible
- **Patch** (`v1.0.1`): Bug fixes

### Release Artifacts

Each release includes:
- `snipo_linux_amd64.tar.gz` - Linux x86_64 binary
- `snipo_linux_arm64.tar.gz` - Linux ARM64 binary
- Docker images: `ghcr.io/mohamedelashri/snipo:v1.0.0`, `:v1.0`, `:v1`, `:latest`

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make changes and add tests
4. Run `make test` and `make lint`
5. Commit with clear messages
6. Open a pull request

### Code Style

- Follow standard Go conventions
- Run `gofmt` before committing
- Keep functions focused and testable
- Add comments for exported functions

## Keyboard Shortcuts (Web UI)

| Shortcut | Action |
|----------|--------|
| `Ctrl+K` / `Cmd+K` | Focus search |
| `Ctrl+N` / `Cmd+N` | New snippet |
| `Escape` | Close editor/modal |
