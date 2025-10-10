package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"

	"social-network/services/users/handlers"
	"social-network/services/users/middleware"
	"social-network/services/users/services"
)

func main() {
	// Get database path from environment or use default
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./social_network.db"
	}

	// Open database connection
	db, err := OpenDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize services
	userService := services.NewUserService(db)

	// Initialize handlers
	userHandlers := handlers.NewUserHandlers(userService)

	// Get auth service URL from environment
	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	if authServiceURL == "" {
		authServiceURL = "http://auth-service:8081"
	}

	// Setup routes
	mux := http.NewServeMux()

	// Health check (no auth required)
	mux.HandleFunc("/health", handlers.HealthHandler)

	// Profile routes (auth required)
	mux.HandleFunc("/profile/", userHandlers.GetProfile)
	mux.HandleFunc("/profile", userHandlers.UpdateProfile)

	// Follow routes (auth required)
	mux.HandleFunc("/follow", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			userHandlers.FollowUser(w, r)
		case "DELETE":
			userHandlers.UnfollowUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/followers", userHandlers.GetFollowers)
	mux.HandleFunc("/following", userHandlers.GetFollowing)

	// Search route (auth required)
	mux.HandleFunc("/search", userHandlers.SearchUsers)

	// Apply middleware
	authMiddleware := middleware.AuthMiddleware(authServiceURL)

	// Protected routes (everything except health)
	protectedMux := http.NewServeMux()
	protectedMux.Handle("/profile/", authMiddleware(http.HandlerFunc(userHandlers.GetProfile)))
	protectedMux.Handle("/profile", authMiddleware(http.HandlerFunc(userHandlers.UpdateProfile)))
	protectedMux.Handle("/follow", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			userHandlers.FollowUser(w, r)
		case "DELETE":
			userHandlers.UnfollowUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))
	protectedMux.Handle("/followers", authMiddleware(http.HandlerFunc(userHandlers.GetFollowers)))
	protectedMux.Handle("/following", authMiddleware(http.HandlerFunc(userHandlers.GetFollowing)))
	protectedMux.Handle("/search", authMiddleware(http.HandlerFunc(userHandlers.SearchUsers)))

	// Health check without auth
	protectedMux.HandleFunc("/health", handlers.HealthHandler)

	// Apply common middleware (CORS and Logging)
	handler := middleware.CORS(
		middleware.Logging(protectedMux),
	)

	// Start server
	log.Println("User Service starting on port :8082")
	log.Fatal(http.ListenAndServe(":8082", handler))
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

	log.Printf("Connected to database: %s", dbPath)
	return db, nil
}
