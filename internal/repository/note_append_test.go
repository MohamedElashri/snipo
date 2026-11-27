package repository

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MohamedElashri/snipo/internal/database"
	"github.com/MohamedElashri/snipo/internal/testutil"
	"github.com/MohamedElashri/snipo/pkg/models"
)

func TestAppendNoteConcurrentAndFileIsolation(t *testing.T) {
	ctx := t.Context()
	db, err := database.New(database.Config{
		Path: filepath.Join(t.TempDir(), "notes.db"), MaxOpenConns: 8,
		BusyTimeout: 5000, JournalMode: "WAL", SynchronousMode: "NORMAL",
	}, testutil.TestLogger())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	repo := NewSnippetRepository(db.DB)
	files := NewSnippetFileRepository(db.DB)
	note, err := repo.Create(ctx, &models.SnippetInput{Title: "Notebook", Description: "keep this", Language: "markdown", Content: "Main"})
	if err != nil {
		t.Fatal(err)
	}
	saved, err := files.SyncFiles(ctx, note.ID, []models.SnippetFileInput{
		{Filename: "main.md", Language: "markdown", Content: "Main"},
		{Filename: "notes.md", Language: "markdown", Content: "Intro  \r\n"},
		{Filename: "run.sh", Language: "bash", Content: "echo hello"},
	})
	if err != nil || len(saved) != 3 {
		t.Fatalf("files: %+v, %v", saved, err)
	}

	const captures = 16
	var group sync.WaitGroup
	for i := range captures {
		group.Go(func() {
			_, err := repo.AppendNote(ctx, note.ID, models.NoteAppendInput{
				FileID: saved[1].ID, Content: fmt.Sprintf("capture-%02d", i),
			})
			if err != nil {
				t.Error(err)
			}
		})
	}
	group.Wait()
	got, err := files.GetBySnippetID(ctx, note.ID)
	if err != nil || len(got) != 3 {
		t.Fatalf("files: %+v, %v", got, err)
	}
	for i := range captures {
		if strings.Count(got[1].Content, fmt.Sprintf("capture-%02d", i)) != 1 {
			t.Fatalf("capture %d was lost or duplicated in %q", i, got[1].Content)
		}
	}
	if !strings.HasPrefix(got[1].Content, "Intro  \r\n\r\n") || got[0].Content != "Main" || got[2].Content != "echo hello" {
		t.Fatalf("unexpected file changes: %+v", got)
	}
	before, err := repo.AppendNote(ctx, note.ID, models.NoteAppendInput{FileID: saved[0].ID, Content: "primary update"})
	if err != nil || before == nil || before.Files[0].Content != "Main" {
		t.Fatalf("history snapshot: %+v, %v", before, err)
	}
	note, err = repo.GetByID(ctx, note.ID)
	if err != nil || note.Content != "Main\n\nprimary update" || note.Description != "keep this" {
		t.Fatalf("primary content and metadata: %+v, %v", note, err)
	}
	for _, fileID := range []int64{0, saved[2].ID, saved[2].ID + 100} {
		if _, err := repo.AppendNote(ctx, note.ID, models.NoteAppendInput{FileID: fileID, Content: "invalid"}); !errors.Is(err, ErrNotFound) {
			t.Errorf("expected unavailable file %d, got %v", fileID, err)
		}
	}
}

func TestNoteTargetsAndAppendEligibility(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewSnippetRepository(db)
	files := NewSnippetFileRepository(db)
	ctx := t.Context()
	create := func(input models.SnippetInput) *models.Snippet {
		t.Helper()
		snippet, err := repo.Create(ctx, &input)
		if err != nil {
			t.Fatal(err)
		}
		return snippet
	}
	legacy := create(models.SnippetInput{Title: "Legacy notes", Language: "markdown", Content: "old\n\n"})
	future := time.Now().Add(time.Hour).In(time.FixedZone("PDT", -7*60*60))
	expiring := create(models.SnippetInput{Title: "Scheduled expiration", Language: "markdown", Content: "later", ExpiresAt: &future})
	targets, err := repo.ListNoteTargets(ctx, "scheduled", 1, 10)
	if err != nil || len(targets.Targets) != 1 || targets.Targets[0].SnippetID != expiring.ID {
		t.Fatalf("unexpired note with an expiration: %+v, %v", targets, err)
	}
	if _, err := repo.AppendNote(ctx, expiring.ID, models.NoteAppendInput{Content: "still available"}); err != nil {
		t.Fatalf("append before expiration: %v", err)
	}
	bundle := create(models.SnippetInput{Title: "Music", Language: "plaintext", Content: "unchanged"})
	bundleFiles, err := files.SyncFiles(ctx, bundle.ID, []models.SnippetFileInput{{Filename: "NOTES.md", Language: "markdown", Content: "new"}})
	if err != nil {
		t.Fatal(err)
	}
	expired := time.Now().Add(-time.Hour)
	for _, input := range []models.SnippetInput{
		{Title: "Public notes", Language: "markdown", IsPublic: true},
		{Title: "Archived notes", Language: "markdown", IsArchived: true},
		{Title: "Expired notes", Language: "markdown", ExpiresAt: &expired},
		{Title: "Text notes", Language: "plaintext"},
	} {
		id := create(input).ID
		if _, err := repo.AppendNote(ctx, id, models.NoteAppendInput{Content: "skip"}); !errors.Is(err, ErrNotFound) {
			t.Errorf("append to %s: %v", input.Title, err)
		}
	}
	deleted := create(models.SnippetInput{Title: "Deleted notes", Language: "markdown"})
	if _, err := db.Exec("UPDATE snippets SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?", deleted.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.AppendNote(ctx, deleted.ID, models.NoteAppendInput{Content: "skip"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("append to deleted note: %v", err)
	}
	first, err := repo.ListNoteTargets(ctx, "notes", 1, 1)
	if err != nil || len(first.Targets) != 1 || !first.HasMore {
		t.Fatalf("first page: %+v, %v", first, err)
	}
	second, err := repo.ListNoteTargets(ctx, "notes", 2, 1)
	if err != nil || len(second.Targets) != 1 || second.HasMore || second.Targets[0] == first.Targets[0] {
		t.Fatalf("second page: %+v, %v", second, err)
	}
	if _, err := repo.AppendNote(ctx, legacy.ID, models.NoteAppendInput{FileID: bundleFiles[0].ID, Content: "invalid"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("append using another snippet's file: %v", err)
	}
	if _, err := repo.AppendNote(ctx, legacy.ID, models.NoteAppendInput{Content: strings.Repeat("é", models.MaxNoteContentBytes/2)}); !errors.Is(err, ErrNoteTooLarge) {
		t.Fatalf("oversized capture: %v", err)
	}
	previous, err := repo.AppendNote(ctx, legacy.ID, models.NoteAppendInput{Content: "new"})
	if err != nil || previous.Content != "old\n\n" {
		t.Fatalf("pre-append state: %+v, %v", previous, err)
	}
	after, err := repo.GetByID(ctx, legacy.ID)
	if err != nil || after.Content != "old\n\nnew" {
		t.Fatalf("legacy result: %+v, %v", after, err)
	}
}
