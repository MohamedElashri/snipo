package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/MohamedElashri/snipo/internal/models"
	"github.com/MohamedElashri/snipo/internal/repository"
)

// TokenHandler handles API token-related HTTP requests
type TokenHandler struct {
	repo *repository.TokenRepository
}

// NewTokenHandler creates a new token handler
func NewTokenHandler(repo *repository.TokenRepository) *TokenHandler {
	return &TokenHandler{repo: repo}
}

// List handles GET /api/v1/tokens
func (h *TokenHandler) List(w http.ResponseWriter, r *http.Request) {
	tokens, err := h.repo.List(r.Context())
	if err != nil {
		InternalError(w)
		return
	}

	OK(w, map[string]interface{}{"data": tokens})
}

// Create handles POST /api/v1/tokens
func (h *TokenHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input models.APITokenInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
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

	// Validate permissions
	if input.Permissions != "" && input.Permissions != "read" && input.Permissions != "write" && input.Permissions != "admin" {
		ValidationErrors(w, []ValidationError{{Field: "permissions", Message: "Permissions must be 'read', 'write', or 'admin'"}})
		return
	}

	token, err := h.repo.Create(r.Context(), &input)
	if err != nil {
		InternalError(w)
		return
	}

	// Return the token with the plain text token (only time it's shown)
	Created(w, token)
}

// Get handles GET /api/v1/tokens/{id}
func (h *TokenHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid token ID")
		return
	}

	token, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			NotFound(w, "Token not found")
			return
		}
		InternalError(w)
		return
	}

	OK(w, token)
}

// Delete handles DELETE /api/v1/tokens/{id}
func (h *TokenHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid token ID")
		return
	}

	err = h.repo.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			NotFound(w, "Token not found")
			return
		}
		InternalError(w)
		return
	}

	NoContent(w)
}
