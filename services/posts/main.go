package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"

	"social-network/services/posts/handlers"
	"social-network/services/posts/middleware"
	"social-network/services/posts/services"
)

// OpenDB opens a connection to the SQLite database
func OpenDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Set connection pool settings for better performance
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return db, nil
}

func main() {
	// Get database path from environment variable or use default
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "/app/social_network.db"
	}

	// Open database connection
	database, err := OpenDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()
	log.Printf("Connected to database: %s", dbPath)

	// Get auth service URL from environment variable
	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	if authServiceURL == "" {
		authServiceURL = "http://auth-service:8081"
	}

	// Initialize services
	postService := services.NewPostService(database)

	// Initialize handlers
	postHandlers := handlers.NewPostHandlers(postService)

	// Initialize middleware
	rateLimiter := middleware.NewRateLimiter()
	authMiddleware := middleware.AuthMiddleware(authServiceURL)

	// Setup routes
	mux := http.NewServeMux()

	// Health check (no auth, no rate limiting)
	mux.HandleFunc("/health", handlers.HealthHandler)

	// Post endpoints
	mux.Handle("/posts", authMiddleware(rateLimiter.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			postHandlers.CreatePost(w, r)
		} else if r.Method == "GET" {
			postHandlers.GetFeed(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))))

	mux.Handle("/posts/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			postHandlers.GetPost(w, r)
		} else if r.Method == "PUT" {
			postHandlers.UpdatePost(w, r)
		} else if r.Method == "DELETE" {
			postHandlers.DeletePost(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Comment endpoints
	mux.Handle("/comments", authMiddleware(rateLimiter.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			postHandlers.CreateComment(w, r)
		} else if r.Method == "GET" {
			postHandlers.GetComments(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))))

	// Apply common middleware
	handler := middleware.CORS(
		middleware.Logging(mux),
	)

	// Start server
	log.Println("Post Service starting on port :8083")
	log.Fatal(http.ListenAndServe(":8083", handler))
}
