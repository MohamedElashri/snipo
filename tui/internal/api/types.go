package api

import (
	"time"

	"github.com/MohamedElashri/snipo/pkg/models"
)

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

type Snippet = models.Snippet
type File = models.SnippetFile
type Tag = models.Tag
type Folder = models.Folder
type SnippetInput = models.SnippetInput
type FileInput = models.SnippetFileInput
type TagInput = models.TagInput
type FolderInput = models.FolderInput

type HealthResponse struct {
	Status   string          `json:"status"`
	Database string          `json:"database"`
	Version  string          `json:"version"`
	Features map[string]bool `json:"features"`
}
