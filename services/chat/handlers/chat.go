package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"social-network/services/chat/db"
	"social-network/services/chat/middleware"
	"social-network/services/chat/models"
	"social-network/services/chat/utils"
	"strconv"
	"strings"
	"time"
)

// ChatHandlers handles HTTP endpoints for chat
type ChatHandlers struct {
	database *sql.DB
	hub      *Hub
}

// NewChatHandlers creates a new ChatHandlers instance
func NewChatHandlers(database *sql.DB, hub *Hub) *ChatHandlers {
	return &ChatHandlers{
		database: database,
		hub:      hub,
	}
}

// GetChatHistory handles GET /chat/history/:userId
// Retrieves chat history between current user and specified user
func (h *ChatHandlers) GetChatHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract other user ID from URL
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		utils.ErrorResponse(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	otherUserID, err := strconv.Atoi(pathParts[2])
	if err != nil {
		utils.ErrorResponse(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Check if user can chat with this person
	canChat, err := db.CanChat(h.database, userID, otherUserID)
	if err != nil {
		log.Printf("Error checking chat permission: %v", err)
		utils.ErrorResponse(w, "Failed to check permissions", http.StatusInternalServerError)
		return
	}

	if !canChat {
		utils.ErrorResponse(w, "You cannot access this chat", http.StatusForbidden)
		return
	}

	// Get limit from query param (default 50)
	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 200 {
			limit = l
		}
	}

	// Retrieve chat history
	messages, err := db.GetChatHistory(h.database, userID, otherUserID, limit)
	if err != nil {
		log.Printf("Error getting chat history: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve messages", http.StatusInternalServerError)
		return
	}

	utils.SuccessResponse(w, map[string]interface{}{
		"messages": messages,
		"count":    len(messages),
	})
}

// GetConversations handles GET /chat/conversations
// Retrieves all conversations for the current user
func (h *ChatHandlers) GetConversations(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conversations, err := db.GetConversations(h.database, userID)
	if err != nil {
		log.Printf("Error getting conversations: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve conversations", http.StatusInternalServerError)
		return
	}

	// Update online status for each conversation
	for i := range conversations {
		conversations[i].IsOnline = h.hub.IsUserOnline(conversations[i].UserID)
	}

	utils.SuccessResponse(w, map[string]interface{}{
		"conversations": conversations,
		"count":         len(conversations),
	})
}

// MarkAsRead handles POST /chat/read/:userId
// Marks all messages from specified user as read
func (h *ChatHandlers) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract other user ID from URL
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		utils.ErrorResponse(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	otherUserID, err := strconv.Atoi(pathParts[2])
	if err != nil {
		utils.ErrorResponse(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	err = db.MarkAsRead(h.database, otherUserID, userID)
	if err != nil {
		log.Printf("Error marking messages as read: %v", err)
		utils.ErrorResponse(w, "Failed to mark messages as read", http.StatusInternalServerError)
		return
	}

	utils.SuccessResponse(w, map[string]interface{}{
		"message": "Messages marked as read",
	})
}

// GetUnreadCount handles GET /chat/unread
// Returns total unread message count for current user
func (h *ChatHandlers) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	count, err := db.GetUnreadCount(h.database, userID)
	if err != nil {
		log.Printf("Error getting unread count: %v", err)
		utils.ErrorResponse(w, "Failed to get unread count", http.StatusInternalServerError)
		return
	}

	utils.SuccessResponse(w, map[string]interface{}{
		"unread_count": count,
	})
}

// SendMessage handles POST /chat/send
// Sends a message via HTTP (alternative to WebSocket)
func (h *ChatHandlers) SendMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		ReceiverID int    `json:"receiver_id"`
		Content    string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Content == "" {
		utils.ErrorResponse(w, "Message content is required", http.StatusBadRequest)
		return
	}

	// Check if user can chat with receiver
	canChat, err := db.CanChat(h.database, userID, req.ReceiverID)
	if err != nil {
		log.Printf("Error checking chat permission: %v", err)
		utils.ErrorResponse(w, "Failed to check permissions", http.StatusInternalServerError)
		return
	}

	if !canChat {
		utils.ErrorResponse(w, "You cannot send messages to this user", http.StatusForbidden)
		return
	}

	// Save message
	msg := &models.Message{
		SenderID:   userID,
		ReceiverID: req.ReceiverID,
		Content:    req.Content,
		IsRead:     false,
		CreatedAt:  time.Now(),
	}

	if err := db.SaveMessage(h.database, msg); err != nil {
		log.Printf("Error saving message: %v", err)
		utils.ErrorResponse(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	// Broadcast via WebSocket if receiver is online
	wsMsg := &models.WebSocketMessage{
		Type:       "message",
		MessageID:  msg.ID,
		SenderID:   userID,
		ReceiverID: req.ReceiverID,
		Content:    req.Content,
		Timestamp:  msg.CreatedAt,
	}
	h.hub.broadcast <- wsMsg

	utils.SuccessResponse(w, map[string]interface{}{
		"message": msg,
	})
}

// HealthCheck handles GET /health
func (h *ChatHandlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, map[string]interface{}{
		"status":  "healthy",
		"service": "chat-service",
	})
}

// GetGroupChatHistory handles GET /chat/groups/:groupId/history
// Retrieves chat history for a group
func (h *ChatHandlers) GetGroupChatHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract group ID from URL
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 {
		utils.ErrorResponse(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	groupID, err := strconv.Atoi(pathParts[2])
	if err != nil {
		utils.ErrorResponse(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	// Check if user is a member
	isMember, err := db.IsGroupMember(h.database, groupID, userID)
	if err != nil {
		log.Printf("Error checking group membership: %v", err)
		utils.ErrorResponse(w, "Failed to check membership", http.StatusInternalServerError)
		return
	}

	if !isMember {
		utils.ErrorResponse(w, "You are not a member of this group", http.StatusForbidden)
		return
	}

	// Get limit from query param (default 50)
	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 200 {
			limit = l
		}
	}

	// Retrieve group chat history
	messages, err := db.GetGroupChatHistory(h.database, groupID, limit)
	if err != nil {
		log.Printf("Error getting group chat history: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve messages", http.StatusInternalServerError)
		return
	}

	utils.SuccessResponse(w, map[string]interface{}{
		"messages": messages,
		"count":    len(messages),
	})
}

// SendGroupMessage handles POST /chat/groups/:groupId/messages
// Sends a message to a group via HTTP (fallback if WebSocket fails)
func (h *ChatHandlers) SendGroupMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract group ID from URL
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 {
		utils.ErrorResponse(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	groupID, err := strconv.Atoi(pathParts[2])
	if err != nil {
		utils.ErrorResponse(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Content == "" {
		utils.ErrorResponse(w, "Message content is required", http.StatusBadRequest)
		return
	}

	// Check if user is a member
	isMember, err := db.IsGroupMember(h.database, groupID, userID)
	if err != nil {
		log.Printf("Error checking group membership: %v", err)
		utils.ErrorResponse(w, "Failed to check membership", http.StatusInternalServerError)
		return
	}

	if !isMember {
		utils.ErrorResponse(w, "You are not a member of this group", http.StatusForbidden)
		return
	}

	// Save message
	msg := &models.GroupMessage{
		GroupID:   groupID,
		SenderID:  userID,
		Content:   req.Content,
		CreatedAt: time.Now(),
	}

	if err := db.SaveGroupMessage(h.database, msg); err != nil {
		log.Printf("Error saving group message: %v", err)
		utils.ErrorResponse(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	// Broadcast via WebSocket to online group members
	wsMsg := &models.WebSocketMessage{
		Type:      "group_message",
		MessageID: msg.ID,
		SenderID:  userID,
		GroupID:   groupID,
		Content:   req.Content,
		Timestamp: msg.CreatedAt,
	}
	h.hub.broadcast <- wsMsg

	utils.SuccessResponse(w, map[string]interface{}{
		"message": msg,
	})
}
