package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/MohamedElashri/snipo/internal/models"
	"github.com/MohamedElashri/snipo/internal/repository"
)

// FolderHandler handles folder-related HTTP requests
type FolderHandler struct {
	repo *repository.FolderRepository
}

// NewFolderHandler creates a new folder handler
func NewFolderHandler(repo *repository.FolderRepository) *FolderHandler {
	return &FolderHandler{repo: repo}
}

// List handles GET /api/v1/folders
func (h *FolderHandler) List(w http.ResponseWriter, r *http.Request) {
	// Check if tree format is requested
	tree := r.URL.Query().Get("tree") == "true"

	var folders []models.Folder
	var err error

	if tree {
		folders, err = h.repo.ListTree(r.Context())
	} else {
		folders, err = h.repo.List(r.Context())
	}

	if err != nil {
		InternalError(w)
		return
	}

	// Get snippet counts for each folder (only for flat list)
	if !tree {
		for i := range folders {
			count, err := h.repo.GetFolderSnippetCount(r.Context(), folders[i].ID)
			if err == nil {
				folders[i].SnippetCount = count
			}
		}
	}

	OK(w, map[string]interface{}{"data": folders})
}

// Create handles POST /api/v1/folders
func (h *FolderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input models.FolderInput
	if err := DecodeJSON(r, &input); err != nil {
		Error(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON payload")
		return
	}

	// Validate input
	if input.Name == "" {
		ValidationErrors(w, []ValidationError{{Field: "name", Message: "Name is required"}})
		return
	}

	if len(input.Name) > 100 {
		ValidationErrors(w, []ValidationError{{Field: "name", Message: "Name must be 100 characters or less"}})
		return
	}

	// Validate parent exists if provided
	if input.ParentID != nil {
		_, err := h.repo.GetByID(r.Context(), *input.ParentID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				ValidationErrors(w, []ValidationError{{Field: "parent_id", Message: "Parent folder not found"}})
				return
			}
			InternalError(w)
			return
		}
	}

	folder, err := h.repo.Create(r.Context(), &input)
	if err != nil {
		InternalError(w)
		return
	}

	Created(w, folder)
}

// Get handles GET /api/v1/folders/{id}
func (h *FolderHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid folder ID")
		return
	}

	folder, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			NotFound(w, "Folder not found")
			return
		}
		InternalError(w)
		return
	}

	// Get snippet count
	count, err := h.repo.GetFolderSnippetCount(r.Context(), folder.ID)
	if err == nil {
		folder.SnippetCount = count
	}

	OK(w, folder)
}

// Update handles PUT /api/v1/folders/{id}
func (h *FolderHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid folder ID")
		return
	}

	var input models.FolderInput
	if err := DecodeJSON(r, &input); err != nil {
		Error(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON payload")
		return
	}

	// Validate input
	if input.Name == "" {
		ValidationErrors(w, []ValidationError{{Field: "name", Message: "Name is required"}})
		return
	}

	if len(input.Name) > 100 {
		ValidationErrors(w, []ValidationError{{Field: "name", Message: "Name must be 100 characters or less"}})
		return
	}

	// Validate parent exists if provided and not self-referencing
	if input.ParentID != nil {
		if *input.ParentID == id {
			ValidationErrors(w, []ValidationError{{Field: "parent_id", Message: "Folder cannot be its own parent"}})
			return
		}

		_, err := h.repo.GetByID(r.Context(), *input.ParentID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				ValidationErrors(w, []ValidationError{{Field: "parent_id", Message: "Parent folder not found"}})
				return
			}
			InternalError(w)
			return
		}
	}

	folder, err := h.repo.Update(r.Context(), id, &input)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			NotFound(w, "Folder not found")
			return
		}
		InternalError(w)
		return
	}

	OK(w, folder)
}

// Delete handles DELETE /api/v1/folders/{id}
func (h *FolderHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid folder ID")
		return
	}

	err = h.repo.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			NotFound(w, "Folder not found")
			return
		}
		InternalError(w)
		return
	}

	NoContent(w)
}

// MoveRequest represents a request to move a folder
type MoveRequest struct {
	ParentID *int64 `json:"parent_id"`
}

// Move handles PUT /api/v1/folders/{id}/move
func (h *FolderHandler) Move(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid folder ID")
		return
	}

	var req MoveRequest
	if err := DecodeJSON(r, &req); err != nil {
		Error(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON payload")
		return
	}

	// Validate not moving to self
	if req.ParentID != nil && *req.ParentID == id {
		ValidationErrors(w, []ValidationError{{Field: "parent_id", Message: "Folder cannot be its own parent"}})
		return
	}

	// Validate parent exists if provided
	if req.ParentID != nil {
		_, err := h.repo.GetByID(r.Context(), *req.ParentID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				ValidationErrors(w, []ValidationError{{Field: "parent_id", Message: "Parent folder not found"}})
				return
			}
			InternalError(w)
			return
		}
	}

	folder, err := h.repo.Move(r.Context(), id, req.ParentID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			NotFound(w, "Folder not found")
			return
		}
		// Check for circular reference error
		if err.Error() == "cannot move folder: would create circular reference" {
			ValidationErrors(w, []ValidationError{{Field: "parent_id", Message: "Cannot move folder: would create circular reference"}})
			return
		}
		InternalError(w)
		return
	}

	OK(w, folder)
}
