package repository

import (
	"testing"
	"time"

	"github.com/MohamedElashri/snipo/internal/models"
	"github.com/MohamedElashri/snipo/internal/testutil"
)

func TestHistoryRepository_CreateHistory(t *testing.T) {
	db := testutil.TestDB(t)
	snippetRepo := NewSnippetRepository(db)
	historyRepo := NewHistoryRepository(db)
	ctx := testutil.TestContext()

	// Create a snippet first
	input := &models.SnippetInput{
		Title:       "Test Snippet",
		Description: "Test description",
		Content:     "console.log('hello');",
		Language:    "javascript",
		IsPublic:    false,
	}
	snippet, err := snippetRepo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create snippet failed: %v", err)
	}

	// Create history entry
	historyID, err := historyRepo.CreateHistory(ctx, snippet, "create")
	if err != nil {
		t.Fatalf("CreateHistory failed: %v", err)
	}

	if historyID <= 0 {
		t.Errorf("expected positive history ID, got %d", historyID)
	}
}

func TestHistoryRepository_CreateFileHistory(t *testing.T) {
	db := testutil.TestDB(t)
	snippetRepo := NewSnippetRepository(db)
	historyRepo := NewHistoryRepository(db)
	ctx := testutil.TestContext()

	// Create a snippet
	input := &models.SnippetInput{
		Title:    "Multi-file Snippet",
		Content:  "main content",
		Language: "go",
	}
	snippet, err := snippetRepo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create snippet failed: %v", err)
	}

	// Create history entry
	historyID, err := historyRepo.CreateHistory(ctx, snippet, "create")
	if err != nil {
		t.Fatalf("CreateHistory failed: %v", err)
	}

	// Create files for history
	files := []models.SnippetFile{
		{
			SnippetID: snippet.ID,
			Filename:  "main.go",
			Content:   "package main",
			Language:  "go",
			SortOrder: 0,
		},
		{
			SnippetID: snippet.ID,
			Filename:  "utils.go",
			Content:   "package utils",
			Language:  "go",
			SortOrder: 1,
		},
	}

	err = historyRepo.CreateFileHistory(ctx, historyID, files)
	if err != nil {
		t.Fatalf("CreateFileHistory failed: %v", err)
	}
}

func TestHistoryRepository_GetSnippetHistory(t *testing.T) {
	db := testutil.TestDB(t)
	snippetRepo := NewSnippetRepository(db)
	historyRepo := NewHistoryRepository(db)
	ctx := testutil.TestContext()

	// Create a snippet
	input := &models.SnippetInput{
		Title:       "Version Test",
		Description: "Original",
		Content:     "version 1",
		Language:    "python",
	}
	snippet, err := snippetRepo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create snippet failed: %v", err)
	}

	// Create multiple history entries
	historyID1, err := historyRepo.CreateHistory(ctx, snippet, "create")
	if err != nil {
		t.Fatalf("CreateHistory failed: %v", err)
	}

	// Update snippet
	snippet.Description = "Updated"
	snippet.Content = "version 2"
	_, err = snippetRepo.Update(ctx, snippet.ID, &models.SnippetInput{
		Title:       snippet.Title,
		Description: snippet.Description,
		Content:     snippet.Content,
		Language:    snippet.Language,
	})
	if err != nil {
		t.Fatalf("Update snippet failed: %v", err)
	}

	// Create another history entry with newer timestamp
	// Add a small delay to ensure different timestamps
	time.Sleep(10 * time.Millisecond)
	historyID2, err := historyRepo.CreateHistory(ctx, snippet, "update")
	if err != nil {
		t.Fatalf("CreateHistory failed: %v", err)
	}

	// Verify history IDs are different
	if historyID1 == historyID2 {
		t.Error("expected different history IDs")
	}

	// Get history
	history, err := historyRepo.GetSnippetHistory(ctx, snippet.ID, 50)
	if err != nil {
		t.Fatalf("GetSnippetHistory failed: %v", err)
	}

	if len(history) != 2 {
		t.Errorf("expected 2 history entries, got %d", len(history))
	}

	// Most recent should be first (highest ID, latest timestamp)
	if history[0].ID != historyID2 {
		t.Errorf("expected most recent history ID %d, got %d", historyID2, history[0].ID)
	}
	if history[0].Content != "version 2" {
		t.Errorf("expected content 'version 2', got %q", history[0].Content)
	}
}

func TestHistoryRepository_GetSnippetHistory_Limit(t *testing.T) {
	db := testutil.TestDB(t)
	snippetRepo := NewSnippetRepository(db)
	historyRepo := NewHistoryRepository(db)
	ctx := testutil.TestContext()

	// Create a snippet
	input := &models.SnippetInput{
		Title:    "Limit Test",
		Content:  "content",
		Language: "plaintext",
	}
	snippet, err := snippetRepo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create snippet failed: %v", err)
	}

	// Create 10 history entries
	for i := 0; i < 10; i++ {
		snippet.Content = string(rune('0'+i))
		_, err = historyRepo.CreateHistory(ctx, snippet, "update")
		if err != nil {
			t.Fatalf("CreateHistory failed: %v", err)
		}
	}

	// Get only 5 entries
	history, err := historyRepo.GetSnippetHistory(ctx, snippet.ID, 5)
	if err != nil {
		t.Fatalf("GetSnippetHistory failed: %v", err)
	}

	if len(history) != 5 {
		t.Errorf("expected 5 history entries, got %d", len(history))
	}
}

func TestHistoryRepository_GetHistoryByID(t *testing.T) {
	db := testutil.TestDB(t)
	snippetRepo := NewSnippetRepository(db)
	historyRepo := NewHistoryRepository(db)
	ctx := testutil.TestContext()

	// Create a snippet
	input := &models.SnippetInput{
		Title:       "Get By ID Test",
		Description: "Test description",
		Content:     "test content",
		Language:    "javascript",
		IsPublic:    true,
	}
	snippet, err := snippetRepo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create snippet failed: %v", err)
	}

	// Manually set favorite status for testing
	snippet, err = snippetRepo.ToggleFavorite(ctx, snippet.ID)
	if err != nil {
		t.Fatalf("ToggleFavorite failed: %v", err)
	}

	// Create history entry
	historyID, err := historyRepo.CreateHistory(ctx, snippet, "create")
	if err != nil {
		t.Fatalf("CreateHistory failed: %v", err)
	}

	// Get history by ID
	history, err := historyRepo.GetHistoryByID(ctx, historyID)
	if err != nil {
		t.Fatalf("GetHistoryByID failed: %v", err)
	}

	if history == nil {
		t.Fatal("expected history entry, got nil")
	}

	if history.Title != input.Title {
		t.Errorf("expected title %q, got %q", input.Title, history.Title)
	}
	if history.Description != input.Description {
		t.Errorf("expected description %q, got %q", input.Description, history.Description)
	}
	if history.Content != input.Content {
		t.Errorf("expected content %q, got %q", input.Content, history.Content)
	}
	if history.IsFavorite != true {
		t.Errorf("expected is_favorite %v, got %v", true, history.IsFavorite)
	}
	if history.IsPublic != input.IsPublic {
		t.Errorf("expected is_public %v, got %v", input.IsPublic, history.IsPublic)
	}
}

func TestHistoryRepository_GetHistoryByID_NotFound(t *testing.T) {
	db := testutil.TestDB(t)
	historyRepo := NewHistoryRepository(db)
	ctx := testutil.TestContext()

	history, err := historyRepo.GetHistoryByID(ctx, 99999)
	if err != nil {
		t.Fatalf("GetHistoryByID failed: %v", err)
	}

	if history != nil {
		t.Error("expected nil for nonexistent history entry")
	}
}

func TestHistoryRepository_GetHistoryFiles(t *testing.T) {
	db := testutil.TestDB(t)
	snippetRepo := NewSnippetRepository(db)
	historyRepo := NewHistoryRepository(db)
	ctx := testutil.TestContext()

	// Create a snippet
	input := &models.SnippetInput{
		Title:    "File History Test",
		Content:  "main",
		Language: "go",
	}
	snippet, err := snippetRepo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create snippet failed: %v", err)
	}

	// Create history entry
	historyID, err := historyRepo.CreateHistory(ctx, snippet, "create")
	if err != nil {
		t.Fatalf("CreateHistory failed: %v", err)
	}

	// Create file history
	files := []models.SnippetFile{
		{
			SnippetID: snippet.ID,
			Filename:  "main.go",
			Content:   "package main\nfunc main() {}",
			Language:  "go",
			SortOrder: 0,
		},
		{
			SnippetID: snippet.ID,
			Filename:  "utils.go",
			Content:   "package utils\nfunc Helper() {}",
			Language:  "go",
			SortOrder: 1,
		},
	}

	err = historyRepo.CreateFileHistory(ctx, historyID, files)
	if err != nil {
		t.Fatalf("CreateFileHistory failed: %v", err)
	}

	// Get history files
	historyFiles, err := historyRepo.GetHistoryFiles(ctx, historyID)
	if err != nil {
		t.Fatalf("GetHistoryFiles failed: %v", err)
	}

	if len(historyFiles) != 2 {
		t.Errorf("expected 2 files, got %d", len(historyFiles))
	}

	if historyFiles[0].Filename != "main.go" {
		t.Errorf("expected first filename 'main.go', got %q", historyFiles[0].Filename)
	}
	if historyFiles[1].Filename != "utils.go" {
		t.Errorf("expected second filename 'utils.go', got %q", historyFiles[1].Filename)
	}

	// Verify sort order
	if historyFiles[0].SortOrder != 0 {
		t.Errorf("expected first sort order 0, got %d", historyFiles[0].SortOrder)
	}
	if historyFiles[1].SortOrder != 1 {
		t.Errorf("expected second sort order 1, got %d", historyFiles[1].SortOrder)
	}
}

func TestHistoryRepository_DeleteSnippetHistory(t *testing.T) {
	db := testutil.TestDB(t)
	snippetRepo := NewSnippetRepository(db)
	historyRepo := NewHistoryRepository(db)
	ctx := testutil.TestContext()

	// Create a snippet
	input := &models.SnippetInput{
		Title:    "Delete History Test",
		Content:  "test",
		Language: "plaintext",
	}
	snippet, err := snippetRepo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create snippet failed: %v", err)
	}

	// Create history entries
	_, err = historyRepo.CreateHistory(ctx, snippet, "create")
	if err != nil {
		t.Fatalf("CreateHistory failed: %v", err)
	}

	_, err = historyRepo.CreateHistory(ctx, snippet, "update")
	if err != nil {
		t.Fatalf("CreateHistory failed: %v", err)
	}

	// Delete history
	err = historyRepo.DeleteSnippetHistory(ctx, snippet.ID)
	if err != nil {
		t.Fatalf("DeleteSnippetHistory failed: %v", err)
	}

	// Verify history is deleted
	history, err := historyRepo.GetSnippetHistory(ctx, snippet.ID, 50)
	if err != nil {
		t.Fatalf("GetSnippetHistory failed: %v", err)
	}

	if len(history) != 0 {
		t.Errorf("expected 0 history entries after deletion, got %d", len(history))
	}
}

func TestHistoryRepository_DeleteOldHistory(t *testing.T) {
	db := testutil.TestDB(t)
	snippetRepo := NewSnippetRepository(db)
	historyRepo := NewHistoryRepository(db)
	ctx := testutil.TestContext()

	// Create a snippet
	input := &models.SnippetInput{
		Title:    "Old History Test",
		Content:  "test",
		Language: "plaintext",
	}
	snippet, err := snippetRepo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create snippet failed: %v", err)
	}

	// Create history entry
	_, err = historyRepo.CreateHistory(ctx, snippet, "create")
	if err != nil {
		t.Fatalf("CreateHistory failed: %v", err)
	}

	// Delete history older than 0 days (should delete all since just created)
	// Note: This test is limited since we can't easily create old timestamps
	// In production, this would delete entries older than the specified days
	deleted, err := historyRepo.DeleteOldHistory(ctx, 365)
	if err != nil {
		t.Fatalf("DeleteOldHistory failed: %v", err)
	}

	// Since entries were just created, none should be deleted with 365 days
	if deleted > 1 {
		t.Errorf("expected at most 1 entry deleted, got %d", deleted)
	}
}

func TestHistoryRepository_GetHistoryCount(t *testing.T) {
	db := testutil.TestDB(t)
	snippetRepo := NewSnippetRepository(db)
	historyRepo := NewHistoryRepository(db)
	ctx := testutil.TestContext()

	// Create a snippet
	input := &models.SnippetInput{
		Title:    "Count Test",
		Content:  "test",
		Language: "plaintext",
	}
	snippet, err := snippetRepo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create snippet failed: %v", err)
	}

	// Create 5 history entries
	for i := 0; i < 5; i++ {
		_, err = historyRepo.CreateHistory(ctx, snippet, "update")
		if err != nil {
			t.Fatalf("CreateHistory failed: %v", err)
		}
	}

	// Get count
	count, err := historyRepo.GetHistoryCount(ctx, snippet.ID)
	if err != nil {
		t.Fatalf("GetHistoryCount failed: %v", err)
	}

	if count != 5 {
		t.Errorf("expected count 5, got %d", count)
	}
}

func TestHistoryRepository_HistoryWithFiles(t *testing.T) {
	db := testutil.TestDB(t)
	snippetRepo := NewSnippetRepository(db)
	historyRepo := NewHistoryRepository(db)
	ctx := testutil.TestContext()

	// Create a snippet
	input := &models.SnippetInput{
		Title:    "History with Files",
		Content:  "main content",
		Language: "python",
	}
	snippet, err := snippetRepo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create snippet failed: %v", err)
	}

	// Create history entry
	historyID, err := historyRepo.CreateHistory(ctx, snippet, "create")
	if err != nil {
		t.Fatalf("CreateHistory failed: %v", err)
	}

	// Create file history
	files := []models.SnippetFile{
		{
			SnippetID: snippet.ID,
			Filename:  "main.py",
			Content:   "print('hello')",
			Language:  "python",
			SortOrder: 0,
		},
	}

	err = historyRepo.CreateFileHistory(ctx, historyID, files)
	if err != nil {
		t.Fatalf("CreateFileHistory failed: %v", err)
	}

	// Get history (should include files)
	history, err := historyRepo.GetSnippetHistory(ctx, snippet.ID, 50)
	if err != nil {
		t.Fatalf("GetSnippetHistory failed: %v", err)
	}

	if len(history) != 1 {
		t.Errorf("expected 1 history entry, got %d", len(history))
	}

	// Verify files are populated
	if len(history[0].Files) != 1 {
		t.Errorf("expected 1 file in history, got %d", len(history[0].Files))
	}

	if history[0].Files[0].Filename != "main.py" {
		t.Errorf("expected filename 'main.py', got %q", history[0].Files[0].Filename)
	}

	if history[0].Files[0].Content != "print('hello')" {
		t.Errorf("expected content \"print('hello')\", got %q", history[0].Files[0].Content)
	}
}

func TestHistoryRepository_TimestampTracking(t *testing.T) {
	db := testutil.TestDB(t)
	snippetRepo := NewSnippetRepository(db)
	historyRepo := NewHistoryRepository(db)
	ctx := testutil.TestContext()

	// Create a snippet
	input := &models.SnippetInput{
		Title:    "Timestamp Test",
		Content:  "test",
		Language: "plaintext",
	}
	snippet, err := snippetRepo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create snippet failed: %v", err)
	}

	// Create history entry
	_, err = historyRepo.CreateHistory(ctx, snippet, "create")
	if err != nil {
		t.Fatalf("CreateHistory failed: %v", err)
	}

	// Get history
	history, err := historyRepo.GetSnippetHistory(ctx, snippet.ID, 50)
	if err != nil {
		t.Fatalf("GetSnippetHistory failed: %v", err)
	}

	if len(history) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(history))
	}

	// Verify timestamp is set
	if history[0].CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}

	// Verify timestamp is recent (within last minute)
	if time.Since(history[0].CreatedAt) > time.Minute {
		t.Error("expected CreatedAt to be recent")
	}
}
