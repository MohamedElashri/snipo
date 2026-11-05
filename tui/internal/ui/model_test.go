package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MohamedElashri/snipo/tui/internal/api"
)

func TestSubmitPreservesCodeAndVisibility(t *testing.T) {
	for _, mode := range []ViewMode{ViewCreate, ViewEdit} {
		t.Run(map[ViewMode]string{ViewCreate: "create", ViewEdit: "edit"}[mode], func(t *testing.T) {
			const content = "    indented()\n\n"
			var saved api.SnippetInput
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := json.NewDecoder(r.Body).Decode(&saved); err != nil {
					t.Error(err)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"data":{"id":"example","title":"Example"}}`))
			}))
			defer server.Close()
			expires := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
			snippet := &api.Snippet{ID: "example", Title: "Example", Content: content, IsPublic: true,
				IsArchived: true, ExpiresAt: &expires, Folders: []api.Folder{{ID: 7}},
				Files: []api.File{{ID: 1, Filename: "main.txt", Content: "old"}, {ID: 2, Filename: "other.txt", Content: "unchanged"}}}

			m := Model{client: api.NewClient(server.URL, "token"), mode: mode, width: 100, height: 40, detailSnippet: snippet}
			m.initEditForm(snippet)
			_, cmd := m.submitForm()
			if cmd == nil {
				t.Fatal("expected a save command")
			}
			if err, ok := cmd().(errMsg); ok {
				t.Fatalf("save failed: %v", err)
			}
			if saved.Content != content {
				t.Errorf("content = %q, want %q", saved.Content, content)
			}

			if mode == ViewEdit {
				if !saved.IsArchived || saved.ExpiresAt == nil || !saved.ExpiresAt.Equal(expires) || saved.FolderID == nil || *saved.FolderID != 7 {
					t.Errorf("metadata was lost: %+v", saved)
				}
				if len(saved.Files) != 2 || saved.Files[0].ID != 1 || saved.Files[0].Filename != "main.txt" || saved.Files[0].Content != content || saved.Files[1].Content != "unchanged" {
					t.Errorf("files were not preserved: %+v", saved.Files)
				}
			} else if saved.IsArchived || saved.ExpiresAt != nil || saved.FolderID != nil || len(saved.Files) != 0 {
				t.Errorf("new snippet inherited existing metadata: %+v", saved)
			}
			if saved.IsPublic != (mode == ViewEdit) {
				t.Errorf("public = %v for mode %v", saved.IsPublic, mode)
			}
		})
	}
}
