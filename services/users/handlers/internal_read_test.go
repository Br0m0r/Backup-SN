package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"social-network/services/common/serviceauth"
	"social-network/services/users/services"

	_ "github.com/mattn/go-sqlite3"
)

func TestInternalUsersReadContract(t *testing.T) {
	const token = "test-internal-token-at-least-32-characters"
	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	database.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = database.Close() })
	if _, err := database.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY, username TEXT, first_name TEXT,
			last_name TEXT, avatar_path TEXT
		);
		CREATE TABLE follows (follower_id INTEGER, following_id INTEGER, status TEXT);
		INSERT INTO users VALUES (7,'ada','Ada','Lovelace','/ada.png');
		INSERT INTO follows VALUES (42,7,'accepted'),(42,8,'pending');
	`); err != nil {
		t.Fatal(err)
	}
	handler := NewInternalReadHandlers(services.NewUserService(database, nil))
	protected := serviceauth.Authenticate(token, handler)

	unauthorized := httptest.NewRecorder()
	protected.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/internal/v1/users/profiles?ids=7", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d", unauthorized.Code)
	}

	request := httptest.NewRequest(http.MethodGet, "/internal/v1/users/42/following", nil)
	request.Header.Set(serviceauth.HeaderName, token)
	response := httptest.NewRecorder()
	protected.ServeHTTP(response, request)
	var body struct {
		Data struct {
			FollowingIDs []int `json:"following_ids"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(body.Data.FollowingIDs, []int{7}) {
		t.Fatalf("following IDs = %v", body.Data.FollowingIDs)
	}
}
