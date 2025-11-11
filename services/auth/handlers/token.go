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

// VerifyToken handles GET/POST /internal/verify-token requests
// Called by other microservices for token validation
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

// GetSession handles GET /session requests
// Called by frontend to get current authenticated user info
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

// GetUserByID handles GET /internal/user/{id} requests
// Internal endpoint for other microservices to get user info by ID
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

// GetUsersBatch handles POST /internal/users/batch requests
// Internal endpoint for other microservices to get multiple users at once
func (h *TokenHandlers) GetUsersBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserIDs []int `json:"user_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if len(req.UserIDs) == 0 {
		utils.SuccessResponse(w, map[string]interface{}{
			"users": []interface{}{},
		})
		return
	}

	// Get users by IDs
	users, err := h.authService.GetUsersByIDs(req.UserIDs)
	if err != nil {
		utils.ErrorResponse(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}

	utils.SuccessResponse(w, map[string]interface{}{
		"users": users,
	})
}
