package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MohamedElashri/snipo/internal/api"
	"github.com/MohamedElashri/snipo/internal/api/middleware"
	"github.com/MohamedElashri/snipo/internal/auth"
	"github.com/MohamedElashri/snipo/internal/config"
	"github.com/MohamedElashri/snipo/internal/database"
	"github.com/MohamedElashri/snipo/internal/demo"
	"github.com/MohamedElashri/snipo/internal/repository"
	"github.com/MohamedElashri/snipo/internal/services"
)

// Build-time variables
var (
	Version = "dev"
	Commit  = "unknown"
)

func main() {
	// Check for subcommands
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "serve":
			runServer()
		case "migrate":
			runMigrations()
		case "version":
			fmt.Printf("snipo %s (commit: %s)\n", Version, Commit)
			os.Exit(0)
		case "health":
			checkHealth()
		case "hash-password":
			hashPassword()
		default:
			fmt.Printf("Unknown command: %s\n", os.Args[1])
			fmt.Println("Available commands: serve, migrate, version, health, hash-password")
			os.Exit(1)
		}
	} else {
		runServer()
	}
}

func runServer() {
	// Setup logger
	logger := setupLogger()

	logger.Info("starting snipo", "version", Version, "commit", Commit)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Configure proxy trust setting
	middleware.TrustProxy = cfg.Server.TrustProxy

	// Security warnings
	if cfg.Auth.Disabled {
		logger.Warn("⚠️  ⚠️  ⚠️  CRITICAL SECURITY WARNING ⚠️  ⚠️  ⚠️")
		logger.Warn("Authentication is COMPLETELY DISABLED (SNIPO_DISABLE_AUTH=true)")
		logger.Warn("ALL requests will be accepted WITHOUT any verification")
		logger.Warn("This should ONLY be used:",
			"use_case_1", "Behind a trusted authentication proxy (Authelia, Authentik, OAuth2 Proxy)",
			"use_case_2", "In a completely trusted local environment with no network access",
			"use_case_3", "For development/testing purposes only")
		logger.Warn("⚠️  NEVER expose this configuration directly to the internet ⚠️")
	} else if cfg.Auth.SessionSecretGenerated {
		logger.Warn("SECURITY WARNING: SNIPO_SESSION_SECRET not set - using auto-generated secret",
			"recommendation", "Set SNIPO_SESSION_SECRET environment variable for production. Generate with: openssl rand -hex 32")
	}

	if cfg.Auth.EncryptionSaltGenerated {
		logger.Warn("SECURITY WARNING: SNIPO_ENCRYPTION_SALT not set - using auto-generated salt",
			"recommendation", "Set SNIPO_ENCRYPTION_SALT environment variable for production. Generate with: openssl rand -hex 32",
			"impact", "GitHub sync tokens will not persist across restarts without a persistent encryption salt")
	}

	// Connect to database
	db, err := database.New(database.Config{
		Path:            cfg.Database.Path,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		BusyTimeout:     cfg.Database.BusyTimeout,
		JournalMode:     cfg.Database.JournalMode,
		SynchronousMode: cfg.Database.SynchronousMode,
	}, logger)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("failed to close database", "error", err)
		}
	}()

	// Run migrations
	ctx := context.Background()
	if err := db.Migrate(ctx); err != nil {
		logger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Create auth service
	// Use pre-hashed password if available, otherwise use plain password
	masterPasswordForAuth := cfg.Auth.MasterPasswordHash
	if masterPasswordForAuth == "" {
		masterPasswordForAuth = cfg.Auth.MasterPassword
	}

	authService := auth.NewService(
		db.DB,
		masterPasswordForAuth,
		cfg.Auth.SessionSecret,
		cfg.Auth.SessionDuration,
		logger,
		cfg.Auth.Disabled,
	)

	// Start session cleanup goroutine
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		for range ticker.C {
			if err := authService.CleanupExpiredSessions(); err != nil {
				logger.Warn("failed to cleanup sessions", "error", err)
			}
		}
	}()

	// Initialize gist sync worker
	var gistSyncWorker *services.GistSyncWorker
	gistSyncRepo := repository.NewGistSyncRepository(db.DB)
	snippetRepo := repository.NewSnippetRepository(db.DB)
	fileRepo := repository.NewSnippetFileRepository(db.DB)

	encryptionKey := services.DeriveEncryptionKey(cfg.Auth.EncryptionSalt)
	if encryptionSvc, err := services.NewEncryptionService(encryptionKey); err == nil {
		gistSyncWorker = services.NewGistSyncWorker(gistSyncRepo, snippetRepo, fileRepo, encryptionSvc, logger)
		if err := gistSyncWorker.Start(ctx); err != nil {
			logger.Warn("failed to start gist sync worker", "error", err)
		}
	}

	// Initialize demo mode if enabled
	if cfg.Demo.Enabled {
		// Create repositories and services for demo mode
		snippetRepo := repository.NewSnippetRepository(db.DB)
		tagRepo := repository.NewTagRepository(db.DB)
		folderRepo := repository.NewFolderRepository(db.DB)
		fileRepo := repository.NewSnippetFileRepository(db.DB)
		historyRepo := repository.NewHistoryRepository(db.DB)
		settingsRepo := repository.NewSettingsRepository(db.DB)

		snippetService := services.NewSnippetService(snippetRepo, logger).
			WithTagRepo(tagRepo).
			WithFolderRepo(folderRepo).
			WithFileRepo(fileRepo).
			WithHistoryRepo(historyRepo).
			WithSettingsRepo(settingsRepo).
			WithMaxFiles(cfg.Server.MaxFilesPerSnippet)

		demoService := demo.NewService(db.DB, snippetService, logger, cfg.Demo.ResetInterval, cfg.Demo.Enabled)
		demoService.StartPeriodicReset(ctx)
	}

	// Create router
	router := api.NewRouter(api.RouterConfig{
		DB:                 db.DB,
		Logger:             logger,
		AuthService:        authService,
		Config:             cfg, // Pass full config
		Version:            Version,
		Commit:             Commit,
		RateLimit:          cfg.Auth.RateLimit,
		RateLimitWindow:    int(cfg.Auth.RateLimitWindow.Seconds()),
		MaxFilesPerSnippet: cfg.Server.MaxFilesPerSnippet,
		S3Config:           &cfg.S3,
		BasePath:           cfg.Server.BasePath,
	})

	// Create server
	server := &http.Server{
		Addr:         cfg.Server.Addr(),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("server listening", "addr", cfg.Server.Addr())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	// Stop gist sync worker if running
	if gistSyncWorker != nil {
		if err := gistSyncWorker.Stop(); err != nil {
			logger.Warn("failed to stop gist sync worker", "error", err)
		}
	}

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
	}

	logger.Info("server stopped")
}

func runMigrations() {
	logger := setupLogger()

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	db, err := database.New(database.Config{
		Path:            cfg.Database.Path,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		BusyTimeout:     cfg.Database.BusyTimeout,
		JournalMode:     cfg.Database.JournalMode,
		SynchronousMode: cfg.Database.SynchronousMode,
	}, logger)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("failed to close database", "error", err)
		}
	}()

	ctx := context.Background()
	if err := db.Migrate(ctx); err != nil {
		logger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	logger.Info("migrations completed successfully")
}

func checkHealth() {
	// Simple health check for Docker HEALTHCHECK
	resp, err := http.Get("http://localhost:8080/ping")
	if err != nil {
		os.Exit(1)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		os.Exit(1)
	}
	os.Exit(0)
}

func hashPassword() {
	// Check if password is provided as argument
	var password string
	if len(os.Args) > 2 {
		password = os.Args[2]
	} else {
		// Prompt for password
		fmt.Print("Enter password to hash: ")
		if _, err := fmt.Scanln(&password); err != nil {
			fmt.Printf("Error reading password: %v\n", err)
			os.Exit(1)
		}
	}

	if password == "" {
		fmt.Println("Error: Password cannot be empty")
		os.Exit(1)
	}

	// Generate hash using auth package
	hash, err := auth.HashPassword(password)
	if err != nil {
		fmt.Printf("Error hashing password: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nGenerated Argon2id password hash:")
	fmt.Println(hash)
	fmt.Println("\nAdd this to your environment or .env file:")
	fmt.Printf("SNIPO_MASTER_PASSWORD_HASH=%s\n", hash)
	fmt.Println("\nNote: Remove SNIPO_MASTER_PASSWORD if you're using SNIPO_MASTER_PASSWORD_HASH")
}

func setupLogger() *slog.Logger {
	logLevel := os.Getenv("SNIPO_LOG_LEVEL")
	logFormat := os.Getenv("SNIPO_LOG_FORMAT")

	var level slog.Level
	switch logLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	if logFormat == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
