// Package testutil provides testing utilities for the snipo application.
package testutil

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"testing"

	_ "modernc.org/sqlite"
)

// TestDB creates an in-memory SQLite database for testing.
// It runs migrations and returns the database connection.
// The database is automatically closed when the test completes.
func TestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:?_foreign_keys=ON")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Run schema
	if err := runTestMigrations(db); err != nil {
		_ = db.Close()
		t.Fatalf("failed to run migrations: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

// runTestMigrations applies the database schema for testing
func runTestMigrations(db *sql.DB) error {
	schema := `
		-- Snippets table
		CREATE TABLE IF NOT EXISTS snippets (
			id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(8)))),
			title TEXT NOT NULL,
			description TEXT DEFAULT '',
			content TEXT NOT NULL DEFAULT '',
			language TEXT DEFAULT 'plaintext',
			is_favorite INTEGER DEFAULT 0,
			is_public INTEGER DEFAULT 0,
			is_archived INTEGER DEFAULT 0,
			view_count INTEGER DEFAULT 0,
			s3_key TEXT DEFAULT NULL,
			checksum TEXT DEFAULT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME DEFAULT NULL
		);

		-- Settings table
		CREATE TABLE IF NOT EXISTS settings (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			archive_enabled INTEGER DEFAULT 0,
			trash_enabled INTEGER DEFAULT 1,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		INSERT OR IGNORE INTO settings (id, archive_enabled, trash_enabled) VALUES (1, 0, 1);

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

		-- Folders table
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

		-- API tokens
		CREATE TABLE IF NOT EXISTS api_tokens (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			token_hash TEXT UNIQUE NOT NULL,
			permissions TEXT DEFAULT 'read',
			last_used_at DATETIME DEFAULT NULL,
			expires_at DATETIME DEFAULT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		-- Sessions table
		CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			token_hash TEXT UNIQUE NOT NULL,
			expires_at DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		-- Snippet files (multi-file support)
		CREATE TABLE IF NOT EXISTS snippet_files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			snippet_id TEXT NOT NULL,
			filename TEXT NOT NULL,
			content TEXT DEFAULT '',
			language TEXT DEFAULT 'plaintext',
			sort_order INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (snippet_id) REFERENCES snippets(id) ON DELETE CASCADE
		);

		-- Indexes
		CREATE INDEX IF NOT EXISTS idx_snippets_language ON snippets(language);
		CREATE INDEX IF NOT EXISTS idx_snippets_favorite ON snippets(is_favorite);
		CREATE INDEX IF NOT EXISTS idx_snippets_public ON snippets(is_public);
		CREATE INDEX IF NOT EXISTS idx_snippets_archived ON snippets(is_archived);
		CREATE INDEX IF NOT EXISTS idx_snippets_deleted ON snippets(deleted_at);
		CREATE INDEX IF NOT EXISTS idx_snippets_created ON snippets(created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_snippets_updated ON snippets(updated_at DESC);
		CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);
		CREATE INDEX IF NOT EXISTS idx_folders_parent ON folders(parent_id);
		CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);
		CREATE INDEX IF NOT EXISTS idx_snippet_files_snippet ON snippet_files(snippet_id);

		-- Full-text search
		CREATE VIRTUAL TABLE IF NOT EXISTS snippets_fts USING fts5(
			snippet_id,
			title,
			description,
			content,
			content='snippets',
			content_rowid='rowid'
		);

		-- FTS triggers
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
	`

	_, err := db.Exec(schema)
	return err
}

// TestLogger returns a no-op logger for testing
func TestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

// TestContext returns a context for testing
func TestContext() context.Context {
	return context.Background()
}
