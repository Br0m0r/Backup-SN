package main

import (
	"context"
	"log"
	"net/http"

	"social-network/services/auth/handlers"
	"social-network/services/auth/middleware"
	"social-network/services/auth/services"
	"social-network/services/auth/usersclient"
	"social-network/services/common/httpserver"
	"social-network/services/common/postgres"
	"social-network/services/common/serviceauth"
)

func main() {
	config, err := postgres.FromEnvironment()
	if err != nil {
		log.Fatalf("Invalid Auth database configuration: %v", err)
	}
	db, err := postgres.Open(context.Background(), config)
	if err != nil {
		log.Fatalf("Failed to open Auth database: %v", err)
	}
	defer db.Close()
	log.Printf("Connected to %s", postgres.Description(config.URL))
	internalServiceToken, err := serviceauth.TokenFromEnvironment()
	if err != nil {
		log.Fatalf("Invalid internal service authentication configuration: %v", err)
	}
	profileProvisioner, err := usersclient.FromEnvironment(internalServiceToken)
	if err != nil {
		log.Fatalf("Invalid Users client configuration: %v", err)
	}

	// Initialize services
	authService := services.NewAuthService(db, profileProvisioner)

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
	internalMux.HandleFunc("/health", authHandlers.HealthCheck)

	// Main router
	mainMux := http.NewServeMux()

	// Browser traffic reaches this service through the same-origin gateway.
	publicHandler := middleware.Logging(publicMux)

	// No CORS for internal routes (just logging)
	internalHandler := middleware.Logging(internalMux)

	// Route based on path
	mainMux.Handle("/internal/", internalHandler)
	mainMux.Handle("/health", internalHandler)
	mainMux.Handle("/", publicHandler)

	address := httpserver.Address("8081")
	log.Printf("Auth Service starting on %s", address)
	if err := httpserver.Run(httpserver.New(address, mainMux)); err != nil {
		log.Fatalf("Auth Service stopped with error: %v", err)
	}
}
