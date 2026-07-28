package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"social-network/services/auth/models"
	"social-network/services/auth/services"
	"social-network/services/auth/utils"
)

// AuthHandlers handles authentication-related HTTP requests
type AuthHandlers struct {
	authService *services.AuthService
}

// NewAuthHandlers creates a new auth handlers instance
func NewAuthHandlers(authService *services.AuthService) *AuthHandlers {
	return &AuthHandlers{
		authService: authService,
	}
}

func (h *AuthHandlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := h.authService.Ping(r.Context()); err != nil {
		utils.ErrorResponse(w, "Database unavailable", http.StatusServiceUnavailable)
		return
	}
	utils.SuccessResponse(w, map[string]string{"status": "healthy", "service": "auth"})
}

// Register handles POST /register requests
func (h *AuthHandlers) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	authResponse, err := h.authService.Register(&req)
	if err != nil {
		utils.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.SuccessResponse(w, authResponse)
}

// Login handles POST /login requests
func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	authResponse, err := h.authService.Login(&req)
	if err != nil {
		utils.ErrorResponse(w, err.Error(), http.StatusUnauthorized)
		return
	}

	utils.SuccessResponse(w, authResponse)
}

// Logout handles POST /logout requests
func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		utils.ErrorResponse(w, "Authorization header required", http.StatusBadRequest)
		return
	}

	// Remove "Bearer " prefix if present
	token := strings.TrimPrefix(authHeader, "Bearer ")

	err := h.authService.Logout(token)
	if err != nil {
		utils.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SuccessResponse(w, map[string]string{"message": "Logged out successfully"})
}
