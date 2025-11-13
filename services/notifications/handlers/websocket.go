package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"social-network/services/notifications/middleware"
	"social-network/services/notifications/models"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// NotificationHub manages WebSocket connections for notifications
type NotificationHub struct {
	clients    map[int]*NotificationClient // userID -> Client
	broadcast  chan *models.Notification
	register   chan *NotificationClient
	unregister chan *NotificationClient
	mu         sync.RWMutex
	database   *sql.DB
}

// NotificationClient represents a WebSocket client
type NotificationClient struct {
	hub    *NotificationHub
	conn   *websocket.Conn
	send   chan []byte
	userID int
}

// NewNotificationHub creates a new NotificationHub
func NewNotificationHub(database *sql.DB) *NotificationHub {
	return &NotificationHub{
		clients:    make(map[int]*NotificationClient),
		broadcast:  make(chan *models.Notification, 256),
		register:   make(chan *NotificationClient),
		unregister: make(chan *NotificationClient),
		database:   database,
	}
}

// Run starts the hub's main loop
func (h *NotificationHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.userID] = client
			h.mu.Unlock()
			log.Printf("Client registered for notifications: user %d", client.userID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.userID]; ok {
				delete(h.clients, client.userID)
				close(client.send)
				log.Printf("Client unregistered from notifications: user %d", client.userID)
			}
			h.mu.Unlock()

		case notification := <-h.broadcast:
			h.mu.RLock()
			if client, ok := h.clients[notification.UserID]; ok {
				wsNotif := models.WebSocketNotification{
					Type:         "notification",
					Notification: *notification,
				}
				data, err := json.Marshal(wsNotif)
				if err == nil {
					select {
					case client.send <- data:
					default:
						close(client.send)
						delete(h.clients, client.userID)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastNotification sends a notification to a specific user if online
func (h *NotificationHub) BroadcastNotification(notification *models.Notification) {
	h.broadcast <- notification
}

// HandleWebSocket handles WebSocket connections for notifications
func (h *NotificationHub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
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

	// Create new client
	client := &NotificationClient{
		hub:    h,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: userID,
	}

	// Register client
	h.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// readPump handles incoming messages from WebSocket
func (c *NotificationClient) readPump() {
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
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
	}
}

// writePump sends messages to WebSocket
func (c *NotificationClient) writePump() {
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

			// Add queued messages
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
