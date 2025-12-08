package repository

import (
	"testing"

	"github.com/MohamedElashri/snipo/internal/models"
	"github.com/MohamedElashri/snipo/internal/testutil"
)

func TestSnippetRepository_Create(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewSnippetRepository(db)
	ctx := testutil.TestContext()

	input := &models.SnippetInput{
		Title:       "Test Snippet",
		Description: "A test description",
		Content:     "console.log('hello');",
		Language:    "javascript",
		IsPublic:    false,
	}

	snippet, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if snippet.ID == "" {
		t.Error("expected snippet ID to be set")
	}
	if snippet.Title != input.Title {
		t.Errorf("expected title %q, got %q", input.Title, snippet.Title)
	}
	if snippet.Content != input.Content {
		t.Errorf("expected content %q, got %q", input.Content, snippet.Content)
	}
	if snippet.Language != input.Language {
		t.Errorf("expected language %q, got %q", input.Language, snippet.Language)
	}
	if snippet.IsFavorite {
		t.Error("expected is_favorite to be false")
	}
	if snippet.IsPublic {
		t.Error("expected is_public to be false")
	}
}

func TestSnippetRepository_GetByID(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewSnippetRepository(db)
	ctx := testutil.TestContext()

	// Create a snippet first
	input := &models.SnippetInput{
		Title:    "Test Snippet",
		Content:  "test content",
		Language: "plaintext",
	}
	created, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Get by ID
	snippet, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if snippet == nil {
		t.Fatal("expected snippet, got nil")
	}
	if snippet.ID != created.ID {
		t.Errorf("expected ID %q, got %q", created.ID, snippet.ID)
	}
	if snippet.Title != input.Title {
		t.Errorf("expected title %q, got %q", input.Title, snippet.Title)
	}
}

func TestSnippetRepository_GetByID_NotFound(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewSnippetRepository(db)
	ctx := testutil.TestContext()

	snippet, err := repo.GetByID(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if snippet != nil {
		t.Error("expected nil for nonexistent snippet")
	}
}

func TestSnippetRepository_Update(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewSnippetRepository(db)
	ctx := testutil.TestContext()

	// Create a snippet
	input := &models.SnippetInput{
		Title:    "Original Title",
		Content:  "original content",
		Language: "plaintext",
	}
	created, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Update it
	updateInput := &models.SnippetInput{
		Title:       "Updated Title",
		Description: "new description",
		Content:     "updated content",
		Language:    "go",
		IsPublic:    true,
	}
	updated, err := repo.Update(ctx, created.ID, updateInput)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if updated.Title != updateInput.Title {
		t.Errorf("expected title %q, got %q", updateInput.Title, updated.Title)
	}
	if updated.Description != updateInput.Description {
		t.Errorf("expected description %q, got %q", updateInput.Description, updated.Description)
	}
	if updated.Content != updateInput.Content {
		t.Errorf("expected content %q, got %q", updateInput.Content, updated.Content)
	}
	if updated.Language != updateInput.Language {
		t.Errorf("expected language %q, got %q", updateInput.Language, updated.Language)
	}
	if !updated.IsPublic {
		t.Error("expected is_public to be true")
	}
}

func TestSnippetRepository_Update_NotFound(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewSnippetRepository(db)
	ctx := testutil.TestContext()

	input := &models.SnippetInput{
		Title:    "Test",
		Content:  "test",
		Language: "plaintext",
	}
	updated, err := repo.Update(ctx, "nonexistent", input)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated != nil {
		t.Error("expected nil for nonexistent snippet")
	}
}

func TestSnippetRepository_Delete(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewSnippetRepository(db)
	ctx := testutil.TestContext()

	// Create a snippet
	input := &models.SnippetInput{
		Title:    "To Delete",
		Content:  "content",
		Language: "plaintext",
	}
	created, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Delete it
	err = repo.Delete(ctx, created.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify it's gone
	snippet, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if snippet != nil {
		t.Error("expected snippet to be deleted")
	}
}

func TestSnippetRepository_Delete_NotFound(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewSnippetRepository(db)
	ctx := testutil.TestContext()

	err := repo.Delete(ctx, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent snippet")
	}
}

func TestSnippetRepository_List(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewSnippetRepository(db)
	ctx := testutil.TestContext()

	// Create multiple snippets
	for i := 0; i < 5; i++ {
		input := &models.SnippetInput{
			Title:    "Snippet " + string(rune('A'+i)),
			Content:  "content",
			Language: "plaintext",
		}
		_, err := repo.Create(ctx, input)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	// List all
	filter := models.DefaultSnippetFilter()
	result, err := repo.List(ctx, filter)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if result.Pagination.Total != 5 {
		t.Errorf("expected 5 total, got %d", result.Pagination.Total)
	}
	if len(result.Data) != 5 {
		t.Errorf("expected 5 snippets, got %d", len(result.Data))
	}
}

func TestSnippetRepository_List_Pagination(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewSnippetRepository(db)
	ctx := testutil.TestContext()

	// Create 10 snippets
	for i := 0; i < 10; i++ {
		input := &models.SnippetInput{
			Title:    "Snippet",
			Content:  "content",
			Language: "plaintext",
		}
		_, err := repo.Create(ctx, input)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	// Get first page
	filter := models.SnippetFilter{
		Page:      1,
		Limit:     3,
		SortBy:    "created_at",
		SortOrder: "asc",
	}
	result, err := repo.List(ctx, filter)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if result.Pagination.Total != 10 {
		t.Errorf("expected 10 total, got %d", result.Pagination.Total)
	}
	if len(result.Data) != 3 {
		t.Errorf("expected 3 snippets on page 1, got %d", len(result.Data))
	}
	if result.Pagination.TotalPages != 4 {
		t.Errorf("expected 4 total pages, got %d", result.Pagination.TotalPages)
	}
}

func TestSnippetRepository_List_FilterByLanguage(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewSnippetRepository(db)
	ctx := testutil.TestContext()

	// Create snippets with different languages
	languages := []string{"go", "python", "go", "javascript"}
	for _, lang := range languages {
		input := &models.SnippetInput{
			Title:    "Snippet",
			Content:  "content",
			Language: lang,
		}
		_, err := repo.Create(ctx, input)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	// Filter by Go
	filter := models.DefaultSnippetFilter()
	filter.Language = "go"
	result, err := repo.List(ctx, filter)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if result.Pagination.Total != 2 {
		t.Errorf("expected 2 Go snippets, got %d", result.Pagination.Total)
	}
}

func TestSnippetRepository_List_FilterByFavorite(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewSnippetRepository(db)
	ctx := testutil.TestContext()

	// Create snippets
	for i := 0; i < 3; i++ {
		input := &models.SnippetInput{
			Title:    "Snippet",
			Content:  "content",
			Language: "plaintext",
		}
		created, err := repo.Create(ctx, input)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		// Make first one a favorite
		if i == 0 {
			_, err = repo.ToggleFavorite(ctx, created.ID)
			if err != nil {
				t.Fatalf("ToggleFavorite failed: %v", err)
			}
		}
	}

	// Filter by favorite
	filter := models.DefaultSnippetFilter()
	isFav := true
	filter.IsFavorite = &isFav
	result, err := repo.List(ctx, filter)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if result.Pagination.Total != 1 {
		t.Errorf("expected 1 favorite, got %d", result.Pagination.Total)
	}
}

func TestSnippetRepository_ToggleFavorite(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewSnippetRepository(db)
	ctx := testutil.TestContext()

	// Create a snippet
	input := &models.SnippetInput{
		Title:    "Test",
		Content:  "content",
		Language: "plaintext",
	}
	created, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if created.IsFavorite {
		t.Error("expected is_favorite to be false initially")
	}

	// Toggle to true
	toggled, err := repo.ToggleFavorite(ctx, created.ID)
	if err != nil {
		t.Fatalf("ToggleFavorite failed: %v", err)
	}
	if !toggled.IsFavorite {
		t.Error("expected is_favorite to be true after toggle")
	}

	// Toggle back to false
	toggled, err = repo.ToggleFavorite(ctx, created.ID)
	if err != nil {
		t.Fatalf("ToggleFavorite failed: %v", err)
	}
	if toggled.IsFavorite {
		t.Error("expected is_favorite to be false after second toggle")
	}
}

func TestSnippetRepository_Search(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewSnippetRepository(db)
	ctx := testutil.TestContext()

	// Create snippets with searchable content
	snippets := []models.SnippetInput{
		{Title: "Hello World", Content: "print('hello')", Language: "python"},
		{Title: "Goodbye World", Content: "console.log('bye')", Language: "javascript"},
		{Title: "Test Snippet", Content: "hello there", Language: "plaintext"},
	}
	for _, s := range snippets {
		_, err := repo.Create(ctx, &s)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	// Search for "hello"
	results, err := repo.Search(ctx, "hello", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results for 'hello', got %d", len(results))
	}
}

func TestSnippetRepository_IncrementViewCount(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewSnippetRepository(db)
	ctx := testutil.TestContext()

	// Create a snippet
	input := &models.SnippetInput{
		Title:    "Test",
		Content:  "content",
		Language: "plaintext",
	}
	created, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if created.ViewCount != 0 {
		t.Errorf("expected view_count 0, got %d", created.ViewCount)
	}

	// Increment view count
	err = repo.IncrementViewCount(ctx, created.ID)
	if err != nil {
		t.Fatalf("IncrementViewCount failed: %v", err)
	}

	// Verify
	snippet, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if snippet.ViewCount != 1 {
		t.Errorf("expected view_count 1, got %d", snippet.ViewCount)
	}
}
