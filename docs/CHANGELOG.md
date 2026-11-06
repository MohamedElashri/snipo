# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.2.1] - 2026-11-06

### Security

- **Web app:** Protect Markdown previews and keep sign-in redirects on your Snipo instance.
- **Server:** Strengthen protection against unwanted browser requests and repeated password attempts.
- **Server:** Stop serving shared snippets and their files after they are moved to Trash.
- **Web app:** Clear local drafts and cached settings on logout, and lock other open Snipo tabs. Save unfinished edits before logging out.
- **Server:** Keep private Gist edits from being sent to an existing public Gist.
- **Server:** Limit backup imports and S3 restores to 50 MiB, including decoded ZIP metadata.
- **Browser extension, VS Code extension, and Snippy:** Keep authenticated requests on the configured Snipo server without following redirects.
- **Browser extension:** Keep save dialogs and private folder and tag suggestions separate from visited websites.
- **VS Code extension:** Prevent workspaces from changing the server used with your saved API token.
- **Snippy:** Display terminal control characters safely while preserving the original code when saving or copying.

### Fixed

- **Server:** Include all active and archived snippets in backups, including collections larger than 100 snippets.
- **Server:** Reject invalid replacement backups before clearing existing data.
- **Server:** Prevent folder moves and edits from creating circular folder hierarchies.
- **Web app:** Keep sign-in usable after an incorrect password and show a retry message if logout fails.
- **Server:** Allow instance owners to authorize administrative requests with their master password when the login page is disabled.
- **Server:** Apply the configured login rate-limit window and use trusted proxy addresses consistently.
- **Browser extension:** Remove pending drafts when save windows close, including windows closed while still opening.

### Changed

- **VS Code extension:** Configure the server URL in user settings; workspace overrides are ignored.
- **Browser extension:** Open save dialogs in a separate extension window, with a tab fallback when popup windows are unavailable.

## [2.2.0] - 2026-11-06

### Added

- **Server, Browser extension, and VS Code extension:** Save CUDA, TeX, and BibTeX snippets.
- **Browser extension:** Dismiss the save dialog with Escape and return focus to its save button.
- **Web app:** Choose Ace, CodeMirror 6, or Monaco in Editor settings, with a remembered browser preference and safe switching between editors.
- **Web app:** Use Monaco’s language assistance for JavaScript, TypeScript, JSON, HTML, and CSS without relying on external services.
- **Web app:** See which editor is active, including when Snipo falls back to Ace.

### Fixed

- **VS Code extension:** Keep offline snippet, tag, and folder caches separate for each Snipo server.
- **Browser extension:** Ignore surrounding spaces in pasted connection settings.
- **Snippy:** Clear old errors after a successful retry.
- **VS Code extension:** Keep the status indicator matched to the latest connection settings.
- **Snippy:** Return to the same snippet and file after closing Help.
- **Browser extension:** Save snippets from pages whose titles exceed 200 characters.
- **VS Code extension:** Refresh sidebar lists after creating, updating, or deleting snippets.
- **Snippy:** Keep programming-language suggestions out of the API-key field.
- **Browser extension:** Show nested folders in the save dialog.
- **VS Code extension:** Open the dashboard consistently for server URLs without a scheme or with surrounding spaces.
- **Snippy:** Preserve tag capitalization when creating or editing snippets.
- **Browser extension:** Keep the context-menu save dialog working on pages where Snipo is already loaded.
- **VS Code extension:** Open the latest snippet contents from the sidebar and use cached results when native network requests fail.
- **Snippy:** Allow new lines while editing snippets longer than 99 lines.
- **Browser extension:** Preserve code indentation and trailing newlines when saving code blocks.
- **Browser extension:** Prevent duplicate save dialogs and recover from stalled connection checks.
- **VS Code extension:** Show snippets beyond the first 50 and open configured server URLs without a scheme.
- **Snippy:** Copy and edit the selected snippet file, and allow question marks while filling out forms.
- **Web app:** Prevent a late Monaco completion response from causing an error after switching files or editors.
- **Server and web app:** Highlight TeX and BibTeX in all three editors and preserve CUDA, TeX, and BibTeX when saving.
- **Web app:** Recover from stalled editor downloads instead of leaving the editor loading indefinitely.
- **Web app:** Keep undo from restoring a different file’s code after switching snippet files.
- **Web app:** Restore unsaved code when refreshing an editing page.
- **Web app:** Keep each tab’s draft separate so saving in another tab does not erase its recovery copy.
- **Web app:** Make Ace’s manual completion shortcut available alongside its snippet and live-completion support.
- **Web app:** Keep CodeMirror’s find and replace controls compact on smaller screens.
- **Web app:** Keep the snippet title and code usable when browser zoom or side panels narrow the editor.

## [2.1.1] - 2026-11-05

### Fixed

- **Server and web app:** Prevent outdated cached styles and scripts from breaking the web interface after an update.
- **Server and web app:** Fix styles, fonts, and editor assets when Snipo is hosted under a URL subpath.
- **Web app:** Use Cards as the default view while remembering your last Cards, List, or Compact choice.
- **Web app:** Hide tag labels in Cards view to keep dates and card controls aligned.
- **Web app:** Make text sizes more consistent and keep the editor’s description field compact and aligned.
- **VS Code extension:** Cancel saving or updating a snippet without making unwanted changes.
- **Browser extension:** Keep descriptions blank when you clear them.
- **Snippy:** Preserve code indentation, trailing newlines, and existing public visibility when saving.
- **VS Code extension:** Make “Save File to Snipo” save the entire file, even when text is selected.
- **VS Code extension:** Fix the settings shortcut so it shows Snipo’s settings.
- **VS Code extension:** Fix updating existing snippets.
- **VS Code extension and Snippy:** Preserve folders, expiration, archive status, and additional files when editing snippets.

## [2.1.0] - 2026-11-05

### Added

- **Web app:** Choose Cards, List, or Compact views to browse your snippets.
- **Web app:** Search tags, keep your filters while opening snippets, and hide the snippet list
  when you want more room for code.
- **Web app:** Enjoy a more compact snippet toolbar and improved mobile layouts.
- **Web app:** Keep snippets open after saving and recover unfinished edits from local drafts.
- **Web app:** Find settings on a dedicated page with direct links to each section.
- **Server and web app:** Browse large snippets with lighter previews while opening and copying the full code.

### Fixed

- **Web app:** Fix outdated search results, briefly visible password prompts on refresh, and
  display issues on login and shared snippet pages.
- **Web app:** Fix shared links and themes when Snipo is hosted under a URL subpath.
- **Browser extension:** Prevent repeated clicks from saving the same snippet while a save is in progress.
- **Browser extension:** Make the save button available again after a failed save.
- **Browser extension:** Stop waiting indefinitely when the server does not respond.
- **VS Code extension:** Fix opening and saving snippets so their contents are returned correctly.
- **VS Code extension:** Keep search results matched to your latest query.
- **VS Code extension:** Stop waiting indefinitely for an unresponsive server and stop using credentials
  after they are removed from settings.
- **Snippy:** Fix navigation beyond the first page of snippets.
- **Snippy:** Apply all selected tags and folders when filtering.
- **Snippy:** Accept server URLs with trailing slashes.

## [2.0.2] - 2026-10-31

### Fixed

- **Web app:** Restore the light theme's text, card surfaces, and borders after the v2.0.1
  CSS migration.
- **Web app:** Keep theme colors inherited inside controls, links, and headings, and prevent
  native focus/hover styles from overriding component colors in either theme.
- **Web app:** Keep public snippet controls readable on hover in the light theme.

## [2.0.1] - 2026-10-31

### Changed

- **Web app:** Replace Pico CSS with local styles while preserving the existing layout,
  native controls, typography, and light/dark themes.
- **Web app:** Update the customization guide and Settings help with current variables,
  working selectors, theme behavior, RTL examples, and recovery instructions.
- **Web app:** Use Snipo CSS variables in place of `--pico-*` variables for custom styles;
  see the [customization guide](customization.md#migrating-custom-css-from-pico).

## [2.0.0] - 2026-10-25

### Changed

- **All components:** Release the server, web app, browser extension, VS Code extension,
  and Snippy together under one version.
- **Browser extension and VS Code extension:** Use shared HTTP connection handling
  when communicating with your Snipo server.
- **VS Code extension:** Activate when Snipo commands or views are used instead of
  loading in every workspace.
- **Snippy:** Use native system clipboard tools to copy code.
- **All components:** Update dependencies, including the web app's bundled libraries.

### Security

- **Browser extension (Firefox):** Limit page access to the active tab instead of
  requesting access to all websites.

### Fixed

- **Browser extension (Firefox):** Include the files needed to build the supplied
  source package.
- **Browser extension (Firefox):** Correct compatibility information for Firefox for Android.
- **Snippy:** Provide Linux, macOS, and Windows downloads, Linux deb/rpm packages,
  and Homebrew installation.
- **Server:** Include the API reference in server archives and Docker images.

## [1.7.7] - 2026-09-26
- **Go/JS:** Update the Go and NPM dependencies

## [1.7.6] - 2026-09-06

### Changed
- **Upgrade Go and JS depedencies:**


## [1.7.5] - 2026-08-24

### Added
- **Gist Deletion Choice:** Added the ability to permanently delete the corresponding GitHub Gist when permanently deleting a mapped snippet.
- **Unlink with Deletion:** When unlinking a Gist from the Gist Sync settings, users are now presented with a custom modal that provides the option to also delete the remote Gist.
- **Dedicated Sync Delete API:** Introduced `DELETE /api/v1/gist/sync/snippet/{id}` to cleanly handle the atomic deletion of both remote Gists and local mappings.

### Changed
- **Soft-Delete (Trash) Lifecycle:** Moving a synced snippet to the Trash no longer severs its GitHub Gist sync mapping. The sync worker simply pauses syncing for trashed snippets, allowing syncing to seamlessly resume upon restoration without duplicating Gists.
- **Dynamic Delete Modal:** The delete confirmation modal now dynamically changes its title, text, and button (e.g., "Move to Trash" vs. "Permanently Delete") based on whether the action is a soft or hard delete.
- Update dependencies

## [1.7.4] - 2026-07-18

### Fixed
- Prevent aggressive caching of static JS and CSS files by setting Cache-Control headers - - Support `auto` theme selection and fix a UI bug that wiped checkboxes and prevented saving settings

### Changed

- Restrict API key creation/deletion and Gist sync in demo mode
- Debounce settings updates with a 500ms timeout to reduce API request frequency and prevent race conditions

## [1.7.3] - 2026-07-14

### Fixed
- Fix bug in token generation process because of the previous change to HMAC keys

## [1.7.2] - 2026-07-13

### Added
- Footer version label now links to the CHANGELOG section on click (#298)

### Security
- Sanitize FTS5 query input to prevent SQL injection attacks (#302)
- Redact sensitive query parameters from logs and reject passwords in URLs (#304)
- Clear user-controlled data from log entries (#311)
- Remove `X-API-Key` from session token extraction (#305)
- Sanitize `Host` header in pagination links to prevent injection (#307)
- Sanitize user input in log entries to prevent log injection (#300)
- Enforce `Content-Type: application/json` in `DecodeJSON` (#301)

### Fixed
- Derive HMAC keys from session secret instead of hardcoding a static value (#308)
- Move S3 delete key from query parameter to URL path for proper signing (#303)
- Require `TrustProxy` setting for `X-Forwarded-Proto` in `scheme()` (#310)
- Resolve OpenAPI spec path relative to executable (#309)
- Fix logging for snippet tag setting errors by truncating overly long messages (#312)

### Changed
- Update GitHub Actions using sanad policy (#299)
- Upgrade Go depedencies

## [1.7.1] - 2026-07-11

### Fixed
- Resolve mobile editor layout overlap and toolbar row stacking

## [1.7.0] - 2026-07-04

### Removed

- Removed decryption with the pre-1.5 Gist credential key. Existing Gist
  credentials encrypted with that key must be saved again.
- Removed the deprecated `tag_id` and `folder_id` list filters. Use the
  comma-separated `tag_ids` and `folder_ids` parameters.
- Removed support for non-enveloped API responses in the web client and browser
  extension.
- Removed the retired Gist metadata-file exception, frontend compatibility
  globals, the no-op schema migration 6, and migration-specific repair paths.

### Fixed

- Editing items does appear crammed on mobile phones

### Development

-  Updated actions version to the latest accepted by sanad policy
- Removed the non-functional `migrate-down` Make target; schema downgrades are
  not supported.
- Removed the unused top-level `migrations` directory. Runtime migrations are
  defined only in `internal/database/migrations.go`.

## [1.6.1] - 2026-06-16

- Update Frontend vendor dependencies to latest versions, including Ace editor and Alpine.js.


## [1.6.0] - 2026-06-06

### Added
- Added `Ctrl+S` / `Cmd+S` save support in edit mode, including Ace editor command handling and duplicate-save protection.

### Changed
- Reworked edit mode so the code editor remains primary while language, folder, sharing, expiration, and tags live in a right-side details sidebar.
- Moved edit-mode description under the title and kept toolbar actions aligned with the title row.
- Updated edit-mode tags to render as wrapping chips, including quick-add suggestions.
- Expanded the edit-mode language dropdown to full width with clearer language labels.
- Made edit-mode file tabs more distinct with stronger borders.
- Removed tag rendering from snippet cards and moved copy controls plus single-file/multi-file indicators directly under the description or preview.
- Hydrated snippet files, tags, and folders in list responses so browse metadata and file indicators are available without opening each snippet.
- Updated Go dependencies: `golang.org/x/crypto`, `modernc.org/sqlite`, `github.com/go-chi/chi/v5`, and AWS SDK config/credentials packages.

### Fixed
- Fixed edit-mode file tab management for legacy single-file snippets, including Add File, switching between files, empty-editor states, and stale file-operation locks.
- Fixed edit-mode single-file snippets showing `+ Add file` instead of a filename when the snippet had not recently been edited.
- Improved Escape key handling so delete modals and search help close before the editor.


### Development
- Added Dependabot coverage for GitHub Actions updates.
- Updated GitHub Actions versions for Node setup, artifact upload/download, Docker login/build, Codecov, and golangci-lint workflows.
- Pinned GitHub Actions to full commit SHAs and added the action lock file.

## [1.5.0] - 2026-05-10

### Added
- Added POST endpoint for exporting encrypted backups with an optional password.
- Updated OpenAPI documentation to reflect new backup export options.

### Changed
- Improved backup export functionality adding encryption support.
- Refactored middleware to support anonymous access with admin password verification.

### Fixed
- Fixed API allowed origins (`SNIPO_ALLOWED_ORIGINS`) configuration parsing to resolve linting issues.
- Improved validation and error handling for backup operations.

### Security
- Improved security by sanitizing HTML in markdown rendering.

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

## [1.3.13] - 2026-03-22

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
