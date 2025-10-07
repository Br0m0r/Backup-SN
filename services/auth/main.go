package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3" // SQLite driver - needed for database connection
)

func main() {
	db, err := OpenDB("/app/social_network.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// ONLY auth-related routes - AUTH SERVICE RESPONSIBILITY
	http.HandleFunc("/register", registerHandler(db))
	http.HandleFunc("/login", loginHandler(db))
	http.HandleFunc("/logout", logoutHandler(db))
	http.HandleFunc("/verify-token", verifyTokenHandler(db))
	http.HandleFunc("/health", healthHandler)

	// Auth Service runs on port 8081
	log.Println("Auth Service listening on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

// OpenDB opens a connection to the SQLite database
// Each service uses this same function to connect to the SHARED database
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

	return db, nil
}

// AUTH SERVICE HANDLERS - Only authentication logic

func registerHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Registration logic here
		// 1. Parse request body for email/password
		// 2. Hash password
		// 3. Insert into users table
		// 4. Generate JWT token or create session
		// 5. Return token to client
		log.Println("AUTH SERVICE: Registration request received")
	}
}

func loginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Login logic here
		log.Println("AUTH SERVICE: Login request received")
	}
}

func logoutHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Logout logic here
		log.Println("AUTH SERVICE: Logout request received")
	}
}

func verifyTokenHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// CRITICAL: OTHER SERVICES call this endpoint to verify authentication
		token := r.Header.Get("Authorization")

		// Verify token logic (JWT validation, session check, etc.)
		if isValidToken(token) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "valid",
				"userId":  getUserIdFromToken(token),
				"service": "auth",
			})
			log.Println("AUTH SERVICE: Token verified successfully")
		} else {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			log.Println("AUTH SERVICE: Token verification failed")
		}
	}
}

// Health check endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy", "service": "auth"})
}

// Helper functions
func isValidToken(token string) bool {
	// For study purposes, let's just check if token is not empty
	return token != ""
}

func getUserIdFromToken(token string) string {
	// For study purposes, return a dummy user ID
	return "user123"
}
