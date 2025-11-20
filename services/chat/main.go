package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"

	"social-network/services/chat/handlers"
	"social-network/services/chat/middleware"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	log.Println("Starting Chat Service...")

	// Get database path from environment or use default
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./social_network.db"
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

	// Create WebSocket hub
	hub := handlers.NewHub(database)
	go hub.Run()

	// Create handlers
	chatHandlers := handlers.NewChatHandlers(database, hub)
	uploadHandlers := handlers.NewUploadHandlers()

	// Create auth middleware
	authMiddleware := middleware.AuthMiddleware(authServiceURL)

	// Setup routes
	mux := http.NewServeMux()

	// Health check (no auth required)
	mux.HandleFunc("/health", chatHandlers.HealthCheck)

	// WebSocket endpoint (auth required via query param or header)
	mux.Handle("/ws", authMiddleware(http.HandlerFunc(hub.HandleWebSocket)))

	// REST endpoints (auth required)
	mux.Handle("/chat/conversations", authMiddleware(http.HandlerFunc(chatHandlers.GetConversations)))
	mux.Handle("/chat/contacts", authMiddleware(http.HandlerFunc(chatHandlers.GetAvailableContacts)))
	mux.Handle("/chat/history/", authMiddleware(http.HandlerFunc(chatHandlers.GetChatHistory)))
	mux.Handle("/chat/read/", authMiddleware(http.HandlerFunc(chatHandlers.MarkAsRead)))
	mux.Handle("/chat/unread", authMiddleware(http.HandlerFunc(chatHandlers.GetUnreadCount)))
	mux.Handle("/chat/send", authMiddleware(http.HandlerFunc(chatHandlers.SendMessage)))

	// Upload endpoints (auth required)
	mux.Handle("/upload/image", authMiddleware(http.HandlerFunc(uploadHandlers.UploadImage)))
	mux.Handle("/upload/delete", authMiddleware(http.HandlerFunc(uploadHandlers.DeleteImage)))

	// Group chat endpoints (auth required)
	mux.Handle("/chat/groups/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Route based on path pattern
		path := r.URL.Path

		if strings.HasSuffix(path, "/history") {
			chatHandlers.GetGroupChatHistory(w, r)
		} else if strings.HasSuffix(path, "/messages") {
			if r.Method == "POST" {
				chatHandlers.SendGroupMessage(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	})))

	// Apply common middleware (CORS and Logging)
	handler := middleware.CORS(
		middleware.Logging(mux),
	)

	// Start server
	log.Println("Chat Service starting on port :8085")
	log.Fatal(http.ListenAndServe(":8085", handler))
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

	// Set journal mode to DELETE for simplicity
	_, err = db.Exec("PRAGMA journal_mode=DELETE;")
	if err != nil {
		return nil, err
	}

	return db, nil
}
