package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
)

type contextKey string

const UserIDKey contextKey = "userID"

// CORS middleware
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Logging middleware
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[%s] %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// GetUserIDFromContext retrieves the user ID from the request context
func GetUserIDFromContext(r *http.Request) (int, bool) {
	userID, ok := r.Context().Value(UserIDKey).(int)
	return userID, ok
}

// SetUserIDInContext sets the user ID in the request context
func SetUserIDInContext(r *http.Request, userID int) *http.Request {
	ctx := context.WithValue(r.Context(), UserIDKey, userID)
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
