package models

import "time"

// Message represents a chat message between users
type Message struct {
	ID         int       `json:"id"`
	SenderID   int       `json:"sender_id"`
	ReceiverID int       `json:"receiver_id"`
	Content    string    `json:"content"`
	IsRead     bool      `json:"is_read"`
	CreatedAt  time.Time `json:"created_at"`
}

// GroupMessage represents a message in a group chat
type GroupMessage struct {
	ID        int       `json:"id"`
	GroupID   int       `json:"group_id"`
	SenderID  int       `json:"sender_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateMessageRequest represents the payload for creating a message
type CreateMessageRequest struct {
	ReceiverID int    `json:"receiver_id"`
	Content    string `json:"content"`
}

// CreateGroupMessageRequest represents the payload for creating a group message
type CreateGroupMessageRequest struct {
	GroupID int    `json:"group_id"`
	Content string `json:"content"`
}

// Conversation represents a chat conversation summary
type Conversation struct {
	UserID        int       `json:"user_id"`
	Username      string    `json:"username"`
	FirstName     *string   `json:"first_name,omitempty"`
	LastName      *string   `json:"last_name,omitempty"`
	Nickname      *string   `json:"nickname,omitempty"`
	LastMessage   string    `json:"last_message"`
	LastMessageAt time.Time `json:"last_message_at"`
	UnreadCount   int       `json:"unread_count"`
	IsOnline      bool      `json:"is_online"`
}

// WebSocketMessage represents messages sent/received via WebSocket
type WebSocketMessage struct {
	Type       string    `json:"type"` // "message", "group_message", "typing", "read", "error"
	MessageID  int       `json:"message_id,omitempty"`
	SenderID   int       `json:"sender_id"`
	ReceiverID int       `json:"receiver_id,omitempty"` // For 1-on-1 chat
	GroupID    int       `json:"group_id,omitempty"`    // For group chat
	Content    string    `json:"content,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	Error      string    `json:"error,omitempty"`
}

// ChatContact represents an available chat contact
type ChatContact struct {
	UserID         int    `json:"user_id"`
	Username       string `json:"username"`
	FirstName      string `json:"first_name,omitempty"`
	LastName       string `json:"last_name,omitempty"`
	Nickname       string `json:"nickname,omitempty"`
	Avatar         string `json:"avatar,omitempty"`
	IsOnline       bool   `json:"is_online"`
	UnreadCount    int    `json:"unread_count"`
	HasChatHistory bool   `json:"has_chat_history"`
}
