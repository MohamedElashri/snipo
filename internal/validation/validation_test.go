package validation

import (
	"strings"
	"testing"

	"github.com/MohamedElashri/snipo/internal/models"
)

func TestValidateSnippetInput_Valid(t *testing.T) {
	input := &models.SnippetInput{
		Title:       "Valid Title",
		Description: "A valid description",
		Content:     "console.log('hello');",
		Language:    "javascript",
		Tags:        []string{"test", "example"},
	}

	errs := ValidateSnippetInput(input)
	if errs.HasErrors() {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidateSnippetInput_EmptyTitle(t *testing.T) {
	input := &models.SnippetInput{
		Title:    "",
		Content:  "some content",
		Language: "plaintext",
	}

	errs := ValidateSnippetInput(input)
	if !errs.HasErrors() {
		t.Error("expected error for empty title")
	}

	found := false
	for _, e := range errs {
		if e.Field == "title" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected error on 'title' field")
	}
}

func TestValidateSnippetInput_TitleTooLong(t *testing.T) {
	input := &models.SnippetInput{
		Title:    strings.Repeat("a", 201),
		Content:  "content",
		Language: "plaintext",
	}

	errs := ValidateSnippetInput(input)
	if !errs.HasErrors() {
		t.Error("expected error for title too long")
	}

	found := false
	for _, e := range errs {
		if e.Field == "title" && strings.Contains(e.Message, "200") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected title length error")
	}
}

func TestValidateSnippetInput_EmptyContent(t *testing.T) {
	input := &models.SnippetInput{
		Title:    "Valid Title",
		Content:  "",
		Language: "plaintext",
	}

	errs := ValidateSnippetInput(input)
	if !errs.HasErrors() {
		t.Error("expected error for empty content")
	}

	found := false
	for _, e := range errs {
		if e.Field == "content" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected error on 'content' field")
	}
}

func TestValidateSnippetInput_ContentWithFiles(t *testing.T) {
	// When files are provided, content can be empty
	input := &models.SnippetInput{
		Title:    "Valid Title",
		Content:  "", // Empty content is OK when files exist
		Language: "plaintext",
		Files: []models.SnippetFileInput{
			{Filename: "main.go", Content: "package main", Language: "go"},
		},
	}

	errs := ValidateSnippetInput(input)
	// Should not have content error since files are provided
	for _, e := range errs {
		if e.Field == "content" {
			t.Error("should not have content error when files are provided")
		}
	}
}

func TestValidateSnippetInput_InvalidLanguage(t *testing.T) {
	input := &models.SnippetInput{
		Title:    "Valid Title",
		Content:  "content",
		Language: "invalid-language",
	}

	errs := ValidateSnippetInput(input)
	if !errs.HasErrors() {
		t.Error("expected error for invalid language")
	}

	found := false
	for _, e := range errs {
		if e.Field == "language" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected error on 'language' field")
	}
}

func TestValidateSnippetInput_EmptyLanguage(t *testing.T) {
	input := &models.SnippetInput{
		Title:    "Valid Title",
		Content:  "content",
		Language: "",
	}

	errs := ValidateSnippetInput(input)
	// Empty language should default to plaintext, not error
	if errs.HasErrors() {
		t.Errorf("expected no errors for empty language, got: %v", errs)
	}
	if input.Language != "plaintext" {
		t.Errorf("expected language to default to 'plaintext', got %q", input.Language)
	}
}

func TestValidateSnippetInput_ValidLanguages(t *testing.T) {
	validLanguages := []string{
		"javascript", "typescript", "python", "go", "rust",
		"java", "c", "cpp", "csharp", "php", "ruby",
		"html", "css", "json", "yaml", "sql", "bash",
	}

	for _, lang := range validLanguages {
		input := &models.SnippetInput{
			Title:    "Test",
			Content:  "content",
			Language: lang,
		}

		errs := ValidateSnippetInput(input)
		if errs.HasErrors() {
			t.Errorf("expected no errors for language %q, got: %v", lang, errs)
		}
	}
}

func TestValidateSnippetInput_DescriptionTooLong(t *testing.T) {
	input := &models.SnippetInput{
		Title:       "Valid Title",
		Description: strings.Repeat("a", 501),
		Content:     "content",
		Language:    "plaintext",
	}

	errs := ValidateSnippetInput(input)
	if !errs.HasErrors() {
		t.Error("expected error for description too long")
	}

	found := false
	for _, e := range errs {
		if e.Field == "description" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected error on 'description' field")
	}
}

func TestValidateSnippetInput_ValidTags(t *testing.T) {
	input := &models.SnippetInput{
		Title:    "Valid Title",
		Content:  "content",
		Language: "plaintext",
		Tags:     []string{"valid-tag", "another_tag", "tag123"},
	}

	errs := ValidateSnippetInput(input)
	if errs.HasErrors() {
		t.Errorf("expected no errors for valid tags, got: %v", errs)
	}
}

func TestValidateSnippetInput_InvalidTagCharacters(t *testing.T) {
	input := &models.SnippetInput{
		Title:    "Valid Title",
		Content:  "content",
		Language: "plaintext",
		Tags:     []string{"invalid tag"}, // Space not allowed
	}

	errs := ValidateSnippetInput(input)
	if !errs.HasErrors() {
		t.Error("expected error for invalid tag characters")
	}

	found := false
	for _, e := range errs {
		if e.Field == "tags" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected error on 'tags' field")
	}
}

func TestValidateSnippetInput_TagTooLong(t *testing.T) {
	input := &models.SnippetInput{
		Title:    "Valid Title",
		Content:  "content",
		Language: "plaintext",
		Tags:     []string{strings.Repeat("a", 51)},
	}

	errs := ValidateSnippetInput(input)
	if !errs.HasErrors() {
		t.Error("expected error for tag too long")
	}
}

func TestValidateSnippetInput_FileValidation(t *testing.T) {
	input := &models.SnippetInput{
		Title:    "Valid Title",
		Language: "plaintext",
		Files: []models.SnippetFileInput{
			{Filename: "", Content: "content", Language: "go"}, // Empty filename
		},
	}

	errs := ValidateSnippetInput(input)
	if !errs.HasErrors() {
		t.Error("expected error for empty filename")
	}

	found := false
	for _, e := range errs {
		if e.Field == "files" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected error on 'files' field")
	}
}

func TestValidateSnippetInput_FileLanguageDefault(t *testing.T) {
	input := &models.SnippetInput{
		Title:    "Valid Title",
		Language: "plaintext",
		Files: []models.SnippetFileInput{
			{Filename: "test.txt", Content: "content", Language: ""}, // Empty language
		},
	}

	_ = ValidateSnippetInput(input)
	// Should not error, language should default to plaintext
	if input.Files[0].Language != "plaintext" {
		t.Errorf("expected file language to default to 'plaintext', got %q", input.Files[0].Language)
	}
}

func TestValidateSnippetInput_FileInvalidLanguage(t *testing.T) {
	input := &models.SnippetInput{
		Title:    "Valid Title",
		Language: "plaintext",
		Files: []models.SnippetFileInput{
			{Filename: "test.txt", Content: "content", Language: "invalid-lang"},
		},
	}

	_ = ValidateSnippetInput(input)
	// Invalid language should default to plaintext, not error
	if input.Files[0].Language != "plaintext" {
		t.Errorf("expected invalid file language to default to 'plaintext', got %q", input.Files[0].Language)
	}
}

func TestValidateSnippetInput_TrimWhitespace(t *testing.T) {
	input := &models.SnippetInput{
		Title:       "  Trimmed Title  ",
		Description: "  Trimmed Description  ",
		Content:     "content",
		Language:    "  javascript  ",
		Tags:        []string{"  trimmed-tag  "},
	}

	errs := ValidateSnippetInput(input)
	if errs.HasErrors() {
		t.Errorf("expected no errors, got: %v", errs)
	}

	if input.Title != "Trimmed Title" {
		t.Errorf("expected title to be trimmed, got %q", input.Title)
	}
	if input.Description != "Trimmed Description" {
		t.Errorf("expected description to be trimmed, got %q", input.Description)
	}
	if input.Language != "javascript" {
		t.Errorf("expected language to be trimmed, got %q", input.Language)
	}
	if input.Tags[0] != "trimmed-tag" {
		t.Errorf("expected tag to be trimmed, got %q", input.Tags[0])
	}
}

func TestValidationErrors_Error(t *testing.T) {
	errs := ValidationErrors{
		{Field: "title", Message: "Title is required"},
		{Field: "content", Message: "Content is required"},
	}

	errStr := errs.Error()
	if !strings.Contains(errStr, "title") {
		t.Error("expected error string to contain 'title'")
	}
	if !strings.Contains(errStr, "content") {
		t.Error("expected error string to contain 'content'")
	}
}

func TestValidationErrors_Empty(t *testing.T) {
	errs := ValidationErrors{}

	if errs.HasErrors() {
		t.Error("expected HasErrors to be false for empty errors")
	}
	if errs.Error() != "" {
		t.Error("expected empty error string for empty errors")
	}
}

func TestGetAllowedLanguages(t *testing.T) {
	languages := GetAllowedLanguages()

	if len(languages) == 0 {
		t.Error("expected at least one allowed language")
	}

	// Check some expected languages
	expected := []string{"javascript", "python", "go", "plaintext"}
	for _, lang := range expected {
		found := false
		for _, l := range languages {
			if l == lang {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %q to be in allowed languages", lang)
		}
	}
}
