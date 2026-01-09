package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/MohamedElashri/snipo/internal/auth"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService *auth.Service
	demoMode    bool
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *auth.Service) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		demoMode:    false,
	}
}

// WithDemoMode sets the demo mode flag
func (h *AuthHandler) WithDemoMode(enabled bool) *AuthHandler {
	h.demoMode = enabled
	return h
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
	if err := DecodeJSON(r, &req); err != nil {
		Error(w, r, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON payload")
		return
	}

	if req.Password == "" {
		Error(w, r, http.StatusBadRequest, "MISSING_PASSWORD", "Password is required")
		return
	}

	// Get client IP for rate limiting
	clientIP := getClientIPForAuth(r)

	// Verify password with progressive delay enforcement
	valid, delay := h.authService.VerifyPasswordWithDelay(req.Password, clientIP)
	if delay > 0 {
		w.Header().Set("Retry-After", fmt.Sprintf("%d", int(delay.Seconds())+1))
		Error(w, r, http.StatusTooManyRequests, "RATE_LIMITED",
			fmt.Sprintf("Too many failed attempts. Please wait %d seconds.", int(delay.Seconds())+1))
		return
	}

	if !valid {
		Error(w, r, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid password")
		return
	}

	// Create session
	token, err := h.authService.CreateSession()
	if err != nil {
		InternalError(w, r)
		return
	}

	// Set session cookie
	h.authService.SetSessionCookie(w, token)

	OK(w, r, LoginResponse{
		Success: true,
		Message: "Login successful",
	})
}

// getClientIPForAuth extracts client IP for authentication rate limiting
func getClientIPForAuth(r *http.Request) string {
	// Check X-Forwarded-For header (if behind proxy)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// Logout handles POST /api/v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	token := auth.GetSessionFromRequest(r)
	if token != "" {
		_ = h.authService.InvalidateSession(token)
	}

	h.authService.ClearSessionCookie(w)

	OK(w, r, LoginResponse{
		Success: true,
		Message: "Logout successful",
	})
}

// Check handles GET /api/v1/auth/check
func (h *AuthHandler) Check(w http.ResponseWriter, r *http.Request) {
	token := auth.GetSessionFromRequest(r)
	if token == "" || !h.authService.ValidateSession(token) {
		Unauthorized(w, r)
		return
	}

	OK(w, r, map[string]bool{"authenticated": true})
}
