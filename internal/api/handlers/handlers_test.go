package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/MohamedElashri/snipo/internal/models"
	"github.com/MohamedElashri/snipo/internal/repository"
	"github.com/MohamedElashri/snipo/internal/services"
	"github.com/MohamedElashri/snipo/internal/testutil"
)

// withChiURLParams adds chi URL params to a request context
func withChiURLParams(r *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for key, val := range params {
		rctx.URLParams.Add(key, val)
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
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
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	var response models.Snippet
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
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

	w := httptest.NewRecorder()
	handler.Get(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var response models.Snippet
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
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
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var response models.SnippetListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(response.Data) != 3 {
		t.Errorf("expected 3 snippets, got %d", len(response.Data))
	}
	if response.Pagination.Total != 3 {
		t.Errorf("expected total 3, got %d", response.Pagination.Total)
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
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response models.SnippetListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(response.Data) != 3 {
		t.Errorf("expected 3 snippets on page, got %d", len(response.Data))
	}
	if response.Pagination.Total != 10 {
		t.Errorf("expected total 10, got %d", response.Pagination.Total)
	}
	if response.Pagination.TotalPages != 4 {
		t.Errorf("expected 4 total pages, got %d", response.Pagination.TotalPages)
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

	w := httptest.NewRecorder()
	handler.Update(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var response models.Snippet
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
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

	// Verify deleted
	deleted, _ := repo.GetByID(ctx, snippet.ID)
	if deleted != nil {
		t.Error("expected snippet to be deleted")
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

	w := httptest.NewRecorder()
	handler.ToggleFavorite(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response models.Snippet
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
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
	w := httptest.NewRecorder()

	handler.Search(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var response map[string][]models.Snippet
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(response["data"]) != 1 {
		t.Errorf("expected 1 result for 'hello', got %d", len(response["data"]))
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
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	var response models.Tag
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
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
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string][]models.Tag
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(response["data"]) != 3 {
		t.Errorf("expected 3 tags, got %d", len(response["data"]))
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
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	var response models.Folder
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
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
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string][]models.Folder
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(response["data"]) != 3 {
		t.Errorf("expected 3 folders, got %d", len(response["data"]))
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
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string][]models.Folder
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Should have 1 root folder with 1 child
	if len(response["data"]) != 1 {
		t.Errorf("expected 1 root folder, got %d", len(response["data"]))
	}
}

// Health Handler Tests

func TestHealthHandler_Ping(t *testing.T) {
	db := testutil.TestDB(t)
	handler := NewHealthHandler(db, "1.0.0", "abc123")

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
	handler := NewHealthHandler(db, "1.0.0", "abc123")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.Health(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("expected status 'healthy', got %v", response["status"])
	}
	if response["version"] != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %v", response["version"])
	}
}
