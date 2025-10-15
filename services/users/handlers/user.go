package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"social-network/services/users/middleware"
	"social-network/services/users/models"
	"social-network/services/users/services"
	"social-network/services/users/utils"
)

// Helper functions for safe pointer logging
func getStrValue(s *string) string {
	if s == nil {
		return "nil"
	}
	return *s
}

func getBoolValue(b *bool) string {
	if b == nil {
		return "nil"
	}
	return fmt.Sprintf("%v", *b)
}

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

// GetCurrentUserProfile handles GET /profile (current authenticated user)
func (h *UserHandlers) GetCurrentUserProfile(w http.ResponseWriter, r *http.Request) {
	log.Printf("GetCurrentUserProfile: called for path=%s", r.URL.Path)

	// Get authenticated user ID from context
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		log.Printf("GetCurrentUserProfile: Failed to get user ID from context")
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Printf("GetCurrentUserProfile: Fetching profile for user %d", userID)

	// Get user profile
	user, err := h.userService.GetProfile(userID)
	if err != nil {
		log.Printf("GetCurrentUserProfile: Error fetching user %d: %v", userID, err)
		utils.ErrorResponse(w, err.Error(), http.StatusNotFound)
		return
	}

	// Return full profile for current user
	utils.SuccessResponse(w, map[string]interface{}{
		"user": user,
	})
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
		log.Printf("UpdateProfile: JSON decode error for user %d: %v", userID, err)
		utils.ErrorResponse(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("UpdateProfile: Request for user %d: nickname=%s, about_me=%s, is_public=%s",
		userID,
		getStrValue(req.Nickname),
		getStrValue(req.AboutMe),
		getBoolValue(req.IsPublicProfile))

	// Update profile
	user, err := h.userService.UpdateProfile(userID, &req)
	if err != nil {
		log.Printf("UpdateProfile: Database error for user %d: %v", userID, err)
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

// GetFollowStatus handles GET /follow/status/:id requests
func (h *UserHandlers) GetFollowStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get authenticated user ID from context
	followerID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract user ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/follow/status/")
	followingID, err := strconv.Atoi(path)
	if err != nil {
		utils.ErrorResponse(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get follow status
	status, err := h.userService.GetFollowStatus(followerID, followingID)
	if err != nil {
		utils.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SuccessResponse(w, map[string]interface{}{
		"status": status, // "none", "pending", or "accepted"
	})
}

// GetPendingFollowRequests handles GET /follow/requests
func (h *UserHandlers) GetPendingFollowRequests(w http.ResponseWriter, r *http.Request) {
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

	// Get pending follow requests
	requests, err := h.userService.GetPendingFollowRequests(userID)
	if err != nil {
		utils.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SuccessResponse(w, map[string]interface{}{
		"requests": requests,
		"count":    len(requests),
	})
}

// RespondToFollowRequest handles POST /follow/respond
func (h *UserHandlers) RespondToFollowRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get authenticated user ID from context (this is the user receiving the request)
	followingID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		FollowerID int  `json:"follower_id"`
		Accept     bool `json:"accept"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Respond to follow request
	err := h.userService.RespondToFollowRequest(req.FollowerID, followingID, req.Accept)
	if err != nil {
		utils.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	message := "Follow request rejected"
	if req.Accept {
		message = "Follow request accepted"
	}

	utils.SuccessResponse(w, map[string]interface{}{
		"message": message,
	})
}

// SearchUsers handles GET /search requests
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

// GetUserProfileByID handles GET /users/:id/profile
// Returns comprehensive profile with posts, followers, following
// Respects privacy settings (public vs private profiles)
func (h *UserHandlers) GetUserProfileByID(w http.ResponseWriter, r *http.Request) {
	// Get authenticated viewer ID from context
	viewerID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract user ID from URL path
	// Path format: /users/:id/profile
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		utils.ErrorResponse(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(pathParts[1])
	if err != nil {
		utils.ErrorResponse(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	log.Printf("GetUserProfileByID: viewer=%d requesting profile of user=%d", viewerID, userID)

	// Get comprehensive profile with privacy enforcement
	profile, err := h.userService.GetUserProfile(userID, viewerID)
	if err != nil {
		log.Printf("GetUserProfileByID: Error getting profile: %v", err)
		utils.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// If viewer cannot access profile, return limited info with can_view=false
	if !profile.CanView {
		log.Printf("GetUserProfileByID: Access denied for viewer=%d to user=%d profile (private)", viewerID, userID)
		utils.SuccessResponse(w, map[string]interface{}{
			"profile": profile,
			"message": "This profile is private. Follow this user to see their content.",
		})
		return
	}

	log.Printf("GetUserProfileByID: Returning profile for user=%d with %d posts, %d followers, %d following",
		userID, profile.PostCount, profile.FollowerCount, profile.FollowingCount)

	utils.SuccessResponse(w, map[string]interface{}{
		"profile": profile,
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
