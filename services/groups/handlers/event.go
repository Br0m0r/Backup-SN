package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"social-network/services/groups/middleware"
	"social-network/services/groups/models"
	"social-network/services/groups/utils"
	"strconv"
	"strings"
)

// CreateEvent handles POST /events
func (h *GroupHandlers) CreateEvent(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.SendError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req models.CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	event, err := h.service.CreateEvent(&req, userID)
	if err != nil {
		log.Printf("Error creating event: %v", err)
		utils.SendError(w, http.StatusForbidden, err.Error())
		return
	}

	utils.SendJSON(w, http.StatusCreated, event)
}

// GetEvent handles GET /events/:id
func (h *GroupHandlers) GetEvent(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.SendError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Extract event ID from path
	path := strings.TrimPrefix(r.URL.Path, "/events/")
	eventID, err := strconv.Atoi(path)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid event ID")
		return
	}

	event, err := h.service.GetEvent(eventID, userID)
	if err != nil {
		log.Printf("Error fetching event: %v", err)
		utils.SendError(w, http.StatusNotFound, "Event not found")
		return
	}

	utils.SendJSON(w, http.StatusOK, event)
}

// GetGroupEvents handles GET /groups/:id/events
func (h *GroupHandlers) GetGroupEvents(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.SendError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Extract group ID from path
	path := strings.TrimPrefix(r.URL.Path, "/groups/")
	path = strings.TrimSuffix(path, "/events")
	groupID, err := strconv.Atoi(path)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	events, err := h.service.GetGroupEvents(groupID, userID)
	if err != nil {
		log.Printf("Error fetching events: %v", err)
		utils.SendError(w, http.StatusForbidden, err.Error())
		return
	}

	utils.SendJSON(w, http.StatusOK, events)
}

// RespondToEvent handles POST /events/respond
func (h *GroupHandlers) RespondToEvent(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.SendError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req models.EventResponseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.RespondToEvent(&req, userID); err != nil {
		log.Printf("Error responding to event: %v", err)
		utils.SendError(w, http.StatusForbidden, err.Error())
		return
	}

	utils.SendJSON(w, http.StatusOK, map[string]string{"message": "Response recorded successfully"})
}

// CreateGroupMessage handles POST /groups/:id/messages
func (h *GroupHandlers) CreateGroupMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.SendError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Extract group ID from path
	path := strings.TrimPrefix(r.URL.Path, "/groups/")
	path = strings.TrimSuffix(path, "/messages")
	groupID, err := strconv.Atoi(path)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	message, err := h.service.CreateGroupMessage(groupID, userID, req.Content)
	if err != nil {
		log.Printf("Error creating message: %v", err)
		utils.SendError(w, http.StatusForbidden, err.Error())
		return
	}

	utils.SendJSON(w, http.StatusCreated, message)
}

// GetGroupMessages handles GET /groups/:id/messages
func (h *GroupHandlers) GetGroupMessages(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.SendError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Extract group ID from path
	path := strings.TrimPrefix(r.URL.Path, "/groups/")
	path = strings.TrimSuffix(path, "/messages")
	groupID, err := strconv.Atoi(path)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	// Get limit from query params
	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	messages, err := h.service.GetGroupMessages(groupID, userID, limit)
	if err != nil {
		log.Printf("Error fetching messages: %v", err)
		utils.SendError(w, http.StatusForbidden, err.Error())
		return
	}

	utils.SendJSON(w, http.StatusOK, messages)
}
