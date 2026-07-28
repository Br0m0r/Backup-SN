package main

import (
	"context"
	"log"

	"social-network/services/auth/migrations"
	"social-network/services/common/postgres"
)

func main() {
	config, err := postgres.FromEnvironment()
	if err != nil {
		log.Fatal(err)
	}
	database, err := postgres.Open(context.Background(), config)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	if err := migrations.Apply(context.Background(), database); err != nil {
		log.Fatal(err)
	}
	log.Println("Auth PostgreSQL migrations applied")
}
