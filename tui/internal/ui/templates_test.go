package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MohamedElashri/snipo/pkg/models"
	"github.com/MohamedElashri/snipo/tui/internal/api"
	"github.com/MohamedElashri/snipo/tui/internal/config"
	tea "github.com/charmbracelet/bubbletea"
)

func TestTemplateWorkflowPreservesFilesAndOrigin(t *testing.T) {
	origin := models.TemplateOrigin{ID: "runbook", Revision: 2, Name: "Runbook"}
	files := []api.FileInput{
		{Filename: "runbook.md", Language: "markdown", Content: "# Redis"},
		{Filename: ".env.example", Language: "plaintext", Content: "PORT=6379"},
		{Filename: "check.sh", Language: "bash", Content: "redis-cli ping"},
	}
	var saved api.SnippetInput
	folder := int64(3)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/templates":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data":       []api.TemplateSummary{{ID: origin.ID, Revision: 2, Name: "Runbook", Filenames: []string{"runbook.md"}}},
				"pagination": map[string]int{"totalPages": 1},
			})
		case "/api/v1/templates/runbook":
			if r.URL.Query().Get("revision") != "2" {
				t.Error("get should pin a revision")
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": api.Template{
				ID: origin.ID, Revision: 2, Definition: models.TemplateDefinition{Name: "Runbook"},
			}})
		case "/api/v1/templates/runbook/render":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": api.SnippetInput{
				Title: "Redis runbook", Files: files, TemplateOrigin: &origin, FolderID: &folder,
			}})
		case "/api/v1/snippets":
			if err := json.NewDecoder(r.Body).Decode(&saved); err != nil {
				t.Error(err)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": api.Snippet{ID: "saved", Title: "Redis runbook"}})
		default:
			t.Errorf("unexpected request %s", r.URL)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	m := NewModel(&config.Config{ServerURL: server.URL, APIKey: "token"})
	m.width, m.height = 100, 35
	m.allowedLanguages = []string{"markdown", "plaintext", "bash"}
	m.folders = []api.Folder{{ID: folder, Name: "Services"}}

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	next, _ = next.Update(cmd())
	next, cmd = next.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next, cmd = next.Update(cmd())
	next, _ = next.Update(cmd())

	draft := next.(Model)
	if draft.mode != ViewCreate || draft.templates.draft == nil {
		t.Fatalf("template draft did not open: %v", draft.err)
	}
	rendered := draft.View()
	if !strings.Contains(rendered, "F2 Previous file") || strings.Count(rendered, "\n")+1 > draft.height {
		t.Fatal("template draft controls should fit the terminal viewport")
	}

	next, _ = draft.Update(tea.KeyMsg{Type: tea.KeyF3})
	draft = next.(Model)
	draft.textarea.SetValue("PORT=6380")
	next, cmd = draft.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if cmd == nil {
		t.Fatalf("save failed: %v", next.(Model).err)
	}
	if result := cmd().(templateSaved); result.err != nil {
		t.Fatal(result.err)
	}
	if len(saved.Files) != 3 || saved.Files[1].Content != "PORT=6380" || saved.Files[0].Content != files[0].Content ||
		saved.Files[2].Content != files[2].Content || saved.TemplateOrigin == nil || *saved.TemplateOrigin != origin ||
		saved.FolderID == nil || *saved.FolderID != folder || saved.IsPublic {
		t.Fatalf("incomplete template draft: %+v", saved)
	}
}
