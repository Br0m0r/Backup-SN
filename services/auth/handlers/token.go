package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"social-network/services/auth/services"
	"social-network/services/auth/utils"
)

// TokenHandlers handles token-related HTTP requests
type TokenHandlers struct {
	authService *services.AuthService
}

// NewTokenHandlers creates a new token handlers instance
func NewTokenHandlers(authService *services.AuthService) *TokenHandlers {
	return &TokenHandlers{
		authService: authService,
	}
}

// VerifyToken handles token verification requests
func (h *TokenHandlers) VerifyToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "POST" {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		utils.ErrorResponse(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	// Remove "Bearer " prefix if present
	token := strings.TrimPrefix(authHeader, "Bearer ")

	user, err := h.authService.VerifyToken(token)
	if err != nil {
		utils.ErrorResponse(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	// Return user information (without password hash)
	response := map[string]interface{}{
		"valid": true,
		"user":  user,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
