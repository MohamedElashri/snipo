package ui

import (
	"encoding/json"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/MohamedElashri/snipo/pkg/models"
	"github.com/MohamedElashri/snipo/tui/internal/api"
	"github.com/MohamedElashri/snipo/tui/internal/config"
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
				Tags: []api.Tag{{Name: "API"}}, IsArchived: true, ExpiresAt: &expires, Folders: []api.Folder{{ID: 7}},
				Files: []api.File{{ID: 1, Filename: "main.txt", Content: "old"}, {ID: 2, Filename: "other.txt", Content: "unchanged"}}}

			m := Model{client: api.NewClient(server.URL, "token"), mode: mode, width: 100, height: 40, detailSnippet: snippet}
			m.initEditForm(snippet)
			m.textarea.SetValue(content)
			_, cmd := m.submitForm()
			if cmd == nil {
				t.Fatal("expected a save command")
			}
			if err, ok := cmd().(errMsg); ok {
				t.Fatalf("save failed: %v", err)
			}
			if len(saved.Tags) != 1 || saved.Tags[0] != "API" {
				t.Errorf("tag capitalization lost: %v", saved.Tags)
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

func TestQuestionMarkInForm(t *testing.T) {
	m := Model{mode: ViewCreate, width: 100, height: 40}
	m.initCreateForm()
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	got := next.(Model)
	if got.mode != ViewCreate || got.inputs[0].Value() != "?" {
		t.Fatalf("question mark lost draft: mode %v value %q", got.mode, got.inputs[0].Value())
	}
}

func TestNewNoteMarkdownPreviewPreservesDraft(t *testing.T) {
	m := Model{mode: ViewList, libraryKind: models.KindSnippet, width: 80, height: 24}
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}})
	m = next.(Model)
	if m.mode != ViewCreate || !m.noteForm || m.inputs[1].Value() != "markdown" {
		t.Fatal("N should start a Markdown note from the Snippets library")
	}
	m.inputs[0].SetValue("Release notes")
	const body = "# Preview heading\n\n"
	m.textarea.SetValue(body + strings.Repeat("Line of text\n\n", 30))
	draft := m.textarea.Value()
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyF2})
	m = next.(Model)
	if view := m.View(); !m.notePreview || !strings.Contains(view, " heading") || !strings.Contains(view, "back to writing") || strings.Contains(view, "# Preview") {
		t.Fatalf("expected a rendered Markdown preview: %q", view)
	}
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = next.(Model)
	if m.detailScroll != 1 {
		t.Fatal("expected preview to scroll")
	}
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(Model)
	if m.notePreview || m.mode != ViewCreate || m.textarea.Value() != draft {
		t.Fatal("Escape should return to the original writing buffer")
	}
	m.inputs[1].SetValue("plaintext")
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyF2})
	if next.(Model).notePreview {
		t.Fatal("plain-text notes should stay in the writer")
	}
}

func TestPlainTextNoteReaderUsesBodyWithoutFileTabs(t *testing.T) {
	m := Model{mode: ViewDetail, width: 36, height: 24, detailSnippet: &api.Snippet{
		Kind: models.KindNote, Language: "plaintext", Title: "Notes", Content: "# literal text",
		Files: []api.File{{Filename: "hidden.txt", Content: "stale file content"}},
	}}
	view := m.View()
	if !strings.Contains(view, "# literal text") || strings.Contains(view, "hidden.txt") || strings.Contains(view, "stale file content") {
		t.Fatalf("expected the note's plain text without file tabs: %q", view)
	}
}
func TestCopySelectedFile(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("macOS clipboard integration")
	}
	dir := t.TempDir()
	out := filepath.Join(dir, "copied")
	if err := os.WriteFile(filepath.Join(dir, "pbcopy"), []byte("#!/bin/sh\ncat > "+out+"\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
	m := Model{mode: ViewDetail, selectedFileIdx: 1, detailSnippet: &api.Snippet{Content: "primary", Files: []api.File{{Content: "primary"}, {Content: "secondary"}}}}
	_, cmd := m.updateDetail(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	cmd()
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "secondary" {
		t.Fatalf("copied %q", data)
	}
}
func TestEditSelectedFile(t *testing.T) {
	var saved api.SnippetInput
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&saved)
		_, _ = w.Write([]byte(`{"data":{"title":"saved"}}`))
	}))
	defer server.Close()
	snippet := &api.Snippet{ID: "files", Title: "Files", Content: "primary", Language: "go", Files: []api.File{{Filename: "main.go", Content: "primary", Language: "go"}, {Filename: "other.py", Content: "secondary", Language: "python"}}}
	m := Model{mode: ViewDetail, allowedLanguages: []string{"go", "python"}, width: 100, height: 40, selectedFileIdx: 1, detailSnippet: snippet, client: api.NewClient(server.URL, "token")}
	next, _ := m.updateDetail(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m = next.(Model)
	if m.textarea.Value() != "secondary" || m.inputs[1].Value() != "python" {
		t.Fatalf("editing wrong file: %q %q", m.textarea.Value(), m.inputs[1].Value())
	}
	m.textarea.SetValue("updated")
	_, cmd := m.submitForm()
	if cmd == nil {
		t.Fatal(m.err)
	}
	cmd()
	if saved.Files[0].Content != "primary" || saved.Files[1].Content != "updated" {
		t.Fatalf("wrong file overwritten: %+v", saved.Files)
	}
}

func TestLongSnippetAllowsNewlines(t *testing.T) {
	content := strings.Repeat("line\n", 250)
	for _, editing := range []bool{false, true} {
		m := Model{width: 100, height: 40}
		if editing {
			m.initEditForm(&api.Snippet{Content: content})
		} else {
			m.initCreateForm()
			m.textarea.SetValue(content)
		}
		m.textarea.Focus()
		before := m.textarea.Value()
		m.textarea, _ = m.textarea.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if len(m.textarea.Value()) != len(before)+1 {
			t.Fatalf("editing=%v: Enter stopped working in a long snippet", editing)
		}
	}
}

func TestLanguageMetadataDoesNotAutocompleteAPIKey(t *testing.T) {
	m := Model{mode: ViewSettings, config: &config.Config{APIKey: "go"}}
	m.initSettingsForm()
	next, _ := m.Update(languagesLoadedMsg{languages: []string{"golang"}})
	got := next.(Model)
	if got.inputs[1].ShowSuggestions {
		t.Fatal("API key field acquired language suggestions")
	}
}

func TestHelpReturnsToSelectedSnippet(t *testing.T) {
	for _, key := range []tea.KeyMsg{{Type: tea.KeyRunes, Runes: []rune{'?'}}, {Type: tea.KeyEsc}} {
		m := Model{mode: ViewDetail, selectedFileIdx: 1, detailScroll: 12, detailSnippet: &api.Snippet{ID: "keep"}}
		next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
		next, _ = next.(Model).Update(key)
		got := next.(Model)
		if got.mode != ViewDetail || got.selectedFileIdx != 1 || got.detailScroll != 12 || got.detailSnippet.ID != "keep" {
			t.Fatalf("help lost position: %+v", got)
		}
	}
}

func TestSuccessfulRetryClearsOldError(t *testing.T) {
	for _, message := range []tea.Msg{snippetsLoadedMsg{}, snippetLoadedMsg{snippet: &api.Snippet{ID: "ok"}}, successMsg{message: "Saved"}} {
		m := Model{mode: ViewList, err: fmt.Errorf("connection refused")}
		next, _ := m.Update(message)
		if next.(Model).err != nil {
			t.Fatalf("successful %T kept old error", message)
		}
	}
}
