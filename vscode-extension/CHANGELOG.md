# Changelog

All notable changes to the "snipo" extension will be documented in this file.

## [2.0.1] - Unreleased

- Keep the extension version aligned with the coordinated Snipo 2.0.1 release.
- Update the nock development dependency to 15.0.5. Extension behavior is unchanged.

## [2.0.0] - 2026-10-25

- Join the coordinated Snipo release version and use the shared native-fetch API client.
- Include dependency updates and security hardening since 0.0.3.
- Build the VSIX with locked packaging tools and validate it in release checks.
- Marketplace submission is pending; the VSIX is available in the GitHub release.

## [0.0.3] - 2026-08-24

### Changed

- Updated dependencies.

## [0.0.2] - 2026-07-16

### Added
- **Tags & Collections View**: Added a new tree view in the sidebar for browsing snippets by tags and folders (collections).
- **Offline Support**: Implemented read-only offline caching. You can now view and search your cached snippets even when disconnected from the internet.
- **CodeLens Provider**: Added an editor CodeLens ("Save to Snipo") that automatically appears above functions and classes for supported languages, as well as a global "Save File to Snipo" action at the top of the file.
- **Hover Provider**: Added an editor hover provider. Hovering over a `snipo:[snippet-id]` reference will dynamically display a rich markdown preview of the snippet.

### Fixed
- **Security**: Fixed potential command injection vulnerability in the Hover Provider.
- **Security**: Updated dependencies to resolve High RCE and Moderate DoS vulnerabilities (patched `serialize-javascript`).
- **UI**: Reordered the Activity Bar views so that "Configuration" is consistently at the bottom.
