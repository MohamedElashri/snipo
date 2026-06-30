# Deployment and configuration

This is the canonical reference for running Snipo. Security implications are
covered in [SECURITY.md](../SECURITY.md).

## Docker Compose

```bash
cp .env.example .env
chmod 600 .env
```

Replace the password and secret placeholders, then start Snipo:

```bash
docker compose up -d
docker compose ps
```

The standard image runs as UID 1000 and stores persistent state in `/data`.
The supplied Compose file uses a named volume, a read-only root filesystem,
dropped capabilities, and a temporary `/tmp`.

To use the hardened image, replace the image tag with `:hardened`. It runs as
UID 65532. Bind-mounted data directories must be writable by the corresponding
UID:

```bash
sudo chown -R 65532:65532 ./snipo-data
```

Versioned hardened tags use `vX.Y.Z-hardened`, `vX.Y-hardened`, and
`vX-hardened`.

## Binary

Release archives are available from the
[GitHub releases page](https://github.com/MohamedElashri/snipo/releases).

```bash
export SNIPO_MASTER_PASSWORD_HASH='$argon2id$...'
export SNIPO_SESSION_SECRET='...'
export SNIPO_ENCRYPTION_SALT='...'
export SNIPO_DB_PATH='./data/snipo.db'
./snipo serve
```

Run `./snipo hash-password` to generate the password hash. In shell exports,
single-quote the hash so `$` characters remain literal. Snipo applies embedded
database migrations when it starts.

## Reverse proxy and HTTPS

Keep Snipo's port private and terminate TLS at the proxy. A root-path Nginx
configuration is:

```nginx
location / {
    proxy_pass http://127.0.0.1:8080;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

Set `SNIPO_TRUST_PROXY=true` only if clients cannot bypass this proxy and the
proxy replaces untrusted forwarding headers.

For a subpath such as `https://example.com/snipo`, set:

```bash
SNIPO_BASE_PATH=/snipo
```

The proxy must strip the prefix before forwarding:

```bash
location /snipo/ {
    proxy_pass http://127.0.0.1:8080/;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

Do not add a trailing slash to `SNIPO_BASE_PATH`.

## Authentication

Exactly one password source is required unless authentication is completely
disabled:

| Variable | Default | Purpose |
|---|---:|---|
| `SNIPO_MASTER_PASSWORD_HASH` | — | Argon2id master-password hash; preferred |
| `SNIPO_MASTER_PASSWORD` | — | Plaintext master password |
| `SNIPO_DISABLE_AUTH` | `false` | Bypass all Snipo authentication |
| `SNIPO_SESSION_SECRET` | generated | Stable secret used by encrypted application data |
| `SNIPO_SESSION_DURATION` | `168h` | Session lifetime |
| `SNIPO_ENCRYPTION_SALT` | generated/persisted | Salt used for backups and stored GitHub credentials |
| `SNIPO_RATE_LIMIT` | `100` | Login requests per minute |

If both password variables are set, the hash takes precedence. Generated
session secrets do not survive restarts; explicitly configure one in any
persistent deployment. An omitted encryption salt is written to
`.encryption_salt` beside the database when possible.

See [Authentication modes](../SECURITY.md#authentication-modes) before enabling
`SNIPO_DISABLE_AUTH` or the Settings → General “Disable login” option.

## Server and database

| Variable | Default | Purpose |
|---|---:|---|
| `SNIPO_HOST` | `0.0.0.0` | Bind address |
| `SNIPO_PORT` | `8080` | HTTP port |
| `SNIPO_READ_TIMEOUT` | `30s` | HTTP read timeout |
| `SNIPO_WRITE_TIMEOUT` | `30s` | HTTP write timeout |
| `SNIPO_TRUST_PROXY` | `false` | Trust client-IP forwarding headers |
| `SNIPO_BASE_PATH` | empty | External URL prefix |
| `SNIPO_MAX_FILES_PER_SNIPPET` | `10` | Maximum files in one snippet |
| `SNIPO_DB_PATH` | `/data/snipo.db` | SQLite file |
| `SNIPO_DB_MAX_CONNS` | `1` | Maximum open SQLite connections |
| `SNIPO_DB_BUSY_TIMEOUT` | `5000` | Busy timeout in milliseconds |
| `SNIPO_DB_JOURNAL` | `WAL` | SQLite journal mode |
| `SNIPO_DB_SYNC` | `NORMAL` | SQLite synchronous mode |
| `SNIPO_DB_MMAP_SIZE` | `268435456` | Memory-mapped I/O bytes |
| `SNIPO_DB_CACHE_SIZE` | `-2000` | SQLite cache pages; negative values are KiB |

For memory-constrained systems, reduce or disable memory mapping:

```bash
SNIPO_DB_MMAP_SIZE=67108864
SNIPO_DB_CACHE_SIZE=-1000
```

## API and feature flags

| Variable | Default | Purpose |
|---|---:|---|
| `SNIPO_ALLOWED_ORIGINS` | empty | Comma-separated CORS origins; empty means same-origin |
| `SNIPO_RATE_LIMIT_READ` | `1000` | Read requests per token/IP per hour |
| `SNIPO_RATE_LIMIT_WRITE` | `500` | Write requests per token/IP per hour |
| `SNIPO_RATE_LIMIT_ADMIN` | `100` | Admin requests per token/IP per hour |
| `SNIPO_ENABLE_PUBLIC_SNIPPETS` | `true` | Public snippet pages and file endpoints |
| `SNIPO_ENABLE_API_TOKENS` | `true` | API token management endpoints |
| `SNIPO_ENABLE_BACKUP_RESTORE` | `true` | Backup and S3 endpoints |

## S3 backups

| Variable | Default | Purpose |
|---|---:|---|
| `SNIPO_S3_ENABLED` | `false` | Enable S3-compatible backup storage |
| `SNIPO_S3_ENDPOINT` | empty | Service endpoint |
| `SNIPO_S3_ACCESS_KEY` | empty | Access-key ID |
| `SNIPO_S3_SECRET_KEY` | empty | Secret access key |
| `SNIPO_S3_BUCKET` | empty | Bucket name |
| `SNIPO_S3_REGION` | `us-east-1` | Region |
| `SNIPO_S3_SSL` | `true` | Use TLS for the S3 connection |

Use a dedicated bucket credential with only the permissions Snipo needs.

## Logging and health

| Variable | Default | Purpose |
|---|---:|---|
| `SNIPO_LOG_LEVEL` | `info` | `debug`, `info`, `warn`, or `error` |
| `SNIPO_LOG_FORMAT` | `json` | `json` or `text` |

`GET /ping` checks the process; `GET /health` also checks the database. The
binary provides `snipo health` for container health checks.

## Upgrade and backup

Before upgrading:

1. Export a backup from Settings and verify that it can be read.
2. Back up the database and `.encryption_salt` while Snipo is stopped, or use a
   SQLite-aware snapshot that includes WAL state.
3. Pull a pinned release tag, recreate the container, and check `/health`.
4. Review the [changelog](CHANGELOG.md) for migration or configuration notes.
