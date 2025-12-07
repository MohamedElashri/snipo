package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/MohamedElashri/snipo/internal/models"
	"github.com/MohamedElashri/snipo/internal/services"
	"github.com/MohamedElashri/snipo/internal/validation"
)

// SnippetHandler handles snippet-related HTTP requests
type SnippetHandler struct {
	service *services.SnippetService
}

// NewSnippetHandler creates a new snippet handler
func NewSnippetHandler(service *services.SnippetService) *SnippetHandler {
	return &SnippetHandler{service: service}
}

// List handles GET /api/v1/snippets
func (h *SnippetHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := models.DefaultSnippetFilter()

	// Parse query parameters
	if page := r.URL.Query().Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			filter.Page = p
		}
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			filter.Limit = l
		}
	}

	if lang := r.URL.Query().Get("language"); lang != "" {
		filter.Language = lang
	}

	if fav := r.URL.Query().Get("favorite"); fav != "" {
		isFav := fav == "true" || fav == "1"
		filter.IsFavorite = &isFav
	}

	if tagID := r.URL.Query().Get("tag_id"); tagID != "" {
		if id, err := strconv.ParseInt(tagID, 10, 64); err == nil && id > 0 {
			filter.TagID = id
		}
	}

	if folderID := r.URL.Query().Get("folder_id"); folderID != "" {
		if id, err := strconv.ParseInt(folderID, 10, 64); err == nil && id > 0 {
			filter.FolderID = id
		}
	}

	if sortBy := r.URL.Query().Get("sort"); sortBy != "" {
		filter.SortBy = sortBy
	}

	if order := r.URL.Query().Get("order"); order != "" {
		filter.SortOrder = order
	}

	result, err := h.service.List(r.Context(), filter)
	if err != nil {
		InternalError(w)
		return
	}

	OK(w, result)
}

// Create handles POST /api/v1/snippets
func (h *SnippetHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input models.SnippetInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		Error(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON payload")
		return
	}

	snippet, err := h.service.Create(r.Context(), &input)
	if err != nil {
		// Check if it's a validation error
		var validationErrs validation.ValidationErrors
		if errors.As(err, &validationErrs) {
			errs := make([]ValidationError, len(validationErrs))
			for i, e := range validationErrs {
				errs[i] = ValidationError{Field: e.Field, Message: e.Message}
			}
			ValidationErrors(w, errs)
			return
		}
		InternalError(w)
		return
	}

	Created(w, snippet)
}

// Get handles GET /api/v1/snippets/{id}
func (h *SnippetHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		Error(w, http.StatusBadRequest, "MISSING_ID", "Snippet ID is required")
		return
	}

	snippet, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrSnippetNotFound) {
			NotFound(w, "Snippet not found")
			return
		}
		InternalError(w)
		return
	}

	OK(w, snippet)
}

// Update handles PUT /api/v1/snippets/{id}
func (h *SnippetHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		Error(w, http.StatusBadRequest, "MISSING_ID", "Snippet ID is required")
		return
	}

	var input models.SnippetInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		Error(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON payload")
		return
	}

	snippet, err := h.service.Update(r.Context(), id, &input)
	if err != nil {
		if errors.Is(err, services.ErrSnippetNotFound) {
			NotFound(w, "Snippet not found")
			return
		}
		var validationErrs validation.ValidationErrors
		if errors.As(err, &validationErrs) {
			errs := make([]ValidationError, len(validationErrs))
			for i, e := range validationErrs {
				errs[i] = ValidationError{Field: e.Field, Message: e.Message}
			}
			ValidationErrors(w, errs)
			return
		}
		InternalError(w)
		return
	}

	OK(w, snippet)
}

// Delete handles DELETE /api/v1/snippets/{id}
func (h *SnippetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		Error(w, http.StatusBadRequest, "MISSING_ID", "Snippet ID is required")
		return
	}

	err := h.service.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrSnippetNotFound) {
			NotFound(w, "Snippet not found")
			return
		}
		InternalError(w)
		return
	}

	NoContent(w)
}

// ToggleFavorite handles POST /api/v1/snippets/{id}/favorite
func (h *SnippetHandler) ToggleFavorite(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		Error(w, http.StatusBadRequest, "MISSING_ID", "Snippet ID is required")
		return
	}

	snippet, err := h.service.ToggleFavorite(r.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrSnippetNotFound) {
			NotFound(w, "Snippet not found")
			return
		}
		InternalError(w)
		return
	}

	OK(w, snippet)
}

// Duplicate handles POST /api/v1/snippets/{id}/duplicate
func (h *SnippetHandler) Duplicate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		Error(w, http.StatusBadRequest, "MISSING_ID", "Snippet ID is required")
		return
	}

	snippet, err := h.service.Duplicate(r.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrSnippetNotFound) {
			NotFound(w, "Snippet not found")
			return
		}
		InternalError(w)
		return
	}

	Created(w, snippet)
}

// Search handles GET /api/v1/snippets/search
func (h *SnippetHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		OK(w, map[string]interface{}{"data": []models.Snippet{}})
		return
	}

	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	snippets, err := h.service.Search(r.Context(), query, limit)
	if err != nil {
		InternalError(w)
		return
	}

	OK(w, map[string]interface{}{"data": snippets})
}

// GetPublic handles GET /api/v1/snippets/public/{id}
func (h *SnippetHandler) GetPublic(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		Error(w, http.StatusBadRequest, "MISSING_ID", "Snippet ID is required")
		return
	}

	snippet, err := h.service.GetByIDPublic(r.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrSnippetNotFound) {
			NotFound(w, "Snippet not found")
			return
		}
		InternalError(w)
		return
	}

	OK(w, snippet)
}
