package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
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

// VerifyToken handles token verification requests(Called by other microservices via /internal/verify-token)
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

// GetSession handles session info requests (replaces VerifyToken for frontend)(Called by React frontend or whatever)
func (h *TokenHandlers) GetSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
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

	// Return current session user information
	utils.SuccessResponse(w, map[string]interface{}{
		"user": user,
	})
}

// GetUserByID handles internal requests for user information by ID
func (h *TokenHandlers) GetUserByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/internal/user/")
	userID, err := strconv.Atoi(path)
	if err != nil {
		utils.ErrorResponse(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get user by ID (internal service call, no token needed)
	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		utils.ErrorResponse(w, "User not found", http.StatusNotFound)
		return
	}

	// Return user information for internal service use
	utils.SuccessResponse(w, map[string]interface{}{
		"user": user,
	})
}
