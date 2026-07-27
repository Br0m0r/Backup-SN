package main

import (
	"context"
	"log"

	"social-network/services/common/postgres"
	"social-network/services/notifications/migrations"
)

func main() {
	config, err := postgres.FromEnvironment()
	if err != nil {
		log.Fatalf("Invalid notification database configuration: %v", err)
	}

	database, err := postgres.Open(context.Background(), config)
	if err != nil {
		log.Fatalf("Failed to open notification database: %v", err)
	}
	defer database.Close()

	if err := migrations.Apply(context.Background(), database); err != nil {
		log.Fatalf("Failed to migrate notification database: %v", err)
	}
	log.Printf("Notification database migrations applied to %s", postgres.Description(config.URL))
}
