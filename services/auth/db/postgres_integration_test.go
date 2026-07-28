package db

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"social-network/services/auth/migrations"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestAccountRepositoryAgainstPostgres(t *testing.T) {
	databaseURL := os.Getenv("AUTH_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("AUTH_TEST_DATABASE_URL is not configured")
	}
	database, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := migrations.Apply(context.Background(), database); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec("TRUNCATE TABLE sessions, accounts RESTART IDENTITY CASCADE"); err != nil {
		t.Fatal(err)
	}
	user, err := CreateUser(database, "ada", "ada@example.com", "hash", "Ada", "Lovelace", "1815-12-10", nil, nil)
	if err != nil || user.ID != 1 {
		t.Fatalf("CreateUser()=%v, %v", user, err)
	}
	loaded, err := GetUserByEmail(database, "ada@example.com")
	if err != nil || loaded.ID != user.ID || loaded.PasswordHash != "hash" {
		t.Fatalf("GetUserByEmail()=%v, %v", loaded, err)
	}
}
