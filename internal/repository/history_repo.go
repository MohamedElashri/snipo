package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/MohamedElashri/snipo/internal/models"
)

// HistoryRepository handles snippet history database operations
type HistoryRepository struct {
	db *sql.DB
}

// NewHistoryRepository creates a new history repository
func NewHistoryRepository(db *sql.DB) *HistoryRepository {
	return &HistoryRepository{db: db}
}

// CreateHistory creates a new history entry for a snippet
func (r *HistoryRepository) CreateHistory(ctx context.Context, snippet *models.Snippet, changeType string) (int64, error) {
	query := `
		INSERT INTO snippet_history 
		(snippet_id, title, description, content, language, is_favorite, is_public, is_archived, change_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		snippet.ID,
		snippet.Title,
		snippet.Description,
		snippet.Content,
		snippet.Language,
		snippet.IsFavorite,
		snippet.IsPublic,
		snippet.IsArchived,
		changeType,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to create snippet history: %w", err)
	}

	historyID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get history ID: %w", err)
	}

	return historyID, nil
}

// CreateFileHistory creates history entries for snippet files
func (r *HistoryRepository) CreateFileHistory(ctx context.Context, historyID int64, files []models.SnippetFile) error {
	if len(files) == 0 {
		return nil
	}

	query := `
		INSERT INTO snippet_files_history 
		(history_id, snippet_id, filename, content, language, sort_order)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare file history statement: %w", err)
	}
	defer stmt.Close()

	for _, file := range files {
		_, err := stmt.ExecContext(ctx,
			historyID,
			file.SnippetID,
			file.Filename,
			file.Content,
			file.Language,
			file.SortOrder,
		)
		if err != nil {
			return fmt.Errorf("failed to create file history: %w", err)
		}
	}

	return nil
}

// GetSnippetHistory retrieves all history entries for a snippet
func (r *HistoryRepository) GetSnippetHistory(ctx context.Context, snippetID string, limit int) ([]models.SnippetHistory, error) {
	if limit <= 0 {
		limit = 50 // Default limit
	}

	query := `
		SELECT id, snippet_id, title, description, content, language, 
		       is_favorite, is_public, is_archived, change_type, created_at
		FROM snippet_history
		WHERE snippet_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, snippetID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get snippet history: %w", err)
	}
	defer rows.Close()

	var history []models.SnippetHistory
	for rows.Next() {
		var h models.SnippetHistory
		err := rows.Scan(
			&h.ID,
			&h.SnippetID,
			&h.Title,
			&h.Description,
			&h.Content,
			&h.Language,
			&h.IsFavorite,
			&h.IsPublic,
			&h.IsArchived,
			&h.ChangeType,
			&h.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan history: %w", err)
		}
		history = append(history, h)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating history rows: %w", err)
	}

	// Fetch files for each history entry
	for i := range history {
		files, err := r.GetHistoryFiles(ctx, history[i].ID)
		if err != nil {
			// Log error but continue
			continue
		}
		history[i].Files = files
	}

	return history, nil
}

// GetHistoryByID retrieves a specific history entry by ID
func (r *HistoryRepository) GetHistoryByID(ctx context.Context, historyID int64) (*models.SnippetHistory, error) {
	query := `
		SELECT id, snippet_id, title, description, content, language,
		       is_favorite, is_public, is_archived, change_type, created_at
		FROM snippet_history
		WHERE id = ?
	`

	var h models.SnippetHistory
	err := r.db.QueryRowContext(ctx, query, historyID).Scan(
		&h.ID,
		&h.SnippetID,
		&h.Title,
		&h.Description,
		&h.Content,
		&h.Language,
		&h.IsFavorite,
		&h.IsPublic,
		&h.IsArchived,
		&h.ChangeType,
		&h.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get history entry: %w", err)
	}

	// Fetch files
	files, err := r.GetHistoryFiles(ctx, h.ID)
	if err == nil {
		h.Files = files
	}

	return &h, nil
}

// GetHistoryFiles retrieves files for a specific history entry
func (r *HistoryRepository) GetHistoryFiles(ctx context.Context, historyID int64) ([]models.SnippetFileHistory, error) {
	query := `
		SELECT id, history_id, snippet_id, filename, content, language, sort_order, created_at
		FROM snippet_files_history
		WHERE history_id = ?
		ORDER BY sort_order ASC
	`

	rows, err := r.db.QueryContext(ctx, query, historyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get history files: %w", err)
	}
	defer rows.Close()

	var files []models.SnippetFileHistory
	for rows.Next() {
		var f models.SnippetFileHistory
		err := rows.Scan(
			&f.ID,
			&f.HistoryID,
			&f.SnippetID,
			&f.Filename,
			&f.Content,
			&f.Language,
			&f.SortOrder,
			&f.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan history file: %w", err)
		}
		files = append(files, f)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating history file rows: %w", err)
	}

	return files, nil
}

// DeleteSnippetHistory deletes all history entries for a snippet
func (r *HistoryRepository) DeleteSnippetHistory(ctx context.Context, snippetID string) error {
	query := `DELETE FROM snippet_history WHERE snippet_id = ?`

	_, err := r.db.ExecContext(ctx, query, snippetID)
	if err != nil {
		return fmt.Errorf("failed to delete snippet history: %w", err)
	}

	return nil
}

// DeleteOldHistory deletes history entries older than a specific date
func (r *HistoryRepository) DeleteOldHistory(ctx context.Context, daysToKeep int) (int64, error) {
	query := `
		DELETE FROM snippet_history 
		WHERE created_at < datetime('now', '-' || ? || ' days')
	`

	result, err := r.db.ExecContext(ctx, query, daysToKeep)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old history: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get affected rows: %w", err)
	}

	return affected, nil
}

// GetHistoryCount returns the total number of history entries for a snippet
func (r *HistoryRepository) GetHistoryCount(ctx context.Context, snippetID string) (int, error) {
	query := `SELECT COUNT(*) FROM snippet_history WHERE snippet_id = ?`

	var count int
	err := r.db.QueryRowContext(ctx, query, snippetID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get history count: %w", err)
	}

	return count, nil
}
