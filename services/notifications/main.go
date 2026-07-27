package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	"social-network/services/common/authcache"
	"social-network/services/common/httpserver"
	"social-network/services/common/origin"
	"social-network/services/common/postgres"
	"social-network/services/common/realtime"
	"social-network/services/common/redisstore"
	"social-network/services/common/serviceauth"
	"social-network/services/notifications/handlers"
	"social-network/services/notifications/middleware"
)

func main() {
	log.Println("Starting Notification Service...")

	databaseConfig, err := postgres.FromEnvironment()
	if err != nil {
		log.Fatalf("Invalid notification database configuration: %v", err)
	}

	// Schema changes are applied by the separate notification migration job.
	database, err := OpenDB(databaseConfig)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	log.Printf("Connected to %s", postgres.Description(databaseConfig.URL))

	// Get auth service URL
	authServiceURL := middleware.GetAuthServiceURL()
	log.Printf("Auth service URL: %s", authServiceURL)

	originValidator, err := origin.FromEnvironment()
	if err != nil {
		log.Fatalf("Invalid WebSocket origin configuration: %v", err)
	}
	redisConfig, err := redisstore.FromEnvironment()
	if err != nil {
		log.Fatalf("Invalid Redis configuration: %v", err)
	}
	redisState, err := redisstore.Open(context.Background(), redisConfig)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisState.Close()
	realtimeBus, err := realtime.New(redisState, "notifications")
	if err != nil {
		log.Fatalf("Invalid Notification realtime configuration: %v", err)
	}
	serviceContext, cancelService := context.WithCancel(context.Background())
	defer cancelService()

	// Create notification hub for WebSocket
	hub := handlers.NewNotificationHub(database, originValidator.Check, realtimeBus)
	go hub.Run(serviceContext)

	// Create handlers
	notifHandlers := handlers.NewNotificationHandlers(database, hub)

	// Create auth middleware and rate limiter
	authMiddleware := authcache.AuthMiddleware(authServiceURL)
	internalServiceToken, err := serviceauth.TokenFromEnvironment()
	if err != nil {
		log.Fatalf("Invalid internal service authentication configuration: %v", err)
	}
	rateLimiter := middleware.NewRateLimiter()
	log.Printf("Using simple auth cache with 5-minute TTL")

	// Setup routes
	mux := http.NewServeMux()

	// Health check (no auth required)
	mux.HandleFunc("/health", notifHandlers.HealthCheck)

	// Create notification (called only by authenticated internal services).
	// The shared token is an interim control until workload identity or mTLS is available.
	mux.Handle("/notifications", serviceauth.Authenticate(
		internalServiceToken,
		rateLimiter.RateLimit(http.HandlerFunc(notifHandlers.CreateNotification)),
	))

	// Get notifications (auth required)
	mux.Handle("/notifications/list", authMiddleware(http.HandlerFunc(notifHandlers.GetNotifications)))

	// Get unread count (auth required)
	mux.Handle("/notifications/unread-count", authMiddleware(http.HandlerFunc(notifHandlers.GetUnreadCount)))

	// Mark as read (auth required + rate limited)
	mux.Handle("/notifications/read/", authMiddleware(rateLimiter.RateLimit(http.HandlerFunc(notifHandlers.MarkAsRead))))

	// Mark all as read (auth required + rate limited)
	mux.Handle("/notifications/read-all", authMiddleware(rateLimiter.RateLimit(http.HandlerFunc(notifHandlers.MarkAllAsRead))))

	// Delete notification (auth required + rate limited)
	mux.Handle("/notifications/delete/", authMiddleware(rateLimiter.RateLimit(http.HandlerFunc(notifHandlers.DeleteNotification))))

	// WebSocket endpoint (auth required via query param)
	mux.Handle("/ws", authMiddleware(http.HandlerFunc(hub.HandleWebSocket)))

	// Browser traffic reaches this private service through the gateway.
	handler := middleware.Logging(mux)

	// Start server
	address := httpserver.Address("8086")
	log.Printf("Notification Service starting on %s", address)
	if err := httpserver.Run(httpserver.New(address, handler)); err != nil {
		log.Fatalf("Notification Service stopped with error: %v", err)
	}
}

// OpenDB opens the notification service's PostgreSQL connection pool.
func OpenDB(config postgres.Config) (*sql.DB, error) {
	return postgres.Open(context.Background(), config)
}
