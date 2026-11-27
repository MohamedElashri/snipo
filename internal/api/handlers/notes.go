package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/MohamedElashri/snipo/internal/repository"
	"github.com/MohamedElashri/snipo/internal/services"
	"github.com/MohamedElashri/snipo/internal/utils"
	"github.com/MohamedElashri/snipo/internal/validation"
	"github.com/MohamedElashri/snipo/pkg/models"
)

func (h *SnippetHandler) ListNoteTargets(w http.ResponseWriter, r *http.Request) {
	page := 1
	var err error
	if value := r.URL.Query().Get("page"); value != "" {
		page, err = strconv.Atoi(value)
	}
	if err != nil || page <= 0 || page > 1000000 {
		Error(w, r, http.StatusBadRequest, "INVALID_PAGE", "Invalid page number")
		return
	}
	result, err := h.service.ListNoteTargets(r.Context(), r.URL.Query().Get("q"), page)
	if err != nil {
		InternalError(w, r)
		return
	}
	OK(w, r, result)
}

func (h *SnippetHandler) AppendNote(w http.ResponseWriter, r *http.Request) {
	var input models.NoteAppendInput
	if err := DecodeJSON(r, &input); err != nil {
		Error(w, r, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON payload")
		return
	}
	var invalid validation.ValidationErrors
	switch err := h.service.AppendNote(r.Context(), utils.URLParam(r, "id"), input); {
	case errors.Is(err, services.ErrSnippetNotFound):
		NotFound(w, r, "This note is no longer available for capture")
	case errors.Is(err, repository.ErrNoteTooLarge):
		Error(w, r, http.StatusRequestEntityTooLarge, "NOTE_TOO_LARGE", "This capture exceeds the note's 1 MiB size limit")
	case errors.As(err, &invalid):
		ValidationErrors(w, r, invalid)
	case err != nil:
		InternalError(w, r)
	default:
		NoContent(w)
	}
}
