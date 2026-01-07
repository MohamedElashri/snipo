-- Add GitHub Gist sync tables
-- Migration: 007

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
