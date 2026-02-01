# Deployment & Configuration Guide

This guide covers everything you need to know about deploying and configuring Snipo, including environment variables, database settings, and advanced deployment scenarios.

## Essential Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `SNIPO_MASTER_PASSWORD` | Yes* | - | Login password (plain text) |
| `SNIPO_MASTER_PASSWORD_HASH` | Yes* | - | Pre-hashed password (Argon2id) - **recommended** |
| `SNIPO_DISABLE_AUTH` | No | `false` | Disable authentication entirely |
| `SNIPO_SESSION_SECRET` | Yes | - | Session signing key (32+ chars) |
| `SNIPO_ENCRYPTION_SALT` | Recommended | Auto-generated | Encryption key for backups & GitHub tokens |
| `SNIPO_PORT` | No | `8080` | Server port |
| `SNIPO_DB_PATH` | No | `/data/snipo.db` | SQLite database path |
| `SNIPO_BASE_PATH` | No | - | Base path for reverse proxy (e.g., `/snipo`) |

*Either `SNIPO_MASTER_PASSWORD` or `SNIPO_MASTER_PASSWORD_HASH` is required (unless `SNIPO_DISABLE_AUTH=true`). Using the hash is recommended for security.

## Hardened Image Variant

For better security, a hardened image variant is available based on [Docker Hardened Images](https://dhi.io). This variant:
- Runs as a non-root user with UID **65532** (nonroot)
- Contains minimal packages (no shell, no package manager)
- Reduces attack surface to absolute minimum
- Tracks CVEs fixes automatically

**Versioning:**
In addition to the `:hardened` rolling tag, versioned tags are available matching the standard release cycles:
- `:vX.Y.Z-hardened` (e.g., `v1.0.0-hardened`)
- `:vX.Y-hardened` (e.g., `v1.0-hardened`)

**Usage:**

```bash
docker run -d \
  -p 8080:8080 \
  -v snipo-data:/data \
  -e SNIPO_MASTER_PASSWORD=your-secure-password \
  -e SNIPO_SESSION_SECRET=$(openssl rand -hex 32) \
  --name snipo \
  ghcr.io/mohamedelashri/snipo:hardened
```

**Important: Permissions**
Since the hardened image runs as `UID 65532`, you must ensure the data volume is writable by this user:

```bash
# Set ownership for the data directory
sudo chown -R 65532:65532 ./snipo-data
```

## Reverse Proxy Configuration

Snipo supports deployment behind a reverse proxy with a custom subpath. This is useful when you want to host Snipo under a specific path like `https://example.com/snipo/`.

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
example.com {
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
      rule: "Host(`example.com`) && PathPrefix(`/snipo`)"
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

## Database Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `SNIPO_DB_PATH` | `/data/snipo.db` | SQLite database path |
| `SNIPO_DB_MAX_CONNS` | `1` | Maximum database connections |
| `SNIPO_DB_BUSY_TIMEOUT` | `5000` | Database busy timeout (ms) |
| `SNIPO_DB_JOURNAL` | `WAL` | Journal mode (WAL/DELETE/MEMORY) |
| `SNIPO_DB_SYNC` | `NORMAL` | Synchronous mode (OFF/NORMAL/FULL) |
| `SNIPO_DB_MMAP_SIZE` | `268435456` | Memory-mapped I/O size (256MB) |
| `SNIPO_DB_CACHE_SIZE` | `-2000` | Cache size in KB (2MB, negative = KB) |

### Database Memory Settings

If you encounter "out of memory" database errors, reduce the memory settings:

```bash
# For systems with limited memory
SNIPO_DB_MMAP_SIZE=67108864    # 64MB instead of 256MB
SNIPO_DB_CACHE_SIZE=-1000      # 1MB instead of 2MB

# For very constrained systems
SNIPO_DB_MMAP_SIZE=33554432    # 32MB
SNIPO_DB_CACHE_SIZE=-500       # 512KB

# Disable memory-mapped I/O if issues persist
SNIPO_DB_MMAP_SIZE=0           # Disable mmap
```

**Note:** The SQLite "out of memory (14)" error is misleading - it's about SQLite's internal memory allocation, not system RAM. Reducing these values typically resolves the issue.

### Database Permission Issues

The "out of memory (14)" error can also be caused by filesystem permission problems:

**Common Causes:**
- Docker volume mount with insufficient permissions
- Read-only filesystem preventing database creation
- User/UID mismatch between host and container

**Solutions:**

**1. Fix Docker Volume Permissions:**
```bash
# Ensure the data directory has proper permissions
sudo chown -R 1000:1000 /path/to/your/data
sudo chmod -R 755 /path/to/your/data

# Or use a bind mount with proper ownership
mkdir -p ./snipo-data
chmod 755 ./snipo-data
```

**2. Docker Compose Configuration:**
```yaml
services:
  snipo:
    volumes:
      - ./snipo-data:/data  # Ensure host directory is writable
    # Remove user mapping if causing permission issues
    # user: "1000:1000"
```

**3. Check Volume Mount:**
```bash
# Verify the container can write to the data directory
docker run --rm -v ./snipo-data:/data alpine touch /data/test
```

**4. Use Named Volumes (Recommended):**
```yaml
services:
  snipo:
    volumes:
      - snipo_data:/data  # Docker manages permissions

volumes:
  snipo_data:
```

## API Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `SNIPO_RATE_LIMIT_READ` | `1000` | API read operations (per hour) |
| `SNIPO_RATE_LIMIT_WRITE` | `500` | API write operations (per hour) |
| `SNIPO_RATE_LIMIT_ADMIN` | `100` | API admin operations (per hour) |
| `SNIPO_ALLOWED_ORIGINS` | - | CORS allowed origins (comma-separated) |
| `SNIPO_ENABLE_PUBLIC_SNIPPETS` | `true` | Enable public snippet sharing |
| `SNIPO_ENABLE_API_TOKENS` | `true` | Enable API token creation |
| `SNIPO_ENABLE_BACKUP_RESTORE` | `true` | Enable backup/restore |

See [`.env.example`](../.env.example) for all available options including S3 backup configuration.

## Password Security

For enhanced security, use a pre-hashed password instead of plain text:

```bash
# Generate a password hash
./snipo hash-password your-secure-password

# Or with Docker
docker run --rm ghcr.io/mohamedelashri/snipo:latest hash-password >
```

Then use the generated hash:

```bash
# In .env file
SNIPO_MASTER_PASSWORD_HASH=$argon2id$base64salt$base64hash

# Or in docker-compose.yml
environment:
  - SNIPO_MASTER_PASSWORD_HASH=$argon2id$base64salt$base64hash
```

> **Docker Compose Warning**: When using `SNIPO_MASTER_PASSWORD_HASH` in docker-compose.yml, the `$` characters in Argon2id hashes will be interpreted as variable substitution. Either:
> - Use double dollar signs: `$$argon2id$$base64salt$$base64hash`
> - Quote the value: `"SNIPO_MASTER_PASSWORD_HASH=$argon2id$base64salt$base64hash"`
> - Use a `.env` file and reference it: `SNIPO_MASTER_PASSWORD_HASH=${SNIPO_MASTER_PASSWORD_HASH}`

See [SECURITY.md](../SECURITY.md) for detailed password security practices.
