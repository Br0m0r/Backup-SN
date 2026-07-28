package usersclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"social-network/services/common/serviceauth"
)

func TestClientReadsAuthenticatedProfiles(t *testing.T) {
	const token = "test-internal-token-at-least-32-characters"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(serviceauth.HeaderName) != token {
			t.Fatal("missing service token")
		}
		if r.URL.Path != "/internal/v1/users/profiles" || r.URL.Query().Get("ids") != "7,8" {
			t.Fatalf("request = %s?%s", r.URL.Path, r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"success":true,"data":{"profiles":[{"id":7,"username":"ada","first_name":"Ada","last_name":"Lovelace","nickname":"Enchantress"}]}}`))
	}))
	defer server.Close()
	t.Setenv("USERS_SERVICE_URL", server.URL)

	client, err := FromEnvironment(token)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	profiles, err := client.Profiles(context.Background(), []int{7, 8})
	if err != nil || len(profiles) != 1 || profiles[0].Nickname == nil || *profiles[0].Nickname != "Enchantress" {
		t.Fatalf("Profiles()=%v, %v", profiles, err)
	}
}
