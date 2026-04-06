-- Snipo Migration: Add Snippet Expiration & Auto-Archive
-- Version: 9

-- Add expires_at column to snippets (nullable, no default)
ALTER TABLE snippets ADD COLUMN expires_at DATETIME DEFAULT NULL;

-- Add auto_archive_enabled to settings (default 0 = disabled)
ALTER TABLE settings ADD COLUMN auto_archive_enabled INTEGER DEFAULT 0;

-- Add default_expiration_days to settings (default 0 = no expiration)
ALTER TABLE settings ADD COLUMN default_expiration_days INTEGER DEFAULT 0;

-- Index for expires_at to speed up auto-archive queries
CREATE INDEX IF NOT EXISTS idx_snippets_expires_at ON snippets(expires_at);
