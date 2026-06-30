# Development

## Prerequisites

- Go version declared by the root `go.mod`
- Node.js and npm for vendored frontend assets
- Make for the documented commands
- Docker only for image and Compose work

## Build and run

```bash
make build
SNIPO_MASTER_PASSWORD=development-only \
SNIPO_SESSION_SECRET=development-session-secret \
SNIPO_DB_PATH=./snipo.db \
./bin/snipo serve
```

`make dev` runs the server with `go run`; it does not provide hot reload.
`make run-test` starts a local unauthenticated instance and should never be used
on a shared network.

Useful binary commands:

```bash
./bin/snipo version
./bin/snipo migrate
./bin/snipo health
./bin/snipo hash-password
```

For containers:

```bash
make docker
docker build -t snipo:local .
docker compose up -d
```

Runtime configuration belongs in the
[deployment reference](deployment.md), not in this contributor guide.

## Test and static checks

Run the checks relevant to a change:

```bash
make test
make test-coverage
make lint
make govulncheck
make vendor-verify
```

The main test target includes the race detector. Package-level tests are useful
during iteration:

```bash
go test ./internal/api/handlers
go test ./internal/services
```

The terminal client is a separate Go module; its
[README](../tui/README.md#build-and-test) owns its build commands. Extension
packaging is documented in the [extension README](../extension/README.md#build).

## Project layout

| Path | Responsibility |
|---|---|
| `cmd/server` | Server CLI and process lifecycle |
| `internal/api` | Router, middleware, and JSON handlers |
| `internal/auth` | Password, session, and authentication logic |
| `internal/repository` | SQLite data access |
| `internal/services` | Snippets, backups, encryption, and Gist sync |
| `internal/web` | Templates and browser assets |
| `migrations` | Embedded, ordered SQLite migrations |
| `docs/openapi.yaml` | Public API contract |
| `extension` | Chrome and Firefox extension |
| `tui` | Snippy terminal client module |

Migrations run automatically at startup. Add a new numbered migration rather
than editing a migration that may already be installed.

## Frontend dependencies

Frontend libraries are installed with npm, copied into
`internal/web/static/vendor`, and served locally. The copied assets are
committed; `node_modules` is not.

```bash
make vendor          # install, sync, and verify
make vendor-check    # list available updates
make vendor-status   # show installed and expected versions
make vendor-update   # update compatible versions
```

When adding a library:

1. Add it to `package.json`.
2. Add exact file mappings to `scripts/sync-vendor.js`.
3. Run `make vendor`.
4. Update templates and tests.

Do not introduce a runtime CDN dependency.

## API changes

Routes are defined in `internal/api/router.go`. A route change should include:

- handler and permission/rate-limit tests;
- input validation and bounded request sizes;
- consistent response envelopes;
- an update to `docs/openapi.yaml`;
- client updates when the extension or Snippy uses the route.

Authentication and response contracts belong in `docs/openapi.yaml`; token
permissions are enforced by route middleware.

## Security-sensitive changes

Authentication, proxy trust, public routes, custom CSS, backup import, and
credential encryption require explicit negative tests. Do not weaken the
container controls or broaden CORS defaults without documenting the boundary in
[SECURITY.md](../SECURITY.md).

The application does not implement `SNIPO_MASTER_PASSWORD_FILE` or
`SNIPO_SESSION_SECRET_FILE`; adding secret-file examples without implementing
those variables creates an unsafe deployment failure.

## Releases

Releases are built by GitHub Actions from version tags:

```bash
git tag vX.Y.Z
git push origin vX.Y.Z
```

Before tagging, update the version constant, changelog, OpenAPI metadata, and
client versions as applicable. Run the full test, lint, vulnerability, vendor,
extension, and TUI checks.

## Contribution checklist

1. Keep the change focused and add tests for behavior changes.
2. Run `gofmt` on Go code and the checks above.
3. Update one canonical documentation page instead of copying guidance.
4. Note user-visible changes in [CHANGELOG.md](CHANGELOG.md).
