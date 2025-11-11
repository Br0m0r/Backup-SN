package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"social-network/services/notifications/handlers"
	"social-network/services/notifications/middleware"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	log.Println("Starting Notification Service...")

	// Get database path from environment
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "/app/database.db"
	}

	// Open database connection
	database, err := OpenDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	log.Printf("Connected to database at %s", dbPath)

	// Get auth service URL
	authServiceURL := middleware.GetAuthServiceURL()
	log.Printf("Auth service URL: %s", authServiceURL)

	// Create notification hub for WebSocket
	hub := handlers.NewNotificationHub(database)
	go hub.Run()

	// Create handlers
	notifHandlers := handlers.NewNotificationHandlers(database, hub)

	// Create auth middleware
	authMiddleware := middleware.AuthMiddleware(authServiceURL)

	// Setup routes
	mux := http.NewServeMux()

	// Health check (no auth required)
	mux.HandleFunc("/health", notifHandlers.HealthCheck)

	// Create notification (no auth required - called by other services)
	// In production, you'd want to secure this with API keys or service-to-service auth
	mux.HandleFunc("/notifications", notifHandlers.CreateNotification)

	// Get notifications (auth required)
	mux.Handle("/notifications/list", authMiddleware(http.HandlerFunc(notifHandlers.GetNotifications)))

	// Get unread count (auth required)
	mux.Handle("/notifications/unread-count", authMiddleware(http.HandlerFunc(notifHandlers.GetUnreadCount)))

	// Mark as read (auth required)
	mux.Handle("/notifications/read/", authMiddleware(http.HandlerFunc(notifHandlers.MarkAsRead)))

	// Mark all as read (auth required)
	mux.Handle("/notifications/read-all", authMiddleware(http.HandlerFunc(notifHandlers.MarkAllAsRead)))

	// Delete notification (auth required)
	mux.Handle("/notifications/delete/", authMiddleware(http.HandlerFunc(notifHandlers.DeleteNotification)))

	// WebSocket endpoint (auth required via query param)
	mux.Handle("/ws", authMiddleware(http.HandlerFunc(hub.HandleWebSocket)))

	// Apply common middleware (CORS and Logging)
	handler := middleware.CORS(
		middleware.Logging(mux),
	)

	// Start server
	log.Println("Notification Service starting on port :8086")
	log.Fatal(http.ListenAndServe(":8086", handler))
}

// OpenDB opens a connection to the SQLite database
func OpenDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Test the connection
	if err = db.Ping(); err != nil {
		return nil, err
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return db, nil
}
