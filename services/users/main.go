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
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "/app/social_network.db"
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

	// Apply middleware
	authMiddleware := middleware.AuthMiddleware(authServiceURL)
	rateLimiter := middleware.NewRateLimiter()

	// Setup routes with middleware
	mux := http.NewServeMux()

	// Health check (no auth required)
	mux.HandleFunc("/health", handlers.HealthHandler)

	// Profile routes (auth required)
	mux.Handle("/profile/", authMiddleware(http.HandlerFunc(userHandlers.GetProfile)))
	mux.Handle("/profile", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Route /profile matched: method=%s, path=%s", r.Method, r.URL.Path)
		switch r.Method {
		case "GET":
			log.Printf("Routing GET /profile to GetCurrentUserProfile")
			userHandlers.GetCurrentUserProfile(w, r)
		case "PUT":
			log.Printf("Routing PUT /profile to UpdateProfile")
			userHandlers.UpdateProfile(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Follow routes (auth required + rate limited)
	mux.Handle("/follow", authMiddleware(rateLimiter.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			userHandlers.FollowUser(w, r)
		case "DELETE":
			userHandlers.UnfollowUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))))

	// User relationship routes (auth required)
	mux.Handle("/followers", authMiddleware(http.HandlerFunc(userHandlers.GetFollowers)))
	mux.Handle("/following", authMiddleware(http.HandlerFunc(userHandlers.GetFollowing)))

	// Search route (auth required)
	mux.Handle("/search", authMiddleware(http.HandlerFunc(userHandlers.SearchUsers)))

	// Apply common middleware (CORS and Logging)
	handler := middleware.CORS(
		middleware.Logging(mux),
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
