package api

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/MohamedElashri/snipo/internal/api/handlers"
	"github.com/MohamedElashri/snipo/internal/api/middleware"
	"github.com/MohamedElashri/snipo/internal/auth"
	"github.com/MohamedElashri/snipo/internal/config"
	"github.com/MohamedElashri/snipo/internal/repository"
	"github.com/MohamedElashri/snipo/internal/services"
	"github.com/MohamedElashri/snipo/internal/storage"
	"github.com/MohamedElashri/snipo/internal/web"
)

// RouterConfig holds router configuration
type RouterConfig struct {
	DB                 *sql.DB
	Logger             *slog.Logger
	AuthService        *auth.Service
	Version            string
	Commit             string
	RateLimit          int
	RateLimitWindow    int // in seconds
	MaxFilesPerSnippet int
	S3Config           *config.S3Config
}

// NewRouter creates and configures the HTTP router
func NewRouter(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.Recovery(cfg.Logger))
	r.Use(middleware.Logger(cfg.Logger))
	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.CORS)

	// Rate limiting for auth endpoints
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimit, 60*1000*1000*1000) // 1 minute in nanoseconds (nothing planck scale)

	// Create repositories
	snippetRepo := repository.NewSnippetRepository(cfg.DB)
	tagRepo := repository.NewTagRepository(cfg.DB)
	folderRepo := repository.NewFolderRepository(cfg.DB)
	tokenRepo := repository.NewTokenRepository(cfg.DB)
	fileRepo := repository.NewSnippetFileRepository(cfg.DB)
	settingsRepo := repository.NewSettingsRepository(cfg.DB)
	historyRepo := repository.NewHistoryRepository(cfg.DB)

	// Create services
	snippetService := services.NewSnippetService(snippetRepo, cfg.Logger).
		WithTagRepo(tagRepo).
		WithFolderRepo(folderRepo).
		WithFileRepo(fileRepo).
		WithHistoryRepo(historyRepo).
		WithSettingsRepo(settingsRepo).
		WithMaxFiles(cfg.MaxFilesPerSnippet)

	// Create backup service
	backupService := services.NewBackupService(cfg.DB, snippetService, tagRepo, folderRepo, fileRepo, cfg.Logger)

	// Create S3 sync service if configured
	var s3SyncService *services.S3SyncService
	if cfg.S3Config != nil && cfg.S3Config.Enabled {
		s3Storage, err := storage.NewS3Storage(storage.S3Config{
			Endpoint:        cfg.S3Config.Endpoint,
			AccessKeyID:     cfg.S3Config.AccessKeyID,
			SecretAccessKey: cfg.S3Config.SecretAccessKey,
			Bucket:          cfg.S3Config.Bucket,
			Region:          cfg.S3Config.Region,
			UseSSL:          cfg.S3Config.UseSSL,
		})
		if err != nil {
			cfg.Logger.Warn("failed to initialize S3 storage", "error", err)
		} else {
			s3SyncService = services.NewS3SyncService(s3Storage, backupService, cfg.Logger)
			cfg.Logger.Info("S3 storage initialized", "bucket", cfg.S3Config.Bucket)
		}
	}

	// Create handlers
	snippetHandler := handlers.NewSnippetHandler(snippetService)
	tagHandler := handlers.NewTagHandler(tagRepo)
	folderHandler := handlers.NewFolderHandler(folderRepo)
	tokenHandler := handlers.NewTokenHandler(tokenRepo)
	authHandler := handlers.NewAuthHandler(cfg.AuthService)
	healthHandler := handlers.NewHealthHandler(cfg.DB, cfg.Version, cfg.Commit)
	backupHandler := handlers.NewBackupHandler(backupService, s3SyncService)
	settingsHandler := handlers.NewSettingsHandler(settingsRepo)

	// Public routes (no auth required)
	r.Group(func(r chi.Router) {
		// Health checks
		r.Get("/health", healthHandler.Health)
		r.Get("/ping", healthHandler.Ping)

		// Public snippet access
		r.Get("/api/v1/snippets/public/{id}", snippetHandler.GetPublic)

		// Auth endpoints (with rate limiting)
		r.Group(func(r chi.Router) {
			r.Use(rateLimiter.Middleware)
			r.Post("/api/v1/auth/login", authHandler.Login)
		})

		r.Post("/api/v1/auth/logout", authHandler.Logout)
		r.Get("/api/v1/auth/check", authHandler.Check)
	})

	// Protected routes (auth required)
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuthWithTokenRepo(cfg.AuthService, tokenRepo))

		// Auth management (protected)
		r.Post("/api/v1/auth/change-password", authHandler.ChangePassword)

		// Settings management
		r.Route("/api/v1/settings", func(r chi.Router) {
			r.Get("/", settingsHandler.Get)
			r.Put("/", settingsHandler.Update)
		})

		// Snippet CRUD
		r.Route("/api/v1/snippets", func(r chi.Router) {
			r.Get("/", snippetHandler.List)
			r.Post("/", snippetHandler.Create)
			r.Get("/search", snippetHandler.Search)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", snippetHandler.Get)
				r.Put("/", snippetHandler.Update)
				r.Delete("/", snippetHandler.Delete)
				r.Post("/favorite", snippetHandler.ToggleFavorite)
				r.Post("/archive", snippetHandler.ToggleArchive)
				r.Post("/duplicate", snippetHandler.Duplicate)
				
				// History routes
				r.Get("/history", snippetHandler.GetHistory)
				r.Post("/history/{history_id}/restore", snippetHandler.RestoreFromHistory)
			})
		})

		// Tag CRUD
		r.Route("/api/v1/tags", func(r chi.Router) {
			r.Get("/", tagHandler.List)
			r.Post("/", tagHandler.Create)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", tagHandler.Get)
				r.Put("/", tagHandler.Update)
				r.Delete("/", tagHandler.Delete)
			})
		})

		// Folder CRUD
		r.Route("/api/v1/folders", func(r chi.Router) {
			r.Get("/", folderHandler.List)
			r.Post("/", folderHandler.Create)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", folderHandler.Get)
				r.Put("/", folderHandler.Update)
				r.Delete("/", folderHandler.Delete)
				r.Put("/move", folderHandler.Move)
			})
		})

		// API Token management
		r.Route("/api/v1/tokens", func(r chi.Router) {
			r.Get("/", tokenHandler.List)
			r.Post("/", tokenHandler.Create)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", tokenHandler.Get)
				r.Delete("/", tokenHandler.Delete)
			})
		})

		// Backup & Restore
		r.Route("/api/v1/backup", func(r chi.Router) {
			r.Get("/export", backupHandler.Export)
			r.Post("/import", backupHandler.Import)

			// S3 operations
			r.Get("/s3/status", backupHandler.S3Status)
			r.Post("/s3/sync", backupHandler.S3Sync)
			r.Get("/s3/list", backupHandler.S3List)
			r.Post("/s3/restore", backupHandler.S3Restore)
			r.Delete("/s3/delete", backupHandler.S3Delete)
		})
	})

	// Web UI routes
	webHandler, err := web.NewHandler(cfg.AuthService)
	if err != nil {
		cfg.Logger.Error("failed to create web handler", "error", err)
	} else {
		// Static files
		r.Handle("/static/*", web.StaticHandler())

		// Web pages
		r.Get("/", webHandler.Index)
		r.Get("/login", webHandler.Login)
		r.Get("/s/{id}", webHandler.PublicSnippet) // Public snippet share page
	}

	return r
}
