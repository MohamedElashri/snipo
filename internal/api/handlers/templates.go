package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"regexp"
	"strconv"

	"github.com/MohamedElashri/snipo/internal/repository"
	"github.com/MohamedElashri/snipo/internal/services"
	"github.com/MohamedElashri/snipo/internal/utils"
	"github.com/MohamedElashri/snipo/internal/validation"
	"github.com/MohamedElashri/snipo/pkg/models"
)

type TemplateHandler struct {
	service *services.TemplateService
}

func NewTemplateHandler(service *services.TemplateService) *TemplateHandler {
	return &TemplateHandler{service: service}
}

func templateResult(w http.ResponseWriter, r *http.Request, err error) bool {
	if err == nil {
		return true
	}
	var fields validation.ValidationErrors
	switch {
	case errors.As(err, &fields):
		ValidationErrors(w, r, fields)
	case errors.Is(err, sql.ErrNoRows):
		NotFound(w, r, "Template or revision unavailable")
	case errors.Is(err, repository.ErrTemplateConflict):
		Error(w, r, http.StatusConflict, "TEMPLATE_CONFLICT", err.Error())
	default:
		InternalError(w, r)
	}
	return false
}

func decodeTemplate(w http.ResponseWriter, r *http.Request, input any) bool {
	if err := DecodeJSON(r, input); err != nil {
		Error(w, r, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON payload")
		return false
	}
	return true
}

func (h *TemplateHandler) List(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	state := params.Get("state")
	if state == "" {
		state = "active"
	}
	if state != "active" && state != "archived" && state != "trash" {
		Error(w, r, http.StatusBadRequest, "INVALID_STATE", "Choose active, archived, or trash")
		return
	}
	page, _ := strconv.Atoi(params.Get("page"))
	limit, _ := strconv.Atoi(params.Get("limit"))
	if page < 1 || page > 1000000 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	items, total, err := h.service.List(r.Context(), state, params.Get("q"), limit, (page-1)*limit)
	if templateResult(w, r, err) {
		SuccessList(w, r, items, page, limit, total)
	}
}

func (h *TemplateHandler) Get(w http.ResponseWriter, r *http.Request) {
	var revision int
	if value := r.URL.Query().Get("revision"); value != "" {
		var err error
		revision, err = strconv.Atoi(value)
		if err != nil || revision < 1 {
			Error(w, r, http.StatusBadRequest, "INVALID_REVISION", "Choose a valid revision")
			return
		}
	}
	item, err := h.service.Get(r.Context(), utils.URLParam(r, "id"), revision)
	if templateResult(w, r, err) {
		OK(w, r, item)
	}
}

func (h *TemplateHandler) Create(w http.ResponseWriter, r *http.Request) {
	var def models.TemplateDefinition
	if !decodeTemplate(w, r, &def) {
		return
	}
	item, err := h.service.SaveDefinition(r.Context(), "", 0, def)
	if templateResult(w, r, err) {
		Created(w, r, item)
	}
}

func (h *TemplateHandler) Update(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Revision   int                       `json:"revision"`
		Definition models.TemplateDefinition `json:"definition"`
	}
	if !decodeTemplate(w, r, &input) {
		return
	}
	if input.Revision < 1 {
		Error(w, r, http.StatusBadRequest, "INVALID_REVISION", "Specify the revision you opened")
		return
	}
	item, err := h.service.SaveDefinition(r.Context(), utils.URLParam(r, "id"), input.Revision, input.Definition)
	if templateResult(w, r, err) {
		OK(w, r, item)
	}
}

func (h *TemplateHandler) Status(w http.ResponseWriter, r *http.Request) {
	var input struct {
		State    *string `json:"state"`
		Favorite *bool   `json:"favorite"`
	}
	if !decodeTemplate(w, r, &input) {
		return
	}
	if input.State != nil && *input.State != "active" && *input.State != "archived" && *input.State != "trash" {
		Error(w, r, http.StatusBadRequest, "INVALID_STATE", "Choose active, archived, or trash")
		return
	}
	if templateResult(w, r, h.service.SetStatus(r.Context(), utils.URLParam(r, "id"), input.State, input.Favorite)) {
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *TemplateHandler) Purge(w http.ResponseWriter, r *http.Request) {
	if templateResult(w, r, h.service.Purge(r.Context(), utils.URLParam(r, "id"))) {
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *TemplateHandler) History(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.History(r.Context(), utils.URLParam(r, "id"))
	if err == nil && len(items) == 0 {
		err = sql.ErrNoRows
	}
	if templateResult(w, r, err) {
		OK(w, r, items)
	}
}

func (h *TemplateHandler) RestoreRevision(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Revision        int `json:"revision"`
		CurrentRevision int `json:"current_revision"`
	}
	if !decodeTemplate(w, r, &input) {
		return
	}
	if input.Revision < 1 || input.CurrentRevision < 1 {
		Error(w, r, http.StatusBadRequest, "INVALID_REVISION", "Specify the current revision and the revision to restore")
		return
	}
	id := utils.URLParam(r, "id")
	item, err := h.service.Get(r.Context(), id, input.Revision)
	if !templateResult(w, r, err) {
		return
	}
	item, err = h.service.RestoreDefinition(r.Context(), id, input.CurrentRevision, item.Definition)
	if templateResult(w, r, err) {
		OK(w, r, item)
	}
}

func (h *TemplateHandler) Render(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Revision int               `json:"revision"`
		Values   map[string]string `json:"values"`
	}
	if !decodeTemplate(w, r, &input) {
		return
	}
	item, err := h.service.Render(r.Context(), utils.URLParam(r, "id"), input.Revision, input.Values)
	if templateResult(w, r, err) {
		OK(w, r, item)
	}
}

func (h *TemplateHandler) Test(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Definition models.TemplateDefinition `json:"definition"`
		Values     map[string]string         `json:"values"`
	}
	if !decodeTemplate(w, r, &input) {
		return
	}
	item, err := h.service.Test(input.Definition, input.Values)
	if templateResult(w, r, err) {
		OK(w, r, item)
	}
}

func (h *TemplateHandler) Starters(w http.ResponseWriter, r *http.Request) {
	OK(w, r, services.TemplateStarters())
}

func (h *TemplateHandler) Export(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.Get(r.Context(), utils.URLParam(r, "id"), 0)
	if !templateResult(w, r, err) {
		return
	}
	def := item.Definition
	def.FolderID, def.Presets = nil, nil
	OK(w, r, models.TemplateExport{Version: 1, ID: item.ID, Definition: def})
}

var portableTemplateID = regexp.MustCompile(`^[a-f0-9]{32}$`)

func (h *TemplateHandler) Import(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Template models.TemplateExport `json:"template"`
		Mode     string                `json:"mode"`
		Revision int                   `json:"revision,omitempty"`
	}
	if !decodeTemplate(w, r, &input) {
		return
	}
	if input.Template.Version != 1 || (input.Template.ID != "" && !portableTemplateID.MatchString(input.Template.ID)) {
		Error(w, r, http.StatusBadRequest, "INVALID_TEMPLATE_EXPORT", "Choose a supported Snipo template export")
		return
	}
	def := input.Template.Definition
	def.FolderID, def.Presets = nil, nil
	if !templateResult(w, r, h.service.Validate(&def)) {
		return
	}
	id := input.Template.ID
	var existing *models.Template
	if id != "" {
		item, err := h.service.Get(r.Context(), id, 0)
		if err == nil {
			existing = item
		} else if !errors.Is(err, sql.ErrNoRows) {
			templateResult(w, r, err)
			return
		}
	}
	switch input.Mode {
	case "preview":
		OK(w, r, struct {
			Definition models.TemplateDefinition `json:"definition"`
			Conflict   *models.Template          `json:"conflict,omitempty"`
		}{def, existing})
	case "copy":
		if existing != nil {
			id = ""
		}
		item, err := h.service.SaveDefinition(r.Context(), id, 0, def)
		if templateResult(w, r, err) {
			Created(w, r, item)
		}
	case "replace":
		if existing == nil || existing.Revision != input.Revision {
			templateResult(w, r, repository.ErrTemplateConflict)
			return
		}
		item, err := h.service.ReplacePortableDefinition(r.Context(), existing, def)
		if templateResult(w, r, err) {
			OK(w, r, item)
		}
	default:
		Error(w, r, http.StatusBadRequest, "INVALID_IMPORT_MODE", "Choose preview, copy, or replace")
	}
}
