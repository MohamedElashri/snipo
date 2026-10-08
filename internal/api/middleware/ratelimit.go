package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// APIRateLimiter implements rate limiting for API endpoints with proper headers
type APIRateLimiter struct {
	requests   map[string][]time.Time // key = IP or token ID
	mu         sync.RWMutex
	readLimit  int // requests per hour for read operations
	writeLimit int // requests per hour for write operations
	adminLimit int // requests per hour for admin operations
	window     time.Duration
}

// RateLimitConfig holds rate limit configuration
type RateLimitConfig struct {
	ReadLimit  int           // default: 1000 req/hour
	WriteLimit int           // default: 500 req/hour
	AdminLimit int           // default: 100 req/hour
	Window     time.Duration // default: 1 hour
}

// NewAPIRateLimiter creates a new API rate limiter with permission-based limits
func NewAPIRateLimiter(config RateLimitConfig) *APIRateLimiter {
	if config.ReadLimit == 0 {
		config.ReadLimit = 1000
	}
	if config.WriteLimit == 0 {
		config.WriteLimit = 500
	}
	if config.AdminLimit == 0 {
		config.AdminLimit = 100
	}
	if config.Window == 0 {
		config.Window = time.Hour
	}

	rl := &APIRateLimiter{
		requests:   make(map[string][]time.Time),
		readLimit:  config.ReadLimit,
		writeLimit: config.WriteLimit,
		adminLimit: config.AdminLimit,
		window:     config.Window,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// RateLimitByPermission returns middleware that rate limits based on permission level
func (rl *APIRateLimiter) RateLimitByPermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Determine the limit based on permission
			var limit int
			switch permission {
			case PermissionAdmin:
				limit = rl.adminLimit
			case PermissionWrite:
				limit = rl.writeLimit
			case PermissionRead:
				limit = rl.readLimit
			default:
				limit = rl.readLimit
			}

			// Get identifier (token ID or IP)
			identifier := rl.getIdentifier(r)
			now := time.Now()

			rl.mu.Lock()

			// Clean old requests for this identifier
			var recent []time.Time
			for _, t := range rl.requests[identifier] {
				if now.Sub(t) < rl.window {
					recent = append(recent, t)
				}
			}

			// Check if limit is exceeded
			if len(recent) >= limit {
				rl.mu.Unlock()
				retryAfter := int(rl.window.Seconds())
				reset := now.Add(rl.window).Unix()

				// Set rate limit headers
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", reset))
				w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))

				http.Error(w, `{"error":{"code":"RATE_LIMIT_EXCEEDED","message":"Rate limit exceeded. Please try again later."}}`, http.StatusTooManyRequests)
				return
			}

			// Add current request
			rl.requests[identifier] = append(recent, now)

			// Calculate remaining after adding this request
			remaining := limit - len(rl.requests[identifier])

			// Set rate limit headers
			reset := now.Add(rl.window).Unix()
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, remaining)))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", reset))

			rl.mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}

// getIdentifier returns a unique identifier for rate limiting
// Uses token ID if available (for more accurate per-user limits), otherwise IP
func (rl *APIRateLimiter) getIdentifier(r *http.Request) string {
	// Try to get token from context
	token := GetTokenFromContext(r.Context())
	if token != nil {
		return fmt.Sprintf("token:%d", token.ID)
	}

	// Fall back to IP address (for session-based auth)
	return "ip:" + getClientIP(r)
}

// cleanup periodically removes old entries to prevent memory leaks
func (rl *APIRateLimiter) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for identifier, times := range rl.requests {
			var recent []time.Time
			for _, t := range times {
				if now.Sub(t) < rl.window {
					recent = append(recent, t)
				}
			}
			if len(recent) == 0 {
				delete(rl.requests, identifier)
			} else {
				rl.requests[identifier] = recent
			}
		}
		rl.mu.Unlock()
	}
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// RateLimitRead is a convenience function for read operation rate limiting
func (rl *APIRateLimiter) RateLimitRead(next http.Handler) http.Handler {
	return rl.RateLimitByPermission(PermissionRead)(next)
}

// RateLimitWrite is a convenience function for write operation rate limiting
func (rl *APIRateLimiter) RateLimitWrite(next http.Handler) http.Handler {
	return rl.RateLimitByPermission(PermissionWrite)(next)
}

// RateLimitAdmin is a convenience function for admin operation rate limiting
func (rl *APIRateLimiter) RateLimitAdmin(next http.Handler) http.Handler {
	return rl.RateLimitByPermission(PermissionAdmin)(next)
}
