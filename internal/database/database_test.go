package database

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"testing"
)

func TestMigrateFromEverySchemaVersion(t *testing.T) {
	for version := 0; version <= 11; version++ {
		t.Run(fmt.Sprintf("version_%d", version), func(t *testing.T) {
			db := newMigrationTestDB(t)
			prepareSchemaVersion(t, db.DB, version)

			if err := db.Migrate(context.Background()); err != nil {
				t.Fatalf("migrate schema %d: %v", version, err)
			}

			assertCurrentSchema(t, db.DB)

			// Starting an already-migrated application must remain a no-op.
			if err := db.Migrate(context.Background()); err != nil {
				t.Fatalf("repeat migration from schema %d: %v", version, err)
			}

			if version >= 1 {
				var title string
				if err := db.QueryRow("SELECT title FROM snippets WHERE id = 'migration-test'").Scan(&title); err != nil {
					t.Fatalf("read data after schema %d migration: %v", version, err)
				}
				if title != "Preserved snippet" {
					t.Fatalf("unexpected migrated title %q", title)
				}
			}
		})
	}
}

func TestMigrateSkipsRemovedNoOpVersionSix(t *testing.T) {
	t.Run("fresh database", func(t *testing.T) {
		db := newMigrationTestDB(t)
		if err := db.Migrate(context.Background()); err != nil {
			t.Fatalf("migrate fresh database: %v", err)
		}

		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = 6").Scan(&count); err != nil {
			t.Fatalf("query migration 6: %v", err)
		}
		if count != 0 {
			t.Fatal("fresh database unexpectedly recorded removed migration 6")
		}
	})

	t.Run("existing version six database", func(t *testing.T) {
		db := newMigrationTestDB(t)
		prepareSchemaVersion(t, db.DB, 6)

		if err := db.Migrate(context.Background()); err != nil {
			t.Fatalf("migrate version 6 database: %v", err)
		}

		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = 6").Scan(&count); err != nil {
			t.Fatalf("query migration 6: %v", err)
		}
		if count != 1 {
			t.Fatalf("expected existing migration 6 record to remain, got %d", count)
		}
		assertCurrentSchema(t, db.DB)
	})
}

func newMigrationTestDB(t *testing.T) *DB {
	t.Helper()

	raw, err := sql.Open("sqlite", ":memory:?_foreign_keys=ON")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	raw.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = raw.Close()
	})

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return &DB{DB: raw, logger: logger}
}

func prepareSchemaVersion(t *testing.T, db *sql.DB, target int) {
	t.Helper()
	if target == 0 {
		return
	}

	if _, err := db.Exec(`
		CREATE TABLE schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		t.Fatalf("create migration table: %v", err)
	}

	for _, migration := range getMigrations() {
		if migration.Version > target {
			break
		}
		if _, err := db.Exec(migration.SQL); err != nil {
			t.Fatalf("apply migration %d: %v", migration.Version, err)
		}
		if _, err := db.Exec(
			"INSERT INTO schema_migrations (version, name) VALUES (?, ?)",
			migration.Version,
			migration.Name,
		); err != nil {
			t.Fatalf("record migration %d: %v", migration.Version, err)
		}
	}

	// Version 6 was a released no-op. It is no longer in getMigrations, but
	// databases that recorded it must continue upgrading from their max version.
	if target >= 6 {
		if _, err := db.Exec(
			"INSERT INTO schema_migrations (version, name) VALUES (6, 'add_markdown_settings')",
		); err != nil {
			t.Fatalf("record migration 6: %v", err)
		}
	}

	if _, err := db.Exec(`
		INSERT INTO snippets (id, title, content, language)
		VALUES ('migration-test', 'Preserved snippet', 'hello', 'plaintext')
	`); err != nil {
		t.Fatalf("seed schema %d: %v", target, err)
	}
}

func assertCurrentSchema(t *testing.T, db *sql.DB) {
	t.Helper()

	var version int
	if err := db.QueryRow("SELECT MAX(version) FROM schema_migrations").Scan(&version); err != nil {
		t.Fatalf("read current version: %v", err)
	}
	if version != 11 {
		t.Fatalf("expected schema version 11, got %d", version)
	}

	for table, column := range map[string]string{
		"snippets": "expires_at",
		"settings": "default_expiration_days",
	} {
		var count int
		query := fmt.Sprintf("SELECT COUNT(*) FROM pragma_table_info('%s') WHERE name = ?", table)
		if err := db.QueryRow(query, column).Scan(&count); err != nil {
			t.Fatalf("inspect %s.%s: %v", table, column, err)
		}
		if count != 1 {
			t.Fatalf("expected column %s.%s", table, column)
		}
	}

	for _, table := range []string{
		"snippet_files",
		"snippet_history",
		"gist_sync_config",
		"snippet_gist_mappings",
	} {
		var count int
		if err := db.QueryRow(
			"SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?",
			table,
		).Scan(&count); err != nil {
			t.Fatalf("inspect table %s: %v", table, err)
		}
		if count != 1 {
			t.Fatalf("expected table %s", table)
		}
	}
}
