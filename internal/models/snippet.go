package models

import (
	"time"
)

// SnippetFile represents a file within a snippet
type SnippetFile struct {
	ID        int64     `json:"id"`
	SnippetID string    `json:"snippet_id,omitempty"`
	Filename  string    `json:"filename"`
	Content   string    `json:"content"`
	Language  string    `json:"language"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// Snippet represents a code snippet
type Snippet struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Content     string    `json:"content"`  // Primary/legacy content (first file)
	Language    string    `json:"language"` // Primary/legacy language
	IsFavorite  bool      `json:"is_favorite"`
	IsPublic    bool      `json:"is_public"`
	IsArchived  bool      `json:"is_archived"`
	ViewCount   int       `json:"view_count"`
	S3Key       *string   `json:"s3_key,omitempty"`
	Checksum    *string   `json:"checksum,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships (populated when needed)
	Tags    []Tag         `json:"tags,omitempty"`
	Folders []Folder      `json:"folders,omitempty"`
	Files   []SnippetFile `json:"files,omitempty"` // Multi-file support
}

// SnippetFileInput represents input for a file within a snippet
type SnippetFileInput struct {
	ID       int64  `json:"id,omitempty"` // 0 for new files
	Filename string `json:"filename"`
	Content  string `json:"content"`
	Language string `json:"language"`
}

// SnippetInput represents input for creating/updating a snippet
type SnippetInput struct {
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Content     string             `json:"content"`  // Legacy single-file content
	Language    string             `json:"language"` // Legacy single-file language
	Tags        []string           `json:"tags,omitempty"`
	FolderID    *int64             `json:"folder_id,omitempty"`
	IsPublic    bool               `json:"is_public"`
	IsArchived  bool               `json:"is_archived,omitempty"`
	Files       []SnippetFileInput `json:"files,omitempty"` // Multi-file support
}

// SnippetFilter represents filter options for listing snippets
type SnippetFilter struct {
	Query      string
	Language   string
	TagID      int64
	FolderID   int64
	IsFavorite *bool
	IsPublic   *bool
	IsArchived *bool
	Page       int
	Limit      int
	SortBy     string
	SortOrder  string
}

// DefaultSnippetFilter returns default filter values
func DefaultSnippetFilter() SnippetFilter {
	return SnippetFilter{
		Page:      1,
		Limit:     20,
		SortBy:    "updated_at",
		SortOrder: "desc",
	}
}

// Tag represents a tag for organizing snippets
type Tag struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Color        string    `json:"color"`
	CreatedAt    time.Time `json:"created_at"`
	SnippetCount int       `json:"snippet_count,omitempty"`
}

// TagInput represents input for creating/updating a tag
type TagInput struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// Folder represents a folder for organizing snippets
type Folder struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	ParentID     *int64    `json:"parent_id,omitempty"`
	Icon         string    `json:"icon"`
	SortOrder    int       `json:"sort_order"`
	CreatedAt    time.Time `json:"created_at"`
	SnippetCount int       `json:"snippet_count,omitempty"`
	Children     []Folder  `json:"children,omitempty"`
}

// FolderInput represents input for creating/updating a folder
type FolderInput struct {
	Name      string `json:"name"`
	ParentID  *int64 `json:"parent_id,omitempty"`
	Icon      string `json:"icon,omitempty"`
	SortOrder int    `json:"sort_order,omitempty"`
}

// APIToken represents an API token for external access
type APIToken struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Token       string     `json:"token,omitempty"` // Only returned on creation
	TokenHash   string     `json:"-"`
	Permissions string     `json:"permissions"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// APITokenInput struct here represents input for creating an API token
type APITokenInput struct {
	Name          string `json:"name"`
	Permissions   string `json:"permissions"` // "read", "write", "admin"
	ExpiresInDays *int   `json:"expires_in_days,omitempty"`
}

// Pagination holds pagination info for list responses (ايه ده ؟)
type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// SnippetListResponse represents a paginated list of snippets
type SnippetListResponse struct {
	Data       []Snippet  `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// BackupData represents a complete backup of all data
type BackupData struct {
	Version   string    `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	Snippets  []Snippet `json:"snippets"`
	Tags      []Tag     `json:"tags"`
	Folders   []Folder  `json:"folders"`
}

// ExportOptions configures backup export behavior
type ExportOptions struct {
	Format   string `json:"format"`   // "json" or "zip"
	Password string `json:"password"` // Optional encryption password
}

// ImportOptions configures backup import behavior
type ImportOptions struct {
	Strategy string `json:"strategy"` // "replace", "merge", "skip"
	Password string `json:"password"` // Decryption password if encrypted
}

// ImportResult contains the results of an import operation
type ImportResult struct {
	SnippetsImported int      `json:"snippets_imported"`
	TagsImported     int      `json:"tags_imported"`
	FoldersImported  int      `json:"folders_imported"`
	Errors           []string `json:"errors,omitempty"`
}

// S3BackupInfo represents info about a backup stored in S3
type S3BackupInfo struct {
	Key          string    `json:"key"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"last_modified"`
}

// S3SyncResult contains the results of an S3 sync operation
type S3SyncResult struct {
	Uploaded   int       `json:"uploaded"`
	Errors     []string  `json:"errors,omitempty"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
}

// S3RestoreResult contains the results of an S3 restore operation
type S3RestoreResult struct {
	Restored   int       `json:"restored"`
	Errors     []string  `json:"errors,omitempty"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
}

// SnippetHistory represents a historical version of a snippet
type SnippetHistory struct {
	ID          int64              `json:"id"`
	SnippetID   string             `json:"snippet_id"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Content     string             `json:"content"`
	Language    string             `json:"language"`
	IsFavorite  bool               `json:"is_favorite"`
	IsPublic    bool               `json:"is_public"`
	IsArchived  bool               `json:"is_archived"`
	ChangeType  string             `json:"change_type"` // 'create', 'update', 'delete'
	CreatedAt   time.Time          `json:"created_at"`
	Files       []SnippetFileHistory `json:"files,omitempty"`
}

// SnippetFileHistory represents a historical version of a snippet file
type SnippetFileHistory struct {
	ID         int64     `json:"id"`
	HistoryID  int64     `json:"history_id"`
	SnippetID  string    `json:"snippet_id"`
	Filename   string    `json:"filename"`
	Content    string    `json:"content"`
	Language   string    `json:"language"`
	SortOrder  int       `json:"sort_order"`
	CreatedAt  time.Time `json:"created_at"`
}
