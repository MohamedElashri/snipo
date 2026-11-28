package repository

import (
	"context"
	"database/sql/driver"
	"errors"
	"strings"

	"github.com/MohamedElashri/snipo/pkg/models"
)

var ErrNoteTooLarge = errors.New("appended content exceeds the 1 MiB file limit")

const appendableNote = `s.deleted_at IS NULL AND s.is_archived = 0 AND s.is_public = 0
	AND ` + unexpiredSnippet

// ListNoteTargets returns private Markdown files and Markdown snippets without a file list.
func (r *SnippetRepository) ListNoteTargets(ctx context.Context, query string, page, limit int, kinds ...string) (models.NoteTargets, error) {
	result := models.NoteTargets{Targets: make([]models.NoteTarget, 0, limit)}
	kind := ""
	if len(kinds) > 0 {
		kind = kinds[0]
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT s.id, s.title, COALESCE(f.id, 0), COALESCE(f.filename, ''), s.kind, COALESCE(f.language, s.language)
		FROM snippets s LEFT JOIN snippet_files f ON f.snippet_id = s.id
		WHERE `+appendableNote+`
			AND (? = '' OR s.kind = ?)
			AND (COALESCE(f.language, s.language) = 'markdown'
				OR (? = 'note' AND s.kind = 'note' AND COALESCE(f.language, s.language) = 'plaintext'))
			AND (instr(lower(s.title), lower(?)) > 0 OR instr(lower(f.filename), lower(?)) > 0 OR ? = '')
		ORDER BY s.updated_at DESC, s.id, f.sort_order, f.id LIMIT ? OFFSET ?`,
		kind, kind, kind, query, query, query, limit+1, (page-1)*limit)
	if err != nil {
		return result, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		if len(result.Targets) == limit {
			result.HasMore = true
			break
		}
		var target models.NoteTarget
		if err := rows.Scan(&target.SnippetID, &target.Title, &target.FileID, &target.Filename, &target.Kind, &target.Language); err != nil {
			return result, err
		}
		result.Targets = append(result.Targets, target)
	}
	return result, rows.Err()
}

// AppendNote updates one Markdown file and returns the previous state for history.
func (r *SnippetRepository) AppendNote(ctx context.Context, id string, input models.NoteAppendInput) (*models.Snippet, error) {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()

	// Acquire the write lock before taking a read snapshot. This lets SQLite
	// wait for a concurrent writer instead of failing to upgrade a stale read.
	if _, err := conn.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		return nil, err
	}
	committed := false
	defer func() {
		if !committed {
			if _, err := conn.ExecContext(context.Background(), "ROLLBACK"); err != nil {
				_ = conn.Raw(func(any) error { return driver.ErrBadConn })
			}
		}
	}()

	before, err := snippetByID(ctx, conn, id)
	if err != nil {
		return nil, err
	}
	if before == nil || before.DeletedAt != nil || before.IsArchived || before.IsPublic || before.IsExpired() {
		return nil, ErrNotFound
	}
	before.Files, err = filesBySnippetID(ctx, conn, id)
	if err != nil {
		return nil, err
	}
	primary := before.Content
	if len(before.Files) == 0 {
		if input.FileID != 0 || !acceptsCapture(before.Kind, before.Language, input.Language) {
			return nil, ErrNotFound
		}
		primary, err = joinNoteContent(primary, input.Content)
		if err != nil {
			return nil, err
		}
	} else {
		index := -1
		for i, file := range before.Files {
			if file.ID == input.FileID && acceptsCapture(before.Kind, file.Language, input.Language) {
				index = i
				break
			}
		}
		if index == -1 {
			return nil, ErrNotFound
		}
		content, err := joinNoteContent(before.Files[index].Content, input.Content)
		if err != nil {
			return nil, err
		}
		if _, err := conn.ExecContext(ctx, `
			UPDATE snippet_files SET content = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
			content, input.FileID); err != nil {
			return nil, err
		}
		if index == 0 {
			primary = content
		}
	}
	if _, err := conn.ExecContext(ctx, `
		UPDATE snippets SET content = ?, revision = revision + 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		primary, id); err != nil {
		return nil, err
	}
	if _, err := conn.ExecContext(ctx, "COMMIT"); err != nil {
		return nil, err
	}
	committed = true
	return before, nil
}

func acceptsCapture(kind, language, requested string) bool {
	if requested != "" && requested != language {
		return false
	}
	return language == "markdown" || (kind == models.KindNote && requested == "plaintext" && language == requested)
}

func joinNoteContent(existing, added string) (string, error) {
	separator := ""
	if existing != "" {
		// Retain existing whitespace, including CRLF lines and Markdown hard breaks.
		switch {
		case strings.HasSuffix(existing, "\r\n\r\n") || strings.HasSuffix(existing, "\n\n") || strings.HasSuffix(existing, "\n\r\n"):
		case strings.HasSuffix(existing, "\r\n"):
			separator = "\r\n"
		case strings.HasSuffix(existing, "\n"):
			separator = "\n"
		default:
			separator = "\n\n"
		}
	}
	if len(existing)+len(separator)+len(added) > models.MaxNoteContentBytes {
		return "", ErrNoteTooLarge
	}
	return existing + separator + added, nil
}
