package middleware

import (
	"net/http"
	"os"
	"strings"
)

// CORS returns a middleware that sets CORS headers for cross-origin browser requests.
// Allowed origins come from CORS_ALLOWED_ORIGINS (comma-separated). Falls back to "*"
// when unset (development only).
func CORS(next http.Handler) http.Handler {
	allowedOrigins := parseOrigins(os.Getenv("CORS_ALLOWED_ORIGINS"))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if len(allowedOrigins) == 0 {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else if originAllowed(allowedOrigins, origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-Request-ID")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func parseOrigins(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func normalizeOrigin(origin string) string {
	return strings.TrimRight(strings.TrimSpace(origin), "/")
}

func originAllowed(allowedOrigins []string, origin string) bool {
	if origin == "" {
		return false
	}
	normalized := normalizeOrigin(origin)
	for _, allowed := range allowedOrigins {
		if allowed == "*" {
			return true
		}
		if normalizeOrigin(allowed) == normalized {
			return true
		}
	}
	return false
}
