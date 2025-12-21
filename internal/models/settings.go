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
}
