package auth

import (
	"crypto/hmac"
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
	"sync"
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
	db                 *sql.DB
	masterPasswordHash string
	sessionSecret      string
	sessionDuration    time.Duration
	logger             *slog.Logger
	failedAttempts     *FailedLoginTracker
	authDisabled       bool // If true, authentication is completely bypassed
}

// FailedLoginTracker tracks failed login attempts per IP for progressive delays
type FailedLoginTracker struct {
	attempts map[string]*loginAttempt
	mu       sync.RWMutex
}

type loginAttempt struct {
	count    int
	lastFail time.Time
}

// NewFailedLoginTracker creates a new tracker
func NewFailedLoginTracker() *FailedLoginTracker {
	tracker := &FailedLoginTracker{
		attempts: make(map[string]*loginAttempt),
	}
	// Start cleanup goroutine
	go tracker.cleanup()
	return tracker
}

// RecordFailure records a failed login attempt
func (t *FailedLoginTracker) RecordFailure(ip string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.attempts[ip] == nil {
		t.attempts[ip] = &loginAttempt{}
	}
	t.attempts[ip].count++
	t.attempts[ip].lastFail = time.Now()
}

// RecordSuccess clears failed attempts for an IP
func (t *FailedLoginTracker) RecordSuccess(ip string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.attempts, ip)
}

// GetDelay returns the delay duration before next login attempt is allowed
// Progressive delays: 0s, 1s, 2s, 4s, 8s, 16s, 30s (max)
func (t *FailedLoginTracker) GetDelay(ip string) time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()

	attempt := t.attempts[ip]
	if attempt == nil || attempt.count == 0 {
		return 0
	}

	// Calculate delay: 2^(attempts-1) seconds, max 30 seconds
	delaySeconds := 1 << (attempt.count - 1) // 1, 2, 4, 8, 16, 32...
	if delaySeconds > 30 {
		delaySeconds = 30
	}

	elapsed := time.Since(attempt.lastFail)
	requiredDelay := time.Duration(delaySeconds) * time.Second

	if elapsed >= requiredDelay {
		return 0
	}
	return requiredDelay - elapsed
}

// cleanup removes old entries periodically
func (t *FailedLoginTracker) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	for range ticker.C {
		t.mu.Lock()
		now := time.Now()
		for ip, attempt := range t.attempts {
			// Remove entries older than 1 hour
			if now.Sub(attempt.lastFail) > time.Hour {
				delete(t.attempts, ip)
			}
		}
		t.mu.Unlock()
	}
}

// NewService creates a new authentication service
// The master password is hashed at startup using Argon2id for secure storage in memory
// If a pre-hashed password is provided, it's used directly without re-hashing
// If authDisabled is true, authentication is completely bypassed (use with external auth)
func NewService(db *sql.DB, masterPassword, sessionSecret string, sessionDuration time.Duration, logger *slog.Logger, authDisabled bool) *Service {
	var passwordHash string

	// If auth is disabled, skip all password processing
	if authDisabled {
		logger.Warn("⚠️  AUTHENTICATION DISABLED - All requests will be accepted without verification",
			"security_risk", "high",
			"recommendation", "Only use this behind a trusted authentication proxy (e.g., Authelia, Authentik) or in secure local environments")
	} else {
		// Check if password is already hashed (Argon2id format)
		if strings.HasPrefix(masterPassword, "$argon2id$") {
			passwordHash = masterPassword
			logger.Info("using pre-hashed master password (Argon2id)")
		} else {
			// Hash the master password at startup so plaintext is never stored in memory
			var err error
			passwordHash, err = HashPassword(masterPassword)
			if err != nil {
				logger.Error("failed to hash master password", "error", err)
				// Fall back to plaintext comparison if hashing fails (should never happen)
				passwordHash = masterPassword
			} else {
				logger.Info("master password hashed with Argon2id")
			}
		}
	}

	return &Service{
		db:                 db,
		masterPasswordHash: passwordHash,
		sessionSecret:      sessionSecret,
		sessionDuration:    sessionDuration,
		logger:             logger,
		failedAttempts:     NewFailedLoginTracker(),
		authDisabled:       authDisabled,
	}
}

// IsAuthDisabled returns whether authentication is disabled
func (s *Service) IsAuthDisabled() bool {
	return s.authDisabled
}

// VerifyPassword checks if the provided password matches the master password
func (s *Service) VerifyPassword(password string) bool {
	// If auth is disabled, always return true
	if s.authDisabled {
		return true
	}
	return VerifyPasswordHash(password, s.masterPasswordHash)
}

// VerifyPasswordWithDelay checks password and enforces progressive delays
// Returns: valid bool, remainingDelay time.Duration
func (s *Service) VerifyPasswordWithDelay(password, clientIP string) (bool, time.Duration) {
	// Check if client needs to wait
	delay := s.failedAttempts.GetDelay(clientIP)
	if delay > 0 {
		return false, delay
	}

	if s.VerifyPassword(password) {
		s.failedAttempts.RecordSuccess(clientIP)
		return true, 0
	}

	s.failedAttempts.RecordFailure(clientIP)
	s.logger.Warn("failed login attempt", "ip", clientIP)
	return false, 0
}

// UpdatePassword updates the master password (in-memory only, resets on restart)
// For persistent password storage, this would need to be stored in the database
func (s *Service) UpdatePassword(newPassword string) error {
	passwordHash, err := HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}
	s.masterPasswordHash = passwordHash
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

	// ALWAYS use the secure HMAC-SHA256 hash for new sessions
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
// MIGRATION STRATEGY: Supports both HMAC-SHA256 (new) and SHA256 (legacy) for backward compatibility
// - Tries HMAC-SHA256 first (all new sessions)
// - Falls back to SHA256 only for old sessions
// - Automatically upgrades old sessions to HMAC-SHA256 on first use
func (s *Service) ValidateSession(token string) bool {
	if token == "" {
		return false
	}

	// Try new HMAC-SHA256 hash first
	tokenHash := hashToken(token)
	var expiresAt time.Time
	var sessionID string
	err := s.db.QueryRow(
		"SELECT id, expires_at FROM sessions WHERE token_hash = ?",
		tokenHash,
	).Scan(&sessionID, &expiresAt)

	if err == nil {
		if time.Now().After(expiresAt) {
			_, _ = s.db.Exec("DELETE FROM sessions WHERE token_hash = ?", tokenHash)
			return false
		}
		return true
	}

	// Fall back to legacy SHA256 hash for old sessions
	if err == sql.ErrNoRows {
		tokenHashLegacy := hashTokenLegacy(token)
		err = s.db.QueryRow(
			"SELECT id, expires_at FROM sessions WHERE token_hash = ?",
			tokenHashLegacy,
		).Scan(&sessionID, &expiresAt)

		if err == nil {
			if time.Now().After(expiresAt) {
				_, _ = s.db.Exec("DELETE FROM sessions WHERE token_hash = ?", tokenHashLegacy)
				return false
			}
			// Upgrade the session hash to new format
			_, _ = s.db.Exec("UPDATE sessions SET token_hash = ? WHERE id = ?", tokenHash, sessionID)
			return true
		}
	}

	return false
}

// InvalidateSession removes a session
// MIGRATION STRATEGY: Supports both hash formats to ensure old sessions can be properly invalidated
func (s *Service) InvalidateSession(token string) error {
	// Try new hash first
	tokenHash := hashToken(token)
	result, err := s.db.Exec("DELETE FROM sessions WHERE token_hash = ?", tokenHash)
	if err != nil {
		return err
	}

	// Check if any rows were affected
	rows, _ := result.RowsAffected()
	if rows > 0 {
		return nil
	}

	// Try legacy hash
	tokenHashLegacy := hashTokenLegacy(token)
	_, err = s.db.Exec("DELETE FROM sessions WHERE token_hash = ?", tokenHashLegacy)
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

// hashToken creates an HMAC-SHA256 hash of the token (new secure method)
// ALL NEW SESSIONS use this method exclusively.
func hashToken(token string) string {
	key := []byte("snipo-session-hmac-key-v1")
	h := hmac.New(sha256.New, key)
	h.Write([]byte(token))
	return hex.EncodeToString(h.Sum(nil))
}

// hashTokenLegacy creates a SHA256 hash of the token (for backward compatibility)
// MIGRATION STRATEGY:
//   - This is ONLY used for validating existing sessions created before the security upgrade
//   - When an old session is validated, it's automatically upgraded to HMAC-SHA256
//   - New sessions NEVER use this method
//   - This fallback can be removed after the session duration has passed (default: 168h/7 days)
//     (recommended: remove after 1-2 major version releases)
func hashTokenLegacy(token string) string {
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
