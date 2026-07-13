package services

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/MohamedElashri/snipo/internal/models"
	"github.com/MohamedElashri/snipo/internal/storage"
)

// S3SyncService handles S3 backup operations
type S3SyncService struct {
	storage   *storage.S3Storage
	backupSvc *BackupService
	logger    *slog.Logger
}

// NewS3SyncService creates a new S3 sync service
func NewS3SyncService(storage *storage.S3Storage, backupSvc *BackupService, logger *slog.Logger) *S3SyncService {
	return &S3SyncService{
		storage:   storage,
		backupSvc: backupSvc,
		logger:    logger,
	}
}

// SyncToS3 uploads a backup to S3
func (s *S3SyncService) SyncToS3(ctx context.Context, opts models.ExportOptions) (*models.S3SyncResult, error) {
	result := &models.S3SyncResult{
		StartedAt: time.Now().UTC(),
	}

	// Create backup
	content, filename, err := s.backupSvc.Export(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup: %w", err)
	}

	// Determine content type
	contentType := "application/json"
	if strings.HasSuffix(filename, ".zip") || strings.HasSuffix(filename, ".zip.enc") {
		contentType = "application/zip"
	}
	if strings.HasSuffix(filename, ".enc") {
		contentType = "application/octet-stream"
	}

	// Upload to S3
	key := "backups/" + filename
	if err := s.storage.Upload(ctx, key, content, contentType); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to upload: %v", err))
		result.FinishedAt = time.Now().UTC()
		return result, fmt.Errorf("failed to upload backup: %w", err)
	}

	result.Uploaded = 1
	result.FinishedAt = time.Now().UTC()

	s.logger.Info("backup synced to S3",
		"size", len(content),
		"duration", result.FinishedAt.Sub(result.StartedAt),
	)

	return result, nil
}

// ListBackups returns all backups stored in S3
func (s *S3SyncService) ListBackups(ctx context.Context) ([]models.S3BackupInfo, error) {
	objects, err := s.storage.List(ctx, "backups/")
	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}

	var backups []models.S3BackupInfo
	for _, obj := range objects {
		backups = append(backups, models.S3BackupInfo{
			Key:          obj.Key,
			Size:         obj.Size,
			LastModified: obj.LastModified,
		})
	}

	return backups, nil
}

// RestoreFromS3 downloads and restores a backup from S3
func (s *S3SyncService) RestoreFromS3(ctx context.Context, key string, opts models.ImportOptions) (*models.S3RestoreResult, error) {
	result := &models.S3RestoreResult{
		StartedAt: time.Now().UTC(),
	}

	// Download backup from S3
	content, err := s.storage.Download(ctx, key)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to download: %v", err))
		result.FinishedAt = time.Now().UTC()
		return result, fmt.Errorf("failed to download backup: %w", err)
	}

	// Import backup
	importResult, err := s.backupSvc.Import(ctx, content, opts)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to import: %v", err))
		result.FinishedAt = time.Now().UTC()
		return result, fmt.Errorf("failed to import backup: %w", err)
	}

	result.Restored = importResult.SnippetsImported + importResult.TagsImported + importResult.FoldersImported
	result.Errors = append(result.Errors, importResult.Errors...)
	result.FinishedAt = time.Now().UTC()

	s.logger.Info("backup restored from S3",
		"snippets", importResult.SnippetsImported,
		"tags", importResult.TagsImported,
		"folders", importResult.FoldersImported,
		"duration", result.FinishedAt.Sub(result.StartedAt),
	)

	return result, nil
}

// DeleteBackup removes a backup from S3
func (s *S3SyncService) DeleteBackup(ctx context.Context, key string) error {
	if err := s.storage.Delete(ctx, key); err != nil {
		return fmt.Errorf("failed to delete backup: %w", err)
	}

	s.logger.Info("backup deleted from S3")
	return nil
}

// GetBackupURL generates a presigned URL for downloading a backup
func (s *S3SyncService) GetBackupURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	url, err := s.storage.GetPresignedURL(ctx, key, expiry)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return url, nil
}
