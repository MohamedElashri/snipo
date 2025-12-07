package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/MohamedElashri/snipo/internal/models"
)

// TagRepository handles tag database operations
type TagRepository struct {
	db *sql.DB
}

// NewTagRepository creates a new tag repository
func NewTagRepository(db *sql.DB) *TagRepository {
	return &TagRepository{db: db}
}

// Create creates a new tag
func (r *TagRepository) Create(ctx context.Context, input *models.TagInput) (*models.Tag, error) {
	query := `
		INSERT INTO tags (name, color)
		VALUES (?, ?)
		RETURNING id, name, color, created_at
	`

	tag := &models.Tag{}
	err := r.db.QueryRowContext(ctx, query, input.Name, input.Color).Scan(
		&tag.ID,
		&tag.Name,
		&tag.Color,
		&tag.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}

	return tag, nil
}

// GetByID retrieves a tag by ID
func (r *TagRepository) GetByID(ctx context.Context, id int64) (*models.Tag, error) {
	query := `SELECT id, name, color, created_at FROM tags WHERE id = ?`

	tag := &models.Tag{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tag.ID,
		&tag.Name,
		&tag.Color,
		&tag.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	return tag, nil
}

// GetByName retrieves a tag by name
func (r *TagRepository) GetByName(ctx context.Context, name string) (*models.Tag, error) {
	query := `SELECT id, name, color, created_at FROM tags WHERE name = ?`

	tag := &models.Tag{}
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&tag.ID,
		&tag.Name,
		&tag.Color,
		&tag.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get tag by name: %w", err)
	}

	return tag, nil
}

// List retrieves all tags
func (r *TagRepository) List(ctx context.Context) ([]models.Tag, error) {
	query := `SELECT id, name, color, created_at FROM tags ORDER BY name ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.Color, &tag.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tags: %w", err)
	}

	return tags, nil
}

// Update updates an existing tag
func (r *TagRepository) Update(ctx context.Context, id int64, input *models.TagInput) (*models.Tag, error) {
	query := `
		UPDATE tags
		SET name = ?, color = ?
		WHERE id = ?
		RETURNING id, name, color, created_at
	`

	tag := &models.Tag{}
	err := r.db.QueryRowContext(ctx, query, input.Name, input.Color, id).Scan(
		&tag.ID,
		&tag.Name,
		&tag.Color,
		&tag.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to update tag: %w", err)
	}

	return tag, nil
}

// Delete deletes a tag
func (r *TagRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM tags WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// GetSnippetTags retrieves all tags for a snippet
func (r *TagRepository) GetSnippetTags(ctx context.Context, snippetID string) ([]models.Tag, error) {
	query := `
		SELECT t.id, t.name, t.color, t.created_at
		FROM tags t
		JOIN snippet_tags st ON t.id = st.tag_id
		WHERE st.snippet_id = ?
		ORDER BY t.name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, snippetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get snippet tags: %w", err)
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.Color, &tag.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// SetSnippetTags sets the tags for a snippet (replaces existing)
func (r *TagRepository) SetSnippetTags(ctx context.Context, snippetID string, tagNames []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Remove existing tags
	_, err = tx.ExecContext(ctx, `DELETE FROM snippet_tags WHERE snippet_id = ?`, snippetID)
	if err != nil {
		return fmt.Errorf("failed to remove existing tags: %w", err)
	}

	// Add new tags
	for _, name := range tagNames {
		// Get or create tag
		var tagID int64
		err := tx.QueryRowContext(ctx, `SELECT id FROM tags WHERE name = ?`, name).Scan(&tagID)
		if err == sql.ErrNoRows {
			// Create new tag with default color
			err = tx.QueryRowContext(ctx,
				`INSERT INTO tags (name, color) VALUES (?, '#6366f1') RETURNING id`,
				name,
			).Scan(&tagID)
			if err != nil {
				return fmt.Errorf("failed to create tag %s: %w", name, err)
			}
		} else if err != nil {
			return fmt.Errorf("failed to get tag %s: %w", name, err)
		}

		// Link tag to snippet
		_, err = tx.ExecContext(ctx,
			`INSERT OR IGNORE INTO snippet_tags (snippet_id, tag_id) VALUES (?, ?)`,
			snippetID, tagID,
		)
		if err != nil {
			return fmt.Errorf("failed to link tag %s to snippet: %w", name, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetTagSnippetCount returns the number of snippets for each tag
func (r *TagRepository) GetTagSnippetCount(ctx context.Context, tagID int64) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM snippet_tags WHERE tag_id = ?`,
		tagID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count snippets for tag: %w", err)
	}
	return count, nil
}
