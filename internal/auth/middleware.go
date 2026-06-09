package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// contextKey is a private type for context keys to avoid collisions.
type contextKey string

const (
	// ContextUserKey holds the authenticated user info in the request context.
	ContextUserKey contextKey = "user"
	// ContextUserIDKey holds the authenticated user ID in the request context.
	ContextUserIDKey contextKey = "user_id"
)

// Middleware returns an HTTP middleware that validates session tokens.
// The token is extracted from the Authorization header (Bearer <token>).
func Middleware(svc *Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token == "" {
				writeAuthError(w, "missing authorization header")
				return
			}

			user, err := svc.ValidateSession(token)
			if err != nil {
				writeAuthError(w, "invalid or expired session")
				return
			}

			ctx := context.WithValue(r.Context(), ContextUserKey, user)
			ctx = context.WithValue(ctx, ContextUserIDKey, user.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalMiddleware returns middleware that extracts user info from the
// session token if present, but does not reject unauthenticated requests.
func OptionalMiddleware(svc *Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token != "" {
				if user, err := svc.ValidateSession(token); err == nil {
					ctx := context.WithValue(r.Context(), ContextUserKey, user)
					ctx = context.WithValue(ctx, ContextUserIDKey, user.ID)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// GetUser extracts the authenticated user info from the request context.
func GetUser(r *http.Request) *UserInfo {
	user, _ := r.Context().Value(ContextUserKey).(*UserInfo)
	return user
}

// GetUserID extracts the authenticated user ID from the request context.
func GetUserID(r *http.Request) string {
	id, _ := r.Context().Value(ContextUserIDKey).(string)
	return id
}

// extractToken pulls the Bearer token from the Authorization header.
func extractToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		// Fall back to query param for WebSocket connections.
		return r.URL.Query().Get("token")
	}

	if !strings.HasPrefix(auth, "Bearer ") {
		return ""
	}

	return strings.TrimSpace(auth[7:])
}

// writeAuthError sends a 401 JSON error response.
func writeAuthError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"code":    "unauthorized",
			"message": message,
		},
	})
}
