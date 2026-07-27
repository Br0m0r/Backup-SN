package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"testing"
	"time"

	"social-network/services/chat/models"
)

type fakeRealtime struct {
	mu        sync.Mutex
	online    map[int]bool
	published chan []byte
	ready     chan struct{}
	handler   func([]byte)
	readyOnce sync.Once
}

func newFakeRealtime() *fakeRealtime {
	return &fakeRealtime{
		online:    make(map[int]bool),
		published: make(chan []byte, 8),
		ready:     make(chan struct{}),
	}
}

func (f *fakeRealtime) Publish(_ context.Context, data []byte) error {
	f.published <- append([]byte(nil), data...)
	return nil
}

func (f *fakeRealtime) Subscribe(ctx context.Context, handler func([]byte)) error {
	f.mu.Lock()
	f.handler = handler
	f.mu.Unlock()
	f.readyOnce.Do(func() { close(f.ready) })
	<-ctx.Done()
	return nil
}

func (f *fakeRealtime) MarkOnline(_ context.Context, userID int) error {
	f.mu.Lock()
	f.online[userID] = true
	f.mu.Unlock()
	return nil
}

func (f *fakeRealtime) MarkOffline(_ context.Context, userID int) error {
	f.mu.Lock()
	delete(f.online, userID)
	f.mu.Unlock()
	return nil
}

func (f *fakeRealtime) IsOnline(_ context.Context, userID int) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.online[userID], nil
}

func (f *fakeRealtime) PresenceRefreshInterval() time.Duration {
	return time.Hour
}

func (f *fakeRealtime) emit(data []byte) {
	f.mu.Lock()
	handler := f.handler
	f.mu.Unlock()
	handler(data)
}

func TestHubSupportsMultipleConnectionsAndRemoteFanout(t *testing.T) {
	transport := newFakeRealtime()
	hub := NewHub(nil, func(*http.Request) bool { return true }, transport)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)
	<-transport.ready

	first := &Client{hub: hub, send: make(chan []byte, 4), userID: 42}
	second := &Client{hub: hub, send: make(chan []byte, 4), userID: 42}
	hub.register <- first
	hub.register <- second
	eventually(t, func() bool {
		hub.mu.RLock()
		defer hub.mu.RUnlock()
		return len(hub.clients[42]) == 2
	})

	local := &models.WebSocketMessage{Type: "message", ReceiverID: 42, Content: "local"}
	hub.Broadcast(local)
	assertChatPayload(t, first.send, "local")
	assertChatPayload(t, second.send, "local")
	select {
	case <-transport.published:
	case <-time.After(time.Second):
		t.Fatal("local message was not published")
	}

	remote, _ := json.Marshal(&models.WebSocketMessage{Type: "message", ReceiverID: 42, Content: "remote"})
	transport.emit(remote)
	assertChatPayload(t, first.send, "remote")
	assertChatPayload(t, second.send, "remote")

	hub.unregister <- first
	eventually(t, func() bool {
		hub.mu.RLock()
		defer hub.mu.RUnlock()
		return len(hub.clients[42]) == 1
	})
	if !hub.IsUserOnline(42) {
		t.Fatal("user should remain online while the second socket is connected")
	}
	hub.unregister <- second
	eventually(t, func() bool { return !hub.IsUserOnline(42) })
}

func assertChatPayload(t *testing.T, messages <-chan []byte, content string) {
	t.Helper()
	select {
	case data := <-messages:
		var message models.WebSocketMessage
		if err := json.Unmarshal(data, &message); err != nil {
			t.Fatalf("unmarshal message: %v", err)
		}
		if message.Content != content {
			t.Fatalf("content = %q, want %q", message.Content, content)
		}
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for %q", content)
	}
}

func eventually(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for !condition() {
		if time.Now().After(deadline) {
			t.Fatal("condition was not satisfied")
		}
		time.Sleep(time.Millisecond)
	}
}
