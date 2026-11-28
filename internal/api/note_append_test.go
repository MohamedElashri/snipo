package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MohamedElashri/snipo/internal/auth"
	"github.com/MohamedElashri/snipo/internal/config"
	"github.com/MohamedElashri/snipo/internal/repository"
	"github.com/MohamedElashri/snipo/internal/testutil"
	"github.com/MohamedElashri/snipo/pkg/models"
)

func TestNoteCaptureHTTPPermissionsAndHistory(t *testing.T) {
	ctx := t.Context()
	db := testutil.TestDB(t)
	logger := testutil.TestLogger()
	const secret = "test-note-secret"
	authService := auth.NewService(db, "test-note-password", secret, time.Hour, logger, false)
	router := NewRouter(RouterConfig{DB: db, Logger: logger, AuthService: authService,
		Config: &config.Config{
			Auth: config.AuthConfig{SessionSecret: secret, EncryptionSalt: "test-note-salt"},
			API:  config.APIConfig{RateLimitRead: 100, RateLimitWrite: 100, RateLimitAdmin: 100},
		},
	})
	tokens := repository.NewTokenRepository(db, deriveHMACKey(secret, "snipo-token-hmac"))
	makeToken := func(permission string) string {
		t.Helper()
		token, err := tokens.Create(ctx, &models.APITokenInput{Name: permission, Permissions: permission})
		if err != nil {
			t.Fatal(err)
		}
		return token.Token
	}
	writeToken, readToken := makeToken("write"), makeToken("read")
	call := func(method, path, token, body string, want int) *httptest.ResponseRecorder {
		t.Helper()
		request := httptest.NewRequest(method, "/api/v1/snippets"+path, strings.NewReader(body))
		if token != "" {
			request.Header.Set("Authorization", "Bearer "+token)
		}
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != want {
			t.Fatalf("%s %s: want status %d, got %d: %s", method, path, want, response.Code, response.Body)
		}
		return response
	}
	if _, err := db.ExecContext(ctx, "UPDATE settings SET history_enabled = 1 WHERE id = 1"); err != nil {
		t.Fatal(err)
	}
	created := call(http.MethodPost, "", writeToken, `{"title":"Runbook","language":"markdown","content":"# Notes"}`, http.StatusCreated)
	var snippet struct {
		Data models.Snippet
	}
	if err := json.Unmarshal(created.Body.Bytes(), &snippet); err != nil {
		t.Fatal(err)
	}
	id := "/" + snippet.Data.ID
	call(http.MethodGet, "/note-targets", "", "", http.StatusUnauthorized)
	call(http.MethodGet, "/note-targets?page=0", readToken, "", http.StatusBadRequest)
	list := call(http.MethodGet, "/note-targets?q=run", readToken, "", http.StatusOK)
	var picks struct {
		Data models.NoteTargets
	}
	if err := json.Unmarshal(list.Body.Bytes(), &picks); err != nil || len(picks.Data.Targets) != 1 || picks.Data.Targets[0].SnippetID != snippet.Data.ID {
		t.Fatalf("listed notes: %+v, %v", picks, err)
	}
	call(http.MethodPost, id+"/append", readToken, `{"content":"denied"}`, http.StatusForbidden)
	call(http.MethodPost, id+"/append", writeToken, `{"content":"  "}`, http.StatusBadRequest)
	call(http.MethodPost, id+"/append", writeToken, `{"content":"capture","file_id":-5}`, http.StatusBadRequest)
	call(http.MethodPost, id+"/append", writeToken, `{"content":"capture","file_id":5}`, http.StatusNotFound)
	call(http.MethodPost, id+"/append", writeToken, `{"content":"Capture"}`, http.StatusNoContent)
	detail := call(http.MethodGet, id, readToken, "", http.StatusOK)
	if err := json.Unmarshal(detail.Body.Bytes(), &snippet); err != nil || snippet.Data.Content != "# Notes\n\nCapture" {
		t.Fatalf("saved note: %+v, %v", snippet, err)
	}
	history := call(http.MethodGet, id+"/history", readToken, "", http.StatusOK)
	if !strings.Contains(history.Body.String(), `"content":"# Notes"`) {
		t.Fatalf("missing original history: %s", history.Body)
	}

	// A legacy Markdown snippet appears in the old list until explicitly converted.
	call(http.MethodGet, "?kind=snippet&q=Runbook", readToken, "", http.StatusOK)
	if result := call(http.MethodGet, "?kind=note&q=Runbook", readToken, "", http.StatusOK); strings.Contains(result.Body.String(), snippet.Data.ID) {
		t.Fatalf("legacy item appeared in Notes: %s", result.Body)
	}
	call(http.MethodPost, id+"/convert", readToken, `{"kind":"note"}`, http.StatusForbidden)
	converted := call(http.MethodPost, id+"/convert", writeToken, `{"kind":"note"}`, http.StatusOK)
	if err := json.Unmarshal(converted.Body.Bytes(), &snippet); err != nil || snippet.Data.Kind != models.KindNote || "/"+snippet.Data.ID != id {
		t.Fatalf("conversion changed identity: %+v, %v", snippet, err)
	}
	call(http.MethodPut, id, writeToken, `{"kind":"snippet","title":"Runbook","language":"markdown"}`, http.StatusBadRequest)
	call(http.MethodPut, id, writeToken, `{"title":"Runbook","language":"python","content":"body"}`, http.StatusBadRequest)
	call(http.MethodPut, id, writeToken, `{"title":"Runbook","files":[{"filename":"a.md","language":"markdown"},{"filename":"b.txt","language":"plaintext"}]}`, http.StatusBadRequest)
	call(http.MethodPut, id, writeToken, `{"title":"Runbook","language":"plaintext","content":"Keep  spaces"}`, http.StatusOK)
	if result := call(http.MethodGet, "?q=Runbook", readToken, "", http.StatusOK); strings.Contains(result.Body.String(), snippet.Data.ID) {
		t.Fatalf("old client list returned notes: %s", result.Body)
	}
	if result := call(http.MethodGet, "/note-targets?kind=note&q=Runbook", readToken, "", http.StatusOK); !strings.Contains(result.Body.String(), `"language":"plaintext"`) {
		t.Fatalf("plain text note missing from capture picker: %s", result.Body)
	}
	call(http.MethodPost, id+"/append", writeToken, `{"language":"plaintext","content":"captured"}`, http.StatusNoContent)
	stale := fmt.Sprintf(`{"title":"Runbook","revision":%d,"language":"plaintext","content":"my draft"}`, snippet.Data.Revision)
	call(http.MethodPut, id, writeToken, stale, http.StatusConflict)
	detail = call(http.MethodGet, id, readToken, "", http.StatusOK)
	if err := json.Unmarshal(detail.Body.Bytes(), &snippet); err != nil || snippet.Data.Content != "Keep  spaces\n\ncaptured" {
		t.Fatalf("capture was not retained: %+v, %v", snippet, err)
	}
	call(http.MethodPost, id+"/convert", writeToken, `{"kind":"snippet"}`, http.StatusOK)
}
