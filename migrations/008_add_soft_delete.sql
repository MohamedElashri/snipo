-- Add deleted_at column to snippets table
ALTER TABLE snippets ADD COLUMN deleted_at DATETIME DEFAULT NULL;

-- Add trash_enabled column to settings table (default 1 = enabled)
ALTER TABLE settings ADD COLUMN trash_enabled INTEGER DEFAULT 1;

-- Index for deleted_at to speed up cleanup and filtering
CREATE INDEX IF NOT EXISTS idx_snippets_deleted_at ON snippets(deleted_at);
