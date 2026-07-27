package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"social-network/services/common/realtime"
	"social-network/services/notifications/middleware"
	"social-network/services/notifications/models"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// NotificationHub manages WebSocket connections for notifications
type NotificationHub struct {
	clients    map[int]map[*NotificationClient]struct{}
	broadcast  chan *models.Notification
	publish    chan []byte
	register   chan *NotificationClient
	unregister chan *NotificationClient
	mu         sync.RWMutex
	database   *sql.DB
	realtime   realtime.Transport
	upgrader   websocket.Upgrader
}

// NotificationClient represents a WebSocket client
type NotificationClient struct {
	hub    *NotificationHub
	conn   *websocket.Conn
	send   chan []byte
	userID int
}

// NewNotificationHub creates a new NotificationHub
func NewNotificationHub(database *sql.DB, checkOrigin func(*http.Request) bool, transport realtime.Transport) *NotificationHub {
	return &NotificationHub{
		clients:    make(map[int]map[*NotificationClient]struct{}),
		broadcast:  make(chan *models.Notification, 256),
		publish:    make(chan []byte, 256),
		register:   make(chan *NotificationClient),
		unregister: make(chan *NotificationClient),
		database:   database,
		realtime:   transport,
		upgrader: websocket.Upgrader{
			CheckOrigin: checkOrigin,
		},
	}
}

// Run starts the hub's main loop and Redis workers.
func (h *NotificationHub) Run(ctx context.Context) {
	if h.realtime != nil {
		go h.runPublisher(ctx)
		go h.runSubscriber(ctx)
	}
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.register:
			h.mu.Lock()
			connections := h.clients[client.userID]
			if connections == nil {
				connections = make(map[*NotificationClient]struct{})
				h.clients[client.userID] = connections
			}
			connections[client] = struct{}{}
			h.mu.Unlock()
			log.Printf("Client registered for notifications: user %d", client.userID)

		case client := <-h.unregister:
			h.mu.Lock()
			if h.removeClientLocked(client) {
				log.Printf("Client unregistered from notifications: user %d", client.userID)
			}
			h.mu.Unlock()

		case notification := <-h.broadcast:
			wsNotif := models.WebSocketNotification{
				Type:         "notification",
				Notification: *notification,
			}
			data, err := json.Marshal(wsNotif)
			if err != nil {
				log.Printf("Error marshaling WebSocket notification: %v", err)
				continue
			}
			h.mu.Lock()
			for client := range h.clients[notification.UserID] {
				select {
				case client.send <- data:
				default:
					h.removeClientLocked(client)
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *NotificationHub) removeClientLocked(client *NotificationClient) bool {
	connections := h.clients[client.userID]
	if _, ok := connections[client]; !ok {
		return false
	}
	delete(connections, client)
	close(client.send)
	if len(connections) == 0 {
		delete(h.clients, client.userID)
	}
	return true
}

func (h *NotificationHub) runPublisher(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-h.publish:
			if err := h.realtime.Publish(ctx, data); err != nil {
				log.Printf("Failed to publish notification realtime message: %v", err)
			}
		}
	}
}

func (h *NotificationHub) runSubscriber(ctx context.Context) {
	delay := time.Second
	for ctx.Err() == nil {
		err := h.realtime.Subscribe(ctx, func(data []byte) {
			var notification models.Notification
			if err := json.Unmarshal(data, &notification); err != nil {
				log.Printf("Ignoring malformed notification realtime message: %v", err)
				return
			}
			select {
			case h.broadcast <- &notification:
			case <-ctx.Done():
			}
		})
		if err == nil || ctx.Err() != nil {
			return
		}
		log.Printf("Notification realtime subscription interrupted: %v; retrying in %s", err, delay)
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

// BroadcastNotification sends a notification locally and to other replicas.
func (h *NotificationHub) BroadcastNotification(notification *models.Notification) {
	copyOfNotification := *notification
	h.broadcast <- &copyOfNotification
	if h.realtime == nil {
		return
	}
	data, err := json.Marshal(&copyOfNotification)
	if err != nil {
		log.Printf("Error marshaling notification realtime message: %v", err)
		return
	}
	select {
	case h.publish <- data:
	default:
		log.Printf("Notification realtime publish queue is full; clients must recover from persisted state")
	}
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
	conn, err := h.upgrader.Upgrade(w, r, nil)
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
