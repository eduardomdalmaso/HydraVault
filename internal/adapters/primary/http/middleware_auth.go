package http

import (
	"net/http"
	"strings"
)

// AuthMiddleware validates Bearer token or X-API-Key.
type AuthMiddleware struct {
	expectedToken string
}

// NewAuthMiddleware creates a new AuthMiddleware. If expectedToken is empty, accepts any non-empty bearer/key.
func NewAuthMiddleware(expectedToken string) *AuthMiddleware {
	return &AuthMiddleware{expectedToken: expectedToken}
}

// Wrap wraps an http.Handler with token authentication enforcement.
func (m *AuthMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow public healthcheck endpoint
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		token := extractToken(r)
		if token == "" {
			writeError(w, http.StatusUnauthorized, "missing authentication token (Authorization: Bearer <token> or X-API-Key required)")
			return
		}

		if m.expectedToken != "" && token != m.expectedToken {
			writeError(w, http.StatusForbidden, "invalid authentication token")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return strings.TrimSpace(parts[1])
		}
	}

	apiKey := r.Header.Get("X-API-Key")
	if apiKey != "" {
		return strings.TrimSpace(apiKey)
	}

	return ""
}
