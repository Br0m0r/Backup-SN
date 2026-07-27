package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type migration struct {
	version int
	name    string
	path    string
}

func main() {
	databasePath := flag.String("database", "/data/social_network.db", "SQLite database path")
	migrationsPath := flag.String("migrations", "/app/db/migrations", "directory containing .up.sql migrations")
	flag.Parse()

	if err := apply(*databasePath, *migrationsPath); err != nil {
		log.Fatalf("Failed to migrate SQLite database: %v", err)
	}
	log.Printf("SQLite migrations applied to %s", *databasePath)
}

func apply(databasePath, migrationsPath string) error {
	migrations, err := discover(migrationsPath)
	if err != nil {
		return err
	}
	if len(migrations) == 0 {
		return fmt.Errorf("no .up.sql migrations found in %s", migrationsPath)
	}

	db, err := sql.Open("sqlite3", databasePath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		return fmt.Errorf("enable foreign keys: %w", err)
	}
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at DATETIME NOT NULL DEFAULT (datetime('now'))
		)
	`); err != nil {
		return fmt.Errorf("create migration history: %w", err)
	}

	for _, item := range migrations {
		var applied int
		err := db.QueryRow(`SELECT 1 FROM schema_migrations WHERE version = ?`, item.version).Scan(&applied)
		if err == nil {
			continue
		}
		if err != sql.ErrNoRows {
			return fmt.Errorf("check migration %s: %w", item.name, err)
		}

		contents, err := os.ReadFile(item.path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", item.name, err)
		}
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", item.name, err)
		}
		if _, err := tx.Exec(string(contents)); err != nil {
			tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", item.name, err)
		}
		if _, err := tx.Exec(`INSERT INTO schema_migrations (version, name) VALUES (?, ?)`, item.version, item.name); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %s: %w", item.name, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", item.name, err)
		}
		log.Printf("Applied SQLite migration %s", item.name)
	}

	return nil
}

func discover(directory string) ([]migration, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("read migrations directory: %w", err)
	}

	var migrations []migration
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}
		prefix, _, ok := strings.Cut(entry.Name(), "_")
		if !ok {
			return nil, fmt.Errorf("migration filename %q has no numeric prefix", entry.Name())
		}
		version, err := strconv.Atoi(prefix)
		if err != nil {
			return nil, fmt.Errorf("migration filename %q has invalid version: %w", entry.Name(), err)
		}
		migrations = append(migrations, migration{
			version: version,
			name:    entry.Name(),
			path:    filepath.Join(directory, entry.Name()),
		})
	}
	sort.Slice(migrations, func(i, j int) bool { return migrations[i].version < migrations[j].version })
	for i := 1; i < len(migrations); i++ {
		if migrations[i-1].version == migrations[i].version {
			return nil, fmt.Errorf("duplicate migration version %d", migrations[i].version)
		}
	}
	return migrations, nil
}
