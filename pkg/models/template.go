package models

import "time"

type TemplateField struct {
	Key      string   `json:"key"`
	Label    string   `json:"label"`
	Help     string   `json:"help,omitempty"`
	Type     string   `json:"type"`
	Required bool     `json:"required,omitempty"`
	Default  *string  `json:"default,omitempty"`
	Choices  []string `json:"choices,omitempty"`
	Source   string   `json:"source,omitempty"` // selection, page_title, or url
}

type TemplatePreset struct {
	Name   string            `json:"name"`
	Values map[string]string `json:"values"`
}

type TemplateDefinition struct {
	Name        string             `json:"name"`
	Purpose     string             `json:"purpose"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Files       []SnippetFileInput `json:"files"`
	Fields      []TemplateField    `json:"fields"`
	Presets     []TemplatePreset   `json:"presets,omitempty"`
	Tags        []string           `json:"tags,omitempty"`
	FolderID    *int64             `json:"folder_id,omitempty"`
}

type Template struct {
	ID         string             `json:"id"`
	Revision   int                `json:"revision"`
	State      string             `json:"state"`
	Favorite   bool               `json:"favorite"`
	UpdatedAt  time.Time          `json:"updated_at"`
	LastUsedAt *time.Time         `json:"last_used_at,omitempty"`
	Definition TemplateDefinition `json:"definition"`
}

type TemplateSummary struct {
	ID         string     `json:"id"`
	Revision   int        `json:"revision"`
	State      string     `json:"state"`
	Favorite   bool       `json:"favorite"`
	Name       string     `json:"name"`
	Purpose    string     `json:"purpose"`
	Filenames  []string   `json:"filenames"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

type TemplateRevision struct {
	Revision   int                `json:"revision"`
	CreatedAt  time.Time          `json:"created_at"`
	Definition TemplateDefinition `json:"definition,omitzero"`
}

type TemplateOrigin struct {
	ID       string `json:"id"`
	Revision int    `json:"revision"`
	Name     string `json:"name"`
}

type TemplateSnapshot struct {
	Template
	Revisions []TemplateRevision `json:"revisions"`
}

type TemplateExport struct {
	Version    int                `json:"version"`
	ID         string             `json:"id,omitempty"`
	Definition TemplateDefinition `json:"definition"`
}
