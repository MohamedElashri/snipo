package services

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/MohamedElashri/snipo/internal/repository"
	"github.com/MohamedElashri/snipo/internal/testutil"
	"github.com/MohamedElashri/snipo/pkg/models"
)

func TestImportRejectsOversizedBackups(t *testing.T) {
	const limit = 50 << 20
	svc := &BackupService{}
	raw := bytes.Repeat([]byte(" "), limit+1)
	for _, zipped := range []bool{false, true} {
		content := raw
		if zipped {
			var archive bytes.Buffer
			zw := zip.NewWriter(&archive)
			metadata, err := zw.Create("metadata.json")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := metadata.Write(raw); err != nil {
				t.Fatal(err)
			}
			if err := zw.Close(); err != nil {
				t.Fatal(err)
			}
			content = archive.Bytes()
		}
		_, err := svc.Import(context.Background(), content, models.ImportOptions{})
		if err == nil || !strings.Contains(err.Error(), "size limit") {
			t.Errorf("zipped=%v: expected size limit error, got %v", zipped, err)
		}
	}
}

func TestBoundedImportPreservesNormalBackups(t *testing.T) {
	for _, format := range []string{"json", "zip"} {
		for _, password := range []string{"", "fixture-password"} {
			t.Run(format+"/"+map[bool]string{false: "plain", true: "encrypted"}[password != ""], func(t *testing.T) {
				db := testutil.TestDB(t)
				logger := testutil.TestLogger()
				tags := repository.NewTagRepository(db)
				folders := repository.NewFolderRepository(db)
				files := repository.NewSnippetFileRepository(db)
				snippets := NewSnippetService(repository.NewSnippetRepository(db), logger).WithTagRepo(tags).WithFolderRepo(folders).WithFileRepo(files)
				svc := NewBackupService(db, snippets, tags, folders, files, logger, "fixture-salt")
				ctx := context.Background()
				_, err := snippets.Create(ctx, &models.SnippetInput{Title: "Original", Language: "plaintext", Files: []models.SnippetFileInput{{Filename: "example.txt", Content: "private backup marker", Language: "plaintext"}}})
				if err != nil {
					t.Fatal(err)
				}
				content, _, err := svc.Export(ctx, models.ExportOptions{Format: format, Password: password})
				if err != nil {
					t.Fatal(err)
				}
				result, err := svc.Import(ctx, content, models.ImportOptions{Strategy: "replace", Password: password})
				if err != nil {
					t.Fatal(err)
				}
				if result.SnippetsImported != 1 || len(result.Errors) != 0 {
					t.Fatalf("import: %+v", result)
				}
				list, err := snippets.List(ctx, models.SnippetFilter{Limit: 10})
				if err != nil || len(list.Data) != 1 {
					t.Fatalf("restored snippets: %+v, %v", list, err)
				}
				restored, err := snippets.GetByID(ctx, list.Data[0].ID)
				if err != nil || len(restored.Files) != 1 || restored.Files[0].Content != "private backup marker" {
					t.Fatalf("restored content: %+v, %v", restored, err)
				}
			})
		}
	}
}

func backupFixture(t *testing.T) (*BackupService, *SnippetService) {
	t.Helper()
	db := testutil.TestDB(t)
	logger := testutil.TestLogger()
	tags := repository.NewTagRepository(db)
	folders := repository.NewFolderRepository(db)
	files := repository.NewSnippetFileRepository(db)
	snippets := NewSnippetService(repository.NewSnippetRepository(db), logger).WithTagRepo(tags).WithFolderRepo(folders).WithFileRepo(files)
	return NewBackupService(db, snippets, tags, folders, files, logger, "fixture-salt"), snippets
}

func TestInvalidReplacementBackupPreservesExistingData(t *testing.T) {
	for _, content := range []string{`null`, `{}`, `{"version":"1.0"}`, `{"version":"unexpected"}`, `{"version":"1.0","snippets":[{"title":"Invalid","content":""}]}`} {
		svc, snippets := backupFixture(t)
		ctx := context.Background()
		original, err := snippets.Create(ctx, &models.SnippetInput{Title: "Keep me", Content: "Original code", Language: "plaintext"})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := svc.Import(ctx, []byte(content), models.ImportOptions{Strategy: "replace"}); err == nil {
			t.Errorf("invalid replacement accepted: %s", content)
		}
		if _, err := snippets.GetByID(ctx, original.ID); err != nil {
			t.Errorf("invalid replacement deleted original data: %s: %v", content, err)
		}
	}
}

func TestZipWithoutMetadataCannotReplaceData(t *testing.T) {
	for _, withFile := range []bool{false, true} {
		t.Run(fmt.Sprintf("with-file=%t", withFile), func(t *testing.T) {
			svc, snippets := backupFixture(t)
			ctx := context.Background()
			original, err := snippets.Create(ctx, &models.SnippetInput{Title: "Keep me", Content: "Original code", Language: "plaintext"})
			if err != nil {
				t.Fatal(err)
			}
			var archive bytes.Buffer
			zw := zip.NewWriter(&archive)
			if withFile {
				file, err := zw.Create("unrelated.txt")
				if err != nil {
					t.Fatal(err)
				}
				if _, err := file.Write([]byte("This is not a Snipo backup")); err != nil {
					t.Fatal(err)
				}
			}
			if err := zw.Close(); err != nil {
				t.Fatal(err)
			}
			if _, err := svc.Import(ctx, archive.Bytes(), models.ImportOptions{Strategy: "replace"}); !errors.Is(err, ErrInvalidBackupFormat) {
				t.Errorf("ZIP without metadata must be rejected: %v", err)
			}
			if _, err := snippets.GetByID(ctx, original.ID); err != nil {
				t.Errorf("ZIP without metadata deleted existing data: %v", err)
			}
		})
	}
}

func TestBackupExportIncludesEveryPage(t *testing.T) {
	svc, snippets := backupFixture(t)
	ctx := context.Background()
	for i := 0; i < 101; i++ {
		if _, err := snippets.Create(ctx, &models.SnippetInput{Title: fmt.Sprintf("Snippet %03d", i), Content: "code", Language: "plaintext"}); err != nil {
			t.Fatal(err)
		}
	}
	archived, err := snippets.Create(ctx, &models.SnippetInput{Title: "Archived", Content: "Archived code", Language: "plaintext", IsArchived: true})
	if err != nil {
		t.Fatal(err)
	}
	content, _, err := svc.Export(ctx, models.ExportOptions{Format: "json"})
	if err != nil {
		t.Fatal(err)
	}
	var backup models.BackupData
	if err := json.Unmarshal(content, &backup); err != nil {
		t.Fatal(err)
	}
	if len(backup.Snippets) != 102 {
		t.Fatalf("backup silently omitted snippets: got %d, want 102 including archived", len(backup.Snippets))
	}
	foundArchived := false
	for _, snippet := range backup.Snippets {
		if snippet.ID == archived.ID && snippet.IsArchived {
			foundArchived = true
		}
	}
	if !foundArchived {
		t.Fatal("archived snippet was omitted")
	}
	merged, err := svc.Import(ctx, content, models.ImportOptions{Strategy: "merge"})
	if err != nil || merged.SnippetsImported != 0 {
		t.Fatalf("merge missed existing snippets beyond the first page: %+v %v", merged, err)
	}
}

func TestGenuineEmptyBackupCanReplaceData(t *testing.T) {
	svc, snippets := backupFixture(t)
	ctx := context.Background()
	content, _, err := svc.Export(ctx, models.ExportOptions{Format: "json"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := snippets.Create(ctx, &models.SnippetInput{Title: "Temporary", Content: "code", Language: "plaintext"}); err != nil {
		t.Fatal(err)
	}
	result, err := svc.Import(ctx, content, models.ImportOptions{Strategy: "replace"})
	if err != nil || result.SnippetsImported != 0 {
		t.Fatalf("genuine empty backup rejected: %+v %v", result, err)
	}
	remaining, err := svc.listBackupSnippets(ctx)
	if err != nil || len(remaining) != 0 {
		t.Fatalf("empty replacement failed: %d %v", len(remaining), err)
	}
}

func TestExportFailsRatherThanReturningIncompleteMetadata(t *testing.T) {
	for _, table := range []string{"tags", "folders"} {
		svc, _ := backupFixture(t)
		if _, err := svc.db.Exec("DROP TABLE " + table); err != nil {
			t.Fatal(err)
		}
		content, filename, err := svc.Export(context.Background(), models.ExportOptions{Format: "json"})
		if err == nil || content != nil || filename != "" {
			t.Fatalf("failed %s lookup produced a partial backup: filename=%q err=%v", table, filename, err)
		}
	}
}
