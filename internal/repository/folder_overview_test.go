package repository

import (
	"errors"
	"testing"
	"time"

	"github.com/MohamedElashri/snipo/internal/testutil"
	"github.com/MohamedElashri/snipo/pkg/models"
)

func TestFolderOverviewLifecycle(t *testing.T) {
	db := testutil.TestDB(t)
	ctx := t.Context()
	folders := NewFolderRepository(db)
	snippets := NewSnippetRepository(db)
	files := NewSnippetFileRepository(db)
	root, err := folders.Create(ctx, &models.FolderInput{Name: "Root"})
	if err != nil {
		t.Fatal(err)
	}
	child, err := folders.Create(ctx, &models.FolderInput{Name: "Runbooks", ParentID: &root.ID})
	if err != nil {
		t.Fatal(err)
	}
	create := func(input models.SnippetInput) string {
		t.Helper()
		snippet, err := snippets.Create(ctx, &input)
		if err != nil {
			t.Fatal(err)
		}
		if err := folders.SetSnippetFolder(ctx, snippet.ID, &child.ID); err != nil {
			t.Fatal(err)
		}
		return snippet.ID
	}
	legacy := create(models.SnippetInput{Title: "Legacy note", Language: "markdown"})
	bundle := create(models.SnippetInput{Title: "Mixed runbook", Language: "bash"})
	if _, err := files.SyncFiles(ctx, bundle, []models.SnippetFileInput{
		{Filename: "run.sh", Language: "bash", Content: "echo hi"},
		{Filename: "notes.md", Language: "markdown", Content: "# Notes"},
	}); err != nil {
		t.Fatal(err)
	}
	plain := create(models.SnippetInput{Title: "Code", Language: "bash"})
	past := time.Now().Add(-time.Hour)
	expired := create(models.SnippetInput{Title: "Expired", Language: "markdown", ExpiresAt: &past})
	archived := create(models.SnippetInput{Title: "Archived", Language: "markdown", IsArchived: true})
	future := time.Now().Add(time.Hour).In(time.FixedZone("PDT", -7*60*60))
	expiring := create(models.SnippetInput{Title: "Expires later", Language: "markdown", ExpiresAt: &future})

	set := func(id *string) {
		t.Helper()
		if err := folders.SetOverview(ctx, child.ID, id); err != nil {
			t.Fatal(err)
		}
	}
	assertRead := func(want *string) {
		t.Helper()
		got, err := folders.GetByID(ctx, child.ID)
		if err != nil {
			t.Fatal(err)
		}
		match := func(id *string) bool {
			return (id == nil && want == nil) || (id != nil && want != nil && *id == *want)
		}
		if !match(got.OverviewSnippetID) {
			t.Fatalf("unexpected assigned note: %+v", got)
		}
		tree, err := folders.ListTree(ctx)
		if err != nil || len(tree) != 1 || len(tree[0].Children) != 1 || !match(tree[0].Children[0].OverviewSnippetID) {
			t.Fatalf("unexpected folder tree: %+v, %v", tree, err)
		}
	}
	set(&expiring)
	assertRead(&expiring)
	set(&legacy)
	assertRead(&legacy)
	for _, invalid := range []string{plain, expired, archived} {
		if err := folders.SetOverview(ctx, child.ID, &invalid); !errors.Is(err, ErrInvalidOverview) {
			t.Errorf("invalid note %s: %v", invalid, err)
		}
	}
	if err := folders.SetOverview(ctx, root.ID, &legacy); !errors.Is(err, ErrInvalidOverview) {
		t.Fatalf("note outside the folder: %v", err)
	}
	if err := folders.SetOverview(ctx, 9999, nil); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing folder: %v", err)
	}
	assertRead(&legacy)

	set(&bundle)
	assertRead(&bundle)
	if err := folders.SetSnippetFolder(ctx, bundle, &root.ID); err != nil {
		t.Fatal(err)
	}
	assertRead(nil)
	if err := folders.SetSnippetFolder(ctx, bundle, &child.ID); err != nil {
		t.Fatal(err)
	}
	assertRead(&bundle)
	if err := snippets.Delete(ctx, bundle, true); err != nil {
		t.Fatal(err)
	}
	assertRead(nil)

	set(&legacy)
	set(nil)
	assertRead(nil)
	if snippet, err := snippets.GetByID(ctx, legacy); err != nil || snippet == nil {
		t.Fatalf("clear removed the note: %v", err)
	}
}
