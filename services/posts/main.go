package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"social-network/services/common/authcache"
	"social-network/services/common/httpserver"
	"social-network/services/common/notify"
	"social-network/services/common/objectstore"
	"social-network/services/common/postgres"
	"social-network/services/common/serviceauth"
	"social-network/services/posts/handlers"
	"social-network/services/posts/middleware"
	"social-network/services/posts/services"
	"social-network/services/posts/usersclient"
)

func main() {
	databaseConfig, err := postgres.FromEnvironment()
	if err != nil {
		log.Fatalf("Invalid Posts database configuration: %v", err)
	}
	database, err := postgres.Open(context.Background(), databaseConfig)
	if err != nil {
		log.Fatalf("Failed to connect to Posts database: %v", err)
	}
	defer database.Close()
	log.Printf("Connected to %s", postgres.Description(databaseConfig.URL))

	// Get auth service URL from environment variable
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
	userDirectory, err := usersclient.FromEnvironment(internalServiceToken)
	if err != nil {
		log.Fatalf("Invalid Users client configuration: %v", err)
	}
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
	postService := services.NewPostService(database, userDirectory)

	// Initialize handlers
	postHandlers := handlers.NewPostHandlers(postService)
	internalUserPostHandlers := handlers.NewInternalUserPostHandlers(postService)
	uploadHandlers := handlers.NewUploadHandlers(mediaStore)

	// Initialize middleware
	rateLimiter := middleware.NewRateLimiter()

	authMiddleware := authcache.AuthMiddleware(authServiceURL)
	log.Printf("Using simple auth cache with 5-minute TTL")

	// Setup routes
	mux := http.NewServeMux()

	// Health check (no auth, no rate limiting)
	mux.HandleFunc("/health", postHandlers.HealthCheck)
	mux.Handle("/internal/v1/users/", serviceauth.Authenticate(internalServiceToken, internalUserPostHandlers))

	// Post endpoints
	mux.Handle("/posts", authMiddleware(rateLimiter.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			postHandlers.CreatePost(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))))

	// Feed endpoint
	mux.Handle("/posts/feed", authMiddleware(http.HandlerFunc(postHandlers.GetFeed)))

	// Search endpoint
	mux.Handle("/posts/search", authMiddleware(http.HandlerFunc(postHandlers.SearchPosts)))

	// Group posts endpoint
	mux.Handle("/posts/group/", authMiddleware(http.HandlerFunc(postHandlers.GetGroupPosts)))

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

	// Comment by ID endpoints (update, delete)
	mux.Handle("/comments/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PUT":
			postHandlers.UpdateComment(w, r)
		case "DELETE":
			postHandlers.DeleteComment(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Upload endpoints
	mux.Handle("/upload/image", authMiddleware(rateLimiter.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			uploadHandlers.UploadImage(w, r)
		case "DELETE":
			uploadHandlers.DeleteImage(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))))

	// Browser traffic reaches this private service through the gateway.
	handler := middleware.Logging(mux)

	// Start server
	address := httpserver.Address("8083")
	log.Printf("Post Service starting on %s", address)
	if err := httpserver.Run(httpserver.New(address, handler)); err != nil {
		log.Fatalf("Post Service stopped with error: %v", err)
	}
}
