package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// AuthMiddleware verifies the token with the auth service
func AuthMiddleware(authServiceURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			// Remove "Bearer " prefix
			token := strings.TrimPrefix(authHeader, "Bearer ")

			// Verify token with auth service
			req, err := http.NewRequest("GET", authServiceURL+"/internal/verify-token", nil)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			req.Header.Set("Authorization", "Bearer "+token)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil || resp.StatusCode != http.StatusOK {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}
			defer resp.Body.Close()

			// Parse user info from auth response
			var authResp struct {
				Valid bool `json:"valid"`
				User  struct {
					ID       int    `json:"id"`
					Username string `json:"username"`
					Email    string `json:"email"`
				} `json:"user"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
				http.Error(w, "Invalid response from auth service", http.StatusInternalServerError)
				return
			}

			if !authResp.Valid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Add user ID to request context
			ctx := context.WithValue(r.Context(), "userID", authResp.User.ID)
			ctx = context.WithValue(ctx, "username", authResp.User.Username)

			// Call next handler with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserIDFromContext extracts user ID from request context
func GetUserIDFromContext(r *http.Request) (int, bool) {
	userID, ok := r.Context().Value("userID").(int)
	return userID, ok
}
