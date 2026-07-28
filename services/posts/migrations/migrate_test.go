package migrations

import (
	"strings"
	"testing"
)

func TestDefinitionsIncludesPostsSchema(t *testing.T) {
	definitions, err := Definitions()
	if err != nil || len(definitions) != 1 {
		t.Fatalf("Definitions()=%v, %v", definitions, err)
	}
	for _, expected := range []string{"CREATE TABLE posts", "CREATE TABLE post_viewers", "CREATE TABLE comments", "TIMESTAMPTZ"} {
		if !strings.Contains(definitions[0].SQL, expected) {
			t.Errorf("migration does not contain %q", expected)
		}
	}
}
