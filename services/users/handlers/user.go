package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"social-network/services/users/middleware"
	"social-network/services/users/models"
	"social-network/services/users/services"
	"social-network/services/users/utils"
)

// UserHandlers contains all user-related HTTP handlers
type UserHandlers struct {
	userService *services.UserService
}

// NewUserHandlers creates a new user handlers instance
func NewUserHandlers(userService *services.UserService) *UserHandlers {
	return &UserHandlers{
		userService: userService,
	}
}

// GetProfile handles GET /profile/:id requests
func (h *UserHandlers) GetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/profile/")
	userID, err := strconv.Atoi(path)
	if err != nil {
		utils.ErrorResponse(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get authenticated user ID from context
	authUserID, _ := middleware.GetUserIDFromContext(r)

	// Get user profile
	user, err := h.userService.GetProfile(userID)
	if err != nil {
		utils.ErrorResponse(w, err.Error(), http.StatusNotFound)
		return
	}

	// Only show full profile (with email and DOB) to the user themselves
	if authUserID == userID {
		utils.SuccessResponse(w, map[string]interface{}{
			"user": user,
		})
	} else {
		// Show public profile to others (no email, no DOB)
		utils.SuccessResponse(w, map[string]interface{}{
			"user": user.PublicProfile(),
		})
	}
}

// UpdateProfile handles PUT /profile requests
func (h *UserHandlers) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get authenticated user ID from context
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req models.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update profile
	user, err := h.userService.UpdateProfile(userID, &req)
	if err != nil {
		utils.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SuccessResponse(w, map[string]interface{}{
		"user":    user,
		"message": "Profile updated successfully",
	})
}

// FollowUser handles POST /follow requests
func (h *UserHandlers) FollowUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get authenticated user ID from context
	followerID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req models.FollowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Follow user
	err := h.userService.FollowUser(followerID, req.UserID)
	if err != nil {
		utils.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.SuccessResponse(w, map[string]interface{}{
		"message": "Follow request sent successfully",
	})
}

// UnfollowUser handles DELETE /follow requests
func (h *UserHandlers) UnfollowUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get authenticated user ID from context
	followerID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req models.FollowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Unfollow user
	err := h.userService.UnfollowUser(followerID, req.UserID)
	if err != nil {
		utils.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.SuccessResponse(w, map[string]interface{}{
		"message": "Unfollowed successfully",
	})
}

// GetFollowers handles GET /followers requests
func (h *UserHandlers) GetFollowers(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get authenticated user ID from context
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get followers
	followers, err := h.userService.GetFollowers(userID)
	if err != nil {
		utils.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SuccessResponse(w, map[string]interface{}{
		"followers": followers,
		"count":     len(followers),
	})
}

// GetFollowing handles GET /following requests
func (h *UserHandlers) GetFollowing(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get authenticated user ID from context
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get following
	following, err := h.userService.GetFollowing(userID)
	if err != nil {
		utils.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SuccessResponse(w, map[string]interface{}{
		"following": following,
		"count":     len(following),
	})
}

// SearchUsers handles GET /search?q=searchTerm requests
func (h *UserHandlers) SearchUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get search term from query params
	searchTerm := r.URL.Query().Get("q")
	if searchTerm == "" {
		utils.ErrorResponse(w, "Search term is required", http.StatusBadRequest)
		return
	}

	// Search users
	users, err := h.userService.SearchUsers(searchTerm)
	if err != nil {
		utils.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SuccessResponse(w, map[string]interface{}{
		"users": users,
		"count": len(users),
	})
}

// HealthHandler handles health check requests
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "healthy",
		"service": "user",
		"message": "User service is running",
	})
}
