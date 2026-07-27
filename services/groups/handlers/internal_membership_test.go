package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"social-network/services/common/serviceauth"
	"social-network/services/groups/services"

	_ "github.com/mattn/go-sqlite3"
)

func TestInternalMembershipContractIsAuthenticatedAndAcceptedOnly(t *testing.T) {
	const token = "test-internal-token-at-least-32-characters"
	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	database.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = database.Close() })
	if _, err := database.Exec(`
		CREATE TABLE group_members (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			group_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			status TEXT NOT NULL
		);
		INSERT INTO group_members (group_id, user_id, status) VALUES
			(7, 42, 'accepted'),
			(7, 43, 'pending'),
			(7, 44, 'accepted');
	`); err != nil {
		t.Fatalf("create fixtures: %v", err)
	}

	handler := NewInternalMembershipHandlers(services.NewGroupService(database))
	protected := serviceauth.Authenticate(token, http.HandlerFunc(handler.GetMembership))

	unauthorized := httptest.NewRecorder()
	protected.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/internal/v1/groups/7/members", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d", unauthorized.Code)
	}

	memberRequest := httptest.NewRequest(http.MethodGet, "/internal/v1/groups/7/members/42", nil)
	memberRequest.Header.Set(serviceauth.HeaderName, token)
	memberResponse := httptest.NewRecorder()
	protected.ServeHTTP(memberResponse, memberRequest)
	var memberBody struct {
		Data struct {
			IsMember bool `json:"is_member"`
		} `json:"data"`
	}
	if err := json.Unmarshal(memberResponse.Body.Bytes(), &memberBody); err != nil {
		t.Fatalf("decode membership response: %v", err)
	}
	if memberResponse.Code != http.StatusOK || !memberBody.Data.IsMember {
		t.Fatalf("membership response = %d, %s", memberResponse.Code, memberResponse.Body.String())
	}

	listRequest := httptest.NewRequest(http.MethodGet, "/internal/v1/groups/7/members", nil)
	listRequest.Header.Set(serviceauth.HeaderName, token)
	listResponse := httptest.NewRecorder()
	protected.ServeHTTP(listResponse, listRequest)
	var listBody struct {
		Data struct {
			MemberIDs []int `json:"member_ids"`
		} `json:"data"`
	}
	if err := json.Unmarshal(listResponse.Body.Bytes(), &listBody); err != nil {
		t.Fatalf("decode member list: %v", err)
	}
	if !reflect.DeepEqual(listBody.Data.MemberIDs, []int{42, 44}) {
		t.Fatalf("member IDs = %v", listBody.Data.MemberIDs)
	}
}
