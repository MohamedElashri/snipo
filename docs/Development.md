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

To build all monorepo components (Server, TUI, Browser Extension, VS Code Extension), run:
```bash
make build-all
```

To start a unified development environment that spins up the server in development mode, and runs watchers for the extensions, run:
```bash
make dev-all
```

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

`make govulncheck` downloads and runs the latest scanner, as CI does; no separate
installation or PATH setup is needed. It requires network access and scans using
the active Go toolchain. CI selects the version in `go.mod`, so use
`GOTOOLCHAIN=go1.26.6 make govulncheck` to reproduce its standard-library findings.

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
| `internal/database` | Connection setup and ordered SQLite migrations |
| `docs/openapi.yaml` | Public API contract |
| `extension` | Chrome and Firefox extension |
| `vscode-extension` | VS Code extension |
| `tui` | Snippy terminal client module |
| `packages/api-client` | Shared TypeScript API client for extensions |

Migrations run automatically at startup. Add a new numbered migration to
`internal/database/migrations.go` rather than editing one that may already be
installed.

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

Releases are built by GitHub Actions via the Unified Release pipeline from version tags. Before tagging, use the bump version script to synchronize versions across all components, and then push the tag:

```bash
./scripts/bump-version.sh vX.Y.Z
git tag vX.Y.Z
git push origin vX.Y.Z
```

The `unified-release.yml` pipeline will automatically build the Go server matrix (Linux/Docker), GoReleaser artifacts (TUI + Homebrew tap), web-ext (Browser), and vsce (VS Code) and aggregate all outputs securely into a single GitHub Release. Run the full test, lint, vulnerability, vendor, extension, and TUI checks before releasing.

## Contribution checklist

1. Keep the change focused and add tests for behavior changes.
2. Run `gofmt` on Go code and the checks above.
3. Update one canonical documentation page instead of copying guidance.
4. Note user-visible changes in [CHANGELOG.md](CHANGELOG.md).
