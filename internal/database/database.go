package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB wraps the sql.DB with additional functionality
type DB struct {
	*sql.DB
	logger *slog.Logger
}

// Config holds database configuration
type Config struct {
	Path            string
	MaxOpenConns    int
	BusyTimeout     int
	JournalMode     string
	SynchronousMode string
}

// New creates a new database connection
func New(cfg Config, logger *slog.Logger) (*DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(cfg.Path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	// Build connection string with pragmas
	dsn := fmt.Sprintf("%s?_busy_timeout=%d&_journal_mode=%s&_synchronous=%s&_foreign_keys=ON",
		cfg.Path,
		cfg.BusyTimeout,
		cfg.JournalMode,
		cfg.SynchronousMode,
	)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(cfg.MaxOpenConns)

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Apply additional pragmas
	pragmas := []string{
		"PRAGMA cache_size = -2000",    // 2MB cache
		"PRAGMA temp_store = MEMORY",   // Temp tables in memory
		"PRAGMA mmap_size = 268435456", // 256MB memory-mapped I/O
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			logger.Warn("failed to set pragma", "pragma", pragma, "error", err)
		}
	}

	logger.Info("database connected", "path", cfg.Path)

	return &DB{DB: db, logger: logger}, nil
}

// Migrate runs all pending migrations
func (db *DB) Migrate(ctx context.Context) error {
	// Create migrations table if not exists
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get current version
	var currentVersion int
	row := db.QueryRowContext(ctx, "SELECT COALESCE(MAX(version), 0) FROM schema_migrations")
	if err := row.Scan(&currentVersion); err != nil {
		return fmt.Errorf("failed to get current migration version: %w", err)
	}

	db.logger.Info("current schema version", "version", currentVersion)

	// Run migrations
	migrations := getMigrations()
	for _, m := range migrations {
		if m.Version <= currentVersion {
			continue
		}

		db.logger.Info("applying migration", "version", m.Version, "name", m.Name)

		// Execute migration in a transaction
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		if _, err := tx.ExecContext(ctx, m.SQL); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to apply migration %d (%s): %w", m.Version, m.Name, err)
		}

		// Record migration
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO schema_migrations (version, name) VALUES (?, ?)",
			m.Version, m.Name,
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to record migration: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration: %w", err)
		}

		db.logger.Info("migration applied", "version", m.Version, "name", m.Name)
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	db.logger.Info("closing database connection")
	return db.DB.Close()
}
