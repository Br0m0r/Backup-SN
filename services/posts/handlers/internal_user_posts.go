package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"social-network/services/posts/services"
	"social-network/services/posts/utils"
)

type InternalUserPostHandlers struct {
	service *services.PostService
}

func NewInternalUserPostHandlers(service *services.PostService) *InternalUserPostHandlers {
	return &InternalUserPostHandlers{service: service}
}

func (h *InternalUserPostHandlers) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/internal/v1/users/"), "/")
	if len(parts) != 2 || parts[1] != "posts" {
		utils.ErrorResponse(w, "Not found", http.StatusNotFound)
		return
	}
	userID, err := strconv.Atoi(parts[0])
	if err != nil || userID <= 0 {
		utils.ErrorResponse(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	posts, err := h.service.GetUserPosts(userID)
	if err != nil {
		log.Printf("Failed to retrieve user posts: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve posts", http.StatusInternalServerError)
		return
	}
	utils.SuccessResponse(w, map[string]any{"posts": posts, "count": len(posts)})
}
