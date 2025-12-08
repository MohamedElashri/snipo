# Snipo

A lightweight, self-hosted snippet manager designed for single-user deployments. This is still work in progress.

![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

## Quick Start

### Docker (Recommended)

```bash
# Create environment file
cat > .env << EOF
SNIPO_MASTER_PASSWORD=secure-password
SNIPO_SESSION_SECRET=$(openssl rand -hex 32)
EOF

# Run with Docker Compose
docker compose up -d
```

Access at http://localhost:8080

### Binary

```bash
# Download latest release
curl -LO https://github.com/MohamedElashri/snipo/releases/latest/download/snipo_linux_amd64.tar.gz
tar xzf snipo_linux_amd64.tar.gz

# Configure
export SNIPO_MASTER_PASSWORD="secure-password"
export SNIPO_SESSION_SECRET=$(openssl rand -hex 32)

# Run
./snipo serve
```

### Build from Source

```bash
git clone https://github.com/MohamedElashri/snipo
cd snipo
make build
./bin/snipo serve
```

## Configuration

All configuration is done via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `SNIPO_HOST` | `0.0.0.0` | Server bind address |
| `SNIPO_PORT` | `8080` | Server port |
| `SNIPO_DB_PATH` | `./data/snipo.db` | SQLite database path |
| `SNIPO_MASTER_PASSWORD` | **required** | Login password |
| `SNIPO_SESSION_SECRET` | **required** | Session signing key (32+ chars) |
| `SNIPO_SESSION_DURATION` | `168h` | Session lifetime |
| `SNIPO_LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `SNIPO_LOG_FORMAT` | `json` | Log format (json, text) |

### S3 Backup (Optional)

| Variable | Default | Description |
|----------|---------|-------------|
| `SNIPO_S3_ENABLED` | `false` | Enable S3 backup |
| `SNIPO_S3_ENDPOINT` | `s3.amazonaws.com` | S3 endpoint URL |
| `SNIPO_S3_ACCESS_KEY` | | Access key ID |
| `SNIPO_S3_SECRET_KEY` | | Secret access key |
| `SNIPO_S3_BUCKET` | `snipo-backups` | Bucket name |
| `SNIPO_S3_REGION` | `us-east-1` | AWS region |
| `SNIPO_S3_SSL` | `true` | Use HTTPS |

Works with AWS S3, MinIO, Backblaze B2, DigitalOcean Spaces, etc.

## Usage

### Web Interface

1. Open http://localhost:8080
2. Login with the master password
3. Create snippets using the "+" button
4. Organize with tags and folders
5. Share public snippets via the link icon

### API

All endpoints require authentication via:
- **Session cookie** (web UI)
- **Bearer token**: `Authorization: Bearer <token>`
- **API key header**: `X-API-Key: <key>`

Create API tokens in Settings → API Tokens.

#### Example: Create a snippet

```bash
curl -X POST http://localhost:8080/api/v1/snippets \
  -H "Authorization: Bearer AUTH_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Hello World",
    "files": [
      {"filename": "main.py", "content": "print(\"Hello!\")", "language": "python"}
    ],
    "tags": ["python", "example"]
  }'
```

#### Example: Search snippets

```bash
curl "http://localhost:8080/api/v1/snippets/search?q=hello" \
  -H "Authorization: Bearer AUTH_TOKEN"
```

See [API Documentation](docs/openapi.yaml) for the complete OpenAPI spec.

## Backup & Restore

### Local Export

```bash
# Export as JSON
curl -o backup.json "http://localhost:8080/api/v1/backup/export" \
  -H "Authorization: Bearer AUTH_TOKEN"

# Export encrypted
curl -o backup.enc "http://localhost:8080/api/v1/backup/export?password=secret" \
  -H "Authorization: Bearer AUTH_TOKEN"
```

### S3 Sync

Configure S3 environment variables, then use the Settings → Backup tab or API:

```bash
# Sync to S3
curl -X POST "http://localhost:8080/api/v1/backup/s3/sync" \
  -H "Authorization: Bearer AUTH_TOKEN"

# List S3 backups
curl "http://localhost:8080/api/v1/backup/s3/list" \
  -H "Authorization: Bearer AUTH_TOKEN"
```

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+K` / `Cmd+K` | Focus search |
| `Ctrl+N` / `Cmd+N` | New snippet |
| `Escape` | Close editor/modal |

## Development

```bash
# Run in development mode
make dev

# Run tests
make test

# Build Docker image
make docker

# Lint code
make lint
```

## Releasing (Maintainer Notes)

Releases are automated via GitHub Actions. When a version tag is pushed, the workflow will:

1. Run tests
2. Build binaries for `linux/amd64` and `linux/arm64`
3. Build and push multi-arch Docker images to `ghcr.io`
4. Create a GitHub Release with auto-generated changelog

### Creating a Release

```bash
# Ensure you're on main with latest changes
git checkout main
git pull origin main

# Create and push a version tag (triggers release workflow)
git tag v1.0.0
git push origin v1.0.0
```

### Version Format

Use [Semantic Versioning](https://semver.org/):
- **Major** (`v2.0.0`): Breaking changes
- **Minor** (`v1.1.0`): New features, backward compatible
- **Patch** (`v1.0.1`): Bug fixes

### Release Artifacts

Each release includes:
- `snipo_linux_amd64.tar.gz` - Linux x86_64 binary
- `snipo_linux_arm64.tar.gz` - Linux ARM64 binary
- Docker images tagged as `v1.0.0`, `v1.0`, `v1`, and `latest`

### For Forks

If you fork this repository:

1. Enable GitHub Actions in your fork
2. Ensure `GITHUB_TOKEN` has `contents: write` and `packages: write` permissions
3. Update the Docker image name in `.github/workflows/release.yml` if needed
4. Push a tag to trigger your own release

## License

GPLv3 License - see [LICENSE](LICENSE) for details.
