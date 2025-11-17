package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"

	"social-network/services/common/authcache"
	"social-network/services/posts/handlers"
	"social-network/services/posts/middleware"
	"social-network/services/posts/services"
)

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
	authMiddleware := authcache.AuthMiddleware(authServiceURL)
	log.Printf("Using simple auth cache with 5-minute TTL")

	// Setup routes
	mux := http.NewServeMux()

	// Health check (no auth, no rate limiting)
	mux.HandleFunc("/health", handlers.HealthHandler)

	// Post endpoints
	mux.Handle("/posts", authMiddleware(rateLimiter.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			postHandlers.CreatePost(w, r)
		case "GET":
			postHandlers.GetFeed(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))))

	mux.Handle("/posts/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			postHandlers.GetPost(w, r)
		case "PUT":
			postHandlers.UpdatePost(w, r)
		case "DELETE":
			postHandlers.DeletePost(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Comment endpoints
	mux.Handle("/comments", authMiddleware(rateLimiter.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			postHandlers.CreateComment(w, r)
		case "GET":
			postHandlers.GetComments(w, r)
		default:
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

// OpenDB opens a connection to the SQLite database
func OpenDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Enable foreign key constraints
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		db.Close()
		return nil, err
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	log.Printf("Connected to database: %s", dbPath)
	return db, nil
}
