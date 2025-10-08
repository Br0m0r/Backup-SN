package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"

	"social-network/services/auth/handlers"
	"social-network/services/auth/middleware"
	"social-network/services/auth/services"
)

func main() {
	// Initialize database connection
	db, err := OpenDB("/app/social_network.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Initialize services
	authService := services.NewAuthService(db)

	// Initialize handlers
	authHandlers := handlers.NewAuthHandlers(authService)
	tokenHandlers := handlers.NewTokenHandlers(authService)

	// Initialize middleware
	rateLimiter := middleware.NewRateLimiter()

	// Setup routes
	mux := http.NewServeMux()

	// Authentication routes (as per flowchart)
	mux.HandleFunc("/register", authHandlers.Register)
	mux.HandleFunc("/login", authHandlers.Login)
	mux.HandleFunc("/logout", authHandlers.Logout)
	mux.HandleFunc("/session", tokenHandlers.GetSession)

	// Internal routes (for microservice communication)
	mux.HandleFunc("/internal/verify-token", tokenHandlers.VerifyToken)
	mux.HandleFunc("/internal/user/", tokenHandlers.GetUserByID)

	// Health check
	mux.HandleFunc("/health", handlers.HealthHandler)

	// Apply middleware chain
	handler := middleware.CORS(
		middleware.Logging(
			rateLimiter.RateLimit(mux),
		),
	)

	// Start server
	log.Println("Auth Service starting on port :8081")
	log.Fatal(http.ListenAndServe(":8081", handler))
}

// OpenDB opens a connection to the SQLite database
func OpenDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
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

	return db, nil
}
