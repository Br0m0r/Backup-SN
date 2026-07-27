package migrations

import (
	"strings"
	"testing"
)

func TestDefinitionsIncludesChatPostgreSQLSchema(t *testing.T) {
	definitions, err := Definitions()
	if err != nil {
		t.Fatalf("Definitions returned an error: %v", err)
	}
	if len(definitions) != 1 {
		t.Fatalf("expected one migration, got %d", len(definitions))
	}
	migration := definitions[0]
	for _, expected := range []string{"CREATE TABLE messages", "CREATE TABLE group_messages", "TIMESTAMPTZ"} {
		if !strings.Contains(migration.SQL, expected) {
			t.Errorf("migration does not contain %q", expected)
		}
	}
	if len(migration.Checksum) != 64 {
		t.Fatalf("expected SHA-256 checksum, got %q", migration.Checksum)
	}
}
