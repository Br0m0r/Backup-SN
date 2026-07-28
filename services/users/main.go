package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"social-network/services/common/authcache"
	"social-network/services/common/httpserver"
	"social-network/services/common/notify"
	"social-network/services/common/objectstore"
	"social-network/services/common/serviceauth"
	"social-network/services/users/handlers"
	"social-network/services/users/middleware"
	"social-network/services/users/postsclient"
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
	objectStoreConfig, err := objectstore.FromEnvironment()
	if err != nil {
		log.Fatalf("Invalid object storage configuration: %v", err)
	}
	objectStoreContext, cancelObjectStore := context.WithTimeout(context.Background(), 10*time.Second)
	mediaStore, err := objectstore.Open(objectStoreContext, objectStoreConfig)
	cancelObjectStore()
	if err != nil {
		log.Fatalf("Failed to connect to object storage: %v", err)
	}

	// Get auth service URL from environment
	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	if authServiceURL == "" {
		authServiceURL = "http://auth-service:8081"
	}
	if err := notify.ValidateConfig(); err != nil {
		log.Fatalf("Invalid notification client configuration: %v", err)
	}
	internalServiceToken, err := serviceauth.TokenFromEnvironment()
	if err != nil {
		log.Fatalf("Invalid internal service authentication configuration: %v", err)
	}
	postReader, err := postsclient.FromEnvironment(internalServiceToken)
	if err != nil {
		log.Fatalf("Invalid Posts client configuration: %v", err)
	}

	// Initialize services
	userService := services.NewUserService(db, postReader)

	// Initialize handlers
	userHandlers := handlers.NewUserHandlers(userService)
	internalReadHandlers := handlers.NewInternalReadHandlers(userService)
	uploadHandlers := handlers.NewUploadHandlers(userService, mediaStore)

	// Apply middleware
	authMiddleware := authcache.AuthMiddleware(authServiceURL)
	rateLimiter := middleware.NewRateLimiter()
	log.Printf("Using simple auth cache with 5-minute TTL")

	// Setup routes with middleware
	mux := http.NewServeMux()

	// Health check (no auth required)
	mux.HandleFunc("/health", handlers.HealthHandler)
	mux.Handle("/internal/v1/users/", serviceauth.Authenticate(internalServiceToken, internalReadHandlers))

	// Upload routes (auth required + rate limited)
	mux.Handle("/upload/avatar", authMiddleware(rateLimiter.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			uploadHandlers.UploadAvatar(w, r)
		case "DELETE":
			uploadHandlers.DeleteAvatar(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))))

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

	// Follow status route (auth required)
	mux.Handle("/follow/status/", authMiddleware(http.HandlerFunc(userHandlers.GetFollowStatus)))

	// Follow requests routes (auth required)
	mux.Handle("/follow/requests", authMiddleware(http.HandlerFunc(userHandlers.GetPendingFollowRequests)))
	mux.Handle("/follow/respond", authMiddleware(http.HandlerFunc(userHandlers.RespondToFollowRequest)))

	// Search route (auth required + rate limited)
	mux.Handle("/search", authMiddleware(rateLimiter.RateLimit(http.HandlerFunc(userHandlers.SearchUsers))))

	// Search for group invites route (auth required + rate limited)
	mux.Handle("/search/group", authMiddleware(rateLimiter.RateLimit(http.HandlerFunc(userHandlers.SearchUsersForGroup))))

	// User profile by ID route (auth required)
	// Pattern: /users/:id/profile
	mux.Handle("/users/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Route requests based on path suffix
		if strings.HasSuffix(r.URL.Path, "/followers") {
			userHandlers.GetUserFollowers(w, r)
		} else if strings.HasSuffix(r.URL.Path, "/following") {
			userHandlers.GetUserFollowing(w, r)
		} else {
			userHandlers.GetUserProfileByID(w, r)
		}
	})))

	// User stats route (auth required)
	mux.Handle("/users/me/stats", authMiddleware(http.HandlerFunc(userHandlers.GetStats)))

	// User "me" routes (auth required) - alias for /profile
	mux.Handle("/users/me/privacy", authMiddleware(http.HandlerFunc(userHandlers.UpdatePrivacy)))
	mux.Handle("/users/me", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			userHandlers.GetCurrentUserProfile(w, r)
		case "PUT":
			userHandlers.UpdateProfile(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Browser traffic reaches this private service through the gateway.
	handler := middleware.Logging(mux)

	// Start server
	address := httpserver.Address("8082")
	log.Printf("User Service starting on %s", address)
	if err := httpserver.Run(httpserver.New(address, handler)); err != nil {
		log.Fatalf("User Service stopped with error: %v", err)
	}
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
	if err = db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	log.Printf("Connected to database: %s", dbPath)
	return db, nil
}
