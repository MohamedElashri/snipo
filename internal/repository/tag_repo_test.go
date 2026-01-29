package repository

import (
	"testing"

	"github.com/MohamedElashri/snipo/internal/models"
	"github.com/MohamedElashri/snipo/internal/testutil"
)

func TestTagRepository_Create(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewTagRepository(db)
	ctx := testutil.TestContext()

	input := &models.TagInput{
		Name:  "golang",
		Color: "#00ADD8",
	}

	tag, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if tag.ID == 0 {
		t.Error("expected tag ID to be set")
	}
	if tag.Name != input.Name {
		t.Errorf("expected name %q, got %q", input.Name, tag.Name)
	}
	if tag.Color != input.Color {
		t.Errorf("expected color %q, got %q", input.Color, tag.Color)
	}
}

func TestTagRepository_Create_Duplicate(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewTagRepository(db)
	ctx := testutil.TestContext()

	input := &models.TagInput{
		Name:  "duplicate",
		Color: "#FF0000",
	}

	// Create first
	_, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("First Create failed: %v", err)
	}

	// Try to create duplicate
	_, err = repo.Create(ctx, input)
	if err == nil {
		t.Error("expected error for duplicate tag name")
	}
}

func TestTagRepository_GetByID(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewTagRepository(db)
	ctx := testutil.TestContext()

	// Create a tag
	input := &models.TagInput{
		Name:  "test-tag",
		Color: "#123456",
	}
	created, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Get by ID
	tag, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if tag.ID != created.ID {
		t.Errorf("expected ID %d, got %d", created.ID, tag.ID)
	}
	if tag.Name != input.Name {
		t.Errorf("expected name %q, got %q", input.Name, tag.Name)
	}
}

func TestTagRepository_GetByID_NotFound(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewTagRepository(db)
	ctx := testutil.TestContext()

	_, err := repo.GetByID(ctx, 99999)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestTagRepository_GetByName(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewTagRepository(db)
	ctx := testutil.TestContext()

	// Create a tag
	input := &models.TagInput{
		Name:  "find-me",
		Color: "#ABCDEF",
	}
	created, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Get by name
	tag, err := repo.GetByName(ctx, "find-me")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}

	if tag.ID != created.ID {
		t.Errorf("expected ID %d, got %d", created.ID, tag.ID)
	}
}

func TestTagRepository_GetByName_NotFound(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewTagRepository(db)
	ctx := testutil.TestContext()

	_, err := repo.GetByName(ctx, "nonexistent")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestTagRepository_List(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewTagRepository(db)
	ctx := testutil.TestContext()

	// Create multiple tags
	names := []string{"alpha", "beta", "gamma"}
	for _, name := range names {
		input := &models.TagInput{Name: name, Color: "#000000"}
		_, err := repo.Create(ctx, input)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	// List all
	tags, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(tags) != 3 {
		t.Errorf("expected 3 tags, got %d", len(tags))
	}

	// Should be sorted by name
	if tags[0].Name != "alpha" {
		t.Errorf("expected first tag to be 'alpha', got %q", tags[0].Name)
	}
}

func TestTagRepository_List_OrderBySnippetCount(t *testing.T) {
	db := testutil.TestDB(t)
	tagRepo := NewTagRepository(db)
	snippetRepo := NewSnippetRepository(db)
	ctx := testutil.TestContext()

	// Create tags with special characters
	tagNames := []string{"C#", "Python", "Go", "JavaScript", "C++"}
	for _, name := range tagNames {
		input := &models.TagInput{Name: name, Color: "#000000"}
		_, err := tagRepo.Create(ctx, input)
		if err != nil {
			t.Fatalf("Create tag failed: %v", err)
		}
	}

	// Create snippets and assign tags with different counts
	// C# - 5 snippets (most popular)
	for i := 0; i < 5; i++ {
		snippet, err := snippetRepo.Create(ctx, &models.SnippetInput{
			Title:    "Test Snippet",
			Content:  "content",
			Language: "plaintext",
		})
		if err != nil {
			t.Fatalf("Create snippet failed: %v", err)
		}
		if err := tagRepo.SetSnippetTags(ctx, snippet.ID, []string{"C#"}); err != nil {
			t.Fatalf("SetSnippetTags failed: %v", err)
		}
	}

	// Python - 3 snippets
	for i := 0; i < 3; i++ {
		snippet, err := snippetRepo.Create(ctx, &models.SnippetInput{
			Title:    "Test Snippet",
			Content:  "content",
			Language: "plaintext",
		})
		if err != nil {
			t.Fatalf("Create snippet failed: %v", err)
		}
		if err := tagRepo.SetSnippetTags(ctx, snippet.ID, []string{"Python"}); err != nil {
			t.Fatalf("SetSnippetTags failed: %v", err)
		}
	}

	// JavaScript - 3 snippets (same count as Python)
	for i := 0; i < 3; i++ {
		snippet, err := snippetRepo.Create(ctx, &models.SnippetInput{
			Title:    "Test Snippet",
			Content:  "content",
			Language: "plaintext",
		})
		if err != nil {
			t.Fatalf("Create snippet failed: %v", err)
		}
		if err := tagRepo.SetSnippetTags(ctx, snippet.ID, []string{"JavaScript"}); err != nil {
			t.Fatalf("SetSnippetTags failed: %v", err)
		}
	}

	// Go - 1 snippet
	snippet, err := snippetRepo.Create(ctx, &models.SnippetInput{
		Title:    "Test Snippet",
		Content:  "content",
		Language: "plaintext",
	})
	if err != nil {
		t.Fatalf("Create snippet failed: %v", err)
	}
	if err := tagRepo.SetSnippetTags(ctx, snippet.ID, []string{"Go"}); err != nil {
		t.Fatalf("SetSnippetTags failed: %v", err)
	}

	// C++ - 0 snippets (no assignment)

	// List all tags
	tags, err := tagRepo.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(tags) != 5 {
		t.Errorf("expected 5 tags, got %d", len(tags))
	}

	// Verify ordering: should be ordered by snippet count DESC, then alphabetically ASC
	// Expected order: C# (5), JavaScript (3), Python (3), Go (1), C++ (0)
	expectedOrder := []struct {
		name  string
		count int
	}{
		{"C#", 5},
		{"JavaScript", 3}, // JavaScript comes before Python alphabetically
		{"Python", 3},
		{"Go", 1},
		{"C++", 0},
	}

	for i, expected := range expectedOrder {
		if i >= len(tags) {
			t.Fatalf("not enough tags returned")
		}
		if tags[i].Name != expected.name {
			t.Errorf("position %d: expected tag %q, got %q", i, expected.name, tags[i].Name)
		}
		if tags[i].SnippetCount != expected.count {
			t.Errorf("tag %q: expected count %d, got %d", tags[i].Name, expected.count, tags[i].SnippetCount)
		}
	}

	// Verify that tags with special characters like C# are not sorted purely alphabetically
	// (which would put them at the top regardless of count)
	if tags[0].Name != "C#" {
		t.Errorf("C# with most snippets should be first, but got %q", tags[0].Name)
	}
}

func TestTagRepository_Update(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewTagRepository(db)
	ctx := testutil.TestContext()

	// Create a tag
	input := &models.TagInput{
		Name:  "original",
		Color: "#111111",
	}
	created, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Update it
	updateInput := &models.TagInput{
		Name:  "updated",
		Color: "#222222",
	}
	updated, err := repo.Update(ctx, created.ID, updateInput)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if updated.Name != updateInput.Name {
		t.Errorf("expected name %q, got %q", updateInput.Name, updated.Name)
	}
	if updated.Color != updateInput.Color {
		t.Errorf("expected color %q, got %q", updateInput.Color, updated.Color)
	}
}

func TestTagRepository_Update_NotFound(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewTagRepository(db)
	ctx := testutil.TestContext()

	input := &models.TagInput{Name: "test", Color: "#000000"}
	_, err := repo.Update(ctx, 99999, input)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestTagRepository_Delete(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewTagRepository(db)
	ctx := testutil.TestContext()

	// Create a tag
	input := &models.TagInput{Name: "to-delete", Color: "#000000"}
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
	_, err = repo.GetByID(ctx, created.ID)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestTagRepository_Delete_NotFound(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewTagRepository(db)
	ctx := testutil.TestContext()

	err := repo.Delete(ctx, 99999)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestTagRepository_SetSnippetTags(t *testing.T) {
	db := testutil.TestDB(t)
	tagRepo := NewTagRepository(db)
	snippetRepo := NewSnippetRepository(db)
	ctx := testutil.TestContext()

	// Create a snippet
	snippetInput := &models.SnippetInput{
		Title:    "Test Snippet",
		Content:  "content",
		Language: "plaintext",
	}
	snippet, err := snippetRepo.Create(ctx, snippetInput)
	if err != nil {
		t.Fatalf("Create snippet failed: %v", err)
	}

	// Set tags (should auto-create)
	err = tagRepo.SetSnippetTags(ctx, snippet.ID, []string{"tag1", "tag2", "tag3"})
	if err != nil {
		t.Fatalf("SetSnippetTags failed: %v", err)
	}

	// Get snippet tags
	tags, err := tagRepo.GetSnippetTags(ctx, snippet.ID)
	if err != nil {
		t.Fatalf("GetSnippetTags failed: %v", err)
	}

	if len(tags) != 3 {
		t.Errorf("expected 3 tags, got %d", len(tags))
	}
}

func TestTagRepository_SetSnippetTags_Replace(t *testing.T) {
	db := testutil.TestDB(t)
	tagRepo := NewTagRepository(db)
	snippetRepo := NewSnippetRepository(db)
	ctx := testutil.TestContext()

	// Create a snippet
	snippetInput := &models.SnippetInput{
		Title:    "Test Snippet",
		Content:  "content",
		Language: "plaintext",
	}
	snippet, err := snippetRepo.Create(ctx, snippetInput)
	if err != nil {
		t.Fatalf("Create snippet failed: %v", err)
	}

	// Set initial tags
	err = tagRepo.SetSnippetTags(ctx, snippet.ID, []string{"old1", "old2"})
	if err != nil {
		t.Fatalf("SetSnippetTags failed: %v", err)
	}

	// Replace with new tags
	err = tagRepo.SetSnippetTags(ctx, snippet.ID, []string{"new1"})
	if err != nil {
		t.Fatalf("SetSnippetTags (replace) failed: %v", err)
	}

	// Verify
	tags, err := tagRepo.GetSnippetTags(ctx, snippet.ID)
	if err != nil {
		t.Fatalf("GetSnippetTags failed: %v", err)
	}

	if len(tags) != 1 {
		t.Errorf("expected 1 tag after replace, got %d", len(tags))
	}
	if tags[0].Name != "new1" {
		t.Errorf("expected tag 'new1', got %q", tags[0].Name)
	}
}

func TestTagRepository_GetTagSnippetCount(t *testing.T) {
	db := testutil.TestDB(t)
	tagRepo := NewTagRepository(db)
	snippetRepo := NewSnippetRepository(db)
	ctx := testutil.TestContext()

	// Create a tag
	tagInput := &models.TagInput{Name: "count-test", Color: "#000000"}
	tag, err := tagRepo.Create(ctx, tagInput)
	if err != nil {
		t.Fatalf("Create tag failed: %v", err)
	}

	// Create snippets and assign tag
	for i := 0; i < 3; i++ {
		snippet, err := snippetRepo.Create(ctx, &models.SnippetInput{
			Title:    "Snippet",
			Content:  "content",
			Language: "plaintext",
		})
		if err != nil {
			t.Fatalf("Create snippet failed: %v", err)
		}
		err = tagRepo.SetSnippetTags(ctx, snippet.ID, []string{"count-test"})
		if err != nil {
			t.Fatalf("SetSnippetTags failed: %v", err)
		}
	}

	// Get count
	count, err := tagRepo.GetTagSnippetCount(ctx, tag.ID)
	if err != nil {
		t.Fatalf("GetTagSnippetCount failed: %v", err)
	}

	if count != 3 {
		t.Errorf("expected count 3, got %d", count)
	}
}

func TestTagRepository_GetTagSnippetCount_Archived(t *testing.T) {
	db := testutil.TestDB(t)
	tagRepo := NewTagRepository(db)
	snippetRepo := NewSnippetRepository(db)
	ctx := testutil.TestContext()

	// Create a tag
	tagInput := &models.TagInput{Name: "archive-count-test", Color: "#000000"}
	tag, err := tagRepo.Create(ctx, tagInput)
	if err != nil {
		t.Fatalf("Create tag failed: %v", err)
	}

	// Create snippet and assign tag
	snippet, err := snippetRepo.Create(ctx, &models.SnippetInput{
		Title:    "Snippet",
		Content:  "content",
		Language: "plaintext",
	})
	if err != nil {
		t.Fatalf("Create snippet failed: %v", err)
	}
	err = tagRepo.SetSnippetTags(ctx, snippet.ID, []string{"archive-count-test"})
	if err != nil {
		t.Fatalf("SetSnippetTags failed: %v", err)
	}

	// Initial count should be 1
	count, err := tagRepo.GetTagSnippetCount(ctx, tag.ID)
	if err != nil {
		t.Fatalf("GetTagSnippetCount failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}

	// Archive the snippet
	_, err = snippetRepo.ToggleArchive(ctx, snippet.ID)
	if err != nil {
		t.Fatalf("ToggleArchive failed: %v", err)
	}

	// Count should now be 0
	count, err = tagRepo.GetTagSnippetCount(ctx, tag.ID)
	if err != nil {
		t.Fatalf("GetTagSnippetCount failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected count 0 after archiving, got %d", count)
	}

	// Unarchive
	_, err = snippetRepo.ToggleArchive(ctx, snippet.ID)
	if err != nil {
		t.Fatalf("ToggleArchive failed: %v", err)
	}

	// Count should be 1 again
	count, err = tagRepo.GetTagSnippetCount(ctx, tag.ID)
	if err != nil {
		t.Fatalf("GetTagSnippetCount failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected count 1 after unarchiving, got %d", count)
	}
}
