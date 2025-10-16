package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"social-network/services/chat/db"
	"social-network/services/chat/middleware"
	"social-network/services/chat/models"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// Hub manages active WebSocket connections
type Hub struct {
	clients    map[int]*Client // userID -> Client
	broadcast  chan *models.WebSocketMessage
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	database   *sql.DB
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
func NewHub(database *sql.DB) *Hub {
	return &Hub{
		clients:    make(map[int]*Client),
		broadcast:  make(chan *models.WebSocketMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		database:   database,
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.userID] = client
			h.mu.Unlock()
			log.Printf("Client registered: user %d", client.userID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.userID]; ok {
				delete(h.clients, client.userID)
				close(client.send)
				log.Printf("Client unregistered: user %d", client.userID)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()

			if message.Type == "group_message" {
				// Broadcast to all group members
				members, err := db.GetGroupMembers(h.database, message.GroupID)
				if err != nil {
					log.Printf("Error getting group members: %v", err)
					h.mu.RUnlock()
					continue
				}

				data, err := json.Marshal(message)
				if err != nil {
					log.Printf("Error marshaling group message: %v", err)
					h.mu.RUnlock()
					continue
				}

				// Send to all online group members
				for _, memberID := range members {
					if client, ok := h.clients[memberID]; ok {
						select {
						case client.send <- data:
						default:
							close(client.send)
							delete(h.clients, memberID)
						}
					}
				}
			} else {
				// Send to receiver if online (1-on-1 chat)
				if client, ok := h.clients[message.ReceiverID]; ok {
					data, err := json.Marshal(message)
					if err == nil {
						select {
						case client.send <- data:
						default:
							close(client.send)
							delete(h.clients, client.userID)
						}
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// IsUserOnline checks if a user is currently connected
func (h *Hub) IsUserOnline(userID int) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[userID]
	return ok
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
	conn, err := upgrader.Upgrade(w, r, nil)
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
	// Check if sender can chat with receiver
	canChat, err := db.CanChat(c.hub.database, c.userID, wsMsg.ReceiverID)
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

	// Save message to database
	msg := &models.Message{
		SenderID:   c.userID,
		ReceiverID: wsMsg.ReceiverID,
		Content:    wsMsg.Content,
		IsRead:     false,
		CreatedAt:  time.Now(),
	}

	if err := db.SaveMessage(c.hub.database, msg); err != nil {
		log.Printf("Error saving message: %v", err)
		c.sendError("Failed to save message")
		return
	}

	// Update WebSocket message with database ID
	wsMsg.MessageID = msg.ID
	wsMsg.Type = "message"

	// Broadcast to receiver if online
	c.hub.broadcast <- wsMsg

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
	c.hub.broadcast <- wsMsg
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
	// Check if sender is a member of the group
	isMember, err := db.IsGroupMember(c.hub.database, wsMsg.GroupID, c.userID)
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

	// Save message to database
	msg := &models.GroupMessage{
		GroupID:   wsMsg.GroupID,
		SenderID:  c.userID,
		Content:   wsMsg.Content,
		CreatedAt: time.Now(),
	}

	if err := db.SaveGroupMessage(c.hub.database, msg); err != nil {
		log.Printf("Error saving group message: %v", err)
		c.sendError("Failed to save message")
		return
	}

	// Update WebSocket message with database ID
	wsMsg.MessageID = msg.ID
	wsMsg.Type = "group_message"
	wsMsg.SenderID = c.userID
	wsMsg.Timestamp = msg.CreatedAt

	// Broadcast to all group members (including sender)
	c.hub.broadcast <- wsMsg
}
