package main

import (
	"context"
	"log"

	"social-network/services/common/postgres"
	"social-network/services/groups/migrations"
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
	log.Println("Groups PostgreSQL migrations applied")
}
