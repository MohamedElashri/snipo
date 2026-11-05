package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
			snippet := &api.Snippet{ID: "example", Title: "Example", Content: content, IsPublic: true}
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
			if saved.IsPublic != (mode == ViewEdit) {
				t.Errorf("public = %v for mode %v", saved.IsPublic, mode)
			}
		})
	}
}
