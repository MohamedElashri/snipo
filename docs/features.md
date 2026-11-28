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

Open a Markdown snippet in a folder and choose **More → Set as folder
overview**. Snipo previews the first Markdown file at the top of that folder.
Use **Remove overview** to clear the assignment. The note stays in the folder.

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

In the Markdown preview, use **Copy** above a code block to copy that block on
its own, including its indentation.

Snipo also supports Markdown preview, custom CSS, and right-to-left text.
Custom CSS examples are in the [customization guide](customization.md).

Ace and CodeMirror adapt text direction to the content. Monaco keeps its standard
code layout. Text is saved as entered, including Arabic and mixed-language text.

## Notes

Choose **New note** and start writing. Markdown is the default format, and
**Plain text** preserves spacing and line breaks. Use **Preview** while writing
Markdown. If you leave the title blank, Snipo uses the first line of the body.
Notes start private, with no expiry or automatic Gist sync.

Use **Notes** in the sidebar to browse all notes. A note created from a folder
inherits that folder. Notes and snippets share folders and tags, and you can
filter a folder or tag by type. A Markdown note can also be a folder overview.

To bring an existing single-body Markdown or plain text snippet into Notes,
open **More → Convert to note**. The item keeps its link, history, folder, and
tags. **Convert to snippet** changes it back. Download a note as Markdown or
plain text from the same menu.

The browser save dialog can create a note or append a selection and its source
link to an existing private note. In VS Code, use the **Notes** view and
**Snipo: Save Note**. In the terminal, use **Tab** to switch to Notes and **n**
to create a note.

## Keyboard shortcuts

| Shortcut | Action |
|---|---|
| `/` or `Ctrl/Cmd+K` | Focus search |
| `Ctrl/Cmd+N` | New snippet |
| `Ctrl/Cmd+S` | Save while editing |
| `Escape` | Close the active editor, modal, or search help |

## Templates

Choose **From template…** to create a Markdown note, a script, or a multi-file
snippet. Search by name, purpose, or filename. Favorites appear first, followed
by recently saved templates. Fill the requested fields, pick a preset if one
is available, and review the rendered draft before saving. Saving creates a
private snippet with its own files, tags, and folder. Changes to the template
apply to future drafts.

Open **Templates** in the web sidebar to author templates, copy a starter, or
import a template file. From a snippet's **More** menu, choose **Save as
template** to start with all of its files. The library supports favorites,
archiving, Trash, revision history, and restoring an earlier revision. **Test
template** previews the current definition. **Export** downloads a portable
definition. Replacing a template from an import keeps its folder and any
preset values that still match the imported fields. Instance backups also
include presets, revision history, folder defaults, and the origin of generated
snippets. Restoring a pre-template
backup keeps existing templates and reconnects their defaults to matching
folder paths in the restored library.

Insert `{{snipo.service}}` to use a field named `service` in an output title,
description, filename, or file body. Use `{{!snipo.service}}` to write that
placeholder literally. Snipo substitutes text once and preserves ordinary
shell, Compose, and Helm expressions. Fields can be text, multiline text,
numbers, choices, toggles, or dates. Define fields and presets in the library
so the same form works across clients.

- In **Snippy**, press `t` in the snippet list. Use F2/F3 to switch files, F4 to
  copy, and Ctrl+S to save the draft.
- In **VS Code**, run **Snipo: Use Template**. Review files, open them as untitled
  editors, insert a file, or save the bundle to Snipo.
- In **Chrome and Firefox**, use the template picker in interactive **Save to
  Snipo**. A template can prefill fields from selected text, the page title,
  or the source URL.
