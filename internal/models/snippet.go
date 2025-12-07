package models

import (
	"time"
)

// Snippet represents a code snippet
type Snippet struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Content     string    `json:"content"`
	Language    string    `json:"language"`
	IsFavorite  bool      `json:"is_favorite"`
	IsPublic    bool      `json:"is_public"`
	ViewCount   int       `json:"view_count"`
	S3Key       *string   `json:"s3_key,omitempty"`
	Checksum    *string   `json:"checksum,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships (populated when needed)
	Tags    []Tag    `json:"tags,omitempty"`
	Folders []Folder `json:"folders,omitempty"`
}

// SnippetInput represents input for creating/updating a snippet
type SnippetInput struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Content     string   `json:"content"`
	Language    string   `json:"language"`
	Tags        []string `json:"tags,omitempty"`
	FolderID    *int64   `json:"folder_id,omitempty"`
	IsPublic    bool     `json:"is_public"`
}

// SnippetFilter represents filter options for listing snippets
type SnippetFilter struct {
	Query      string
	Language   string
	TagID      int64
	FolderID   int64
	IsFavorite *bool
	IsPublic   *bool
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
	Name        string     `json:"name"`
	Permissions string     `json:"permissions"` // "read", "write", "admin"
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
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
