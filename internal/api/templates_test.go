package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/MohamedElashri/snipo/internal/auth"
	"github.com/MohamedElashri/snipo/internal/config"
	"github.com/MohamedElashri/snipo/internal/database"
	"github.com/MohamedElashri/snipo/internal/repository"
	"github.com/MohamedElashri/snipo/internal/services"
	"github.com/MohamedElashri/snipo/pkg/models"
)

func TestTemplateLifecycle(t *testing.T) {
	ctx := t.Context()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	db, err := database.New(database.Config{
		Path: filepath.Join(t.TempDir(), "templates.db"), MaxOpenConns: 1, BusyTimeout: 5000,
		JournalMode: "WAL", SynchronousMode: "NORMAL",
	}, logger)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	owner := auth.NewService(db.DB, "template-test-password", "template-test-secret", time.Hour, logger, false)
	cfg := &config.Config{Auth: config.AuthConfig{SessionSecret: "template-test-secret", EncryptionSalt: "template-test-salt"}}
	router := NewRouter(RouterConfig{DB: db.DB, Logger: logger, AuthService: owner, Config: cfg, RateLimit: 100, MaxFilesPerSnippet: 3})
	tokenRepo := repository.NewTokenRepository(db.DB, deriveHMACKey(cfg.Auth.SessionSecret, "snipo-token-hmac"))
	read, err := tokenRepo.Create(ctx, &models.APITokenInput{Name: "reader", Permissions: "read"})
	if err != nil {
		t.Fatal(err)
	}
	write, err := tokenRepo.Create(ctx, &models.APITokenInput{Name: "writer", Permissions: "write"})
	if err != nil {
		t.Fatal(err)
	}
	request := func(method, path, token string, body any, status int, target any) {
		t.Helper()
		var payload []byte
		if body != nil {
			var err error
			payload, err = json.Marshal(body)
			if err != nil {
				t.Fatal(err)
			}
		}
		req := httptest.NewRequest(method, path, bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != status {
			t.Fatalf("%s %s: got %d, want %d: %s", method, path, rec.Code, status, rec.Body.String())
		}
		if target != nil {
			var env struct {
				Data json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(env.Data, target); err != nil {
				t.Fatal(err)
			}
		}
	}
	folders := repository.NewFolderRepository(db.DB)
	parent, err := folders.Create(ctx, &models.FolderInput{Name: "Infrastructure"})
	if err != nil {
		t.Fatal(err)
	}
	folder, err := folders.Create(ctx, &models.FolderInput{Name: "Services", ParentID: &parent.ID})
	if err != nil {
		t.Fatal(err)
	}
	snipRepo := repository.NewSnippetRepository(db.DB)
	fileRepo := repository.NewSnippetFileRepository(db.DB)
	tagRepo := repository.NewTagRepository(db.DB)
	snips := services.NewSnippetService(snipRepo, logger).WithFileRepo(fileRepo).WithTagRepo(tagRepo).WithFolderRepo(folders)
	backup := services.NewBackupService(db.DB, snips, tagRepo, folders, fileRepo, logger, "template-test-salt").
		WithTemplateRepo(repository.NewTemplateRepository(db.DB))

	const base = "/api/v1/templates"
	request("GET", base, "", nil, 401, nil)
	def := services.TemplateStarters()[2]
	def.FolderID = &folder.ID
	def.Tags = []string{"infra"}
	def.Presets = []models.TemplatePreset{{Name: "Local", Values: map[string]string{"service": "redis"}}}
	def.Files[2].Content += "\n{{!snipo.service}} {{ .Values.name }}\n"
	request("POST", base, read.Token, def, 403, nil)
	var first models.Template
	request("POST", base, write.Token, def, 201, &first)
	var summaries []models.TemplateSummary
	request("GET", base+"?q=env.example", read.Token, nil, 200, &summaries)
	if len(summaries) != 1 || len(summaries[0].Filenames) != 3 {
		t.Fatalf("unexpected summaries: %+v", summaries)
	}

	var changed models.Template
	path := base + "/" + first.ID
	def.Files[2].Content = "New revision for {{snipo.service}}"
	request("PUT", path, write.Token, map[string]any{"revision": 1, "definition": def}, 200, &changed)
	request("PUT", path, write.Token, map[string]any{"revision": 1, "definition": def}, 409, nil)
	if changed.Revision != 2 {
		t.Fatalf("unexpected revision: %d", changed.Revision)
	}
	render := path + "/render"
	request("POST", render, read.Token, map[string]any{"revision": 1}, 400, nil)
	var draft models.SnippetInput
	request("POST", render, read.Token,
		map[string]any{"revision": 1, "values": map[string]string{"service": "redis"}}, 200, &draft)
	if draft.TemplateOrigin == nil || draft.TemplateOrigin.Revision != 1 || len(draft.Files) != 3 ||
		!strings.Contains(draft.Files[0].Content, "${SERVICE_IMAGE}") ||
		!strings.Contains(draft.Files[2].Content, "{{snipo.service}} {{ .Values.name }}") {
		t.Fatalf("unexpected draft: %+v", draft)
	}
	bad := draft
	unavailable := int64(999)
	bad.FolderID = &unavailable
	request("POST", "/api/v1/snippets", write.Token, bad, 400, nil)
	var count int
	if err := db.QueryRow("SELECT count(*) FROM snippets").Scan(&count); err != nil || count != 0 {
		t.Fatalf("failed save left %d snippets: %v", count, err)
	}

	draft.Files[1].Content = "SERVICE_IMAGE=redis:alpine"
	var saved models.Snippet
	request("POST", "/api/v1/snippets", write.Token, draft, 201, &saved)
	request("GET", "/api/v1/snippets/"+saved.ID, read.Token, nil, 200, &saved)
	if len(saved.Files) != 3 || saved.Files[1].Content != draft.Files[1].Content || saved.IsPublic ||
		len(saved.Folders) != 1 || saved.Folders[0].ID != folder.ID || saved.TemplateOrigin.Revision != 1 {
		t.Fatalf("incomplete snippet: %+v", saved)
	}

	request("PUT", path+"/status", write.Token, map[string]string{"state": "archived"}, 204, nil)
	request("GET", base, read.Token, nil, 200, &summaries)
	if len(summaries) != 0 {
		t.Fatalf("archived template remains in the picker")
	}
	request("POST", path+"/restore-revision", write.Token,
		map[string]int{"revision": 1, "current_revision": 2}, 200, &changed)
	if changed.Revision != 3 {
		t.Fatalf("restore should create a new revision: %+v", changed)
	}

	for _, format := range []string{"json", "zip"} {
		t.Run(format, func(t *testing.T) {
			body, _, err := backup.Export(ctx, models.ExportOptions{Format: format})
			if err != nil {
				t.Fatal(err)
			}
			result, err := backup.Import(ctx, body, models.ImportOptions{Strategy: "replace"})
			if err != nil || len(result.Errors) != 0 || result.TemplatesImported != 1 || result.SnippetsImported != 1 {
				t.Fatalf("backup restore failed: %+v; %v", result, err)
			}
			var restored models.Template
			request("GET", path, read.Token, nil, 200, &restored)
			var history []models.TemplateRevision
			request("GET", path+"/history", read.Token, nil, 200, &history)
			if restored.Revision != 3 || len(history) != 3 || len(restored.Definition.Presets) != 1 ||
				restored.Definition.FolderID == nil || *restored.Definition.FolderID == folder.ID {
				t.Fatalf("template history or folder was lost: %+v", restored)
			}
			var restoredID string
			if err := db.QueryRow("SELECT id FROM snippets").Scan(&restoredID); err != nil {
				t.Fatal(err)
			}
			request("GET", "/api/v1/snippets/"+restoredID, read.Token, nil, 200, &saved)
			if saved.TemplateOrigin == nil || saved.TemplateOrigin.ID != first.ID || saved.TemplateOrigin.Revision != 1 ||
				len(saved.Files) != 3 || saved.Files[1].Content != draft.Files[1].Content {
				t.Fatalf("incomplete restored snippet: %+v", saved)
			}
		})
	}

	var exported models.TemplateExport
	request("GET", path+"/export", read.Token, nil, 200, &exported)
	if exported.Version != 1 || exported.ID != first.ID || exported.Definition.FolderID != nil || len(exported.Definition.Presets) != 0 {
		t.Fatalf("unexpected portable definition: %+v", exported)
	}
	importPath := base + "/import"
	var preview struct {
		Conflict *models.Template `json:"conflict"`
	}
	request("POST", importPath, write.Token, map[string]any{"mode": "preview", "template": exported}, 200, &preview)
	if preview.Conflict == nil || preview.Conflict.ID != first.ID {
		t.Fatalf("import preview should identify the existing template: %+v", preview)
	}
	var imported models.Template
	request("POST", importPath, write.Token, map[string]any{"mode": "copy", "template": exported}, 201, &imported)
	if imported.ID == first.ID || imported.Revision != 1 {
		t.Fatalf("keeping both should create a separate template: %+v", imported)
	}
	exported.ID = strings.Repeat("f", 32)
	request("POST", importPath, write.Token, map[string]any{"mode": "copy", "template": exported}, 201, &imported)
	if imported.ID != exported.ID {
		t.Fatalf("first import should preserve the exported ID: %+v", imported)
	}
	portablePath := base + "/" + imported.ID
	localDef := imported.Definition
	localDef.Fields = append(localDef.Fields,
		models.TemplateField{Key: "mode", Label: "Mode", Type: "choice", Choices: []string{"staging"}},
		models.TemplateField{Key: "host", Label: "Host", Type: "text"},
	)
	localPresets := []models.TemplatePreset{{Name: "Staging", Values: map[string]string{"service": "redis", "mode": "staging", "host": "cache"}}}
	localDef.Presets = localPresets
	request("PUT", portablePath, write.Token, map[string]any{"revision": 1, "definition": localDef}, 200, nil)
	exported.Definition.Fields = append(exported.Definition.Fields,
		models.TemplateField{Key: "mode", Label: "Mode", Type: "choice", Choices: []string{"production"}})
	request("POST", importPath, write.Token, map[string]any{"mode": "replace", "revision": 2, "template": exported}, 200, &imported)
	if imported.ID != exported.ID || imported.Revision != 3 {
		t.Fatalf("replacing should create a revision: %+v", imported)
	}
	if expected := []models.TemplatePreset{{Name: "Staging", Values: map[string]string{"service": "redis"}}}; !reflect.DeepEqual(imported.Definition.Presets, expected) {
		t.Fatalf("portable replacement should keep compatible local preset values: %+v", imported.Definition.Presets)
	}
	var previousImport models.Template
	request("GET", portablePath+"?revision=2", read.Token, nil, 200, &previousImport)
	if !reflect.DeepEqual(previousImport.Definition.Presets, localPresets) {
		t.Fatalf("previous revision lost its local presets: %+v", previousImport.Definition.Presets)
	}

	// A rendered draft must fit the save request limit after escaping its text.
	largeDef := models.TemplateDefinition{
		Name: "Large note", Title: "Large note",
		Fields: []models.TemplateField{{Key: "body", Label: "Body", Type: "multiline", Required: true}},
		Files:  []models.SnippetFileInput{{Filename: "note.txt", Content: "{{snipo.body}}", Language: "plaintext"}},
	}
	request("POST", base, write.Token, largeDef, 201, &imported)
	largePath := base + "/" + imported.ID
	var largeDraft models.SnippetInput
	largeValue := strings.Repeat("\\", 550000)
	request("POST", largePath+"/render", read.Token, map[string]any{
		"revision": 1, "values": map[string]string{"body": largeValue},
	}, 200, &largeDraft)
	if largeDraft.Content != "" || largeDraft.Files[0].Content != largeValue {
		t.Fatal("draft must carry the primary content only in its file")
	}
	var largeSnippet models.Snippet
	request("POST", "/api/v1/snippets", write.Token, largeDraft, 201, &largeSnippet)
	if largeSnippet.Content != largeValue || largeSnippet.Files[0].Content != largeValue {
		t.Fatal("the saved snippet did not preserve the file and its legacy primary content")
	}

	largeDef.Files[0].Content = "{{snipo.body}}{{snipo.body}}"
	request("PUT", largePath, write.Token, map[string]any{"revision": 1, "definition": largeDef}, 200, nil)
	request("POST", largePath+"/render", read.Token, map[string]any{
		"revision": 2, "values": map[string]string{"body": strings.Repeat("\x00", 200000)},
	}, 400, nil)

	invalidOrigin := *largeSnippet.TemplateOrigin
	invalidOrigin.Revision = 0
	broken := largeSnippet
	broken.TemplateOrigin = &invalidOrigin
	brokenBackup, err := json.Marshal(models.BackupData{Version: "1.0", Snippets: []models.Snippet{broken}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := backup.Import(ctx, brokenBackup, models.ImportOptions{Strategy: "replace"}); !errors.Is(err, services.ErrInvalidBackupFormat) {
		t.Errorf("replace accepted an invalid origin: %v", err)
	}
	var preserved int
	if err := db.QueryRow("SELECT COUNT(*) FROM snippets WHERE id = ?", largeSnippet.ID).Scan(&preserved); err != nil || preserved != 1 {
		t.Fatalf("invalid backup deleted an existing snippet: %d, %v", preserved, err)
	}

	request("PUT", "/api/v1/snippets/"+largeSnippet.ID, write.Token, map[string]any{
		"title": "Edited as a plain note", "content": "Content independent from its template",
		"language": "plaintext", "files": []models.SnippetFileInput{},
	}, 200, nil)
	editedBackup, _, err := backup.Export(ctx, models.ExportOptions{Format: "json"})
	if err != nil {
		t.Fatal(err)
	}
	restoreEdited, err := backup.Import(ctx, editedBackup, models.ImportOptions{Strategy: "replace"})
	if err != nil || len(restoreEdited.Errors) != 0 {
		t.Fatalf("restoring an edited template snippet failed: %+v, %v", restoreEdited, err)
	}
	var editedID string
	if err := db.QueryRow("SELECT id FROM snippets WHERE title = 'Edited as a plain note'").Scan(&editedID); err != nil {
		t.Fatal(err)
	}
	var edited models.Snippet
	request("GET", "/api/v1/snippets/"+editedID, read.Token, nil, 200, &edited)
	if len(edited.Files) != 0 || edited.Content != "Content independent from its template" || edited.TemplateOrigin == nil ||
		edited.TemplateOrigin.Revision != 1 {
		t.Fatalf("restoring the edited note lost its content or origin: %+v", edited)
	}

	// Pre-template backups can recreate an existing destination with a new ID.
	request("GET", path, read.Token, nil, 200, &changed)
	before := changed.Definition.FolderID
	if before == nil {
		t.Fatal("missing destination before legacy restore")
	}
	legacyFolders, err := folders.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	legacy, err := json.Marshal(models.BackupData{Version: "1.0", Snippets: []models.Snippet{}, Folders: legacyFolders})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := backup.Import(ctx, legacy, models.ImportOptions{Strategy: "replace"}); err != nil {
		t.Fatal(err)
	}
	var legacyDraft models.SnippetInput
	request("POST", render, read.Token, map[string]any{
		"revision": 1, "values": map[string]string{"service": "redis"},
	}, 200, &legacyDraft)
	if legacyDraft.FolderID == nil || *legacyDraft.FolderID == *before {
		t.Fatalf("legacy restore left the historical destination at %v", legacyDraft.FolderID)
	}
	legacyFolder, err := folders.GetByID(ctx, *legacyDraft.FolderID)
	if err != nil || legacyFolder.ParentID == nil || legacyFolder.Name != "Services" {
		t.Fatalf("legacy restore lost the nested destination: %+v %v", legacyFolder, err)
	}
	if err := folders.Delete(ctx, legacyFolder.ID); err != nil {
		t.Fatal(err)
	}
	legacyDraft = models.SnippetInput{}
	request("POST", render, read.Token, map[string]any{
		"revision": 1, "values": map[string]string{"service": "redis"},
	}, 200, &legacyDraft)
	if legacyDraft.FolderID != nil {
		t.Fatal("a deleted default folder should yield an unfiled draft")
	}

	request("POST", path+"/restore-revision", write.Token, map[string]int{"revision": 1}, 400, nil)
	var restoredWithoutFolder models.Template
	request("POST", path+"/restore-revision", write.Token,
		map[string]int{"revision": 1, "current_revision": 3}, 200, &restoredWithoutFolder)
	if restoredWithoutFolder.Revision != 4 || restoredWithoutFolder.Definition.FolderID != nil {
		t.Fatalf("restoring history should clear a deleted folder association: %+v", restoredWithoutFolder)
	}

	if _, err := backup.Import(ctx, []byte(`{"version":"1.0","snippets":[]}`), models.ImportOptions{Strategy: "replace"}); err != nil {
		t.Fatal(err)
	}
	request("GET", path, read.Token, nil, 200, &changed)
	request("PUT", path+"/status", write.Token, map[string]string{"state": "trash"}, 204, nil)
	request("POST", render, read.Token, map[string]any{"revision": 1, "values": map[string]string{"service": "redis"}}, 404, nil)
	request("DELETE", path, write.Token, nil, 204, nil)
	request("GET", path, read.Token, nil, 404, nil)
}
