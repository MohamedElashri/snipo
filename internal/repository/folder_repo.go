package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/MohamedElashri/snipo/internal/models"
)

// FolderRepository handles folder database operations
type FolderRepository struct {
	db *sql.DB
}

// NewFolderRepository creates a new folder repository
func NewFolderRepository(db *sql.DB) *FolderRepository {
	return &FolderRepository{db: db}
}

// Create creates a new folder
func (r *FolderRepository) Create(ctx context.Context, input *models.FolderInput) (*models.Folder, error) {
	icon := input.Icon
	if icon == "" {
		icon = "folder"
	}

	query := `
		INSERT INTO folders (name, parent_id, icon, sort_order)
		VALUES (?, ?, ?, ?)
		RETURNING id, name, parent_id, icon, sort_order, created_at
	`

	folder := &models.Folder{}
	err := r.db.QueryRowContext(ctx, query, input.Name, input.ParentID, icon, input.SortOrder).Scan(
		&folder.ID,
		&folder.Name,
		&folder.ParentID,
		&folder.Icon,
		&folder.SortOrder,
		&folder.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create folder: %w", err)
	}

	return folder, nil
}

// GetByID retrieves a folder by ID
func (r *FolderRepository) GetByID(ctx context.Context, id int64) (*models.Folder, error) {
	query := `SELECT id, name, parent_id, icon, sort_order, created_at FROM folders WHERE id = ?`

	folder := &models.Folder{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&folder.ID,
		&folder.Name,
		&folder.ParentID,
		&folder.Icon,
		&folder.SortOrder,
		&folder.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get folder: %w", err)
	}

	return folder, nil
}

// List retrieves all folders (flat list) with snippet counts
func (r *FolderRepository) List(ctx context.Context) ([]models.Folder, error) {
	query := `
		SELECT f.id, f.name, f.parent_id, f.icon, f.sort_order, f.created_at,
		       (SELECT COUNT(*) FROM snippet_folders sf 
		        INNER JOIN snippets s ON s.id = sf.snippet_id 
		        WHERE sf.folder_id = f.id AND s.is_archived = 0) as snippet_count
		FROM folders f
		ORDER BY f.sort_order ASC, f.name ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list folders: %w", err)
	}
	defer rows.Close()

	var folders []models.Folder
	for rows.Next() {
		var folder models.Folder
		if err := rows.Scan(
			&folder.ID,
			&folder.Name,
			&folder.ParentID,
			&folder.Icon,
			&folder.SortOrder,
			&folder.CreatedAt,
			&folder.SnippetCount,
		); err != nil {
			return nil, fmt.Errorf("failed to scan folder: %w", err)
		}
		folders = append(folders, folder)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating folders: %w", err)
	}

	return folders, nil
}

// ListTree retrieves folders as a tree structure
func (r *FolderRepository) ListTree(ctx context.Context) ([]models.Folder, error) {
	folders, err := r.List(ctx)
	if err != nil {
		return nil, err
	}

	// Build tree from flat list
	return buildFolderTree(folders), nil
}

// buildFolderTree converts a flat list of folders into a tree structure
func buildFolderTree(folders []models.Folder) []models.Folder {
	// Create a map for quick lookup
	folderMap := make(map[int64]*models.Folder)
	for i := range folders {
		folders[i].Children = []models.Folder{}
		folderMap[folders[i].ID] = &folders[i]
	}

	// Build tree
	var roots []models.Folder
	for i := range folders {
		if folders[i].ParentID == nil {
			roots = append(roots, folders[i])
		} else {
			parent, exists := folderMap[*folders[i].ParentID]
			if exists {
				parent.Children = append(parent.Children, folders[i])
			} else {
				// Orphan folder, add to roots
				roots = append(roots, folders[i])
			}
		}
	}

	// Update roots with children from map
	for i := range roots {
		if mapped, exists := folderMap[roots[i].ID]; exists {
			roots[i].Children = mapped.Children
		}
	}

	return roots
}

// Update updates an existing folder
func (r *FolderRepository) Update(ctx context.Context, id int64, input *models.FolderInput) (*models.Folder, error) {
	icon := input.Icon
	if icon == "" {
		icon = "folder"
	}

	query := `
		UPDATE folders
		SET name = ?, parent_id = ?, icon = ?, sort_order = ?
		WHERE id = ?
		RETURNING id, name, parent_id, icon, sort_order, created_at
	`

	folder := &models.Folder{}
	err := r.db.QueryRowContext(ctx, query, input.Name, input.ParentID, icon, input.SortOrder, id).Scan(
		&folder.ID,
		&folder.Name,
		&folder.ParentID,
		&folder.Icon,
		&folder.SortOrder,
		&folder.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to update folder: %w", err)
	}

	return folder, nil
}

// Delete deletes a folder
func (r *FolderRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM folders WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete folder: %w", err)
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

// Move moves a folder to a new parent
func (r *FolderRepository) Move(ctx context.Context, id int64, newParentID *int64) (*models.Folder, error) {
	// Check for circular reference
	if newParentID != nil {
		if err := r.checkCircularReference(ctx, id, *newParentID); err != nil {
			return nil, err
		}
	}

	query := `
		UPDATE folders
		SET parent_id = ?
		WHERE id = ?
		RETURNING id, name, parent_id, icon, sort_order, created_at
	`

	folder := &models.Folder{}
	err := r.db.QueryRowContext(ctx, query, newParentID, id).Scan(
		&folder.ID,
		&folder.Name,
		&folder.ParentID,
		&folder.Icon,
		&folder.SortOrder,
		&folder.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to move folder: %w", err)
	}

	return folder, nil
}

// checkCircularReference checks if moving a folder would create a circular reference
func (r *FolderRepository) checkCircularReference(ctx context.Context, folderID, newParentID int64) error {
	// Check if newParentID is a descendant of folderID
	currentID := newParentID
	for {
		var parentID *int64
		err := r.db.QueryRowContext(ctx, `SELECT parent_id FROM folders WHERE id = ?`, currentID).Scan(&parentID)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil // Parent doesn't exist, no circular reference
			}
			return fmt.Errorf("failed to check circular reference: %w", err)
		}

		if parentID == nil {
			return nil // Reached root, no circular reference
		}

		if *parentID == folderID {
			return fmt.Errorf("cannot move folder: would create circular reference")
		}

		currentID = *parentID
	}
}

// GetFolderSnippetCount returns the number of snippets in a folder
func (r *FolderRepository) GetFolderSnippetCount(ctx context.Context, folderID int64) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM snippet_folders sf
		 JOIN snippets s ON s.id = sf.snippet_id
		 WHERE sf.folder_id = ? AND s.is_archived = 0`,
		folderID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count snippets in folder: %w", err)
	}
	return count, nil
}

// GetSnippetFolders retrieves all folders for a snippet
func (r *FolderRepository) GetSnippetFolders(ctx context.Context, snippetID string) ([]models.Folder, error) {
	query := `
		SELECT f.id, f.name, f.parent_id, f.icon, f.sort_order, f.created_at
		FROM folders f
		JOIN snippet_folders sf ON f.id = sf.folder_id
		WHERE sf.snippet_id = ?
		ORDER BY f.name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, snippetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get snippet folders: %w", err)
	}
	defer rows.Close()

	var folders []models.Folder
	for rows.Next() {
		var folder models.Folder
		if err := rows.Scan(
			&folder.ID,
			&folder.Name,
			&folder.ParentID,
			&folder.Icon,
			&folder.SortOrder,
			&folder.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan folder: %w", err)
		}
		folders = append(folders, folder)
	}

	return folders, nil
}

// SetSnippetFolder sets the folder for a snippet
func (r *FolderRepository) SetSnippetFolder(ctx context.Context, snippetID string, folderID *int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Remove existing folder associations
	_, err = tx.ExecContext(ctx, `DELETE FROM snippet_folders WHERE snippet_id = ?`, snippetID)
	if err != nil {
		return fmt.Errorf("failed to remove existing folders: %w", err)
	}

	// Add new folder association if provided
	if folderID != nil {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO snippet_folders (snippet_id, folder_id) VALUES (?, ?)`,
			snippetID, *folderID,
		)
		if err != nil {
			return fmt.Errorf("failed to set snippet folder: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
