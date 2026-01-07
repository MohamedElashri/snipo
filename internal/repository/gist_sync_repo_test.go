package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/MohamedElashri/snipo/internal/models"
	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
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

	CREATE TABLE gist_sync_conflicts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		snippet_id TEXT NOT NULL,
		gist_id TEXT NOT NULL,
		snipo_version TEXT,
		gist_version TEXT,
		resolved INTEGER DEFAULT 0,
		resolution_choice TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		resolved_at DATETIME
	);

	CREATE TABLE gist_sync_log (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		snippet_id TEXT,
		gist_id TEXT,
		operation TEXT NOT NULL,
		status TEXT NOT NULL,
		message TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func TestGistSyncRepository_Config(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	repo := NewGistSyncRepository(db)
	ctx := context.Background()

	t.Run("get non-existent config", func(t *testing.T) {
		config, err := repo.GetConfig(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if config != nil {
			t.Error("expected nil config for non-existent record")
		}
	})

	t.Run("create config", func(t *testing.T) {
		config := &models.GistSyncConfig{
			Enabled:                    true,
			GithubTokenEncrypted:       "encrypted-token",
			GithubUsername:             "testuser",
			AutoSyncEnabled:            true,
			SyncIntervalMinutes:        15,
			ConflictResolutionStrategy: models.ConflictStrategyManual,
		}

		err := repo.CreateOrUpdateConfig(ctx, config)
		if err != nil {
			t.Fatalf("failed to create config: %v", err)
		}

		retrieved, err := repo.GetConfig(ctx)
		if err != nil {
			t.Fatalf("failed to get config: %v", err)
		}

		if retrieved.GithubUsername != "testuser" {
			t.Errorf("expected username 'testuser', got '%s'", retrieved.GithubUsername)
		}
	})
}

func TestGistSyncRepository_Mapping(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	repo := NewGistSyncRepository(db)
	ctx := context.Background()

	t.Run("create and get mapping", func(t *testing.T) {
		mapping := &models.SnippetGistMapping{
			SnippetID:     "snippet-123",
			GistID:        "gist-456",
			GistURL:       "https://gist.github.com/user/gist-456",
			SyncEnabled:   true,
			SnipoChecksum: "checksum1",
			GistChecksum:  "checksum2",
			SyncStatus:    models.SyncStatusSynced,
		}

		err := repo.CreateMapping(ctx, mapping)
		if err != nil {
			t.Fatalf("failed to create mapping: %v", err)
		}

		if mapping.ID == 0 {
			t.Error("expected mapping ID to be set")
		}

		retrieved, err := repo.GetMapping(ctx, "snippet-123")
		if err != nil {
			t.Fatalf("failed to get mapping: %v", err)
		}

		if retrieved.GistID != "gist-456" {
			t.Errorf("expected gist ID 'gist-456', got '%s'", retrieved.GistID)
		}
	})

	t.Run("update mapping", func(t *testing.T) {
		mapping, _ := repo.GetMapping(ctx, "snippet-123")
		mapping.SyncStatus = models.SyncStatusPending
		now := time.Now()
		mapping.LastSyncedAt = &now

		err := repo.UpdateMapping(ctx, mapping)
		if err != nil {
			t.Fatalf("failed to update mapping: %v", err)
		}

		retrieved, _ := repo.GetMapping(ctx, "snippet-123")
		if retrieved.SyncStatus != models.SyncStatusPending {
			t.Errorf("expected status 'pending', got '%s'", retrieved.SyncStatus)
		}
	})

	t.Run("list mappings", func(t *testing.T) {
		mappings, err := repo.ListMappings(ctx)
		if err != nil {
			t.Fatalf("failed to list mappings: %v", err)
		}

		if len(mappings) != 1 {
			t.Errorf("expected 1 mapping, got %d", len(mappings))
		}
	})
}

func TestGistSyncRepository_Conflict(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	repo := NewGistSyncRepository(db)
	ctx := context.Background()

	t.Run("create and resolve conflict", func(t *testing.T) {
		conflict := &models.GistSyncConflict{
			SnippetID:    "snippet-123",
			GistID:       "gist-456",
			SnipoVersion: `{"title":"v1"}`,
			GistVersion:  `{"title":"v2"}`,
		}

		err := repo.CreateConflict(ctx, conflict)
		if err != nil {
			t.Fatalf("failed to create conflict: %v", err)
		}

		conflicts, err := repo.ListConflicts(ctx, false)
		if err != nil {
			t.Fatalf("failed to list conflicts: %v", err)
		}

		if len(conflicts) != 1 {
			t.Errorf("expected 1 conflict, got %d", len(conflicts))
		}

		err = repo.ResolveConflict(ctx, conflict.ID, "snipo_wins")
		if err != nil {
			t.Fatalf("failed to resolve conflict: %v", err)
		}

		resolved, err := repo.GetConflict(ctx, conflict.ID)
		if err != nil {
			t.Fatalf("failed to get conflict: %v", err)
		}

		if !resolved.Resolved {
			t.Error("expected conflict to be resolved")
		}
	})
}

func TestGistSyncRepository_Log(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	repo := NewGistSyncRepository(db)
	ctx := context.Background()

	t.Run("create and list logs", func(t *testing.T) {
		snippetID := "snippet-123"
		gistID := "gist-456"
		message := "sync completed"

		log := &models.GistSyncLog{
			SnippetID: &snippetID,
			GistID:    &gistID,
			Operation: models.SyncOpSync,
			Status:    models.SyncOpStatusSuccess,
			Message:   &message,
		}

		err := repo.CreateLog(ctx, log)
		if err != nil {
			t.Fatalf("failed to create log: %v", err)
		}

		logs, err := repo.ListLogs(ctx, 10)
		if err != nil {
			t.Fatalf("failed to list logs: %v", err)
		}

		if len(logs) != 1 {
			t.Errorf("expected 1 log, got %d", len(logs))
		}

		if *logs[0].Message != "sync completed" {
			t.Errorf("expected message 'sync completed', got '%s'", *logs[0].Message)
		}
	})
}
