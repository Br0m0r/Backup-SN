package authcache

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
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
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			// Remove "Bearer " prefix
			token := strings.TrimPrefix(authHeader, "Bearer ")

			// Check cache first
			cacheMutex.RLock()
			cached, exists := cache[token]
			cacheMutex.RUnlock()

			if exists && time.Now().Before(cached.ExpiresAt) {
				// Cache hit! Use cached data
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
			ctx := context.WithValue(r.Context(), "userID", user.UserID)
			ctx = context.WithValue(ctx, "username", user.Username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// verifyToken calls auth service to validate token
func verifyToken(authServiceURL, token string) (*CachedUser, error) {
	// Create HTTP client with 2 second timeout
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	// Create request
	req, err := http.NewRequest("GET", authServiceURL+"/internal/verify-token", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("auth service unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
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
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !authResp.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return &CachedUser{
		UserID:   authResp.User.ID,
		Username: authResp.User.Username,
		Email:    authResp.User.Email,
	}, nil
}

// GetUserIDFromContext extracts user ID from request context
func GetUserIDFromContext(r *http.Request) (int, bool) {
	userID, ok := r.Context().Value("userID").(int)
	return userID, ok
}

// GetUsernameFromContext extracts username from request context
func GetUsernameFromContext(r *http.Request) (string, bool) {
	username, ok := r.Context().Value("username").(string)
	return username, ok
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
