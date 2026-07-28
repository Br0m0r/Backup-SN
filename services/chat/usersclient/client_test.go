package usersclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"social-network/services/common/serviceauth"
)

func TestClientReadsAuthenticatedChatIdentityContracts(t *testing.T) {
	const token = "test-internal-token-at-least-32-characters"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(serviceauth.HeaderName) != token {
			t.Fatal("missing service token")
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/internal/v1/users/profiles":
			_, _ = w.Write([]byte(`{"success":true,"data":{"profiles":[{"id":7,"username":"ada"}]}}`))
		case "/internal/v1/users/chat/permission":
			_, _ = w.Write([]byte(`{"success":true,"data":{"can_start":true}}`))
		case "/internal/v1/users/chat/contacts":
			_, _ = w.Write([]byte(`{"success":true,"data":{"contacts":[{"id":8,"username":"grace","is_message_request":true}]}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	t.Setenv("USERS_SERVICE_URL", server.URL)
	client, err := FromEnvironment(token)
	if err != nil {
		t.Fatal(err)
	}
	profiles, err := client.Profiles(context.Background(), []int{7})
	if err != nil || len(profiles) != 1 || profiles[0].Username != "ada" {
		t.Fatalf("Profiles()=%v, %v", profiles, err)
	}
	allowed, err := client.CanStartConversation(context.Background(), 42, 7)
	if err != nil || !allowed {
		t.Fatalf("CanStartConversation()=%v, %v", allowed, err)
	}
	contacts, err := client.Contacts(context.Background(), 42, []int{8})
	if err != nil || len(contacts) != 1 || !contacts[0].IsMessageRequest {
		t.Fatalf("Contacts()=%v, %v", contacts, err)
	}
}
