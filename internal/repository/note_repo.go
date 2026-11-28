package repository

import (
	"context"
	"errors"

	"github.com/MohamedElashri/snipo/pkg/models"
)

var (
	ErrNoteConflict   = errors.New("the note changed since it was opened")
	ErrNoteOverview   = errors.New("remove or replace the folder overview before changing this note's folder or format")
	ErrNoteConversion = errors.New("only a single-body Markdown or plain text snippet can be converted to a note")
)

// ConvertKind preserves existing file IDs and uses the actual file body when a
// one-file snippet has different legacy content.
func (r *SnippetRepository) ConvertKind(ctx context.Context, id, kind string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE snippets SET kind = ?, revision = revision + 1, updated_at = CURRENT_TIMESTAMP,
			language = CASE WHEN ? = 'note' THEN COALESCE((SELECT language FROM snippet_files WHERE snippet_id = snippets.id), language) ELSE language END,
			content = CASE WHEN ? = 'note' THEN COALESCE((SELECT content FROM snippet_files WHERE snippet_id = snippets.id), content) ELSE content END
		WHERE id = ? AND deleted_at IS NULL AND kind <> ?
		AND (? = 'snippet' OR (
			(SELECT COUNT(*) FROM snippet_files WHERE snippet_id = snippets.id) <= 1
			AND COALESCE((SELECT language FROM snippet_files WHERE snippet_id = snippets.id), language) IN ('markdown', 'plaintext')
		))`, kind, kind, kind, id, kind, kind)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil || count > 0 {
		return err
	}
	item, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if item == nil || item.DeletedAt != nil {
		return ErrNotFound
	}
	if item.Kind == kind {
		return nil
	}
	return ErrNoteConversion
}

// UpdateNote writes the canonical body, any existing file representation and
// folder membership in one transaction.
func (r *SnippetRepository) UpdateNote(ctx context.Context, id string, input *models.SnippetInput, setFolder bool) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	result, err := tx.ExecContext(ctx, `
		UPDATE snippets SET title = ?, description = ?, content = ?, language = ?,
			is_public = ?, is_archived = ?, expires_at = ?, revision = revision + 1,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND deleted_at IS NULL AND kind = 'note' AND revision = ?
		AND (? OR NOT EXISTS (SELECT 1 FROM folders WHERE overview_snippet_id = snippets.id
			AND (? <> 'markdown' OR (? AND id IS NOT ?))))`,
		input.Title, input.Description, input.Content, input.Language, input.IsPublic,
		input.IsArchived, input.ExpiresAt, id, input.Revision, input.ClearOverview, input.Language, setFolder, input.FolderID)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		item, err := snippetByID(ctx, tx, id)
		if err != nil {
			return err
		}
		if item == nil || item.DeletedAt != nil {
			return ErrNotFound
		}
		if item.Kind != models.KindNote || input.Revision == nil || item.Revision != *input.Revision {
			return ErrNoteConflict
		}
		return ErrNoteOverview
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE snippet_files SET content = ?, language = ?, updated_at = CURRENT_TIMESTAMP WHERE snippet_id = ?`,
		input.Content, input.Language, id); err != nil {
		return err
	}
	if input.ClearOverview {
		if _, err := tx.ExecContext(ctx, `
			UPDATE folders SET overview_snippet_id = NULL WHERE overview_snippet_id = ?
				AND (? <> 'markdown' OR (? AND id IS NOT ?))`, id, input.Language, setFolder, input.FolderID); err != nil {
			return err
		}
	}
	if setFolder {
		if err := setSnippetFolder(ctx, tx, id, input.FolderID); err != nil {
			return err
		}
	}
	return tx.Commit()
}
