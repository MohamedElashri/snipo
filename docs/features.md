# Features Guide

## Search

Snipo features powerful fuzzy search that searches across:
- Snippet titles, descriptions, and content
- Multi-file snippet contents
- File names

### Basic Search
Type keywords in the search bar. Multiple words are matched using AND logic:
```
python docker
```
Finds snippets containing both "python" and "docker" anywhere in the metadata or content.

### Filters

**By Tags:**
```
?tag_id=1              # Single tag
?tag_ids=1,2,3         # Multiple tags
```

**By Folders:**
```
?folder_id=1           # Single folder
?folder_ids=1,2,3      # Multiple folders
```

**By Language:**
```
?language=javascript
```

**By Status:**
```
?favorite=true         # Favorites only
?is_archived=true      # Archived snippets
```

### Combining Filters
Mix search with filters for precise results:
```
?q=api&tag_id=1&language=python
```
Searches for "api" in Python snippets with tag 1.

### Sorting
```
?sort=title&order=asc  # A-Z by title
?sort=updated_at       # Recently updated (default)
?sort=created_at       # Recently created
```

**In-app help:** Click the `?` icon next to the search bar for interactive documentation.

## Public Snippets

Share code snippets publicly with granular file-level access.

### Making Snippets Public

**From Editor:**
1. Open or create a snippet
2. Toggle the "Public" switch in the editor header
3. Share the generated URL: `https://localhost:8080/s/{snippet-id}`

**From Preview Page:**
- Click the globe icon in the header to toggle public/private status
- Requires authentication to change visibility

### Accessing Public Snippets

**Web Interface:**
- Multi-file snippets display with file tabs
- Switch between files using the tab interface
- Download individual files with the download button
- Copy file URLs for direct access

**Direct File Access (wget/curl):**

For single-file snippets:
```bash
# Download the file
curl -O https://localhost:8080/api/v1/snippets/public/{snippet-id}/files/{filename}

# Or with wget
wget https://localhost:8080/api/v1/snippets/public/{snippet-id}/files/{filename}
```

For multi-file snippets:
```bash
# Download specific file
curl -O https://localhost:8080/api/v1/snippets/public/abc123/files/config.yaml

# Download all files using the file URLs from the web interface
curl -O https://localhost:8080/api/v1/snippets/public/abc123/files/main.go
curl -O https://localhost:8080/api/v1/snippets/public/abc123/files/README.md
```

**URL Format:**
- Snippet preview: `/s/{snippet-id}`
- Individual file (raw): `/api/v1/snippets/public/{snippet-id}/files/{filename}`

**Permissions:**
- Public snippets are accessible without authentication
- View count is tracked automatically
- Files are returned as plain text with proper Content-Disposition headers

## GitHub Gist Sync

Snipo supports two-way synchronization with GitHub Gists, allowing you to backup your snippets to GitHub and keep them in sync across platforms.

### Setup

1. **Generate GitHub Personal Access Token:**
   - Go to [GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)](https://github.com/settings/tokens/new?scopes=gist&description=Snipo)
   - Create a token with `gist` scope
   - Copy the generated token

2. **Configure in Snipo:**
   - Go to Settings → GitHub Gist tab
   - Paste your GitHub token
   - Click "Save Configuration" (sync is automatically enabled)

### Features

**Sync Options:**
- **Enable Sync for All**: Creates GitHub gists for all your snippets at once
- **Sync Now**: Syncs changes for already-enabled snippets
- **Auto-Sync**: Background sync at configurable intervals (5/15/30/60 minutes)

**Conflict Resolution:**
- **Manual**: Review and resolve conflicts manually
- **Snipo Wins**: Always keep Snipo version
- **Gist Wins**: Always keep GitHub version
- **Newest Wins**: Keep the most recently modified version

**Metadata Preservation:**
- Snippet titles, descriptions, and content are synced
- Snipo-specific metadata (favorites, folders, tags) embedded in gist description
- Multi-file snippets fully supported

### Usage

**Enable sync for all snippets:**
```
Settings → GitHub Gist → Enable Sync for All
```

**View synced snippets:**
- See list of synced snippets with status badges (✓ synced, ⟳ pending, ⚠ conflict, ✗ error)
- Click gist links to view on GitHub
- Remove mappings to stop syncing specific snippets

**Manage conflicts:**
- Conflicts appear when both Snipo and GitHub versions are modified
- Choose "Keep Snipo" or "Keep Gist" to resolve
- Or set automatic conflict resolution strategy in settings

### Limitations

- Requires GitHub Personal Access Token (no OAuth)
- Sync is per-snippet, not automatic for new snippets
- GitHub API rate limit: 5000 requests/hour

## API

Create API tokens in Settings → API Tokens with granular permissions:
- **read**: View snippets, tags, folders
- **write**: Create, update, delete resources
- **admin**: Full access including settings

Authenticate via:
- `Authorization: Bearer <token>`
- `X-API-Key: <key>`

All responses include metadata (request ID, timestamp, version) and pagination for lists.

API documentation:
- OpenAPI spec: [`openapi.yaml`](openapi.yaml)
- Interactive docs: `http://localhost:8080/api/v1/openapi.json`
