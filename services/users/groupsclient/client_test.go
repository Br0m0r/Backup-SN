package groupsclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"social-network/services/common/serviceauth"
)

func TestClientReadsAuthenticatedParticipants(t *testing.T) {
	const token = "test-internal-token-at-least-32-characters"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(serviceauth.HeaderName) != token {
			t.Fatal("missing service token")
		}
		if r.URL.Path != "/internal/v1/groups/7/participants" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"success":true,"data":{"participant_ids":[2,4,8]}}`))
	}))
	defer server.Close()
	t.Setenv("GROUPS_SERVICE_URL", server.URL)

	client, err := FromEnvironment(token)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	participants, err := client.ParticipantIDs(context.Background(), 7)
	if err != nil || !reflect.DeepEqual(participants, []int{2, 4, 8}) {
		t.Fatalf("ParticipantIDs()=%v, %v", participants, err)
	}
}
