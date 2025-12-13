package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"social-network/services/posts/middleware"
	"social-network/services/posts/models"
	"social-network/services/posts/services"
	"social-network/services/posts/utils"
)

// PostHandlers handles post-related HTTP requests
type PostHandlers struct {
	postService *services.PostService
}

// NewPostHandlers creates a new post handlers instance
func NewPostHandlers(postService *services.PostService) *PostHandlers {
	return &PostHandlers{
		postService: postService,
	}
}

// CreatePost handles POST /posts requests
func (h *PostHandlers) CreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get authenticated user ID from context
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	post, err := h.postService.CreatePost(&req, userID)
	if err != nil {
		utils.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.SuccessResponse(w, map[string]interface{}{
		"post": post,
	})
}

// GetFeed handles GET /posts requests
func (h *PostHandlers) GetFeed(w http.ResponseWriter, r *http.Request) {
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

	posts, err := h.postService.GetFeed(userID)
	if err != nil {
		log.Printf("GetFeed error for user %d: %v", userID, err)
		utils.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SuccessResponse(w, map[string]interface{}{
		"posts": posts,
	})
}

// SearchPosts handles GET /posts/search?q=query requests
func (h *PostHandlers) SearchPosts(w http.ResponseWriter, r *http.Request) {
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

	// Get search query from URL parameters
	query := r.URL.Query().Get("q")
	if query == "" {
		utils.SuccessResponse(w, map[string]interface{}{
			"posts": []*models.Post{},
		})
		return
	}

	posts, err := h.postService.SearchPosts(userID, query)
	if err != nil {
		log.Printf("SearchPosts error for user %d: %v", userID, err)
		utils.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SuccessResponse(w, map[string]interface{}{
		"posts": posts,
	})
}

// GetPost handles GET /posts/:id requests
func (h *PostHandlers) GetPost(w http.ResponseWriter, r *http.Request) {
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

	// Extract post ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/posts/")
	postID, err := strconv.Atoi(path)
	if err != nil {
		utils.ErrorResponse(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	post, err := h.postService.GetPost(postID, userID)
	if err != nil {
		if err.Error() == "access denied" {
			utils.ErrorResponse(w, err.Error(), http.StatusForbidden)
		} else {
			utils.ErrorResponse(w, err.Error(), http.StatusNotFound)
		}
		return
	}

	utils.SuccessResponse(w, map[string]interface{}{
		"post": post,
	})
}

// UpdatePost handles PUT /posts/:id requests
func (h *PostHandlers) UpdatePost(w http.ResponseWriter, r *http.Request) {
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

	// Extract post ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/posts/")
	postID, err := strconv.Atoi(path)
	if err != nil {
		utils.ErrorResponse(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	var req models.UpdatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	post, err := h.postService.UpdatePost(postID, userID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			utils.ErrorResponse(w, err.Error(), http.StatusForbidden)
		} else {
			utils.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	utils.SuccessResponse(w, map[string]interface{}{
		"post": post,
	})
}

// DeletePost handles DELETE /posts/:id requests
func (h *PostHandlers) DeletePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get authenticated user ID from context
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract post ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/posts/")
	postID, err := strconv.Atoi(path)
	if err != nil {
		utils.ErrorResponse(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	err = h.postService.DeletePost(postID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			utils.ErrorResponse(w, err.Error(), http.StatusForbidden)
		} else {
			utils.ErrorResponse(w, err.Error(), http.StatusNotFound)
		}
		return
	}

	utils.SuccessResponse(w, map[string]string{
		"message": "Post deleted successfully",
	})
}

// CreateComment handles POST /comments requests
func (h *PostHandlers) CreateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get authenticated user ID from context
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	username, ok := middleware.GetUsernameFromContext(r)
	if !ok {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	comment, err := h.postService.CreateComment(&req, userID, username)
	if err != nil {
		if err.Error() == "access denied: cannot comment on this post" {
			utils.ErrorResponse(w, err.Error(), http.StatusForbidden)
		} else {
			utils.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	utils.SuccessResponse(w, map[string]interface{}{
		"comment": comment,
	})
}

// GetComments handles GET /comments?post_id=:id requests
func (h *PostHandlers) GetComments(w http.ResponseWriter, r *http.Request) {
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

	// Extract post_id from query parameters
	postIDStr := r.URL.Query().Get("post_id")
	if postIDStr == "" {
		utils.ErrorResponse(w, "post_id query parameter required", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		utils.ErrorResponse(w, "Invalid post_id", http.StatusBadRequest)
		return
	}

	comments, err := h.postService.GetComments(postID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "access denied") {
			utils.ErrorResponse(w, err.Error(), http.StatusForbidden)
		} else {
			utils.ErrorResponse(w, err.Error(), http.StatusNotFound)
		}
		return
	}

	utils.SuccessResponse(w, map[string]interface{}{
		"comments": comments,
	})
}

// UpdateComment handles PUT /comments/:id requests
func (h *PostHandlers) UpdateComment(w http.ResponseWriter, r *http.Request) {
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

	// Extract comment ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/comments/")
	commentID, err := strconv.Atoi(path)
	if err != nil {
		utils.ErrorResponse(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req models.UpdateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Update comment
	comment, err := h.postService.UpdateComment(commentID, userID, req.Content, req.ImagePath)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			utils.ErrorResponse(w, err.Error(), http.StatusForbidden)
		} else if strings.Contains(err.Error(), "not found") {
			utils.ErrorResponse(w, err.Error(), http.StatusNotFound)
		} else {
			utils.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	utils.SuccessResponse(w, map[string]interface{}{
		"comment": comment,
	})
}

// DeleteComment handles DELETE /comments/:id requests
func (h *PostHandlers) DeleteComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get authenticated user ID from context
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract comment ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/comments/")
	commentID, err := strconv.Atoi(path)
	if err != nil {
		utils.ErrorResponse(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	// Delete comment
	err = h.postService.DeleteComment(commentID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			utils.ErrorResponse(w, err.Error(), http.StatusForbidden)
		} else if strings.Contains(err.Error(), "not found") {
			utils.ErrorResponse(w, err.Error(), http.StatusNotFound)
		} else {
			utils.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	utils.SuccessResponse(w, map[string]interface{}{
		"message": "Comment deleted successfully",
	})
}

// GetGroupPosts handles GET /posts/group/:id requests
func (h *PostHandlers) GetGroupPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get authenticated user ID from context (for future access control)
	_, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract group ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/posts/group/")
	groupID, err := strconv.Atoi(path)
	if err != nil {
		utils.ErrorResponse(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	posts, err := h.postService.GetGroupPosts(groupID)
	if err != nil {
		log.Printf("GetGroupPosts error for group %d: %v", groupID, err)
		utils.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SuccessResponse(w, map[string]interface{}{
		"posts": posts,
	})
}

// HealthHandler handles GET /health requests
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "healthy",
		"service": "post",
		"message": "Post service is running",
	})
}
