package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/MohamedElashri/snipo/internal/auth"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService *auth.Service
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *auth.Service) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// LoginRequest represents a login request
type LoginRequest struct {
	Password string `json:"password"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON payload")
		return
	}

	if req.Password == "" {
		Error(w, http.StatusBadRequest, "MISSING_PASSWORD", "Password is required")
		return
	}

	if !h.authService.VerifyPassword(req.Password) {
		Error(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid password")
		return
	}

	// Create session
	token, err := h.authService.CreateSession()
	if err != nil {
		InternalError(w)
		return
	}

	// Set session cookie
	h.authService.SetSessionCookie(w, token)

	OK(w, LoginResponse{
		Success: true,
		Message: "Login successful",
	})
}

// Logout handles POST /api/v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	token := auth.GetSessionFromRequest(r)
	if token != "" {
		_ = h.authService.InvalidateSession(token)
	}

	h.authService.ClearSessionCookie(w)

	OK(w, LoginResponse{
		Success: true,
		Message: "Logout successful",
	})
}

// Check handles GET /api/v1/auth/check
func (h *AuthHandler) Check(w http.ResponseWriter, r *http.Request) {
	token := auth.GetSessionFromRequest(r)
	if token == "" || !h.authService.ValidateSession(token) {
		Unauthorized(w)
		return
	}

	OK(w, map[string]bool{"authenticated": true})
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// ChangePassword handles POST /api/v1/auth/change-password
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON payload")
		return
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		Error(w, http.StatusBadRequest, "MISSING_FIELDS", "Current and new password are required")
		return
	}

	if len(req.NewPassword) < 6 {
		Error(w, http.StatusBadRequest, "PASSWORD_TOO_SHORT", "Password must be at least 6 characters")
		return
	}

	// Verify current password
	if !h.authService.VerifyPassword(req.CurrentPassword) {
		Error(w, http.StatusUnauthorized, "INVALID_PASSWORD", "Current password is incorrect")
		return
	}

	// Update password
	if err := h.authService.UpdatePassword(req.NewPassword); err != nil {
		InternalError(w)
		return
	}

	OK(w, map[string]interface{}{
		"success": true,
		"message": "Password changed successfully",
	})
}
