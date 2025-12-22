package models

import "time"

// Settings represents application settings
type Settings struct {
	ID                      int64     `json:"id"`
	AppName                 string    `json:"app_name"`
	CustomCSS               string    `json:"custom_css"`
	Theme                   string    `json:"theme"`
	DefaultLanguage         string    `json:"default_language"`
	S3Enabled               bool      `json:"s3_enabled"`
	S3Endpoint              string    `json:"s3_endpoint"`
	S3Bucket                string    `json:"s3_bucket"`
	S3Region                string    `json:"s3_region"`
	BackupEncryptionEnabled bool      `json:"backup_encryption_enabled"`
	ArchiveEnabled          bool      `json:"archive_enabled"`
	HistoryEnabled          bool      `json:"history_enabled"`
	EditorFontSize          int       `json:"editor_font_size"`
	EditorTabSize           int       `json:"editor_tab_size"`
	EditorTheme             string    `json:"editor_theme"`
	EditorWordWrap          bool      `json:"editor_word_wrap"`
	EditorShowPrintMargin   bool      `json:"editor_show_print_margin"`
	EditorShowGutter        bool      `json:"editor_show_gutter"`
	EditorShowIndentGuides  bool      `json:"editor_show_indent_guides"`
	EditorHighlightActiveLine bool    `json:"editor_highlight_active_line"`
	EditorUseSoftTabs       bool      `json:"editor_use_soft_tabs"`
	EditorEnableSnippets    bool      `json:"editor_enable_snippets"`
	EditorEnableLiveAutocompletion bool `json:"editor_enable_live_autocompletion"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

// SettingsInput represents input for updating settings
type SettingsInput struct {
	AppName                 string `json:"app_name"`
	CustomCSS               string `json:"custom_css"`
	Theme                   string `json:"theme"`
	DefaultLanguage         string `json:"default_language"`
	S3Enabled               bool   `json:"s3_enabled"`
	S3Endpoint              string `json:"s3_endpoint"`
	S3Bucket                string `json:"s3_bucket"`
	S3Region                string `json:"s3_region"`
	S3AccessKeyID           string `json:"s3_access_key_id,omitempty"`     // Optional, only for updates
	S3SecretAccessKey       string `json:"s3_secret_access_key,omitempty"` // Optional, only for updates
	BackupEncryptionEnabled bool   `json:"backup_encryption_enabled"`
	ArchiveEnabled          bool   `json:"archive_enabled"`
	HistoryEnabled          bool   `json:"history_enabled"`
	EditorFontSize          int    `json:"editor_font_size"`
	EditorTabSize           int    `json:"editor_tab_size"`
	EditorTheme             string `json:"editor_theme"`
	EditorWordWrap          bool   `json:"editor_word_wrap"`
	EditorShowPrintMargin   bool   `json:"editor_show_print_margin"`
	EditorShowGutter        bool   `json:"editor_show_gutter"`
	EditorShowIndentGuides  bool   `json:"editor_show_indent_guides"`
	EditorHighlightActiveLine bool `json:"editor_highlight_active_line"`
	EditorUseSoftTabs       bool   `json:"editor_use_soft_tabs"`
	EditorEnableSnippets    bool   `json:"editor_enable_snippets"`
	EditorEnableLiveAutocompletion bool `json:"editor_enable_live_autocompletion"`
}
