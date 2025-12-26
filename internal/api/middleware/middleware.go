package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/MohamedElashri/snipo/internal/auth"
	"github.com/MohamedElashri/snipo/internal/repository"
)

// Context keys for authentication and request tracking
type contextKey string

const (
	// ContextKeyAPIToken is the context key for API token
	ContextKeyAPIToken contextKey = "api_token"
	// ContextKeyRequestID is the context key for request ID
	ContextKeyRequestID contextKey = "request_id"
)

// API version
const APIVersion = "1.0"

// RequestID generates a unique request ID for tracking
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request already has an ID (from proxy/load balancer)
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			// Generate new UUID
			requestID = uuid.New().String()
		}

		// Add to response header
		w.Header().Set("X-Request-ID", requestID)

		// Add to context for use in handlers and logging
		ctx := context.WithValue(r.Context(), ContextKeyRequestID, requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID retrieves the request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(ContextKeyRequestID).(string); ok {
		return requestID
	}
	return ""
}

// SecurityHeaders adds essential security headers to responses
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// API version header
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("X-API-Version", APIVersion)
		}

		// Prevent XSS
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Content Security Policy - differentiate between API and web routes
		if strings.HasPrefix(r.URL.Path, "/api/") {
			// Stricter CSP for API-only routes (JSON responses)
			w.Header().Set("Content-Security-Policy", strings.Join([]string{
				"default-src 'none'",
				"frame-ancestors 'none'",
			}, "; "))
		} else {
			// Full CSP for web UI routes
			w.Header().Set("Content-Security-Policy", strings.Join([]string{
				"default-src 'self'",
				"script-src 'self' 'unsafe-inline' 'unsafe-eval' blob:", // unsafe-eval needed for Alpine.js, blob for Ace workers
				"style-src 'self' 'unsafe-inline'",
				"img-src 'self' data: blob:",
				"font-src 'self'",
				"connect-src 'self'",
				"worker-src 'self' blob:", // Allow Ace Editor web workers
				"frame-ancestors 'none'",
				"form-action 'self'",
				"base-uri 'self'",
			}, "; "))
		}

		// HTTPS enforcement (only in production)
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Referrer policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions policy
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

		next.ServeHTTP(w, r)
	})
}

// Logger logs HTTP requests with request ID
func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)

			// Get request ID from context
			requestID := GetRequestID(r.Context())

			logger.Info("request",
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.statusCode,
				"duration", duration,
				"ip", getClientIP(r),
			)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Recovery recovers from panics and logs the error
func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("panic recovered",
						"error", err,
						"stack", string(debug.Stack()),
						"path", r.URL.Path,
					)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAuth checks for valid authentication (session or API token)
func RequireAuth(authService *auth.Service) func(http.Handler) http.Handler {
	return RequireAuthWithTokenRepo(authService, nil)
}

// RequireAuthWithTokenRepo checks for valid authentication with API token support
func RequireAuthWithTokenRepo(authService *auth.Service, tokenRepo *repository.TokenRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If authentication is disabled, allow all requests
			if authService.IsAuthDisabled() {
				next.ServeHTTP(w, r)
				return
			}
			
			// First, check for API token in header
			if tokenRepo != nil {
				// Check Authorization header (Bearer token)
				authHeader := r.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					token := strings.TrimPrefix(authHeader, "Bearer ")
					apiToken, err := tokenRepo.ValidateToken(r.Context(), token)
					if err == nil && apiToken != nil {
						// Valid API token, add to context and continue
						ctx := context.WithValue(r.Context(), ContextKeyAPIToken, apiToken)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}

				// Check X-API-Key header
				apiKey := r.Header.Get("X-API-Key")
				if apiKey != "" {
					apiToken, err := tokenRepo.ValidateToken(r.Context(), apiKey)
					if err == nil && apiToken != nil {
						// Valid API token, add to context and continue
						ctx := context.WithValue(r.Context(), ContextKeyAPIToken, apiToken)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}
			}

			// Fall back to session authentication
			sessionToken := auth.GetSessionFromRequest(r)
			if sessionToken != "" && authService.ValidateSession(sessionToken) {
				next.ServeHTTP(w, r)
				return
			}

			// No valid authentication found
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
			} else {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
			}
		})
	}
}

// RateLimiter implements a simple in-memory rate limiter
type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// Middleware returns the rate limiting middleware
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		now := time.Now()

		rl.mu.Lock()

		// Clean old requests for this IP
		var recent []time.Time
		for _, t := range rl.requests[ip] {
			if now.Sub(t) < rl.window {
				recent = append(recent, t)
			}
		}

		if len(recent) >= rl.limit {
			rl.mu.Unlock()
			w.Header().Set("Retry-After", "60")
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		rl.requests[ip] = append(recent, now)
		rl.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

// cleanup periodically removes old entries
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, times := range rl.requests {
			var recent []time.Time
			for _, t := range times {
				if now.Sub(t) < rl.window {
					recent = append(recent, t)
				}
			}
			if len(recent) == 0 {
				delete(rl.requests, ip)
			} else {
				rl.requests[ip] = recent
			}
		}
		rl.mu.Unlock()
	}
}

// TrustProxy controls whether to trust X-Forwarded-For headers
// Set to true only when behind a trusted reverse proxy
var TrustProxy = false

// getClientIP extracts the client IP from the request
// WARNING: X-Forwarded-For can be spoofed if not behind a trusted proxy
func getClientIP(r *http.Request) string {
	// Only trust proxy headers if explicitly configured
	if TrustProxy {
		// Check X-Forwarded-For header (take rightmost untrusted IP for security)
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			ips := strings.Split(xff, ",")
			if len(ips) > 0 {
				// Use the first IP (client IP) - in production, consider using
				// the rightmost IP not in your trusted proxy list
				return strings.TrimSpace(ips[0])
			}
		}

		// Check X-Real-IP header
		if xri := r.Header.Get("X-Real-IP"); xri != "" {
			return xri
		}
	}

	// Fall back to RemoteAddr (direct connection IP)
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// CORS adds CORS headers for API requests
// For local-first deployment, CORS is restrictive by default.
// Configure SNIPO_ALLOWED_ORIGINS to allow specific cross-origin requests.
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			if origin != "" {
				allowed := false
				
				// Check if wildcard is configured (development mode)
				for _, allowedOrigin := range allowedOrigins {
					if allowedOrigin == "*" {
						w.Header().Set("Access-Control-Allow-Origin", "*")
						allowed = true
						break
					} else if allowedOrigin == origin {
						w.Header().Set("Access-Control-Allow-Origin", origin)
						w.Header().Set("Access-Control-Allow-Credentials", "true")
						allowed = true
						break
					}
				}

				// If not explicitly allowed, check if same-origin
				if !allowed {
					requestHost := r.Host
					if origin == "http://"+requestHost || origin == "https://"+requestHost {
						w.Header().Set("Access-Control-Allow-Origin", origin)
						w.Header().Set("Access-Control-Allow-Credentials", "true")
					}
					// Otherwise, don't set CORS headers (browser will block cross-origin requests)
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
			w.Header().Set("Access-Control-Max-Age", "86400")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
