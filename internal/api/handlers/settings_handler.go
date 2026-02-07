package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/MohamedElashri/snipo/internal/auth"
	"github.com/MohamedElashri/snipo/internal/models"
	"github.com/MohamedElashri/snipo/internal/repository"
)

// SettingsHandler handles settings related endpoints
type SettingsHandler struct {
	repo        *repository.SettingsRepository
	authService *auth.Service
}

// NewSettingsHandler creates a new settings handler
func NewSettingsHandler(repo *repository.SettingsRepository, authService *auth.Service) *SettingsHandler {
	return &SettingsHandler{repo: repo, authService: authService}
}

// Get retrieves application settings
func (h *SettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	settings, err := h.repo.Get(r.Context())
	if err != nil {
		InternalError(w, r)
		return
	}

	OK(w, r, settings)
}

// Update updates application settings
func (h *SettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	var input models.SettingsInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		Error(w, r, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	// If disable_login is being changed, require password verification
	if h.authService != nil && !h.authService.IsAuthDisabled() {
		current, err := h.repo.Get(r.Context())
		if err == nil && current.DisableLogin != input.DisableLogin {
			if input.Password == "" {
				Error(w, r, http.StatusForbidden, "PASSWORD_REQUIRED", "Password is required to change the login setting")
				return
			}
			if !h.authService.VerifyPassword(input.Password) {
				Error(w, r, http.StatusForbidden, "INVALID_PASSWORD", "Invalid password")
				return
			}
		}
	}

	updated, err := h.repo.Update(r.Context(), &input)
	if err != nil {
		InternalError(w, r)
		return
	}

	OK(w, r, updated)
}
