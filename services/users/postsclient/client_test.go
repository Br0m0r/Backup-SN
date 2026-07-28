package postsclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"social-network/services/common/serviceauth"
)

func TestClientReadsAuthenticatedUserPosts(t *testing.T) {
	const token = "test-internal-token-at-least-32-characters"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(serviceauth.HeaderName) != token {
			t.Fatal("missing service token")
		}
		if r.URL.Path != "/internal/v1/users/7/posts" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"success":true,"data":{"posts":[{"id":9,"user_id":7,"content":"hello","privacy_level":"public","created_at":"2026-01-01T00:00:00Z"}]}}`))
	}))
	defer server.Close()
	t.Setenv("POSTS_SERVICE_URL", server.URL)
	client, err := FromEnvironment(token)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	posts, err := client.UserPosts(context.Background(), 7)
	if err != nil || len(posts) != 1 || posts[0].ID != 9 {
		t.Fatalf("UserPosts()=%v, %v", posts, err)
	}
}
