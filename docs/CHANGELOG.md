# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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