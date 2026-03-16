package services

import (
	"testing"

	"github.com/MohamedElashri/snipo/internal/models"
	"github.com/MohamedElashri/snipo/internal/repository"
	"github.com/MohamedElashri/snipo/internal/testutil"
)

func TestSnippetService_GetHistory(t *testing.T) {
	db := testutil.TestDB(t)
	snippetRepo := repository.NewSnippetRepository(db)
	historyRepo := repository.NewHistoryRepository(db)
	fileRepo := repository.NewSnippetFileRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)
	logger := testutil.TestLogger()

	service := NewSnippetService(snippetRepo, logger).
		WithFileRepo(fileRepo).
		WithHistoryRepo(historyRepo).
		WithSettingsRepo(settingsRepo)

	ctx := testutil.TestContext()

	// Create a snippet
	input := &models.SnippetInput{
		Title:       "History Service Test",
		Description: "Original description",
		Content:     "original content",
		Language:    "javascript",
	}
	snippet, err := service.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Update the snippet to create history
	updateInput := &models.SnippetInput{
		Title:       "History Service Test",
		Description: "Updated description",
		Content:     "updated content",
		Language:    "javascript",
	}
	_, err = service.Update(ctx, snippet.ID, updateInput)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Get history
	history, err := service.GetHistory(ctx, snippet.ID, 50)
	if err != nil {
		t.Fatalf("GetHistory failed: %v", err)
	}

	if len(history) < 1 {
		t.Errorf("expected at least 1 history entry, got %d", len(history))
	}

	// Most recent should have original content (before update was saved to history)
	// Note: The service saves history BEFORE updating, so the most recent entry
	// contains the state before the update
	if len(history) > 0 && history[0].Content != "original content" {
		t.Errorf("expected most recent content 'original content' (pre-update state), got %q", history[0].Content)
	}
}

func TestSnippetService_GetHistory_NotFound(t *testing.T) {
	db := testutil.TestDB(t)
	snippetRepo := repository.NewSnippetRepository(db)
	historyRepo := repository.NewHistoryRepository(db)
	logger := testutil.TestLogger()

	service := NewSnippetService(snippetRepo, logger).
		WithHistoryRepo(historyRepo)

	ctx := testutil.TestContext()

	// Try to get history for non-existent snippet
	_, err := service.GetHistory(ctx, "nonexistent-id", 50)
	if err == nil {
		t.Error("expected error for non-existent snippet")
	}
}

func TestSnippetService_RestoreFromHistory(t *testing.T) {
	db := testutil.TestDB(t)
	snippetRepo := repository.NewSnippetRepository(db)
	historyRepo := repository.NewHistoryRepository(db)
	fileRepo := repository.NewSnippetFileRepository(db)
	tagRepo := repository.NewTagRepository(db)
	folderRepo := repository.NewFolderRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)
	logger := testutil.TestLogger()

	service := NewSnippetService(snippetRepo, logger).
		WithFileRepo(fileRepo).
		WithHistoryRepo(historyRepo).
		WithTagRepo(tagRepo).
		WithFolderRepo(folderRepo).
		WithSettingsRepo(settingsRepo)

	ctx := testutil.TestContext()

	// Create a snippet
	input := &models.SnippetInput{
		Title:       "Restore Test",
		Description: "Original",
		Content:     "original version",
		Language:    "python",
		IsPublic:    false,
	}
	snippet, err := service.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Get the creation history entry
	history, err := historyRepo.GetSnippetHistory(ctx, snippet.ID, 50)
	if err != nil {
		t.Fatalf("GetSnippetHistory failed: %v", err)
	}

	if len(history) == 0 {
		t.Fatal("expected at least one history entry")
	}

	creationHistory := history[len(history)-1] // Oldest entry

	// Update the snippet
	updateInput := &models.SnippetInput{
		Title:       "Restore Test",
		Description: "Updated",
		Content:     "modified version",
		Language:    "python",
		IsPublic:    true,
	}
	_, err = service.Update(ctx, snippet.ID, updateInput)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	updated, err := service.GetByID(ctx, snippet.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if updated.Content != "modified version" {
		t.Errorf("expected updated content 'modified version', got %q", updated.Content)
	}

	// Restore from creation history
	restored, err := service.RestoreFromHistory(ctx, snippet.ID, creationHistory.ID)
	if err != nil {
		t.Fatalf("RestoreFromHistory failed: %v", err)
	}

	// Verify restoration
	if restored.Title != "Restore Test" {
		t.Errorf("expected title 'Restore Test', got %q", restored.Title)
	}
	if restored.Description != "Original" {
		t.Errorf("expected description 'Original', got %q", restored.Description)
	}
	if restored.Content != "original version" {
		t.Errorf("expected content 'original version', got %q", restored.Content)
	}
	if restored.IsPublic != false {
		t.Errorf("expected IsPublic false, got %v", restored.IsPublic)
	}
}

func TestSnippetService_RestoreFromHistory_WithFiles(t *testing.T) {
	db := testutil.TestDB(t)
	snippetRepo := repository.NewSnippetRepository(db)
	historyRepo := repository.NewHistoryRepository(db)
	fileRepo := repository.NewSnippetFileRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)
	logger := testutil.TestLogger()

	service := NewSnippetService(snippetRepo, logger).
		WithFileRepo(fileRepo).
		WithHistoryRepo(historyRepo).
		WithSettingsRepo(settingsRepo)

	ctx := testutil.TestContext()

	// Create a snippet with files
	input := &models.SnippetInput{
		Title:    "Multi-file Restore",
		Content:  "main",
		Language: "go",
		Files: []models.SnippetFileInput{
			{
				Filename: "main.go",
				Content:  "package main\n\nfunc main() {\n\tprintln(\"v1\")\n}",
				Language: "go",
			},
			{
				Filename: "utils.go",
				Content:  "package utils\n\nfunc Helper() {}",
				Language: "go",
			},
		},
	}
	snippet, err := service.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Get history
	history, err := historyRepo.GetSnippetHistory(ctx, snippet.ID, 50)
	if err != nil {
		t.Fatalf("GetSnippetHistory failed: %v", err)
	}

	if len(history) == 0 {
		t.Fatal("expected at least one history entry")
	}

	// Update files
	updateInput := &models.SnippetInput{
		Title:    "Multi-file Restore",
		Content:  "main",
		Language: "go",
		Files: []models.SnippetFileInput{
			{
				Filename: "main.go",
				Content:  "package main\n\nfunc main() {\n\tprintln(\"v2\")\n}",
				Language: "go",
			},
			{
				Filename: "utils.go",
				Content:  "package utils\n\nfunc Helper() { println(\"updated\") }",
				Language: "go",
			},
		},
	}
	_, err = service.Update(ctx, snippet.ID, updateInput)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Restore from first version
	restored, err := service.RestoreFromHistory(ctx, snippet.ID, history[0].ID)
	if err != nil {
		t.Fatalf("RestoreFromHistory failed: %v", err)
	}

	// Verify files are restored
	if len(restored.Files) != 2 {
		t.Errorf("expected 2 files, got %d", len(restored.Files))
	}

	// Check file content is from original version
	for _, file := range restored.Files {
		if file.Filename == "main.go" {
			if file.Content != "package main\n\nfunc main() {\n\tprintln(\"v1\")\n}" {
				t.Errorf("expected main.go v1 content, got %q", file.Content)
			}
		}
		if file.Filename == "utils.go" {
			if file.Content != "package utils\n\nfunc Helper() {}" {
				t.Errorf("expected utils.go original content, got %q", file.Content)
			}
		}
	}
}

func TestSnippetService_RestoreFromHistory_WrongSnippetID(t *testing.T) {
	db := testutil.TestDB(t)
	snippetRepo := repository.NewSnippetRepository(db)
	historyRepo := repository.NewHistoryRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)
	logger := testutil.TestLogger()

	service := NewSnippetService(snippetRepo, logger).
		WithHistoryRepo(historyRepo).
		WithSettingsRepo(settingsRepo)

	ctx := testutil.TestContext()

	// Create a snippet
	input := &models.SnippetInput{
		Title:    "Test",
		Content:  "test",
		Language: "plaintext",
	}
	snippet, err := service.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Get history
	history, err := historyRepo.GetSnippetHistory(ctx, snippet.ID, 50)
	if err != nil {
		t.Fatalf("GetSnippetHistory failed: %v", err)
	}

	if len(history) == 0 {
		t.Fatal("expected at least one history entry")
	}

	// Try to restore with wrong snippet ID
	_, err = service.RestoreFromHistory(ctx, "wrong-id", history[0].ID)
	if err == nil {
		t.Error("expected error for mismatched snippet ID")
	}
}

func TestSnippetService_RestoreFromHistory_HistoryNotFound(t *testing.T) {
	db := testutil.TestDB(t)
	snippetRepo := repository.NewSnippetRepository(db)
	historyRepo := repository.NewHistoryRepository(db)
	logger := testutil.TestLogger()

	service := NewSnippetService(snippetRepo, logger).
		WithHistoryRepo(historyRepo)

	ctx := testutil.TestContext()

	// Create a snippet
	input := &models.SnippetInput{
		Title:    "Test",
		Content:  "test",
		Language: "plaintext",
	}
	snippet, err := service.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Try to restore with non-existent history ID
	_, err = service.RestoreFromHistory(ctx, snippet.ID, 99999)
	if err == nil {
		t.Error("expected error for non-existent history entry")
	}
}

func TestSnippetService_SaveHistory_WhenHistoryDisabled(t *testing.T) {
	db := testutil.TestDB(t)
	snippetRepo := repository.NewSnippetRepository(db)
	historyRepo := repository.NewHistoryRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)
	logger := testutil.TestLogger()

	ctx := testutil.TestContext()

	// Disable history in settings
	_, err := settingsRepo.Update(ctx, &models.SettingsInput{
		HistoryEnabled: false,
	})
	if err != nil {
		t.Fatalf("Update settings failed: %v", err)
	}

	service := NewSnippetService(snippetRepo, logger).
		WithHistoryRepo(historyRepo).
		WithSettingsRepo(settingsRepo)

	// Create a snippet
	input := &models.SnippetInput{
		Title:    "History Disabled Test",
		Content:  "test",
		Language: "plaintext",
	}
	_, err = service.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify no history was created
	count, err := historyRepo.GetHistoryCount(ctx, input.Title)
	if err != nil {
		// Expected - no history for non-existent snippet with this title
		count = 0
	}

	// Get snippet to find its ID
	snippets, _ := snippetRepo.List(ctx, models.SnippetFilter{Query: "History Disabled Test"})
	if len(snippets.Data) > 0 {
		count, err = historyRepo.GetHistoryCount(ctx, snippets.Data[0].ID)
		if err != nil {
			t.Fatalf("GetHistoryCount failed: %v", err)
		}

		if count != 0 {
			t.Errorf("expected 0 history entries when disabled, got %d", count)
		}
	}
}

func TestSnippetService_SaveHistory_WhenHistoryEnabled(t *testing.T) {
	db := testutil.TestDB(t)
	snippetRepo := repository.NewSnippetRepository(db)
	historyRepo := repository.NewHistoryRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)
	logger := testutil.TestLogger()

	ctx := testutil.TestContext()

	// Enable history in settings
	_, err := settingsRepo.Update(ctx, &models.SettingsInput{
		HistoryEnabled: true,
	})
	if err != nil {
		t.Fatalf("Update settings failed: %v", err)
	}

	service := NewSnippetService(snippetRepo, logger).
		WithHistoryRepo(historyRepo).
		WithSettingsRepo(settingsRepo)

	// Create a snippet
	input := &models.SnippetInput{
		Title:    "History Enabled Test",
		Content:  "test",
		Language: "plaintext",
	}
	_, err = service.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Get snippet to find its ID
	snippets, err := snippetRepo.List(ctx, models.SnippetFilter{Query: "History Enabled Test"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(snippets.Data) == 0 {
		t.Fatal("expected snippet to be created")
	}

	// Verify history was created
	count, err := historyRepo.GetHistoryCount(ctx, snippets.Data[0].ID)
	if err != nil {
		t.Fatalf("GetHistoryCount failed: %v", err)
	}

	if count < 1 {
		t.Errorf("expected at least 1 history entry when enabled, got %d", count)
	}
}

func TestSnippetService_Update_CreatesHistoryEntry(t *testing.T) {
	db := testutil.TestDB(t)
	snippetRepo := repository.NewSnippetRepository(db)
	historyRepo := repository.NewHistoryRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)
	logger := testutil.TestLogger()

	ctx := testutil.TestContext()

	// Enable history
	_, err := settingsRepo.Update(ctx, &models.SettingsInput{
		HistoryEnabled: true,
	})
	if err != nil {
		t.Fatalf("Update settings failed: %v", err)
	}

	service := NewSnippetService(snippetRepo, logger).
		WithHistoryRepo(historyRepo).
		WithSettingsRepo(settingsRepo)

	// Create a snippet
	input := &models.SnippetInput{
		Title:    "Update History Test",
		Content:  "version 1",
		Language: "python",
	}
	snippet, err := service.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Get initial history count
	initialCount, err := historyRepo.GetHistoryCount(ctx, snippet.ID)
	if err != nil {
		t.Fatalf("GetHistoryCount failed: %v", err)
	}

	// Update the snippet
	updateInput := &models.SnippetInput{
		Title:    "Update History Test",
		Content:  "version 2",
		Language: "python",
	}
	_, err = service.Update(ctx, snippet.ID, updateInput)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify new history entry was created
	newCount, err := historyRepo.GetHistoryCount(ctx, snippet.ID)
	if err != nil {
		t.Fatalf("GetHistoryCount failed: %v", err)
	}

	if newCount != initialCount+1 {
		t.Errorf("expected history count to increase by 1, from %d to %d, got %d", initialCount, initialCount+1, newCount)
	}
}
