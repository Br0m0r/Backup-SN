package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"social-network/services/users/models"
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
	path := strings.TrimPrefix(r.URL.Path, "/internal/v1/users/")
	if path == "profiles" && r.Method == http.MethodPost {
		h.provisionProfile(w, r)
		return
	}
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if path == "profiles" {
		h.getProfiles(w, r)
		return
	}
	if path == "chat/permission" {
		h.getChatPermission(w, r)
		return
	}
	if path == "chat/contacts" {
		h.getChatContacts(w, r)
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

func (h *InternalReadHandlers) provisionProfile(w http.ResponseWriter, r *http.Request) {
	var request models.ProvisionProfileRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&request); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if request.AccountID <= 0 || strings.TrimSpace(request.Username) == "" ||
		strings.TrimSpace(request.Email) == "" {
		utils.ErrorResponse(w, "account_id, username, and email are required", http.StatusBadRequest)
		return
	}
	if err := h.service.ProvisionProfile(request); err != nil {
		log.Printf("Failed to provision user profile: %v", err)
		utils.ErrorResponse(w, "Failed to provision profile", http.StatusInternalServerError)
		return
	}
	utils.SuccessResponse(w, map[string]any{"profile_id": request.AccountID})
}

func (h *InternalReadHandlers) getChatPermission(w http.ResponseWriter, r *http.Request) {
	senderID, senderErr := strconv.Atoi(r.URL.Query().Get("sender_id"))
	receiverID, receiverErr := strconv.Atoi(r.URL.Query().Get("receiver_id"))
	if senderErr != nil || receiverErr != nil || senderID <= 0 || receiverID <= 0 {
		utils.ErrorResponse(w, "Invalid sender or receiver ID", http.StatusBadRequest)
		return
	}
	canChat, err := h.service.CanStartConversation(senderID, receiverID)
	if err != nil {
		log.Printf("Failed to evaluate direct conversation permission: %v", err)
		utils.ErrorResponse(w, "Failed to check conversation permission", http.StatusInternalServerError)
		return
	}
	utils.SuccessResponse(w, map[string]any{"can_start": canChat})
}

func (h *InternalReadHandlers) getChatContacts(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(r.URL.Query().Get("user_id"))
	if err != nil || userID <= 0 {
		utils.ErrorResponse(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	recentSenderIDs := make([]int, 0)
	if rawIDs := strings.TrimSpace(r.URL.Query().Get("recent_sender_ids")); rawIDs != "" {
		for _, rawID := range strings.Split(rawIDs, ",") {
			senderID, err := strconv.Atoi(strings.TrimSpace(rawID))
			if err != nil || senderID <= 0 {
				utils.ErrorResponse(w, "Invalid recent sender ID", http.StatusBadRequest)
				return
			}
			recentSenderIDs = append(recentSenderIDs, senderID)
			if len(recentSenderIDs) > 500 {
				utils.ErrorResponse(w, "Too many recent sender IDs", http.StatusBadRequest)
				return
			}
		}
	}
	contacts, err := h.service.GetChatContacts(userID, recentSenderIDs)
	if err != nil {
		log.Printf("Failed to retrieve Chat contacts: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve Chat contacts", http.StatusInternalServerError)
		return
	}
	utils.SuccessResponse(w, map[string]any{"contacts": contacts})
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
