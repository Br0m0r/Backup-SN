package usersclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"social-network/services/common/serviceauth"
)

func TestClientReadsAuthenticatedUsersContract(t *testing.T) {
	const token = "test-internal-token-at-least-32-characters"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(serviceauth.HeaderName) != token {
			t.Fatal("missing service token")
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/internal/v1/users/profiles":
			_, _ = w.Write([]byte(`{"success":true,"data":{"profiles":[{"id":7,"username":"ada"}]}}`))
		case "/internal/v1/users/42/following":
			_, _ = w.Write([]byte(`{"success":true,"data":{"following_ids":[7,8]}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	t.Setenv("USERS_SERVICE_URL", server.URL)
	client, err := FromEnvironment(token)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	profiles, err := client.Profiles(context.Background(), []int{7})
	if err != nil || len(profiles) != 1 || profiles[0].Username != "ada" {
		t.Fatalf("Profiles()=%v, %v", profiles, err)
	}
	following, err := client.FollowingIDs(context.Background(), 42)
	if err != nil || !reflect.DeepEqual(following, []int{7, 8}) {
		t.Fatalf("FollowingIDs()=%v, %v", following, err)
	}
}
