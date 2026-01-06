package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/MohamedElashri/snipo/internal/models"
)

// GistSyncRepository handles gist sync database operations
type GistSyncRepository struct {
	db *sql.DB
}

// NewGistSyncRepository creates a new gist sync repository
func NewGistSyncRepository(db *sql.DB) *GistSyncRepository {
	return &GistSyncRepository{db: db}
}

// GetConfig retrieves the gist sync configuration
func (r *GistSyncRepository) GetConfig(ctx context.Context) (*models.GistSyncConfig, error) {
	query := `
		SELECT id, enabled, github_token_encrypted, github_username,
		       auto_sync_enabled, sync_interval_minutes, conflict_strategy,
		       last_full_sync_at, created_at, updated_at
		FROM gist_sync_config
		WHERE id = 1
	`

	config := &models.GistSyncConfig{}
	var lastFullSyncAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query).Scan(
		&config.ID,
		&config.Enabled,
		&config.GithubTokenEncrypted,
		&config.GithubUsername,
		&config.AutoSyncEnabled,
		&config.SyncIntervalMinutes,
		&config.ConflictResolutionStrategy,
		&lastFullSyncAt,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get gist sync config: %w", err)
	}

	if lastFullSyncAt.Valid {
		config.LastFullSyncAt = &lastFullSyncAt.Time
	}

	return config, nil
}

// CreateOrUpdateConfig creates or updates the gist sync configuration
func (r *GistSyncRepository) CreateOrUpdateConfig(ctx context.Context, config *models.GistSyncConfig) error {
	query := `
		INSERT INTO gist_sync_config (
			id, enabled, github_token_encrypted, github_username,
			auto_sync_enabled, sync_interval_minutes, conflict_strategy,
			last_full_sync_at, updated_at
		) VALUES (1, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(id) DO UPDATE SET
			enabled = excluded.enabled,
			github_token_encrypted = excluded.github_token_encrypted,
			github_username = excluded.github_username,
			auto_sync_enabled = excluded.auto_sync_enabled,
			sync_interval_minutes = excluded.sync_interval_minutes,
			conflict_strategy = excluded.conflict_strategy,
			last_full_sync_at = excluded.last_full_sync_at,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := r.db.ExecContext(ctx, query,
		config.Enabled,
		config.GithubTokenEncrypted,
		config.GithubUsername,
		config.AutoSyncEnabled,
		config.SyncIntervalMinutes,
		config.ConflictResolutionStrategy,
		config.LastFullSyncAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create or update gist sync config: %w", err)
	}

	return nil
}

// DeleteConfig deletes the gist sync configuration
func (r *GistSyncRepository) DeleteConfig(ctx context.Context) error {
	query := `DELETE FROM gist_sync_config WHERE id = 1`
	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete gist sync config: %w", err)
	}
	return nil
}

// CreateMapping creates a new snippet-gist mapping
func (r *GistSyncRepository) CreateMapping(ctx context.Context, mapping *models.SnippetGistMapping) error {
	query := `
		INSERT INTO snippet_gist_mappings (
			snippet_id, gist_id, gist_url, sync_enabled,
			snipo_checksum, gist_checksum, sync_status
		) VALUES (?, ?, ?, ?, ?, ?, ?)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		mapping.SnippetID,
		mapping.GistID,
		mapping.GistURL,
		mapping.SyncEnabled,
		mapping.SnipoChecksum,
		mapping.GistChecksum,
		mapping.SyncStatus,
	).Scan(&mapping.ID, &mapping.CreatedAt, &mapping.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create mapping: %w", err)
	}

	return nil
}

// GetMapping retrieves a mapping by snippet ID
func (r *GistSyncRepository) GetMapping(ctx context.Context, snippetID string) (*models.SnippetGistMapping, error) {
	query := `
		SELECT id, snippet_id, gist_id, gist_url, sync_enabled,
		       last_synced_at, snipo_checksum, gist_checksum,
		       sync_status, error_message, created_at, updated_at
		FROM snippet_gist_mappings
		WHERE snippet_id = ?
	`

	mapping := &models.SnippetGistMapping{}
	var lastSyncedAt sql.NullTime
	var errorMessage sql.NullString

	err := r.db.QueryRowContext(ctx, query, snippetID).Scan(
		&mapping.ID,
		&mapping.SnippetID,
		&mapping.GistID,
		&mapping.GistURL,
		&mapping.SyncEnabled,
		&lastSyncedAt,
		&mapping.SnipoChecksum,
		&mapping.GistChecksum,
		&mapping.SyncStatus,
		&errorMessage,
		&mapping.CreatedAt,
		&mapping.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get mapping: %w", err)
	}

	if lastSyncedAt.Valid {
		mapping.LastSyncedAt = &lastSyncedAt.Time
	}
	if errorMessage.Valid {
		mapping.ErrorMessage = &errorMessage.String
	}

	return mapping, nil
}

// GetMappingByGistID retrieves a mapping by gist ID
func (r *GistSyncRepository) GetMappingByGistID(ctx context.Context, gistID string) (*models.SnippetGistMapping, error) {
	query := `
		SELECT id, snippet_id, gist_id, gist_url, sync_enabled,
		       last_synced_at, snipo_checksum, gist_checksum,
		       sync_status, error_message, created_at, updated_at
		FROM snippet_gist_mappings
		WHERE gist_id = ?
	`

	mapping := &models.SnippetGistMapping{}
	var lastSyncedAt sql.NullTime
	var errorMessage sql.NullString

	err := r.db.QueryRowContext(ctx, query, gistID).Scan(
		&mapping.ID,
		&mapping.SnippetID,
		&mapping.GistID,
		&mapping.GistURL,
		&mapping.SyncEnabled,
		&lastSyncedAt,
		&mapping.SnipoChecksum,
		&mapping.GistChecksum,
		&mapping.SyncStatus,
		&errorMessage,
		&mapping.CreatedAt,
		&mapping.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get mapping by gist ID: %w", err)
	}

	if lastSyncedAt.Valid {
		mapping.LastSyncedAt = &lastSyncedAt.Time
	}
	if errorMessage.Valid {
		mapping.ErrorMessage = &errorMessage.String
	}

	return mapping, nil
}

// ListMappings retrieves all mappings
func (r *GistSyncRepository) ListMappings(ctx context.Context) ([]*models.SnippetGistMapping, error) {
	query := `
		SELECT id, snippet_id, gist_id, gist_url, sync_enabled,
		       last_synced_at, snipo_checksum, gist_checksum,
		       sync_status, error_message, created_at, updated_at
		FROM snippet_gist_mappings
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list mappings: %w", err)
	}
	defer rows.Close()

	var mappings []*models.SnippetGistMapping
	for rows.Next() {
		mapping := &models.SnippetGistMapping{}
		var lastSyncedAt sql.NullTime
		var errorMessage sql.NullString

		err := rows.Scan(
			&mapping.ID,
			&mapping.SnippetID,
			&mapping.GistID,
			&mapping.GistURL,
			&mapping.SyncEnabled,
			&lastSyncedAt,
			&mapping.SnipoChecksum,
			&mapping.GistChecksum,
			&mapping.SyncStatus,
			&errorMessage,
			&mapping.CreatedAt,
			&mapping.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan mapping: %w", err)
		}

		if lastSyncedAt.Valid {
			mapping.LastSyncedAt = &lastSyncedAt.Time
		}
		if errorMessage.Valid {
			mapping.ErrorMessage = &errorMessage.String
		}

		mappings = append(mappings, mapping)
	}

	return mappings, nil
}

// UpdateMapping updates an existing mapping
func (r *GistSyncRepository) UpdateMapping(ctx context.Context, mapping *models.SnippetGistMapping) error {
	query := `
		UPDATE snippet_gist_mappings
		SET sync_enabled = ?, last_synced_at = ?, snipo_checksum = ?,
		    gist_checksum = ?, sync_status = ?, error_message = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query,
		mapping.SyncEnabled,
		mapping.LastSyncedAt,
		mapping.SnipoChecksum,
		mapping.GistChecksum,
		mapping.SyncStatus,
		mapping.ErrorMessage,
		mapping.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update mapping: %w", err)
	}

	return nil
}

// DeleteMapping deletes a mapping
func (r *GistSyncRepository) DeleteMapping(ctx context.Context, id int64) error {
	query := `DELETE FROM snippet_gist_mappings WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete mapping: %w", err)
	}
	return nil
}

// CreateConflict creates a new sync conflict
func (r *GistSyncRepository) CreateConflict(ctx context.Context, conflict *models.GistSyncConflict) error {
	query := `
		INSERT INTO gist_sync_conflicts (
			snippet_id, gist_id, snipo_version, gist_version
		) VALUES (?, ?, ?, ?)
		RETURNING id, created_at
	`

	err := r.db.QueryRowContext(ctx, query,
		conflict.SnippetID,
		conflict.GistID,
		conflict.SnipoVersion,
		conflict.GistVersion,
	).Scan(&conflict.ID, &conflict.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create conflict: %w", err)
	}

	return nil
}

// GetConflict retrieves a conflict by ID
func (r *GistSyncRepository) GetConflict(ctx context.Context, id int64) (*models.GistSyncConflict, error) {
	query := `
		SELECT id, snippet_id, gist_id, snipo_version, gist_version,
		       resolved, resolution_choice, created_at, resolved_at
		FROM gist_sync_conflicts
		WHERE id = ?
	`

	conflict := &models.GistSyncConflict{}
	var resolutionChoice sql.NullString
	var resolvedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&conflict.ID,
		&conflict.SnippetID,
		&conflict.GistID,
		&conflict.SnipoVersion,
		&conflict.GistVersion,
		&conflict.Resolved,
		&resolutionChoice,
		&conflict.CreatedAt,
		&resolvedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get conflict: %w", err)
	}

	if resolutionChoice.Valid {
		conflict.ResolutionChoice = &resolutionChoice.String
	}
	if resolvedAt.Valid {
		conflict.ResolvedAt = &resolvedAt.Time
	}

	return conflict, nil
}

// ListConflicts retrieves all unresolved conflicts
func (r *GistSyncRepository) ListConflicts(ctx context.Context, resolvedOnly bool) ([]*models.GistSyncConflict, error) {
	query := `
		SELECT id, snippet_id, gist_id, snipo_version, gist_version,
		       resolved, resolution_choice, created_at, resolved_at
		FROM gist_sync_conflicts
		WHERE resolved = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, resolvedOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to list conflicts: %w", err)
	}
	defer rows.Close()

	var conflicts []*models.GistSyncConflict
	for rows.Next() {
		conflict := &models.GistSyncConflict{}
		var resolutionChoice sql.NullString
		var resolvedAt sql.NullTime

		err := rows.Scan(
			&conflict.ID,
			&conflict.SnippetID,
			&conflict.GistID,
			&conflict.SnipoVersion,
			&conflict.GistVersion,
			&conflict.Resolved,
			&resolutionChoice,
			&conflict.CreatedAt,
			&resolvedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conflict: %w", err)
		}

		if resolutionChoice.Valid {
			conflict.ResolutionChoice = &resolutionChoice.String
		}
		if resolvedAt.Valid {
			conflict.ResolvedAt = &resolvedAt.Time
		}

		conflicts = append(conflicts, conflict)
	}

	return conflicts, nil
}

// ResolveConflict marks a conflict as resolved
func (r *GistSyncRepository) ResolveConflict(ctx context.Context, id int64, resolution string) error {
	query := `
		UPDATE gist_sync_conflicts
		SET resolved = 1, resolution_choice = ?, resolved_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query, resolution, id)
	if err != nil {
		return fmt.Errorf("failed to resolve conflict: %w", err)
	}

	return nil
}

// CreateLog creates a new sync log entry
func (r *GistSyncRepository) CreateLog(ctx context.Context, log *models.GistSyncLog) error {
	query := `
		INSERT INTO gist_sync_log (
			snippet_id, gist_id, operation, status, message
		) VALUES (?, ?, ?, ?, ?)
		RETURNING id, created_at
	`

	err := r.db.QueryRowContext(ctx, query,
		log.SnippetID,
		log.GistID,
		log.Operation,
		log.Status,
		log.Message,
	).Scan(&log.ID, &log.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create log: %w", err)
	}

	return nil
}

// ListLogs retrieves sync logs with optional filters
func (r *GistSyncRepository) ListLogs(ctx context.Context, limit int) ([]*models.GistSyncLog, error) {
	query := `
		SELECT id, snippet_id, gist_id, operation, status, message, created_at
		FROM gist_sync_log
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list logs: %w", err)
	}
	defer rows.Close()

	var logs []*models.GistSyncLog
	for rows.Next() {
		log := &models.GistSyncLog{}
		var snippetID sql.NullString
		var gistID sql.NullString
		var message sql.NullString

		err := rows.Scan(
			&log.ID,
			&snippetID,
			&gistID,
			&log.Operation,
			&log.Status,
			&message,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan log: %w", err)
		}

		if snippetID.Valid {
			log.SnippetID = &snippetID.String
		}
		if gistID.Valid {
			log.GistID = &gistID.String
		}
		if message.Valid {
			log.Message = &message.String
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// UpdateLastFullSyncTime updates the last full sync timestamp
func (r *GistSyncRepository) UpdateLastFullSyncTime(ctx context.Context) error {
	query := `
		UPDATE gist_sync_config
		SET last_full_sync_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = 1
	`

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to update last full sync time: %w", err)
	}

	return nil
}

// GetEnabledMappings retrieves all mappings with sync enabled
func (r *GistSyncRepository) GetEnabledMappings(ctx context.Context) ([]*models.SnippetGistMapping, error) {
	query := `
		SELECT id, snippet_id, gist_id, gist_url, sync_enabled,
		       last_synced_at, snipo_checksum, gist_checksum,
		       sync_status, error_message, created_at, updated_at
		FROM snippet_gist_mappings
		WHERE sync_enabled = 1
		ORDER BY last_synced_at ASC NULLS FIRST
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled mappings: %w", err)
	}
	defer rows.Close()

	var mappings []*models.SnippetGistMapping
	for rows.Next() {
		mapping := &models.SnippetGistMapping{}
		var lastSyncedAt sql.NullTime
		var errorMessage sql.NullString

		err := rows.Scan(
			&mapping.ID,
			&mapping.SnippetID,
			&mapping.GistID,
			&mapping.GistURL,
			&mapping.SyncEnabled,
			&lastSyncedAt,
			&mapping.SnipoChecksum,
			&mapping.GistChecksum,
			&mapping.SyncStatus,
			&errorMessage,
			&mapping.CreatedAt,
			&mapping.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan mapping: %w", err)
		}

		if lastSyncedAt.Valid {
			mapping.LastSyncedAt = &lastSyncedAt.Time
		}
		if errorMessage.Valid {
			mapping.ErrorMessage = &errorMessage.String
		}

		mappings = append(mappings, mapping)
	}

	return mappings, nil
}
