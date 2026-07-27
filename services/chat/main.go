package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"social-network/services/chat/groupsclient"
	"social-network/services/chat/handlers"
	"social-network/services/chat/middleware"
	"social-network/services/common/authcache"
	"social-network/services/common/httpserver"
	"social-network/services/common/notify"
	"social-network/services/common/objectstore"
	"social-network/services/common/origin"
	"social-network/services/common/postgres"
	"social-network/services/common/realtime"
	"social-network/services/common/redisstore"
	"social-network/services/common/serviceauth"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	log.Println("Starting Chat Service...")

	messageDatabaseConfig, err := postgres.FromEnvironment()
	if err != nil {
		log.Fatalf("Invalid Chat database configuration: %v", err)
	}
	messageDatabase, err := postgres.Open(context.Background(), messageDatabaseConfig)
	if err != nil {
		log.Fatalf("Failed to open Chat database: %v", err)
	}
	defer messageDatabase.Close()
	log.Printf("Connected to %s", postgres.Description(messageDatabaseConfig.URL))

	// The temporary identity connection is read-only at the Chat boundary and
	// remains until Users profile/follow contracts replace it.
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./social_network.db"
	}

	// Open database connection
	identityDatabase, err := OpenLegacyDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer identityDatabase.Close()

	log.Printf("Connected to database at %s", dbPath)

	// Get auth service URL
	authServiceURL := middleware.GetAuthServiceURL()
	log.Printf("Auth service URL: %s", authServiceURL)
	if err := notify.ValidateConfig(); err != nil {
		log.Fatalf("Invalid notification client configuration: %v", err)
	}
	internalServiceToken, err := serviceauth.TokenFromEnvironment()
	if err != nil {
		log.Fatalf("Invalid internal service authentication configuration: %v", err)
	}
	groupMembership, err := groupsclient.FromEnvironment(internalServiceToken)
	if err != nil {
		log.Fatalf("Invalid Groups client configuration: %v", err)
	}

	originValidator, err := origin.FromEnvironment()
	if err != nil {
		log.Fatalf("Invalid WebSocket origin configuration: %v", err)
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
	redisConfig, err := redisstore.FromEnvironment()
	if err != nil {
		log.Fatalf("Invalid Redis configuration: %v", err)
	}
	redisState, err := redisstore.Open(context.Background(), redisConfig)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisState.Close()
	realtimeBus, err := realtime.New(redisState, "chat")
	if err != nil {
		log.Fatalf("Invalid Chat realtime configuration: %v", err)
	}
	serviceContext, cancelService := context.WithCancel(context.Background())
	defer cancelService()

	// Create WebSocket hub
	hub := handlers.NewHub(messageDatabase, identityDatabase, originValidator.Check, realtimeBus, groupMembership)
	go hub.Run(serviceContext)

	// Create handlers
	chatHandlers := handlers.NewChatHandlers(messageDatabase, identityDatabase, hub, groupMembership)
	uploadHandlers := handlers.NewUploadHandlers(mediaStore)

	// Create auth middleware and rate limiter
	authMiddleware := authcache.AuthMiddleware(authServiceURL)
	rateLimiter := middleware.NewRateLimiter()
	log.Printf("Using simple auth cache with 5-minute TTL")

	// Setup routes
	mux := http.NewServeMux()

	// Health check (no auth required)
	mux.HandleFunc("/health", chatHandlers.HealthCheck)

	// WebSocket endpoint (auth required via query param or header)
	mux.Handle("/ws", authMiddleware(http.HandlerFunc(hub.HandleWebSocket)))

	// REST endpoints (auth required + rate limited for write operations)
	mux.Handle("/chat/conversations", authMiddleware(http.HandlerFunc(chatHandlers.GetConversations)))
	mux.Handle("/chat/contacts", authMiddleware(http.HandlerFunc(chatHandlers.GetAvailableContacts)))
	mux.Handle("/chat/history/", authMiddleware(http.HandlerFunc(chatHandlers.GetChatHistory)))
	mux.Handle("/chat/read/", authMiddleware(rateLimiter.RateLimit(http.HandlerFunc(chatHandlers.MarkAsRead))))
	mux.Handle("/chat/unread", authMiddleware(http.HandlerFunc(chatHandlers.GetUnreadCount)))
	mux.Handle("/chat/send", authMiddleware(rateLimiter.RateLimit(http.HandlerFunc(chatHandlers.SendMessage))))

	// Upload endpoints (auth required + rate limited)
	mux.Handle("/upload/image", authMiddleware(rateLimiter.RateLimit(http.HandlerFunc(uploadHandlers.UploadImage))))
	mux.Handle("/upload/delete", authMiddleware(rateLimiter.RateLimit(http.HandlerFunc(uploadHandlers.DeleteImage))))

	// Group chat endpoints (auth required + rate limited for writes)
	mux.Handle("/chat/groups/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Route based on path pattern
		path := r.URL.Path

		if strings.HasSuffix(path, "/history") {
			chatHandlers.GetGroupChatHistory(w, r)
		} else if strings.HasSuffix(path, "/messages") {
			if r.Method == "POST" {
				rateLimiter.RateLimit(http.HandlerFunc(chatHandlers.SendGroupMessage)).ServeHTTP(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	})))

	// Browser traffic reaches this private service through the gateway.
	handler := middleware.Logging(mux)

	// Start server
	address := httpserver.Address("8085")
	log.Printf("Chat Service starting on %s", address)
	if err := httpserver.Run(httpserver.New(address, handler)); err != nil {
		log.Fatalf("Chat Service stopped with error: %v", err)
	}
}

// OpenDB opens a connection to the SQLite database
func OpenLegacyDB(dbPath string) (*sql.DB, error) {
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

	// Set journal mode to DELETE for simplicity
	_, err = db.Exec("PRAGMA journal_mode=DELETE;")
	if err != nil {
		return nil, err
	}

	return db, nil
}
