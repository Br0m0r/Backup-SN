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

func main() {
	// Open database connection
	database, err := OpenDB("./social_network.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer database.Close()

	// Initialize services
	postService := services.NewPostService(database)

	// Initialize handlers
	postHandlers := handlers.NewPostHandlers(postService)

	// Get auth service URL from environment
	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	if authServiceURL == "" {
		authServiceURL = "http://auth-service:8081" // Default for Docker
	}

	// Create middleware
	authMiddleware := middleware.AuthMiddleware(authServiceURL)
	rateLimiter := middleware.NewRateLimiter()

	// Create base mux with CORS and Logging
	mux := http.NewServeMux()

	// Apply CORS and Logging to all routes
	var handler http.Handler = mux
	handler = middleware.Logging(handler)
	handler = middleware.CORS(handler)

	// Protected routes (require authentication)
	protectedMux := http.NewServeMux()

	// Post routes
	protectedMux.HandleFunc("/posts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			postHandlers.GetFeed(w, r)
		} else if r.Method == "POST" {
			// Apply rate limiting to post creation
			rateLimiter.RateLimit(http.HandlerFunc(postHandlers.CreatePost)).ServeHTTP(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	protectedMux.HandleFunc("/posts/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			postHandlers.GetPost(w, r)
		} else if r.Method == "PUT" {
			postHandlers.UpdatePost(w, r)
		} else if r.Method == "DELETE" {
			postHandlers.DeletePost(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Comment routes
	protectedMux.HandleFunc("/comments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			postHandlers.GetComments(w, r)
		} else if r.Method == "POST" {
			// Apply rate limiting to comment creation
			rateLimiter.RateLimit(http.HandlerFunc(postHandlers.CreateComment)).ServeHTTP(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Apply auth middleware to all protected routes
	mux.Handle("/", authMiddleware(protectedMux))

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	log.Printf("Post service starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal("Server failed:", err)
	}
}

// OpenDB opens a connection to the SQLite database
func OpenDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	// Test connection
	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
