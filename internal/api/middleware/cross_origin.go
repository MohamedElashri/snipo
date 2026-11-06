package middleware

import "net/http"

// ProtectBrowserWrites prevents cross-origin requests from using ambient session
// credentials. Validated API tokens are explicit credentials, not ambient cookies.
func ProtectBrowserWrites(trustedOrigins []string) func(http.Handler) http.Handler {
	protection := http.NewCrossOriginProtection()
	for _, origin := range trustedOrigins {
		// A CORS wildcard never grants permission to use ambient session credentials.
		if origin != "*" {
			_ = protection.AddTrustedOrigin(origin)
		}
	}
	return func(next http.Handler) http.Handler {
		protected := protection.Handler(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if GetTokenFromContext(r.Context()) != nil && !IsAnonymousAccess(r.Context()) {
				next.ServeHTTP(w, r)
				return
			}
			protected.ServeHTTP(w, r)
		})
	}
}
