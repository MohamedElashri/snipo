# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.3.2] - TBD

### Web App

#### Security

##### Hash Algorithm Upgrade
- **API Tokens**: Upgraded from `SHA256` to `HMAC-SHA256` for improved security
- **Sessions**: Upgraded from `SHA256` to `HMAC-SHA256` for improved security  
- **Backup Encryption**: Upgraded from `SHA256` to `PBKDF2` (100,000 iterations) for password-based key derivation

**Migration:** Fully backward compatible. Existing tokens and sessions continue to work and are automatically upgraded on first use. No user action required.

**Configuration:** Added `SNIPO_ENCRYPTION_SALT` environment variable for production backup encryption (auto-generated if not set).

**Deprecation Notice:** Legacy `SHA256` fallback will be removed in v1.5.0 (Q1/2 2026) - Hopefully I will not be a jerk for that release.

#### Added

- Added CHANGELOG.md to track version history and changes
- Added copy page from the overview page with dropdown support for multi-files snippets
- Add option to exclude the first line `shebang` from copy button (in the settings -> General)

#### Fixed

- NA

#### Removed

- NA

### Snippy TUI

#### Added

- Added basic TUI functionality for reading snippets
- Handle different terminal themes based on `ANSI` colors.

#### Fixed

- Fixed terminal color detection for better theme support
- Fixed TUI layout issues on different terminal sizes
- Fixed TUI responsiveness on small terminals

#### Removed

- NA

### Web Extension

#### Added

- Initial web extension implementation for adding snippets from any page

#### Fixed

- NA
  
#### Removed

- NA
---

I started writing CHANGELOG from `v1.3.2`