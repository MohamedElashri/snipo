package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/MohamedElashri/snipo/pkg/models"
)

// SnippetFileRepository handles snippet file database operations
type SnippetFileRepository struct {
	db *sql.DB
}

// NewSnippetFileRepository creates a new snippet file repository
func NewSnippetFileRepository(db *sql.DB) *SnippetFileRepository {
	return &SnippetFileRepository{db: db}
}

// GetBySnippetID retrieves all files for a snippet
func (r *SnippetFileRepository) GetBySnippetID(ctx context.Context, snippetID string) ([]models.SnippetFile, error) {
	query := `
		SELECT id, snippet_id, filename, content, language, sort_order, created_at, updated_at
		FROM snippet_files
		WHERE snippet_id = ?
		ORDER BY sort_order, id
	`

	rows, err := r.db.QueryContext(ctx, query, snippetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get snippet files: %w", err)
	}
	defer func() {
		_ = rows.Close() // Best effort close
	}()

	var files []models.SnippetFile
	for rows.Next() {
		var f models.SnippetFile
		if err := rows.Scan(
			&f.ID,
			&f.SnippetID,
			&f.Filename,
			&f.Content,
			&f.Language,
			&f.SortOrder,
			&f.CreatedAt,
			&f.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan snippet file: %w", err)
		}
		files = append(files, f)
	}

	return files, nil
}

// Create creates a new snippet file
func (r *SnippetFileRepository) Create(ctx context.Context, snippetID string, file *models.SnippetFileInput, sortOrder int) (*models.SnippetFile, error) {
	query := `
		INSERT INTO snippet_files (snippet_id, filename, content, language, sort_order)
		VALUES (?, ?, ?, ?, ?)
		RETURNING id, snippet_id, filename, content, language, sort_order, created_at, updated_at
	`

	var f models.SnippetFile
	err := r.db.QueryRowContext(ctx, query,
		snippetID,
		file.Filename,
		file.Content,
		file.Language,
		sortOrder,
	).Scan(
		&f.ID,
		&f.SnippetID,
		&f.Filename,
		&f.Content,
		&f.Language,
		&f.SortOrder,
		&f.CreatedAt,
		&f.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create snippet file: %w", err)
	}

	return &f, nil
}

// Update updates an existing snippet file
func (r *SnippetFileRepository) Update(ctx context.Context, file *models.SnippetFileInput, sortOrder int) (*models.SnippetFile, error) {
	query := `
		UPDATE snippet_files
		SET filename = ?, content = ?, language = ?, sort_order = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		RETURNING id, snippet_id, filename, content, language, sort_order, created_at, updated_at
	`

	var f models.SnippetFile
	err := r.db.QueryRowContext(ctx, query,
		file.Filename,
		file.Content,
		file.Language,
		sortOrder,
		file.ID,
	).Scan(
		&f.ID,
		&f.SnippetID,
		&f.Filename,
		&f.Content,
		&f.Language,
		&f.SortOrder,
		&f.CreatedAt,
		&f.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update snippet file: %w", err)
	}

	return &f, nil
}

// Delete deletes a snippet file
func (r *SnippetFileRepository) Delete(ctx context.Context, fileID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM snippet_files WHERE id = ?", fileID)
	if err != nil {
		return fmt.Errorf("failed to delete snippet file: %w", err)
	}
	return nil
}

// SyncFiles synchronizes files for a snippet (creates, updates, deletes as needed)
func (r *SnippetFileRepository) SyncFiles(ctx context.Context, snippetID string, files []models.SnippetFileInput) ([]models.SnippetFile, error) {
	// Get existing files
	existing, err := r.GetBySnippetID(ctx, snippetID)
	if err != nil {
		return nil, err
	}

	// Build map of existing file IDs
	existingMap := make(map[int64]bool)
	for _, f := range existing {
		existingMap[f.ID] = true
	}

	// Track which IDs are in the input
	inputIDs := make(map[int64]bool)
	var result []models.SnippetFile

	for i, file := range files {
		if file.ID > 0 {
			// Update existing file
			inputIDs[file.ID] = true
			updated, err := r.Update(ctx, &file, i)
			if err != nil {
				return nil, err
			}
			result = append(result, *updated)
		} else {
			// Create new file
			created, err := r.Create(ctx, snippetID, &file, i)
			if err != nil {
				return nil, err
			}
			result = append(result, *created)
		}
	}

	// Delete files that are no longer in the input
	for id := range existingMap {
		if !inputIDs[id] {
			if err := r.Delete(ctx, id); err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}
