package notify

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"social-network/services/common/serviceauth"
)

func TestCreateNotificationAuthenticatesInternalRequest(t *testing.T) {
	expectedToken := strings.Repeat("n", 32)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/notifications" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get(serviceauth.HeaderName); got != expectedToken {
			t.Errorf("internal token = %q", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("content type = %q", got)
		}

		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		if payload["user_id"] != float64(7) || payload["type"] != "message" {
			t.Errorf("unexpected payload: %#v", payload)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	previousURL := notificationServiceURL
	previousToken := internalServiceToken
	previousClient := notificationHTTPClient
	notificationServiceURL = server.URL
	internalServiceToken = expectedToken
	notificationHTTPClient = &http.Client{Timeout: time.Second}
	t.Cleanup(func() {
		notificationServiceURL = previousURL
		internalServiceToken = previousToken
		notificationHTTPClient = previousClient
	})

	if err := createNotification(7, "message", "New message", 42); err != nil {
		t.Fatalf("create notification: %v", err)
	}
}

func TestValidateConfigRejectsMissingCredential(t *testing.T) {
	previousToken := internalServiceToken
	internalServiceToken = ""
	t.Cleanup(func() { internalServiceToken = previousToken })

	if err := ValidateConfig(); err == nil {
		t.Fatal("expected missing credential to be rejected")
	}
}
