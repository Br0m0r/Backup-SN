package models

import "time"

// Notification represents a user notification
type Notification struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Type      string    `json:"type"`
	RelatedID int       `json:"related_id"`
	Content   string    `json:"content"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateNotificationRequest is the request body for creating a notification
type CreateNotificationRequest struct {
	UserID    int    `json:"user_id"`
	Type      string `json:"type"`
	RelatedID int    `json:"related_id"`
	Content   string `json:"content"`
}

// NotificationTypes constants
const (
	TypeFollow        = "follow"
	TypeFollowRequest = "follow_request"
	TypeGroupInvite   = "group_invite"
	TypeGroupRequest  = "group_request"
	TypeGroupActivity = "group_activity"
	TypeEvent         = "event"
	TypeMessage       = "message"
	TypeComment       = "comment"
	TypePost          = "post"
)

// WebSocketNotification represents a notification sent via WebSocket
type WebSocketNotification struct {
	Type         string       `json:"type"`
	Notification Notification `json:"notification"`
}
