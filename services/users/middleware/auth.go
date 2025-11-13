package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// AuthMiddleware verifies the token with the auth service
//explanation of this matryoshka pattern:
// We have three layers of functions here, each serving a specific purpose:
// 1. The outermost function `AuthMiddleware` takes the `authServiceURL` as a parameter.
//    This allows us to configure the middleware with the URL of the auth service when we set it up.
//    It returns a function that takes an `http.Handler` and returns another `http.Handler`.
// 2. The middle function (returned by `AuthMiddleware`) takes the next handler in the chain as a parameter.
//    This is the actual handler that will process the request if authentication succeeds.
//    It returns an `http.HandlerFunc`, which is the innermost layer.
// 3. The innermost function (the `http.HandlerFunc`) is where the actual request processing happens.
//    It checks the token, calls the auth service, adds user info to the context, and finally calls the next handler if everything is valid.

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
			resp, err := client.Do(req) //this is the actual HTTP call to auth service,creates an http client
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
