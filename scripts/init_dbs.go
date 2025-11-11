package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Create data directory
	if err := os.MkdirAll("data", 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	databases := []struct {
		name       string
		schemaFile string
	}{
		{"data/auth_service.db", "db/migrations/auth/001_create_users.sql"},
		{"data/user_service.db", "db/migrations/users/001_create_user_profiles.sql"},
		{"data/post_service.db", "db/migrations/posts/001_create_posts.sql"},
		{"data/group_service.db", "db/migrations/groups/001_create_groups.sql"},
		{"data/chat_service.db", "db/migrations/chat/001_create_messages.sql"},
		{"data/notif_service.db", "db/migrations/notifications/001_create_notifications.sql"},
	}

	for _, dbInfo := range databases {
		fmt.Printf("Creating %s...\n", dbInfo.name)

		// Read schema file
		schema, err := os.ReadFile(dbInfo.schemaFile)
		if err != nil {
			log.Fatalf("Failed to read schema file %s: %v", dbInfo.schemaFile, err)
		}

		// Remove existing database file if it exists
		os.Remove(dbInfo.name)

		// Create and initialize database
		db, err := sql.Open("sqlite3", dbInfo.name)
		if err != nil {
			log.Fatalf("Failed to open database %s: %v", dbInfo.name, err)
		}

		// Execute schema
		if _, err := db.Exec(string(schema)); err != nil {
			db.Close()
			log.Fatalf("Failed to execute schema for %s: %v", dbInfo.name, err)
		}

		db.Close()
		fmt.Printf("✓ %s created successfully\n\n", dbInfo.name)
	}

	fmt.Println("All databases created successfully!")
	fmt.Println("\nDatabase files:")
	files, _ := os.ReadDir("data")
	for _, file := range files {
		if !file.IsDir() {
			info, _ := file.Info()
			fmt.Printf("  - %s (%d bytes)\n", file.Name(), info.Size())
		}
	}
}
