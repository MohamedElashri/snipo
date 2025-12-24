package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/MohamedElashri/snipo/internal/models"
	"github.com/MohamedElashri/snipo/internal/repository"
)

// SettingsHandler handles settings related endpoints
type SettingsHandler struct {
	repo *repository.SettingsRepository
}

// NewSettingsHandler creates a new settings handler
func NewSettingsHandler(repo *repository.SettingsRepository) *SettingsHandler {
	return &SettingsHandler{repo: repo}
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

	// Basic validation could go here

	updated, err := h.repo.Update(r.Context(), &input)
	if err != nil {
		InternalError(w, r)
		return
	}

	OK(w, r, updated)
}
