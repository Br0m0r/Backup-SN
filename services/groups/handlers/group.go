package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"social-network/services/groups/middleware"
	"social-network/services/groups/models"
	"social-network/services/groups/services"
	"social-network/services/groups/utils"
	"strconv"
	"strings"
)

type GroupHandlers struct {
	service *services.GroupService
}

func NewGroupHandlers(service *services.GroupService) *GroupHandlers {
	return &GroupHandlers{service: service}
}

// HealthHandler handles GET /health
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	utils.SendJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

// CreateGroup handles POST /groups
func (h *GroupHandlers) CreateGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.SendError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req models.CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	group, err := h.service.CreateGroup(&req, userID)
	if err != nil {
		log.Printf("Error creating group: %v", err)
		// Check for UNIQUE constraint violation
		if strings.Contains(err.Error(), "UNIQUE constraint failed: groups.name") {
			utils.SendError(w, http.StatusConflict, "A group with this name already exists")
			return
		}
		utils.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SendJSON(w, http.StatusCreated, group)
}

// GetGroups handles GET /groups
func (h *GroupHandlers) GetGroups(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.SendError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	groups, err := h.service.GetAllGroups(userID)
	if err != nil {
		log.Printf("Error fetching groups: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "Failed to fetch groups")
		return
	}

	utils.SendJSON(w, http.StatusOK, groups)
}

// GetGroup handles GET /groups/:id
func (h *GroupHandlers) GetGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.SendError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Extract group ID from path
	path := strings.TrimPrefix(r.URL.Path, "/groups/")
	groupID, err := strconv.Atoi(path)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	group, err := h.service.GetGroup(groupID, userID)
	if err != nil {
		log.Printf("Error fetching group: %v", err)
		utils.SendError(w, http.StatusNotFound, "Group not found")
		return
	}

	utils.SendJSON(w, http.StatusOK, group)
}

// UpdateGroup handles PUT /groups/:id
func (h *GroupHandlers) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.SendError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Extract group ID from path
	path := strings.TrimPrefix(r.URL.Path, "/groups/")
	groupID, err := strconv.Atoi(path)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	var req models.UpdateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.UpdateGroup(groupID, userID, &req); err != nil {
		log.Printf("Error updating group: %v", err)
		utils.SendError(w, http.StatusForbidden, err.Error())
		return
	}

	utils.SendJSON(w, http.StatusOK, map[string]string{"message": "Group updated successfully"})
}

// InviteMember handles POST /groups/:id/invite
func (h *GroupHandlers) InviteMember(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.SendError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	username, ok := middleware.GetUsernameFromContext(r)
	if !ok {
		utils.SendError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Extract group ID from path
	path := strings.TrimPrefix(r.URL.Path, "/groups/")
	path = strings.TrimSuffix(path, "/invite")
	groupID, err := strconv.Atoi(path)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	var req models.InviteMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.InviteMember(groupID, userID, req.UserID, username); err != nil {
		log.Printf("Error inviting member: %v", err)
		utils.SendError(w, http.StatusForbidden, err.Error())
		return
	}

	utils.SendJSON(w, http.StatusOK, map[string]string{"message": "Invitation sent successfully"})
}

// RequestToJoin handles POST /groups/:id/request
func (h *GroupHandlers) RequestToJoin(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.SendError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	username, ok := middleware.GetUsernameFromContext(r)
	if !ok {
		utils.SendError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Extract group ID from path
	path := strings.TrimPrefix(r.URL.Path, "/groups/")
	path = strings.TrimSuffix(path, "/request")
	groupID, err := strconv.Atoi(path)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	if err := h.service.RequestToJoin(groupID, userID, username); err != nil {
		log.Printf("Error requesting to join: %v", err)
		utils.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SendJSON(w, http.StatusOK, map[string]string{"message": "Join request sent successfully"})
}

// GetPendingRequests handles GET /groups/:id/requests
func (h *GroupHandlers) GetPendingRequests(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.SendError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Extract group ID from path
	path := strings.TrimPrefix(r.URL.Path, "/groups/")
	path = strings.TrimSuffix(path, "/requests")
	groupID, err := strconv.Atoi(path)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	requests, err := h.service.GetPendingRequests(groupID, userID)
	if err != nil {
		log.Printf("Error fetching requests: %v", err)
		utils.SendError(w, http.StatusForbidden, err.Error())
		return
	}

	utils.SendJSON(w, http.StatusOK, requests)
}

// RespondToRequest handles POST /groups/:id/requests/respond
func (h *GroupHandlers) RespondToRequest(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.SendError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Extract group ID from path
	path := strings.TrimPrefix(r.URL.Path, "/groups/")
	path = strings.TrimSuffix(path, "/requests/respond")
	groupID, err := strconv.Atoi(path)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	var req models.RespondToRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.RespondToRequest(groupID, req.MemberID, userID, req.Accept); err != nil {
		log.Printf("Error responding to request: %v", err)
		utils.SendError(w, http.StatusForbidden, err.Error())
		return
	}

	message := "Request rejected"
	if req.Accept {
		message = "Request accepted"
	}
	utils.SendJSON(w, http.StatusOK, map[string]string{"message": message})
}

// GetMembers handles GET /groups/:id/members
func (h *GroupHandlers) GetMembers(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.SendError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Extract group ID from path
	path := strings.TrimPrefix(r.URL.Path, "/groups/")
	path = strings.TrimSuffix(path, "/members")
	groupID, err := strconv.Atoi(path)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	members, err := h.service.GetGroupMembers(groupID, userID)
	if err != nil {
		log.Printf("Error fetching members: %v", err)
		utils.SendError(w, http.StatusForbidden, err.Error())
		return
	}

	utils.SendJSON(w, http.StatusOK, members)
}

func (h *GroupHandlers) LeaveGroup(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodDelete {
		utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.SendError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}
	// Extract group ID from path
	path := strings.TrimPrefix(r.URL.Path, "/groups/")
	path = strings.TrimSuffix(path, "/leave")
	groupID, err := strconv.Atoi(path)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid group ID")
		return
	}
	if err := h.service.LeaveGroup(groupID, userID); err != nil {
		log.Printf("Error leaving group: %v", err)
		utils.SendError(w, http.StatusForbidden, err.Error())
		return
	}
	utils.SendJSON(w, http.StatusOK, map[string]string{"message": "Left group successfully"})
}
