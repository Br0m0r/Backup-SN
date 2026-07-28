package handlers

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"social-network/services/common/objectstore"
	"social-network/services/groups/middleware"
	"social-network/services/groups/models"
	"social-network/services/groups/services"
	"social-network/services/groups/utils"
	"strconv"
	"strings"
)

type GroupHandlers struct {
	service *services.GroupService
	store   objectstore.Store
}

func NewGroupHandlers(service *services.GroupService, store objectstore.Store) *GroupHandlers {
	return &GroupHandlers{service: service, store: store}
}

const maxGroupImageUploadSize = 5 << 20 // 5 MiB

var errInvalidGroupImage = errors.New("invalid group image")

// HealthCheck reports whether the service-owned database is reachable.
func (h *GroupHandlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Ping(r.Context()); err != nil {
		utils.SendError(w, http.StatusServiceUnavailable, "Database unavailable")
		return
	}
	utils.SendJSON(w, http.StatusOK, map[string]string{"status": "healthy", "service": "group-service"})
}

// CreateGroup handles POST /groups
func (h *GroupHandlers) CreateGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.SendError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse multipart form for image upload.
	r.Body = http.MaxBytesReader(w, r.Body, maxGroupImageUploadSize+(1<<20))
	err := r.ParseMultipartForm(maxGroupImageUploadSize)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "File too large or invalid form data (max 5MB)")
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
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
	file, _, err := r.FormFile("image")
	var uploadedKey string
	if err == nil {
		defer file.Close()
		imageURL, key, err := h.storeGroupImage(r, file, userID)
		if err != nil {
			if errors.Is(err, errInvalidGroupImage) {
				utils.SendError(w, http.StatusBadRequest, "Invalid image. Only JPEG, PNG, GIF, and WebP up to 5MB are allowed")
				return
			}
			log.Printf("Failed to store group image: %v", err)
			utils.SendError(w, http.StatusInternalServerError, "Failed to save image")
			return
		}
		uploadedKey = key
		req.ImageURL = &imageURL
	} else if !errors.Is(err, http.ErrMissingFile) {
		utils.SendError(w, http.StatusBadRequest, "Invalid image upload")
		return
	}

	group, err := h.service.CreateGroup(&req, userID)
	if err != nil {
		if uploadedKey != "" {
			_ = h.store.Delete(r.Context(), uploadedKey)
		}
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

	// Parse multipart form.
	r.Body = http.MaxBytesReader(w, r.Body, maxGroupImageUploadSize+(1<<20))
	err = r.ParseMultipartForm(maxGroupImageUploadSize)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "File too large or invalid form data (max 5MB)")
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	// Handle image upload
	file, _, err := r.FormFile("image")
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Image file is required")
		return
	}
	defer file.Close()

	imageURL, key, err := h.storeGroupImage(r, file, userID)
	if err != nil {
		if errors.Is(err, errInvalidGroupImage) {
			utils.SendError(w, http.StatusBadRequest, "Invalid image. Only JPEG, PNG, GIF, and WebP up to 5MB are allowed")
			return
		}
		log.Printf("Failed to store group image: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "Failed to save image")
		return
	}

	// Update group image in database
	err = h.service.UpdateGroupImage(groupID, imageURL)
	if err != nil {
		_ = h.store.Delete(r.Context(), key)
		log.Printf("Error updating group image: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "Failed to update group image")
		return
	}
	if group.ImageURL != nil && *group.ImageURL != "" {
		if previousKey, err := h.store.KeyFromURL(*group.ImageURL); err == nil && strings.HasPrefix(previousKey, fmt.Sprintf("groups/users/%d/", userID)) {
			if err := h.store.Delete(r.Context(), previousKey); err != nil {
				log.Printf("Failed to remove superseded group image %q: %v", previousKey, err)
			}
		}
	}

	utils.SendJSON(w, http.StatusOK, map[string]string{
		"message":   "Group image updated successfully",
		"image_url": imageURL,
	})
}

func (h *GroupHandlers) storeGroupImage(r *http.Request, reader io.Reader, creatorID int) (string, string, error) {
	contents, err := io.ReadAll(io.LimitReader(reader, maxGroupImageUploadSize+1))
	if err != nil {
		return "", "", err
	}
	if len(contents) == 0 || len(contents) > maxGroupImageUploadSize {
		return "", "", fmt.Errorf("%w: invalid size", errInvalidGroupImage)
	}
	contentType := http.DetectContentType(contents)
	extension, ok := groupImageExtension(contentType)
	if !ok {
		return "", "", fmt.Errorf("%w: unsupported type %q", errInvalidGroupImage, contentType)
	}
	objectName, err := randomGroupObjectName()
	if err != nil {
		return "", "", err
	}
	key := fmt.Sprintf("groups/users/%d/%s%s", creatorID, objectName, extension)
	if err := h.store.Put(r.Context(), key, bytes.NewReader(contents), int64(len(contents)), contentType); err != nil {
		return "", "", err
	}
	return h.store.URL(key), key, nil
}

func randomGroupObjectName() (string, error) {
	var random [16]byte
	if _, err := rand.Read(random[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(random[:]), nil
}

func groupImageExtension(contentType string) (string, bool) {
	switch contentType {
	case "image/jpeg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	case "image/gif":
		return ".gif", true
	case "image/webp":
		return ".webp", true
	default:
		return "", false
	}
}

// GetGroups handles GET /groups
func (h *GroupHandlers) GetGroups(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.SendError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(query) > 100 {
		utils.SendError(w, http.StatusBadRequest, "Search query must be 100 characters or fewer")
		return
	}

	var groups []*models.GroupWithDetails
	var err error
	if query == "" {
		groups, err = h.service.GetAllGroups(userID)
	} else {
		groups, err = h.service.SearchGroups(userID, query)
	}
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
