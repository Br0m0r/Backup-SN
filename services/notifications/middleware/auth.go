package middleware

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
)

// AuthMiddleware verifies the session token with the auth service
func AuthMiddleware(authServiceURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := ExtractToken(r)
			if token == "" {
				http.Error(w, "Unauthorized: No token provided", http.StatusUnauthorized)
				return
			}

			// Verify token with auth service
			req, err := http.NewRequest("GET", authServiceURL+"/session", nil)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			req.Header.Set("Authorization", "Bearer "+token)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				http.Error(w, "Failed to verify token", http.StatusInternalServerError)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
				return
			}

			// Parse response to get user ID
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				http.Error(w, "Failed to read auth response", http.StatusInternalServerError)
				return
			}

			var authResp struct {
				Success bool `json:"success"`
				Data    struct {
					User struct {
						ID int `json:"id"`
					} `json:"user"`
				} `json:"data"`
			}

			if err := json.Unmarshal(body, &authResp); err != nil {
				http.Error(w, "Failed to parse auth response", http.StatusInternalServerError)
				return
			}

			if !authResp.Success || authResp.Data.User.ID == 0 {
				http.Error(w, "Unauthorized: Invalid user", http.StatusUnauthorized)
				return
			}

			// Add user ID to context
			ctx := context.WithValue(r.Context(), "userID", authResp.Data.User.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ExtractToken extracts token from header, query param, or cookie
func ExtractToken(r *http.Request) string {
	// Try query parameter first (for WebSocket)
	token := r.URL.Query().Get("token")
	if token != "" {
		return token
	}

	// Try Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
		return authHeader
	}

	// Try cookie
	cookie, err := r.Cookie("session_token")
	if err == nil {
		return cookie.Value
	}

	return ""
}

// GetUserIDFromContext retrieves user ID from context
func GetUserIDFromContext(r *http.Request) (int, bool) {
	userID, ok := r.Context().Value("userID").(int)
	return userID, ok
}

// GetAuthServiceURL returns the auth service URL
func GetAuthServiceURL() string {
	url := os.Getenv("AUTH_SERVICE_URL")
	if url == "" {
		url = "http://auth-service:8081"
	}
	return url
}
