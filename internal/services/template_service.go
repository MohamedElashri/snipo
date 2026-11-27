package services

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/MohamedElashri/snipo/internal/repository"
	"github.com/MohamedElashri/snipo/internal/validation"
	"github.com/MohamedElashri/snipo/pkg/models"
)

const maxTemplateOutput = 1024 * 1024

var (
	templateKey   = regexp.MustCompile(`^[a-z][a-z0-9_]{0,47}$`)
	templateID    = regexp.MustCompile(`^[a-f0-9]{32}$`)
	templateToken = regexp.MustCompile(`\{\{(!?)snipo\.([a-z][a-z0-9_]{0,47})\}\}|{{!?snipo\.`)
)

type TemplateService struct {
	*repository.TemplateRepository
	folders  *repository.FolderRepository
	maxFiles int
}

func NewTemplateService(repo *repository.TemplateRepository, folders *repository.FolderRepository, maxFiles int) *TemplateService {
	if maxFiles <= 0 {
		maxFiles = 10
	}
	return &TemplateService{TemplateRepository: repo, folders: folders, maxFiles: maxFiles}
}

func templateError(field, message string) validation.ValidationErrors {
	return validation.ValidationErrors{{Field: field, Message: message}}
}

func validTemplateValue(f models.TemplateField, value string) bool {
	if len(value) > maxTemplateOutput {
		return false
	}
	if value == "" {
		return true
	}
	switch f.Type {
	case "multiline":
		return true
	case "text":
		return !strings.ContainsAny(value, "\r\n")
	case "choice":
		return slices.Contains(f.Choices, value)
	case "number":
		n, err := strconv.ParseFloat(value, 64)
		return err == nil && !math.IsNaN(n) && !math.IsInf(n, 0)
	case "toggle":
		return value == "true" || value == "false"
	case "date":
		_, err := time.Parse(time.DateOnly, value)
		return err == nil
	}
	return false
}

func (s *TemplateService) Validate(def *models.TemplateDefinition) error {
	return validateTemplate(def, s.maxFiles)
}

// A backup may have been created on an instance with a different file limit.
func validateTemplate(def *models.TemplateDefinition, maxFiles int) error {
	def.Name = strings.TrimSpace(def.Name)
	if def.Name == "" || utf8.RuneCountInString(def.Name) > 200 {
		return templateError("name", "Enter a template name of up to 200 characters")
	}
	if strings.TrimSpace(def.Title) == "" || len(def.Purpose) > 4000 || len(def.Description) > 4000 || len(def.Title) > 2000 {
		return templateError("title", "Enter an output title and keep template descriptions within the editor limits")
	}
	if len(def.Files) == 0 {
		return templateError("files", "Add at least one file")
	}
	if maxFiles > 0 && len(def.Files) > maxFiles {
		return templateError("files", fmt.Sprintf("Choose up to %d files", maxFiles))
	}
	if len(def.Fields) > 50 || len(def.Presets) > 50 {
		return templateError("fields", "A template can have up to 50 fields and 50 presets")
	}
	fields := make(map[string]models.TemplateField, len(def.Fields))
	samples := make(map[string]string, len(def.Fields))
	for i, field := range def.Fields {
		path := fmt.Sprintf("fields.%d", i)
		if !templateKey.MatchString(field.Key) || fields[field.Key].Key != "" || strings.TrimSpace(field.Label) == "" || len(field.Label) > 200 || len(field.Help) > 1000 {
			return templateError(path, "Fields need a unique lower-case key and a short label")
		}
		if !slices.Contains([]string{"text", "multiline", "choice", "number", "toggle", "date"}, field.Type) {
			return templateError(path, "Choose a supported field type")
		}
		if field.Source != "" && !slices.Contains([]string{"selection", "page_title", "url"}, field.Source) {
			return templateError(path, "Choose a supported browser capture source")
		}
		if field.Type == "choice" && (len(field.Choices) == 0 || len(field.Choices) > 50 || slices.Contains(field.Choices, "")) {
			return templateError(path, "Enter between 1 and 50 nonempty choices")
		}
		if field.Default != nil && !validTemplateValue(field, *field.Default) {
			return templateError(path, "The default value must match the field type")
		}
		fields[field.Key] = field
		samples[field.Key] = "x"
	}
	presets := make(map[string]bool, len(def.Presets))
	for i, preset := range def.Presets {
		path := fmt.Sprintf("presets.%d", i)
		name := strings.ToLower(strings.TrimSpace(preset.Name))
		if name == "" || len(name) > 200 || presets[name] {
			return templateError(path, "Presets need a unique name")
		}
		presets[name] = true
		if preset.Values == nil {
			def.Presets[i].Values = map[string]string{}
		}
		for key, value := range preset.Values {
			if field, ok := fields[key]; !ok || !validTemplateValue(field, value) {
				return templateError(path, "Preset values must match the template fields")
			}
		}
	}
	langs := validation.GetAllowedLanguages()
	names := make(map[string]bool, len(def.Files))
	for i := range def.Files {
		file := &def.Files[i]
		file.ID = 0
		name := strings.ToLower(validation.SanitizeFilename(file.Filename))
		if names[name] {
			return templateError(fmt.Sprintf("files.%d.filename", i), "Choose a unique filename pattern")
		}
		names[name] = true
		if file.Language == "" {
			file.Language = "plaintext"
		}
		if !slices.Contains(langs, file.Language) {
			return templateError(fmt.Sprintf("files.%d.language", i), "Choose a supported language")
		}
	}
	_, err := renderTemplate(*def, samples, false)
	return err
}

func expandTemplate(text string, values map[string]string, path string) (string, error) {
	var output strings.Builder
	start := 0
	for _, match := range templateToken.FindAllStringSubmatchIndex(text, -1) {
		if match[2] < 0 {
			return "", templateError(path, "Invalid Snipo placeholder")
		}
		key := text[match[4]:match[5]]
		value, exists := values[key]
		if text[match[2]:match[3]] == "!" {
			value, exists = "{{snipo."+key+"}}", true
		}
		if !exists {
			return "", templateError(path, "Unknown template field "+key)
		}
		if output.Len()+match[0]-start+len(value) > maxTemplateOutput {
			return "", templateError(path, "Rendered field exceeds the size limit")
		}
		output.WriteString(text[start:match[0]])
		output.WriteString(value)
		start = match[1]
	}
	if output.Len()+len(text)-start > maxTemplateOutput {
		return "", templateError(path, "Rendered field exceeds the size limit")
	}
	output.WriteString(text[start:])
	return output.String(), nil
}

func renderTemplate(def models.TemplateDefinition, values map[string]string, checkUniqueNames bool) (*models.SnippetInput, error) {
	draft := &models.SnippetInput{FolderID: def.FolderID, Tags: slices.Clone(def.Tags)}
	var err error
	draft.Title, err = expandTemplate(def.Title, values, "title")
	if err != nil {
		return nil, err
	}
	draft.Description, err = expandTemplate(def.Description, values, "description")
	if err != nil {
		return nil, err
	}
	size := len(draft.Title) + len(draft.Description)
	for i, file := range def.Files {
		filenamePath := fmt.Sprintf("files.%d.filename", i)
		file.Filename, err = expandTemplate(file.Filename, values, filenamePath)
		if err != nil {
			return nil, err
		}
		file.Content, err = expandTemplate(file.Content, values, fmt.Sprintf("files.%d.content", i))
		if err != nil {
			return nil, err
		}
		size += len(file.Filename) + len(file.Content)
		if size > maxTemplateOutput {
			return nil, templateError("files", "Rendered draft exceeds the size limit")
		}
		draft.Files = append(draft.Files, file)
	}
	if err := validateTemplateSnippet(draft, checkUniqueNames); err != nil {
		return nil, err
	}
	return draft, nil
}

func validateTemplateSnippet(draft *models.SnippetInput, checkUniqueNames bool) error {
	filenames := make(map[string]bool, len(draft.Files))
	for i, file := range draft.Files {
		name := validation.SanitizeFilename(file.Filename)
		if name == "." || strings.ContainsAny(name, "\x00\r\n:*?\"<>|") || (checkUniqueNames && filenames[strings.ToLower(name)]) {
			return templateError(fmt.Sprintf("files.%d.filename", i), "Choose a valid, unique filename")
		}
		filenames[strings.ToLower(name)] = true
	}
	if len(draft.Files) > 0 {
		draft.Content, draft.Language = draft.Files[0].Content, draft.Files[0].Language
	}
	if errs := validation.ValidateSnippetInput(draft); errs.HasErrors() {
		return errs
	}
	return nil
}

func (s *TemplateService) SaveDefinition(ctx context.Context, id string, expected int, def models.TemplateDefinition) (*models.Template, error) {
	return s.saveDefinition(ctx, id, expected, def, false)
}

func (s *TemplateService) RestoreDefinition(ctx context.Context, id string, expected int, def models.TemplateDefinition) (*models.Template, error) {
	if expected < 1 {
		return nil, templateError("current_revision", "Specify the current revision")
	}
	return s.saveDefinition(ctx, id, expected, def, true)
}

// ReplacePortableDefinition keeps local settings that a portable export omits.
func (s *TemplateService) ReplacePortableDefinition(ctx context.Context, previous *models.Template, def models.TemplateDefinition) (*models.Template, error) {
	def.FolderID = previous.Definition.FolderID
	fields := make(map[string]models.TemplateField, len(def.Fields))
	for _, field := range def.Fields {
		fields[field.Key] = field
	}
	def.Presets = make([]models.TemplatePreset, 0, len(previous.Definition.Presets))
	for _, preset := range previous.Definition.Presets {
		values := make(map[string]string, len(preset.Values))
		for key, value := range preset.Values {
			if field, exists := fields[key]; exists && validTemplateValue(field, value) {
				values[key] = value
			}
		}
		def.Presets = append(def.Presets, models.TemplatePreset{Name: preset.Name, Values: values})
	}
	return s.RestoreDefinition(ctx, previous.ID, previous.Revision, def)
}

func (s *TemplateService) saveDefinition(ctx context.Context, id string, expected int, def models.TemplateDefinition, restore bool) (*models.Template, error) {
	if err := s.Validate(&def); err != nil {
		return nil, err
	}
	folderID, err := s.availableFolder(ctx, def.FolderID)
	if err != nil {
		return nil, err
	}
	if def.FolderID != nil && folderID == nil && !restore {
		return nil, templateError("folder_id", "Choose an existing folder")
	}
	def.FolderID = folderID
	return s.Save(ctx, id, expected, def)
}

func (s *TemplateService) availableFolder(ctx context.Context, id *int64) (*int64, error) {
	if id == nil {
		return nil, nil
	}
	_, err := s.folders.GetByID(ctx, *id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, nil
	}
	return id, err
}

func (s *TemplateService) Render(ctx context.Context, id string, revision int, supplied map[string]string) (*models.SnippetInput, error) {
	template, err := s.Get(ctx, id, revision)
	if err != nil {
		return nil, err
	}
	if template.State == "trash" {
		return nil, sql.ErrNoRows
	}
	draft, err := s.Test(template.Definition, supplied)
	if err != nil {
		return nil, err
	}
	draft.TemplateOrigin = &models.TemplateOrigin{ID: id, Revision: template.Revision, Name: template.Definition.Name}
	draft.FolderID, err = s.availableFolder(ctx, draft.FolderID)
	if err != nil {
		return nil, err
	}
	return draft, nil
}

func (s *TemplateService) Test(def models.TemplateDefinition, supplied map[string]string) (*models.SnippetInput, error) {
	if err := s.Validate(&def); err != nil {
		return nil, err
	}
	values := make(map[string]string, len(def.Fields))
	var errs validation.ValidationErrors
	for _, field := range def.Fields {
		value, provided := supplied[field.Key]
		if !provided && field.Default != nil {
			value = *field.Default
		}
		if field.Required && strings.TrimSpace(value) == "" {
			errs = append(errs, validation.ValidationError{Field: field.Key, Message: "Enter a value for " + field.Label})
		} else if !validTemplateValue(field, value) {
			errs = append(errs, validation.ValidationError{Field: field.Key, Message: "Value must be a valid " + field.Type})
		}
		values[field.Key] = value
	}
	if len(errs) > 0 {
		return nil, errs
	}
	for key := range supplied {
		if _, ok := values[key]; !ok {
			return nil, templateError("values", "Values include an unknown template field")
		}
	}
	draft, err := renderTemplate(def, values, true)
	if err != nil {
		return nil, err
	}
	// The files already carry the content. The save path populates the legacy
	// primary content field from the first file.
	draft.Content = ""
	var body bytes.Buffer
	encoder := json.NewEncoder(&body)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(draft); err != nil {
		return nil, err
	}
	// Leave space for template origin and client metadata.
	if body.Len() > validation.MaxJSONRequestBytes-8192 {
		return nil, templateError("files", "Rendered draft is too large to save. Shorten the field values or split the content into separate snippets.")
	}
	return draft, nil
}

func validateTemplateSnapshots(items []models.TemplateSnapshot) error {
	ids := make(map[string]bool, len(items))
	for _, item := range items {
		if !templateID.MatchString(item.ID) || ids[item.ID] || item.Revision < 1 || !slices.Contains([]string{"active", "archived", "trash"}, item.State) || item.UpdatedAt.IsZero() || len(item.Revisions) != item.Revision {
			return fmt.Errorf("invalid template snapshot")
		}
		ids[item.ID] = true
		revisions := make(map[int]bool, len(item.Revisions))
		for _, revision := range item.Revisions {
			if revision.Revision < 1 || revision.Revision > item.Revision || revisions[revision.Revision] || revision.CreatedAt.IsZero() {
				return fmt.Errorf("invalid template history")
			}
			revisions[revision.Revision] = true
			if err := validateTemplate(&revision.Definition, 0); err != nil {
				return err
			}
		}
	}
	return nil
}

func TemplateStarters() []models.TemplateDefinition {
	return []models.TemplateDefinition{
		{
			Name: "Service runbook", Purpose: "Document health checks, recovery steps, and commands for a service.",
			Title: "{{snipo.service}} runbook", Tags: []string{"runbook"},
			Fields: []models.TemplateField{
				{Key: "service", Label: "Service", Type: "text", Required: true},
				{Key: "host", Label: "Host", Type: "text", Required: true},
			},
			Files: []models.SnippetFileInput{{Filename: "runbook.md", Language: "markdown",
				Content: "# {{snipo.service}}\n\nHost: {{snipo.host}}\n\n## Health checks\n\n## Recovery steps\n\n## Useful commands\n"}},
		},
		{
			Name: "Research note", Purpose: "Capture a source and what you learned from it.", Title: "{{snipo.page_title}}", Tags: []string{"research"},
			Fields: []models.TemplateField{
				{Key: "page_title", Label: "Source title", Type: "text", Source: "page_title", Required: true},
				{Key: "url", Label: "Source URL", Type: "text", Source: "url"},
				{Key: "selection", Label: "Excerpt", Type: "multiline", Source: "selection"},
			},
			Files: []models.SnippetFileInput{{Filename: "note.md", Language: "markdown",
				Content: "# {{snipo.page_title}}\n\nSource: {{snipo.url}}\n\n## Excerpt\n\n{{snipo.selection}}\n\n## Notes\n"}},
		},
		{
			Name: "Compose service", Purpose: "Start with Compose, an example environment file, and a README.", Title: "{{snipo.service}} Compose service",
			Fields: []models.TemplateField{{Key: "service", Label: "Service name", Type: "text", Required: true}},
			Files: []models.SnippetFileInput{
				{Filename: "compose.yaml", Language: "yaml", Content: "services:\n  {{snipo.service}}:\n    image: ${SERVICE_IMAGE}\n"},
				{Filename: ".env.example", Language: "plaintext", Content: "SERVICE_IMAGE=example/image:tag\n"},
				{Filename: "README.md", Language: "markdown", Content: "# {{snipo.service}}\n\nCopy .env.example to .env and set SERVICE_IMAGE.\n"},
			},
		},
	}
}
