package api

import (
	"database/sql"
	"log/slog"
	"net/http"
	"time"

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
	Config             *config.Config // Full application config
	Version            string
	Commit             string
	RateLimit          int
	RateLimitWindow    int // in seconds
	MaxFilesPerSnippet int
	S3Config           *config.S3Config
	SnippetService     *services.SnippetService // For demo mode
	BasePath           string                   // Base path for reverse proxy
}

// NewRouter creates and configures the HTTP router
func NewRouter(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	// Global middleware (order matters!)
	r.Use(middleware.RequestID)            // Generate request IDs first
	r.Use(middleware.Recovery(cfg.Logger)) // Catch panics
	r.Use(middleware.Logger(cfg.Logger))   // Log requests (includes request ID)
	r.Use(middleware.SecurityHeaders)      // Security headers (includes X-API-Version)

	// Use configured CORS
	allowedOrigins := []string{"*"} // default
	if cfg.Config != nil {
		allowedOrigins = cfg.Config.API.AllowedOrigins
	}
	r.Use(middleware.CORS(allowedOrigins)) // CORS handling

	// Rate limiting for auth endpoints
	authRateLimiter := middleware.NewRateLimiter(cfg.RateLimit, 60*1000*1000*1000) // 1 minute in nanoseconds

	// API rate limiter with permission-based limits (use config values or defaults)
	readLimit, writeLimit, adminLimit := 1000, 500, 100
	if cfg.Config != nil {
		readLimit = cfg.Config.API.RateLimitRead
		writeLimit = cfg.Config.API.RateLimitWrite
		adminLimit = cfg.Config.API.RateLimitAdmin
	}
	apiRateLimiter := middleware.NewAPIRateLimiter(middleware.RateLimitConfig{
		ReadLimit:  readLimit,
		WriteLimit: writeLimit,
		AdminLimit: adminLimit,
		Window:     time.Hour,
	})

	// Create repositories
	snippetRepo := repository.NewSnippetRepository(cfg.DB)
	tagRepo := repository.NewTagRepository(cfg.DB)
	folderRepo := repository.NewFolderRepository(cfg.DB)
	tokenRepo := repository.NewTokenRepository(cfg.DB)
	fileRepo := repository.NewSnippetFileRepository(cfg.DB)
	settingsRepo := repository.NewSettingsRepository(cfg.DB)
	historyRepo := repository.NewHistoryRepository(cfg.DB)
	gistSyncRepo := repository.NewGistSyncRepository(cfg.DB)

	// Create services
	var snippetService *services.SnippetService
	if cfg.SnippetService != nil {
		// Use provided snippet service (for demo mode)
		snippetService = cfg.SnippetService
	} else {
		// Create new snippet service
		snippetService = services.NewSnippetService(snippetRepo, cfg.Logger).
			WithTagRepo(tagRepo).
			WithFolderRepo(folderRepo).
			WithFileRepo(fileRepo).
			WithHistoryRepo(historyRepo).
			WithSettingsRepo(settingsRepo).
			WithMaxFiles(cfg.MaxFilesPerSnippet)
	}

	// Create backup service
	backupService := services.NewBackupService(cfg.DB, snippetService, tagRepo, folderRepo, fileRepo, cfg.Logger, cfg.Config.Auth.EncryptionSalt)

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
	tokenHandler := handlers.NewTokenHandler(tokenRepo, settingsRepo, cfg.AuthService).WithDemoMode(cfg.Config.Demo.Enabled)
	authHandler := handlers.NewAuthHandler(cfg.AuthService).WithDemoMode(cfg.Config.Demo.Enabled)

	// Create health handler with feature flags
	var featureFlags *config.FeatureFlags
	if cfg.Config != nil {
		featureFlags = &cfg.Config.Features
	}
	healthHandler := handlers.NewHealthHandler(cfg.DB, cfg.Version, cfg.Commit, featureFlags)

	backupHandler := handlers.NewBackupHandler(backupService, s3SyncService)
	settingsHandler := handlers.NewSettingsHandler(settingsRepo, cfg.AuthService)

	// Create encryption service for gist sync (using encryption salt as key for persistence)
	encryptionKey := services.DeriveEncryptionKey(cfg.Config.Auth.EncryptionSalt)
	encryptionSvc, err := services.NewEncryptionService(encryptionKey)
	if err != nil {
		cfg.Logger.Warn("failed to initialize encryption service", "error", err)
	}

	// Create gist sync handler
	var gistSyncHandler *handlers.GistSyncHandler
	if encryptionSvc != nil {
		gistSyncHandler = handlers.NewGistSyncHandler(gistSyncRepo, snippetRepo, fileRepo, encryptionSvc)
	}

	// Public routes (no auth required)
	r.Group(func(r chi.Router) {
		// Health checks
		r.Get("/health", healthHandler.Health)
		r.Get("/ping", healthHandler.Ping)

		// OpenAPI specification
		r.Get("/api/v1/openapi.json", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "docs/openapi.yaml")
		})

		// Public snippet access
		r.Get("/api/v1/snippets/public/{id}", snippetHandler.GetPublic)
		r.Get("/api/v1/snippets/public/{id}/files/{filename}", snippetHandler.GetPublicFile)

		// Auth endpoints (with rate limiting)
		r.Group(func(r chi.Router) {
			r.Use(authRateLimiter.Middleware)
			r.Post("/api/v1/auth/login", authHandler.Login)
		})

		r.Post("/api/v1/auth/logout", authHandler.Logout)
		r.Get("/api/v1/auth/check", authHandler.Check)
	})

	// Protected routes (auth required + rate limiting)
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuthWithSettings(cfg.AuthService, tokenRepo, settingsRepo))

		// Auth management (protected, requires any auth)

		// Settings management (admin only)
		r.Route("/api/v1/settings", func(r chi.Router) {
			r.Use(middleware.RequireAdmin)
			r.Use(apiRateLimiter.RateLimitAdmin)
			r.Get("/", settingsHandler.Get)
			r.Put("/", settingsHandler.Update)
		})

		// Snippet CRUD (read for GET, write for modifications)
		r.Route("/api/v1/snippets", func(r chi.Router) {
			r.With(middleware.RequireRead, apiRateLimiter.RateLimitRead).Get("/", snippetHandler.List)
			r.With(middleware.RequireWrite, apiRateLimiter.RateLimitWrite).Post("/", snippetHandler.Create)
			r.With(middleware.RequireRead, apiRateLimiter.RateLimitRead).Get("/search", snippetHandler.Search)

			r.Route("/{id}", func(r chi.Router) {
				r.With(middleware.RequireRead, apiRateLimiter.RateLimitRead).Get("/", snippetHandler.Get)
				r.With(middleware.RequireWrite, apiRateLimiter.RateLimitWrite).Put("/", snippetHandler.Update)
				r.With(middleware.RequireWrite, apiRateLimiter.RateLimitWrite).Delete("/", snippetHandler.Delete)
				r.With(middleware.RequireWrite, apiRateLimiter.RateLimitWrite).Post("/favorite", snippetHandler.ToggleFavorite)
				r.With(middleware.RequireWrite, apiRateLimiter.RateLimitWrite).Post("/archive", snippetHandler.ToggleArchive)
				r.With(middleware.RequireWrite, apiRateLimiter.RateLimitWrite).Post("/duplicate", snippetHandler.Duplicate)
				r.With(middleware.RequireWrite, apiRateLimiter.RateLimitWrite).Post("/restore", snippetHandler.Restore)

				// History routes
				r.With(middleware.RequireRead, apiRateLimiter.RateLimitRead).Get("/history", snippetHandler.GetHistory)
				r.With(middleware.RequireWrite, apiRateLimiter.RateLimitWrite).Post("/history/{history_id}/restore", snippetHandler.RestoreFromHistory)
			})
		})

		// Tag CRUD (read for GET, write for modifications)
		r.Route("/api/v1/tags", func(r chi.Router) {
			r.With(middleware.RequireRead, apiRateLimiter.RateLimitRead).Get("/", tagHandler.List)
			r.With(middleware.RequireWrite, apiRateLimiter.RateLimitWrite).Post("/", tagHandler.Create)

			r.Route("/{id}", func(r chi.Router) {
				r.With(middleware.RequireRead, apiRateLimiter.RateLimitRead).Get("/", tagHandler.Get)
				r.With(middleware.RequireWrite, apiRateLimiter.RateLimitWrite).Put("/", tagHandler.Update)
				r.With(middleware.RequireWrite, apiRateLimiter.RateLimitWrite).Delete("/", tagHandler.Delete)
			})
		})

		// Folder CRUD (read for GET, write for modifications)
		r.Route("/api/v1/folders", func(r chi.Router) {
			r.With(middleware.RequireRead, apiRateLimiter.RateLimitRead).Get("/", folderHandler.List)
			r.With(middleware.RequireWrite, apiRateLimiter.RateLimitWrite).Post("/", folderHandler.Create)

			r.Route("/{id}", func(r chi.Router) {
				r.With(middleware.RequireRead, apiRateLimiter.RateLimitRead).Get("/", folderHandler.Get)
				r.With(middleware.RequireWrite, apiRateLimiter.RateLimitWrite).Put("/", folderHandler.Update)
				r.With(middleware.RequireWrite, apiRateLimiter.RateLimitWrite).Delete("/", folderHandler.Delete)
				r.With(middleware.RequireWrite, apiRateLimiter.RateLimitWrite).Put("/move", folderHandler.Move)
			})
		})

		// API Token management (admin only)
		r.Route("/api/v1/tokens", func(r chi.Router) {
			r.Use(middleware.RequireAdmin)
			r.Use(apiRateLimiter.RateLimitAdmin)
			r.Get("/", tokenHandler.List)
			r.Post("/", tokenHandler.Create)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", tokenHandler.Get)
				r.Delete("/", tokenHandler.Delete)
			})
		})

		// Backup & Restore (admin only)
		r.Route("/api/v1/backup", func(r chi.Router) {
			r.Use(middleware.RequireAdmin)
			r.Use(apiRateLimiter.RateLimitAdmin)
			r.Get("/export", backupHandler.Export)
			r.Post("/import", backupHandler.Import)

			// S3 operations
			r.Get("/s3/status", backupHandler.S3Status)
			r.Post("/s3/sync", backupHandler.S3Sync)
			r.Get("/s3/list", backupHandler.S3List)
			r.Post("/s3/restore", backupHandler.S3Restore)
			r.Delete("/s3/delete", backupHandler.S3Delete)
		})

		// GitHub Gist Sync (admin only for config, write for sync operations)
		if gistSyncHandler != nil {
			r.Route("/api/v1/gist", func(r chi.Router) {
				// Config endpoints (admin only)
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireAdmin)
					r.Use(apiRateLimiter.RateLimitAdmin)
					r.Get("/config", gistSyncHandler.GetConfig)
					r.Post("/config", gistSyncHandler.UpdateConfig)
					r.Delete("/config", gistSyncHandler.ClearConfig)
					r.Post("/config/test", gistSyncHandler.TestConnection)
				})

				// Sync operations (write permission)
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireWrite)
					r.Use(apiRateLimiter.RateLimitWrite)
					r.Post("/sync/snippet/{id}", gistSyncHandler.SyncSnippet)
					r.Post("/sync/all", gistSyncHandler.SyncAll)
					r.Post("/sync/enable/{id}", gistSyncHandler.EnableSync)
					r.Post("/sync/enable-all", gistSyncHandler.EnableSyncForAll)
					r.Post("/sync/disable/{id}", gistSyncHandler.DisableSync)
				})

				// Mappings and conflicts (read permission)
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireRead)
					r.Use(apiRateLimiter.RateLimitRead)
					r.Get("/mappings", gistSyncHandler.ListMappings)
					r.Get("/conflicts", gistSyncHandler.ListConflicts)
					r.Get("/logs", gistSyncHandler.GetLogs)
				})

				// Mapping deletion and conflict resolution (write permission)
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireWrite)
					r.Use(apiRateLimiter.RateLimitWrite)
					r.Delete("/mappings/{id}", gistSyncHandler.DeleteMapping)
					r.Post("/conflicts/{id}/resolve", gistSyncHandler.ResolveConflict)
				})
			})
		}
	})

	// Web UI routes
	webHandler, err := web.NewHandler(cfg.AuthService, settingsRepo)
	if err != nil {
		cfg.Logger.Error("failed to create web handler", "error", err)
	} else {
		// Set demo mode and base path if enabled
		webHandler = webHandler.WithDemoMode(cfg.Config.Demo.Enabled).WithBasePath(cfg.BasePath)

		// Static files
		r.Handle("/static/*", web.StaticHandler(cfg.BasePath))

		// Web pages
		r.Get("/", webHandler.Index)
		r.Get("/login", webHandler.Login)
		r.Get("/s/{id}", webHandler.PublicSnippet) // Public snippet share page
	}

	// If base path is configured, mount everything under it
	if cfg.BasePath != "" {
		baseRouter := chi.NewRouter()
		baseRouter.Mount(cfg.BasePath, r)
		return baseRouter
	}

	return r
}
