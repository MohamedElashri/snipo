package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MohamedElashri/snipo/internal/models"
	"github.com/MohamedElashri/snipo/internal/repository"
)

// GistSyncService handles gist synchronization operations
type GistSyncService struct {
	githubClient  *GitHubClient
	snippetRepo   *repository.SnippetRepository
	syncRepo      *repository.GistSyncRepository
	encryptionSvc *EncryptionService
}

// NewGistSyncService creates a new gist sync service
func NewGistSyncService(
	githubClient *GitHubClient,
	snippetRepo *repository.SnippetRepository,
	syncRepo *repository.GistSyncRepository,
	encryptionSvc *EncryptionService,
) *GistSyncService {
	return &GistSyncService{
		githubClient:  githubClient,
		snippetRepo:   snippetRepo,
		syncRepo:      syncRepo,
		encryptionSvc: encryptionSvc,
	}
}

// SyncSnippetToGist syncs a snippet to its corresponding gist
func (s *GistSyncService) SyncSnippetToGist(ctx context.Context, snippetID string) error {
	snippet, err := s.snippetRepo.GetByID(ctx, snippetID)
	if err != nil {
		return fmt.Errorf("failed to get snippet: %w", err)
	}

	mapping, err := s.syncRepo.GetMapping(ctx, snippetID)
	if err != nil {
		return fmt.Errorf("failed to get mapping: %w", err)
	}

	gistReq, err := SnippetToGistRequest(snippet)
	if err != nil {
		return fmt.Errorf("failed to convert snippet to gist: %w", err)
	}

	var gist *models.GistResponse
	if mapping == nil {
		gist, err = s.githubClient.CreateGist(ctx, gistReq)
		if err != nil {
			s.logError(ctx, snippetID, "", models.SyncOpCreate, err)
			return fmt.Errorf("failed to create gist: %w", err)
		}

		checksum, _ := CalculateSnippetChecksum(snippet)
		gistChecksum, _ := CalculateGistChecksum(gist)

		mapping = &models.SnippetGistMapping{
			SnippetID:     snippetID,
			GistID:        gist.ID,
			GistURL:       gist.HTMLURL,
			SyncEnabled:   true,
			SnipoChecksum: checksum,
			GistChecksum:  gistChecksum,
			SyncStatus:    models.SyncStatusSynced,
		}
		now := time.Now()
		mapping.LastSyncedAt = &now

		if err := s.syncRepo.CreateMapping(ctx, mapping); err != nil {
			return fmt.Errorf("failed to create mapping: %w", err)
		}

		s.logSuccess(ctx, snippetID, gist.ID, models.SyncOpCreate, "Gist created successfully")
	} else {
		gist, err = s.githubClient.UpdateGist(ctx, mapping.GistID, gistReq)
		if err != nil {
			s.logError(ctx, snippetID, mapping.GistID, models.SyncOpUpdate, err)
			errMsg := err.Error()
			mapping.ErrorMessage = &errMsg
			mapping.SyncStatus = models.SyncStatusError
			s.syncRepo.UpdateMapping(ctx, mapping)
			return fmt.Errorf("failed to update gist: %w", err)
		}

		checksum, _ := CalculateSnippetChecksum(snippet)
		gistChecksum, _ := CalculateGistChecksum(gist)

		mapping.SnipoChecksum = checksum
		mapping.GistChecksum = gistChecksum
		mapping.SyncStatus = models.SyncStatusSynced
		mapping.ErrorMessage = nil
		now := time.Now()
		mapping.LastSyncedAt = &now

		if err := s.syncRepo.UpdateMapping(ctx, mapping); err != nil {
			return fmt.Errorf("failed to update mapping: %w", err)
		}

		s.logSuccess(ctx, snippetID, gist.ID, models.SyncOpUpdate, "Gist updated successfully")
	}

	return nil
}

// SyncGistToSnippet syncs a gist to its corresponding snippet
func (s *GistSyncService) SyncGistToSnippet(ctx context.Context, gistID string) error {
	mapping, err := s.syncRepo.GetMappingByGistID(ctx, gistID)
	if err != nil {
		return fmt.Errorf("failed to get mapping: %w", err)
	}
	if mapping == nil {
		return fmt.Errorf("no mapping found for gist %s", gistID)
	}

	gist, err := s.githubClient.GetGist(ctx, gistID)
	if err != nil {
		s.logError(ctx, mapping.SnippetID, gistID, models.SyncOpSync, err)
		return fmt.Errorf("failed to get gist: %w", err)
	}

	existingSnippet, err := s.snippetRepo.GetByID(ctx, mapping.SnippetID)
	if err != nil {
		return fmt.Errorf("failed to get snippet: %w", err)
	}

	snippet, err := GistToSnippet(gist, existingSnippet)
	if err != nil {
		return fmt.Errorf("failed to convert gist to snippet: %w", err)
	}

	snippetInput := &models.SnippetInput{
		Title:       snippet.Title,
		Description: snippet.Description,
		Content:     snippet.Content,
		Language:    snippet.Language,
		IsPublic:    snippet.IsPublic,
		IsArchived:  snippet.IsArchived,
		Files:       make([]models.SnippetFileInput, 0),
	}

	for _, file := range snippet.Files {
		snippetInput.Files = append(snippetInput.Files, models.SnippetFileInput{
			Filename: file.Filename,
			Content:  file.Content,
			Language: file.Language,
		})
	}

	updatedSnippet, err := s.snippetRepo.Update(ctx, mapping.SnippetID, snippetInput)
	if err != nil {
		s.logError(ctx, mapping.SnippetID, gistID, models.SyncOpUpdate, err)
		return fmt.Errorf("failed to update snippet: %w", err)
	}

	checksum, _ := CalculateSnippetChecksum(updatedSnippet)
	gistChecksum, _ := CalculateGistChecksum(gist)

	mapping.SnipoChecksum = checksum
	mapping.GistChecksum = gistChecksum
	mapping.SyncStatus = models.SyncStatusSynced
	mapping.ErrorMessage = nil
	now := time.Now()
	mapping.LastSyncedAt = &now

	if err := s.syncRepo.UpdateMapping(ctx, mapping); err != nil {
		return fmt.Errorf("failed to update mapping: %w", err)
	}

	s.logSuccess(ctx, mapping.SnippetID, gistID, models.SyncOpSync, "Snippet updated from gist")
	return nil
}

// DetectChanges detects what changed between snippet and gist
func (s *GistSyncService) DetectChanges(ctx context.Context, snippetID string) (models.SyncDirection, error) {
	mapping, err := s.syncRepo.GetMapping(ctx, snippetID)
	if err != nil {
		return models.NoSync, fmt.Errorf("failed to get mapping: %w", err)
	}
	if mapping == nil {
		return models.NoSync, fmt.Errorf("no mapping found for snippet %s", snippetID)
	}

	snippet, err := s.snippetRepo.GetByID(ctx, snippetID)
	if err != nil {
		return models.NoSync, fmt.Errorf("failed to get snippet: %w", err)
	}

	gist, err := s.githubClient.GetGist(ctx, mapping.GistID)
	if err != nil {
		return models.NoSync, fmt.Errorf("failed to get gist: %w", err)
	}

	currentSnipoChecksum, err := CalculateSnippetChecksum(snippet)
	if err != nil {
		return models.NoSync, fmt.Errorf("failed to calculate snippet checksum: %w", err)
	}

	currentGistChecksum, err := CalculateGistChecksum(gist)
	if err != nil {
		return models.NoSync, fmt.Errorf("failed to calculate gist checksum: %w", err)
	}

	snipoChanged := currentSnipoChecksum != mapping.SnipoChecksum
	gistChanged := currentGistChecksum != mapping.GistChecksum

	if !snipoChanged && !gistChanged {
		return models.NoSync, nil
	}
	if snipoChanged && !gistChanged {
		return models.SnipoToGist, nil
	}
	if !snipoChanged && gistChanged {
		return models.GistToSnipo, nil
	}
	return models.Conflict, nil
}

// SyncAll syncs all enabled mappings
func (s *GistSyncService) SyncAll(ctx context.Context) (*models.SyncResult, error) {
	startTime := time.Now()
	result := &models.SyncResult{
		ErrorMessages: make([]string, 0),
	}

	config, err := s.syncRepo.GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}
	if config == nil || !config.Enabled {
		return result, fmt.Errorf("gist sync is not enabled")
	}

	mappings, err := s.syncRepo.GetEnabledMappings(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled mappings: %w", err)
	}

	result.TotalProcessed = len(mappings)

	for _, mapping := range mappings {
		direction, err := s.DetectChanges(ctx, mapping.SnippetID)
		if err != nil {
			result.Errors++
			result.ErrorMessages = append(result.ErrorMessages, fmt.Sprintf("snippet %s: %v", mapping.SnippetID, err))
			continue
		}

		switch direction {
		case models.NoSync:
			result.Synced++
		case models.SnipoToGist:
			if err := s.SyncSnippetToGist(ctx, mapping.SnippetID); err != nil {
				result.Errors++
				result.ErrorMessages = append(result.ErrorMessages, fmt.Sprintf("snippet %s: %v", mapping.SnippetID, err))
			} else {
				result.Synced++
			}
		case models.GistToSnipo:
			if err := s.SyncGistToSnippet(ctx, mapping.GistID); err != nil {
				result.Errors++
				result.ErrorMessages = append(result.ErrorMessages, fmt.Sprintf("gist %s: %v", mapping.GistID, err))
			} else {
				result.Synced++
			}
		case models.Conflict:
			if err := s.handleConflict(ctx, mapping); err != nil {
				result.Errors++
				result.ErrorMessages = append(result.ErrorMessages, fmt.Sprintf("conflict %s: %v", mapping.SnippetID, err))
			} else {
				result.Conflicts++
			}
		}
	}

	result.Duration = time.Since(startTime).String()
	s.syncRepo.UpdateLastFullSyncTime(ctx)

	return result, nil
}

// handleConflict handles a sync conflict
func (s *GistSyncService) handleConflict(ctx context.Context, mapping *models.SnippetGistMapping) error {
	snippet, err := s.snippetRepo.GetByID(ctx, mapping.SnippetID)
	if err != nil {
		return fmt.Errorf("failed to get snippet: %w", err)
	}

	gist, err := s.githubClient.GetGist(ctx, mapping.GistID)
	if err != nil {
		return fmt.Errorf("failed to get gist: %w", err)
	}

	snipoVersion, err := json.Marshal(snippet)
	if err != nil {
		return fmt.Errorf("failed to marshal snippet: %w", err)
	}

	gistVersion, err := json.Marshal(gist)
	if err != nil {
		return fmt.Errorf("failed to marshal gist: %w", err)
	}

	conflict := &models.GistSyncConflict{
		SnippetID:    mapping.SnippetID,
		GistID:       mapping.GistID,
		SnipoVersion: string(snipoVersion),
		GistVersion:  string(gistVersion),
	}

	if err := s.syncRepo.CreateConflict(ctx, conflict); err != nil {
		return fmt.Errorf("failed to create conflict: %w", err)
	}

	mapping.SyncStatus = models.SyncStatusConflict
	if err := s.syncRepo.UpdateMapping(ctx, mapping); err != nil {
		return fmt.Errorf("failed to update mapping: %w", err)
	}

	s.logSuccess(ctx, mapping.SnippetID, mapping.GistID, models.SyncOpConflict, "Conflict detected")
	return nil
}

// ResolveConflict resolves a conflict with the given strategy
func (s *GistSyncService) ResolveConflict(ctx context.Context, conflictID int64, resolution string) error {
	conflict, err := s.syncRepo.GetConflict(ctx, conflictID)
	if err != nil {
		return fmt.Errorf("failed to get conflict: %w", err)
	}
	if conflict == nil {
		return fmt.Errorf("conflict not found")
	}

	switch resolution {
	case models.ConflictStrategySnipoWins:
		if err := s.SyncSnippetToGist(ctx, conflict.SnippetID); err != nil {
			return fmt.Errorf("failed to sync snippet to gist: %w", err)
		}
	case models.ConflictStrategyGistWins:
		if err := s.SyncGistToSnippet(ctx, conflict.GistID); err != nil {
			return fmt.Errorf("failed to sync gist to snippet: %w", err)
		}
	default:
		return fmt.Errorf("invalid resolution strategy: %s", resolution)
	}

	if err := s.syncRepo.ResolveConflict(ctx, conflictID, resolution); err != nil {
		return fmt.Errorf("failed to resolve conflict: %w", err)
	}

	return nil
}

// EnableSyncForSnippet enables sync for a snippet
func (s *GistSyncService) EnableSyncForSnippet(ctx context.Context, snippetID string) error {
	mapping, err := s.syncRepo.GetMapping(ctx, snippetID)
	if err != nil {
		return fmt.Errorf("failed to get mapping: %w", err)
	}
	if mapping == nil {
		return s.SyncSnippetToGist(ctx, snippetID)
	}

	mapping.SyncEnabled = true
	if err := s.syncRepo.UpdateMapping(ctx, mapping); err != nil {
		return fmt.Errorf("failed to update mapping: %w", err)
	}

	return nil
}

// DisableSyncForSnippet disables sync for a snippet
func (s *GistSyncService) DisableSyncForSnippet(ctx context.Context, snippetID string) error {
	mapping, err := s.syncRepo.GetMapping(ctx, snippetID)
	if err != nil {
		return fmt.Errorf("failed to get mapping: %w", err)
	}
	if mapping == nil {
		return fmt.Errorf("no mapping found for snippet %s", snippetID)
	}

	mapping.SyncEnabled = false
	if err := s.syncRepo.UpdateMapping(ctx, mapping); err != nil {
		return fmt.Errorf("failed to update mapping: %w", err)
	}

	return nil
}

// logSuccess logs a successful sync operation
func (s *GistSyncService) logSuccess(ctx context.Context, snippetID, gistID, operation, message string) {
	log := &models.GistSyncLog{
		SnippetID: &snippetID,
		GistID:    &gistID,
		Operation: operation,
		Status:    models.SyncOpStatusSuccess,
		Message:   &message,
	}
	s.syncRepo.CreateLog(ctx, log)
}

// logError logs a failed sync operation
func (s *GistSyncService) logError(ctx context.Context, snippetID, gistID, operation string, err error) {
	message := err.Error()
	log := &models.GistSyncLog{
		SnippetID: &snippetID,
		GistID:    &gistID,
		Operation: operation,
		Status:    models.SyncOpStatusFailed,
		Message:   &message,
	}
	s.syncRepo.CreateLog(ctx, log)
}
