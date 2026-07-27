package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestApplyRunsPendingMigrationsOnce(t *testing.T) {
	directory := t.TempDir()
	migrations := filepath.Join(directory, "migrations")
	if err := os.Mkdir(migrations, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(migrations, "000001_create_items.up.sql"), []byte(`CREATE TABLE items (id INTEGER PRIMARY KEY);`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(migrations, "000002_add_name.up.sql"), []byte(`ALTER TABLE items ADD COLUMN name TEXT;`), 0o600); err != nil {
		t.Fatal(err)
	}

	database := filepath.Join(directory, "database.sqlite")
	if err := apply(database, migrations); err != nil {
		t.Fatalf("first apply failed: %v", err)
	}
	if err := apply(database, migrations); err != nil {
		t.Fatalf("second apply failed: %v", err)
	}

	db, err := sql.Open("sqlite3", database)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("expected 2 applied migrations, got %d", count)
	}
}
