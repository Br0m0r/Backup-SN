package authcache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type contextKey string

const (
	userIDKey   contextKey = "userID"
	usernameKey contextKey = "username"
)

// CachedUser stores validated user information with expiry
type CachedUser struct {
	UserID    int
	Username  string
	Email     string
	ExpiresAt time.Time
}

var (
	cache      = make(map[string]CachedUser)
	cacheMutex sync.RWMutex
)

// AuthMiddleware creates middleware that validates tokens with caching
func AuthMiddleware(authServiceURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from query parameter, Authorization header, or cookie
			token := extractToken(r)
			if token == "" {
				http.Error(w, "Authorization required", http.StatusUnauthorized)
				return
			}

			// Check cache first
			cacheMutex.RLock()
			cached, exists := cache[token]
			cacheMutex.RUnlock()

			if exists && time.Now().Before(cached.ExpiresAt) {
				// Cache hit! Use cached data
				// Use raw string keys for compatibility with middleware packages
				ctx := context.WithValue(r.Context(), "userID", cached.UserID)
				ctx = context.WithValue(ctx, "username", cached.Username)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Cache miss - verify with auth service
			user, err := verifyToken(authServiceURL, token)
			if err != nil {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Cache the result for 5 minutes
			cacheMutex.Lock()
			cache[token] = CachedUser{
				UserID:    user.UserID,
				Username:  user.Username,
				Email:     user.Email,
				ExpiresAt: time.Now().Add(5 * time.Minute),
			}
			cacheMutex.Unlock()

			// Add user info to context and proceed
			// Use raw string keys for compatibility with middleware packages
			ctx := context.WithValue(r.Context(), "userID", user.UserID)
			ctx = context.WithValue(ctx, "username", user.Username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// verifyToken calls auth service to validate token
func verifyToken(authServiceURL, token string) (*CachedUser, error) {
	log.Printf("[AuthCache] Verifying token with auth service: %s", authServiceURL)

	// Create HTTP client with 2 second timeout
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	// Create request
	req, err := http.NewRequest("GET", authServiceURL+"/internal/verify-token", nil)
	if err != nil {
		log.Printf("[AuthCache] Failed to create request: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[AuthCache] Auth service unreachable: %v", err)
		return nil, fmt.Errorf("auth service unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[AuthCache] Invalid token, status: %d", resp.StatusCode)
		return nil, fmt.Errorf("invalid token (status %d)", resp.StatusCode)
	}

	// Parse response
	var authResp struct {
		Valid bool `json:"valid"`
		User  struct {
			ID       int    `json:"id"`
			Username string `json:"username"`
			Email    string `json:"email"`
		} `json:"user"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		log.Printf("[AuthCache] Failed to decode response: %v", err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !authResp.Valid {
		log.Printf("[AuthCache] Token marked as invalid by auth service")
		return nil, fmt.Errorf("invalid token")
	}

	log.Printf("[AuthCache] Token verified successfully for user %d (%s)", authResp.User.ID, authResp.User.Username)

	return &CachedUser{
		UserID:   authResp.User.ID,
		Username: authResp.User.Username,
		Email:    authResp.User.Email,
	}, nil
}

// GetUserIDFromContext extracts user ID from request context
// Try both the typed key and raw string for compatibility
func GetUserIDFromContext(r *http.Request) (int, bool) {
	// Try typed key first
	if userID, ok := r.Context().Value(userIDKey).(int); ok {
		return userID, true
	}
	// Try raw string key for compatibility with other middleware
	if userID, ok := r.Context().Value("userID").(int); ok {
		return userID, true
	}
	return 0, false
}

// GetUsernameFromContext extracts username from request context
func GetUsernameFromContext(r *http.Request) (string, bool) {
	// Try typed key first
	if username, ok := r.Context().Value(usernameKey).(string); ok {
		return username, true
	}
	// Try raw string key for compatibility
	if username, ok := r.Context().Value("username").(string); ok {
		return username, true
	}
	return "", false
}

// extractToken extracts token from query parameter, Authorization header, or cookie
func extractToken(r *http.Request) string {
	// Try query parameter first (for WebSocket connections)
	token := r.URL.Query().Get("token")
	if token != "" {
		return token
	}

	// Try Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		// Support both "Bearer <token>" and direct token
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
		return authHeader
	}

	// Try cookie as fallback
	cookie, err := r.Cookie("session_token")
	if err == nil {
		return cookie.Value
	}

	return ""
}

// InvalidateToken removes a token from the cache (useful for logout)
func InvalidateToken(token string) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	delete(cache, token)
}

// ClearExpiredTokens removes expired tokens from cache (optional cleanup)
func ClearExpiredTokens() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	now := time.Now()
	for token, user := range cache {
		if now.After(user.ExpiresAt) {
			delete(cache, token)
		}
	}
}
