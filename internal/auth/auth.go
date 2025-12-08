package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
)

// Common errors
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrSessionExpired     = errors.New("session expired")
	ErrInvalidToken       = errors.New("invalid token")
)

// Argon2id parameters (OWASP recommended)
const (
	argonTime    = 1
	argonMemory  = 64 * 1024
	argonThreads = 4
	argonKeyLen  = 32
)

// Config holds authentication configuration
type Config struct {
	MasterPasswordHash string
	SessionSecret      string
	SessionDuration    time.Duration
}

// Service handles authentication
type Service struct {
	db              *sql.DB
	masterPassword  string
	sessionSecret   string
	sessionDuration time.Duration
	logger          *slog.Logger
}

// NewService creates a new authentication service
func NewService(db *sql.DB, masterPassword, sessionSecret string, sessionDuration time.Duration, logger *slog.Logger) *Service {
	return &Service{
		db:              db,
		masterPassword:  masterPassword,
		sessionSecret:   sessionSecret,
		sessionDuration: sessionDuration,
		logger:          logger,
	}
}

// VerifyPassword checks if the provided password matches the master password
func (s *Service) VerifyPassword(password string) bool {
	return subtle.ConstantTimeCompare([]byte(password), []byte(s.masterPassword)) == 1
}

// UpdatePassword updates the master password (in-memory only, resets on restart)
// For persistent password storage, this would need to be stored in the database
func (s *Service) UpdatePassword(newPassword string) error {
	s.masterPassword = newPassword
	s.logger.Info("master password updated")
	return nil
}

// CreateSession creates a new session and returns the session token
func (s *Service) CreateSession() (string, error) {
	// Generate random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate session token: %w", err)
	}
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	// Hash token for storage
	tokenHash := hashToken(token)

	// Generate session ID
	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		return "", fmt.Errorf("failed to generate session ID: %w", err)
	}
	sessionID := hex.EncodeToString(idBytes)

	// Calculate expiry
	expiresAt := time.Now().Add(s.sessionDuration)

	// Store session
	_, err := s.db.Exec(
		"INSERT INTO sessions (id, token_hash, expires_at) VALUES (?, ?, ?)",
		sessionID, tokenHash, expiresAt,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	s.logger.Info("session created", "session_id", sessionID, "expires_at", expiresAt)
	return token, nil
}

// ValidateSession checks if a session token is valid
func (s *Service) ValidateSession(token string) bool {
	if token == "" {
		return false
	}

	tokenHash := hashToken(token)

	var expiresAt time.Time
	err := s.db.QueryRow(
		"SELECT expires_at FROM sessions WHERE token_hash = ?",
		tokenHash,
	).Scan(&expiresAt)

	if err != nil {
		return false
	}

	if time.Now().After(expiresAt) {
		// Clean up expired session
		_, _ = s.db.Exec("DELETE FROM sessions WHERE token_hash = ?", tokenHash)
		return false
	}

	return true
}

// InvalidateSession removes a session
func (s *Service) InvalidateSession(token string) error {
	tokenHash := hashToken(token)
	_, err := s.db.Exec("DELETE FROM sessions WHERE token_hash = ?", tokenHash)
	return err
}

// CleanupExpiredSessions removes all expired sessions
func (s *Service) CleanupExpiredSessions() error {
	result, err := s.db.Exec("DELETE FROM sessions WHERE expires_at < ?", time.Now())
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows > 0 {
		s.logger.Info("cleaned up expired sessions", "count", rows)
	}
	return nil
}

// SetSessionCookie sets the session cookie on the response
func (s *Service) SetSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "snipo_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(s.sessionDuration.Seconds()),
	})
}

// ClearSessionCookie clears the session cookie
func (s *Service) ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "snipo_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}

// GetSessionFromRequest extracts the session token from the request
func GetSessionFromRequest(r *http.Request) string {
	// Check cookie first
	cookie, err := r.Cookie("snipo_session")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// Check Authorization header
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	// Check X-API-Key header
	apiKey := r.Header.Get("X-API-Key")
	if apiKey != "" {
		return apiKey
	}

	return ""
}

// hashToken creates a SHA256 hash of the token
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// HashPassword creates an Argon2id hash of a password
func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	// Encode as: $argon2id$salt$hash
	return fmt.Sprintf("$argon2id$%s$%s",
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

// VerifyPasswordHash checks password against an Argon2id hash
func VerifyPasswordHash(password, encodedHash string) bool {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 4 || parts[1] != "argon2id" {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return false
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false
	}

	computedHash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	return subtle.ConstantTimeCompare(hash, computedHash) == 1
}

// GenerateAPIToken creates a secure random API token
func GenerateAPIToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
