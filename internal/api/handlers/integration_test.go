package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MohamedElashri/snipo/internal/models"
	"github.com/MohamedElashri/snipo/internal/repository"
	"github.com/MohamedElashri/snipo/internal/services"
	"github.com/MohamedElashri/snipo/internal/testutil"
)

// TestIntegration_SnippetCRUDFlow tests the complete snippet lifecycle
func TestIntegration_SnippetCRUDFlow(t *testing.T) {
	handler, snippetRepo := setupSnippetHandler(t)

	// Step 1: Create a snippet
	t.Log("Creating snippet...")
	createInput := map[string]interface{}{
		"title":       "Integration Test Snippet",
		"content":     "package main\n\nfunc main() {}\n",
		"language":    "go",
		"description": "A test snippet for integration testing",
	}
	createBody, _ := json.Marshal(createInput)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/snippets", bytes.NewReader(createBody))
	req = withRequestID(req)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Create failed: expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	var createEnvelope testAPIResponse
	json.Unmarshal(w.Body.Bytes(), &createEnvelope)
	dataBytes, _ := json.Marshal(createEnvelope.Data)
	var createdSnippet models.Snippet
	json.Unmarshal(dataBytes, &createdSnippet)
	snippetID := createdSnippet.ID

	// Verify creation
	if createdSnippet.Title != "Integration Test Snippet" {
		t.Errorf("Expected title 'Integration Test Snippet', got '%s'", createdSnippet.Title)
	}

	// Step 2: Get the snippet
	t.Log("Fetching snippet...")
	req = httptest.NewRequest(http.MethodGet, "/api/v1/snippets/"+snippetID, nil)
	req = withRequestID(req)
	req = withChiURLParams(req, map[string]string{"id": snippetID})
	w = httptest.NewRecorder()

	handler.Get(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Get failed: expected status %d, got %d", http.StatusOK, w.Code)
	}

	var getEnvelope testAPIResponse
	json.Unmarshal(w.Body.Bytes(), &getEnvelope)
	dataBytes, _ = json.Marshal(getEnvelope.Data)
	var fetchedSnippet models.Snippet
	json.Unmarshal(dataBytes, &fetchedSnippet)

	if fetchedSnippet.ID != snippetID {
		t.Errorf("Expected ID %s, got %s", snippetID, fetchedSnippet.ID)
	}

	// Step 3: Update the snippet
	t.Log("Updating snippet...")
	updateInput := map[string]interface{}{
		"title":   "Updated Integration Test",
		"content": "package main\n\nfunc main() {\n\tprintln(\"updated\")\n}\n",
	}
	updateBody, _ := json.Marshal(updateInput)

	req = httptest.NewRequest(http.MethodPut, "/api/v1/snippets/"+snippetID, bytes.NewReader(updateBody))
	req = withRequestID(req)
	req = withChiURLParams(req, map[string]string{"id": snippetID})
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	handler.Update(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Update failed: expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	// Verify update
	updatedSnippet, err := snippetRepo.GetByID(context.Background(), snippetID)
	if err != nil {
		t.Fatalf("Failed to fetch updated snippet: %v", err)
	}
	if updatedSnippet.Title != "Updated Integration Test" {
		t.Errorf("Expected title 'Updated Integration Test', got '%s'", updatedSnippet.Title)
	}

	// Step 4: List snippets
	t.Log("Listing snippets...")
	req = httptest.NewRequest(http.MethodGet, "/api/v1/snippets?page=1&limit=10", nil)
	req = withRequestID(req)
	w = httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("List failed: expected status %d, got %d", http.StatusOK, w.Code)
	}

	var listEnvelope testListResponse
	json.Unmarshal(w.Body.Bytes(), &listEnvelope)

	// Verify pagination
	if listEnvelope.Pagination == nil {
		t.Fatal("Expected pagination to be present")
	}
	if listEnvelope.Pagination.Page != 1 {
		t.Errorf("Expected page 1, got %d", listEnvelope.Pagination.Page)
	}
	if listEnvelope.Pagination.Total < 1 {
		t.Error("Expected at least 1 snippet in total")
	}

	// Step 5: Delete the snippet
	t.Log("Deleting snippet...")
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/snippets/"+snippetID, nil)
	req = withRequestID(req)
	req = withChiURLParams(req, map[string]string{"id": snippetID})
	w = httptest.NewRecorder()

	handler.Delete(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("Delete failed: expected status %d, got %d", http.StatusNoContent, w.Code)
	}

	// Verify deletion
	deletedSnippet, err := snippetRepo.GetByID(context.Background(), snippetID)
	if err != nil {
		t.Fatalf("Error checking if snippet was deleted: %v", err)
	}
	if deletedSnippet != nil {
		t.Error("Expected snippet to be deleted, but it still exists")
	}

	t.Log("Integration test completed successfully!")
}

// TestIntegration_ValidationErrors tests validation error responses
func TestIntegration_ValidationErrors(t *testing.T) {
	handler, _ := setupSnippetHandler(t)

	tests := []struct {
		name         string
		input        map[string]interface{}
		expectedCode int
		errorField   string
	}{
		{
			name:         "empty title",
			input:        map[string]interface{}{"title": "", "content": "test"},
			expectedCode: http.StatusBadRequest,
			errorField:   "title",
		},
		{
			name:         "missing title",
			input:        map[string]interface{}{"content": "test"},
			expectedCode: http.StatusBadRequest,
			errorField:   "title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/snippets", bytes.NewReader(body))
			req = withRequestID(req)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Create(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Expected status %d, got %d: %s", tt.expectedCode, w.Code, w.Body.String())
			}

			// Verify error response contains field name
			if w.Body.Len() > 0 {
				bodyStr := w.Body.String()
				if tt.errorField != "" && !contains(bodyStr, tt.errorField) {
					t.Errorf("Expected error message to contain '%s', got: %s", tt.errorField, bodyStr)
				}
			}
		})
	}
}

// TestIntegration_TagsAndFolders tests tag and folder functionality
func TestIntegration_TagsAndFolders(t *testing.T) {
	db := testutil.TestDB(t)
	snippetRepo := repository.NewSnippetRepository(db)
	tagRepo := repository.NewTagRepository(db)
	folderRepo := repository.NewFolderRepository(db)
	fileRepo := repository.NewSnippetFileRepository(db)
	logger := testutil.TestLogger()

	service := services.NewSnippetService(snippetRepo, logger).
		WithTagRepo(tagRepo).
		WithFolderRepo(folderRepo).
		WithFileRepo(fileRepo).
		WithMaxFiles(10)

	snippetHandler := NewSnippetHandler(service)
	tagHandler := NewTagHandler(tagRepo)
	folderHandler := NewFolderHandler(folderRepo)

	// Create a tag
	t.Log("Creating tag...")
	tagInput := map[string]interface{}{"name": "golang", "color": "#00ADD8"}
	tagBody, _ := json.Marshal(tagInput)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tags", bytes.NewReader(tagBody))
	req = withRequestID(req)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	tagHandler.Create(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Tag creation failed: %d: %s", w.Code, w.Body.String())
	}

	// Create a folder
	t.Log("Creating folder...")
	folderInput := map[string]interface{}{"name": "Work Projects"}
	folderBody, _ := json.Marshal(folderInput)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/folders", bytes.NewReader(folderBody))
	req = withRequestID(req)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	folderHandler.Create(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Folder creation failed: %d: %s", w.Code, w.Body.String())
	}

	var folderEnvelope testAPIResponse
	json.Unmarshal(w.Body.Bytes(), &folderEnvelope)
	dataBytes, _ := json.Marshal(folderEnvelope.Data)
	var folder models.Folder
	json.Unmarshal(dataBytes, &folder)

	// Create snippet with tag and folder
	t.Log("Creating snippet with tag and folder...")
	snippetInput := map[string]interface{}{
		"title":     "Go HTTP Server",
		"content":   "package main",
		"language":  "go",
		"tags":      []string{"golang"},
		"folder_id": folder.ID,
	}
	snippetBody, _ := json.Marshal(snippetInput)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/snippets", bytes.NewReader(snippetBody))
	req = withRequestID(req)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	snippetHandler.Create(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Snippet creation failed: %d: %s", w.Code, w.Body.String())
	}

	var snippetEnvelope testAPIResponse
	json.Unmarshal(w.Body.Bytes(), &snippetEnvelope)
	dataBytes, _ = json.Marshal(snippetEnvelope.Data)
	var snippet models.Snippet
	json.Unmarshal(dataBytes, &snippet)

	// Verify tags
	if len(snippet.Tags) != 1 || snippet.Tags[0].Name != "golang" {
		t.Errorf("Expected tag 'golang', got: %v", snippet.Tags)
	}

	// Verify folder
	if len(snippet.Folders) != 1 || snippet.Folders[0].ID != folder.ID {
		t.Errorf("Expected folder ID %d, got folders: %v", folder.ID, snippet.Folders)
	}

	t.Log("Tags and folders integration test completed!")
}

// TestIntegration_SearchFunctionality tests search
func TestIntegration_SearchFunctionality(t *testing.T) {
	handler, _ := setupSnippetHandler(t)

	// Create test snippets
	testSnippets := []map[string]interface{}{
		{"title": "Go HTTP Server", "content": "package main\nimport \"net/http\""},
		{"title": "Python Flask App", "content": "from flask import Flask"},
		{"title": "JavaScript React", "content": "import React from 'react'"},
	}

	for _, input := range testSnippets {
		body, _ := json.Marshal(input)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/snippets", bytes.NewReader(body))
		req = withRequestID(req)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.Create(w, req)
	}

	// Search for "Go"
	t.Log("Searching for 'Go'...")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/snippets/search?q=Go", nil)
	req = withRequestID(req)
	w := httptest.NewRecorder()

	handler.Search(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Search failed: %d", w.Code)
	}

	var searchEnvelope testListResponse
	json.Unmarshal(w.Body.Bytes(), &searchEnvelope)
	dataBytes, _ := json.Marshal(searchEnvelope.Data)
	var results []models.Snippet
	json.Unmarshal(dataBytes, &results)

	if len(results) == 0 {
		t.Error("Expected search results, got none")
	}

	// Verify result contains "Go"
	found := false
	for _, snippet := range results {
		if contains(snippet.Title, "Go") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Search results don't contain expected snippet with 'Go'")
	}

	t.Log("Search integration test completed!")
}

// Helper function
func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
