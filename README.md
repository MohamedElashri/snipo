# Snipo

A lightweight, self-hosted snippet manager designed for single-user deployments.

> **Note:** This project is intentionally scoped for single-user use. Multi-user features are not planned.

[![CI](https://github.com/MohamedElashri/snipo/actions/workflows/snipo-ci.yml/badge.svg)](https://github.com/MohamedElashri/snipo/actions/workflows/snipo-ci.yml)
[![CI](https://github.com/MohamedElashri/snipo/actions/workflows/snippy-ci.yml/badge.svg)](https://github.com/MohamedElashri/snipo/actions/workflows/snippy-ci.yml)
[![Release](https://github.com/MohamedElashri/snipo/actions/workflows/release.yml/badge.svg)](https://github.com/MohamedElashri/snipo/actions/workflows/release.yml)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
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
  -v snipo-data:/data \
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
| `SNIPO_DISABLE_AUTH` | No | `false` | Disable authentication entirely |
| `SNIPO_SESSION_SECRET` | Yes | - | Session signing key (32+ chars) |
| `SNIPO_PORT` | No | `8080` | Server port |
| `SNIPO_DB_PATH` | No | `/data/snipo.db` | SQLite database path |
| `SNIPO_BASE_PATH` | No | - | Base path for reverse proxy (e.g., `/snipo`) |

*Either `SNIPO_MASTER_PASSWORD` or `SNIPO_MASTER_PASSWORD_HASH` is required (unless `SNIPO_DISABLE_AUTH=true`). Using the hash is recommended for security.

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

### Disabling Authentication

Snipo offers **three authentication modes** to suit different deployment scenarios:

#### 1. Full Authentication (Default)
Normal mode with login page and password protection:
```bash
SNIPO_MASTER_PASSWORD=your-secure-password
# or
SNIPO_MASTER_PASSWORD_HASH=$argon2id$...
```

#### 2. Login Disabled (Settings Option)
Hide the login page while maintaining security for sensitive operations. Useful for:
- Private networks (Tailscale, WireGuard) where login UI is unnecessary
- Trusted environments where you want easy access but protected token management
- Auth proxy deployments where external authentication handles access control

**Configuration:**
1. Log in to your Snipo instance
2. Go to Settings → General
3. Enable "Disable Login Page"

**How it works:**
- Web UI is accessible without logging in
- Login page redirects to home page
- All API operations work without authentication
- **API token creation and deletion always require password verification** for security

**Security Model:**
- **Read/Write Operations**: No password required
- **Create API Token**: Password required (prompted in UI)
- **Delete API Token**: Password required (prompted in UI)
- **Change Settings**: No password required

**Example - Working with API tokens:**
```bash
# Create token via API (password required in body)
curl -X POST http://localhost:8080/api/v1/tokens \
  -H "Content-Type: application/json" \
  -d '{
    "name":"My Token",
    "permissions":"admin",
    "password":"your-secure-password"
  }'

# Delete token via API (password required in body)
curl -X DELETE http://localhost:8080/api/v1/tokens/1 \
  -H "Content-Type: application/json" \
  -d '{"password":"your-secure-password"}'

# Use token for operations (no password needed)
curl http://localhost:8080/api/v1/snippets \
  -H "Authorization: Bearer <api-token>"
```

**Why this matters:**
- Even if someone gains access to your session or network, they cannot create or delete API tokens without the master password
- Provides an additional security layer against session hijacking or XSS attacks
- Protects against unauthorized API token management

#### 3. Authentication Completely Disabled (Environment Variable)
⚠️ **DANGER: Use with extreme caution!**

Disable **all authentication and password requirements** when deploying behind an external authentication layer:

```bash
SNIPO_DISABLE_AUTH=true
```

**Security Impact:**
- **No login required** - Direct access to web UI
- **No password verification** - API token operations don't require passwords
- **No session authentication** - All API endpoints are open
- **Complete trust** in external authentication layer

**Only use this when:**
- Behind a trusted authentication proxy (Authelia, Authentik, OAuth2 Proxy, Cloudflare Access)
- In a completely isolated local environment with no network access
- For development/testing purposes

**Never use this when:**
- Directly exposed to the internet
- In untrusted networks  
- Without understanding the complete security implications
- Unless you have a properly configured authentication proxy

See [SECURITY.md](SECURITY.md#authentication-modes) for detailed guidance and best practices.

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

## Search

Snipo features powerful fuzzy search that searches across:
- Snippet titles, descriptions, and content
- Multi-file snippet contents
- File names

### Basic Search
Type keywords in the search bar. Multiple words are matched using AND logic:
```
python docker
```
Finds snippets containing both "python" and "docker" anywhere in the metadata or content.

### Filters

**By Tags:**
```
?tag_id=1              # Single tag
?tag_ids=1,2,3         # Multiple tags
```

**By Folders:**
```
?folder_id=1           # Single folder
?folder_ids=1,2,3      # Multiple folders
```

**By Language:**
```
?language=javascript
```

**By Status:**
```
?favorite=true         # Favorites only
?is_archived=true      # Archived snippets
```

### Combining Filters
Mix search with filters for precise results:
```
?q=api&tag_id=1&language=python
```
Searches for "api" in Python snippets with tag 1.

### Sorting
```
?sort=title&order=asc  # A-Z by title
?sort=updated_at       # Recently updated (default)
?sort=created_at       # Recently created
```

**In-app help:** Click the `?` icon next to the search bar for interactive documentation.

## Public Snippets

Share code snippets publicly with granular file-level access.

### Making Snippets Public

**From Editor:**
1. Open or create a snippet
2. Toggle the "Public" switch in the editor header
3. Share the generated URL: `https://localhost:8080/s/{snippet-id}`

**From Preview Page:**
- Click the globe icon in the header to toggle public/private status
- Requires authentication to change visibility

### Accessing Public Snippets

**Web Interface:**
- Multi-file snippets display with file tabs
- Switch between files using the tab interface
- Download individual files with the download button
- Copy file URLs for direct access

**Direct File Access (wget/curl):**

For single-file snippets:
```bash
# Download the file
curl -O https://localhost:8080/api/v1/snippets/public/{snippet-id}/files/{filename}

# Or with wget
wget https://localhost:8080/api/v1/snippets/public/{snippet-id}/files/{filename}
```

For multi-file snippets:
```bash
# Download specific file
curl -O https://localhost:8080/api/v1/snippets/public/abc123/files/config.yaml

# Download all files using the file URLs from the web interface
curl -O https://localhost:8080/api/v1/snippets/public/abc123/files/main.go
curl -O https://localhost:8080/api/v1/snippets/public/abc123/files/README.md
```

**URL Format:**
- Snippet preview: `/s/{snippet-id}`
- Individual file (raw): `/api/v1/snippets/public/{snippet-id}/files/{filename}`

**Permissions:**
- Public snippets are accessible without authentication
- View count is tracked automatically
- Files are returned as plain text with proper Content-Disposition headers

## GitHub Gist Sync

Snipo supports two-way synchronization with GitHub Gists, allowing you to backup your snippets to GitHub and keep them in sync across platforms.

### Setup

1. **Generate GitHub Personal Access Token:**
   - Go to [GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)](https://github.com/settings/tokens/new?scopes=gist&description=Snipo)
   - Create a token with `gist` scope
   - Copy the generated token

2. **Configure in Snipo:**
   - Go to Settings → GitHub Gist tab
   - Paste your GitHub token
   - Click "Save Configuration" (sync is automatically enabled)

### Features

**Sync Options:**
- **Enable Sync for All**: Creates GitHub gists for all your snippets at once
- **Sync Now**: Syncs changes for already-enabled snippets
- **Auto-Sync**: Background sync at configurable intervals (5/15/30/60 minutes)

**Conflict Resolution:**
- **Manual**: Review and resolve conflicts manually
- **Snipo Wins**: Always keep Snipo version
- **Gist Wins**: Always keep GitHub version
- **Newest Wins**: Keep the most recently modified version

**Metadata Preservation:**
- Snippet titles, descriptions, and content are synced
- Snipo-specific metadata (favorites, folders, tags) embedded in gist description
- Multi-file snippets fully supported

### Usage

**Enable sync for all snippets:**
```
Settings → GitHub Gist → Enable Sync for All
```

**View synced snippets:**
- See list of synced snippets with status badges (✓ synced, ⟳ pending, ⚠ conflict, ✗ error)
- Click gist links to view on GitHub
- Remove mappings to stop syncing specific snippets

**Manage conflicts:**
- Conflicts appear when both Snipo and GitHub versions are modified
- Choose "Keep Snipo" or "Keep Gist" to resolve
- Or set automatic conflict resolution strategy in settings

### Limitations

- Requires GitHub Personal Access Token (no OAuth)
- Sync is per-snippet, not automatic for new snippets
- GitHub API rate limit: 5000 requests/hour

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

## Reverse Proxy Configuration

Snipo supports deployment behind a reverse proxy with a custom subpath. This is useful when you want to host Snipo under a specific path like `https://yourdomain.com/snipo/`.

### Configuration

Set the `SNIPO_BASE_PATH` environment variable to your desired subpath:

```bash
SNIPO_BASE_PATH=/snipo
```

**Important notes:**
- The path should start with `/` but not end with `/`
- Examples: `/snipo`, `/apps/snippets`, `/code`
- Leave empty (default) for root path deployment

### Nginx Example

```nginx
location /snipo/ {
    proxy_pass http://localhost:8080/;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

Then configure Snipo:
```bash
SNIPO_BASE_PATH=/snipo
SNIPO_TRUST_PROXY=true
```

### Caddy Example

```caddy
yourdomain.com {
    handle /snipo/* {
        reverse_proxy localhost:8080
    }
}
```

Configuration:
```bash
SNIPO_BASE_PATH=/snipo
```

### Traefik Example

```yaml
http:
  routers:
    snipo:
      rule: "Host(`yourdomain.com`) && PathPrefix(`/snipo`)"
      service: snipo
      middlewares:
        - snipo-stripprefix
  
  middlewares:
    snipo-stripprefix:
      stripPrefix:
        prefixes:
          - "/snipo"
  
  services:
    snipo:
      loadBalancer:
        servers:
          - url: "http://localhost:8080"
```

Configuration:
```bash
SNIPO_BASE_PATH=/snipo
```

### Docker Compose with Reverse Proxy

```yaml
services:
  snipo:
    image: ghcr.io/mohamedelashri/snipo:latest
    environment:
      - SNIPO_MASTER_PASSWORD=your-secure-password
      - SNIPO_SESSION_SECRET=${SESSION_SECRET}
      - SNIPO_BASE_PATH=/snipo
      - SNIPO_TRUST_PROXY=true
    volumes:
      - snipo-data:/data
    networks:
      - proxy-network

volumes:
  snipo-data:

networks:
  proxy-network:
    external: true
```

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
