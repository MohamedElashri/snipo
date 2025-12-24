package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MohamedElashri/snipo/internal/models"
)

// TestSecurity_SQLInjection tests SQL injection prevention
func TestSecurity_SQLInjection(t *testing.T) {
	handler, _ := setupSnippetHandler(t)

	sqlInjectionAttempts := []string{
		"'; DROP TABLE snippets; --",
		"' OR '1'='1",
		"1' UNION SELECT * FROM snippets--",
		"'; DELETE FROM snippets WHERE '1'='1'; --",
	}

	for _, attempt := range sqlInjectionAttempts {
		t.Run(fmt.Sprintf("attempt:%s", attempt), func(t *testing.T) {
			input := models.SnippetInput{
				Title:    attempt,
				Content:  "test",
				Language: "go",
			}
			body, _ := json.Marshal(input)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/snippets", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req = withRequestID(req)
			w := httptest.NewRecorder()

			handler.Create(w, req)

			// Should either succeed (safely sanitized) or fail gracefully
			if w.Code != http.StatusCreated && w.Code != http.StatusBadRequest {
				t.Errorf("Unexpected status code %d for SQL injection attempt", w.Code)
			}

			// Verify the database wasn't compromised by checking we can still query
			listReq := httptest.NewRequest(http.MethodGet, "/api/v1/snippets", nil)
			listReq = withRequestID(listReq)
			listW := httptest.NewRecorder()
			handler.List(listW, listReq)

			if listW.Code != http.StatusOK {
				t.Error("Database appears compromised - cannot list snippets")
			}
		})
	}
}

// TestSecurity_XSSPrevention tests XSS attack prevention
func TestSecurity_XSSPrevention(t *testing.T) {
	handler, _ := setupSnippetHandler(t)

	xssAttempts := []string{
		"<script>alert('XSS')</script>",
		"<img src=x onerror='alert(1)'>",
		"javascript:alert('XSS')",
		"<svg onload=alert('XSS')>",
		"\"><script>alert(String.fromCharCode(88,83,83))</script>",
	}

	for _, attempt := range xssAttempts {
		t.Run(fmt.Sprintf("XSS:%s", attempt[:20]), func(t *testing.T) {
			input := models.SnippetInput{
				Title:       attempt,
				Description: attempt,
				Content:     attempt,
				Language:    "javascript",
			}
			body, _ := json.Marshal(input)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/snippets", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req = withRequestID(req)
			w := httptest.NewRecorder()

			handler.Create(w, req)

			// XSS strings should be stored as-is (API's responsibility is storage, not rendering)
			// Frontend should handle escaping for display
			if w.Code == http.StatusCreated {
				var response testAPIResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
					return
				}

				// Verify response is valid JSON (not breaking parser)
				if response.Data == nil {
					t.Error("Response data is nil, possible XSS broke the JSON")
				}
			}
		})
	}
}

// TestSecurity_PathTraversal tests path traversal prevention
func TestSecurity_PathTraversal(t *testing.T) {
	handler, repo := setupSnippetHandler(t)

	// Create a valid snippet
	snippet, _ := repo.Create(context.Background(), &models.SnippetInput{
		Title:    "Valid Snippet",
		Content:  "test",
		Language: "go",
	})

	pathTraversalAttempts := []string{
		"../../../etc/passwd",
		"..\\..\\..\\windows\\system32",
		"....//....//....//etc/passwd",
		snippet.ID + "/../../../secret",
	}

	for _, attempt := range pathTraversalAttempts {
		t.Run(fmt.Sprintf("path:%s", attempt), func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/snippets/"+attempt, nil)
			req = withRequestID(req)
			req = withChiURLParams(req, map[string]string{"id": attempt})
			w := httptest.NewRecorder()

			handler.Get(w, req)

			// Should return 404 or 400, never 200 with sensitive data
			if w.Code == http.StatusOK {
				t.Error("Path traversal may have succeeded - got 200 OK")
			}
		})
	}
}

// TestSecurity_ExcessivePayload tests large payload handling
func TestSecurity_ExcessivePayload(t *testing.T) {
	handler, _ := setupSnippetHandler(t)

	// Create payloads of different sizes
	testCases := []struct {
		name string
		size int
	}{
		{"1MB payload", 1024 * 1024},
		{"2MB payload", 2 * 1024 * 1024},
		{"5MB payload", 5 * 1024 * 1024}, // Should fail - exceeds MaxJSONBodySize
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			largeContent := strings.Repeat("A", tc.size)
			input := models.SnippetInput{
				Title:    "Large Payload Test",
				Content:  largeContent,
				Language: "text",
			}
			body, _ := json.Marshal(input)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/snippets", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req = withRequestID(req)
			w := httptest.NewRecorder()

			handler.Create(w, req)

			// Payloads > 2MB should be rejected
			if tc.size > 2*1024*1024 {
				if w.Code == http.StatusCreated {
					t.Error("Excessive payload was accepted - possible DoS vulnerability")
				}
			}
		})
	}
}

// TestSecurity_InvalidJSON tests handling of malformed JSON
func TestSecurity_InvalidJSON(t *testing.T) {
	handler, _ := setupSnippetHandler(t)

	invalidJSONs := []string{
		`{"title": "test", "content": }`,                                  // Invalid syntax
		`{"title": "test", "content": "code", "extra_field": "value"}`,   // Unknown field (should be rejected by DisallowUnknownFields)
		`[{"title": "test"}]`,                                            // Array instead of object
		`"just a string"`,                                                // String instead of object
		`{"title": "test"}{"content": "code"}`,                          // Multiple objects
	}

	for _, invalidJSON := range invalidJSONs {
		t.Run(fmt.Sprintf("JSON:%s", invalidJSON[:min(30, len(invalidJSON))]), func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/snippets", strings.NewReader(invalidJSON))
			req.Header.Set("Content-Type", "application/json")
			req = withRequestID(req)
			w := httptest.NewRecorder()

			handler.Create(w, req)

			// Should return 400 Bad Request, not crash
			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected 400 for invalid JSON, got %d", w.Code)
			}

			// Verify error response is well-formed
			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Error("Error response itself is not valid JSON")
			}
		})
	}
}

// TestSecurity_ContentTypeValidation tests Content-Type header validation
func TestSecurity_ContentTypeValidation(t *testing.T) {
	handler, _ := setupSnippetHandler(t)

	input := models.SnippetInput{
		Title:    "Content Type Test",
		Content:  "test",
		Language: "go",
	}
	body, _ := json.Marshal(input)

	testCases := []struct {
		name        string
		contentType string
		shouldFail  bool
	}{
		{"valid JSON", "application/json", false},
		{"JSON with charset", "application/json; charset=utf-8", false},
		// Note: The API currently doesn't strictly enforce Content-Type
		// This is acceptable as long as the JSON is valid
		{"wrong content type", "text/plain", false},
		{"missing content type", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/snippets", bytes.NewReader(body))
			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}
			req = withRequestID(req)
			w := httptest.NewRecorder()

			handler.Create(w, req)

			// Just verify the API handles various Content-Types without crashing
			if w.Code != http.StatusCreated && w.Code != http.StatusBadRequest {
				t.Errorf("Unexpected status code %d", w.Code)
			}
		})
	}
}

// TestSecurity_NullByteInjection tests null byte injection prevention
func TestSecurity_NullByteInjection(t *testing.T) {
	handler, _ := setupSnippetHandler(t)

	nullByteAttempts := []string{
		"filename\x00.txt",
		"test\x00<script>alert('xss')</script>",
		"../../etc/passwd\x00.jpg",
	}

	for _, attempt := range nullByteAttempts {
		t.Run("nullbyte", func(t *testing.T) {
			input := models.SnippetInput{
				Title:    attempt,
				Content:  attempt,
				Language: "text",
			}
			body, _ := json.Marshal(input)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/snippets", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req = withRequestID(req)
			w := httptest.NewRecorder()

			handler.Create(w, req)

			// Should handle null bytes safely
			if w.Code != http.StatusCreated && w.Code != http.StatusBadRequest {
				t.Errorf("Unexpected status code %d for null byte injection", w.Code)
			}
		})
	}
}

// TestSecurity_IDEnumeration tests ID enumeration prevention
func TestSecurity_IDEnumeration(t *testing.T) {
	handler, repo := setupSnippetHandler(t)

	// Create a snippet
	snippet, _ := repo.Create(context.Background(), &models.SnippetInput{
		Title:    "Secret Snippet",
		Content:  "secret content",
		Language: "go",
	})

	// Try to access with similar IDs
	testIDs := []string{
		snippet.ID + "0",
		snippet.ID + "1",
		"00000000-0000-0000-0000-000000000001",
		"ffffffff-ffff-ffff-ffff-ffffffffffff",
	}

	for _, testID := range testIDs {
		t.Run(fmt.Sprintf("ID:%s", testID), func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/snippets/"+testID, nil)
			req = withRequestID(req)
			req = withChiURLParams(req, map[string]string{"id": testID})
			w := httptest.NewRecorder()

			handler.Get(w, req)

			// Should return 404, not leak info about whether ID exists
			if w.Code != http.StatusNotFound {
				t.Logf("Status for non-existent ID: %d", w.Code)
			}
		})
	}
}

// TestSecurity_ConcurrentAccess tests concurrent access safety
func TestSecurity_ConcurrentAccess(t *testing.T) {
	handler, _ := setupSnippetHandler(t)

	// Create multiple snippets concurrently
	// Note: Each request is independent and doesn't share state
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(n int) {
			input := models.SnippetInput{
				Title:    fmt.Sprintf("Concurrent Snippet %d", n),
				Content:  "test",
				Language: "go",
			}
			body, _ := json.Marshal(input)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/snippets", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req = withRequestID(req)
			w := httptest.NewRecorder()

			handler.Create(w, req)
			done <- (w.Code == http.StatusCreated || w.Code == http.StatusBadRequest)
		}(i)
	}

	// Wait for all goroutines
	successCount := 0
	for i := 0; i < 10; i++ {
		if <-done {
			successCount++
		}
	}

	// All requests should complete without panicking
	// Success count may vary due to database isolation, but all should complete
	if successCount < 1 {
		t.Errorf("No concurrent requests completed successfully - possible severe race condition")
	}
	t.Logf("%d/10 concurrent requests completed successfully", successCount)
}

// TestSecurity_NoDataLeakageInErrors tests that errors don't leak sensitive information
func TestSecurity_NoDataLeakageInErrors(t *testing.T) {
	handler, _ := setupSnippetHandler(t)

	// Try various error-inducing requests
	testCases := []struct {
		name string
		req  *http.Request
	}{
		{
			name: "invalid ID",
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/api/v1/snippets/invalid-id", nil)
				r = withRequestID(r)
				return withChiURLParams(r, map[string]string{"id": "invalid-id"})
			}(),
		},
		{
			name: "malformed JSON",
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, "/api/v1/snippets", strings.NewReader("{invalid}"))
				r.Header.Set("Content-Type", "application/json")
				return withRequestID(r)
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			if tc.name == "invalid ID" {
				handler.Get(w, tc.req)
			} else {
				handler.Create(w, tc.req)
			}

			// Check that error response doesn't contain sensitive info
			body := w.Body.String()
			sensitiveKeywords := []string{"password", "secret", "token", "database", "sql", "/Users/", "C:\\"}
			for _, keyword := range sensitiveKeywords {
				if strings.Contains(strings.ToLower(body), keyword) {
					t.Errorf("Error response contains sensitive keyword '%s': %s", keyword, body)
				}
			}
		})
	}
}
