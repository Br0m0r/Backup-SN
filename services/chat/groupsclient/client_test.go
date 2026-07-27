package groupsclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"social-network/services/common/serviceauth"
)

func TestClientUsesAuthenticatedVersionedMembershipContract(t *testing.T) {
	const token = "test-internal-token-at-least-32-characters"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(serviceauth.HeaderName) != token {
			t.Fatal("internal service token was not sent")
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/internal/v1/groups/7/members/42":
			_, _ = w.Write([]byte(`{"success":true,"data":{"is_member":true}}`))
		case "/internal/v1/groups/7/members":
			_, _ = w.Write([]byte(`{"success":true,"data":{"member_ids":[42,43]}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	t.Setenv("GROUPS_SERVICE_URL", server.URL+"/")
	t.Setenv("GROUPS_SERVICE_TIMEOUT", "1s")
	client, err := FromEnvironment(token)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	isMember, err := client.IsMember(context.Background(), 7, 42)
	if err != nil || !isMember {
		t.Fatalf("IsMember() = %v, %v", isMember, err)
	}
	memberIDs, err := client.MemberIDs(context.Background(), 7)
	if err != nil {
		t.Fatalf("MemberIDs(): %v", err)
	}
	if !reflect.DeepEqual(memberIDs, []int{42, 43}) {
		t.Fatalf("MemberIDs() = %v", memberIDs)
	}
}

func TestClientRejectsInvalidConfigurationAndResponses(t *testing.T) {
	t.Setenv("GROUPS_SERVICE_URL", "relative")
	if _, err := FromEnvironment("token"); err == nil {
		t.Fatal("expected invalid URL error")
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"success":false,"error":"no"}`, http.StatusUnauthorized)
	}))
	defer server.Close()
	t.Setenv("GROUPS_SERVICE_URL", server.URL)
	client, err := FromEnvironment("token")
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	if _, err := client.IsMember(context.Background(), 7, 42); err == nil {
		t.Fatal("expected upstream response error")
	}
}
