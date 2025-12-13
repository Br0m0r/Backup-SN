package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"social-network/services/groups/middleware"
	"social-network/services/groups/models"
	"social-network/services/groups/services"
	"social-network/services/groups/utils"
	"strconv"
	"strings"
	"time"
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

	// Parse multipart form for image upload
	err := r.ParseMultipartForm(10 << 20) // 10MB max
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")

	if name == "" || description == "" {
		utils.SendError(w, http.StatusBadRequest, "Name and description are required")
		return
	}

	req := models.CreateGroupRequest{
		Name:        name,
		Description: &description,
	}

	// Handle optional image upload
	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()

		// Validate file type
		contentType := header.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "image/") {
			utils.SendError(w, http.StatusBadRequest, "File must be an image")
			return
		}

		// Generate unique filename
		ext := filepath.Ext(header.Filename)
		if ext == "" {
			ext = ".jpg"
		}
		filename := fmt.Sprintf("group_%d_%d%s", userID, time.Now().Unix(), ext)
		filePath := filepath.Join("uploads", "groups", filename)

		// Create directory if not exists
		dirPath := filepath.Dir(filePath)
		os.MkdirAll(dirPath, os.ModePerm)

		// Save file
		dst, err := os.Create(filePath)
		if err != nil {
			log.Printf("Error creating file: %v", err)
			utils.SendError(w, http.StatusInternalServerError, "Failed to save image")
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			log.Printf("Error saving file: %v", err)
			utils.SendError(w, http.StatusInternalServerError, "Failed to save image")
			return
		}

		req.ImageURL = &filePath
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

// UpdateGroupImage handles PUT /groups/:id/image (owner only)
func (h *GroupHandlers) UpdateGroupImage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.SendError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Extract group ID from URL
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		utils.SendError(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	groupID, err := strconv.Atoi(parts[1])
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	// Check if user is the group creator
	group, err := h.service.GetGroup(groupID, userID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "Group not found")
		return
	}

	if group.CreatorID != userID {
		utils.SendError(w, http.StatusForbidden, "Only the group creator can update the image")
		return
	}

	// Parse multipart form
	err = r.ParseMultipartForm(10 << 20) // 10MB max
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	// Handle image upload
	file, header, err := r.FormFile("image")
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Image file is required")
		return
	}
	defer file.Close()

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		utils.SendError(w, http.StatusBadRequest, "File must be an image")
		return
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	filename := fmt.Sprintf("group_%d_%d%s", groupID, time.Now().Unix(), ext)
	filePath := filepath.Join("uploads", "groups", filename)

	// Create directory if not exists
	dirPath := filepath.Dir(filePath)
	os.MkdirAll(dirPath, os.ModePerm)

	// Save file
	dst, err := os.Create(filePath)
	if err != nil {
		log.Printf("Error creating file: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "Failed to save image")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		log.Printf("Error saving file: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "Failed to save image")
		return
	}

	// Delete old image if exists
	if group.ImageURL != nil && *group.ImageURL != "" {
		os.Remove(*group.ImageURL)
	}

	// Update group image in database
	err = h.service.UpdateGroupImage(groupID, filePath)
	if err != nil {
		log.Printf("Error updating group image: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "Failed to update group image")
		return
	}

	utils.SendJSON(w, http.StatusOK, map[string]string{
		"message":   "Group image updated successfully",
		"image_url": filePath,
	})
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

// GetMyInvitations handles GET /invitations
func (h *GroupHandlers) GetMyInvitations(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.SendError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	invitations, err := h.service.GetUserInvitations(userID)
	if err != nil {
		log.Printf("Error fetching invitations: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "Failed to fetch invitations")
		return
	}

	utils.SendJSON(w, http.StatusOK, invitations)
}

// RespondToInvitation handles POST /invitations/:id/respond
func (h *GroupHandlers) RespondToInvitation(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.SendError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Extract invitation ID from path
	path := strings.TrimPrefix(r.URL.Path, "/invitations/")
	path = strings.TrimSuffix(path, "/respond")
	invitationID, err := strconv.Atoi(path)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid invitation ID")
		return
	}

	var req models.RespondToInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.RespondToInvitation(invitationID, userID, req.Accept); err != nil {
		log.Printf("Error responding to invitation: %v", err)
		utils.SendError(w, http.StatusForbidden, err.Error())
		return
	}

	message := "Invitation declined"
	if req.Accept {
		message = "Invitation accepted"
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
