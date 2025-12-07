package repository

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/MohamedElashri/snipo/internal/models"
)

// TokenRepository handles API token database operations
type TokenRepository struct {
	db *sql.DB
}

// NewTokenRepository creates a new token repository
func NewTokenRepository(db *sql.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

// generateToken generates a secure random token
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// hashToken creates a SHA256 hash of a token
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// Create creates a new API token
func (r *TokenRepository) Create(ctx context.Context, input *models.APITokenInput) (*models.APIToken, error) {
	// Generate token
	token, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	tokenHash := hashToken(token)

	// Validate permissions
	if input.Permissions == "" {
		input.Permissions = "read"
	}
	if input.Permissions != "read" && input.Permissions != "write" && input.Permissions != "admin" {
		return nil, fmt.Errorf("invalid permissions: must be 'read', 'write', or 'admin'")
	}

	query := `
		INSERT INTO api_tokens (name, token_hash, permissions, expires_at)
		VALUES (?, ?, ?, ?)
		RETURNING id, name, permissions, last_used_at, expires_at, created_at
	`

	apiToken := &models.APIToken{}
	err = r.db.QueryRowContext(ctx, query, input.Name, tokenHash, input.Permissions, input.ExpiresAt).Scan(
		&apiToken.ID,
		&apiToken.Name,
		&apiToken.Permissions,
		&apiToken.LastUsedAt,
		&apiToken.ExpiresAt,
		&apiToken.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create token: %w", err)
	}

	// Include the plain token in the response (only time it's returned)
	apiToken.Token = token

	return apiToken, nil
}

// GetByID retrieves a token by ID
func (r *TokenRepository) GetByID(ctx context.Context, id int64) (*models.APIToken, error) {
	query := `SELECT id, name, permissions, last_used_at, expires_at, created_at FROM api_tokens WHERE id = ?`

	token := &models.APIToken{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&token.ID,
		&token.Name,
		&token.Permissions,
		&token.LastUsedAt,
		&token.ExpiresAt,
		&token.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	return token, nil
}

// GetByToken retrieves a token by its plain text value (for authentication)
func (r *TokenRepository) GetByToken(ctx context.Context, token string) (*models.APIToken, error) {
	tokenHash := hashToken(token)

	query := `SELECT id, name, permissions, last_used_at, expires_at, created_at FROM api_tokens WHERE token_hash = ?`

	apiToken := &models.APIToken{}
	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&apiToken.ID,
		&apiToken.Name,
		&apiToken.Permissions,
		&apiToken.LastUsedAt,
		&apiToken.ExpiresAt,
		&apiToken.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	return apiToken, nil
}

// List retrieves all API tokens
func (r *TokenRepository) List(ctx context.Context) ([]models.APIToken, error) {
	query := `SELECT id, name, permissions, last_used_at, expires_at, created_at FROM api_tokens ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list tokens: %w", err)
	}
	defer rows.Close()

	var tokens []models.APIToken
	for rows.Next() {
		var token models.APIToken
		if err := rows.Scan(
			&token.ID,
			&token.Name,
			&token.Permissions,
			&token.LastUsedAt,
			&token.ExpiresAt,
			&token.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan token: %w", err)
		}
		tokens = append(tokens, token)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tokens: %w", err)
	}

	return tokens, nil
}

// Delete deletes a token
func (r *TokenRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM api_tokens WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// UpdateLastUsed updates the last_used_at timestamp for a token
func (r *TokenRepository) UpdateLastUsed(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE api_tokens SET last_used_at = ? WHERE id = ?`,
		time.Now().UTC(), id,
	)
	if err != nil {
		return fmt.Errorf("failed to update last used: %w", err)
	}
	return nil
}

// ValidateToken validates a token and returns it if valid
func (r *TokenRepository) ValidateToken(ctx context.Context, token string) (*models.APIToken, error) {
	apiToken, err := r.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	// Check if token is expired
	if apiToken.ExpiresAt != nil && apiToken.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("token expired")
	}

	// Update last used timestamp
	_ = r.UpdateLastUsed(ctx, apiToken.ID)

	return apiToken, nil
}
