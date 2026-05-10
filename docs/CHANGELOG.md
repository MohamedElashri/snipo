# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.5.0] - 2026-05-10

### Removed
- Removed legacy SHA256 weak cryptographic hashing algorithm fallback for backup code.

## [1.4.2] - 2026-04-26

### Changed
- Regular update of snipo Go and vendor dependencies.

## [1.4.1] - 2026-04-08

### Fixed
- **Textarea auto-resize**: `autoResizeTextarea` now reads the element's computed `max-height` instead of a hardcoded cap, so the title input (120 px) and description textarea (80 px) each respect their own CSS constraint
- **Cancel button invisible in light theme**: Modal cancel buttons (`modal-footer button:first-child`) now use `--snipo-text-primary` instead of `--pico-color`, which Pico v2 overrides to white on button elements
- **Modal flicker on load/navigation**: Added `x-cloak` to the Token Password and Disable Login Password confirmation overlays, which were briefly visible before Alpine.js initialised
- **Dropdown text clipping**: Removed the `height: 40px` constraint from `.editor-field` inputs and selects; elements now auto-size from padding alone, eliminating the tight content area that clipped text at the bottom border. `::file-selector-button` vertical padding reduced to match
- **Copy button misalignment**: Added `display: inline-flex; align-items: center` to `.dropdown` so the copy dropdown wrapper aligns with sibling `btn-action` buttons in the editor toolbar
- **Gist toggle in preview mode**: Replaced the interactive checkbox/switch with a read-only status indicator matching the Public/Private indicator style; Gist sync can now only be changed from edit mode

### Changed
- New/Edit Folder modal: inputs and buttons use a compact size (`btn-compact`) for a less heavy feel

## [1.4.0] - 2026-04-07

### Added
- **Snippet Expiration & Auto-Archive**: Set optional expiration dates on snippets with automatic archiving. Supports quick presets (7d, 30d, 90d, 6mo, 1y) or custom date picker
- **Copy dropdown**: Adding `copy as rich` option to copy button that is now a dropdown. This copies the snippet with syntax highlighting and formatting preserved, ideal for pasting into rich text editors or mattermost/slack.

### Changed
- **Vendor workflow**: Simplified package.json scripts, added verification and orphan cleanup
- Moved Public and Gist toggles in editor to symmetric compact field-inline layout
- Preview mode public toggle changed from interactive switch to read-only indicator
- Updated golangci-lint to support Go 1.26
- Updated various vendor JS dependencies 
- Updated Go dependencies and sqlite

## [1.3.13] - 2026-10-22

### Added 
- Add RTL and Arabic support with proper mixed-content handling
- Update vendor JS dependencies


### Changed
- Improve snippet history tracking and restoration feature
- Improve the settings Modal CSS to make it more user friendly

## [1.3.12] - 2026-03-09

### fixed
- **TUI** Fixed n keybinding navigation not showing in the TUI interface
- Removed unecessary information from the `/health` endpoint response to avoid potential security implications.

## [1.3.11] - 2026-02-22

### Added

- Added a new public endpoint (`GET /api/v1/metadata/languages`) to the `Snipo` backend, dynamically exposing the canonical list of languages valid for syntax highlighting to any connected client.
- **TUI**: Added editing and writing capabilities to the TUI interface so it is no longer read-only.
- Update vendor JS dependencies including Ace editor components, Alpine, and Marked.

### Fixed

- Fix `SNIPO_DISABLE_AUTH=true` still prompting for password when we try to delete or create a token.
- Fix Gist sync button does not work with new snippets.


## [1.3.10] - 2026-02-19

### Added

- Display application version in the footer

## [1.3.6] - 2026-02-04

### Fixed

- Corrected hardened image release workflow to include conditional debug pulls

## [1.3.5] - 2026-02-02

### Added

- Hardened Docker image build (#132)
- Soft delete for snippets with trash management (#135)

### Fixed

- Ensure API token lists are always arrays (#134)
- Rename modal CSS classes to `snipo-modal-backdrop` to avoid adblocker detection (#136)
- various other fixes and improvements


## [1.3.4] - 2026-01-29

### Fixed

- Allow special characters in tag names (e.g., C#, C++) (#127)
- Add configurable database memory settings to resolve 'out of memory' errors (#111)

### Security

- Prevent SQL injection in snippet list query (#128)
- Add URL validation and snippet ID sanitization (#112)
- Security updates and hardening (#108, #114)

## [1.3.3] - 2026-01-07

### Added

- Demo mode to showcase the app with sample snippets (configure via `SNIPO_DEMO_MODE` environment variable)
- GitHub Gists sync integration (Settings → GitHub Gists)
- Support for deploying snipo to subpaths (configure via `SNIPO_BASE_PATH` environment variable)
- **Extension**: Published Firefox extension ([Install](https://addons.mozilla.org/addon/snipo-code-snippet-manager/))

### Fixed

- GitHub sync tokens not persisting across restarts

> **Note**: To use GitHub Gist sync, set the `SNIPO_ENCRYPTION_SALT` environment variable. Without this, GitHub tokens will not persist across restarts. Generate a salt with: `openssl rand -base64 32`

## [1.3.2] - 2026-01-03

### Added

- CHANGELOG.md to track version history
- Copy button on overview page with dropdown support for multi-file snippets
- Option to exclude first line (shebang) from copy (Settings → General)
- **TUI**: Basic functionality for reading snippets with terminal theme support
- **Extension**: Initial web extension for adding snippets from any page

### Fixed

- Pagination not showing in some cases
- API response formatting for consistent JSON field naming
- Sorting by title (A-Z and Z-A) returning same results
- **TUI**: Terminal color detection for better theme support
- **TUI**: Layout issues on different terminal sizes
- **TUI**: Responsiveness on small terminals

### Security

- **API Tokens**: Upgraded from SHA256 to HMAC-SHA256
- **Sessions**: Upgraded from SHA256 to HMAC-SHA256
- **Backup Encryption**: Upgraded from SHA256 to PBKDF2 (100,000 iterations)

> **Migration**: Fully backward compatible. Existing tokens and sessions continue to work and are automatically upgraded on first use.
>
> **Configuration**: Added `SNIPO_ENCRYPTION_SALT` environment variable for production backup encryption (auto-generated if not set).
>
> **Deprecation**: Legacy SHA256 fallback will be removed in v1.5.0 (Q1/2 2026).

---

*Changelog started at v1.3.2*
