package usersclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"social-network/services/common/serviceauth"
)

func TestProvisionUsesAuthenticatedUsersContract(t *testing.T) {
	const token = "test-internal-token-at-least-32-characters"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/internal/v1/users/profiles" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get(serviceauth.HeaderName) != token {
			t.Fatal("missing service token")
		}
		_, _ = w.Write([]byte(`{"success":true,"data":{"profile_id":7}}`))
	}))
	defer server.Close()
	t.Setenv("USERS_SERVICE_URL", server.URL)
	client, err := FromEnvironment(token)
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Provision(context.Background(), Profile{
		AccountID: 7, Username: "ada", Email: "ada@example.com",
	}); err != nil {
		t.Fatal(err)
	}
}
