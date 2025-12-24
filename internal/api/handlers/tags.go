package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/MohamedElashri/snipo/internal/models"
	"github.com/MohamedElashri/snipo/internal/repository"
	"github.com/MohamedElashri/snipo/internal/validation"
)

// TagHandler handles tag-related HTTP requests
type TagHandler struct {
	repo *repository.TagRepository
}

// NewTagHandler creates a new tag handler
func NewTagHandler(repo *repository.TagRepository) *TagHandler {
	return &TagHandler{repo: repo}
}

// List handles GET /api/v1/tags
func (h *TagHandler) List(w http.ResponseWriter, r *http.Request) {
	tags, err := h.repo.List(r.Context())
	if err != nil {
		InternalError(w, r)
		return
	}

	// Get snippet counts for each tag
	for i := range tags {
		count, err := h.repo.GetTagSnippetCount(r.Context(), tags[i].ID)
		if err == nil {
			tags[i].SnippetCount = count
		}
	}

	OK(w, r, tags)
}

// Create handles POST /api/v1/tags
func (h *TagHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input models.TagInput
	if err := DecodeJSON(r, &input); err != nil {
		Error(w, r, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON payload")
		return
	}

	// Validate input
	if input.Name == "" {
		ValidationErrors(w, r, validation.ValidationErrors{validation.ValidationError{Field: "name", Message: "Name is required"}})
		return
	}

	if len(input.Name) > 50 {
		ValidationErrors(w, r, validation.ValidationErrors{validation.ValidationError{Field: "name", Message: "Name must be 50 characters or less"}})
		return
	}

	// Set default color if not provided
	if input.Color == "" {
		input.Color = "#6366f1"
	}

	// Check if tag already exists
	existing, err := h.repo.GetByName(r.Context(), input.Name)
	if err == nil && existing != nil {
		Error(w, r, http.StatusConflict, "TAG_EXISTS", "A tag with this name already exists")
		return
	}

	tag, err := h.repo.Create(r.Context(), &input)
	if err != nil {
		InternalError(w, r)
		return
	}

	Created(w, r, tag)
}

// Get handles GET /api/v1/tags/{id}
func (h *TagHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Error(w, r, http.StatusBadRequest, "INVALID_ID", "Invalid tag ID")
		return
	}

	tag, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			NotFound(w, r, "Tag not found")
			return
		}
		InternalError(w, r)
		return
	}

	// Get snippet count
	count, err := h.repo.GetTagSnippetCount(r.Context(), tag.ID)
	if err == nil {
		tag.SnippetCount = count
	}

	OK(w, r, tag)
}

// Update handles PUT /api/v1/tags/{id}
func (h *TagHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Error(w, r, http.StatusBadRequest, "INVALID_ID", "Invalid tag ID")
		return
	}

	var input models.TagInput
	if err := DecodeJSON(r, &input); err != nil {
		Error(w, r, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON payload")
		return
	}

	// Validate input
	if input.Name == "" {
		ValidationErrors(w, r, validation.ValidationErrors{validation.ValidationError{Field: "name", Message: "Name is required"}})
		return
	}

	if len(input.Name) > 50 {
		ValidationErrors(w, r, validation.ValidationErrors{validation.ValidationError{Field: "name", Message: "Name must be 50 characters or less"}})
		return
	}

	// Check if another tag with same name exists
	existing, err := h.repo.GetByName(r.Context(), input.Name)
	if err == nil && existing != nil && existing.ID != id {
		Error(w, r, http.StatusConflict, "TAG_EXISTS", "A tag with this name already exists")
		return
	}

	// Set default color if not provided
	if input.Color == "" {
		input.Color = "#6366f1"
	}

	tag, err := h.repo.Update(r.Context(), id, &input)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			NotFound(w, r, "Tag not found")
			return
		}
		InternalError(w, r)
		return
	}

	OK(w, r, tag)
}

// Delete handles DELETE /api/v1/tags/{id}
func (h *TagHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Error(w, r, http.StatusBadRequest, "INVALID_ID", "Invalid tag ID")
		return
	}

	err = h.repo.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			NotFound(w, r, "Tag not found")
			return
		}
		InternalError(w, r)
		return
	}

	NoContent(w)
}
