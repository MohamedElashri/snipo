package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MohamedElashri/snipo/internal/models"
)

// TestContract_SnippetCreateResponse validates the create snippet response structure
func TestContract_SnippetCreateResponse(t *testing.T) {
	handler, _ := setupSnippetHandler(t)

	input := models.SnippetInput{
		Title:    "Contract Test Snippet",
		Content:  "fmt.Println(\"test\")",
		Language: "go",
	}
	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/snippets", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withRequestID(req)
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	// Parse response
	var response testAPIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Validate envelope structure
	if response.Data == nil {
		t.Fatal("Expected data field to be present")
	}

	// Validate snippet data
	dataBytes, _ := json.Marshal(response.Data)
	var snippet models.Snippet
	if err := json.Unmarshal(dataBytes, &snippet); err != nil {
		t.Fatalf("Failed to unmarshal snippet data: %v", err)
	}

	if snippet.ID == "" {
		t.Error("Expected snippet ID to be present")
	}
	if snippet.Title != input.Title {
		t.Errorf("Expected title '%s', got '%s'", input.Title, snippet.Title)
	}
	if snippet.Content != input.Content {
		t.Errorf("Expected content '%s', got '%s'", input.Content, snippet.Content)
	}
	if snippet.Language != input.Language {
		t.Errorf("Expected language '%s', got '%s'", input.Language, snippet.Language)
	}
	if snippet.CreatedAt.IsZero() {
		t.Error("Expected created_at timestamp to be present")
	}
	if snippet.UpdatedAt.IsZero() {
		t.Error("Expected updated_at timestamp to be present")
	}
}

// TestContract_SnippetListResponse validates the list snippets response structure
func TestContract_SnippetListResponse(t *testing.T) {
	handler, repo := setupSnippetHandler(t)

	// Create a test snippet first
	_, err := repo.Create(context.Background(), &models.SnippetInput{
		Title:    "List Test",
		Content:  "test",
		Language: "go",
	})
	if err != nil {
		t.Fatalf("Failed to create test snippet: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/snippets?page=1&limit=10", nil)
	req = withRequestID(req)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Parse response
	var response testListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Validate envelope structure
	if response.Data == nil {
		t.Fatal("Expected data array to be present")
	}

	// Validate pagination structure
	if response.Pagination == nil {
		t.Fatal("Expected pagination field to be present")
	}
	if response.Pagination.Page <= 0 {
		t.Error("Expected page number to be positive")
	}
	if response.Pagination.Limit <= 0 {
		t.Error("Expected limit to be positive")
	}
	if response.Pagination.Total < 0 {
		t.Error("Expected total to be non-negative")
	}
}

// TestContract_ErrorResponse validates error response structure
func TestContract_ErrorResponse(t *testing.T) {
	handler, _ := setupSnippetHandler(t)

	tests := []struct {
		name           string
		method         string
		path           string
		body           map[string]interface{}
		expectedStatus int
		checkFields    []string
	}{
		{
			name:           "validation error",
			method:         http.MethodPost,
			path:           "/api/v1/snippets",
			body:           map[string]interface{}{"title": ""},
			expectedStatus: http.StatusBadRequest,
			checkFields:    []string{"error", "code", "message"},
		},
		{
			name:           "not found error",
			method:         http.MethodGet,
			path:           "/api/v1/snippets/nonexistent",
			body:           nil,
			expectedStatus: http.StatusNotFound,
			checkFields:    []string{"error", "code", "message"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != nil {
				body, _ := json.Marshal(tt.body)
				req = httptest.NewRequest(tt.method, tt.path, bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}
			req = withRequestID(req)

			if tt.name == "not found error" {
				req = withChiURLParams(req, map[string]string{"id": "nonexistent"})
			}

			w := httptest.NewRecorder()

			switch tt.method {
			case http.MethodPost:
				handler.Create(w, req)
			case http.MethodGet:
				handler.Get(w, req)
			}

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Parse error response
			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to parse error response: %v", err)
			}

			// Validate that error field exists and is an object
			errorObj, ok := response["error"]
			if !ok {
				t.Fatal("Expected error response to contain 'error' field")
			}

			errorDetail, ok := errorObj.(map[string]interface{})
			if !ok {
				t.Fatal("Expected 'error' field to be an object")
			}

			// Validate required error fields are within the error object
			for _, field := range tt.checkFields {
				if field == "error" {
					continue // Already checked above
				}
				if _, ok := errorDetail[field]; !ok {
					t.Errorf("Expected error object to contain field '%s'", field)
				}
			}
		})
	}
}

// TestContract_ContentTypeHeaders validates content type headers
func TestContract_ContentTypeHeaders(t *testing.T) {
	handler, repo := setupSnippetHandler(t)

	// Create a test snippet
	snippet, _ := repo.Create(context.Background(), &models.SnippetInput{
		Title:    "Header Test",
		Content:  "test",
		Language: "go",
	})

	tests := []struct {
		name               string
		method             string
		path               string
		body               interface{}
		expectedStatusCode int
		expectedContentType string
	}{
		{
			name:                "list snippets returns JSON",
			method:              http.MethodGet,
			path:                "/api/v1/snippets",
			expectedStatusCode:  http.StatusOK,
			expectedContentType: "application/json",
		},
		{
			name:                "get snippet returns JSON",
			method:              http.MethodGet,
			path:                "/api/v1/snippets/" + snippet.ID,
			expectedStatusCode:  http.StatusOK,
			expectedContentType: "application/json",
		},
		{
			name:                "create snippet returns JSON",
			method:              http.MethodPost,
			path:                "/api/v1/snippets",
			body:                models.SnippetInput{Title: "Test", Content: "test", Language: "go"},
			expectedStatusCode:  http.StatusCreated,
			expectedContentType: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != nil {
				body, _ := json.Marshal(tt.body)
				req = httptest.NewRequest(tt.method, tt.path, bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}
			req = withRequestID(req)

			if tt.method == http.MethodGet && tt.path != "/api/v1/snippets" {
				req = withChiURLParams(req, map[string]string{"id": snippet.ID})
			}

			w := httptest.NewRecorder()

			switch tt.method {
			case http.MethodGet:
				if tt.path == "/api/v1/snippets" {
					handler.List(w, req)
				} else {
					handler.Get(w, req)
				}
			case http.MethodPost:
				handler.Create(w, req)
			}

			if w.Code != tt.expectedStatusCode {
				t.Errorf("Expected status %d, got %d", tt.expectedStatusCode, w.Code)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != tt.expectedContentType {
				t.Errorf("Expected Content-Type '%s', got '%s'", tt.expectedContentType, contentType)
			}
		})
	}
}

// TestContract_FieldTypes validates that all JSON fields have correct types
func TestContract_FieldTypes(t *testing.T) {
	handler, _ := setupSnippetHandler(t)

	input := models.SnippetInput{
		Title:      "Type Test",
		Content:    "test content",
		Language:   "go",
		IsPublic:   true,
		IsArchived: false,
		Tags:       []string{"test"},
	}
	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/snippets", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withRequestID(req)
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	// Parse as generic map to check types
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be an object")
	}

	// Validate field types
	typeChecks := map[string]string{
		"id":          "string",
		"title":       "string",
		"content":     "string",
		"language":    "string",
		"is_favorite": "bool",
		"is_public":   "bool",
		"is_archived": "bool",
		"view_count":  "number",
		"created_at":  "string",
		"updated_at":  "string",
	}

	for field, expectedType := range typeChecks {
		value, exists := data[field]
		if !exists {
			t.Errorf("Expected field '%s' to exist in response", field)
			continue
		}

		var actualType string
		switch value.(type) {
		case string:
			actualType = "string"
		case bool:
			actualType = "bool"
		case float64:
			actualType = "number"
		default:
			actualType = "unknown"
		}

		if actualType != expectedType {
			t.Errorf("Field '%s': expected type %s, got %s", field, expectedType, actualType)
		}
	}
}
