-- Snipo Migration: Add Snippet History
-- Version: 3

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
    change_type TEXT DEFAULT 'update', -- 'create', 'update', 'delete'
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
