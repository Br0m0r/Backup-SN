package db

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestSearchGroupsFiltersNameAndDescription(t *testing.T) {
	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	database.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = database.Close() })

	schema := `
		CREATE TABLE groups (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			image_url TEXT,
			creator_id INTEGER NOT NULL,
			created_at DATETIME NOT NULL DEFAULT (datetime('now'))
		);
		CREATE TABLE group_members (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			group_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			status TEXT NOT NULL
		);
		INSERT INTO groups (name, description, image_url, creator_id) VALUES
			('Alpha Runners', 'Training together', '', 1),
			('Kitchen Club', 'Recipes and baking', '', 2),
			('Quiet Readers', NULL, '', 3);
		INSERT INTO group_members (group_id, user_id, status) VALUES (1, 1, 'accepted');
	`
	if _, err := database.Exec(schema); err != nil {
		t.Fatalf("create fixtures: %v", err)
	}

	tests := []struct {
		name       string
		query      string
		wantName   string
		wantLength int
	}{
		{name: "name is case insensitive", query: "ALPHA", wantName: "Alpha Runners", wantLength: 1},
		{name: "description is searchable", query: "baking", wantName: "Kitchen Club", wantLength: 1},
		{name: "no match", query: "cycling", wantLength: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			groups, err := SearchGroups(database, 1, test.query)
			if err != nil {
				t.Fatalf("search groups: %v", err)
			}
			if len(groups) != test.wantLength {
				t.Fatalf("result length = %d, want %d", len(groups), test.wantLength)
			}
			if test.wantLength > 0 && groups[0].Name != test.wantName {
				t.Fatalf("group name = %q, want %q", groups[0].Name, test.wantName)
			}
		})
	}
}
