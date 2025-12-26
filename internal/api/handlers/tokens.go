package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/MohamedElashri/snipo/internal/auth"
	"github.com/MohamedElashri/snipo/internal/models"
	"github.com/MohamedElashri/snipo/internal/repository"
	"github.com/MohamedElashri/snipo/internal/validation"
)

// TokenHandler handles API token-related HTTP requests
type TokenHandler struct {
	repo         *repository.TokenRepository
	settingsRepo *repository.SettingsRepository
	authService  *auth.Service
}

// NewTokenHandler creates a new token handler
func NewTokenHandler(repo *repository.TokenRepository, settingsRepo *repository.SettingsRepository, authService *auth.Service) *TokenHandler {
	return &TokenHandler{
		repo:         repo,
		settingsRepo: settingsRepo,
		authService:  authService,
	}
}

// List handles GET /api/v1/tokens
func (h *TokenHandler) List(w http.ResponseWriter, r *http.Request) {
	tokens, err := h.repo.List(r.Context())
	if err != nil {
		InternalError(w, r)
		return
	}

	OK(w, r, map[string]interface{}{"data": tokens})
}

// Create handles POST /api/v1/tokens
func (h *TokenHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input models.APITokenInput
	if err := DecodeJSON(r, &input); err != nil {
		// Provide more detailed error message for debugging
		Error(w, r, http.StatusBadRequest, "INVALID_JSON", fmt.Sprintf("Invalid JSON payload: %v", err))
		return
	}

	// Always require password for token creation (unless auth is completely disabled)
	if h.authService != nil && !h.authService.IsAuthDisabled() {
		if input.Password == "" {
			Error(w, r, http.StatusUnauthorized, "PASSWORD_REQUIRED", "Password is required to create API tokens")
			return
		}
		// Verify password
		if !h.authService.VerifyPassword(input.Password) {
			Error(w, r, http.StatusUnauthorized, "INVALID_PASSWORD", "Invalid password")
			return
		}
	}

	// Validate input
	if input.Name == "" {
		ValidationErrors(w, r, validation.ValidationErrors{validation.ValidationError{Field: "name", Message: "Name is required"}})
		return
	}

	if len(input.Name) > 100 {
		ValidationErrors(w, r, validation.ValidationErrors{validation.ValidationError{Field: "name", Message: "Name must be 100 characters or less"}})
		return
	}

	// Validate permissions
	if input.Permissions != "" && input.Permissions != "read" && input.Permissions != "write" && input.Permissions != "admin" {
		ValidationErrors(w, r, validation.ValidationErrors{validation.ValidationError{Field: "permissions", Message: "Permissions must be 'read', 'write', or 'admin'"}})
		return
	}

	token, err := h.repo.Create(r.Context(), &input)
	if err != nil {
		InternalError(w, r)
		return
	}

	// Return the token with the plain text token (only time it's shown)
	Created(w, r, token)
}

// Get handles GET /api/v1/tokens/{id}
func (h *TokenHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Error(w, r, http.StatusBadRequest, "INVALID_ID", "Invalid token ID")
		return
	}

	token, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			NotFound(w, r, "Token not found")
			return
		}
		InternalError(w, r)
		return
	}

	OK(w, r, token)
}

// Delete handles DELETE /api/v1/tokens/{id}
func (h *TokenHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Error(w, r, http.StatusBadRequest, "INVALID_ID", "Invalid token ID")
		return
	}

	// Always require password for token deletion (unless auth is completely disabled)
	if h.authService != nil && !h.authService.IsAuthDisabled() {
		var input struct {
			Password string `json:"password"`
		}
		if err := DecodeJSON(r, &input); err != nil || input.Password == "" {
			Error(w, r, http.StatusUnauthorized, "PASSWORD_REQUIRED", "Password is required to delete API tokens")
			return
		}
		// Verify password
		if !h.authService.VerifyPassword(input.Password) {
			Error(w, r, http.StatusUnauthorized, "INVALID_PASSWORD", "Invalid password")
			return
		}
	}

	err = h.repo.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			NotFound(w, r, "Token not found")
			return
		}
		InternalError(w, r)
		return
	}

	NoContent(w)
}
