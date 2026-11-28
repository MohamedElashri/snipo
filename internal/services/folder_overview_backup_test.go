package services

import (
	"testing"

	"github.com/MohamedElashri/snipo/pkg/models"
)

func TestBackupRestoresFolderOverviewToImportedNote(t *testing.T) {
	backup, snippets := backupFixture(t)
	ctx := t.Context()
	folder, err := backup.folderRepo.Create(ctx, &models.FolderInput{Name: "Runbooks"})
	if err != nil {
		t.Fatal(err)
	}
	note, err := snippets.Create(ctx, &models.SnippetInput{
		Title: "README", Language: "markdown", Content: "# Runbooks", FolderID: &folder.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := backup.folderRepo.SetOverview(ctx, folder.ID, &note.ID); err != nil {
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
			restored, err := newBackup.folderRepo.List(ctx)
			if err != nil || len(restored) != 1 || restored[0].OverviewSnippetID == nil || *restored[0].OverviewSnippetID == note.ID {
				t.Fatalf("restored OverviewSnippetID: %+v, %v", restored, err)
			}
			snippet, err := newSnippets.GetByID(ctx, *restored[0].OverviewSnippetID)
			if err != nil || snippet.Content != "# Runbooks" || len(snippet.Folders) != 1 || snippet.Folders[0].ID != restored[0].ID {
				t.Fatalf("restored overview note: %+v, %v", snippet, err)
			}
		})
	}
}
