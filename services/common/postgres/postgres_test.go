package postgres

import (
	"strings"
	"testing"
	"time"
)

func TestFromLookupUsesDefaults(t *testing.T) {
	config, err := FromLookup(mapLookup(map[string]string{
		"DATABASE_URL": "postgres://user:secret@db:5432/notifications?sslmode=disable",
	}))
	if err != nil {
		t.Fatalf("FromLookup returned an error: %v", err)
	}

	if config.MaxOpenConns != 25 || config.MaxIdleConns != 5 {
		t.Fatalf("unexpected pool defaults: %+v", config)
	}
	if config.ConnectTimeout != 10*time.Second {
		t.Fatalf("unexpected connect timeout: %s", config.ConnectTimeout)
	}
}

func TestFromLookupRejectsInvalidPool(t *testing.T) {
	_, err := FromLookup(mapLookup(map[string]string{
		"DATABASE_URL":      "postgres://db/notifications",
		"DB_MAX_OPEN_CONNS": "2",
		"DB_MAX_IDLE_CONNS": "3",
	}))
	if err == nil || !strings.Contains(err.Error(), "must not exceed") {
		t.Fatalf("expected invalid pool error, got %v", err)
	}
}

func TestDescriptionOmitsCredentials(t *testing.T) {
	description := Description("postgres://notifications:super-secret@db:5432/notifications?sslmode=disable")
	if strings.Contains(description, "super-secret") || strings.Contains(description, "notifications:super") {
		t.Fatalf("description leaked credentials: %q", description)
	}
	if description != "PostgreSQL host=db:5432 database=notifications" {
		t.Fatalf("unexpected description: %q", description)
	}
}

func mapLookup(values map[string]string) func(string) (string, bool) {
	return func(name string) (string, bool) {
		value, ok := values[name]
		return value, ok
	}
}
