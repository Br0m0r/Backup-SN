package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"social-network/services/users/services"
	"social-network/services/users/utils"
)

type InternalReadHandlers struct {
	service *services.UserService
}

func NewInternalReadHandlers(service *services.UserService) *InternalReadHandlers {
	return &InternalReadHandlers{service: service}
}

func (h *InternalReadHandlers) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/internal/v1/users/")
	if path == "profiles" {
		h.getProfiles(w, r)
		return
	}
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[1] != "following" {
		utils.ErrorResponse(w, "Not found", http.StatusNotFound)
		return
	}
	userID, err := strconv.Atoi(parts[0])
	if err != nil || userID <= 0 {
		utils.ErrorResponse(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	followingIDs, err := h.service.GetAcceptedFollowingIDs(userID)
	if err != nil {
		log.Printf("Failed to list accepted following IDs: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve following IDs", http.StatusInternalServerError)
		return
	}
	utils.SuccessResponse(w, map[string]any{"following_ids": followingIDs})
}

func (h *InternalReadHandlers) getProfiles(w http.ResponseWriter, r *http.Request) {
	search := strings.TrimSpace(r.URL.Query().Get("q"))
	rawIDs := strings.TrimSpace(r.URL.Query().Get("ids"))
	if search == "" && rawIDs == "" {
		utils.ErrorResponse(w, "ids or q is required", http.StatusBadRequest)
		return
	}
	userIDs := make([]int, 0)
	if rawIDs != "" {
		for _, rawID := range strings.Split(rawIDs, ",") {
			userID, err := strconv.Atoi(strings.TrimSpace(rawID))
			if err != nil || userID <= 0 {
				utils.ErrorResponse(w, "Invalid user ID", http.StatusBadRequest)
				return
			}
			userIDs = append(userIDs, userID)
			if len(userIDs) > 200 {
				utils.ErrorResponse(w, "Too many user IDs", http.StatusBadRequest)
				return
			}
		}
	}
	profiles, err := h.service.GetProfileSummaries(userIDs, search)
	if err != nil {
		log.Printf("Failed to retrieve profile summaries: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve profiles", http.StatusInternalServerError)
		return
	}
	utils.SuccessResponse(w, map[string]any{"profiles": profiles})
}
