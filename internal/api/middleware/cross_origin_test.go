package middleware

import (
	"context"
	"github.com/MohamedElashri/snipo/pkg/models"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBrowserWriteBoundary(t *testing.T) {
	for _, tc := range []struct {
		name, method, origin string
		token, anonymous     bool
		trusted              []string
		want                 int
	}{
		{name: "same-site foreign origin", method: "POST", origin: "https://other.example.test", want: 403},
		{name: "safe public read", method: "GET", origin: "https://other.example.test", want: 204},
		{name: "non-browser client", method: "POST", want: 204},
		{name: "same origin", method: "POST", origin: "https://snipo.example.test", want: 204},
		{name: "validated API token", method: "POST", origin: "https://other.example.test", token: true, want: 204},
		{name: "anonymous synthetic token", method: "POST", origin: "https://other.example.test", token: true, anonymous: true, want: 403},
		{name: "explicit trusted origin", method: "POST", origin: "https://other.example.test", trusted: []string{"https://other.example.test"}, want: 204},
		{name: "wildcard is not session trust", method: "POST", origin: "https://other.example.test", trusted: []string{"*"}, want: 403},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "https://snipo.example.test/api/v1/snippets", nil)
			req.Header.Set("Origin", tc.origin)
			if tc.token {
				req = req.WithContext(context.WithValue(req.Context(), ContextKeyAPIToken, &models.APIToken{Permissions: PermissionWrite}))
			}
			if tc.anonymous {
				req = req.WithContext(context.WithValue(req.Context(), ContextKeyAnonymousAccess, true))
			}
			rr := httptest.NewRecorder()
			ProtectBrowserWrites(tc.trusted)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })).ServeHTTP(rr, req)
			if rr.Code != tc.want {
				t.Fatalf("status=%d want=%d", rr.Code, tc.want)
			}
		})
	}
}
