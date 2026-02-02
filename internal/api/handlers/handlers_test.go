package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/MohamedElashri/snipo/internal/api/middleware"
	"github.com/MohamedElashri/snipo/internal/config"
	"github.com/MohamedElashri/snipo/internal/models"
	"github.com/MohamedElashri/snipo/internal/repository"
	"github.com/MohamedElashri/snipo/internal/services"
	"github.com/MohamedElashri/snipo/internal/testutil"
)

// Test response wrapper structs for new API envelope format
type testAPIResponse struct {
	Data interface{} `json:"data"`
	Meta *testMeta   `json:"meta,omitempty"`
}

type testMeta struct {
	RequestID string    `json:"request_id"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

type testListResponse struct {
	Data       interface{}     `json:"data"`
	Pagination *testPagination `json:"pagination,omitempty"`
	Meta       *testMeta       `json:"meta,omitempty"`
}

type testPagination struct {
	Page       int                  `json:"page"`
	Limit      int                  `json:"limit"`
	Total      int                  `json:"total"`
	TotalPages int                  `json:"totalPages"`
	Links      *testPaginationLinks `json:"links,omitempty"`
}

type testPaginationLinks struct {
	Self string  `json:"self"`
	Next *string `json:"next"`
	Prev *string `json:"prev"`
}

// withChiURLParams adds chi URL params to a request context
func withChiURLParams(r *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for key, val := range params {
		rctx.URLParams.Add(key, val)
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

// withRequestID adds a request ID to the context for testing
func withRequestID(r *http.Request) *http.Request {
	requestID := uuid.New().String()
	ctx := context.WithValue(r.Context(), middleware.ContextKeyRequestID, requestID)
	return r.WithContext(ctx)
}

// setupSnippetHandler creates a snippet handler with test database
func setupSnippetHandler(t *testing.T) (*SnippetHandler, *repository.SnippetRepository) {
	t.Helper()
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

	return NewSnippetHandler(service), snippetRepo
}

func TestSnippetHandler_Create(t *testing.T) {
	handler, _ := setupSnippetHandler(t)

	input := map[string]interface{}{
		"title":    "Test Snippet",
		"content":  "console.log('hello');",
		"language": "javascript",
	}
	body, _ := json.Marshal(input)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/snippets", bytes.NewReader(body))
	req = withRequestID(req)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	var envelope testAPIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Check meta fields
	if envelope.Meta == nil {
		t.Fatal("expected meta to be present")
	}
	if envelope.Meta.RequestID == "" {
		t.Error("expected request_id to be set")
	}
	if envelope.Meta.Version != "1.0" {
		t.Errorf("expected version '1.0', got %q", envelope.Meta.Version)
	}

	// Unmarshal data field
	dataBytes, _ := json.Marshal(envelope.Data)
	var response models.Snippet
	if err := json.Unmarshal(dataBytes, &response); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	if response.ID == "" {
		t.Error("expected snippet ID to be set")
	}
	if response.Title != "Test Snippet" {
		t.Errorf("expected title 'Test Snippet', got %q", response.Title)
	}
}

func TestSnippetHandler_Create_InvalidJSON(t *testing.T) {
	handler, _ := setupSnippetHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/snippets", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestSnippetHandler_Create_ValidationError(t *testing.T) {
	handler, _ := setupSnippetHandler(t)

	input := map[string]interface{}{
		"title":    "", // Empty title should fail validation
		"content":  "content",
		"language": "plaintext",
	}
	body, _ := json.Marshal(input)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/snippets", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d: %s", http.StatusBadRequest, w.Code, w.Body.String())
	}
}

func TestSnippetHandler_Get(t *testing.T) {
	handler, repo := setupSnippetHandler(t)
	ctx := testutil.TestContext()

	// Create a snippet first
	snippet, err := repo.Create(ctx, &models.SnippetInput{
		Title:    "Test Snippet",
		Content:  "content",
		Language: "plaintext",
	})
	if err != nil {
		t.Fatalf("failed to create snippet: %v", err)
	}

	// Create request with chi URL param
	req := httptest.NewRequest(http.MethodGet, "/api/v1/snippets/"+snippet.ID, nil)
	req = withChiURLParams(req, map[string]string{"id": snippet.ID})
	req = withRequestID(req)

	w := httptest.NewRecorder()
	handler.Get(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var envelope testAPIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Unmarshal data field
	dataBytes, _ := json.Marshal(envelope.Data)
	var response models.Snippet
	if err := json.Unmarshal(dataBytes, &response); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	if response.ID != snippet.ID {
		t.Errorf("expected ID %q, got %q", snippet.ID, response.ID)
	}
}

func TestSnippetHandler_Get_NotFound(t *testing.T) {
	handler, _ := setupSnippetHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/snippets/nonexistent", nil)
	req = withChiURLParams(req, map[string]string{"id": "nonexistent"})

	w := httptest.NewRecorder()
	handler.Get(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestSnippetHandler_List(t *testing.T) {
	handler, repo := setupSnippetHandler(t)
	ctx := testutil.TestContext()

	// Create some snippets
	for i := 0; i < 3; i++ {
		_, err := repo.Create(ctx, &models.SnippetInput{
			Title:    "Snippet",
			Content:  "content",
			Language: "plaintext",
		})
		if err != nil {
			t.Fatalf("failed to create snippet: %v", err)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/snippets", nil)
	req = withRequestID(req)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var envelope testListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Check pagination
	if envelope.Pagination == nil {
		t.Fatal("expected pagination to be present")
	}
	if envelope.Pagination.Total != 3 {
		t.Errorf("expected total 3, got %d", envelope.Pagination.Total)
	}

	// Unmarshal data field
	dataBytes, _ := json.Marshal(envelope.Data)
	var snippets []models.Snippet
	if err := json.Unmarshal(dataBytes, &snippets); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	if len(snippets) != 3 {
		t.Errorf("expected 3 snippets, got %d", len(snippets))
	}
}

func TestSnippetHandler_List_WithPagination(t *testing.T) {
	handler, repo := setupSnippetHandler(t)
	ctx := testutil.TestContext()

	// Create 10 snippets
	for i := 0; i < 10; i++ {
		_, err := repo.Create(ctx, &models.SnippetInput{
			Title:    "Snippet",
			Content:  "content",
			Language: "plaintext",
		})
		if err != nil {
			t.Fatalf("failed to create snippet: %v", err)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/snippets?page=1&limit=3", nil)
	req = withRequestID(req)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var envelope testListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Check pagination metadata
	if envelope.Pagination == nil {
		t.Fatal("expected pagination to be present")
	}
	if envelope.Pagination.Total != 10 {
		t.Errorf("expected total 10, got %d", envelope.Pagination.Total)
	}
	if envelope.Pagination.TotalPages != 4 {
		t.Errorf("expected 4 total pages, got %d", envelope.Pagination.TotalPages)
	}
	if envelope.Pagination.Page != 1 {
		t.Errorf("expected page 1, got %d", envelope.Pagination.Page)
	}
	if envelope.Pagination.Limit != 3 {
		t.Errorf("expected limit 3, got %d", envelope.Pagination.Limit)
	}

	// Check navigation links
	if envelope.Pagination.Links == nil {
		t.Fatal("expected pagination links to be present")
	}
	if envelope.Pagination.Links.Next == nil {
		t.Error("expected next link to be present")
	}
	if envelope.Pagination.Links.Prev != nil {
		t.Error("expected prev link to be nil on first page")
	}

	// Unmarshal data field
	dataBytes, _ := json.Marshal(envelope.Data)
	var snippets []models.Snippet
	if err := json.Unmarshal(dataBytes, &snippets); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	if len(snippets) != 3 {
		t.Errorf("expected 3 snippets on page, got %d", len(snippets))
	}
}

func TestSnippetHandler_Update(t *testing.T) {
	handler, repo := setupSnippetHandler(t)
	ctx := testutil.TestContext()

	// Create a snippet
	snippet, err := repo.Create(ctx, &models.SnippetInput{
		Title:    "Original",
		Content:  "original content",
		Language: "plaintext",
	})
	if err != nil {
		t.Fatalf("failed to create snippet: %v", err)
	}

	// Update it
	input := map[string]interface{}{
		"title":    "Updated",
		"content":  "updated content",
		"language": "go",
	}
	body, _ := json.Marshal(input)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/snippets/"+snippet.ID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withChiURLParams(req, map[string]string{"id": snippet.ID})
	req = withRequestID(req)

	w := httptest.NewRecorder()
	handler.Update(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var envelope testAPIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Unmarshal data field
	dataBytes, _ := json.Marshal(envelope.Data)
	var response models.Snippet
	if err := json.Unmarshal(dataBytes, &response); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	if response.Title != "Updated" {
		t.Errorf("expected title 'Updated', got %q", response.Title)
	}
}

func TestSnippetHandler_Delete(t *testing.T) {
	handler, repo := setupSnippetHandler(t)
	ctx := testutil.TestContext()

	// Create a snippet
	snippet, err := repo.Create(ctx, &models.SnippetInput{
		Title:    "To Delete",
		Content:  "content",
		Language: "plaintext",
	})
	if err != nil {
		t.Fatalf("failed to create snippet: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/snippets/"+snippet.ID, nil)
	req = withChiURLParams(req, map[string]string{"id": snippet.ID})

	w := httptest.NewRecorder()
	handler.Delete(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}

	// Verify deleted (soft delete)
	deleted, _ := repo.GetByID(ctx, snippet.ID)
	if deleted == nil {
		t.Error("expected snippet to still exist (soft deleted)")
	} else if deleted.DeletedAt == nil {
		t.Error("expected snippet to be marked as deleted")
	}
}

func TestSnippetHandler_ToggleFavorite(t *testing.T) {
	handler, repo := setupSnippetHandler(t)
	ctx := testutil.TestContext()

	// Create a snippet
	snippet, err := repo.Create(ctx, &models.SnippetInput{
		Title:    "Test",
		Content:  "content",
		Language: "plaintext",
	})
	if err != nil {
		t.Fatalf("failed to create snippet: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/snippets/"+snippet.ID+"/favorite", nil)
	req = withChiURLParams(req, map[string]string{"id": snippet.ID})
	req = withRequestID(req)

	w := httptest.NewRecorder()
	handler.ToggleFavorite(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var envelope testAPIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Unmarshal data field
	dataBytes, _ := json.Marshal(envelope.Data)
	var response models.Snippet
	if err := json.Unmarshal(dataBytes, &response); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	if !response.IsFavorite {
		t.Error("expected is_favorite to be true after toggle")
	}
}

func TestSnippetHandler_Search(t *testing.T) {
	handler, repo := setupSnippetHandler(t)
	ctx := testutil.TestContext()

	// Create snippets with searchable content
	_, err := repo.Create(ctx, &models.SnippetInput{
		Title:    "Hello World",
		Content:  "print('hello')",
		Language: "python",
	})
	if err != nil {
		t.Fatalf("failed to create snippet: %v", err)
	}

	_, err = repo.Create(ctx, &models.SnippetInput{
		Title:    "Goodbye World",
		Content:  "console.log('bye')",
		Language: "javascript",
	})
	if err != nil {
		t.Fatalf("failed to create snippet: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/snippets/search?q=hello", nil)
	req = withRequestID(req)
	w := httptest.NewRecorder()

	handler.Search(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var envelope testAPIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Unmarshal data field
	dataBytes, _ := json.Marshal(envelope.Data)
	var snippets []models.Snippet
	if err := json.Unmarshal(dataBytes, &snippets); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	if len(snippets) != 1 {
		t.Errorf("expected 1 result for 'hello', got %d", len(snippets))
	}
}

// Tag Handler Tests

func setupTagHandler(t *testing.T) (*TagHandler, *repository.TagRepository) {
	t.Helper()
	db := testutil.TestDB(t)
	repo := repository.NewTagRepository(db)
	return NewTagHandler(repo), repo
}

func TestTagHandler_Create(t *testing.T) {
	handler, _ := setupTagHandler(t)

	input := map[string]interface{}{
		"name":  "golang",
		"color": "#00ADD8",
	}
	body, _ := json.Marshal(input)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tags", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withRequestID(req)
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	var envelope testAPIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Unmarshal data field
	dataBytes, _ := json.Marshal(envelope.Data)
	var response models.Tag
	if err := json.Unmarshal(dataBytes, &response); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	if response.Name != "golang" {
		t.Errorf("expected name 'golang', got %q", response.Name)
	}
}

func TestTagHandler_List(t *testing.T) {
	handler, repo := setupTagHandler(t)
	ctx := testutil.TestContext()

	// Create tags
	for _, name := range []string{"alpha", "beta", "gamma"} {
		_, err := repo.Create(ctx, &models.TagInput{Name: name, Color: "#000000"})
		if err != nil {
			t.Fatalf("failed to create tag: %v", err)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tags", nil)
	req = withRequestID(req)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var envelope testAPIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Unmarshal data field
	dataBytes, _ := json.Marshal(envelope.Data)
	var tags []models.Tag
	if err := json.Unmarshal(dataBytes, &tags); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	if len(tags) != 3 {
		t.Errorf("expected 3 tags, got %d", len(tags))
	}
}

// Folder Handler Tests

func setupFolderHandler(t *testing.T) (*FolderHandler, *repository.FolderRepository) {
	t.Helper()
	db := testutil.TestDB(t)
	repo := repository.NewFolderRepository(db)
	return NewFolderHandler(repo), repo
}

func TestFolderHandler_Create(t *testing.T) {
	handler, _ := setupFolderHandler(t)

	input := map[string]interface{}{
		"name": "Projects",
	}
	body, _ := json.Marshal(input)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/folders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withRequestID(req)
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	var envelope testAPIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Unmarshal data field
	dataBytes, _ := json.Marshal(envelope.Data)
	var response models.Folder
	if err := json.Unmarshal(dataBytes, &response); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	if response.Name != "Projects" {
		t.Errorf("expected name 'Projects', got %q", response.Name)
	}
}

func TestFolderHandler_List(t *testing.T) {
	handler, repo := setupFolderHandler(t)
	ctx := testutil.TestContext()

	// Create folders
	for _, name := range []string{"Alpha", "Beta", "Gamma"} {
		_, err := repo.Create(ctx, &models.FolderInput{Name: name})
		if err != nil {
			t.Fatalf("failed to create folder: %v", err)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/folders", nil)
	req = withRequestID(req)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var envelope testAPIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Unmarshal data field
	dataBytes, _ := json.Marshal(envelope.Data)
	var folders []models.Folder
	if err := json.Unmarshal(dataBytes, &folders); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	if len(folders) != 3 {
		t.Errorf("expected 3 folders, got %d", len(folders))
	}
}

func TestFolderHandler_ListTree(t *testing.T) {
	handler, repo := setupFolderHandler(t)
	ctx := testutil.TestContext()

	// Create parent and child
	parent, err := repo.Create(ctx, &models.FolderInput{Name: "Parent"})
	if err != nil {
		t.Fatalf("failed to create parent: %v", err)
	}
	_, err = repo.Create(ctx, &models.FolderInput{Name: "Child", ParentID: &parent.ID})
	if err != nil {
		t.Fatalf("failed to create child: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/folders?tree=true", nil)
	req = withRequestID(req)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var envelope testAPIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Unmarshal data field
	dataBytes, _ := json.Marshal(envelope.Data)
	var folders []models.Folder
	if err := json.Unmarshal(dataBytes, &folders); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	// Should have 1 root folder with 1 child
	if len(folders) != 1 {
		t.Errorf("expected 1 root folder, got %d", len(folders))
	}
}

// Health Handler Tests

func TestHealthHandler_Ping(t *testing.T) {
	db := testutil.TestDB(t)
	handler := NewHealthHandler(db, "1.0.0", "abc123", nil)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	handler.Ping(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != "pong" {
		t.Errorf("expected 'pong', got %q", w.Body.String())
	}
}

func TestHealthHandler_Health(t *testing.T) {
	db := testutil.TestDB(t)
	handler := NewHealthHandler(db, "1.0.0", "abc123", nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req = withRequestID(req)
	w := httptest.NewRecorder()

	handler.Health(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var envelope testAPIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Check meta fields
	if envelope.Meta == nil {
		t.Fatal("expected meta to be present")
	}

	// Unmarshal data field
	dataBytes, _ := json.Marshal(envelope.Data)
	var response map[string]interface{}
	if err := json.Unmarshal(dataBytes, &response); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("expected status 'healthy', got %v", response["status"])
	}
	if response["version"] != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %v", response["version"])
	}
}

func TestHealthHandler_Health_WithFeatureFlags(t *testing.T) {
	db := testutil.TestDB(t)

	// Test with all features enabled
	features := &config.FeatureFlags{
		PublicSnippets: true,
		S3Sync:         true,
		APITokens:      true,
		BackupRestore:  true,
	}
	handler := NewHealthHandler(db, "1.0.0", "abc123", features)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req = withRequestID(req)
	w := httptest.NewRecorder()

	handler.Health(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var envelope testAPIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Unmarshal data field
	dataBytes, _ := json.Marshal(envelope.Data)
	var response map[string]interface{}
	if err := json.Unmarshal(dataBytes, &response); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	// Check feature flags
	featuresMap, ok := response["features"].(map[string]interface{})
	if !ok {
		t.Fatal("expected features field to be present")
	}

	if featuresMap["public_snippets"] != true {
		t.Errorf("expected public_snippets to be true, got %v", featuresMap["public_snippets"])
	}
	if featuresMap["s3_sync"] != true {
		t.Errorf("expected s3_sync to be true, got %v", featuresMap["s3_sync"])
	}
	if featuresMap["api_tokens"] != true {
		t.Errorf("expected api_tokens to be true, got %v", featuresMap["api_tokens"])
	}
	if featuresMap["backup_restore"] != true {
		t.Errorf("expected backup_restore to be true, got %v", featuresMap["backup_restore"])
	}
}
