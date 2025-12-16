# Security Guide for Snipo

This document outlines security considerations and best practices for deploying Snipo.

## Security Model

Snipo is designed as a **local-first, self-hosted** application. The security model assumes:

1. **Single-user deployment** - One master password protects all data
2. **Local network or VPN access** - Not exposed directly to the internet
3. **No CDN dependencies** - All assets served locally to prevent supply chain attacks

## Current Security Features

### Authentication
- **Master password** hashed at startup with Argon2id (OWASP recommended parameters)
- **Progressive login delays** - exponential backoff after failed attempts (1s, 2s, 4s, 8s, 16s, 30s max)
- **Session tokens** hashed with SHA256 before database storage
- **Secure cookies**: `HttpOnly`, `Secure`, `SameSite=Strict`
- **Session expiration** with automatic cleanup
- **API tokens** with SHA256 hashing and optional expiration
- **Rate limiting** on authentication endpoints (configurable)
- **Session secret warning** - logs warning if `SNIPO_SESSION_SECRET` not explicitly set

### HTTP Security Headers
- `Content-Security-Policy` - Restricts resource loading to same-origin
- `X-Content-Type-Options: nosniff` - Prevents MIME sniffing
- `X-Frame-Options: DENY` - Prevents clickjacking
- `X-XSS-Protection: 1; mode=block` - Legacy XSS protection
- `Strict-Transport-Security` - Enforces HTTPS
- `Referrer-Policy: strict-origin-when-cross-origin`
- `Permissions-Policy` - Disables camera, microphone, geolocation

### Input Validation
- JSON request body size limits (2MB max)
- Content size limits (1MB per file)
- Tag name validation (alphanumeric, underscores, hyphens)
- Language allowlist validation

### Database Security
- SQLite with foreign key constraints enabled
- Parameterized queries (SQL injection protection)
- WAL mode for crash recovery

## Configuration Best Practices

### Environment Variables

```bash
# REQUIRED: Strong master password (12+ characters recommended)
SNIPO_MASTER_PASSWORD=your-very-secure-password-here

# REQUIRED: Random session secret (generate with: openssl rand -hex 32)
SNIPO_SESSION_SECRET=$(openssl rand -hex 32)

# Rate limiting (adjust based on expected usage)
SNIPO_RATE_LIMIT=100
SNIPO_RATE_WINDOW=1m

# Only enable if behind a trusted reverse proxy (nginx, traefik, etc.)
SNIPO_TRUST_PROXY=false
```

### Reverse Proxy Configuration

If deploying behind a reverse proxy:

1. Set `SNIPO_TRUST_PROXY=true` to trust `X-Forwarded-For` headers
2. Configure your proxy to set proper headers:

**Nginx example:**
```nginx
location / {
    proxy_pass http://localhost:8080;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

**Traefik example:**
```yaml
http:
  middlewares:
    secure-headers:
      headers:
        forceSTSHeader: true
        stsSeconds: 31536000
```

### Docker Security

The Docker image runs as non-root user (`snipo`, UID 1000):

```yaml
services:
  snipo:
    image: ghcr.io/mohamedelashri/snipo:latest
    security_opt:
      - no-new-privileges:true
    read_only: true
    tmpfs:
      - /tmp
    volumes:
      - snipo_data:/data
```

## Known Limitations

### CSP Relaxations
The Content Security Policy includes:
- `'unsafe-inline'` for styles (required for dynamic styling)
- `'unsafe-eval'` for scripts (required for Alpine.js)

These are necessary for the current frontend stack but reduce XSS protection. The plan is to migrate to a more secure frontend stack in the future.

### Single-User Model
- No role-based access control
- All authenticated users have full access
- Password changes are in-memory only (reset on restart)

## Dependency Management

### Go Dependencies
All dependencies are vendored and version-pinned in `go.mod`:

| Package | Version | Purpose |
|---------|---------|---------|
| `go-chi/chi` | v5.1.0 | HTTP router |
| `golang.org/x/crypto` | v0.28.0 | Argon2id password hashing |
| `modernc.org/sqlite` | v1.33.1 | Pure-Go SQLite driver |
| `aws-sdk-go-v2` | v1.40.1 | S3 backup support |

### Frontend Dependencies (Vendored)
All frontend assets are served locally from `/static/vendor/`:

| Library | Version | File |
|---------|---------|------|
| Alpine.js | 3.x | `alpine.min.js` |
| htmx | 2.x | `htmx.min.js` |
| Ace Editor | 5.x | `ace.js` |
| Prism.js | 1.x | `prism.min.js` |
| Pico CSS | 2.x | `pico.min.css` |
| Fira Code | - | `FiraCode-*.woff2` |

### Updating Dependencies

**Go dependencies:**
```bash
# Check for updates
go list -u -m all

# Update all dependencies
go get -u ./...
go mod tidy

# Update specific package
go get -u golang.org/x/crypto@latest
```

**Frontend dependencies:**
Download new versions and replace files in `internal/web/static/vendor/`:

```bash
# Example: Update Alpine.js
curl -o internal/web/static/vendor/js/alpine.min.js \
  https://cdn.jsdelivr.net/npm/alpinejs@3/dist/cdn.min.js
```

## Security Checklist

- [ ] Set strong master password (12+ characters)
- [ ] Generate random session secret
- [ ] Configure rate limiting appropriately
- [ ] Use HTTPS in production (via reverse proxy)
- [ ] Set `SNIPO_TRUST_PROXY=false` unless behind trusted proxy
- [ ] Restrict network access (firewall/VPN)
- [ ] Regular backups with encryption enabled
- [ ] Keep dependencies updated

## Reporting Security Issues

If you discover a security vulnerability, please report it privately via GitHub Security Advisories or email. Do not create public issues for security vulnerabilities.
