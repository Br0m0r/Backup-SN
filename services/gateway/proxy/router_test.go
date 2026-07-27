package proxy

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"social-network/services/common/observability"
	"social-network/services/common/serviceauth"
	"social-network/services/gateway/edge"
)

func TestRouterRoutesAndStripsPrivateCredentials(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get(serviceauth.HeaderName); got != "" {
			t.Errorf("private credential reached upstream: %q", got)
		}
		w.Header().Set("Content-Type", "text/plain")
		_, _ = fmt.Fprintf(w, "%s?%s", r.URL.Path, r.URL.RawQuery)
	}))
	defer upstream.Close()

	target := mustParseURL(t, upstream.URL)
	gateway := httptest.NewServer(NewRouter(allTargets(target), edge.New(edge.DefaultConfig())))
	defer gateway.Close()

	request, err := http.NewRequest(http.MethodGet, gateway.URL+"/api/posts/posts/feed?page=2", nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	request.Header.Set(serviceauth.HeaderName, strings.Repeat("x", 32))
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("gateway request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", response.StatusCode)
	}
	if got := response.Header.Get(observability.RequestIDHeader); got == "" {
		t.Fatal("gateway response has no request ID")
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	if got, want := string(body), "/posts/feed?page=2"; got != want {
		t.Fatalf("upstream request = %q, want %q", got, want)
	}
}

func TestRouterSendsNonAPIRoutesToFrontend(t *testing.T) {
	frontend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("frontend:" + r.URL.Path))
	}))
	defer frontend.Close()

	target := mustParseURL(t, frontend.URL)
	gateway := httptest.NewServer(NewRouter(allTargets(target), edge.New(edge.DefaultConfig())))
	defer gateway.Close()

	response, err := http.Get(gateway.URL + "/profile/7")
	if err != nil {
		t.Fatalf("gateway request: %v", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	if got, want := string(body), "frontend:/profile/7"; got != want {
		t.Fatalf("frontend response = %q, want %q", got, want)
	}
}

func TestRouterProxiesMediaReadsAndRejectsWrites(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Method + ":" + r.URL.Path))
	}))
	defer upstream.Close()

	target := mustParseURL(t, upstream.URL)
	gateway := httptest.NewServer(NewRouter(allTargets(target), edge.New(edge.DefaultConfig())))
	defer gateway.Close()

	response, err := http.Get(gateway.URL + "/media/social-network-media/chat/image.png")
	if err != nil {
		t.Fatalf("gateway media request: %v", err)
	}
	body, err := io.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		t.Fatalf("read media response: %v", err)
	}
	if got, want := string(body), "GET:/social-network-media/chat/image.png"; got != want {
		t.Fatalf("media upstream request = %q, want %q", got, want)
	}

	request, err := http.NewRequest(http.MethodPut, gateway.URL+"/media/social-network-media/chat/image.png", strings.NewReader("blocked"))
	if err != nil {
		t.Fatalf("create media write request: %v", err)
	}
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("gateway media write request: %v", err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("media write status = %d, want %d", response.StatusCode, http.StatusMethodNotAllowed)
	}
}

func TestRouterProxiesWebSocketUpgrade(t *testing.T) {
	upgrader := websocket.Upgrader{CheckOrigin: func(_ *http.Request) bool { return true }}
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connection, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer connection.Close()
		_ = connection.WriteMessage(websocket.TextMessage, []byte(r.URL.Path+"?"+r.URL.RawQuery))
	}))
	defer upstream.Close()

	target := mustParseURL(t, upstream.URL)
	gateway := httptest.NewServer(NewRouter(allTargets(target), edge.New(edge.DefaultConfig())))
	defer gateway.Close()

	websocketURL := "ws" + strings.TrimPrefix(gateway.URL, "http") + "/api/chat/ws?token=test"
	connection, _, err := websocket.DefaultDialer.Dial(websocketURL, nil)
	if err != nil {
		t.Fatalf("dial gateway websocket: %v", err)
	}
	defer connection.Close()

	_, message, err := connection.ReadMessage()
	if err != nil {
		t.Fatalf("read websocket message: %v", err)
	}
	if got, want := string(message), "/ws?token=test"; got != want {
		t.Fatalf("upstream websocket path = %q, want %q", got, want)
	}
}

func TestTargetsFromEnvironmentRejectsInvalidURL(t *testing.T) {
	t.Setenv("AUTH_SERVICE_URL", "://invalid")
	if _, err := TargetsFromEnvironment(); err == nil {
		t.Fatal("expected invalid upstream URL to be rejected")
	}
}

func mustParseURL(t *testing.T, value string) *url.URL {
	t.Helper()
	parsed, err := url.Parse(value)
	if err != nil {
		t.Fatalf("parse URL: %v", err)
	}
	return parsed
}

func allTargets(target *url.URL) Targets {
	return Targets{
		Auth:          target,
		Users:         target,
		Posts:         target,
		Groups:        target,
		Chat:          target,
		Notifications: target,
		Media:         target,
		Frontend:      target,
	}
}
