package migrations

import "testing"

func TestDefinitionsIncludeInitialMigration(t *testing.T) {
	definitions, err := Definitions()
	if err != nil {
		t.Fatalf("load definitions: %v", err)
	}
	if len(definitions) != 1 {
		t.Fatalf("definition count = %d, want 1", len(definitions))
	}
	if definitions[0].Version != "000001_initial" || definitions[0].Checksum == "" {
		t.Fatalf("unexpected definition: %+v", definitions[0])
	}
}
