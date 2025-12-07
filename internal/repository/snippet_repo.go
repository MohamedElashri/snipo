package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

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
		INSERT INTO snippets (title, description, content, language, is_public)
		VALUES (?, ?, ?, ?, ?)
		RETURNING id, title, description, content, language, is_favorite, is_public, 
		          view_count, s3_key, checksum, created_at, updated_at
	`

	snippet := &models.Snippet{}
	err := r.db.QueryRowContext(ctx, query,
		input.Title,
		input.Description,
		input.Content,
		input.Language,
		input.IsPublic,
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
		&snippet.CreatedAt,
		&snippet.UpdatedAt,
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
		       view_count, s3_key, checksum, created_at, updated_at
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
		&snippet.CreatedAt,
		&snippet.UpdatedAt,
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
		SET title = ?, description = ?, content = ?, language = ?, is_public = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		RETURNING id, title, description, content, language, is_favorite, is_public,
		          view_count, s3_key, checksum, created_at, updated_at
	`

	snippet := &models.Snippet{}
	err := r.db.QueryRowContext(ctx, query,
		input.Title,
		input.Description,
		input.Content,
		input.Language,
		input.IsPublic,
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
		&snippet.CreatedAt,
		&snippet.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update snippet: %w", err)
	}

	return snippet, nil
}

// Delete removes a snippet by ID
func (r *SnippetRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM snippets WHERE id = ?", id)
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

	return nil
}

// List retrieves snippets with filtering and pagination
func (r *SnippetRepository) List(ctx context.Context, filter models.SnippetFilter) (*models.SnippetListResponse, error) {
	// Build query
	var conditions []string
	var args []interface{}

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

	// Filter by tag
	if filter.TagID > 0 {
		conditions = append(conditions, "s.id IN (SELECT snippet_id FROM snippet_tags WHERE tag_id = ?)")
		args = append(args, filter.TagID)
	}

	// Filter by folder
	if filter.FolderID > 0 {
		conditions = append(conditions, "s.id IN (SELECT snippet_id FROM snippet_folders WHERE folder_id = ?)")
		args = append(args, filter.FolderID)
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

	// Validate sort column
	validSortColumns := map[string]bool{
		"created_at": true,
		"updated_at": true,
		"title":      true,
		"language":   true,
	}
	if !validSortColumns[filter.SortBy] {
		filter.SortBy = "updated_at"
	}

	sortOrder := "DESC"
	if strings.ToLower(filter.SortOrder) == "asc" {
		sortOrder = "ASC"
	}

	// Calculate offset
	offset := (filter.Page - 1) * filter.Limit

	// Build main query
	query := fmt.Sprintf(`
		SELECT s.id, s.title, s.description, s.content, s.language, s.is_favorite, s.is_public,
		       s.view_count, s.s3_key, s.checksum, s.created_at, s.updated_at
		FROM snippets s
		%s
		ORDER BY s.%s %s
		LIMIT ? OFFSET ?
	`, whereClause, filter.SortBy, sortOrder)

	args = append(args, filter.Limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list snippets: %w", err)
	}
	defer rows.Close()

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
			&s.CreatedAt,
			&s.UpdatedAt,
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
		          view_count, s3_key, checksum, created_at, updated_at
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
		&snippet.CreatedAt,
		&snippet.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to toggle favorite: %w", err)
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
		       s.view_count, s.s3_key, s.checksum, s.created_at, s.updated_at
		FROM snippets s
		WHERE s.rowid IN (
			SELECT rowid FROM snippets_fts WHERE snippets_fts MATCH ?
		)
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, sqlQuery, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search snippets: %w", err)
	}
	defer rows.Close()

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
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan snippet: %w", err)
		}
		snippets = append(snippets, s)
	}

	return snippets, rows.Err()
}
