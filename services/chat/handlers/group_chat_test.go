package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"social-network/services/chat/middleware"
	"social-network/services/chat/models"

	_ "github.com/mattn/go-sqlite3"
)

func TestChatOwnsGroupMessageHTTPFlow(t *testing.T) {
	database := openGroupChatTestDatabase(t)
	membership := fakeGroupMembership{members: map[int][]int{7: {42}}}
	hub := NewHub(database, database, nil, nil, membership)
	handler := NewChatHandlers(database, database, hub, membership)

	sendRequest := httptest.NewRequest(
		http.MethodPost,
		"/chat/groups/7/messages",
		bytes.NewBufferString(`{"content":"hello group"}`),
	)
	sendRequest = middleware.SetUserIDInContext(sendRequest, 42)
	sendResponse := httptest.NewRecorder()
	handler.SendGroupMessage(sendResponse, sendRequest)
	if sendResponse.Code != http.StatusOK {
		t.Fatalf("send status = %d, body = %s", sendResponse.Code, sendResponse.Body.String())
	}

	var storedCount int
	if err := database.QueryRow(`SELECT COUNT(*) FROM group_messages WHERE group_id = 7 AND sender_id = 42 AND content = 'hello group'`).Scan(&storedCount); err != nil {
		t.Fatalf("count stored messages: %v", err)
	}
	if storedCount != 1 {
		t.Fatalf("stored message count = %d, want 1", storedCount)
	}

	historyRequest := httptest.NewRequest(http.MethodGet, "/chat/groups/7/history?limit=20", nil)
	historyRequest = middleware.SetUserIDInContext(historyRequest, 42)
	historyResponse := httptest.NewRecorder()
	handler.GetGroupChatHistory(historyResponse, historyRequest)
	if historyResponse.Code != http.StatusOK {
		t.Fatalf("history status = %d, body = %s", historyResponse.Code, historyResponse.Body.String())
	}

	var body struct {
		Success bool `json:"success"`
		Data    struct {
			Messages []models.GroupMessage `json:"messages"`
			Count    int                   `json:"count"`
		} `json:"data"`
	}
	if err := json.Unmarshal(historyResponse.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode history response: %v", err)
	}
	if !body.Success || body.Data.Count != 1 || len(body.Data.Messages) != 1 {
		t.Fatalf("unexpected history response: %+v", body)
	}
	if body.Data.Messages[0].Content != "hello group" {
		t.Fatalf("message content = %q", body.Data.Messages[0].Content)
	}
}

func TestChatRejectsGroupMessageFromNonMember(t *testing.T) {
	database := openGroupChatTestDatabase(t)
	membership := fakeGroupMembership{members: map[int][]int{7: {42}}}
	handler := NewChatHandlers(database, database, NewHub(database, database, nil, nil, membership), membership)

	request := httptest.NewRequest(
		http.MethodPost,
		"/chat/groups/7/messages",
		bytes.NewBufferString(`{"content":"not allowed"}`),
	)
	request = middleware.SetUserIDInContext(request, 99)
	response := httptest.NewRecorder()
	handler.SendGroupMessage(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusForbidden, response.Body.String())
	}
}

func openGroupChatTestDatabase(t *testing.T) *sql.DB {
	t.Helper()
	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	database.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = database.Close() })

	schema := `
		CREATE TABLE group_messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			group_id INTEGER NOT NULL,
			sender_id INTEGER NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME NOT NULL
		);
	`
	if _, err := database.Exec(schema); err != nil {
		t.Fatalf("create fixtures: %v", err)
	}
	return database
}

type fakeGroupMembership struct {
	members map[int][]int
}

func (f fakeGroupMembership) IsMember(_ context.Context, groupID, userID int) (bool, error) {
	for _, memberID := range f.members[groupID] {
		if memberID == userID {
			return true, nil
		}
	}
	return false, nil
}

func (f fakeGroupMembership) MemberIDs(_ context.Context, groupID int) ([]int, error) {
	return f.members[groupID], nil
}
