package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/MohamedElashri/snipo/internal/models"
	"github.com/MohamedElashri/snipo/internal/repository"
	"github.com/MohamedElashri/snipo/internal/services"
	"github.com/go-chi/chi/v5"
)

// GistSyncHandler handles gist sync related endpoints
type GistSyncHandler struct {
	syncRepo      *repository.GistSyncRepository
	snippetRepo   *repository.SnippetRepository
	fileRepo      *repository.SnippetFileRepository
	encryptionSvc *services.EncryptionService
}

// NewGistSyncHandler creates a new gist sync handler
func NewGistSyncHandler(
	syncRepo *repository.GistSyncRepository,
	snippetRepo *repository.SnippetRepository,
	fileRepo *repository.SnippetFileRepository,
	encryptionSvc *services.EncryptionService,
) *GistSyncHandler {
	return &GistSyncHandler{
		syncRepo:      syncRepo,
		snippetRepo:   snippetRepo,
		fileRepo:      fileRepo,
		encryptionSvc: encryptionSvc,
	}
}

// ConfigInput represents the input for configuring gist sync
type ConfigInput struct {
	Enabled                    bool   `json:"enabled"`
	GithubToken                string `json:"github_token"`
	AutoSyncEnabled            bool   `json:"auto_sync_enabled"`
	SyncIntervalMinutes        int    `json:"sync_interval_minutes"`
	ConflictResolutionStrategy string `json:"conflict_resolution_strategy"`
}

// ConfigResponse represents the gist sync configuration response (token masked)
type ConfigResponse struct {
	Enabled                    bool   `json:"enabled"`
	GithubUsername             string `json:"github_username"`
	HasToken                   bool   `json:"has_token"`
	AutoSyncEnabled            bool   `json:"auto_sync_enabled"`
	SyncIntervalMinutes        int    `json:"sync_interval_minutes"`
	ConflictResolutionStrategy string `json:"conflict_resolution_strategy"`
	LastFullSyncAt             string `json:"last_full_sync_at,omitempty"`
}

// GetConfig retrieves the gist sync configuration
func (h *GistSyncHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	config, err := h.syncRepo.GetConfig(r.Context())
	if err != nil {
		InternalError(w, r)
		return
	}

	if config == nil {
		OK(w, r, ConfigResponse{
			Enabled:                    false,
			HasToken:                   false,
			AutoSyncEnabled:            true,
			SyncIntervalMinutes:        15,
			ConflictResolutionStrategy: models.ConflictStrategyManual,
		})
		return
	}

	response := ConfigResponse{
		Enabled:                    config.Enabled,
		GithubUsername:             config.GithubUsername,
		HasToken:                   config.GithubTokenEncrypted != "",
		AutoSyncEnabled:            config.AutoSyncEnabled,
		SyncIntervalMinutes:        config.SyncIntervalMinutes,
		ConflictResolutionStrategy: config.ConflictResolutionStrategy,
	}

	if config.LastFullSyncAt != nil {
		response.LastFullSyncAt = config.LastFullSyncAt.Format("2006-01-02 15:04:05")
	}

	OK(w, r, response)
}

// UpdateConfig updates the gist sync configuration
func (h *GistSyncHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var input ConfigInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		Error(w, r, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	if input.SyncIntervalMinutes < 5 {
		Error(w, r, http.StatusBadRequest, "INVALID_INTERVAL", "Sync interval must be at least 5 minutes")
		return
	}

	validStrategies := map[string]bool{
		models.ConflictStrategyManual:     true,
		models.ConflictStrategySnipoWins:  true,
		models.ConflictStrategyGistWins:   true,
		models.ConflictStrategyNewestWins: true,
	}
	if !validStrategies[input.ConflictResolutionStrategy] {
		Error(w, r, http.StatusBadRequest, "INVALID_STRATEGY", "Invalid conflict resolution strategy")
		return
	}

	var encryptedToken string
	var username string

	if input.GithubToken != "" {
		githubClient := services.NewGitHubClient(input.GithubToken)
		var err error
		username, err = githubClient.GetAuthenticatedUser(r.Context())
		if err != nil {
			// Log detailed error for debugging
			if logger := r.Context().Value("logger"); logger != nil {
				logger.(*slog.Logger).Error("failed to validate GitHub token",
					"error", err,
					"token_prefix", input.GithubToken[:min(10, len(input.GithubToken))])
			}
			Error(w, r, http.StatusBadRequest, "INVALID_TOKEN", fmt.Sprintf("Failed to validate GitHub token: %v", err))
			return
		}

		encryptedToken, err = h.encryptionSvc.Encrypt(input.GithubToken)
		if err != nil {
			if logger := r.Context().Value("logger"); logger != nil {
				logger.(*slog.Logger).Error("failed to encrypt token", "error", err)
			}
			InternalError(w, r)
			return
		}
	} else {
		existingConfig, err := h.syncRepo.GetConfig(r.Context())
		if err != nil {
			InternalError(w, r)
			return
		}
		if existingConfig != nil {
			encryptedToken = existingConfig.GithubTokenEncrypted
			username = existingConfig.GithubUsername
		}
	}

	config := &models.GistSyncConfig{
		Enabled:                    input.Enabled,
		GithubTokenEncrypted:       encryptedToken,
		GithubUsername:             username,
		AutoSyncEnabled:            input.AutoSyncEnabled,
		SyncIntervalMinutes:        input.SyncIntervalMinutes,
		ConflictResolutionStrategy: input.ConflictResolutionStrategy,
	}

	if err := h.syncRepo.CreateOrUpdateConfig(r.Context(), config); err != nil {
		InternalError(w, r)
		return
	}

	OK(w, r, map[string]interface{}{
		"message":  "Configuration updated successfully",
		"username": username,
	})
}

// TestConnection tests the GitHub token validity
func (h *GistSyncHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	config, err := h.syncRepo.GetConfig(r.Context())
	if err != nil {
		InternalError(w, r)
		return
	}

	if config == nil || config.GithubTokenEncrypted == "" {
		Error(w, r, http.StatusBadRequest, "NO_TOKEN", "No GitHub token configured")
		return
	}

	token, err := h.encryptionSvc.Decrypt(config.GithubTokenEncrypted)
	if err != nil {
		InternalError(w, r)
		return
	}

	githubClient := services.NewGitHubClient(token)
	username, err := githubClient.GetAuthenticatedUser(r.Context())
	if err != nil {
		Error(w, r, http.StatusBadRequest, "INVALID_TOKEN", "GitHub token is invalid or expired")
		return
	}

	OK(w, r, map[string]interface{}{
		"valid":    true,
		"username": username,
		"message":  "Connection successful",
	})
}

// ClearConfig clears the GitHub token and disables sync
func (h *GistSyncHandler) ClearConfig(w http.ResponseWriter, r *http.Request) {
	if err := h.syncRepo.DeleteConfig(r.Context()); err != nil {
		InternalError(w, r)
		return
	}

	OK(w, r, map[string]string{
		"message": "Configuration cleared successfully",
	})
}

// SyncSnippet syncs a specific snippet to gist
func (h *GistSyncHandler) SyncSnippet(w http.ResponseWriter, r *http.Request) {
	snippetID := chi.URLParam(r, "id")
	if snippetID == "" {
		Error(w, r, http.StatusBadRequest, "MISSING_ID", "Snippet ID is required")
		return
	}

	syncService, err := h.createSyncService(r.Context())
	if err != nil {
		Error(w, r, http.StatusBadRequest, "SYNC_NOT_CONFIGURED", err.Error())
		return
	}

	if err := syncService.SyncSnippetToGist(r.Context(), snippetID); err != nil {
		Error(w, r, http.StatusInternalServerError, "SYNC_FAILED", err.Error())
		return
	}

	OK(w, r, map[string]string{
		"message": "Snippet synced successfully",
	})
}

// SyncAll syncs all enabled snippets
func (h *GistSyncHandler) SyncAll(w http.ResponseWriter, r *http.Request) {
	syncService, err := h.createSyncService(r.Context())
	if err != nil {
		Error(w, r, http.StatusBadRequest, "SYNC_NOT_CONFIGURED", err.Error())
		return
	}

	result, err := syncService.SyncAll(r.Context())
	if err != nil {
		Error(w, r, http.StatusInternalServerError, "SYNC_FAILED", err.Error())
		return
	}

	OK(w, r, result)
}

// EnableSync enables sync for a snippet
func (h *GistSyncHandler) EnableSync(w http.ResponseWriter, r *http.Request) {
	snippetID := chi.URLParam(r, "id")
	if snippetID == "" {
		Error(w, r, http.StatusBadRequest, "MISSING_ID", "Snippet ID is required")
		return
	}

	syncService, err := h.createSyncService(r.Context())
	if err != nil {
		Error(w, r, http.StatusBadRequest, "SYNC_NOT_CONFIGURED", err.Error())
		return
	}

	if err := syncService.EnableSyncForSnippet(r.Context(), snippetID); err != nil {
		Error(w, r, http.StatusInternalServerError, "ENABLE_FAILED", err.Error())
		return
	}

	OK(w, r, map[string]string{
		"message": "Sync enabled for snippet",
	})
}

// DisableSync disables sync for a snippet
func (h *GistSyncHandler) DisableSync(w http.ResponseWriter, r *http.Request) {
	snippetID := chi.URLParam(r, "id")
	if snippetID == "" {
		Error(w, r, http.StatusBadRequest, "MISSING_ID", "Snippet ID is required")
		return
	}

	syncService, err := h.createSyncService(r.Context())
	if err != nil {
		Error(w, r, http.StatusBadRequest, "SYNC_NOT_CONFIGURED", err.Error())
		return
	}

	if err := syncService.DisableSyncForSnippet(r.Context(), snippetID); err != nil {
		Error(w, r, http.StatusInternalServerError, "DISABLE_FAILED", err.Error())
		return
	}

	OK(w, r, map[string]string{
		"message": "Sync disabled for snippet",
	})
}

// EnableSyncForAll enables sync for all snippets
func (h *GistSyncHandler) EnableSyncForAll(w http.ResponseWriter, r *http.Request) {
	syncService, err := h.createSyncService(r.Context())
	if err != nil {
		Error(w, r, http.StatusBadRequest, "SYNC_NOT_CONFIGURED", err.Error())
		return
	}

	// Get all snippets using List with high limit
	result, err := h.snippetRepo.List(r.Context(), models.SnippetFilter{
		Limit: 10000, // High limit to get all snippets
	})
	if err != nil {
		Error(w, r, http.StatusInternalServerError, "FETCH_FAILED", "Failed to fetch snippets")
		return
	}

	enabled := 0
	errors := 0
	errorMessages := []string{}

	for _, snippet := range result.Data {
		if err := syncService.EnableSyncForSnippet(r.Context(), snippet.ID); err != nil {
			errors++
			errorMessages = append(errorMessages, fmt.Sprintf("%s: %v", snippet.ID, err))
		} else {
			enabled++
		}
	}

	OK(w, r, map[string]interface{}{
		"message":        fmt.Sprintf("Enabled sync for %d snippets", enabled),
		"enabled":        enabled,
		"errors":         errors,
		"error_messages": errorMessages,
	})
}

// ListMappings lists all snippet-gist mappings
func (h *GistSyncHandler) ListMappings(w http.ResponseWriter, r *http.Request) {
	mappings, err := h.syncRepo.ListMappings(r.Context())
	if err != nil {
		InternalError(w, r)
		return
	}

	OK(w, r, mappings)
}

// DeleteMapping deletes a snippet-gist mapping
func (h *GistSyncHandler) DeleteMapping(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, r, http.StatusBadRequest, "INVALID_ID", "Invalid mapping ID")
		return
	}

	if err := h.syncRepo.DeleteMapping(r.Context(), id); err != nil {
		InternalError(w, r)
		return
	}

	OK(w, r, map[string]string{
		"message": "Mapping deleted successfully",
	})
}

// ListConflicts lists all unresolved conflicts
func (h *GistSyncHandler) ListConflicts(w http.ResponseWriter, r *http.Request) {
	conflicts, err := h.syncRepo.ListConflicts(r.Context(), false)
	if err != nil {
		InternalError(w, r)
		return
	}

	OK(w, r, conflicts)
}

// ResolveConflict resolves a conflict
func (h *GistSyncHandler) ResolveConflict(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, r, http.StatusBadRequest, "INVALID_ID", "Invalid conflict ID")
		return
	}

	var input struct {
		Resolution string `json:"resolution"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		Error(w, r, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	validResolutions := map[string]bool{
		models.ConflictStrategySnipoWins: true,
		models.ConflictStrategyGistWins:  true,
	}
	if !validResolutions[input.Resolution] {
		Error(w, r, http.StatusBadRequest, "INVALID_RESOLUTION", "Invalid resolution choice")
		return
	}

	syncService, err := h.createSyncService(r.Context())
	if err != nil {
		Error(w, r, http.StatusBadRequest, "SYNC_NOT_CONFIGURED", err.Error())
		return
	}

	if err := syncService.ResolveConflict(r.Context(), id, input.Resolution); err != nil {
		Error(w, r, http.StatusInternalServerError, "RESOLVE_FAILED", err.Error())
		return
	}

	OK(w, r, map[string]string{
		"message": "Conflict resolved successfully",
	})
}

// GetLogs retrieves sync operation logs
func (h *GistSyncHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 200 {
			limit = l
		}
	}

	logs, err := h.syncRepo.ListLogs(r.Context(), limit)
	if err != nil {
		InternalError(w, r)
		return
	}

	OK(w, r, logs)
}

// createSyncService creates a sync service with the current configuration
func (h *GistSyncHandler) createSyncService(ctx context.Context) (*services.GistSyncService, error) {
	config, err := h.syncRepo.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	if config == nil || config.GithubTokenEncrypted == "" {
		return nil, fmt.Errorf("github token not configured")
	}

	token, err := h.encryptionSvc.Decrypt(config.GithubTokenEncrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt token: %w", err)
	}

	githubClient := services.NewGitHubClient(token)
	return services.NewGistSyncService(githubClient, h.snippetRepo, h.fileRepo, h.syncRepo, h.encryptionSvc), nil
}
