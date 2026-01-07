package services

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/MohamedElashri/snipo/internal/repository"
	_ "modernc.org/sqlite"
)

func setupTestWorker(t *testing.T) (*GistSyncWorker, *sql.DB) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	schema := `
	CREATE TABLE gist_sync_config (
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

	CREATE TABLE snippet_gist_mappings (
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
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE snippets (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		description TEXT DEFAULT '',
		content TEXT NOT NULL,
		language TEXT DEFAULT 'plaintext',
		is_favorite INTEGER DEFAULT 0,
		is_public INTEGER DEFAULT 0,
		view_count INTEGER DEFAULT 0,
		is_archived INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	syncRepo := repository.NewGistSyncRepository(db)
	snippetRepo := repository.NewSnippetRepository(db)

	key := make([]byte, 32)
	encryptionSvc, err := NewEncryptionService(key)
	if err != nil {
		t.Fatalf("failed to create encryption service: %v", err)
	}

	fileRepo := repository.NewSnippetFileRepository(db)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	worker := NewGistSyncWorker(syncRepo, snippetRepo, fileRepo, encryptionSvc, logger)

	return worker, db
}

func TestGistSyncWorker_StartStop(t *testing.T) {
	worker, db := setupTestWorker(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	if worker.IsRunning() {
		t.Error("worker should not be running initially")
	}

	if err := worker.Start(ctx); err != nil {
		t.Fatalf("failed to start worker: %v", err)
	}

	if !worker.IsRunning() {
		t.Error("worker should be running after start")
	}

	time.Sleep(100 * time.Millisecond)

	if err := worker.Stop(); err != nil {
		t.Fatalf("failed to stop worker: %v", err)
	}

	if worker.IsRunning() {
		t.Error("worker should not be running after stop")
	}
}

func TestGistSyncWorker_MultipleStarts(t *testing.T) {
	worker, db := setupTestWorker(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	if err := worker.Start(ctx); err != nil {
		t.Fatalf("failed to start worker: %v", err)
	}

	if err := worker.Start(ctx); err != nil {
		t.Fatalf("second start should not error: %v", err)
	}

	if err := worker.Stop(); err != nil {
		t.Fatalf("failed to stop worker: %v", err)
	}
}

func TestGistSyncWorker_StopWithoutStart(t *testing.T) {
	worker, db := setupTestWorker(t)
	defer func() { _ = db.Close() }()

	if err := worker.Stop(); err != nil {
		t.Fatalf("stop without start should not error: %v", err)
	}
}
