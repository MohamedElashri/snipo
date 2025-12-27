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
		"html", "css", "json", "yaml", "sql", "bash", "powershell",
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
		Description: strings.Repeat("a", 1001),
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
		Tags:     []string{"valid-tag", "another_tag", "tag123", "hi there"},
	}

	errs := ValidateSnippetInput(input)
	if errs.HasErrors() {
		t.Errorf("expected no errors for valid tags, got: %v", errs)
	}
}

func TestValidateSnippetInput_MultiWordTags(t *testing.T) {
	input := &models.SnippetInput{
		Title:    "Valid Title",
		Content:  "content",
		Language: "plaintext",
		Tags:     []string{"multi word tag", "another tag", "one more"},
	}

	errs := ValidateSnippetInput(input)
	if errs.HasErrors() {
		t.Errorf("expected no errors for multi-word tags, got: %v", errs)
	}
}

func TestValidateSnippetInput_InvalidTagCharacters(t *testing.T) {
	input := &models.SnippetInput{
		Title:    "Valid Title",
		Content:  "content",
		Language: "plaintext",
		Tags:     []string{"invalid@tag"}, // Special characters not allowed
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

// TestValidateSettingsInput tests settings validation
func TestValidateSettingsInput_Valid(t *testing.T) {
	input := &models.SettingsInput{
		AppName:          "My Snipo",
		Theme:            "dark",
		EditorTheme:      "monokai",
		EditorFontSize:   14,
		EditorTabSize:    2,
		MarkdownFontSize: 16,
		DefaultLanguage:  "javascript",
	}

	errs := ValidateSettingsInput(input)
	if errs.HasErrors() {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidateSettingsInput_InvalidTheme(t *testing.T) {
	input := &models.SettingsInput{
		Theme: "invalid-theme",
	}

	errs := ValidateSettingsInput(input)
	if !errs.HasErrors() {
		t.Error("expected error for invalid theme")
	}

	found := false
	for _, e := range errs {
		if e.Field == "theme" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected error on 'theme' field")
	}
}

func TestValidateSettingsInput_InvalidEditorTheme(t *testing.T) {
	input := &models.SettingsInput{
		EditorTheme: "invalid-editor-theme",
	}

	errs := ValidateSettingsInput(input)
	if !errs.HasErrors() {
		t.Error("expected error for invalid editor theme")
	}

	found := false
	for _, e := range errs {
		if e.Field == "editor_theme" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected error on 'editor_theme' field")
	}
}

func TestValidateSettingsInput_FontSizeBoundaries(t *testing.T) {
	tests := []struct {
		name     string
		fontSize int
		wantErr  bool
	}{
		{"too small", 7, true},
		{"min valid", 8, false},
		{"mid range", 16, false},
		{"max valid", 32, false},
		{"too large", 33, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &models.SettingsInput{
				EditorFontSize: tt.fontSize,
			}

			errs := ValidateSettingsInput(input)
			if tt.wantErr && !errs.HasErrors() {
				t.Errorf("expected error for font size %d", tt.fontSize)
			}
			if !tt.wantErr && errs.HasErrors() {
				t.Errorf("unexpected error for font size %d: %v", tt.fontSize, errs)
			}
		})
	}
}

func TestValidateSettingsInput_TabSizeBoundaries(t *testing.T) {
	tests := []struct {
		name    string
		tabSize int
		wantErr bool
	}{
		{"zero (use default)", 0, false}, // 0 means "not set" / use default
		{"min valid", 1, false},
		{"mid range", 4, false},
		{"max valid", 8, false},
		{"too large", 9, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &models.SettingsInput{
				EditorTabSize: tt.tabSize,
			}

			errs := ValidateSettingsInput(input)
			if tt.wantErr && !errs.HasErrors() {
				t.Errorf("expected error for tab size %d", tt.tabSize)
			}
			if !tt.wantErr && errs.HasErrors() {
				t.Errorf("unexpected error for tab size %d: %v", tt.tabSize, errs)
			}
		})
	}
}

func TestValidateSettingsInput_S3Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   *models.SettingsInput
		wantErr bool
	}{
		{
			name: "S3 disabled - no validation",
			input: &models.SettingsInput{
				S3Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "S3 enabled - all fields valid",
			input: &models.SettingsInput{
				S3Enabled:  true,
				S3Endpoint: "s3.amazonaws.com",
				S3Bucket:   "my-bucket",
				S3Region:   "us-east-1",
			},
			wantErr: false,
		},
		{
			name: "S3 enabled - missing endpoint",
			input: &models.SettingsInput{
				S3Enabled: true,
				S3Bucket:  "my-bucket",
				S3Region:  "us-east-1",
			},
			wantErr: true,
		},
		{
			name: "S3 enabled - missing bucket",
			input: &models.SettingsInput{
				S3Enabled:  true,
				S3Endpoint: "s3.amazonaws.com",
				S3Region:   "us-east-1",
			},
			wantErr: true,
		},
		{
			name: "S3 enabled - missing region",
			input: &models.SettingsInput{
				S3Enabled:  true,
				S3Endpoint: "s3.amazonaws.com",
				S3Bucket:   "my-bucket",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateSettingsInput(tt.input)
			if tt.wantErr && !errs.HasErrors() {
				t.Error("expected validation errors")
			}
			if !tt.wantErr && errs.HasErrors() {
				t.Errorf("unexpected errors: %v", errs)
			}
		})
	}
}

// TestValidateTagInput tests tag name validation
func TestValidateTagInput(t *testing.T) {
	tests := []struct {
		name    string
		tagName string
		wantErr bool
	}{
		{"valid tag", "test-tag", false},
		{"valid with underscore", "test_tag", false},
		{"valid alphanumeric", "test123", false},
		{"empty tag", "", true},
		{"too long", strings.Repeat("a", 51), true},
		{"with spaces", "test tag", false},
		{"with special chars", "test@tag", true},
		{"with dots", "test.tag", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateTagInput(tt.tagName)
			if tt.wantErr && !errs.HasErrors() {
				t.Errorf("expected error for tag name %q", tt.tagName)
			}
			if !tt.wantErr && errs.HasErrors() {
				t.Errorf("unexpected error for tag name %q: %v", tt.tagName, errs)
			}
		})
	}
}

// TestValidateFolderInput tests folder name validation
func TestValidateFolderInput(t *testing.T) {
	tests := []struct {
		name       string
		folderName string
		wantErr    bool
	}{
		{"valid folder", "My Folder", false},
		{"valid with special chars", "Folder-2023", false},
		{"empty folder", "", true},
		{"too long", strings.Repeat("a", 101), true},
		{"max length", strings.Repeat("a", 100), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateFolderInput(tt.folderName)
			if tt.wantErr && !errs.HasErrors() {
				t.Errorf("expected error for folder name %q", tt.folderName)
			}
			if !tt.wantErr && errs.HasErrors() {
				t.Errorf("unexpected error for folder name %q: %v", tt.folderName, errs)
			}
		})
	}
}

// TestValidateTokenInput tests token name validation
func TestValidateTokenInput(t *testing.T) {
	tests := []struct {
		name      string
		tokenName string
		wantErr   bool
	}{
		{"valid token", "My API Token", false},
		{"valid with special chars", "Token-2023", false},
		{"empty token", "", true},
		{"too long", strings.Repeat("a", 101), true},
		{"max length", strings.Repeat("a", 100), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateTokenInput(tt.tokenName)
			if tt.wantErr && !errs.HasErrors() {
				t.Errorf("expected error for token name %q", tt.tokenName)
			}
			if !tt.wantErr && errs.HasErrors() {
				t.Errorf("unexpected error for token name %q: %v", tt.tokenName, errs)
			}
		})
	}
}

// TestValidateFilename tests filename validation
func TestValidateFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{"valid filename", "main.go", false},
		{"valid with dashes", "my-file.js", false},
		{"empty filename", "", true},
		{"too long", strings.Repeat("a", 256), true},
		{"max length", strings.Repeat("a", 255), false},
		{"path traversal", "../etc/passwd", true},
		{"forward slash", "path/to/file.txt", true},
		{"backslash", "path\\to\\file.txt", true},
		{"double dots", "file..txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateFilename(tt.filename)
			if tt.wantErr && !errs.HasErrors() {
				t.Errorf("expected error for filename %q", tt.filename)
			}
			if !tt.wantErr && errs.HasErrors() {
				t.Errorf("unexpected error for filename %q: %v", tt.filename, errs)
			}
		})
	}
}

// TestSnippetInput_SQLInjectionPrevention tests that validation doesn't allow SQL injection patterns
func TestSnippetInput_SQLInjectionPrevention(t *testing.T) {
	// The validation doesn't specifically block SQL injection patterns because:
	// 1. We use parameterized queries in the repository layer (proper defense)
	// 2. These are just text fields that users might legitimately want to include SQL in snippets
	// However, we do validate length limits to prevent abuse

	sqlInjectionPatterns := []string{
		"'; DROP TABLE snippets; --",
		"1' OR '1'='1",
		"admin'--",
		"' UNION SELECT * FROM users --",
	}

	for _, pattern := range sqlInjectionPatterns {
		input := &models.SnippetInput{
			Title:    "Test Snippet",
			Content:  pattern, // Should be allowed in content (it's code!)
			Language: "sql",
		}

		errs := ValidateSnippetInput(input)
		// Should not have errors - SQL content is legitimate for code snippets
		if errs.HasErrors() {
			t.Errorf("validation should not block SQL patterns in content: %v", errs)
		}
	}
}

// TestSnippetInput_XSSPrevention tests XSS pattern handling
func TestSnippetInput_XSSPrevention(t *testing.T) {
	// Similar to SQL injection, XSS content is legitimate in code snippets
	// The frontend/template rendering layer should handle escaping, not validation

	xssPatterns := []string{
		"<script>alert('XSS')</script>",
		"<img src=x onerror=alert('XSS')>",
		"javascript:alert('XSS')",
		"<iframe src='javascript:alert(1)'>",
	}

	for _, pattern := range xssPatterns {
		input := &models.SnippetInput{
			Title:    "XSS Test",
			Content:  pattern, // Should be allowed in content (might be HTML/JS snippets!)
			Language: "html",
		}

		errs := ValidateSnippetInput(input)
		// Should not have errors - XSS content is legitimate for code snippets
		if errs.HasErrors() {
			t.Errorf("validation should not block XSS patterns in content: %v", errs)
		}
	}

	// However, we should enforce length limits
	longXSS := "<script>" + strings.Repeat("alert('x');", 100000) + "</script>"
	input := &models.SnippetInput{
		Title:    "Test",
		Content:  longXSS,
		Language: "html",
	}

	errs := ValidateSnippetInput(input)
	if !errs.HasErrors() {
		t.Error("expected error for content exceeding size limit")
	}
}

// TestSnippetInput_DescriptionLengthUpdate tests new 1000 char limit
func TestSnippetInput_DescriptionLengthUpdate(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		wantErr bool
	}{
		{"under limit", 999, false},
		{"at limit", 1000, false},
		{"over limit", 1001, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &models.SnippetInput{
				Title:       "Test",
				Description: strings.Repeat("a", tt.length),
				Content:     "content",
				Language:    "plaintext",
			}

			errs := ValidateSnippetInput(input)
			hasDescErr := false
			for _, e := range errs {
				if e.Field == "description" {
					hasDescErr = true
					break
				}
			}

			if tt.wantErr && !hasDescErr {
				t.Errorf("expected description error for length %d", tt.length)
			}
			if !tt.wantErr && hasDescErr {
				t.Errorf("unexpected description error for length %d", tt.length)
			}
		})
	}
}
