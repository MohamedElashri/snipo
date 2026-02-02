package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/MohamedElashri/snipo/internal/models"
)

// SnippetRepository handles snippet database operations
type SnippetRepository struct {
	db *sql.DB
}

// NewSnippetRepository creates a new snippet repository
func NewSnippetRepository(db *sql.DB) *SnippetRepository {
	return &SnippetRepository{db: db}
}

// Create inserts a new snippet
func (r *SnippetRepository) Create(ctx context.Context, input *models.SnippetInput) (*models.Snippet, error) {
	query := `
		INSERT INTO snippets (title, description, content, language, is_public, is_archived)
		VALUES (?, ?, ?, ?, ?, ?)
		RETURNING id, title, description, content, language, is_favorite, is_public, 
		          view_count, s3_key, checksum, is_archived, created_at, updated_at, deleted_at
	`

	snippet := &models.Snippet{}
	err := r.db.QueryRowContext(ctx, query,
		input.Title,
		input.Description,
		input.Content,
		input.Language,
		input.IsPublic,
		input.IsArchived,
	).Scan(
		&snippet.ID,
		&snippet.Title,
		&snippet.Description,
		&snippet.Content,
		&snippet.Language,
		&snippet.IsFavorite,
		&snippet.IsPublic,
		&snippet.ViewCount,
		&snippet.S3Key,
		&snippet.Checksum,
		&snippet.IsArchived,
		&snippet.CreatedAt,
		&snippet.UpdatedAt,
		&snippet.DeletedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create snippet: %w", err)
	}

	return snippet, nil
}

// GetByID retrieves a snippet by ID
func (r *SnippetRepository) GetByID(ctx context.Context, id string) (*models.Snippet, error) {
	query := `
		SELECT id, title, description, content, language, is_favorite, is_public,
		       view_count, s3_key, checksum, is_archived, created_at, updated_at, deleted_at
		FROM snippets
		WHERE id = ?
	`

	snippet := &models.Snippet{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&snippet.ID,
		&snippet.Title,
		&snippet.Description,
		&snippet.Content,
		&snippet.Language,
		&snippet.IsFavorite,
		&snippet.IsPublic,
		&snippet.ViewCount,
		&snippet.S3Key,
		&snippet.Checksum,
		&snippet.IsArchived,
		&snippet.CreatedAt,
		&snippet.UpdatedAt,
		&snippet.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get snippet: %w", err)
	}

	return snippet, nil
}

// Update updates an existing snippet
func (r *SnippetRepository) Update(ctx context.Context, id string, input *models.SnippetInput) (*models.Snippet, error) {
	query := `
		UPDATE snippets
		SET title = ?, description = ?, content = ?, language = ?, is_public = ?, is_archived = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		RETURNING id, title, description, content, language, is_favorite, is_public,
		          view_count, s3_key, checksum, is_archived, created_at, updated_at, deleted_at
	`

	snippet := &models.Snippet{}
	err := r.db.QueryRowContext(ctx, query,
		input.Title,
		input.Description,
		input.Content,
		input.Language,
		input.IsPublic,
		input.IsArchived,
		id,
	).Scan(
		&snippet.ID,
		&snippet.Title,
		&snippet.Description,
		&snippet.Content,
		&snippet.Language,
		&snippet.IsFavorite,
		&snippet.IsPublic,
		&snippet.ViewCount,
		&snippet.S3Key,
		&snippet.Checksum,
		&snippet.IsArchived,
		&snippet.CreatedAt,
		&snippet.UpdatedAt,
		&snippet.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update snippet: %w", err)
	}

	return snippet, nil
}

// Delete removes a snippet by ID (soft delete if trash enabled)
// If permanent is true, it forces a hard delete regardless of settings
func (r *SnippetRepository) Delete(ctx context.Context, id string, permanent bool) error {
	// Check if trash is enabled
	var trashEnabled bool
	err := r.db.QueryRowContext(ctx, "SELECT trash_enabled FROM settings WHERE id = 1").Scan(&trashEnabled)
	if err != nil {
		return fmt.Errorf("failed to check trash settings: %w", err)
	}

	// Soft delete if enabled and not forced permanent
	if trashEnabled && !permanent {
		query := `
            UPDATE snippets 
            SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP 
            WHERE id = ? AND deleted_at IS NULL
        `
		result, err := r.db.ExecContext(ctx, query, id)
		if err != nil {
			return fmt.Errorf("failed to soft delete snippet: %w", err)
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}
		if rows == 0 {
			return sql.ErrNoRows
		}
		return nil
	}

	// Hard delete (original logic)
	// Start transaction for atomic delete
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Delete related data first (in case CASCADE doesn't work)
	_, _ = tx.ExecContext(ctx, "DELETE FROM snippet_tags WHERE snippet_id = ?", id)
	_, _ = tx.ExecContext(ctx, "DELETE FROM snippet_folders WHERE snippet_id = ?", id)
	_, _ = tx.ExecContext(ctx, "DELETE FROM snippet_files WHERE snippet_id = ?", id)

	// Delete the snippet
	result, err := tx.ExecContext(ctx, "DELETE FROM snippets WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete snippet: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Restore restores a soft-deleted snippet
func (r *SnippetRepository) Restore(ctx context.Context, id string) error {
	query := `
        UPDATE snippets 
        SET deleted_at = NULL, updated_at = CURRENT_TIMESTAMP 
        WHERE id = ? AND deleted_at IS NOT NULL
    `
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to restore snippet: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// CleanupDeleted permanently deletes snippets older than the specified duration
func (r *SnippetRepository) CleanupDeleted(ctx context.Context, days int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -days)

	// Using transaction for safety
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Find IDs to delete
	rows, err := tx.QueryContext(ctx, "SELECT id FROM snippets WHERE deleted_at < ?", cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to query old snippets: %w", err)
	}

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			_ = rows.Close()
			return 0, err
		}
		ids = append(ids, id)
	}
	_ = rows.Close()

	if len(ids) == 0 {
		return 0, nil
	}

	// Delete related data
	// Note: This could be optimized with batch deletes or ensuring cascading deletes work
	for _, id := range ids {
		_, _ = tx.ExecContext(ctx, "DELETE FROM snippet_tags WHERE snippet_id = ?", id)
		_, _ = tx.ExecContext(ctx, "DELETE FROM snippet_folders WHERE snippet_id = ?", id)
		_, _ = tx.ExecContext(ctx, "DELETE FROM snippet_files WHERE snippet_id = ?", id)
	}

	// Delete snippets
	query := fmt.Sprintf("DELETE FROM snippets WHERE id IN ('%s')", strings.Join(ids, "','"))
	result, err := tx.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete snippets: %w", err)
	}

	deletedCount, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return deletedCount, nil
}

// Allowed sort columns - maps user input to safe SQL column identifiers
// This prevents SQL injection by only allowing predefined column names
var allowedSortColumns = map[string]string{
	"id":          "id",
	"title":       "title",
	"description": "description",
	"content":     "content",
	"language":    "language",
	"is_favorite": "is_favorite",
	"is_public":   "is_public",
	"view_count":  "view_count",
	"created_at":  "created_at",
	"updated_at":  "updated_at",
	"deleted_at":  "deleted_at",
}

// List retrieves snippets with filtering and pagination
func (r *SnippetRepository) List(ctx context.Context, filter models.SnippetFilter) (*models.SnippetListResponse, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}

	// Map user-provided sort column to safe SQL column name
	// This prevents SQL injection by using a constant value from allowedSortColumns
	sortColumn, ok := allowedSortColumns[filter.SortBy]
	if !ok {
		sortColumn = "updated_at"
	}

	// Validate sort order using constant values
	sortOrder := "DESC"
	if filter.SortOrder == "asc" {
		sortOrder = "ASC"
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}

	// Build query
	var conditions []string
	var args []interface{}

	// Filter by deletion status
	if filter.IsDeleted != nil && *filter.IsDeleted {
		conditions = append(conditions, "s.deleted_at IS NOT NULL")
	} else {
		conditions = append(conditions, "s.deleted_at IS NULL")
	}

	// Fuzzy search on title, description, content, and snippet files
	if filter.Query != "" {
		// Split query into words for fuzzy matching
		words := strings.Fields(filter.Query)
		var searchConditions []string
		for _, word := range words {
			fuzzyPattern := "%" + word + "%"
			// Search in snippet metadata and files
			searchConditions = append(searchConditions,
				"(s.title LIKE ? OR s.description LIKE ? OR s.content LIKE ? OR "+
					"s.id IN (SELECT snippet_id FROM snippet_files WHERE content LIKE ? OR filename LIKE ?))")
			args = append(args, fuzzyPattern, fuzzyPattern, fuzzyPattern, fuzzyPattern, fuzzyPattern)
		}
		if len(searchConditions) > 0 {
			conditions = append(conditions, "("+strings.Join(searchConditions, " AND ")+")")
		}
	}

	if filter.Language != "" {
		conditions = append(conditions, "s.language = ?")
		args = append(args, filter.Language)
	}

	if filter.IsFavorite != nil {
		conditions = append(conditions, "s.is_favorite = ?")
		if *filter.IsFavorite {
			args = append(args, 1)
		} else {
			args = append(args, 0)
		}
	}

	if filter.IsPublic != nil {
		conditions = append(conditions, "s.is_public = ?")
		if *filter.IsPublic {
			args = append(args, 1)
		} else {
			args = append(args, 0)
		}
	}

	if filter.IsArchived != nil {
		conditions = append(conditions, "s.is_archived = ?")
		if *filter.IsArchived {
			args = append(args, 1)
		} else {
			args = append(args, 0)
		}
	} else if filter.IsDeleted == nil || !*filter.IsDeleted {
		// Default: hide archived unless we are looking at deleted items (which might be archived?) or explicitly filtering
		// Actually typical behavior: Archive view shows archived, All view hides archived.
		// If IsArchived is nil, we default to hidden.
		conditions = append(conditions, "s.is_archived = 0")
	}

	// Filter by tag (support both single and multiple tags)
	if filter.TagID > 0 {
		conditions = append(conditions, "s.id IN (SELECT snippet_id FROM snippet_tags WHERE tag_id = ?)")
		args = append(args, filter.TagID)
	} else if len(filter.TagIDs) > 0 {
		placeholders := make([]string, len(filter.TagIDs))
		for i, tagID := range filter.TagIDs {
			placeholders[i] = "?"
			args = append(args, tagID)
		}
		conditions = append(conditions, fmt.Sprintf("s.id IN (SELECT snippet_id FROM snippet_tags WHERE tag_id IN (%s))", strings.Join(placeholders, ",")))
	}

	// Filter by folder (support both single and multiple folders)
	if filter.FolderID > 0 {
		conditions = append(conditions, "s.id IN (SELECT snippet_id FROM snippet_folders WHERE folder_id = ?)")
		args = append(args, filter.FolderID)
	} else if len(filter.FolderIDs) > 0 {
		placeholders := make([]string, len(filter.FolderIDs))
		for i, folderID := range filter.FolderIDs {
			placeholders[i] = "?"
			args = append(args, folderID)
		}
		conditions = append(conditions, fmt.Sprintf("s.id IN (SELECT snippet_id FROM snippet_folders WHERE folder_id IN (%s))", strings.Join(placeholders, ",")))
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM snippets s %s", whereClause)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count snippets: %w", err)
	}

	// Calculate offset
	offset := (filter.Page - 1) * filter.Limit

	// Build main query using safe column names from allowedSortColumns map
	query := fmt.Sprintf(`
		SELECT s.id, s.title, s.description, s.content, s.language, s.is_favorite, s.is_public,
		       s.view_count, s.s3_key, s.checksum, s.is_archived, s.created_at, s.updated_at, s.deleted_at
		FROM snippets s
		%s
		ORDER BY s.%s %s
		LIMIT ? OFFSET ?
	`, whereClause, sortColumn, sortOrder)

	args = append(args, filter.Limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list snippets: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			slog.Error("failed to close rows", "error", err)
		}
	}()

	var snippets []models.Snippet
	for rows.Next() {
		var s models.Snippet
		if err := rows.Scan(
			&s.ID,
			&s.Title,
			&s.Description,
			&s.Content,
			&s.Language,
			&s.IsFavorite,
			&s.IsPublic,
			&s.ViewCount,
			&s.S3Key,
			&s.Checksum,
			&s.IsArchived,
			&s.CreatedAt,
			&s.UpdatedAt,
			&s.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan snippet: %w", err)
		}
		snippets = append(snippets, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating snippets: %w", err)
	}

	// Calculate total pages
	totalPages := total / filter.Limit
	if total%filter.Limit > 0 {
		totalPages++
	}

	return &models.SnippetListResponse{
		Data: snippets,
		Pagination: models.Pagination{
			Page:       filter.Page,
			Limit:      filter.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// ToggleFavorite toggles the favorite status of a snippet
func (r *SnippetRepository) ToggleFavorite(ctx context.Context, id string) (*models.Snippet, error) {
	query := `
		UPDATE snippets
		SET is_favorite = NOT is_favorite
		WHERE id = ?
		RETURNING id, title, description, content, language, is_favorite, is_public,
		          view_count, s3_key, checksum, is_archived, created_at, updated_at, deleted_at
	`

	snippet := &models.Snippet{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&snippet.ID,
		&snippet.Title,
		&snippet.Description,
		&snippet.Content,
		&snippet.Language,
		&snippet.IsFavorite,
		&snippet.IsPublic,
		&snippet.ViewCount,
		&snippet.S3Key,
		&snippet.Checksum,
		&snippet.IsArchived,
		&snippet.CreatedAt,
		&snippet.UpdatedAt,
		&snippet.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to toggle favorite: %w", err)
	}

	return snippet, nil
}

// ToggleArchive toggles the archive status of a snippet
func (r *SnippetRepository) ToggleArchive(ctx context.Context, id string) (*models.Snippet, error) {
	query := `
		UPDATE snippets
		SET is_archived = NOT is_archived,
		    is_public = CASE WHEN (NOT is_archived) = 1 THEN 0 ELSE is_public END,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		RETURNING id, title, description, content, language, is_favorite, is_public,
		          view_count, s3_key, checksum, is_archived, created_at, updated_at, deleted_at
	`

	snippet := &models.Snippet{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&snippet.ID,
		&snippet.Title,
		&snippet.Description,
		&snippet.Content,
		&snippet.Language,
		&snippet.IsFavorite,
		&snippet.IsPublic,
		&snippet.ViewCount,
		&snippet.S3Key,
		&snippet.Checksum,
		&snippet.IsArchived,
		&snippet.CreatedAt,
		&snippet.UpdatedAt,
		&snippet.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to toggle archive: %w", err)
	}

	return snippet, nil
}

// IncrementViewCount increments the view count for a snippet
func (r *SnippetRepository) IncrementViewCount(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE snippets SET view_count = view_count + 1 WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to increment view count: %w", err)
	}
	return nil
}

// Search performs full-text search on snippets
func (r *SnippetRepository) Search(ctx context.Context, query string, limit int) ([]models.Snippet, error) {
	if limit <= 0 {
		limit = 10
	}

	sqlQuery := `
		SELECT s.id, s.title, s.description, s.content, s.language, s.is_favorite, s.is_public,
		       s.view_count, s.s3_key, s.checksum, s.is_archived, s.created_at, s.updated_at, s.deleted_at
		FROM snippets s
		WHERE s.rowid IN (
			SELECT rowid FROM snippets_fts WHERE snippets_fts MATCH ?
		)
        AND s.deleted_at IS NULL
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, sqlQuery, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search snippets: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			slog.Error("failed to close rows", "error", err)
		}
	}()

	var snippets []models.Snippet
	for rows.Next() {
		var s models.Snippet
		if err := rows.Scan(
			&s.ID,
			&s.Title,
			&s.Description,
			&s.Content,
			&s.Language,
			&s.IsFavorite,
			&s.IsPublic,
			&s.ViewCount,
			&s.S3Key,
			&s.Checksum,
			&s.IsArchived,
			&s.CreatedAt,
			&s.UpdatedAt,
			&s.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan snippet: %w", err)
		}
		snippets = append(snippets, s)
	}

	return snippets, rows.Err()
}
