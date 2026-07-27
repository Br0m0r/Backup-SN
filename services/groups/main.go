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
	"social-network/services/groups/handlers"
	"social-network/services/groups/middleware"
	"social-network/services/groups/services"
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

	// Initialize services
	groupService := services.NewGroupService(db)

	// Initialize handlers
	groupHandlers := handlers.NewGroupHandlers(groupService, mediaStore)

	// Get auth service URL from environment
	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	if authServiceURL == "" {
		authServiceURL = "http://auth-service:8081"
	}
	if err := notify.ValidateConfig(); err != nil {
		log.Fatalf("Invalid notification client configuration: %v", err)
	}

	// Apply middleware
	authMiddleware := authcache.AuthMiddleware(authServiceURL)
	rateLimiter := middleware.NewRateLimiter()
	log.Printf("Using simple auth cache with 5-minute TTL")

	// Setup routes
	mux := http.NewServeMux()

	// Health check (no auth required)
	mux.HandleFunc("/health", handlers.HealthHandler)

	// Group routes (auth required + rate limited for write operations)
	mux.Handle("/groups", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			rateLimiter.RateLimit(http.HandlerFunc(groupHandlers.CreateGroup)).ServeHTTP(w, r)
		case "GET":
			groupHandlers.GetGroups(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/groups/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Route based on path pattern
		path := r.URL.Path

		if strings.HasSuffix(path, "/image") {
			if r.Method == "PUT" {
				rateLimiter.RateLimit(http.HandlerFunc(groupHandlers.UpdateGroupImage)).ServeHTTP(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if strings.HasSuffix(path, "/invite") {
			rateLimiter.RateLimit(http.HandlerFunc(groupHandlers.InviteMember)).ServeHTTP(w, r)
		} else if strings.HasSuffix(path, "/request") {
			rateLimiter.RateLimit(http.HandlerFunc(groupHandlers.RequestToJoin)).ServeHTTP(w, r)
		} else if strings.HasSuffix(path, "/requests/respond") {
			rateLimiter.RateLimit(http.HandlerFunc(groupHandlers.RespondToRequest)).ServeHTTP(w, r)
		} else if strings.HasSuffix(path, "/requests") {
			groupHandlers.GetPendingRequests(w, r)
		} else if strings.HasSuffix(path, "/members") {
			groupHandlers.GetMembers(w, r)
		} else if strings.HasSuffix(path, "/leave") {
			if r.Method == "DELETE" {
				rateLimiter.RateLimit(http.HandlerFunc(groupHandlers.LeaveGroup)).ServeHTTP(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if strings.HasSuffix(path, "/events") {
			if r.Method == "GET" {
				groupHandlers.GetGroupEvents(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if strings.HasSuffix(path, "/messages") {
			switch r.Method {
			case "POST":
				rateLimiter.RateLimit(http.HandlerFunc(groupHandlers.CreateGroupMessage)).ServeHTTP(w, r)
			case "GET":
				groupHandlers.GetGroupMessages(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			// Single group operations
			switch r.Method {
			case "GET":
				groupHandlers.GetGroup(w, r)
			case "PUT":
				rateLimiter.RateLimit(http.HandlerFunc(groupHandlers.UpdateGroup)).ServeHTTP(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		}
	})))

	// Invitation routes (auth required)
	mux.Handle("/invitations", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			groupHandlers.GetMyInvitations(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/invitations/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/respond") {
			rateLimiter.RateLimit(http.HandlerFunc(groupHandlers.RespondToInvitation)).ServeHTTP(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Event routes (auth required + rate limited for writes)
	mux.Handle("/events", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			rateLimiter.RateLimit(http.HandlerFunc(groupHandlers.CreateEvent)).ServeHTTP(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/events/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/respond") {
			rateLimiter.RateLimit(http.HandlerFunc(groupHandlers.RespondToEvent)).ServeHTTP(w, r)
		} else if r.Method == "GET" {
			groupHandlers.GetEvent(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Browser traffic reaches this private service through the gateway.
	handler := middleware.Logging(mux)

	// Start server
	address := httpserver.Address("8084")
	log.Printf("Group Service starting on %s", address)
	if err := httpserver.Run(httpserver.New(address, handler)); err != nil {
		log.Fatalf("Group Service stopped with error: %v", err)
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
