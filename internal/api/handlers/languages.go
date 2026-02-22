package handlers

import (
	"net/http"
	"sort"

	"github.com/MohamedElashri/snipo/internal/validation"
)

// LanguageHandler handles terminology / metadata requests
type LanguageHandler struct{}

// NewLanguageHandler creates a new LanguageHandler
func NewLanguageHandler() *LanguageHandler {
	return &LanguageHandler{}
}

// GetLanguages returns the list of allowed snippet programming languages
func (h *LanguageHandler) GetLanguages(w http.ResponseWriter, r *http.Request) {
	langs := validation.GetAllowedLanguages()
	sort.Strings(langs) // Sort for consistent UI rendering

	response := map[string][]string{
		"languages": langs,
	}

	Success(w, r, http.StatusOK, response)
}
