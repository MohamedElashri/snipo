package api

import "time"

type Meta struct {
	RequestID string    `json:"request_id"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

type PaginationLinks struct {
	Self string  `json:"self"`
	Next *string `json:"next"`
	Prev *string `json:"prev"`
}

type Pagination struct {
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	Total      int             `json:"total"`
	TotalPages int             `json:"total_pages"`
	Links      PaginationLinks `json:"links"`
}

type APIResponse struct {
	Data interface{} `json:"data"`
	Meta Meta        `json:"meta"`
}

type ListResponse struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
	Meta       Meta        `json:"meta"`
}

type ErrorResponse struct {
	Error struct {
		Code      string      `json:"code"`
		Message   string      `json:"message"`
		Details   interface{} `json:"details,omitempty"`
		RequestID string      `json:"request_id"`
		Timestamp time.Time   `json:"timestamp"`
	} `json:"error"`
}

type Snippet struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Language    string    `json:"language"`
	Content     string    `json:"content"`
	IsFavorite  bool      `json:"is_favorite"`
	IsArchived  bool      `json:"is_archived"`
	IsPublic    bool      `json:"is_public"`
	ViewCount   int       `json:"view_count"`
	FolderID    *int      `json:"folder_id"`
	Tags        []Tag     `json:"tags"`
	Files       []File    `json:"files,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type File struct {
	ID        int       `json:"id"`
	SnippetID string    `json:"snippet_id"`
	Filename  string    `json:"filename"`
	Content   string    `json:"content"`
	Language  string    `json:"language"`
	Size      int       `json:"size"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Tag struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Folder struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	ParentID  *int      `json:"parent_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SnippetInput struct {
	Title       string      `json:"title"`
	Description string      `json:"description,omitempty"`
	Language    string      `json:"language"`
	Content     string      `json:"content"`
	Tags        []string    `json:"tags,omitempty"`
	FolderID    *int64      `json:"folder_id,omitempty"`
	IsPublic    bool        `json:"is_public"`
	IsArchived  bool        `json:"is_archived,omitempty"`
	Files       []FileInput `json:"files,omitempty"`
}

type FileInput struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
	Language string `json:"language"`
}

type TagInput struct {
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
}

type FolderInput struct {
	Name     string `json:"name"`
	ParentID *int   `json:"parent_id,omitempty"`
}

type HealthResponse struct {
	Status   string          `json:"status"`
	Database string          `json:"database"`
	Version  string          `json:"version"`
	Features map[string]bool `json:"features"`
}
