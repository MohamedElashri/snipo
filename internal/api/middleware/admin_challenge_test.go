package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MohamedElashri/snipo/internal/auth"
	"github.com/MohamedElashri/snipo/internal/repository"
	"github.com/MohamedElashri/snipo/internal/testutil"
	"github.com/MohamedElashri/snipo/pkg/models"
)

func TestDisableLoginAdminChallengeThroughAuthentication(t *testing.T) {
	db := testutil.TestDB(t)
	ctx := context.Background()
	if _, err := db.Exec("UPDATE settings SET disable_login=1 WHERE id=1"); err != nil {
		t.Fatal(err)
	}
	service := auth.NewService(db, "correct-password", "fixture-secret", time.Hour, testutil.TestLogger(), false)
	tokens := repository.NewTokenRepository(db, []byte("fixture-hmac-key"))
	fixtures := map[string]string{}
	for _, permission := range []string{"read", "write", "admin"} {
		token, err := tokens.Create(ctx, &models.APITokenInput{Name: permission, Permissions: permission})
		if err != nil {
			t.Fatal(err)
		}
		fixtures[permission] = token.Token
	}
	handler := RequireAuthWithSettings(service, tokens, repository.NewSettingsRepository(db))(
		RequireAdminWithPassword(service)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })))
	for i, tc := range []struct {
		name, permission, password string
		want                       int
	}{
		{name: "anonymous requires password", want: 401},
		{name: "anonymous correct password", password: "correct-password", want: 204},
		{name: "anonymous incorrect password", password: "incorrect", want: 403},
		{name: "read token cannot escalate with password", permission: "read", password: "correct-password", want: 403},
		{name: "write token cannot escalate with password", permission: "write", password: "correct-password", want: 403},
		{name: "admin token remains sufficient", permission: "admin", want: 204},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/admin", nil)
			req.RemoteAddr = fmt.Sprintf("192.0.2.%d:9000", i+1)
			if tc.permission != "" {
				req.Header.Set("Authorization", "Bearer "+fixtures[tc.permission])
			}
			if tc.password != "" {
				req.Header.Set(adminPasswordHeader, tc.password)
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, req)
			if response.Code != tc.want {
				t.Fatalf("want %d, got %d: %s", tc.want, response.Code, response.Body.String())
			}
		})
	}
}
