package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/MohamedElashri/snipo/pkg/models"
)

func TestAPIRateLimiter_Basic(t *testing.T) {
	config := RateLimitConfig{
		ReadLimit:  5, // 5 requests
		WriteLimit: 3,
		AdminLimit: 2,
		Window:     time.Second,
	}

	rl := NewAPIRateLimiter(config)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test read limit
	t.Run("read_limit", func(t *testing.T) {
		handler := rl.RateLimitByPermission(PermissionRead)(testHandler)

		// Make requests up to limit
		for i := 0; i < 5; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("request %d: expected 200, got %d", i+1, rr.Code)
			}

			// Check rate limit headers
			if limit := rr.Header().Get("X-RateLimit-Limit"); limit != "5" {
				t.Errorf("expected X-RateLimit-Limit: 5, got %s", limit)
			}
		}

		// Next request should be rate limited
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusTooManyRequests {
			t.Errorf("expected 429, got %d", rr.Code)
		}

		if retryAfter := rr.Header().Get("Retry-After"); retryAfter == "" {
			t.Error("expected Retry-After header")
		}
	})
}

func TestAPIRateLimiter_PerToken(t *testing.T) {
	config := RateLimitConfig{
		ReadLimit: 3,
		Window:    time.Second,
	}

	rl := NewAPIRateLimiter(config)
	handler := rl.RateLimitByPermission(PermissionRead)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Create two different tokens
	token1 := &models.APIToken{ID: 1, Name: "token1", Permissions: PermissionRead}
	token2 := &models.APIToken{ID: 2, Name: "token2", Permissions: PermissionRead}

	// Each token should have its own limit
	for i := 0; i < 3; i++ {
		// Token 1
		req1 := httptest.NewRequest("GET", "/test", nil)
		ctx1 := context.WithValue(req1.Context(), ContextKeyAPIToken, token1)
		req1 = req1.WithContext(ctx1)
		rr1 := httptest.NewRecorder()
		handler.ServeHTTP(rr1, req1)

		if rr1.Code != http.StatusOK {
			t.Errorf("token1 request %d: expected 200, got %d", i+1, rr1.Code)
		}

		// Token 2
		req2 := httptest.NewRequest("GET", "/test", nil)
		ctx2 := context.WithValue(req2.Context(), ContextKeyAPIToken, token2)
		req2 = req2.WithContext(ctx2)
		rr2 := httptest.NewRecorder()
		handler.ServeHTTP(rr2, req2)

		if rr2.Code != http.StatusOK {
			t.Errorf("token2 request %d: expected 200, got %d", i+1, rr2.Code)
		}
	}

	// Next request for token1 should be limited
	req1 := httptest.NewRequest("GET", "/test", nil)
	ctx1 := context.WithValue(req1.Context(), ContextKeyAPIToken, token1)
	req1 = req1.WithContext(ctx1)
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 for token1, got %d", rr1.Code)
	}

	// But token2 should also be limited (both used up their 3 requests)
	req2 := httptest.NewRequest("GET", "/test", nil)
	ctx2 := context.WithValue(req2.Context(), ContextKeyAPIToken, token2)
	req2 = req2.WithContext(ctx2)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 for token2, got %d", rr2.Code)
	}
}

func TestAPIRateLimiter_Headers(t *testing.T) {
	config := RateLimitConfig{
		ReadLimit: 10,
		Window:    time.Hour,
	}

	rl := NewAPIRateLimiter(config)
	handler := rl.RateLimitByPermission(PermissionRead)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check headers
	headers := []string{"X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"}
	for _, h := range headers {
		if val := rr.Header().Get(h); val == "" {
			t.Errorf("expected %s header to be set", h)
		}
	}

	// Check specific values
	if limit := rr.Header().Get("X-RateLimit-Limit"); limit != "10" {
		t.Errorf("expected limit 10, got %s", limit)
	}

	remaining, _ := strconv.Atoi(rr.Header().Get("X-RateLimit-Remaining"))
	if remaining != 9 { // 10 - 1 request
		t.Errorf("expected remaining 9, got %d", remaining)
	}
}

func TestAPIRateLimiter_DifferentPermissions(t *testing.T) {
	config := RateLimitConfig{
		ReadLimit:  10,
		WriteLimit: 5,
		AdminLimit: 2,
		Window:     time.Second,
	}

	rl := NewAPIRateLimiter(config)

	tests := []struct {
		permission string
		limit      int
	}{
		{PermissionRead, 10},
		{PermissionWrite, 5},
		{PermissionAdmin, 2},
	}

	for _, tt := range tests {
		t.Run(tt.permission, func(t *testing.T) {
			handler := rl.RateLimitByPermission(tt.permission)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			// Use unique identifier for each test
			token := &models.APIToken{
				ID:          int64(time.Now().UnixNano()),
				Name:        fmt.Sprintf("test-%s", tt.permission),
				Permissions: tt.permission,
			}

			// Make requests up to limit
			for i := 0; i < tt.limit; i++ {
				req := httptest.NewRequest("GET", "/test", nil)
				ctx := context.WithValue(req.Context(), ContextKeyAPIToken, token)
				req = req.WithContext(ctx)
				rr := httptest.NewRecorder()
				handler.ServeHTTP(rr, req)

				if rr.Code != http.StatusOK {
					t.Errorf("request %d: expected 200, got %d", i+1, rr.Code)
				}
			}

			// Next request should fail
			req := httptest.NewRequest("GET", "/test", nil)
			ctx := context.WithValue(req.Context(), ContextKeyAPIToken, token)
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusTooManyRequests {
				t.Errorf("expected 429, got %d", rr.Code)
			}
		})
	}
}

func TestAPIRateLimiter_ConvenienceFunctions(t *testing.T) {
	config := RateLimitConfig{
		ReadLimit:  10,
		WriteLimit: 5,
		AdminLimit: 2,
		Window:     time.Second,
	}

	rl := NewAPIRateLimiter(config)

	tests := []struct {
		name     string
		handler  http.Handler
		expected string
	}{
		{"RateLimitRead", rl.RateLimitRead(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})), "10"},
		{"RateLimitWrite", rl.RateLimitWrite(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})), "5"},
		{"RateLimitAdmin", rl.RateLimitAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})), "2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()
			tt.handler.ServeHTTP(rr, req)

			if limit := rr.Header().Get("X-RateLimit-Limit"); limit != tt.expected {
				t.Errorf("expected limit %s, got %s", tt.expected, limit)
			}
		})
	}
}

func TestAPIRateLimiter_DefaultConfig(t *testing.T) {
	// Test that zero values get set to defaults
	rl := NewAPIRateLimiter(RateLimitConfig{})

	if rl.readLimit != 1000 {
		t.Errorf("expected default readLimit 1000, got %d", rl.readLimit)
	}
	if rl.writeLimit != 500 {
		t.Errorf("expected default writeLimit 500, got %d", rl.writeLimit)
	}
	if rl.adminLimit != 100 {
		t.Errorf("expected default adminLimit 100, got %d", rl.adminLimit)
	}
	if rl.window != time.Hour {
		t.Errorf("expected default window 1h, got %v", rl.window)
	}
}
