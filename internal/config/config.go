package config

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	S3       S3Config
	Logging  LoggingConfig
	API      APIConfig
	Features FeatureFlags
}

// ServerConfig holds HTTP server settings
type ServerConfig struct {
	Host               string
	Port               int
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	TrustProxy         bool
	MaxFilesPerSnippet int
}

// DatabaseConfig holds SQLite settings
type DatabaseConfig struct {
	Path            string
	MaxOpenConns    int
	BusyTimeout     int
	JournalMode     string
	SynchronousMode string
}

// AuthConfig holds authentication settings
type AuthConfig struct {
	MasterPassword         string
	MasterPasswordHash     string // Pre-hashed password (Argon2id format)
	Disabled               bool   // Disable authentication entirely (use with external auth like Authelia)
	SessionSecret          string
	SessionSecretGenerated bool // True if session secret was auto-generated (not recommended for production)
	SessionDuration        time.Duration
	RateLimit              int
	RateLimitWindow        time.Duration
}

// S3Config holds S3 storage settings
type S3Config struct {
	Enabled         bool
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	Region          string
	UseSSL          bool
}

// LoggingConfig holds logging settings
type LoggingConfig struct {
	Level  string
	Format string
}

// APIConfig holds API-specific settings
type APIConfig struct {
	AllowedOrigins []string // CORS allowed origins
	RateLimitRead  int      // requests per hour for read operations
	RateLimitWrite int      // requests per hour for write operations
	RateLimitAdmin int      // requests per hour for admin operations
}

// FeatureFlags holds feature toggle settings
type FeatureFlags struct {
	PublicSnippets bool
	S3Sync         bool
	APITokens      bool
	BackupRestore  bool
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{}

	// Server
	cfg.Server.Host = getEnv("SNIPO_HOST", "0.0.0.0")
	cfg.Server.Port = getEnvInt("SNIPO_PORT", 8080)
	cfg.Server.ReadTimeout = getEnvDuration("SNIPO_READ_TIMEOUT", 30*time.Second)
	cfg.Server.WriteTimeout = getEnvDuration("SNIPO_WRITE_TIMEOUT", 30*time.Second)
	cfg.Server.TrustProxy = getEnvBool("SNIPO_TRUST_PROXY", false)
	cfg.Server.MaxFilesPerSnippet = getEnvInt("SNIPO_MAX_FILES_PER_SNIPPET", 10)

	// Database
	cfg.Database.Path = getEnv("SNIPO_DB_PATH", "./data/snipo.db")
	cfg.Database.MaxOpenConns = getEnvInt("SNIPO_DB_MAX_CONNS", 1)
	cfg.Database.BusyTimeout = getEnvInt("SNIPO_DB_BUSY_TIMEOUT", 5000)
	cfg.Database.JournalMode = getEnv("SNIPO_DB_JOURNAL", "WAL")
	cfg.Database.SynchronousMode = getEnv("SNIPO_DB_SYNC", "NORMAL")

	// Auth - Check if authentication is disabled
	cfg.Auth.Disabled = getEnvBool("SNIPO_DISABLE_AUTH", false)
	
	// If auth is disabled, skip password requirements
	if cfg.Auth.Disabled {
		// Clear any password config when disabled
		cfg.Auth.MasterPassword = ""
		cfg.Auth.MasterPasswordHash = ""
	} else {
		// Auth enabled - Support both plain text password and pre-hashed password
		cfg.Auth.MasterPassword = os.Getenv("SNIPO_MASTER_PASSWORD")
		cfg.Auth.MasterPasswordHash = os.Getenv("SNIPO_MASTER_PASSWORD_HASH")
		
		// At least one password method must be provided when auth is enabled
		if cfg.Auth.MasterPassword == "" && cfg.Auth.MasterPasswordHash == "" {
			return nil, errors.New("either SNIPO_MASTER_PASSWORD or SNIPO_MASTER_PASSWORD_HASH is required (or set SNIPO_DISABLE_AUTH=true)")
		}
		
		// If both are provided, prefer the hash
		if cfg.Auth.MasterPasswordHash != "" {
			// Validate hash format (basic check for Argon2id format)
			if !strings.HasPrefix(cfg.Auth.MasterPasswordHash, "$argon2id$") {
				return nil, errors.New("SNIPO_MASTER_PASSWORD_HASH must be in Argon2id format ($argon2id$...)")
			}
			cfg.Auth.MasterPassword = "" // Clear plain text if hash is provided
		}
	}

	sessionSecret := os.Getenv("SNIPO_SESSION_SECRET")
	if sessionSecret == "" {
		secret, err := generateSecret()
		if err != nil {
			return nil, err
		}
		sessionSecret = secret
		cfg.Auth.SessionSecretGenerated = true
	}
	cfg.Auth.SessionSecret = sessionSecret
	cfg.Auth.SessionDuration = getEnvDuration("SNIPO_SESSION_DURATION", 168*time.Hour)
	cfg.Auth.RateLimit = getEnvInt("SNIPO_RATE_LIMIT", 100)
	cfg.Auth.RateLimitWindow = getEnvDuration("SNIPO_RATE_WINDOW", 1*time.Minute)

	// S3
	cfg.S3.Enabled = getEnvBool("SNIPO_S3_ENABLED", false)
	cfg.S3.Endpoint = os.Getenv("SNIPO_S3_ENDPOINT")
	cfg.S3.AccessKeyID = os.Getenv("SNIPO_S3_ACCESS_KEY")
	cfg.S3.SecretAccessKey = os.Getenv("SNIPO_S3_SECRET_KEY")
	cfg.S3.Bucket = os.Getenv("SNIPO_S3_BUCKET")
	cfg.S3.Region = getEnv("SNIPO_S3_REGION", "us-east-1")
	cfg.S3.UseSSL = getEnvBool("SNIPO_S3_SSL", true)

	// Logging
	cfg.Logging.Level = getEnv("SNIPO_LOG_LEVEL", "info")
	cfg.Logging.Format = getEnv("SNIPO_LOG_FORMAT", "json")

	// API
	originsStr := getEnv("SNIPO_ALLOWED_ORIGINS", "*")
	if originsStr == "*" {
		cfg.API.AllowedOrigins = []string{"*"}
	} else {
		cfg.API.AllowedOrigins = strings.Split(originsStr, ",")
		// Trim whitespace from each origin
		for i, origin := range cfg.API.AllowedOrigins {
			cfg.API.AllowedOrigins[i] = strings.TrimSpace(origin)
		}
	}
	cfg.API.RateLimitRead = getEnvInt("SNIPO_RATE_LIMIT_READ", 1000)
	cfg.API.RateLimitWrite = getEnvInt("SNIPO_RATE_LIMIT_WRITE", 500)
	cfg.API.RateLimitAdmin = getEnvInt("SNIPO_RATE_LIMIT_ADMIN", 100)

	// Feature Flags
	cfg.Features.PublicSnippets = getEnvBool("SNIPO_ENABLE_PUBLIC_SNIPPETS", true)
	cfg.Features.S3Sync = cfg.S3.Enabled // S3Sync follows S3.Enabled
	cfg.Features.APITokens = getEnvBool("SNIPO_ENABLE_API_TOKENS", true)
	cfg.Features.BackupRestore = getEnvBool("SNIPO_ENABLE_BACKUP_RESTORE", true)

	return cfg, nil
}

// Addr returns the server address string
func (c *ServerConfig) Addr() string {
	return c.Host + ":" + strconv.Itoa(c.Port)
}

// Helper functions

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		return val == "true" || val == "1" || val == "yes"
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultVal
}

func generateSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
