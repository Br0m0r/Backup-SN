package middleware

import (
	"context"
	"net/http"
	"strings"

	"social-network/services/common/observability"
)

// Logging middleware
func Logging(next http.Handler) http.Handler {
	return observability.HTTPLogging("chat", next)
}

// GetUserIDFromContext retrieves the user ID from the request context
func GetUserIDFromContext(r *http.Request) (int, bool) {
	userID, ok := r.Context().Value("userID").(int)
	return userID, ok
}

// SetUserIDInContext sets the user ID in the request context
func SetUserIDInContext(r *http.Request, userID int) *http.Request {
	ctx := context.WithValue(r.Context(), "userID", userID)
	return r.WithContext(ctx)
}

// ExtractTokenFromHeader extracts the token from Authorization header
func ExtractTokenFromHeader(r *http.Request) string {
	// First, try query parameter (for WebSocket connections)
	token := r.URL.Query().Get("token")
	if token != "" {
		return token
	}

	// Then try Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		// Try cookie as fallback
		cookie, err := r.Cookie("session_token")
		if err == nil {
			return cookie.Value
		}
		return ""
	}

	// Support both "Bearer <token>" and direct token
	parts := strings.Split(authHeader, " ")
	if len(parts) == 2 && parts[0] == "Bearer" {
		return parts[1]
	}
	return authHeader
}
