package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

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

	if q := r.URL.Query().Get("q"); q != "" {
		filter.Query = q
	}

	if lang := r.URL.Query().Get("language"); lang != "" {
		filter.Language = lang
	}

	if fav := r.URL.Query().Get("favorite"); fav != "" {
		isFav := fav == "true" || fav == "1"
		filter.IsFavorite = &isFav
	}

	if archived := r.URL.Query().Get("is_archived"); archived != "" {
		isArchived := archived == "true" || archived == "1"
		filter.IsArchived = &isArchived
	}

	if deleted := r.URL.Query().Get("is_deleted"); deleted != "" {
		isDeleted := deleted == "true" || deleted == "1"
		filter.IsDeleted = &isDeleted
	}

	if tagID := r.URL.Query().Get("tag_id"); tagID != "" {
		if id, err := strconv.ParseInt(tagID, 10, 64); err == nil && id > 0 {
			filter.TagID = id
		}
	}

	// Support multiple tag filtering (tag_ids=1,2,3)
	if tagIDs := r.URL.Query().Get("tag_ids"); tagIDs != "" {
		idStrs := strings.Split(tagIDs, ",")
		for _, idStr := range idStrs {
			if id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64); err == nil && id > 0 {
				filter.TagIDs = append(filter.TagIDs, id)
			}
		}
	}

	if folderID := r.URL.Query().Get("folder_id"); folderID != "" {
		if id, err := strconv.ParseInt(folderID, 10, 64); err == nil && id > 0 {
			filter.FolderID = id
		}
	}

	// Support multiple folder filtering (folder_ids=1,2,3)
	if folderIDs := r.URL.Query().Get("folder_ids"); folderIDs != "" {
		idStrs := strings.Split(folderIDs, ",")
		for _, idStr := range idStrs {
			if id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64); err == nil && id > 0 {
				filter.FolderIDs = append(filter.FolderIDs, id)
			}
		}
	}

	if sortBy := r.URL.Query().Get("sort"); sortBy != "" {
		validSortColumns := map[string]bool{
			"id":          true,
			"title":       true,
			"description": true,
			"content":     true,
			"language":    true,
			"is_favorite": true,
			"is_public":   true,
			"view_count":  true,
			"created_at":  true,
			"updated_at":  true,
		}
		if validSortColumns[sortBy] {
			filter.SortBy = sortBy
		}
	}

	if order := r.URL.Query().Get("order"); order != "" {
		if order == "asc" || order == "desc" {
			filter.SortOrder = order
		}
	}

	result, err := h.service.List(r.Context(), filter)
	if err != nil {
		InternalError(w, r)
		return
	}

	// Use SuccessList to include pagination metadata
	SuccessList(w, r, result.Data, result.Pagination.Page, result.Pagination.Limit, result.Pagination.Total)
}

// Create handles POST /api/v1/snippets
func (h *SnippetHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input models.SnippetInput
	if err := DecodeJSON(r, &input); err != nil {
		Error(w, r, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON payload")
		return
	}

	snippet, err := h.service.Create(r.Context(), &input)
	if err != nil {
		// Check if it's a validation error
		var validationErrs validation.ValidationErrors
		if errors.As(err, &validationErrs) {
			ValidationErrors(w, r, validationErrs)
			return
		}
		InternalError(w, r)
		return
	}

	Created(w, r, snippet)
}

// Get handles GET /api/v1/snippets/{id}
func (h *SnippetHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		Error(w, r, http.StatusBadRequest, "MISSING_ID", "Snippet ID is required")
		return
	}

	snippet, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrSnippetNotFound) {
			NotFound(w, r, "Snippet not found")
			return
		}
		InternalError(w, r)
		return
	}

	OK(w, r, snippet)
}

// Update handles PUT /api/v1/snippets/{id}
func (h *SnippetHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		Error(w, r, http.StatusBadRequest, "MISSING_ID", "Snippet ID is required")
		return
	}

	var input models.SnippetInput
	if err := DecodeJSON(r, &input); err != nil {
		Error(w, r, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON payload")
		return
	}

	snippet, err := h.service.Update(r.Context(), id, &input)
	if err != nil {
		if errors.Is(err, services.ErrSnippetNotFound) {
			NotFound(w, r, "Snippet not found")
			return
		}
		var validationErrs validation.ValidationErrors
		if errors.As(err, &validationErrs) {
			ValidationErrors(w, r, validationErrs)
			return
		}
		InternalError(w, r)
		return
	}

	OK(w, r, snippet)
}

// Delete handles DELETE /api/v1/snippets/{id}
func (h *SnippetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		Error(w, r, http.StatusBadRequest, "MISSING_ID", "Snippet ID is required")
		return
	}

	permanent := r.URL.Query().Get("permanent") == "true"
	err := h.service.Delete(r.Context(), id, permanent)
	if err != nil {
		if errors.Is(err, services.ErrSnippetNotFound) {
			NotFound(w, r, "Snippet not found")
			return
		}
		InternalError(w, r)
		return
	}

	NoContent(w)
}

// Restore handles POST /api/v1/snippets/{id}/restore
func (h *SnippetHandler) Restore(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		Error(w, r, http.StatusBadRequest, "MISSING_ID", "Snippet ID is required")
		return
	}

	err := h.service.Restore(r.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrSnippetNotFound) {
			NotFound(w, r, "Snippet not found")
			return
		}
		InternalError(w, r)
		return
	}

	OK(w, r, map[string]string{"status": "restored"})
}

// ToggleFavorite handles POST /api/v1/snippets/{id}/favorite
func (h *SnippetHandler) ToggleFavorite(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		Error(w, r, http.StatusBadRequest, "MISSING_ID", "Snippet ID is required")
		return
	}

	snippet, err := h.service.ToggleFavorite(r.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrSnippetNotFound) {
			NotFound(w, r, "Snippet not found")
			return
		}
		InternalError(w, r)
		return
	}

	OK(w, r, snippet)
}

// ToggleArchive handles POST /api/v1/snippets/{id}/archive
func (h *SnippetHandler) ToggleArchive(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		Error(w, r, http.StatusBadRequest, "MISSING_ID", "Snippet ID is required")
		return
	}

	snippet, err := h.service.ToggleArchive(r.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrSnippetNotFound) {
			NotFound(w, r, "Snippet not found")
			return
		}
		InternalError(w, r)
		return
	}

	OK(w, r, snippet)
}

// Duplicate handles POST /api/v1/snippets/{id}/duplicate
func (h *SnippetHandler) Duplicate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		Error(w, r, http.StatusBadRequest, "MISSING_ID", "Snippet ID is required")
		return
	}

	snippet, err := h.service.Duplicate(r.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrSnippetNotFound) {
			NotFound(w, r, "Snippet not found")
			return
		}
		InternalError(w, r)
		return
	}

	Created(w, r, snippet)
}

// Search handles GET /api/v1/snippets/search
func (h *SnippetHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		OK(w, r, []models.Snippet{})
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
		InternalError(w, r)
		return
	}

	OK(w, r, snippets)
}

// GetPublic handles GET /api/v1/snippets/public/{id}
func (h *SnippetHandler) GetPublic(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		Error(w, r, http.StatusBadRequest, "MISSING_ID", "Snippet ID is required")
		return
	}

	snippet, err := h.service.GetByIDPublic(r.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrSnippetNotFound) {
			NotFound(w, r, "Snippet not found")
			return
		}
		InternalError(w, r)
		return
	}

	OK(w, r, snippet)
}

// GetPublicFile handles GET /api/v1/snippets/public/{id}/files/{filename}
// Returns raw file content for downloading individual files from public snippets
func (h *SnippetHandler) GetPublicFile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		Error(w, r, http.StatusBadRequest, "MISSING_ID", "Snippet ID is required")
		return
	}

	filename := chi.URLParam(r, "filename")
	if filename == "" {
		Error(w, r, http.StatusBadRequest, "MISSING_FILENAME", "Filename is required")
		return
	}

	// Get the public snippet (this also checks if it's public and increments view count)
	snippet, err := h.service.GetByIDPublic(r.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrSnippetNotFound) {
			NotFound(w, r, "File not found")
			return
		}
		InternalError(w, r)
		return
	}

	// Find the requested file
	var targetFile *models.SnippetFile
	for i := range snippet.Files {
		if snippet.Files[i].Filename == filename {
			targetFile = &snippet.Files[i]
			break
		}
	}

	// If not found, try with sanitized filename (to handle legacy files with spaces)
	if targetFile == nil {
		sanitizedFilename := validation.SanitizeFilename(filename)
		for i := range snippet.Files {
			if validation.SanitizeFilename(snippet.Files[i].Filename) == sanitizedFilename {
				targetFile = &snippet.Files[i]
				break
			}
		}
	}

	// If not found in files array, check if it's the legacy single-file content
	if targetFile == nil && len(snippet.Files) == 0 {
		// For legacy single-file snippets, use the main content
		targetFile = &models.SnippetFile{
			Filename: filename,
			Content:  snippet.Content,
			Language: snippet.Language,
		}
	}

	if targetFile == nil {
		NotFound(w, r, "File not found in snippet")
		return
	}

	// Set appropriate headers for raw file download
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", "inline; filename=\""+filename+"\"")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	// Write the file content
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(targetFile.Content)); err != nil {
		// Log the error but don't return as headers are already sent
		// This is best effort - connection may have been closed
		return
	}
}

// GetHistory handles GET /api/v1/snippets/{id}/history
func (h *SnippetHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		Error(w, r, http.StatusBadRequest, "MISSING_ID", "Snippet ID is required")
		return
	}

	// Parse limit from query parameter
	limit := 50 // default
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 200 {
			limit = l
		}
	}

	history, err := h.service.GetHistory(r.Context(), id, limit)
	if err != nil {
		if errors.Is(err, services.ErrSnippetNotFound) {
			NotFound(w, r, "Snippet not found")
			return
		}
		InternalError(w, r)
		return
	}

	OK(w, r, history)
}

// RestoreFromHistory handles POST /api/v1/snippets/{id}/history/{history_id}/restore
func (h *SnippetHandler) RestoreFromHistory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		Error(w, r, http.StatusBadRequest, "MISSING_ID", "Snippet ID is required")
		return
	}

	historyIDStr := chi.URLParam(r, "history_id")
	if historyIDStr == "" {
		Error(w, r, http.StatusBadRequest, "MISSING_HISTORY_ID", "History ID is required")
		return
	}

	historyID, err := strconv.ParseInt(historyIDStr, 10, 64)
	if err != nil || historyID <= 0 {
		Error(w, r, http.StatusBadRequest, "INVALID_HISTORY_ID", "Invalid history ID")
		return
	}

	snippet, err := h.service.RestoreFromHistory(r.Context(), id, historyID)
	if err != nil {
		if errors.Is(err, services.ErrSnippetNotFound) {
			NotFound(w, r, "Snippet not found")
			return
		}
		Error(w, r, http.StatusBadRequest, "RESTORE_FAILED", err.Error())
		return
	}

	OK(w, r, snippet)
}
