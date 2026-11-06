package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MohamedElashri/snipo/internal/auth"
	"github.com/MohamedElashri/snipo/internal/config"
	"github.com/MohamedElashri/snipo/internal/testutil"
)

func TestLoginRateLimitHonorsConfiguredWindow(t *testing.T) {
	db := testutil.TestDB(t)
	logger := testutil.TestLogger()
	service := auth.NewService(db, "fixture-password", "fixture-secret", time.Hour, logger, false)
	router := NewRouter(RouterConfig{DB: db, Logger: logger, AuthService: service, Config: &config.Config{Auth: config.AuthConfig{SessionSecret: "fixture-secret", EncryptionSalt: "fixture-salt"}}, RateLimit: 1, RateLimitWindow: time.Second})
	attempt := func() int {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		return response.Code
	}
	// Empty passwords avoid any password-hash work while exercising the real route.
	if got := attempt(); got != http.StatusBadRequest {
		t.Fatalf("first attempt: %d", got)
	}
	if got := attempt(); got != http.StatusTooManyRequests {
		t.Fatalf("limit not enforced: %d", got)
	}
	time.Sleep(1100 * time.Millisecond)
	if got := attempt(); got != http.StatusBadRequest {
		t.Fatalf("configured one-second window was ignored: %d", got)
	}
}

func TestDisableLoginTokenCreationStillRequiresOwnerPassword(t *testing.T) {
	db := testutil.TestDB(t)
	logger := testutil.TestLogger()
	if _, err := db.Exec("UPDATE settings SET disable_login=1 WHERE id=1"); err != nil {
		t.Fatal(err)
	}
	service := auth.NewService(db, "fixture-password", "fixture-secret", time.Hour, logger, false)
	router := NewRouter(RouterConfig{DB: db, Logger: logger, AuthService: service, Config: &config.Config{Auth: config.AuthConfig{SessionSecret: "fixture-secret", EncryptionSalt: "fixture-salt"}, Features: config.FeatureFlags{APITokens: true}}, RateLimit: 100})
	attempt := func(header string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/tokens", bytes.NewBufferString(`{"name":"owner-created","permissions":"read","password":"fixture-password"}`))
		req.Header.Set("Content-Type", "application/json")
		if header != "" {
			req.Header.Set("X-Snipo-Master-Password", header)
		}
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		return response
	}
	if response := attempt(""); response.Code != http.StatusUnauthorized {
		t.Fatalf("missing challenge: %d %s", response.Code, response.Body.String())
	}
	if response := attempt("fixture-password"); response.Code != http.StatusCreated {
		t.Fatalf("owner challenge rejected: %d %s", response.Code, response.Body.String())
	}
}
