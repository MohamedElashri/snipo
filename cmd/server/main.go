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
		default:
			fmt.Printf("Unknown command: %s\n", os.Args[1])
			fmt.Println("Available commands: serve, migrate, version, health")
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

	// Warn if session secret was auto-generated
	if cfg.Auth.SessionSecretGenerated {
		logger.Warn("SECURITY WARNING: SNIPO_SESSION_SECRET not set - using auto-generated secret",
			"recommendation", "Set SNIPO_SESSION_SECRET environment variable for production. Generate with: openssl rand -hex 32")
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
	defer db.Close()

	// Run migrations
	ctx := context.Background()
	if err := db.Migrate(ctx); err != nil {
		logger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Create auth service
	authService := auth.NewService(
		db.DB,
		cfg.Auth.MasterPassword,
		cfg.Auth.SessionSecret,
		cfg.Auth.SessionDuration,
		logger,
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

	// Create router
	router := api.NewRouter(api.RouterConfig{
		DB:                 db.DB,
		Logger:             logger,
		AuthService:        authService,
		Version:            Version,
		Commit:             Commit,
		RateLimit:          cfg.Auth.RateLimit,
		RateLimitWindow:    int(cfg.Auth.RateLimitWindow.Seconds()),
		MaxFilesPerSnippet: cfg.Server.MaxFilesPerSnippet,
		S3Config:           &cfg.S3,
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
	defer db.Close()

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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		os.Exit(1)
	}
	os.Exit(0)
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
