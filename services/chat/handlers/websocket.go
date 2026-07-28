package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"social-network/services/chat/db"
	"social-network/services/chat/groupsclient"
	"social-network/services/chat/middleware"
	"social-network/services/chat/models"
	"social-network/services/chat/usersclient"
	"social-network/services/chat/utils"
	"social-network/services/common/notify"
	"social-network/services/common/realtime"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Hub manages active WebSocket connections
type Hub struct {
	clients       map[int]map[*Client]struct{}
	broadcast     chan *models.WebSocketMessage
	publish       chan []byte
	register      chan *Client
	unregister    chan *Client
	mu            sync.RWMutex
	database      *sql.DB
	userDirectory usersclient.Directory
	membership    groupsclient.Membership
	realtime      realtime.Transport
	upgrader      websocket.Upgrader
}

// Client represents a WebSocket client connection
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	userID   int
	username string
}

// NewHub creates a new Hub instance
func NewHub(database *sql.DB, userDirectory usersclient.Directory, checkOrigin func(*http.Request) bool, transport realtime.Transport, membership groupsclient.Membership) *Hub {
	return &Hub{
		clients:       make(map[int]map[*Client]struct{}),
		broadcast:     make(chan *models.WebSocketMessage, 256),
		publish:       make(chan []byte, 256),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		database:      database,
		userDirectory: userDirectory,
		membership:    membership,
		realtime:      transport,
		upgrader: websocket.Upgrader{
			CheckOrigin: checkOrigin,
		},
	}
}

// Run starts the hub's main loop and Redis workers.
func (h *Hub) Run(ctx context.Context) {
	if h.realtime != nil {
		go h.runPublisher(ctx)
		go h.runSubscriber(ctx)
		go h.runPresenceHeartbeat(ctx)
	}
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.register:
			h.mu.Lock()
			connections := h.clients[client.userID]
			if connections == nil {
				connections = make(map[*Client]struct{})
				h.clients[client.userID] = connections
			}
			wasOffline := len(connections) == 0
			connections[client] = struct{}{}
			h.mu.Unlock()
			if wasOffline {
				h.markOnline(ctx, client.userID)
			}
			log.Printf("Client registered: user %d", client.userID)

		case client := <-h.unregister:
			h.mu.Lock()
			removed, nowOffline := h.removeClientLocked(client)
			h.mu.Unlock()
			if removed {
				log.Printf("Client unregistered: user %d", client.userID)
			}
			if nowOffline {
				h.markOffline(ctx, client.userID)
			}

		case message := <-h.broadcast:
			h.deliverLocal(ctx, message)
		}
	}
}

// Broadcast delivers locally and publishes to other Chat replicas.
func (h *Hub) Broadcast(message *models.WebSocketMessage) {
	copyOfMessage := *message
	h.broadcast <- &copyOfMessage
	if h.realtime == nil {
		return
	}
	data, err := json.Marshal(&copyOfMessage)
	if err != nil {
		log.Printf("Error marshaling Chat realtime message: %v", err)
		return
	}
	select {
	case h.publish <- data:
	default:
		log.Printf("Chat realtime publish queue is full; clients must recover from persisted history")
	}
}

// IsUserOnline checks if a user is currently connected
func (h *Hub) IsUserOnline(userID int) bool {
	h.mu.RLock()
	local := len(h.clients[userID]) > 0
	h.mu.RUnlock()
	if local || h.realtime == nil {
		return local
	}
	online, err := h.realtime.IsOnline(context.Background(), userID)
	if err != nil {
		log.Printf("Failed to read Chat presence for user %d: %v", userID, err)
		return false
	}
	return online
}

func (h *Hub) deliverLocal(ctx context.Context, message *models.WebSocketMessage) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling Chat message: %v", err)
		return
	}
	recipients := []int{message.ReceiverID}
	if message.Type == "group_message" {
		recipients, err = h.membership.MemberIDs(ctx, message.GroupID)
		if err != nil {
			log.Printf("Error getting group members: %v", err)
			return
		}
	}

	var offline []int
	h.mu.Lock()
	for _, userID := range recipients {
		for client := range h.clients[userID] {
			select {
			case client.send <- data:
			default:
				_, nowOffline := h.removeClientLocked(client)
				if nowOffline {
					offline = append(offline, userID)
				}
			}
		}
	}
	h.mu.Unlock()
	for _, userID := range offline {
		h.markOffline(ctx, userID)
	}
}

func (h *Hub) removeClientLocked(client *Client) (bool, bool) {
	connections := h.clients[client.userID]
	if _, ok := connections[client]; !ok {
		return false, false
	}
	delete(connections, client)
	close(client.send)
	if len(connections) == 0 {
		delete(h.clients, client.userID)
		return true, true
	}
	return true, false
}

func (h *Hub) runPublisher(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-h.publish:
			if err := h.realtime.Publish(ctx, data); err != nil {
				log.Printf("Failed to publish Chat realtime message: %v", err)
			}
		}
	}
}

func (h *Hub) runSubscriber(ctx context.Context) {
	delay := time.Second
	for ctx.Err() == nil {
		err := h.realtime.Subscribe(ctx, func(data []byte) {
			var message models.WebSocketMessage
			if err := json.Unmarshal(data, &message); err != nil {
				log.Printf("Ignoring malformed Chat realtime message: %v", err)
				return
			}
			select {
			case h.broadcast <- &message:
			case <-ctx.Done():
			}
		})
		if err == nil || ctx.Err() != nil {
			return
		}
		log.Printf("Chat realtime subscription interrupted: %v; retrying in %s", err, delay)
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}
		if delay < 30*time.Second {
			delay *= 2
		}
	}
}

func (h *Hub) runPresenceHeartbeat(ctx context.Context) {
	ticker := time.NewTicker(h.realtime.PresenceRefreshInterval())
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.mu.RLock()
			userIDs := make([]int, 0, len(h.clients))
			for userID := range h.clients {
				userIDs = append(userIDs, userID)
			}
			h.mu.RUnlock()
			for _, userID := range userIDs {
				h.markOnline(ctx, userID)
			}
		}
	}
}

func (h *Hub) markOnline(ctx context.Context, userID int) {
	if h.realtime != nil {
		if err := h.realtime.MarkOnline(ctx, userID); err != nil {
			log.Printf("Failed to refresh Chat presence for user %d: %v", userID, err)
		}
	}
}

func (h *Hub) markOffline(ctx context.Context, userID int) {
	if h.realtime != nil {
		if err := h.realtime.MarkOffline(ctx, userID); err != nil {
			log.Printf("Failed to clear Chat presence for user %d: %v", userID, err)
		}
	}
}

// HandleWebSocket handles WebSocket connection upgrades
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Get username
	username := r.URL.Query().Get("username")
	if username == "" {
		username = "User"
	}

	// Create new client
	client := &Client{
		hub:      h,
		conn:     conn,
		send:     make(chan []byte, 256),
		userID:   userID,
		username: username,
	}

	// Register client
	h.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, messageData, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Parse incoming message
		var wsMsg models.WebSocketMessage
		if err := json.Unmarshal(messageData, &wsMsg); err != nil {
			log.Printf("Error parsing message: %v", err)
			continue
		}

		// Set sender ID from authenticated user
		wsMsg.SenderID = c.userID
		wsMsg.Timestamp = time.Now()

		// Handle different message types
		switch wsMsg.Type {
		case "message":
			c.handleChatMessage(&wsMsg)
		case "group_message":
			c.handleGroupChatMessage(&wsMsg)
		case "read":
			c.handleReadReceipt(&wsMsg)
		case "typing":
			c.handleTypingIndicator(&wsMsg)
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to current write
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleChatMessage processes incoming chat messages
func (c *Client) handleChatMessage(wsMsg *models.WebSocketMessage) {
	// Validate message content (allow empty if image is provided)
	allowEmpty := wsMsg.ImagePath != nil && *wsMsg.ImagePath != ""
	sanitizedContent, err := utils.ValidateMessageContent(wsMsg.Content, allowEmpty)
	if err != nil {
		log.Printf("Message validation failed: %v", err)
		c.sendError(err.Error())
		return
	}

	// Check if sender can chat with receiver
	canChat, err := canChat(context.Background(), c.hub.database, c.hub.userDirectory, c.userID, wsMsg.ReceiverID)
	if err != nil {
		log.Printf("Error checking chat permission: %v", err)
		c.sendError("Failed to check chat permissions")
		return
	}

	if !canChat {
		log.Printf("User %d cannot chat with user %d", c.userID, wsMsg.ReceiverID)
		c.sendError("You cannot send messages to this user")
		return
	}

	// Save message to database with sanitized content
	msg := &models.Message{
		SenderID:   c.userID,
		ReceiverID: wsMsg.ReceiverID,
		Content:    sanitizedContent,
		ImagePath:  wsMsg.ImagePath,
		IsRead:     false,
		CreatedAt:  time.Now(),
	}

	if err := db.SaveMessage(c.hub.database, msg); err != nil {
		log.Printf("Error saving message: %v", err)
		c.sendError("Failed to save message")
		return
	}

	// Update WebSocket message with database ID and sanitized content
	wsMsg.MessageID = msg.ID
	wsMsg.Content = sanitizedContent
	wsMsg.Type = "message"

	// Broadcast to receiver if online
	c.hub.Broadcast(wsMsg)

	// Send notification if receiver is offline
	if !c.hub.IsUserOnline(wsMsg.ReceiverID) {
		notify.NewMessage(wsMsg.ReceiverID, msg.ID, c.username)
	}

	// Send confirmation back to sender
	confirmation := *wsMsg
	data, err := json.Marshal(confirmation)
	if err == nil {
		c.send <- data
	}
}

// handleReadReceipt marks messages as read
func (c *Client) handleReadReceipt(wsMsg *models.WebSocketMessage) {
	err := db.MarkAsRead(c.hub.database, wsMsg.SenderID, c.userID)
	if err != nil {
		log.Printf("Error marking messages as read: %v", err)
	}
}

// handleTypingIndicator forwards typing status to receiver
func (c *Client) handleTypingIndicator(wsMsg *models.WebSocketMessage) {
	wsMsg.Type = "typing"
	c.hub.Broadcast(wsMsg)
}

// sendError sends an error message to the client
func (c *Client) sendError(errMsg string) {
	wsMsg := models.WebSocketMessage{
		Type:      "error",
		Error:     errMsg,
		Timestamp: time.Now(),
	}
	data, err := json.Marshal(wsMsg)
	if err == nil {
		c.send <- data
	}
}

// handleGroupChatMessage processes incoming group chat messages
func (c *Client) handleGroupChatMessage(wsMsg *models.WebSocketMessage) {
	// Validate message content (allow empty if image is provided)
	allowEmpty := wsMsg.ImagePath != nil && *wsMsg.ImagePath != ""
	sanitizedContent, err := utils.ValidateMessageContent(wsMsg.Content, allowEmpty)
	if err != nil {
		log.Printf("Message validation failed: %v", err)
		c.sendError(err.Error())
		return
	}

	// Check if sender is a member of the group
	isMember, err := c.hub.membership.IsMember(context.Background(), wsMsg.GroupID, c.userID)
	if err != nil {
		log.Printf("Error checking group membership: %v", err)
		c.sendError("Failed to verify group membership")
		return
	}

	if !isMember {
		log.Printf("User %d is not a member of group %d", c.userID, wsMsg.GroupID)
		c.sendError("You are not a member of this group")
		return
	}

	// Save message to database with sanitized content
	msg := &models.GroupMessage{
		GroupID:   wsMsg.GroupID,
		SenderID:  c.userID,
		Content:   sanitizedContent,
		CreatedAt: time.Now(),
	}

	if err := db.SaveGroupMessage(c.hub.database, msg); err != nil {
		log.Printf("Error saving group message: %v", err)
		c.sendError("Failed to save message")
		return
	}

	// Update WebSocket message with database ID and sanitized content
	wsMsg.MessageID = msg.ID
	wsMsg.Content = sanitizedContent
	wsMsg.Type = "group_message"
	wsMsg.SenderID = c.userID
	wsMsg.Timestamp = msg.CreatedAt

	// Broadcast to all group members (including sender)
	c.hub.Broadcast(wsMsg)

	// Get offline members for notifications
	groupMembers, err := c.hub.membership.MemberIDs(context.Background(), wsMsg.GroupID)
	if err == nil {
		var offlineMemberIDs []int
		for _, memberID := range groupMembers {
			// Only notify offline members (excluding sender)
			if memberID != c.userID && !c.hub.IsUserOnline(memberID) {
				offlineMemberIDs = append(offlineMemberIDs, memberID)
			}
		}
		if len(offlineMemberIDs) > 0 {
			// Use generic group name since chat service doesn't have access to group details
			notify.NewGroupMessage(offlineMemberIDs, msg.ID, c.userID, c.username, "group chat")
		}
	}
}
