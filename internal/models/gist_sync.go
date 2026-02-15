package models

import (
	"time"
)

// GistSyncConfig represents the global gist sync configuration
type GistSyncConfig struct {
	ID                         int        `json:"id"`
	Enabled                    bool       `json:"enabled"`
	GithubTokenEncrypted       string     `json:"-"`
	GithubUsername             string     `json:"github_username"`
	AutoSyncEnabled            bool       `json:"auto_sync_enabled"`
	SyncIntervalMinutes        int        `json:"sync_interval_minutes"`
	ConflictResolutionStrategy string     `json:"conflict_resolution_strategy"`
	LastFullSyncAt             *time.Time `json:"last_full_sync_at,omitempty"`
	CreatedAt                  time.Time  `json:"created_at"`
	UpdatedAt                  time.Time  `json:"updated_at"`
}

// SnippetGistMapping represents the mapping between a snippet and a gist
type SnippetGistMapping struct {
	ID            int64      `json:"id"`
	SnippetID     string     `json:"snippet_id"`
	GistID        string     `json:"gist_id"`
	GistURL       string     `json:"gist_url"`
	SyncEnabled   bool       `json:"sync_enabled"`
	LastSyncedAt  *time.Time `json:"last_synced_at,omitempty"`
	SnipoChecksum string     `json:"snipo_checksum"`
	GistChecksum  string     `json:"gist_checksum"`
	SyncStatus    string     `json:"sync_status"`
	ErrorMessage  *string    `json:"error_message,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// GistSyncConflict represents a sync conflict that needs resolution
type GistSyncConflict struct {
	ID               int64      `json:"id"`
	SnippetID        string     `json:"snippet_id"`
	GistID           string     `json:"gist_id"`
	SnipoVersion     string     `json:"snipo_version"`
	GistVersion      string     `json:"gist_version"`
	Resolved         bool       `json:"resolved"`
	ResolutionChoice *string    `json:"resolution_choice,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	ResolvedAt       *time.Time `json:"resolved_at,omitempty"`
}

// GistSyncLog represents a log entry for sync operations
type GistSyncLog struct {
	ID        int64     `json:"id"`
	SnippetID *string   `json:"snippet_id,omitempty"`
	GistID    *string   `json:"gist_id,omitempty"`
	Operation string    `json:"operation"`
	Status    string    `json:"status"`
	Message   *string   `json:"message,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// SyncResult represents the result of a sync operation
type SyncResult struct {
	TotalProcessed int      `json:"total_processed"`
	Synced         int      `json:"synced"`
	Conflicts      int      `json:"conflicts"`
	Errors         int      `json:"errors"`
	ErrorMessages  []string `json:"error_messages,omitempty"`
	Duration       string   `json:"duration"`
}

// GistRequest represents a request to create or update a gist
type GistRequest struct {
	Description string              `json:"description"`
	Public      bool                `json:"public"`
	Files       map[string]GistFile `json:"files"`
}

// GistFile represents a file in a gist
type GistFile struct {
	Content  string  `json:"content"`
	Filename *string `json:"filename,omitempty"`
}

// GistResponse represents a response from GitHub Gist API
type GistResponse struct {
	ID          string              `json:"id"`
	URL         string              `json:"url"`
	HTMLURL     string              `json:"html_url"`
	Description string              `json:"description"`
	Public      bool                `json:"public"`
	Files       map[string]GistFile `json:"files"`
	Owner       *GistOwner          `json:"owner,omitempty"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

// GistOwner represents the owner of a gist
type GistOwner struct {
	Login string `json:"login"`
	ID    int64  `json:"id"`
}

// SnipoMetadata represents Snipo-specific metadata stored in gists
type SnipoMetadata struct {
	Version      string   `json:"version"`
	SnipoID      string   `json:"snipo_id"`
	Folders      []Folder `json:"folders,omitempty"`
	TagsOverflow []string `json:"tags_overflow,omitempty"`
	IsFavorite   bool     `json:"is_favorite"`
	IsArchived   bool     `json:"is_archived"`
}

// SyncDirection represents the direction of sync
type SyncDirection int

const (
	NoSync SyncDirection = iota
	SnipoToGist
	GistToSnipo
	Conflict
	GistDeleted
)

// Sync status constants
const (
	SyncStatusSynced   = "synced"
	SyncStatusPending  = "pending"
	SyncStatusConflict = "conflict"
	SyncStatusError    = "error"
)

// Conflict resolution strategies
const (
	ConflictStrategyManual     = "manual"
	ConflictStrategySnipoWins  = "snipo_wins"
	ConflictStrategyGistWins   = "gist_wins"
	ConflictStrategyNewestWins = "newest_wins"
)

// Sync operations
const (
	SyncOpCreate   = "create"
	SyncOpUpdate   = "update"
	SyncOpDelete   = "delete"
	SyncOpSync     = "sync"
	SyncOpConflict = "conflict"
)

// Sync operation statuses
const (
	SyncOpStatusSuccess = "success"
	SyncOpStatusFailed  = "failed"
)
