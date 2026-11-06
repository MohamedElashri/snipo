# VS Code extension changelog

Changes below apply to Snipo's VS Code extension.

## [2.2.1] - 2026-11-06

- **VS Code extension:** Keep the server used with your saved API token controlled by user settings; workspace overrides are ignored.
- **VS Code extension:** Stop authenticated requests from following redirects to another destination.

## [2.2.0] - 2026-11-06

- **VS Code extension:** Keep each server’s offline cache separate when switching Snipo instances.
- **VS Code extension:** Keep the status indicator matched to the latest connection settings.
- **VS Code extension:** Refresh sidebar lists after creating, updating, or deleting snippets.
- **VS Code extension:** Handle server URLs consistently when opening the dashboard or connecting to Snipo.
- **VS Code extension:** Open current snippet contents from the sidebar and recover cached results after network failures.
- **VS Code extension:** Browse snippets beyond the first 50 items.
- **VS Code extension:** Save snippets as CUDA, TeX, or BibTeX, with LaTeX files detected as TeX.
- **VS Code extension:** Open snippets in the browser when the configured server URL omits its scheme.

## [2.1.1] - 2026-11-05

- **VS Code extension:** Cancel saving or updating a snippet without making unwanted changes.
- **VS Code extension:** Make “Save File to Snipo” save the entire file, even when text is selected.
- **VS Code extension:** Fix the settings shortcut so it shows Snipo’s settings.
- **VS Code extension:** Fix updating existing snippets while preserving their metadata and additional files.

## [2.1.0] - 2026-11-05

- **VS Code extension:** Fix opening and saving snippets so their contents are returned correctly.
- **VS Code extension:** Keep search results matched to your latest query, even on slow connections.
- **VS Code extension:** Stop waiting indefinitely when the server does not respond.
- **VS Code extension:** Stop using saved credentials after they are removed from settings.

## [2.0.2] - 2026-10-31

- **VS Code extension:** No user-visible changes in this release.

## [2.0.1] - 2026-10-31

- **VS Code extension:** No user-visible changes in this release.

## [2.0.0] - 2026-10-25

- **VS Code extension:** Join the coordinated releases for the server and clients.
- **VS Code extension:** Use shared HTTP connection handling when communicating with Snipo.
- **VS Code extension:** Activate when Snipo commands or views are used instead of loading in every workspace.

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
