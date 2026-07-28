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
			last_name TEXT, avatar_path TEXT, nickname TEXT,
			is_public_profile INTEGER NOT NULL
		);
		CREATE TABLE follows (follower_id INTEGER, following_id INTEGER, status TEXT);
		INSERT INTO users VALUES
			(7,'ada','Ada','Lovelace','/ada.png','Enchantress',1),
			(8,'grace','Grace','Hopper',NULL,NULL,0),
			(9,'linus','Linus','Torvalds',NULL,NULL,1),
			(42,'current','Current','User',NULL,NULL,1);
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
}
