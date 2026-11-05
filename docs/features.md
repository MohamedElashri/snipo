# Feature guide

This page covers behavior that is not obvious in the interface. API details are
in the [OpenAPI specification](openapi.yaml).

## Snippets and organization

A snippet can contain one legacy content field or multiple named files. Snipo
supports folders, tags, favorites, archiving, soft deletion, expiration, and
duplication. The default maximum is 10 files per snippet; operators can change
it with `SNIPO_MAX_FILES_PER_SNIPPET`.

History records pre-update snapshots when history is enabled in Settings.
Restoring an entry creates a new update; permanently deleting a snippet also
deletes its history.

## Search and filtering

The snippet list accepts a text query and filters:

```bash
?q=api
?tag_ids=1,2
?folder_ids=3
?language=python
?favorite=true
?is_archived=true
?sort=updated_at&order=desc
```

The web interface exposes these through search and the sidebar. Press `/` or
`Ctrl/Cmd+K` to focus search.

## Public snippets

Enable sharing on a snippet and distribute `/s/{id}`. Individual files are
available at:

```bash
/api/v1/snippets/public/{id}/files/{filename}
```

These routes require no authentication. Anyone with the URL can read the
published content, and disabling login does not add protection. Operators can
remove the public routes entirely with:

```bash
SNIPO_ENABLE_PUBLIC_SNIPPETS=false
```

## API tokens

Create tokens under Settings → API Tokens. The
[security guide](../SECURITY.md#sessions-and-api-tokens) explains permission and
credential handling; [openapi.yaml](openapi.yaml) defines authentication and is
also served at `/api/v1/openapi.json`.

## Backups

Settings can export JSON or ZIP and import using `replace`, `merge`, or `skip`.
Supplying a backup password produces an AES-256-GCM encrypted `.enc` file.
Restoration requires the password and the same `SNIPO_ENCRYPTION_SALT` used for
export; see [Data and backups](../SECURITY.md#data-and-backups).

S3-compatible storage can upload, list, restore, and delete backups when its
environment variables are configured. See the [deployment guide](deployment.md#s3-backups).

## GitHub Gist sync

Gist sync uses a GitHub personal access token with Gist access. Configure it
under Settings → GitHub Gist.

- Sync is enabled per snippet or in bulk; new snippets are not automatically
  enrolled.
- Background sync runs only when both sync and automatic sync are enabled.
- Conflict policies are manual, Snipo wins, Gist wins, or newest wins.
- Snipo metadata is embedded in the Gist description; files remain normal Gist
  files.

Removing a mapping stops synchronization; it does not delete either copy.

## Appearance and editing

Choose **Ace**, **CodeMirror 6**, or **Monaco** in **Settings → Editor → Code editor**.
Ace remains the default. Your choice is remembered in this browser, and switching
keeps unsaved code and local drafts. All three support multi-file snippets,
syntax highlighting, search, font size, line numbers, tabs, wrapping, and
Ctrl/Cmd+S to save. Switching editors or files starts a fresh undo history.

CodeMirror offers a compact editor; Monaco adds a VS Code-style experience with
JavaScript/TypeScript, JSON, HTML, and CSS language assistance. Both follow the
app’s light/dark theme. Ace also offers its existing named editor themes;
indent guides and print margins are available in Ace and Monaco.

Editor assets and language workers are served by your own Snipo instance.
CodeMirror and Monaco load only when selected. If a selected editor cannot load,
Snipo uses Ace and keeps your code.

Snipo also supports Markdown preview, custom CSS, and right-to-left text.
Custom CSS examples are in the [customization guide](customization.md).

RTL direction is selected from the text content. Mixed technical text uses the
browser's bidirectional layout rather than changing the stored snippet.

## Keyboard shortcuts

| Shortcut | Action |
|---|---|
| `/` or `Ctrl/Cmd+K` | Focus search |
| `Ctrl/Cmd+N` | New snippet |
| `Ctrl/Cmd+S` | Save while editing |
| `Escape` | Close the active editor, modal, or search help |
