package api

import (
	"encoding/json"
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
}
