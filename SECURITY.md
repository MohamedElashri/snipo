# Security

This page describes Snipo's security boundaries and the deployment choices that
operators must make. For environment variables and proxy examples, see the
[deployment guide](docs/deployment.md).

## Security model

Snipo is a single-user application. It has one master password, shared settings,
and one database. API token permissions limit client capabilities, but they do
not create isolated users or private workspaces.

Use Snipo on a private network or behind an HTTPS reverse proxy. Do not expose
the container port directly to the public internet.

## Authentication modes

| Mode | Configuration | Anonymous access | Appropriate use |
|---|---|---|---|
| Standard | Default | Public snippets and public metadata only | Normal deployments |
| Login page disabled | Settings → General | Snippet read/write operations | A trusted private network |
| Authentication disabled | `SNIPO_DISABLE_AUTH=true` | All routes, including admin routes | Only behind an access-control proxy or in isolated development |

When the login page is disabled in settings, anonymous requests receive
write-level access. Admin operations still require the master password through
`X-Snipo-Master-Password`, and changing the setting requires the password.

`SNIPO_DISABLE_AUTH=true` is different: it bypasses authentication and password
checks completely. Network isolation must prevent access around the external
authentication proxy.

## Credentials and stored secrets

- A plaintext `SNIPO_MASTER_PASSWORD` is converted to an Argon2id hash at
  startup. Prefer `SNIPO_MASTER_PASSWORD_HASH` so the plaintext password is not
  present in the process environment.
- Session and API tokens are random values. Snipo stores HMAC-SHA-256 digests,
  not the raw tokens, in SQLite.
- API tokens are shown once at creation. Give each client a separate,
  expiring token with the least privilege required.
- GitHub tokens are encrypted in the database with AES-256-GCM using keys
  derived from `SNIPO_ENCRYPTION_SALT` and `SNIPO_SESSION_SECRET`.
- The VS Code client uses the Snipo URL from user settings. Workspace settings
  cannot override the destination used with your saved API token.
- Password-protected exports use AES-256-GCM with a PBKDF2-SHA-256-derived key.

Set stable values for `SNIPO_SESSION_SECRET` and `SNIPO_ENCRYPTION_SALT` before
first use. If the encryption salt is omitted, Snipo generates it and attempts to
store it as `.encryption_salt` beside the database. Back up that file with the
database. Changing either configured secret can affect access to stored
third-party credentials.

Generation and quoting instructions are in the
[deployment guide](docs/deployment.md#authentication) and `.env.example`.

## Sessions and API tokens

Session cookies are `HttpOnly`, `Secure`, and `SameSite=Strict`, with a default
lifetime of seven days. HTTPS is therefore required outside local development.

Logging out clears local recovery drafts and cached settings, and other open
Snipo tabs reload without the ended session. Save your changes before logging
out. Appearance and layout preferences are retained. Private pages and API
responses must not be cached by a reverse proxy.

Permissions are cumulative:

| Permission | Access |
|---|---|
| `read` | Read snippets, tags, folders, history, and Gist status |
| `write` | `read` plus create, update, delete, restore, and sync operations |
| `admin` | `write` plus settings, token management, backup, and Gist configuration |

Creating or deleting an API token also requires the master password while
Snipo authentication is enabled. Authentication syntax and endpoint
requirements are defined in the [OpenAPI specification](docs/openapi.yaml).

## Network controls

- Terminate TLS at a reverse proxy and keep port 8080 private.
- Leave `SNIPO_TRUST_PROXY=false` for direct connections. Enable it only when a
  trusted proxy overwrites `X-Forwarded-For` and clients cannot reach Snipo
  directly. Snipo uses the final forwarded address; earlier entries are not
  trusted. Multiple proxy hops need a trusted client-IP policy at the final proxy.
- Browser writes using a session cookie require the same origin or an exact
  trusted origin from `SNIPO_ALLOWED_ORIGINS`. A wildcard does not grant cookie
  write access. Valid API-token clients authenticate separately.
- Clients must use the final Snipo URL; authenticated API requests do not follow
  redirects, and the shared JavaScript client restricts requests to its configured
  origin.
- CORS is same-origin by default. Set exact origins in
  `SNIPO_ALLOWED_ORIGINS`; use `*` only for controlled development.
- Disable unused public sharing, API token, or backup endpoints with the
  corresponding feature flags.
- Password verification runs one memory-heavy check at a time. Concurrent login
  attempts receive a short retry delay instead of starting additional checks.
- `SNIPO_RATE_LIMIT` controls login requests per client over `SNIPO_RATE_WINDOW`
  (one minute by default).
- Rate limits are in memory and local to one Snipo process. They are not a
  replacement for proxy-level limits or denial-of-service protection.

Snipo sends CSP, clickjacking, MIME-sniffing, referrer, permissions, and HSTS
headers. HSTS only protects connections that already use HTTPS. The web CSP
currently allows inline scripts, inline styles, and `unsafe-eval` for the
frontend/editor stack, so it reduces but does not eliminate XSS risk.

## Browser extension

Code-block save buttons require a real user click. Interactive saves open a
separate extension window (or a tab when popups are unavailable) so websites cannot read or change the dialog,
private folder names, tag suggestions, or the text you type there. Pending save
drafts stay in trusted browser-session storage and are removed when the dialog
closes. They are not synced with your extension settings.

## Data and backups

SQLite files, generated salts, exported backups, and API client configuration
contain sensitive data. Restrict their filesystem permissions and include the
database WAL/SHM files when taking a live filesystem-level backup. Prefer the
built-in export for portable backups. An encrypted export requires both its
password and the `SNIPO_ENCRYPTION_SALT` used when it was created.

Backups include active and archived snippets across all pages; items in Trash
are excluded. Unsupported backup formats and invalid replacement records are
rejected before existing data is cleared.

Backup imports and S3 restores accept at most 50 MiB of file data. Compressed
backups also have a 50 MiB limit on their decoded JSON metadata. Larger backups
must be split before importing.

Public snippets and their files are available without credentials to anyone
who has the URL. Treat a public link as published data, not as an access secret.

GitHub secret Gists are unlisted, and anyone with their URL can read them. Making
an already published Snipo snippet private does not withdraw its GitHub history.
Private edits cannot sync to an existing public Gist: unlink that Gist and create
a new secret one first. Previously published copies remain outside Snipo's
control.

Snippy displays terminal control characters in snippet content and metadata as
visible escape text. Saving and explicitly copying code preserve the original
content.

Custom CSS is stored in the shared database and injected into the web page.
Only the instance owner should be able to change it. The browser performs basic
checks, but this is not a security boundary.

Container users, filesystem controls, and image variants are documented once in
the [Docker Compose deployment section](docs/deployment.md#docker-compose).

## Reporting a vulnerability

Report vulnerabilities privately through
[GitHub Security Advisories](https://github.com/MohamedElashri/snipo/security/advisories/new).
Do not include exploit details or sensitive deployment data in a public issue.
