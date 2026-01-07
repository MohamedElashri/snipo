package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MohamedElashri/snipo/internal/models"
)

func TestCheckPermission(t *testing.T) {
	tests := []struct {
		name          string
		tokenPerm     string
		requiredPerm  string
		expectAllowed bool
	}{
		// Admin can do everything
		{"admin can read", PermissionAdmin, PermissionRead, true},
		{"admin can write", PermissionAdmin, PermissionWrite, true},
		{"admin can admin", PermissionAdmin, PermissionAdmin, true},

		// Write can read and write
		{"write can read", PermissionWrite, PermissionRead, true},
		{"write can write", PermissionWrite, PermissionWrite, true},
		{"write cannot admin", PermissionWrite, PermissionAdmin, false},

		// Read can only read
		{"read can read", PermissionRead, PermissionRead, true},
		{"read cannot write", PermissionRead, PermissionWrite, false},
		{"read cannot admin", PermissionRead, PermissionAdmin, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test handler
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("success"))
			})

			// Create token with specified permission
			token := &models.APIToken{
				ID:          1,
				Name:        "test-token",
				Permissions: tt.tokenPerm,
			}

			// Create request with token in context
			req := httptest.NewRequest("GET", "/test", nil)
			ctx := context.WithValue(req.Context(), ContextKeyAPIToken, token)
			req = req.WithContext(ctx)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Apply permission middleware
			handler := CheckPermission(tt.requiredPerm)(testHandler)
			handler.ServeHTTP(rr, req)

			// Check result
			if tt.expectAllowed {
				if rr.Code != http.StatusOK {
					t.Errorf("expected status OK, got %d", rr.Code)
				}
			} else {
				if rr.Code != http.StatusForbidden {
					t.Errorf("expected status Forbidden, got %d", rr.Code)
				}
			}
		})
	}
}

func TestCheckPermission_NoToken(t *testing.T) {
	// When no token in context (session-based auth), should allow
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler := CheckPermission(PermissionAdmin)(testHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status OK for session auth, got %d", rr.Code)
	}
}

func TestGetTokenFromContext(t *testing.T) {
	token := &models.APIToken{
		ID:          42,
		Name:        "test-token",
		Permissions: PermissionRead,
	}

	ctx := context.WithValue(context.Background(), ContextKeyAPIToken, token)

	retrieved := GetTokenFromContext(ctx)
	if retrieved == nil {
		t.Fatal("expected token, got nil")
		return
	}

	if retrieved.ID != token.ID {
		t.Errorf("expected token ID %d, got %d", token.ID, retrieved.ID)
	}

	// Test with no token
	emptyCtx := context.Background()
	retrieved = GetTokenFromContext(emptyCtx)
	if retrieved != nil {
		t.Errorf("expected nil for empty context, got %v", retrieved)
	}
}

func TestHasPermission(t *testing.T) {
	tests := []struct {
		tokenPerm string
		required  string
		expected  bool
	}{
		{PermissionAdmin, PermissionRead, true},
		{PermissionAdmin, PermissionWrite, true},
		{PermissionAdmin, PermissionAdmin, true},
		{PermissionWrite, PermissionRead, true},
		{PermissionWrite, PermissionWrite, true},
		{PermissionWrite, PermissionAdmin, false},
		{PermissionRead, PermissionRead, true},
		{PermissionRead, PermissionWrite, false},
		{PermissionRead, PermissionAdmin, false},
		{"invalid", PermissionRead, false},
	}

	for _, tt := range tests {
		t.Run(tt.tokenPerm+"_requires_"+tt.required, func(t *testing.T) {
			result := hasPermission(tt.tokenPerm, tt.required)
			if result != tt.expected {
				t.Errorf("hasPermission(%s, %s) = %v, expected %v",
					tt.tokenPerm, tt.required, result, tt.expected)
			}
		})
	}
}

func TestConvenienceMiddleware(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name         string
		middleware   func(http.Handler) http.Handler
		tokenPerm    string
		expectStatus int
	}{
		{"RequireRead with read token", RequireRead, PermissionRead, http.StatusOK},
		{"RequireRead with write token", RequireRead, PermissionWrite, http.StatusOK},
		{"RequireWrite with read token", RequireWrite, PermissionRead, http.StatusForbidden},
		{"RequireWrite with write token", RequireWrite, PermissionWrite, http.StatusOK},
		{"RequireAdmin with read token", RequireAdmin, PermissionRead, http.StatusForbidden},
		{"RequireAdmin with admin token", RequireAdmin, PermissionAdmin, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &models.APIToken{
				ID:          1,
				Name:        "test",
				Permissions: tt.tokenPerm,
			}

			req := httptest.NewRequest("GET", "/test", nil)
			ctx := context.WithValue(req.Context(), ContextKeyAPIToken, token)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler := tt.middleware(testHandler)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectStatus {
				t.Errorf("expected status %d, got %d", tt.expectStatus, rr.Code)
			}
		})
	}
}
