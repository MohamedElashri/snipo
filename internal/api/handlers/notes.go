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
	kind := r.URL.Query().Get("kind")
	if kind != "" && kind != models.KindNote && kind != models.KindSnippet {
		Error(w, r, http.StatusBadRequest, "INVALID_KIND", "Choose note or snippet")
		return
	}
	result, err := h.service.ListNoteTargets(r.Context(), r.URL.Query().Get("q"), page, kind)
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

func readItemKind(w http.ResponseWriter, r *http.Request) (string, bool) {
	switch kind := r.URL.Query().Get("kind"); kind {
	case "":
		return models.KindSnippet, true
	case "all":
		return "", true
	case models.KindNote, models.KindSnippet:
		return kind, true
	default:
		Error(w, r, http.StatusBadRequest, "INVALID_KIND", "Choose all, note, or snippet")
		return "", false
	}
}

func writeNoteError(w http.ResponseWriter, r *http.Request, err error) bool {
	switch {
	case errors.Is(err, repository.ErrNoteConflict):
		Error(w, r, http.StatusConflict, "NOTE_CONFLICT", err.Error())
	case errors.Is(err, repository.ErrNoteOverview):
		Error(w, r, http.StatusConflict, "NOTE_OVERVIEW", err.Error())
	case errors.Is(err, repository.ErrNoteConversion):
		Error(w, r, http.StatusUnprocessableEntity, "NOTE_CONVERSION", err.Error())
	default:
		return false
	}
	return true
}

func (h *SnippetHandler) ConvertKind(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Kind string `json:"kind"`
	}
	if err := DecodeJSON(r, &input); err != nil {
		Error(w, r, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON payload")
		return
	}
	item, err := h.service.ConvertKind(r.Context(), utils.URLParam(r, "id"), input.Kind)
	if err != nil {
		var invalid validation.ValidationErrors
		switch {
		case writeNoteError(w, r, err):
		case errors.Is(err, services.ErrSnippetNotFound):
			NotFound(w, r, "Item not found")
		case errors.As(err, &invalid):
			ValidationErrors(w, r, invalid)
		default:
			InternalError(w, r)
		}
		return
	}
	OK(w, r, item)
}
