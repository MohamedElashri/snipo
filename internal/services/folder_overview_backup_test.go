package services

import (
	"testing"

	"github.com/MohamedElashri/snipo/pkg/models"
)

func TestBackupRestoresFolderOverviewToImportedNote(t *testing.T) {
	backup, snippets := backupFixture(t)
	ctx := t.Context()
	parent, err := backup.folderRepo.Create(ctx, &models.FolderInput{Name: "Projects"})
	if err != nil {
		t.Fatal(err)
	}
	folder, err := backup.folderRepo.Create(ctx, &models.FolderInput{Name: "Runbooks", ParentID: &parent.ID})
	if err != nil {
		t.Fatal(err)
	}
	note, err := snippets.Create(ctx, &models.SnippetInput{
		Kind: models.KindNote, Title: "README", Tags: []string{"runbook"}, Language: "markdown", Content: "# Runbooks", FolderID: &folder.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := backup.folderRepo.SetOverview(ctx, folder.ID, &note.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := snippets.Create(ctx, &models.SnippetInput{Title: "README", Language: "markdown", Content: "A snippet with the same title"}); err != nil {
		t.Fatal(err)
	}
	for _, format := range []string{"json", "zip"} {
		t.Run(format, func(t *testing.T) {
			content, _, err := backup.Export(ctx, models.ExportOptions{Format: format})
			if err != nil {
				t.Fatal(err)
			}
			newBackup, newSnippets := backupFixture(t)
			result, err := newBackup.Import(ctx, content, models.ImportOptions{Strategy: "merge"})
			if err != nil || len(result.Errors) != 0 {
				t.Fatalf("import: %+v, %v", result, err)
			}
			restored, err := newBackup.folderRepo.ListTree(ctx)
			if err != nil || len(restored) != 1 || len(restored[0].Children) != 1 {
				t.Fatalf("restored nested folders: %+v, %v", restored, err)
			}
			runbooks := restored[0].Children[0]
			if runbooks.OverviewSnippetID == nil || *runbooks.OverviewSnippetID == note.ID {
				t.Fatalf("restored overview: %+v", runbooks)
			}
			snippet, err := newSnippets.GetByID(ctx, *runbooks.OverviewSnippetID)
			if err != nil || snippet.Kind != models.KindNote || snippet.Content != "# Runbooks" ||
				len(snippet.Tags) != 1 || snippet.Tags[0].Name != "runbook" ||
				len(snippet.Folders) != 1 || snippet.Folders[0].ID != runbooks.ID {
				t.Fatalf("restored overview note: %+v, %v", snippet, err)
			}
			items, err := newSnippets.List(ctx, models.SnippetFilter{Query: "README"})
			if err != nil || items.Pagination.Total != 2 {
				t.Fatalf("expected both items named README: %+v, %v", items, err)
			}
		})
	}
}
