package migrations

import (
	"testing"
)

func TestInitialMigrationIsEmbedded(t *testing.T) {
	contents, err := files.ReadFile("000001_initial.up.sql")
	if err != nil || len(contents) == 0 {
		t.Fatalf("embedded migration: %v", err)
	}
}
