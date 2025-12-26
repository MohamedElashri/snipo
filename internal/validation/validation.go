package validation

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/MohamedElashri/snipo/internal/models"
)

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return ""
	}
	var msgs []string
	for _, e := range ve {
		msgs = append(msgs, e.Field+": "+e.Message)
	}
	return strings.Join(msgs, "; ")
}

// HasErrors returns true if there are validation errors
func (ve ValidationErrors) HasErrors() bool {
	return len(ve) > 0
}

// Allowed languages (subset of Prism.js supported languages)
var allowedLanguages = map[string]bool{
	"plaintext": true, "javascript": true, "typescript": true, "python": true,
	"go": true, "rust": true, "java": true, "c": true, "cpp": true, "csharp": true,
	"php": true, "ruby": true, "swift": true, "kotlin": true, "scala": true,
	"html": true, "css": true, "scss": true, "json": true, "yaml": true, "xml": true,
	"markdown": true, "sql": true, "bash": true, "shell": true, "powershell": true,
	"dockerfile": true, "nginx": true, "toml": true, "ini": true, "makefile": true,
	"lua": true, "perl": true, "r": true, "haskell": true, "elixir": true,
	"clojure": true, "graphql": true, "protobuf": true, "terraform": true,
}

// tagRegex validates tag names
var tagRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// ValidateSnippetInput validates snippet input
func ValidateSnippetInput(input *models.SnippetInput) ValidationErrors {
	var errs ValidationErrors

	// Title validation
	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		errs = append(errs, ValidationError{Field: "title", Message: "Title is required"})
	} else if utf8.RuneCountInString(input.Title) > 200 {
		errs = append(errs, ValidationError{Field: "title", Message: "Title must be less than 200 characters"})
	}

	// Content validation (skip if multi-file snippet with files)
	hasFiles := len(input.Files) > 0
	if !hasFiles && strings.TrimSpace(input.Content) == "" {
		errs = append(errs, ValidationError{Field: "content", Message: "Content is required"})
	} else if len(input.Content) > 1024*1024 { // 1MB limit
		errs = append(errs, ValidationError{Field: "content", Message: "Content must be less than 1MB"})
	}

	// Validate files if present
	for i, file := range input.Files {
		// Sanitize filename by replacing spaces with underscores
		input.Files[i].Filename = SanitizeFilename(file.Filename)
		
		if strings.TrimSpace(file.Filename) == "" {
			errs = append(errs, ValidationError{Field: "files", Message: "Filename is required for all files"})
		}
		if len(file.Content) > 1024*1024 { // 1MB limit per file
			errs = append(errs, ValidationError{Field: "files", Message: "File content must be less than 1MB each"})
		}
		// Validate file language
		lang := strings.ToLower(strings.TrimSpace(file.Language))
		if lang == "" {
			input.Files[i].Language = "plaintext"
		} else if !allowedLanguages[lang] {
			input.Files[i].Language = "plaintext" // Default to plaintext if invalid
		}
	}

	// Language validation
	input.Language = strings.ToLower(strings.TrimSpace(input.Language))
	if input.Language == "" {
		input.Language = "plaintext"
	} else if !allowedLanguages[input.Language] {
		errs = append(errs, ValidationError{Field: "language", Message: "Invalid language"})
	}

	// Description length
	input.Description = strings.TrimSpace(input.Description)
	if utf8.RuneCountInString(input.Description) > 1000 {
		errs = append(errs, ValidationError{Field: "description", Message: "Description must be less than 1000 characters"})
	}

	// Tag validation
	for i, tag := range input.Tags {
		tag = strings.TrimSpace(tag)
		input.Tags[i] = tag
		if tag == "" {
			continue
		}
		if len(tag) > 50 {
			errs = append(errs, ValidationError{Field: "tags", Message: "Tag name must be less than 50 characters"})
		} else if !tagRegex.MatchString(tag) {
			errs = append(errs, ValidationError{Field: "tags", Message: "Tag can only contain letters, numbers, underscores, and hyphens"})
		}
	}

	// Validate filenames in files
	for _, file := range input.Files {
		if fileErrs := ValidateFilename(file.Filename); fileErrs.HasErrors() {
			errs = append(errs, fileErrs...)
		}
	}

	return errs
}

// GetAllowedLanguages returns a list of allowed language identifiers
func GetAllowedLanguages() []string {
	languages := make([]string, 0, len(allowedLanguages))
	for lang := range allowedLanguages {
		languages = append(languages, lang)
	}
	return languages
}

// Allowed editor themes
var allowedEditorThemes = map[string]bool{
	"chaos": true, "clouds": true, "clouds_midnight": true, "cobalt": true,
	"crimson_editor": true, "dawn": true, "dracula": true, "dreamweaver": true,
	"eclipse": true, "github": true, "gob": true, "gruvbox": true, "idle_fingers": true,
	"iplastic": true, "katzenmilch": true, "kr_theme": true, "kuroir": true,
	"merbivore": true, "merbivore_soft": true, "monokai": true, "mono_industrial": true,
	"nord_dark": true, "one_dark": true, "pastel_on_dark": true, "solarized_dark": true,
	"solarized_light": true, "sqlserver": true, "terminal": true, "textmate": true,
	"tomorrow": true, "tomorrow_night": true, "tomorrow_night_blue": true,
	"tomorrow_night_bright": true, "tomorrow_night_eighties": true, "twilight": true,
	"vibrant_ink": true, "xcode": true,
}

// Allowed UI themes
var allowedUIThemes = map[string]bool{
	"light": true,
	"dark":  true,
}

// ValidateSettingsInput validates settings input
func ValidateSettingsInput(input *models.SettingsInput) ValidationErrors {
	var errs ValidationErrors

	// App name validation
	input.AppName = strings.TrimSpace(input.AppName)
	if input.AppName != "" && utf8.RuneCountInString(input.AppName) > 100 {
		errs = append(errs, ValidationError{Field: "app_name", Message: "App name must be less than 100 characters"})
	}

	// Theme validation (UI theme)
	input.Theme = strings.ToLower(strings.TrimSpace(input.Theme))
	if input.Theme != "" && !allowedUIThemes[input.Theme] {
		errs = append(errs, ValidationError{Field: "theme", Message: "Theme must be 'light' or 'dark'"})
	}

	// Editor theme validation
	input.EditorTheme = strings.ToLower(strings.TrimSpace(input.EditorTheme))
	if input.EditorTheme != "" && !allowedEditorThemes[input.EditorTheme] {
		errs = append(errs, ValidationError{Field: "editor_theme", Message: "Invalid editor theme"})
	}

	// Editor font size validation (8-32)
	if input.EditorFontSize != 0 && (input.EditorFontSize < 8 || input.EditorFontSize > 32) {
		errs = append(errs, ValidationError{Field: "editor_font_size", Message: "Editor font size must be between 8 and 32"})
	}

	// Editor tab size validation (1-8)
	if input.EditorTabSize != 0 && (input.EditorTabSize < 1 || input.EditorTabSize > 8) {
		errs = append(errs, ValidationError{Field: "editor_tab_size", Message: "Editor tab size must be between 1 and 8"})
	}

	// Markdown font size validation (8-32)
	if input.MarkdownFontSize != 0 && (input.MarkdownFontSize < 8 || input.MarkdownFontSize > 32) {
		errs = append(errs, ValidationError{Field: "markdown_font_size", Message: "Markdown font size must be between 8 and 32"})
	}

	// Default language validation
	input.DefaultLanguage = strings.ToLower(strings.TrimSpace(input.DefaultLanguage))
	if input.DefaultLanguage != "" && !allowedLanguages[input.DefaultLanguage] {
		errs = append(errs, ValidationError{Field: "default_language", Message: "Invalid default language"})
	}

	// S3 configuration validation
	if input.S3Enabled {
		input.S3Endpoint = strings.TrimSpace(input.S3Endpoint)
		input.S3Bucket = strings.TrimSpace(input.S3Bucket)
		input.S3Region = strings.TrimSpace(input.S3Region)

		if input.S3Endpoint == "" {
			errs = append(errs, ValidationError{Field: "s3_endpoint", Message: "S3 endpoint is required when S3 is enabled"})
		}
		if input.S3Bucket == "" {
			errs = append(errs, ValidationError{Field: "s3_bucket", Message: "S3 bucket is required when S3 is enabled"})
		}
		if input.S3Region == "" {
			errs = append(errs, ValidationError{Field: "s3_region", Message: "S3 region is required when S3 is enabled"})
		}
	}

	return errs
}

// ValidateTagInput validates tag input
func ValidateTagInput(name string) ValidationErrors {
	var errs ValidationErrors

	name = strings.TrimSpace(name)
	if name == "" {
		errs = append(errs, ValidationError{Field: "name", Message: "Tag name is required"})
	} else if len(name) > 50 {
		errs = append(errs, ValidationError{Field: "name", Message: "Tag name must be less than 50 characters"})
	} else if !tagRegex.MatchString(name) {
		errs = append(errs, ValidationError{Field: "name", Message: "Tag can only contain letters, numbers, underscores, and hyphens"})
	}

	return errs
}

// ValidateFolderInput validates folder input
func ValidateFolderInput(name string) ValidationErrors {
	var errs ValidationErrors

	name = strings.TrimSpace(name)
	if name == "" {
		errs = append(errs, ValidationError{Field: "name", Message: "Folder name is required"})
	} else if utf8.RuneCountInString(name) > 100 {
		errs = append(errs, ValidationError{Field: "name", Message: "Folder name must be less than 100 characters"})
	}

	return errs
}

// ValidateTokenInput validates API token input
func ValidateTokenInput(name string) ValidationErrors {
	var errs ValidationErrors

	name = strings.TrimSpace(name)
	if name == "" {
		errs = append(errs, ValidationError{Field: "name", Message: "Token name is required"})
	} else if utf8.RuneCountInString(name) > 100 {
		errs = append(errs, ValidationError{Field: "name", Message: "Token name must be less than 100 characters"})
	}

	return errs
}

// SanitizeFilename removes or replaces problematic characters in filenames
func SanitizeFilename(filename string) string {
	filename = strings.TrimSpace(filename)
	
	// Replace spaces with underscores
	filename = strings.ReplaceAll(filename, " ", "_")
	
	// Replace multiple underscores with single underscore
	for strings.Contains(filename, "__") {
		filename = strings.ReplaceAll(filename, "__", "_")
	}
	
	return filename
}

// ValidateFilename validates a filename for length and basic safety
func ValidateFilename(filename string) ValidationErrors {
	var errs ValidationErrors

	filename = strings.TrimSpace(filename)
	if filename == "" {
		errs = append(errs, ValidationError{Field: "filename", Message: "Filename is required"})
	} else if utf8.RuneCountInString(filename) > 255 {
		errs = append(errs, ValidationError{Field: "filename", Message: "Filename must be less than 255 characters"})
	} else if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		errs = append(errs, ValidationError{Field: "filename", Message: "Filename contains invalid characters"})
	}

	return errs
}
