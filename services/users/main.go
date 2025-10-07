package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

func main() {
	db, err := OpenDB("/app/social_network.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Get auth service URL from environment
	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	if authServiceURL == "" {
		authServiceURL = "http://auth-service:8081" // Default for Docker
	}

	// ONLY user-related routes - USER SERVICE RESPONSIBILITY
	http.HandleFunc("/profile", profileHandler(db, authServiceURL))
	http.HandleFunc("/follow", followHandler(db, authServiceURL))
	http.HandleFunc("/unfollow", unfollowHandler(db, authServiceURL))
	http.HandleFunc("/followers", followersHandler(db, authServiceURL))
	http.HandleFunc("/following", followingHandler(db, authServiceURL))
	http.HandleFunc("/search", searchHandler(db, authServiceURL))
	http.HandleFunc("/health", healthHandler)

	// User Service runs on port 8082
	log.Println("User Service listening on :8082")
	log.Fatal(http.ListenAndServe(":8082", nil))
}

// Same OpenDB function as Auth Service - connects to SHARED database
func OpenDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// USER SERVICE HANDLERS - Only user profile and relationship logic

func profileHandler(db *sql.DB, authServiceURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Step 1: Verify authentication by calling Auth Service
		authToken := r.Header.Get("Authorization")

		if !verifyWithAuthService(authToken, authServiceURL) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Step 2: Handle user profile logic
		switch r.Method {
		case "GET":
			// Get user profile from database
			log.Println("USER SERVICE: Get profile request")
		case "PUT":
			// Update user profile in database
			log.Println("USER SERVICE: Update profile request")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func followHandler(db *sql.DB, authServiceURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		authToken := r.Header.Get("Authorization")
		if !verifyWithAuthService(authToken, authServiceURL) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Follow logic - insert into follows table
		log.Println("USER SERVICE: Follow request")
	}
}

func unfollowHandler(db *sql.DB, authServiceURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		authToken := r.Header.Get("Authorization")
		if !verifyWithAuthService(authToken, authServiceURL) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Unfollow logic - delete from follows table
		log.Println("USER SERVICE: Unfollow request")
	}
}

func followersHandler(db *sql.DB, authServiceURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authToken := r.Header.Get("Authorization")
		if !verifyWithAuthService(authToken, authServiceURL) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get followers list from database
		log.Println("USER SERVICE: Get followers request")
	}
}

func followingHandler(db *sql.DB, authServiceURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authToken := r.Header.Get("Authorization")
		if !verifyWithAuthService(authToken, authServiceURL) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get following list from database
		log.Println("USER SERVICE: Get following request")
	}
}

func searchHandler(db *sql.DB, authServiceURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authToken := r.Header.Get("Authorization")
		if !verifyWithAuthService(authToken, authServiceURL) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Search users in database
		log.Println("USER SERVICE: Search users request")
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy", "service": "users"})
}

// Helper function - User Service calls Auth Service to verify tokens
// This is INTER-SERVICE COMMUNICATION in microservices
func verifyWithAuthService(token, authServiceURL string) bool {
	if token == "" {
		return false
	}

	// Make HTTP call to Auth Service
	req, err := http.NewRequest("GET", authServiceURL+"/verify-token", nil)
	if err != nil {
		log.Printf("USER SERVICE: Error creating auth request: %v", err)
		return false
	}

	req.Header.Set("Authorization", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("USER SERVICE: Error calling auth service: %v", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		log.Println("USER SERVICE: Token verified by auth service")
		return true
	}

	log.Println("USER SERVICE: Token rejected by auth service")
	return false
}
