package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/MohamedElashri/snipo/pkg/models"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestSummaryPreservesFullContentAndFindsSecondaryFile(t *testing.T) {
	original := []models.Snippet{{Title: "Example", Content: "legacy", Files: []models.SnippetFile{{Filename: "main.sh", Content: "echo hello"}, {Filename: "README.md", Content: "Introduction\nİstanbul\nCanary-000\nDetails\nMore"}}}}
	got := summarizeSnippets(original, "canary-000")
	if len(got) != 1 || !got[0].IsSummary || got[0].PreviewFilename != "README.md" || got[0].Preview != "Canary-000\nDetails" {
		t.Fatalf("unexpected summary: %+v", got)
	}
	if got[0].Content != "" || got[0].Files[1].Content != "" {
		t.Fatal("summary includes full content")
	}
	if original[0].Content != "legacy" || original[0].Files[1].Content == "" {
		t.Fatal("projection mutated source")
	}
}
func TestPreviewBoundsUnicodeAndLines(t *testing.T) {
	got := snippetPreview(strings.Repeat("界", 600), "")
	if !utf8.ValidString(got) || len([]rune(got)) != 501 {
		t.Fatal("preview must be bounded without splitting Unicode")
	}
	if got := snippetPreview("İstanbul\nTARGET\nnext\nlast", "target"); got != "TARGET\nnext" {
		t.Fatalf("wrong match: %q", got)
	}
	if got := snippetPreview("one\ntwo\nthree", ""); got != "one\ntwo" {
		t.Fatalf("wrong excerpt: %q", got)
	}
}

func TestListSummaryIsOptInAndDoesNotChangeDetail(t *testing.T) {
	handler, _ := setupSnippetHandler(t)
	body := []byte(`{"title":"Projection","content":"full primary","language":"bash","files":[{"filename":"main.sh","language":"bash","content":"full primary"},{"filename":"README.md","language":"markdown","content":"Introduction\nneedle\nMore"}]}`)
	created := httptest.NewRecorder()
	handler.Create(created, withRequestID(httptest.NewRequest(http.MethodPost, "/api/v1/snippets", bytes.NewReader(body))))
	if created.Code != http.StatusCreated {
		t.Fatalf("create: %s", created.Body.String())
	}
	var initial struct {
		Data models.Snippet `json:"data"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &initial); err != nil {
		t.Fatal(err)
	}
	for _, summary := range []bool{false, true} {
		url := "/api/v1/snippets?limit=20"
		if summary {
			url += "&view=summary&q=needle"
		}
		response := httptest.NewRecorder()
		handler.List(response, withRequestID(httptest.NewRequest(http.MethodGet, url, nil)))
		if response.Code != http.StatusOK {
			t.Fatalf("list: %s", response.Body.String())
		}
		var list struct {
			Data []snippetSummary `json:"data"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &list); err != nil {
			t.Fatal(err)
		}
		if len(list.Data) != 1 {
			t.Fatalf("expected one result: %s", response.Body.String())
		}
		got := list.Data[0]
		if got.IsSummary != summary {
			t.Fatal("wrong representation")
		}
		if summary {
			if got.Content != "" || got.Files[1].Content != "" || got.PreviewFilename != "README.md" || got.Preview != "needle\nMore" {
				t.Fatalf("wrong projection: %+v", got)
			}
		} else if got.Content != "full primary" || got.Files[1].Content != "Introduction\nneedle\nMore" {
			t.Fatal("default list content changed")
		}
	}
	response := httptest.NewRecorder()
	req := withURLParams(withRequestID(httptest.NewRequest(http.MethodGet, "/api/v1/snippets/"+initial.Data.ID, nil)), map[string]string{"id": initial.Data.ID})
	handler.Get(response, req)
	var detail struct {
		Data models.Snippet `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &detail); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusOK || detail.Data.Files[1].Content != "Introduction\nneedle\nMore" {
		t.Fatalf("detail content changed: %s", response.Body.String())
	}
}
