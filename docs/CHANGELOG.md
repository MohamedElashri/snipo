# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## [1.3.4] - 29/01/2026

### Web App

#### Added

- NA

#### Fixed

- Fix: Allow special characters in tag names (e.g., C#, C++) (#127)
- Fix: Add configurable database memory settings to resolve 'out of memory' errors (#111)

#### Removed

- NA
  
#### Security

- Prevent SQL injection in snippet list query (#128)
- Add URL validation and snippet ID sanitization (#112)
- Security updates and hardening (#108, #114)

### Snippy TUI

#### Added

- NA

#### Fixed

- NA

#### Removed

- NA

### Web Extension

#### Added

- NA

#### Fixed

- NA
  
#### Removed

- NA

---

## [1.3.3] - 07/01/2026

### Web App

#### Added

- Added demo mode to show a demo of the app with some snippets (Can be configured via `SNIPO_DEMO_MODE` environment variable)
- Added GitHub Gists sync (Can be configured settings -> GitHub Gists)
- Add support for deploying snipo to subpaths (Can be configured via `SNIPO_BASE_PATH` environment variable)

#### Fixed

- Fix GitHub sync tokens not persisting across restarts

#### Removed

- NA
  
#### Security

- NA

#### Important Notes

- **GitHub Gist Sync**: To use the GitHub Gist sync feature, you must set the `SNIPO_ENCRYPTION_SALT` environment variable. Without this, GitHub tokens will not persist across application restarts and you'll need to reconnect after each restart. Generate a salt with: `openssl rand -base64 32`. This is not a breaking change for existing users, but is required for the GitHub sync feature to work properly.


### Snippy TUI

#### Added

- NA
- 
#### Fixed

- NA
- 
#### Removed

- NA

### Web Extension

#### Added

- Published Firefox extension (https://addons.mozilla.org/addon/snipo-code-snippet-manager/)

#### Fixed

- NA
  
#### Removed

- NA
---
**

## [1.3.2] - 03/01/2026

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

- Fix pagination not showing up in some cases
- Fix API response formatting for consistent JSON field naming
- Fix sorting by title (A-Z) and title (Z-A) being the same.

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