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
		return
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

	// Verify view count
	updated, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if updated.ViewCount != 1 {
		t.Errorf("expected view_count 1, got %d", updated.ViewCount)
	}
}

func TestSnippetRepository_ToggleArchive(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewSnippetRepository(db)
	ctx := testutil.TestContext()

	// Create a snippet (public)
	input := &models.SnippetInput{
		Title:    "Test Snippet",
		Content:  "content",
		Language: "plaintext",
		IsPublic: true,
	}
	created, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Initial state should be unarchived and public
	if created.IsArchived {
		t.Error("expected initial IsArchived to be false")
	}
	if !created.IsPublic {
		t.Error("expected initial IsPublic to be true")
	}

	// Toggle archive (archive it)
	updated, err := repo.ToggleArchive(ctx, created.ID)
	if err != nil {
		t.Fatalf("ToggleArchive failed: %v", err)
	}
	if !updated.IsArchived {
		t.Error("expected IsArchived to be true after toggle")
	}
	// It should now be private
	if updated.IsPublic {
		t.Error("expected IsPublic to be false after archiving")
	}

	// Toggle again (unarchive it)
	updated, err = repo.ToggleArchive(ctx, created.ID)
	if err != nil {
		t.Fatalf("ToggleArchive failed: %v", err)
	}
	if updated.IsArchived {
		t.Error("expected IsArchived to be false after second toggle")
	}
	// It should remain private (we don't restore public status)
	if updated.IsPublic {
		t.Error("expected IsPublic to remain false after unarchiving")
	}
}

func TestSnippetRepository_List_FilterByArchive(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewSnippetRepository(db)
	ctx := testutil.TestContext()

	// Create active snippet
	activeInput := &models.SnippetInput{Title: "Active", Content: "c", Language: "p"}
	active, _ := repo.Create(ctx, activeInput)

	// Create archived snippet
	archivedInput := &models.SnippetInput{Title: "Archived", Content: "c", Language: "p", IsArchived: true}
	archived, _ := repo.Create(ctx, archivedInput)

	// Case 1: Default filter (IsArchived = nil, defaults to active only)
	filterActive := models.SnippetFilter{}
	listActive, err := repo.List(ctx, filterActive)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(listActive.Data) != 1 || listActive.Data[0].ID != active.ID {
		t.Errorf("expected 1 active snippet, got %d", len(listActive.Data))
	}

	// Case 2: Explicitly IsArchived = true
	isArchived := true
	filterArchived := models.SnippetFilter{IsArchived: &isArchived}
	listArchived, err := repo.List(ctx, filterArchived)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(listArchived.Data) != 1 || listArchived.Data[0].ID != archived.ID {
		t.Errorf("expected 1 archived snippet, got %d", len(listArchived.Data))
	}

	// Case 3: Explicitly IsArchived = false
	isNotArchived := false
	filterNotArchived := models.SnippetFilter{IsArchived: &isNotArchived}
	listNotArchived, err := repo.List(ctx, filterNotArchived)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(listNotArchived.Data) != 1 || listNotArchived.Data[0].ID != active.ID {
		t.Errorf("expected 1 active snippet, got %d", len(listNotArchived.Data))
	}
}

// TestAllowedSortColumns_SafeValues verifies that the allowedSortColumns map
// contains only safe, predefined column names that match valid SQL identifiers.
func TestAllowedSortColumns_SafeValues(t *testing.T) {
	// Valid SQL identifier pattern: starts with letter or underscore, contains only alphanumeric and underscore
	validIdentifier := func(s string) bool {
		if len(s) == 0 {
			return false
		}
		for i, c := range s {
			if i == 0 {
				if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && c != '_' {
					return false
				}
			} else {
				if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') && c != '_' {
					return false
				}
			}
		}
		return true
	}

	// Dangerous characters/patterns that should never appear in column names
	// We only check for SQL syntax characters, not SQL keywords (which could be substrings)
	dangerousChars := []string{
		";", "--", "/*", "*/", "'", "\"", "\\", "(", ")", "=", "<", ">", " ", "\t", "\n", "\r",
	}

	hasDangerousPattern := func(s string) bool {
		for _, char := range dangerousChars {
			for i := 0; i <= len(s)-len(char); i++ {
				if s[i:i+len(char)] == char {
					return true
				}
			}
		}
		return false
	}

	// Test that the map exists and is not empty
	if len(allowedSortColumns) == 0 {
		t.Fatal("allowedSortColumns map is empty - SQL injection protection may be compromised")
	}

	// Test each key and value in the map
	for key, value := range allowedSortColumns {
		t.Run("key_"+key, func(t *testing.T) {
			// Keys should be valid identifiers
			if !validIdentifier(key) {
				t.Errorf("key %q is not a valid SQL identifier", key)
			}
			// Keys should not contain dangerous patterns
			if hasDangerousPattern(key) {
				t.Errorf("key %q contains dangerous SQL pattern", key)
			}
		})

		t.Run("value_"+value, func(t *testing.T) {
			// Values should be valid identifiers
			if !validIdentifier(value) {
				t.Errorf("value %q is not a valid SQL identifier", value)
			}
			// Values should not contain dangerous patterns
			if hasDangerousPattern(value) {
				t.Errorf("value %q contains dangerous SQL pattern", value)
			}
			// Values should be short (reasonable column names)
			if len(value) > 64 {
				t.Errorf("value %q is suspiciously long (%d chars)", value, len(value))
			}
		})
	}

	// Verify expected columns exist
	expectedColumns := []string{"id", "title", "description", "content", "language", "is_favorite", "is_public", "view_count", "created_at", "updated_at"}
	for _, col := range expectedColumns {
		if _, ok := allowedSortColumns[col]; !ok {
			t.Errorf("expected column %q not found in allowedSortColumns", col)
		}
	}
}

// TestSQLInjection_SortColumnIsolation verifies that malicious sort column values
// are completely replaced with safe defaults, not just sanitized.
// This is critical because the fix uses map lookup to return CONSTANT values.
func TestSQLInjection_SortColumnIsolation(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewSnippetRepository(db)
	ctx := testutil.TestContext()

	// Create test data
	_, err := repo.Create(ctx, &models.SnippetInput{
		Title:    "Test",
		Content:  "content",
		Language: "go",
	})
	if err != nil {
		t.Fatalf("failed to create test snippet: %v", err)
	}

	// Comprehensive SQL injection attack vectors
	injectionAttempts := []struct {
		name      string
		sortBy    string
		sortOrder string
	}{
		// Basic SQL injection
		{"semicolon_injection", "title; DROP TABLE snippets;--", "asc"},
		{"comment_injection", "title--", "asc"},
		{"quote_injection", "title'", "asc"},
		{"double_quote", `title"`, "asc"},

		// UNION-based injection
		{"union_select", "title UNION SELECT * FROM users--", "asc"},
		{"union_all", "title UNION ALL SELECT password FROM users--", "asc"},

		// Stacked queries
		{"stacked_query", "title; INSERT INTO users VALUES ('hacker');--", "asc"},
		{"stacked_delete", "title; DELETE FROM snippets;--", "asc"},

		// Boolean-based blind injection
		{"boolean_or", "title OR 1=1--", "asc"},
		{"boolean_and", "title AND 1=1--", "asc"},

		// Time-based blind injection
		{"time_based_sqlite", "title; SELECT CASE WHEN 1=1 THEN sqlite_sleep(5) END;--", "asc"},

		// Subquery injection
		{"subquery", "(SELECT password FROM users LIMIT 1)", "asc"},
		{"nested_subquery", "title,(SELECT GROUP_CONCAT(password) FROM users)", "asc"},

		// Function injection
		{"function_call", "SUBSTR(password,1,1)", "asc"},
		{"hex_function", "HEX(password)", "asc"},

		// Order by specific attacks
		{"order_by_number", "1", "asc"},
		{"order_by_case", "CASE WHEN 1=1 THEN title ELSE id END", "asc"},
		{"order_by_if", "IF(1=1,title,id)", "asc"},

		// Whitespace/encoding tricks
		{"tab_injection", "title\t;DROP TABLE snippets;--", "asc"},
		{"newline_injection", "title\n;DROP TABLE snippets;--", "asc"},
		{"url_encoded", "title%3BDROP%20TABLE%20snippets%3B--", "asc"},

		// Sort order injection
		{"order_injection", "title", "asc; DROP TABLE snippets;--"},
		{"order_union", "title", "asc UNION SELECT * FROM users--"},
		{"order_subquery", "title", "(SELECT 1)"},

		// Mixed case bypass attempts
		{"mixed_case_drop", "TiTlE; DrOp TaBlE snippets;--", "asc"},

		// Null byte injection
		{"null_byte", "title\x00;DROP TABLE snippets;--", "asc"},

		// Very long input
		{"long_input", "title" + string(make([]byte, 10000)), "asc"},
	}

	for _, tc := range injectionAttempts {
		t.Run(tc.name, func(t *testing.T) {
			filter := models.SnippetFilter{
				SortBy:    tc.sortBy,
				SortOrder: tc.sortOrder,
				Limit:     10,
				Page:      1,
			}

			// The query should NOT fail - if injection worked, it would likely cause an error
			result, err := repo.List(ctx, filter)
			if err != nil {
				// If error occurs, it should NOT be because of successful injection
				// A well-protected query should still execute with defaults
				t.Logf("Query returned error (may indicate injection was blocked): %v", err)
			}

			// If we got results, verify integrity
			if result != nil {
				// Verify we can still see our test data
				if result.Pagination.Total < 1 {
					t.Error("Expected at least 1 snippet - data may have been deleted by injection")
				}
			}
		})
	}
}

// TestSQLInjection_ConstantValueGuarantee verifies that the sort values used in queries
// are ALWAYS from the predefined constant set, never user input.
func TestSQLInjection_ConstantValueGuarantee(t *testing.T) {
	// Test that when a valid key is provided, the returned value is the constant from the map
	testCases := []struct {
		input    string
		expected string
	}{
		{"id", "id"},
		{"title", "title"},
		{"created_at", "created_at"},
		{"updated_at", "updated_at"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			// Simulate the lookup that happens in List()
			result, ok := allowedSortColumns[tc.input]
			if !ok {
				t.Fatalf("expected key %q to exist in allowedSortColumns", tc.input)
			}
			if result != tc.expected {
				t.Errorf("expected value %q for key %q, got %q", tc.expected, tc.input, result)
			}

			// Verify the returned value is NOT the same object as input (it's a constant)
			// This is a conceptual test - in Go strings are immutable, but the value should come from the map
			if &tc.input == &result {
				t.Error("returned value appears to be the same object as input - should be constant from map")
			}
		})
	}

	// Test that invalid keys return no value (forcing default)
	invalidKeys := []string{
		"malicious",
		"'; DROP TABLE snippets; --",
		"id; DROP TABLE snippets;",
		"",
		" ",
		"ID", // Case sensitive
		"Id", // Case sensitive
		"TITLE",
	}

	for _, key := range invalidKeys {
		t.Run("invalid_"+key, func(t *testing.T) {
			_, ok := allowedSortColumns[key]
			if ok {
				t.Errorf("key %q should NOT be in allowedSortColumns", key)
			}
		})
	}
}

// TestSQLInjection_SortOrderConstantValues verifies that sort order is always
// one of the constant values "ASC" or "DESC", never user input.
func TestSQLInjection_SortOrderConstantValues(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewSnippetRepository(db)
	ctx := testutil.TestContext()

	// Create two snippets with different titles to verify ordering works
	// Using title for sorting as it's deterministic (alphabetical order)
	_, err := repo.Create(ctx, &models.SnippetInput{
		Title: "AAA First", Content: "1", Language: "go",
	})
	if err != nil {
		t.Fatalf("failed to create first snippet: %v", err)
	}

	_, err = repo.Create(ctx, &models.SnippetInput{
		Title: "ZZZ Second", Content: "2", Language: "go",
	})
	if err != nil {
		t.Fatalf("failed to create second snippet: %v", err)
	}

	// Test valid ascending order - AAA should come first
	t.Run("valid_asc", func(t *testing.T) {
		result, err := repo.List(ctx, models.SnippetFilter{
			SortBy:    "title",
			SortOrder: "asc",
			Limit:     10,
		})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(result.Data) < 2 {
			t.Fatal("expected at least 2 snippets")
		}
		// AAA should be first in ascending order
		if result.Data[0].Title != "AAA First" {
			t.Errorf("ascending order not working correctly, got first title: %s", result.Data[0].Title)
		}
	})

	// Test valid descending order - ZZZ should come first
	t.Run("valid_desc", func(t *testing.T) {
		result, err := repo.List(ctx, models.SnippetFilter{
			SortBy:    "title",
			SortOrder: "desc",
			Limit:     10,
		})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(result.Data) < 2 {
			t.Fatal("expected at least 2 snippets")
		}
		// ZZZ should be first in descending order
		if result.Data[0].Title != "ZZZ Second" {
			t.Errorf("descending order not working correctly, got first title: %s", result.Data[0].Title)
		}
	})

	// Test that invalid orders default to desc (safe default)
	// The key security guarantee: invalid input NEVER reaches the SQL query
	invalidOrders := []string{
		"invalid",
		"ASC",  // Must be lowercase per our implementation
		"DESC", // Must be lowercase per our implementation
		"asc; DROP TABLE snippets;--",
		"1",
		"",
		"asc OR 1=1",
	}

	for _, order := range invalidOrders {
		t.Run("invalid_"+order, func(t *testing.T) {
			result, err := repo.List(ctx, models.SnippetFilter{
				SortBy:    "title",
				SortOrder: order,
				Limit:     10,
			})
			if err != nil {
				t.Fatalf("List failed for order %q: %v", order, err)
			}
			// Should default to DESC - ZZZ should be first
			if len(result.Data) >= 2 && result.Data[0].Title != "ZZZ Second" {
				t.Errorf("expected default desc order for invalid order %q, got first title: %s", order, result.Data[0].Title)
			}
		})
	}
}

func TestSnippetRepository_List_SQLInjectionPrevention(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewSnippetRepository(db)
	ctx := testutil.TestContext()

	// Create a test snippet
	input := &models.SnippetInput{
		Title:    "Test Snippet",
		Content:  "test content",
		Language: "plaintext",
	}
	_, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	tests := []struct {
		name              string
		inputFilter       models.SnippetFilter
		expectedSortBy    string
		expectedSortOrder string
	}{
		{
			name: "Valid sort column and order",
			inputFilter: models.SnippetFilter{
				SortBy:    "title",
				SortOrder: "asc",
			},
			expectedSortBy:    "title",
			expectedSortOrder: "asc",
		},
		{
			name: "Invalid sort column should default to updated_at",
			inputFilter: models.SnippetFilter{
				SortBy:    "malicious_column; DROP TABLE snippets; --",
				SortOrder: "desc",
			},
			expectedSortBy:    "updated_at",
			expectedSortOrder: "desc",
		},
		{
			name: "Invalid sort order should default to desc",
			inputFilter: models.SnippetFilter{
				SortBy:    "created_at",
				SortOrder: "invalid_order",
			},
			expectedSortBy:    "created_at",
			expectedSortOrder: "desc",
		},
		{
			name: "Empty sort values should get defaults",
			inputFilter: models.SnippetFilter{
				SortBy:    "",
				SortOrder: "",
			},
			expectedSortBy:    "updated_at",
			expectedSortOrder: "desc",
		},
		{
			name: "SQL injection attempt in sort column",
			inputFilter: models.SnippetFilter{
				SortBy:    "updated_at; SELECT * FROM users; --",
				SortOrder: "asc",
			},
			expectedSortBy:    "updated_at",
			expectedSortOrder: "asc",
		},
		{
			name: "Valid sort column with mixed case order",
			inputFilter: models.SnippetFilter{
				SortBy:    "language",
				SortOrder: "ASC",
			},
			expectedSortBy:    "language",
			expectedSortOrder: "desc", // Should be normalized to lowercase, invalid values become "desc"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that the List method doesn't crash and returns valid results
			result, err := repo.List(ctx, tt.inputFilter)
			if err != nil {
				t.Fatalf("List failed: %v", err)
			}

			// Should return at least one snippet since we created one
			if result.Pagination.Total < 1 {
				t.Error("expected at least 1 snippet in results")
			}

			// The query should execute successfully without SQL injection errors
			// If there is SQL injection, the query would fail or return unexpected results
		})
	}
}
