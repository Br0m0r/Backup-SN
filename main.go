package main

import (
	"database/sql"
	"log"
	"net/http"
)

func main() {
	db, err := OpenDB("social_network.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	log.Println("Database connected and foreign keys enabled!")

	// Start HTTP server listening on port 8081
	log.Println("Server listening on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func OpenDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	// Enable foreign key constraints
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
