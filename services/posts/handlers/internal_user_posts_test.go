package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"social-network/services/common/serviceauth"
	"social-network/services/posts/services"

	_ "github.com/mattn/go-sqlite3"
)

func TestInternalUserPostsContract(t *testing.T) {
	const token = "test-internal-token-at-least-32-characters"
	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	database.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = database.Close() })
	if _, err := database.Exec(`
		CREATE TABLE posts (
			id INTEGER PRIMARY KEY, user_id INTEGER, group_id INTEGER, title TEXT,
			content TEXT, image_path TEXT, privacy_level TEXT, created_at DATETIME
		);
		INSERT INTO posts VALUES (9,7,NULL,NULL,'hello',NULL,'public','2026-01-01T00:00:00Z');
	`); err != nil {
		t.Fatal(err)
	}
	handler := NewInternalUserPostHandlers(services.NewPostService(database, nil))
	protected := serviceauth.Authenticate(token, handler)
	request := httptest.NewRequest(http.MethodGet, "/internal/v1/users/7/posts", nil)
	request.Header.Set(serviceauth.HeaderName, token)
	response := httptest.NewRecorder()
	protected.ServeHTTP(response, request)
	var body struct {
		Data struct {
			Count int `json:"count"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusOK || body.Data.Count != 1 {
		t.Fatalf("response = %d %s", response.Code, response.Body.String())
	}
}
