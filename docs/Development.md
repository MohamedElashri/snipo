# Development Guide

This document covers building, testing, and contributing to Snipo.

## Prerequisites

- **Go 1.24+**
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

# Build with version info
docker build \
  --build-arg VERSION=v1.0.0 \
  --build-arg COMMIT=$(git rev-parse --short HEAD) \
  -t snipo:v1.0.0 .
```

## Security

### Docker Security Features

The Docker deployment implements multiple security layers:

**Container Security:**
- **Non-root user**: Runs as UID 1000 (`snipo` user)
- **Read-only root filesystem**: Prevents tampering with system files
- **Dropped capabilities**: All Linux capabilities removed (`cap_drop: ALL`)
- **No privilege escalation**: `no-new-privileges:true` prevents gaining elevated privileges
- **Minimal base image**: Alpine Linux 3.20 with only essential packages

**Filesystem Security:**
- Binary owned by root with 755 permissions (executable but not writable)
- Data directory (`/data`) owned by snipo user
- Temporary storage via tmpfs (10MB limit, automatically cleared)
- Volume mount for persistent data only

**Network Security:**
- No privileged ports required (uses 8080)
- Container-to-container isolation via Docker networks
- CORS configuration for cross-origin access control

**Resource Limits:**
You can add resource constraints in docker-compose.yml:
```yaml
services:
  snipo:
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M
```

### Production Deployment Checklist

- [ ] Use strong `SNIPO_MASTER_PASSWORD` (16+ characters, mixed case, numbers, symbols)
- [ ] Generate random `SNIPO_SESSION_SECRET` (use `openssl rand -hex 32`)
- [ ] Enable HTTPS (use reverse proxy like Nginx/Caddy/Traefik)
- [ ] Configure `SNIPO_TRUST_PROXY=true` if behind proxy
- [ ] Set restrictive `SNIPO_ALLOWED_ORIGINS` for CORS
- [ ] Use Docker secrets for sensitive environment variables
- [ ] Enable S3 backups with encryption
- [ ] Set up monitoring and health checks
- [ ] Configure log aggregation (`SNIPO_LOG_FORMAT=json`)
- [ ] Keep Docker image updated regularly
- [ ] Review and adjust rate limits based on usage

### Using Docker Secrets (Recommended for Production)

Instead of plain environment variables, use Docker secrets:

```yaml
services:
  snipo:
    secrets:
      - snipo_password
      - snipo_session_secret
    environment:
      - SNIPO_MASTER_PASSWORD_FILE=/run/secrets/snipo_password
      - SNIPO_SESSION_SECRET_FILE=/run/secrets/snipo_session_secret

secrets:
  snipo_password:
    file: ./secrets/password.txt
  snipo_session_secret:
    file: ./secrets/session_secret.txt
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
| `SNIPO_DB_PATH` | `/data/snipo.db` | SQLite database path |
| `SNIPO_MASTER_PASSWORD` | **required** | Login password |
| `SNIPO_SESSION_SECRET` | **required** | Session signing key (32+ chars) |
| `SNIPO_SESSION_DURATION` | `168h` | Session lifetime |
| `SNIPO_TRUST_PROXY` | `false` | Trust X-Forwarded-For headers |

### Rate Limiting

| Variable | Default | Description |
|----------|---------|-------------|
| `SNIPO_RATE_LIMIT` | `100` | Login requests per window |
| `SNIPO_RATE_WINDOW` | `1m` | Rate limit window duration |
| `SNIPO_RATE_LIMIT_READ` | `1000` | API read operations (per hour) |
| `SNIPO_RATE_LIMIT_WRITE` | `500` | API write operations (per hour) |
| `SNIPO_RATE_LIMIT_ADMIN` | `100` | API admin operations (per hour) |

### API Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `SNIPO_ALLOWED_ORIGINS` | - | CORS allowed origins (comma-separated), use `*` for dev |
| `SNIPO_ENABLE_PUBLIC_SNIPPETS` | `true` | Enable public snippet sharing |
| `SNIPO_ENABLE_API_TOKENS` | `true` | Enable API token creation |
| `SNIPO_ENABLE_BACKUP_RESTORE` | `true` | Enable backup/restore features |

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
- Session cookie (from web login) - full admin access
- Bearer token: `Authorization: Bearer <token>`
- API key header: `X-API-Key: <key>`

Create API tokens via Settings → API Tokens in the web UI.

### Token Permissions

API tokens have three permission levels:
- **read**: Can only access GET endpoints (view snippets, tags, folders)
- **write**: Can create, update, and delete snippets, tags, and folders
- **admin**: Full access including token management, settings, and backups

### Rate Limits

API endpoints are rate-limited per token:
- Read operations: 1000 requests/hour (configurable)
- Write operations: 500 requests/hour (configurable)
- Admin operations: 100 requests/hour (configurable)

Rate limit info is included in response headers:
- `X-RateLimit-Limit`: Maximum requests allowed
- `X-RateLimit-Remaining`: Requests remaining
- `X-RateLimit-Reset`: Unix timestamp when limit resets
- `Retry-After`: Seconds to wait (when limit exceeded)

### Response Format

All API responses use standardized envelopes:

**Single resource:**
```json
{
  "data": {...},
  "meta": {
    "request_id": "uuid",
    "timestamp": "2024-12-24T10:30:00Z",
    "version": "1.0"
  }
}
```

**List with pagination:**
```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "total_pages": 8,
    "links": {
      "self": "/api/v1/snippets?page=1",
      "next": "/api/v1/snippets?page=2",
      "prev": null
    }
  },
  "meta": {...}
}
```

### Example Requests

```bash
# Create snippet (returns {data: {...}, meta: {...}})
curl -X POST http://localhost:8080/api/v1/snippets \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Example",
    "files": [{"filename": "main.go", "content": "package main", "language": "go"}]
  }'

# List snippets with pagination
curl "http://localhost:8080/api/v1/snippets?page=1&limit=20" \
  -H "Authorization: Bearer TOKEN"

# Search snippets
curl "http://localhost:8080/api/v1/snippets/search?q=example" \
  -H "Authorization: Bearer TOKEN"

# Export backup
curl -o backup.json "http://localhost:8080/api/v1/backup/export" \
  -H "Authorization: Bearer TOKEN"

# Get API documentation
curl http://localhost:8080/api/v1/openapi.json
```

**Note:** All responses are wrapped in envelopes. Access data via `response.data` instead of directly using the response body.

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

## Vendor Library Management

Snipo serves all frontend JavaScript and CSS libraries locally (no CDN) for privacy. Libraries are managed via npm but served from the `internal/web/static/vendor/` directory.

### Current Libraries

- **htmx** - HTML-over-the-wire interactions
- **Alpine.js** - Reactive UI framework
- **Ace Editor** - Code editor with syntax highlighting
- **Prism.js** - Syntax highlighting for display
- **Marked** - Markdown parser
- **Pico CSS** - Minimal CSS framework

### Setup

```bash
# First time setup
npm install
npm run vendor:sync

# Or use Make
make vendor-install
make vendor-sync
```

### Updating Libraries

```bash
# Check for available updates
make vendor-check

# Update to latest compatible versions (minor/patch)
make vendor-update

# Update to latest including major versions
make vendor-update-major
```

### How It Works

1. Dependencies are declared in `package.json` with semantic versioning
2. `npm install` downloads packages to `node_modules/` (gitignored)
3. `scripts/sync-vendor.js` copies specific files to `internal/web/static/vendor/` (committed)
4. Your app serves files from the vendor directory

### Adding New Libraries

1. Install the package: `npm install <package-name>`
2. Edit `scripts/sync-vendor.js` to add file mappings:
```javascript
const vendorConfig = {
  js: {
    'newlib.min.js': 'package-name/dist/newlib.min.js',
  }
};
```
3. Sync: `npm run vendor:sync`
4. Update HTML templates to include the new library

## Customization

Snipo supports extensive visual customization through custom CSS. See [customization.md](customization.md) for a complete guide on:

- Overriding CSS variables for colors and spacing
- Customizing component styles (sidebar, editor, modals)
- Creating custom themes
- Best practices and examples

Users can add custom CSS through **Settings → Appearance → Custom CSS**.
