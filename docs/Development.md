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
`GOTOOLCHAIN=go1.26.10 make govulncheck` to reproduce its standard-library findings.

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

Follow the [release instructions](releasing.md) for the shared version tag,
preparation steps, preview, and publication commands.

## Contribution checklist

1. Keep the change focused and add tests for behavior changes.
2. Run `gofmt` on Go code and the checks above.
3. Update one canonical documentation page instead of copying guidance.
4. Note user-visible changes in [CHANGELOG.md](CHANGELOG.md).

## Web workspace architecture

The frontend overhaul is organized into three layers:

1. **Workspace and visual foundation:** `workspace.css` composes the navigation,
   independently scrolling library, and preview/editor. Wide screens retain the
   results list while a snippet is open; narrow screens use list-to-detail
   navigation. Density and view preferences are local to the browser.
2. **Retrieval and editing:** `workspace-mixin.js` owns library presentation,
   filter chips, result keyboard navigation, and editor expansion. The existing
   editor and draft mixins own save state and recovery. Save explicitly persists
   to the server; timed draft saves persist only on the current device. Search
   request sequencing prevents stale responses from replacing newer results.
3. **Settings and public reading:** `components/settings.html` is a workspace
   destination with URL-addressable sections (`?settings=backup`), styled by
   `settings.css`. Confirmation dialogs remain in `modals.html`. Public page
   styles live in `public.css`, separate from the template and workspace.

Keep feature behavior in its existing mixin. Do not use getters in spread-in
mixins: JavaScript evaluates them while composing the component, before Alpine
state exists. Use methods for derived mixin state. Keep layout-specific CSS in
`workspace.css` and standalone component rules in their component stylesheet.
User CSS continues to load after application styles.

Run frontend regression tests with Node.js 22.7+:

```bash
npm run test:frontend
```

These cover component construction, search races/errors, URL restoration,
string snippet IDs, draft recovery, save failure/success, and settings navigation.
Run `go test ./...` for backend and embedded-template checks.

For manual verification use a temporary database. Check light and dark themes,
wide and narrow viewports, keyboard-only search → result → edit → save, browser
Back, draft recovery, multi-file copying, settings sections, and public reading.
`Ctrl/Cmd+S` saves an open edit. Arrow keys move through focused library results;
Enter opens the focused result. Escape dismisses the frontmost supported layer
before closing the editor, keeping a local draft when editing.


### Library scale checks

Use `scripts/scale-preview.py` against an **empty temporary localhost instance**;
never point a seed run at your working database. The script refuses non-empty
libraries and non-loopback URLs. Start the server with a separate `SNIPO_DB_PATH`,
then run:

```sh
python3 scripts/scale-preview.py http://127.0.0.1:18081 --snippets 160 --tags 12
python3 scripts/scale-preview.py http://127.0.0.1:18082 --snippets 160 --tags 60
# Repeat without seeding, using the browser's lightweight representation:
python3 scripts/scale-preview.py http://127.0.0.1:18082 --measure-only --summary --output /tmp/scale.json
```

November 2026 local check: each library contained 160 synthetic snippets, including
long (1,200-line) and multi-file examples. Thirty warm serial HTTP requests per
operation gave these results after porting to upstream main `89e694a` (v2.0.2):

| Dataset | First page p95 | Last page p95 | Content search p95 | Tags p95 | First page bytes |
| --- | ---: | ---: | ---: | ---: | ---: |
| 160 snippets / 12 tags | 3.44 ms | 4.41 ms | 9.64 ms | 4.56 ms | 20,358 |
| 160 snippets / 60 tags | 3.18 ms | 4.86 ms | 9.40 ms | 14.85 ms | 20,448 |

The summary representation limits preview text and omits full file contents from
the list payload. It currently still loads full content in the service, so this
is not a database-memory optimization.
Default API list responses and detail responses retain full content. Card copy
fetches detail first. Summary previews can identify matches in secondary files.

Manual desktop checks covered next-page navigation, finding and filtering `zsh`
among 60 tags, and searching `canary-000` in a secondary README. Sidebar tags are
bounded to ten for larger collections (plus the selected tag), searchable, and
expandable. Pagination remains at the top of the scrolling library.
These HTTP numbers exclude browser rendering, WAN latency, concurrent users,
and low-powered self-hosting hardware; they are a reproducible local baseline,
not a production performance guarantee.
