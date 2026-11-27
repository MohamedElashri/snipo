package repository

import (
	"context"
	"database/sql"
	"errors"
)

var ErrInvalidOverview = errors.New("overview must be an active Markdown snippet in this folder")

// All reads and assignments use the same eligibility rules. An unavailable
// overview is hidden, and reappears if the note is restored to this folder.
const overviewCandidates = `
	SELECT s.id FROM snippets s
	JOIN snippet_folders sf ON sf.snippet_id = s.id
	WHERE sf.folder_id = f.id AND s.deleted_at IS NULL AND s.is_archived = 0
		AND ` + unexpiredSnippet + `
		AND (
			EXISTS (SELECT 1 FROM snippet_files sf2 WHERE sf2.snippet_id = s.id AND sf2.language = 'markdown')
			OR (s.language = 'markdown' AND NOT EXISTS (SELECT 1 FROM snippet_files sf2 WHERE sf2.snippet_id = s.id))
		)`

const folderOverviewQuery = overviewCandidates + ` AND s.id = f.overview_snippet_id`

func (r *FolderRepository) SetOverview(ctx context.Context, folderID int64, snippetID *string) error {
	var id int64
	err := r.db.QueryRowContext(ctx, `
		UPDATE folders AS f SET overview_snippet_id = ?
		WHERE f.id = ? AND (? IS NULL OR EXISTS (`+overviewCandidates+` AND s.id = ?))
		RETURNING id`, snippetID, folderID, snippetID, snippetID).Scan(&id)
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if _, err := r.GetByID(ctx, folderID); err != nil {
		return err
	}
	return ErrInvalidOverview
}
