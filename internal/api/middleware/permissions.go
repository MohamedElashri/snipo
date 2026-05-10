package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/MohamedElashri/snipo/internal/auth"
	"github.com/MohamedElashri/snipo/internal/models"
)

const adminPasswordHeader = "X-Snipo-Master-Password"

// Permission levels
const (
	PermissionRead  = "read"
	PermissionWrite = "write"
	PermissionAdmin = "admin"
)

// GetTokenFromContext retrieves the API token from context
func GetTokenFromContext(ctx context.Context) *models.APIToken {
	if token, ok := ctx.Value(ContextKeyAPIToken).(*models.APIToken); ok {
		return token
	}
	return nil
}

// IsAnonymousAccess reports whether the request was allowed by disable-login.
func IsAnonymousAccess(ctx context.Context) bool {
	anonymous, _ := ctx.Value(ContextKeyAnonymousAccess).(bool)
	return anonymous
}

// CheckPermission returns middleware that checks if the request has required permission level
func CheckPermission(required string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from context (set by RequireAuthWithTokenRepo middleware)
			token := GetTokenFromContext(r.Context())

			// If no token, this must be a session-based auth (full admin access)
			if token == nil {
				next.ServeHTTP(w, r)
				return
			}

			// Check if token has required permission
			if !hasPermission(token.Permissions, required) {
				http.Error(w, `{"error":{"code":"INSUFFICIENT_PERMISSIONS","message":"Token does not have required permissions"}}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// hasPermission checks if the token's permission level is sufficient
func hasPermission(tokenPermission, required string) bool {
	// Admin has all permissions
	if tokenPermission == PermissionAdmin {
		return true
	}

	// Write has read + write permissions
	if tokenPermission == PermissionWrite {
		return required == PermissionRead || required == PermissionWrite
	}

	// Read only has read permission
	if tokenPermission == PermissionRead {
		return required == PermissionRead
	}

	return false
}

// RequireRead is a convenience middleware for read operations
func RequireRead(next http.Handler) http.Handler {
	return CheckPermission(PermissionRead)(next)
}

// RequireWrite is a convenience middleware for write operations
func RequireWrite(next http.Handler) http.Handler {
	return CheckPermission(PermissionWrite)(next)
}

// RequireAdmin is a convenience middleware for admin operations
func RequireAdmin(next http.Handler) http.Handler {
	return CheckPermission(PermissionAdmin)(next)
}

// RequireAdminWithPassword allows normal admin sessions/tokens, but requires
// the master password when the request is anonymous because login is disabled.
func RequireAdminWithPassword(authService *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := GetTokenFromContext(r.Context())
			if token != nil {
				if !hasPermission(token.Permissions, PermissionAdmin) {
					http.Error(w, `{"error":{"code":"INSUFFICIENT_PERMISSIONS","message":"Admin permission required"}}`, http.StatusForbidden)
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			if !IsAnonymousAccess(r.Context()) {
				next.ServeHTTP(w, r)
				return
			}

			password := r.Header.Get(adminPasswordHeader)
			if password == "" {
				http.Error(w, `{"error":{"code":"ADMIN_PASSWORD_REQUIRED","message":"Master password is required for admin operations when login is disabled"}}`, http.StatusUnauthorized)
				return
			}

			valid, delay := authService.VerifyPasswordWithDelay(password, ClientIP(r))
			if delay > 0 {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(delay.Seconds())+1))
				http.Error(w, `{"error":{"code":"RATE_LIMITED","message":"Too many failed attempts. Please wait before retrying."}}`, http.StatusTooManyRequests)
				return
			}
			if !valid {
				http.Error(w, `{"error":{"code":"INVALID_PASSWORD","message":"Invalid password"}}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// PermissionByMethod returns middleware that checks permission based on HTTP method
// GET = read, POST/PUT/PATCH/DELETE = write
func PermissionByMethod(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method := strings.ToUpper(r.Method)

		var required string
		switch method {
		case "GET", "HEAD", "OPTIONS":
			required = PermissionRead
		case "POST", "PUT", "PATCH", "DELETE":
			required = PermissionWrite
		default:
			required = PermissionRead
		}

		CheckPermission(required)(next).ServeHTTP(w, r)
	})
}
