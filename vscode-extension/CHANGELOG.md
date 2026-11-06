# Changelog

All notable changes to the "snipo" extension will be documented in this file.

## [Unreleased]

- Keep the server used with your saved API token controlled by user settings; workspace overrides are ignored.
- Stop authenticated requests from following redirects to another destination.

## [2.2.0] - 2026-11-06

- Keep each server’s offline cache separate when switching Snipo instances.

- Keep the status indicator matched to the latest connection settings.

- Refresh sidebar lists after creating, updating, or deleting snippets.

- Handle server URLs consistently when opening the dashboard or connecting to Snipo.

- Open current snippet contents from the sidebar and recover cached results after network failures.

- Browse snippets beyond the first 50 items.
- Save snippets as CUDA, TeX, or BibTeX, with LaTeX files detected as TeX.
- Open snippets in the browser when the configured server URL omits its scheme.

## [2.1.1] - 2026-11-05

- Cancel saving or updating a snippet without making unwanted changes.
- Make VS Code’s “Save File to Snipo” save the entire file, even when text is selected.
- Fix the VS Code settings shortcut so it shows Snipo’s settings.
- Fix updating existing snippets while preserving their metadata and additional files.

## [2.1.0] - 2026-11-05

- Fix opening and saving snippets so their contents are returned correctly.
- Keep search results matched to your latest query, even on slow connections.
- Stop waiting indefinitely when the server does not respond.
- Stop using saved credentials after they are removed from settings.

## [2.0.2] - 2026-10-31

- Align the extension version with the coordinated Snipo 2.0.2 hotfix for web
  theme colors. Extension behavior is unchanged.

## [2.0.1] - 2026-10-31

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
