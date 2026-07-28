package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
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
			id INTEGER PRIMARY KEY, username TEXT UNIQUE, email TEXT UNIQUE,
			password_hash TEXT NOT NULL, first_name TEXT, last_name TEXT,
			date_of_birth TEXT, avatar_path TEXT, nickname TEXT, about_me TEXT,
			is_public_profile INTEGER NOT NULL, created_at DATETIME
		);
		CREATE TABLE follows (follower_id INTEGER, following_id INTEGER, status TEXT);
		INSERT INTO users (id,username,email,password_hash,first_name,last_name,avatar_path,nickname,is_public_profile,created_at) VALUES
			(7,'ada','ada@example.com','hash','Ada','Lovelace','/ada.png','Enchantress',1,datetime('now')),
			(8,'grace','grace@example.com','hash','Grace','Hopper',NULL,NULL,0,datetime('now')),
			(9,'linus','linus@example.com','hash','Linus','Torvalds',NULL,NULL,1,datetime('now')),
			(42,'current','current@example.com','hash','Current','User',NULL,NULL,1,datetime('now'));
		INSERT INTO follows VALUES (42,7,'accepted'),(42,8,'pending');
	`); err != nil {
		t.Fatal(err)
	}
	handler := NewInternalReadHandlers(services.NewUserService(database, nil, nil))
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

	permissionRequest := httptest.NewRequest(http.MethodGet, "/internal/v1/users/chat/permission?sender_id=42&receiver_id=7", nil)
	permissionRequest.Header.Set(serviceauth.HeaderName, token)
	permissionResponse := httptest.NewRecorder()
	protected.ServeHTTP(permissionResponse, permissionRequest)
	var permissionBody struct {
		Data struct {
			CanStart bool `json:"can_start"`
		} `json:"data"`
	}
	if err := json.Unmarshal(permissionResponse.Body.Bytes(), &permissionBody); err != nil {
		t.Fatal(err)
	}
	if !permissionBody.Data.CanStart {
		t.Fatalf("expected accepted follow to permit public-profile conversation: %s", permissionResponse.Body.String())
	}

	contactsRequest := httptest.NewRequest(http.MethodGet, "/internal/v1/users/chat/contacts?user_id=42&recent_sender_ids=9", nil)
	contactsRequest.Header.Set(serviceauth.HeaderName, token)
	contactsResponse := httptest.NewRecorder()
	protected.ServeHTTP(contactsResponse, contactsRequest)
	var contactsBody struct {
		Data struct {
			Contacts []struct {
				ID               int  `json:"id"`
				IsMessageRequest bool `json:"is_message_request"`
			} `json:"contacts"`
		} `json:"data"`
	}
	if err := json.Unmarshal(contactsResponse.Body.Bytes(), &contactsBody); err != nil {
		t.Fatal(err)
	}
	if len(contactsBody.Data.Contacts) != 2 ||
		contactsBody.Data.Contacts[0].ID != 7 ||
		contactsBody.Data.Contacts[1].ID != 9 ||
		!contactsBody.Data.Contacts[1].IsMessageRequest {
		t.Fatalf("contacts = %+v", contactsBody.Data.Contacts)
	}

	provisionRequest := httptest.NewRequest(http.MethodPost, "/internal/v1/users/profiles",
		strings.NewReader(`{"account_id":10,"username":"new","email":"new@example.com","first_name":"New","last_name":"User","date_of_birth":"2000-01-01"}`))
	provisionRequest.Header.Set(serviceauth.HeaderName, token)
	provisionResponse := httptest.NewRecorder()
	protected.ServeHTTP(provisionResponse, provisionRequest)
	if provisionResponse.Code != http.StatusOK {
		t.Fatalf("provision response = %d, %s", provisionResponse.Code, provisionResponse.Body.String())
	}
	var provisionedUsername, provisionedPassword string
	if err := database.QueryRow("SELECT username,password_hash FROM users WHERE id=10").Scan(&provisionedUsername, &provisionedPassword); err != nil {
		t.Fatal(err)
	}
	if provisionedUsername != "new" || provisionedPassword != "" {
		t.Fatalf("provisioned profile = %q, password placeholder %q", provisionedUsername, provisionedPassword)
	}
}
