package database

// Migration represents a database migration
type Migration struct {
	Version int
	Name    string
	SQL     string
}

// Initial schema SQL
const initialSchemaSQL = `
-- Snipo Initial Schema
-- Version: 1

-- Snippets table - core entity
CREATE TABLE IF NOT EXISTS snippets (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(8)))),
    title TEXT NOT NULL,
    description TEXT DEFAULT '',
    content TEXT NOT NULL,
    language TEXT DEFAULT 'plaintext',
    is_favorite INTEGER DEFAULT 0,
    is_public INTEGER DEFAULT 0,
    view_count INTEGER DEFAULT 0,
    s3_key TEXT DEFAULT NULL,
    checksum TEXT DEFAULT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Tags table
CREATE TABLE IF NOT EXISTS tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    color TEXT DEFAULT '#6366f1',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Snippet-Tag junction table
CREATE TABLE IF NOT EXISTS snippet_tags (
    snippet_id TEXT NOT NULL,
    tag_id INTEGER NOT NULL,
    PRIMARY KEY (snippet_id, tag_id),
    FOREIGN KEY (snippet_id) REFERENCES snippets(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

-- Folders/Collections for organization
CREATE TABLE IF NOT EXISTS folders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    parent_id INTEGER DEFAULT NULL,
    icon TEXT DEFAULT 'folder',
    sort_order INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_id) REFERENCES folders(id) ON DELETE CASCADE
);

-- Snippet-Folder relationship
CREATE TABLE IF NOT EXISTS snippet_folders (
    snippet_id TEXT NOT NULL,
    folder_id INTEGER NOT NULL,
    PRIMARY KEY (snippet_id, folder_id),
    FOREIGN KEY (snippet_id) REFERENCES snippets(id) ON DELETE CASCADE,
    FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE CASCADE
);

-- Snippet files (multi-file support)
CREATE TABLE IF NOT EXISTS snippet_files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    snippet_id TEXT NOT NULL,
    filename TEXT NOT NULL,
    content TEXT NOT NULL,
    language TEXT DEFAULT 'plaintext',
    sort_order INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (snippet_id) REFERENCES snippets(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_snippet_files_snippet ON snippet_files(snippet_id);

-- Application settings (single row)
CREATE TABLE IF NOT EXISTS settings (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    app_name TEXT DEFAULT 'snipo',
    custom_css TEXT DEFAULT '',
    theme TEXT DEFAULT 'auto',
    default_language TEXT DEFAULT 'plaintext',
    s3_enabled INTEGER DEFAULT 0,
    s3_endpoint TEXT DEFAULT '',
    s3_bucket TEXT DEFAULT '',
    s3_region TEXT DEFAULT 'us-east-1',
    backup_encryption_enabled INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- API tokens for external access
CREATE TABLE IF NOT EXISTS api_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    token_hash TEXT UNIQUE NOT NULL,
    permissions TEXT DEFAULT 'read',
    last_used_at DATETIME DEFAULT NULL,
    expires_at DATETIME DEFAULT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Sessions table for web authentication
CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    token_hash TEXT UNIQUE NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_snippets_language ON snippets(language);
CREATE INDEX IF NOT EXISTS idx_snippets_favorite ON snippets(is_favorite);
CREATE INDEX IF NOT EXISTS idx_snippets_public ON snippets(is_public);
CREATE INDEX IF NOT EXISTS idx_snippets_created ON snippets(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_snippets_updated ON snippets(updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);
CREATE INDEX IF NOT EXISTS idx_folders_parent ON folders(parent_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);

-- Full-text search (external content FTS5 table)
CREATE VIRTUAL TABLE IF NOT EXISTS snippets_fts USING fts5(
    snippet_id,
    title,
    description,
    content,
    content='snippets',
    content_rowid='rowid'
);

-- Triggers to keep FTS in sync
CREATE TRIGGER IF NOT EXISTS snippets_ai AFTER INSERT ON snippets BEGIN
    INSERT INTO snippets_fts(rowid, snippet_id, title, description, content)
    VALUES (NEW.rowid, NEW.id, NEW.title, NEW.description, NEW.content);
END;

CREATE TRIGGER IF NOT EXISTS snippets_ad AFTER DELETE ON snippets BEGIN
    INSERT INTO snippets_fts(snippets_fts, rowid, snippet_id, title, description, content)
    VALUES('delete', OLD.rowid, OLD.id, OLD.title, OLD.description, OLD.content);
END;

CREATE TRIGGER IF NOT EXISTS snippets_au AFTER UPDATE ON snippets BEGIN
    INSERT INTO snippets_fts(snippets_fts, rowid, snippet_id, title, description, content)
    VALUES('delete', OLD.rowid, OLD.id, OLD.title, OLD.description, OLD.content);
    INSERT INTO snippets_fts(rowid, snippet_id, title, description, content)
    VALUES (NEW.rowid, NEW.id, NEW.title, NEW.description, NEW.content);
END;

-- Insert default settings row
INSERT OR IGNORE INTO settings (id) VALUES (1);
`

// Migration 2: Add snippet_files table for multi-file support
const addSnippetFilesSQL = `
-- Snippet files (multi-file support)
CREATE TABLE IF NOT EXISTS snippet_files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    snippet_id TEXT NOT NULL,
    filename TEXT NOT NULL,
    content TEXT NOT NULL,
    language TEXT DEFAULT 'plaintext',
    sort_order INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (snippet_id) REFERENCES snippets(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_snippet_files_snippet ON snippet_files(snippet_id);
`

// Migration 3: Add archiving support
const addArchivingSQL = `
-- Add is_archived column to snippets
ALTER TABLE snippets ADD COLUMN is_archived INTEGER DEFAULT 0;

-- Create index for archived snippets
CREATE INDEX IF NOT EXISTS idx_snippets_archived ON snippets(is_archived);

-- Add archive_enabled column to settings
ALTER TABLE settings ADD COLUMN archive_enabled INTEGER DEFAULT 0;
`

// Migration 4: Add snippet history support
const addHistorySQL = `
-- Create snippet_history table to track all modifications
CREATE TABLE IF NOT EXISTS snippet_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    snippet_id TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT DEFAULT '',
    content TEXT NOT NULL,
    language TEXT DEFAULT 'plaintext',
    is_favorite INTEGER DEFAULT 0,
    is_public INTEGER DEFAULT 0,
    is_archived INTEGER DEFAULT 0,
    change_type TEXT DEFAULT 'update',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (snippet_id) REFERENCES snippets(id) ON DELETE CASCADE
);

-- Create snippet_files_history table to track file modifications
CREATE TABLE IF NOT EXISTS snippet_files_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    history_id INTEGER NOT NULL,
    snippet_id TEXT NOT NULL,
    filename TEXT NOT NULL,
    content TEXT NOT NULL,
    language TEXT NOT NULL,
    sort_order INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (history_id) REFERENCES snippet_history(id) ON DELETE CASCADE,
    FOREIGN KEY (snippet_id) REFERENCES snippets(id) ON DELETE CASCADE
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_snippet_history_snippet_id ON snippet_history(snippet_id);
CREATE INDEX IF NOT EXISTS idx_snippet_history_created ON snippet_history(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_snippet_files_history_history_id ON snippet_files_history(history_id);
CREATE INDEX IF NOT EXISTS idx_snippet_files_history_snippet_id ON snippet_files_history(snippet_id);

-- Add history_enabled column to settings
ALTER TABLE settings ADD COLUMN history_enabled INTEGER DEFAULT 1;
`

// Migration 5: Add editor settings
const addEditorSettingsSQL = `
-- Add editor-related settings columns
ALTER TABLE settings ADD COLUMN editor_font_size INTEGER DEFAULT 14;
ALTER TABLE settings ADD COLUMN editor_tab_size INTEGER DEFAULT 2;
ALTER TABLE settings ADD COLUMN editor_theme TEXT DEFAULT 'auto';
ALTER TABLE settings ADD COLUMN editor_word_wrap INTEGER DEFAULT 1;
ALTER TABLE settings ADD COLUMN editor_show_print_margin INTEGER DEFAULT 0;
ALTER TABLE settings ADD COLUMN editor_show_gutter INTEGER DEFAULT 1;
ALTER TABLE settings ADD COLUMN editor_show_indent_guides INTEGER DEFAULT 1;
ALTER TABLE settings ADD COLUMN editor_highlight_active_line INTEGER DEFAULT 1;
ALTER TABLE settings ADD COLUMN editor_use_soft_tabs INTEGER DEFAULT 1;
ALTER TABLE settings ADD COLUMN editor_enable_snippets INTEGER DEFAULT 1;
ALTER TABLE settings ADD COLUMN editor_enable_live_autocompletion INTEGER DEFAULT 1;
ALTER TABLE settings ADD COLUMN markdown_font_size INTEGER DEFAULT 14;
`

// Migration 6: Add markdown settings (kept for backward compatibility)
const addMarkdownSettingsSQL = `
-- This migration was already applied in some databases
-- Keeping it as a no-op for backward compatibility
SELECT 1;
`

// Migration 7: Add disable_login setting
const addDisableLoginSQL = `
-- Add disable_login column to settings table
-- This allows users to disable the login page while keeping password protection
-- for administrative operations like API token creation
ALTER TABLE settings ADD COLUMN disable_login INTEGER DEFAULT 0 NOT NULL;
`

// Migration 8: Add exclude_first_line_on_copy setting

const addExcludeFirstLineSQL = `
-- Add exclude_first_line_on_copy column to settings table
ALTER TABLE settings ADD COLUMN exclude_first_line_on_copy INTEGER DEFAULT 0 NOT NULL;
`

// Migration 9: Add GitHub Gist sync tables
const addGistSyncSQL = `
-- Global sync configuration (single row)
CREATE TABLE IF NOT EXISTS gist_sync_config (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    enabled INTEGER DEFAULT 0,
    github_token_encrypted TEXT,
    github_username TEXT,
    auto_sync_enabled INTEGER DEFAULT 1,
    sync_interval_minutes INTEGER DEFAULT 15,
    conflict_strategy TEXT DEFAULT 'manual',
    last_full_sync_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Snippet to Gist mappings
CREATE TABLE IF NOT EXISTS snippet_gist_mappings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    snippet_id TEXT NOT NULL UNIQUE,
    gist_id TEXT NOT NULL UNIQUE,
    gist_url TEXT NOT NULL,
    sync_enabled INTEGER DEFAULT 1,
    last_synced_at DATETIME,
    snipo_checksum TEXT,
    gist_checksum TEXT,
    sync_status TEXT DEFAULT 'synced',
    error_message TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (snippet_id) REFERENCES snippets(id) ON DELETE CASCADE
);

-- Conflicts requiring manual resolution
CREATE TABLE IF NOT EXISTS gist_sync_conflicts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    snippet_id TEXT NOT NULL,
    gist_id TEXT NOT NULL,
    snipo_version TEXT,
    gist_version TEXT,
    resolved INTEGER DEFAULT 0,
    resolution_choice TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    resolved_at DATETIME,
    FOREIGN KEY (snippet_id) REFERENCES snippets(id) ON DELETE CASCADE
);

-- Sync operation logs
CREATE TABLE IF NOT EXISTS gist_sync_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    snippet_id TEXT,
    gist_id TEXT,
    operation TEXT NOT NULL,
    status TEXT NOT NULL,
    message TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_gist_mappings_snippet ON snippet_gist_mappings(snippet_id);
CREATE INDEX IF NOT EXISTS idx_gist_mappings_status ON snippet_gist_mappings(sync_status);
CREATE INDEX IF NOT EXISTS idx_gist_conflicts_resolved ON gist_sync_conflicts(resolved);
`

// getMigrations returns all available migrations in order
func getMigrations() []Migration {
	return []Migration{
		{Version: 1, Name: "initial_schema", SQL: initialSchemaSQL},
		{Version: 2, Name: "add_snippet_files", SQL: addSnippetFilesSQL},
		{Version: 3, Name: "add_archiving", SQL: addArchivingSQL},
		{Version: 4, Name: "add_history", SQL: addHistorySQL},
		{Version: 5, Name: "add_editor_settings", SQL: addEditorSettingsSQL},
		{Version: 6, Name: "add_markdown_settings", SQL: addMarkdownSettingsSQL},
		{Version: 7, Name: "add_disable_login", SQL: addDisableLoginSQL},
		{Version: 8, Name: "add_exclude_first_line", SQL: addExcludeFirstLineSQL},
		{Version: 9, Name: "add_gist_sync", SQL: addGistSyncSQL},
	}
}
