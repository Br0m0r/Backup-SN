package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"

	"social-network/services/auth/handlers"
	"social-network/services/auth/middleware"
	"social-network/services/auth/services"
)

func main() {
	// Get database path from environment or use default
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "/app/database.db"
	}

	// Initialize database connection
	db, err := OpenDB(dbPath)
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

	// Public endpoints (need CORS for browsers)
	publicMux := http.NewServeMux()
	publicMux.HandleFunc("/register", rateLimiter.RateLimit(http.HandlerFunc(authHandlers.Register)).ServeHTTP)
	publicMux.HandleFunc("/login", rateLimiter.RateLimit(http.HandlerFunc(authHandlers.Login)).ServeHTTP)
	publicMux.HandleFunc("/logout", authHandlers.Logout)
	publicMux.HandleFunc("/session", tokenHandlers.GetSession)

	// Internal endpoints (no CORS needed)
	internalMux := http.NewServeMux()
	internalMux.HandleFunc("/internal/verify-token", tokenHandlers.VerifyToken)
	internalMux.HandleFunc("/internal/user/", tokenHandlers.GetUserByID)
	internalMux.HandleFunc("/internal/users/batch", tokenHandlers.GetUsersBatch)
	internalMux.HandleFunc("/health", handlers.HealthHandler)

	// Main router
	mainMux := http.NewServeMux()

	// Apply CORS only to public routes
	publicHandler := middleware.CORS(
		middleware.Logging(publicMux),
	)

	// No CORS for internal routes (just logging)
	internalHandler := middleware.Logging(internalMux)

	// Route based on path
	mainMux.Handle("/internal/", internalHandler)
	mainMux.Handle("/health", internalHandler)
	mainMux.Handle("/", publicHandler)

	log.Println("Auth Service starting on port :8081")
	log.Fatal(http.ListenAndServe(":8081", mainMux))
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

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	log.Printf("Connected to database: %s", path)
	return db, nil
}
