package services

import (
	"context"
	"log/slog"
	"time"

	"github.com/MohamedElashri/snipo/internal/repository"
)

// CleanupService handles background cleanup tasks
type CleanupService struct {
	snippetRepo *repository.SnippetRepository
	logger      *slog.Logger
}

// NewCleanupService creates a new cleanup service
func NewCleanupService(snippetRepo *repository.SnippetRepository, logger *slog.Logger) *CleanupService {
	return &CleanupService{
		snippetRepo: snippetRepo,
		logger:      logger,
	}
}

// Start starts the cleanup service periodic task
func (s *CleanupService) Start(ctx context.Context) {
	s.logger.Info("starting cleanup service")

	// Run immediately on startup
	if err := s.cleanup(ctx); err != nil {
		s.logger.Error("cleanup task failed", "error", err)
	}

	// Then run every 24 hours
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := s.cleanup(ctx); err != nil {
					s.logger.Error("cleanup task failed", "error", err)
				}
			}
		}
	}()
}

func (s *CleanupService) cleanup(ctx context.Context) error {
	s.logger.Info("running cleanup task")

	// Delete snippets deleted more than 30 days ago
	count, err := s.snippetRepo.CleanupDeleted(ctx, 30)
	if err != nil {
		return err
	}

	if count > 0 {
		s.logger.Info("cleaned up deleted snippets", "count", count)
	}

	return nil
}
