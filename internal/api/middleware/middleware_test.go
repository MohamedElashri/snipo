package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestID(t *testing.T) {
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that request ID is in context
		requestID := GetRequestID(r.Context())
		if requestID == "" {
			t.Error("expected request ID in context, got empty string")
		}

		// Check that it looks like a UUID
		if len(requestID) != 36 || strings.Count(requestID, "-") != 4 {
			t.Errorf("request ID doesn't look like UUID: %s", requestID)
		}

		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check that X-Request-ID header is set
	requestID := rr.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Error("expected X-Request-ID header")
	}

	// Verify it's a valid UUID format
	if len(requestID) != 36 {
		t.Errorf("expected UUID format (36 chars), got %d: %s", len(requestID), requestID)
	}
}

func TestRequestID_PreExisting(t *testing.T) {
	// Test that existing request IDs are preserved
	existingID := "existing-request-id-12345"

	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())
		if requestID != existingID {
			t.Errorf("expected request ID %s, got %s", existingID, requestID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", existingID)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check that the existing ID is in the response
	if rr.Header().Get("X-Request-ID") != existingID {
		t.Errorf("expected X-Request-ID: %s, got %s", existingID, rr.Header().Get("X-Request-ID"))
	}
}

func TestGetRequestID(t *testing.T) {
	// Test with request ID in context
	ctx := context.WithValue(context.Background(), ContextKeyRequestID, "test-id-123")
	requestID := GetRequestID(ctx)
	if requestID != "test-id-123" {
		t.Errorf("expected 'test-id-123', got '%s'", requestID)
	}

	// Test with no request ID
	emptyCtx := context.Background()
	requestID = GetRequestID(emptyCtx)
	if requestID != "" {
		t.Errorf("expected empty string, got '%s'", requestID)
	}
}

func TestSecurityHeaders_API(t *testing.T) {
	handler := SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/snippets", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check API-specific headers
	if version := rr.Header().Get("X-API-Version"); version != APIVersion {
		t.Errorf("expected X-API-Version: %s, got %s", APIVersion, version)
	}

	// Check CSP for API routes (should be stricter)
	csp := rr.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "default-src 'none'") {
		t.Errorf("expected strict CSP for API routes, got: %s", csp)
	}

	// Check other security headers
	headers := map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"X-XSS-Protection":          "1; mode=block",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
	}

	for header, expected := range headers {
		if val := rr.Header().Get(header); !strings.Contains(val, expected) {
			t.Errorf("expected %s to contain '%s', got '%s'", header, expected, val)
		}
	}
}

func TestSecurityHeaders_Web(t *testing.T) {
	handler := SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Web routes should NOT have X-API-Version
	if version := rr.Header().Get("X-API-Version"); version != "" {
		t.Errorf("expected no X-API-Version for web routes, got %s", version)
	}

	// Check CSP for web routes (should allow scripts, styles, etc.)
	csp := rr.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "default-src 'self'") {
		t.Errorf("expected permissive CSP for web routes, got: %s", csp)
	}
	if !strings.Contains(csp, "script-src") {
		t.Error("expected script-src in web CSP")
	}
}

func TestSecurityHeaders_Permissions(t *testing.T) {
	handler := SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check Permissions-Policy header
	if pp := rr.Header().Get("Permissions-Policy"); pp == "" {
		t.Error("expected Permissions-Policy header")
	} else if !strings.Contains(pp, "camera=()") {
		t.Errorf("expected camera restriction in Permissions-Policy, got: %s", pp)
	}
}

func TestResponseWriter(t *testing.T) {
	// Test that our wrapped response writer captures status codes
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrapped, ok := w.(*responseWriter)
		if !ok {
			t.Fatal("expected responseWriter type")
		}

		// Default should be 200
		if wrapped.statusCode != http.StatusOK {
			t.Errorf("expected default status 200, got %d", wrapped.statusCode)
		}

		w.WriteHeader(http.StatusCreated)

		// After WriteHeader, should be 201
		if wrapped.statusCode != http.StatusCreated {
			t.Errorf("expected status 201, got %d", wrapped.statusCode)
		}
	})

	req := httptest.NewRequest("GET", "/test", nil)

	// Wrap with our responseWriter
	rr := httptest.NewRecorder()
	wrapped := &responseWriter{ResponseWriter: rr, statusCode: http.StatusOK}

	handler.ServeHTTP(wrapped, req)

	if wrapped.statusCode != http.StatusCreated {
		t.Errorf("expected final status 201, got %d", wrapped.statusCode)
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		xff        string
		xri        string
		trustProxy bool
		expected   string
	}{
		{
			name:       "direct connection",
			remoteAddr: "192.168.1.100:12345",
			trustProxy: false,
			expected:   "192.168.1.100",
		},
		{
			name:       "xff with trust",
			remoteAddr: "10.0.0.1:12345",
			xff:        "203.0.113.1, 198.51.100.1",
			trustProxy: true,
			expected:   "203.0.113.1",
		},
		{
			name:       "xff without trust",
			remoteAddr: "192.168.1.100:12345",
			xff:        "203.0.113.1",
			trustProxy: false,
			expected:   "192.168.1.100",
		},
		{
			name:       "x-real-ip with trust",
			remoteAddr: "10.0.0.1:12345",
			xri:        "203.0.113.5",
			trustProxy: true,
			expected:   "203.0.113.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set trust proxy setting
			oldTrust := TrustProxy
			TrustProxy = tt.trustProxy
			defer func() { TrustProxy = oldTrust }()

			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}
			if tt.xri != "" {
				req.Header.Set("X-Real-IP", tt.xri)
			}

			ip := getClientIP(req)
			if ip != tt.expected {
				t.Errorf("expected IP %s, got %s", tt.expected, ip)
			}
		})
	}
}
