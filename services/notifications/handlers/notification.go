package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"social-network/services/notifications/db"
	"social-network/services/notifications/middleware"
	"social-network/services/notifications/models"
	"social-network/services/notifications/utils"
	"strconv"
	"strings"
	"time"
)

// NotificationHandlers handles notification-related HTTP requests
type NotificationHandlers struct {
	database *sql.DB
	hub      *NotificationHub
}

// NewNotificationHandlers creates a new NotificationHandlers
func NewNotificationHandlers(database *sql.DB, hub *NotificationHub) *NotificationHandlers {
	return &NotificationHandlers{
		database: database,
		hub:      hub,
	}
}

// CreateNotification creates a new notification (called by other services)
func (h *NotificationHandlers) CreateNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req models.CreateNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate notification type
	validTypes := map[string]bool{
		models.TypeFollow:        true,
		models.TypeFollowRequest: true,
		models.TypeGroupInvite:   true,
		models.TypeGroupRequest:  true,
		models.TypeGroupActivity: true,
		models.TypeEvent:         true,
		models.TypeMessage:       true,
		models.TypeComment:       true,
		models.TypePost:          true,
	}

	if !validTypes[req.Type] {
		utils.SendError(w, http.StatusBadRequest, "Invalid notification type")
		return
	}

	// Create notification
	notification, err := db.CreateNotification(h.database, &req)
	if err != nil {
		log.Printf("Error creating notification: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "Failed to create notification")
		return
	}

	// Broadcast to WebSocket if user is online
	h.hub.BroadcastNotification(notification)

	utils.SendSuccess(w, notification)
}

// GetNotifications retrieves notifications for the authenticated user
func (h *NotificationHandlers) GetNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	unreadOnly := r.URL.Query().Get("unread") == "true"

	limit := 20
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	var notifications []models.Notification
	var err error

	if unreadOnly {
		notifications, err = db.GetUnreadNotifications(h.database, userID)
	} else {
		notifications, err = db.GetUserNotifications(h.database, userID, limit, offset)
	}

	if err != nil {
		log.Printf("Error fetching notifications: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "Failed to fetch notifications")
		return
	}

	utils.SendSuccess(w, map[string]interface{}{
		"notifications": notifications,
		"count":         len(notifications),
	})
}

// GetUnreadCount returns the count of unread notifications
func (h *NotificationHandlers) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	count, err := db.GetUnreadCount(h.database, userID)
	if err != nil {
		log.Printf("Error getting unread count: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "Failed to get unread count")
		return
	}

	utils.SendSuccess(w, map[string]int{"unread_count": count})
}

// MarkAsRead marks a notification as read
func (h *NotificationHandlers) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get notification ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		utils.SendError(w, http.StatusBadRequest, "Invalid notification ID")
		return
	}

	notificationID, err := strconv.Atoi(pathParts[2])
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid notification ID")
		return
	}

	err = db.MarkAsRead(h.database, notificationID, userID)
	if err != nil {
		log.Printf("Error marking notification as read: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "Failed to mark notification as read")
		return
	}

	utils.SendSuccess(w, map[string]bool{"marked": true})
}

// MarkAllAsRead marks all notifications as read
func (h *NotificationHandlers) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	err := db.MarkAllAsRead(h.database, userID)
	if err != nil {
		log.Printf("Error marking all notifications as read: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "Failed to mark all as read")
		return
	}

	utils.SendSuccess(w, map[string]bool{"marked": true})
}

// DeleteNotification deletes a notification
func (h *NotificationHandlers) DeleteNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get notification ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		utils.SendError(w, http.StatusBadRequest, "Invalid notification ID")
		return
	}

	notificationID, err := strconv.Atoi(pathParts[2])
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid notification ID")
		return
	}

	err = db.DeleteNotification(h.database, notificationID, userID)
	if err != nil {
		log.Printf("Error deleting notification: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "Failed to delete notification")
		return
	}

	utils.SendSuccess(w, map[string]bool{"deleted": true})
}

// HealthCheck returns service health status
func (h *NotificationHandlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := h.database.PingContext(ctx); err != nil {
		log.Printf("Notification database health check failed: %v", err)
		utils.SendError(w, http.StatusServiceUnavailable, "Notification database unavailable")
		return
	}
	utils.SendSuccess(w, map[string]string{
		"status":  "healthy",
		"service": "notifications",
	})
}
