package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"github.com/MohamedElashri/snipo/pkg/models"
)

var ErrTemplateConflict = errors.New("template changed since it was opened")

type TemplateRepository struct {
	db *sql.DB
}

func NewTemplateRepository(db *sql.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

const selectTemplate = `SELECT t.id, r.revision, t.state, t.favorite, t.updated_at, t.last_used_at, r.definition
	FROM templates t JOIN template_revisions r ON r.template_id = t.id`

func scanTemplate(row interface{ Scan(...any) error }) (*models.Template, error) {
	var t models.Template
	var definition string
	err := row.Scan(&t.ID, &t.Revision, &t.State, &t.Favorite, &t.UpdatedAt, &t.LastUsedAt, &definition)
	if err != nil {
		return nil, err
	}
	return &t, json.Unmarshal([]byte(definition), &t.Definition)
}

func (r *TemplateRepository) Get(ctx context.Context, id string, revision int) (*models.Template, error) {
	return scanTemplate(r.db.QueryRowContext(ctx, selectTemplate+` WHERE t.id = ? AND r.revision = COALESCE(NULLIF(?, 0), t.revision)`, id, revision))
}

func (r *TemplateRepository) List(ctx context.Context, state, query string, limit, offset int) ([]models.TemplateSummary, int, error) {
	filter := ` FROM templates t JOIN template_revisions r ON r.template_id = t.id AND r.revision = t.revision
		WHERE t.state = ? AND (
			? = '' OR instr(lower(json_extract(r.definition, '$.name')), lower(?)) > 0
			OR instr(lower(json_extract(r.definition, '$.purpose')), lower(?)) > 0
			OR EXISTS (
				SELECT 1 FROM json_each(r.definition, '$.files') AS f
				WHERE instr(lower(json_extract(f.value, '$.filename')), lower(?)) > 0
			)
		)`
	query = strings.TrimSpace(query)
	args := []any{state, query, query, query, query}
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT count(*)"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT t.id, t.revision, t.state, t.favorite, t.last_used_at,
		json_extract(r.definition, '$.name'), json_extract(r.definition, '$.purpose'),
		(SELECT json_group_array(json_extract(value, '$.filename')) FROM json_each(r.definition, '$.files'))`+
		filter+` ORDER BY t.favorite DESC, t.last_used_at DESC, lower(json_extract(r.definition, '$.name')), t.id LIMIT ? OFFSET ?`,
		append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	items := []models.TemplateSummary{}
	for rows.Next() {
		var item models.TemplateSummary
		var filenames string
		if err := rows.Scan(&item.ID, &item.Revision, &item.State, &item.Favorite, &item.LastUsedAt, &item.Name, &item.Purpose, &filenames); err != nil {
			return nil, 0, err
		}
		if err := json.Unmarshal([]byte(filenames), &item.Filenames); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *TemplateRepository) Save(ctx context.Context, id string, expected int, def models.TemplateDefinition) (*models.Template, error) {
	body, err := json.Marshal(def)
	if err != nil {
		return nil, err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	if expected == 0 {
		if id == "" {
			err = tx.QueryRowContext(ctx, "INSERT INTO templates DEFAULT VALUES RETURNING id").Scan(&id)
		} else {
			err = tx.QueryRowContext(ctx, `INSERT INTO templates (id) VALUES (?) ON CONFLICT(id) DO NOTHING RETURNING id`, id).Scan(&id)
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrTemplateConflict
			}
		}
	} else {
		var current int
		err = tx.QueryRowContext(ctx, `UPDATE templates SET revision = revision + 1, updated_at = CURRENT_TIMESTAMP
			WHERE id = ? AND revision = ? AND state != 'trash' RETURNING revision`, id, expected).Scan(&current)
		if errors.Is(err, sql.ErrNoRows) {
			var actual int
			err = tx.QueryRowContext(ctx, "SELECT revision FROM templates WHERE id = ? AND state != 'trash'", id).Scan(&actual)
			if err == nil {
				return nil, ErrTemplateConflict
			}
		}
	}
	if err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO template_revisions (template_id, revision, definition)
		VALUES (?, ?, ?)`, id, expected+1, string(body)); err != nil {
		return nil, err
	}
	result, err := scanTemplate(tx.QueryRowContext(ctx, selectTemplate+` WHERE t.id = ? AND r.revision = t.revision`, id))
	if err != nil {
		return nil, err
	}
	return result, tx.Commit()
}

func (r *TemplateRepository) SetStatus(ctx context.Context, id string, state *string, favorite *bool) error {
	result, err := r.db.ExecContext(ctx, `UPDATE templates SET state = COALESCE(?, state), favorite = COALESCE(?, favorite)
		WHERE id = ?`, state, favorite, id)
	return templateAffected(result, err)
}

func (r *TemplateRepository) Purge(ctx context.Context, id string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	result, err := tx.ExecContext(ctx, "DELETE FROM templates WHERE id = ? AND state = 'trash'", id)
	if err := templateAffected(result, err); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM template_revisions WHERE template_id = ?", id); err != nil {
		return err
	}
	return tx.Commit()
}

func templateAffected(result sql.Result, err error) error {
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err == nil && count == 0 {
		return sql.ErrNoRows
	}
	return err
}

func (r *TemplateRepository) History(ctx context.Context, id string) ([]models.TemplateRevision, error) {
	return templateHistory(ctx, r.db,
		"SELECT revision, created_at, '' FROM template_revisions WHERE template_id = ? ORDER BY revision DESC", id)
}

type templateQueryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func templateHistory(ctx context.Context, db templateQueryer, query, id string) ([]models.TemplateRevision, error) {
	rows, err := db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := []models.TemplateRevision{}
	for rows.Next() {
		var item models.TemplateRevision
		var def string
		if err := rows.Scan(&item.Revision, &item.CreatedAt, &def); err != nil {
			return nil, err
		}
		if def != "" {
			if err := json.Unmarshal([]byte(def), &item.Definition); err != nil {
				return nil, err
			}
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *TemplateRepository) Snapshots(ctx context.Context) ([]models.TemplateSnapshot, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	rows, err := tx.QueryContext(ctx, selectTemplate+` WHERE r.revision = t.revision ORDER BY t.id`)
	if err != nil {
		return nil, err
	}
	items := []models.TemplateSnapshot{}
	for rows.Next() {
		item, err := scanTemplate(rows)
		if err != nil {
			_ = rows.Close()
			return nil, err
		}
		items = append(items, models.TemplateSnapshot{Template: *item})
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	for i := range items {
		items[i].Revisions, err = templateHistory(ctx, tx,
			"SELECT revision, created_at, definition FROM template_revisions WHERE template_id = ? ORDER BY revision DESC", items[i].ID)
		if err != nil {
			return nil, err
		}
	}
	return items, tx.Commit()
}

func (r *TemplateRepository) RestoreSnapshots(ctx context.Context, items []models.TemplateSnapshot, replace bool) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	if replace {
		if _, err := tx.ExecContext(ctx, "DELETE FROM template_revisions"); err != nil {
			return 0, err
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM templates"); err != nil {
			return 0, err
		}
	}
	count := 0
	for _, item := range items {
		res, err := tx.ExecContext(ctx, `INSERT INTO templates (id, revision, state, favorite, updated_at, last_used_at)
			VALUES (?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO NOTHING`,
			item.ID, item.Revision, item.State, item.Favorite, item.UpdatedAt, item.LastUsedAt)
		if err != nil {
			return 0, err
		}
		n, err := res.RowsAffected()
		if err != nil {
			return 0, err
		}
		if n == 0 {
			continue
		}
		for _, rev := range item.Revisions {
			body, err := json.Marshal(rev.Definition)
			if err != nil {
				return 0, err
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO template_revisions (template_id, revision, definition, created_at)
				VALUES (?, ?, ?, ?)`, item.ID, rev.Revision, string(body), rev.CreatedAt); err != nil {
				return 0, err
			}
		}
		count++
	}
	return count, tx.Commit()
}

func (r *TemplateRepository) RemapFolders(ctx context.Context, destinations map[int64]int64) error {
	mapping, err := json.Marshal(destinations)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `
		UPDATE template_revisions
		SET definition = CASE folders.value
			WHEN 0 THEN json_remove(definition, '$.folder_id')
			ELSE json_set(definition, '$.folder_id', folders.value)
		END
		FROM json_each(?) AS folders
		WHERE json_extract(definition, '$.folder_id') = CAST(folders.key AS INTEGER)`, string(mapping))
	return err
}

type rowQuerier interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func (r *SnippetRepository) CreateFromTemplate(ctx context.Context, input *models.SnippetInput, trackUse bool) (*models.Snippet, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	snippet, err := createSnippet(ctx, tx, input)
	if err != nil {
		return nil, err
	}
	for i, file := range input.Files {
		created, err := createSnippetFile(ctx, tx, snippet.ID, &file, i)
		if err != nil {
			return nil, err
		}
		snippet.Files = append(snippet.Files, *created)
	}
	if err := setSnippetTag(ctx, tx, snippet.ID, input.Tags); err != nil {
		return nil, err
	}
	if input.FolderID != nil {
		var exists bool
		if err := tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM folders WHERE id = ?)", *input.FolderID).Scan(&exists); err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrNotFound
		}
		if err := setSnippetFolder(ctx, tx, snippet.ID, input.FolderID); err != nil {
			return nil, err
		}
	}
	origin := input.TemplateOrigin
	if _, err := tx.ExecContext(ctx, `INSERT INTO snippet_template_origins (snippet_id, template_id, revision, name)
		VALUES (?, ?, ?, ?)`, snippet.ID, origin.ID, origin.Revision, origin.Name); err != nil {
		return nil, err
	}
	if trackUse {
		if _, err := tx.ExecContext(ctx, "UPDATE templates SET last_used_at = CURRENT_TIMESTAMP WHERE id = ?", origin.ID); err != nil {
			return nil, err
		}
	}
	snippet.TemplateOrigin = origin
	return snippet, tx.Commit()
}

func (r *SnippetRepository) GetTemplateOrigin(ctx context.Context, snippetID string) (*models.TemplateOrigin, error) {
	var origin models.TemplateOrigin
	err := r.db.QueryRowContext(ctx, "SELECT template_id, revision, name FROM snippet_template_origins WHERE snippet_id = ?", snippetID).
		Scan(&origin.ID, &origin.Revision, &origin.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &origin, err
}

func (r *TemplateRepository) SnippetOrigins(ctx context.Context) (map[string]models.TemplateOrigin, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT snippet_id, template_id, revision, name FROM snippet_template_origins")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	result := make(map[string]models.TemplateOrigin)
	for rows.Next() {
		var snippetID string
		var origin models.TemplateOrigin
		if err := rows.Scan(&snippetID, &origin.ID, &origin.Revision, &origin.Name); err != nil {
			return nil, err
		}
		result[snippetID] = origin
	}
	return result, rows.Err()
}
