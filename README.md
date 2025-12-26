# Snipo

A lightweight, self-hosted snippet manager designed for single-user deployments.

> **Note:** This project is intentionally scoped for single-user use. Multi-user features are not planned.

[![CI](https://github.com/MohamedElashri/snipo/actions/workflows/ci.yml/badge.svg)](https://github.com/MohamedElashri/snipo/actions/workflows/ci.yml)
[![Release](https://github.com/MohamedElashri/snipo/actions/workflows/release.yml/badge.svg)](https://github.com/MohamedElashri/snipo/actions/workflows/release.yml)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![GitHub release](https://img.shields.io/github/v/release/MohamedElashri/snipo?include_prereleases)](https://github.com/MohamedElashri/snipo/releases)

<p align="center">
  <img src="docs/demo.png" alt="Snipo Demo" width="800">
</p>

## Quick Start

### Docker (Recommended)

```bash
# Create environment file
cat > .env << EOF
SNIPO_MASTER_PASSWORD=your-secure-password
SNIPO_SESSION_SECRET=$(openssl rand -hex 32)
EOF

# Run with Docker Compose
docker compose up -d
```

Or using Docker directly:

```bash
docker run -d \
  -p 8080:8080 \
  -v snipo-data:/app/data \
  -e SNIPO_MASTER_PASSWORD=your-secure-password \
  -e SNIPO_SESSION_SECRET=$(openssl rand -hex 32) \
  --name snipo \
  ghcr.io/mohamedelashri/snipo:latest
```

Access at http://localhost:8080

### Binary

```bash
# Download latest release
curl -LO https://github.com/MohamedElashri/snipo/releases/latest/download/snipo_linux_amd64.tar.gz
tar xzf snipo_linux_amd64.tar.gz

# Configure and run
export SNIPO_MASTER_PASSWORD="your-secure-password"
export SNIPO_SESSION_SECRET=$(openssl rand -hex 32)
./snipo serve
```

## Configuration

### Essential

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `SNIPO_MASTER_PASSWORD` | Yes* | - | Login password (plain text) |
| `SNIPO_MASTER_PASSWORD_HASH` | Yes* | - | Pre-hashed password (Argon2id) - **recommended** |
| `SNIPO_SESSION_SECRET` | Yes | - | Session signing key (32+ chars) |
| `SNIPO_PORT` | No | `8080` | Server port |
| `SNIPO_DB_PATH` | No | `./data/snipo.db` | SQLite database path |

*Either `SNIPO_MASTER_PASSWORD` or `SNIPO_MASTER_PASSWORD_HASH` is required. Using the hash is recommended for security.

### API Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `SNIPO_RATE_LIMIT_READ` | `1000` | API read operations (per hour) |
| `SNIPO_RATE_LIMIT_WRITE` | `500` | API write operations (per hour) |
| `SNIPO_RATE_LIMIT_ADMIN` | `100` | API admin operations (per hour) |
| `SNIPO_ALLOWED_ORIGINS` | - | CORS allowed origins (comma-separated) |
| `SNIPO_ENABLE_PUBLIC_SNIPPETS` | `true` | Enable public snippet sharing |
| `SNIPO_ENABLE_API_TOKENS` | `true` | Enable API token creation |
| `SNIPO_ENABLE_BACKUP_RESTORE` | `true` | Enable backup/restore |

See [`.env.example`](.env.example) for all available options including S3 backup configuration.

### Password Security

For enhanced security, use a pre-hashed password instead of plain text:

```bash
# Generate a password hash
./snipo hash-password your-secure-password

# Or with Docker
docker run --rm ghcr.io/mohamedelashri/snipo:latest hash-password your-secure-password
```

Then use the generated hash:

```bash
# In .env file
SNIPO_MASTER_PASSWORD_HASH=$argon2id$base64salt$base64hash

# Or in docker-compose.yml
environment:
  - SNIPO_MASTER_PASSWORD_HASH=$argon2id$base64salt$base64hash
```

**Benefits of using hashed passwords:**
- Password never appears in plain text in config files
- Safer for version control (if you encrypt/secure the hash)
- Prevents accidental password exposure in logs or process listings
- Backward compatible - plain text passwords still work

See [SECURITY.md](SECURITY.md) for detailed password security practices.

## API

Create API tokens in Settings → API Tokens with granular permissions:
- **read**: View snippets, tags, folders
- **write**: Create, update, delete resources
- **admin**: Full access including settings

Authenticate via:
- `Authorization: Bearer <token>`
- `X-API-Key: <key>`

All responses include metadata (request ID, timestamp, version) and pagination for lists.

API documentation:
- OpenAPI spec: [`docs/openapi.yaml`](docs/openapi.yaml)
- Interactive docs: `http://localhost:8080/api/v1/openapi.json`

## Security

**Container Security:**
- Runs as non-root user (UID 1000)
- Read-only root filesystem
- All Linux capabilities dropped
- No privilege escalation allowed

**Production Recommendations:**
- Use strong passwords (16+ characters)
- Enable HTTPS via reverse proxy (`Nginx`/`Caddy`/`Traefik`)
- Configure CORS restrictively (`SNIPO_ALLOWED_ORIGINS`)
- Use Docker secrets for sensitive values
- Enable S3 backups with encryption
- Keep image updated regularly

See [Development Guide](docs/Development.md#security) for detailed security configuration.

## Customization

Snipo supports extensive visual customization through custom CSS. Users can personalize the interface by:
- Overriding color schemes and CSS variables
- Customizing component styles (sidebar, editor, modals)
- Creating unique themes and visual effects

Access via **Settings → Appearance → Custom CSS**. See the [Customization Guide](docs/customization.md) for detailed documentation, examples, and best practices.

## Development

See the [Development Guide](docs/Development.md) for build instructions, testing, and contribution guidelines.

## License

[GPLv3](LICENSE)
