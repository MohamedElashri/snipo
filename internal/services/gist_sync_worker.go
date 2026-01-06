package services

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/MohamedElashri/snipo/internal/repository"
)

// GistSyncWorker handles background synchronization
type GistSyncWorker struct {
	syncRepo      *repository.GistSyncRepository
	snippetRepo   *repository.SnippetRepository
	encryptionSvc *EncryptionService
	logger        *slog.Logger
	stopCh        chan struct{}
	wg            sync.WaitGroup
	mu            sync.Mutex
	running       bool
}

// NewGistSyncWorker creates a new background sync worker
func NewGistSyncWorker(
	syncRepo *repository.GistSyncRepository,
	snippetRepo *repository.SnippetRepository,
	encryptionSvc *EncryptionService,
	logger *slog.Logger,
) *GistSyncWorker {
	return &GistSyncWorker{
		syncRepo:      syncRepo,
		snippetRepo:   snippetRepo,
		encryptionSvc: encryptionSvc,
		logger:        logger,
		stopCh:        make(chan struct{}),
	}
}

// Start begins the background sync worker
func (w *GistSyncWorker) Start(ctx context.Context) error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return nil
	}
	w.running = true
	w.mu.Unlock()

	w.wg.Add(1)
	go w.run(ctx)

	w.logger.Info("gist sync worker started")
	return nil
}

// Stop gracefully stops the background sync worker
func (w *GistSyncWorker) Stop() error {
	w.mu.Lock()
	if !w.running {
		w.mu.Unlock()
		return nil
	}
	w.mu.Unlock()

	close(w.stopCh)
	w.wg.Wait()

	w.mu.Lock()
	w.running = false
	w.mu.Unlock()

	w.logger.Info("gist sync worker stopped")
	return nil
}

// run is the main worker loop
func (w *GistSyncWorker) run(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.performSync(ctx)
		}
	}
}

// performSync executes a sync cycle
func (w *GistSyncWorker) performSync(ctx context.Context) {
	config, err := w.syncRepo.GetConfig(ctx)
	if err != nil {
		w.logger.Error("failed to get sync config", "error", err)
		return
	}

	if config == nil || !config.Enabled || !config.AutoSyncEnabled {
		return
	}

	// Check if token exists
	if config.GithubTokenEncrypted == "" {
		w.logger.Debug("no github token configured, skipping sync")
		return
	}

	if config.LastFullSyncAt != nil {
		nextSync := config.LastFullSyncAt.Add(time.Duration(config.SyncIntervalMinutes) * time.Minute)
		if time.Now().Before(nextSync) {
			return
		}
	}

	w.logger.Info("starting automatic sync")

	token, err := w.encryptionSvc.Decrypt(config.GithubTokenEncrypted)
	if err != nil {
		w.logger.Error("failed to decrypt token", "error", err, "token_length", len(config.GithubTokenEncrypted))
		return
	}

	githubClient := NewGitHubClient(token)
	syncService := NewGistSyncService(githubClient, w.snippetRepo, w.syncRepo, w.encryptionSvc)

	result, err := syncService.SyncAll(ctx)
	if err != nil {
		w.logger.Error("sync failed", "error", err)
		return
	}

	w.logger.Info("automatic sync completed",
		"total", result.TotalProcessed,
		"synced", result.Synced,
		"conflicts", result.Conflicts,
		"errors", result.Errors,
		"duration", result.Duration,
	)
}

// IsRunning returns whether the worker is currently running
func (w *GistSyncWorker) IsRunning() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.running
}
