package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"testing"
	"time"

	"social-network/services/notifications/models"
)

type notificationRealtime struct {
	mu        sync.Mutex
	handler   func([]byte)
	ready     chan struct{}
	readyOnce sync.Once
	published chan []byte
}

func newNotificationRealtime() *notificationRealtime {
	return &notificationRealtime{
		ready:     make(chan struct{}),
		published: make(chan []byte, 4),
	}
}

func (f *notificationRealtime) Publish(_ context.Context, data []byte) error {
	f.published <- append([]byte(nil), data...)
	return nil
}

func (f *notificationRealtime) Subscribe(ctx context.Context, handler func([]byte)) error {
	f.mu.Lock()
	f.handler = handler
	f.mu.Unlock()
	f.readyOnce.Do(func() { close(f.ready) })
	<-ctx.Done()
	return nil
}

func (f *notificationRealtime) MarkOnline(context.Context, int) error       { return nil }
func (f *notificationRealtime) MarkOffline(context.Context, int) error      { return nil }
func (f *notificationRealtime) IsOnline(context.Context, int) (bool, error) { return false, nil }
func (f *notificationRealtime) PresenceRefreshInterval() time.Duration      { return time.Hour }

func (f *notificationRealtime) emit(data []byte) {
	f.mu.Lock()
	handler := f.handler
	f.mu.Unlock()
	handler(data)
}

func TestNotificationHubDeliversLocalAndRemoteEventsToAllSockets(t *testing.T) {
	transport := newNotificationRealtime()
	hub := NewNotificationHub(nil, func(*http.Request) bool { return true }, transport)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)
	<-transport.ready

	first := &NotificationClient{hub: hub, send: make(chan []byte, 2), userID: 7}
	second := &NotificationClient{hub: hub, send: make(chan []byte, 2), userID: 7}
	hub.register <- first
	hub.register <- second

	local := &models.Notification{ID: 1, UserID: 7, Type: models.TypeFollow}
	hub.BroadcastNotification(local)
	assertNotificationID(t, first.send, 1)
	assertNotificationID(t, second.send, 1)
	select {
	case <-transport.published:
	case <-time.After(time.Second):
		t.Fatal("notification was not published")
	}

	remote, _ := json.Marshal(&models.Notification{ID: 2, UserID: 7, Type: models.TypeComment})
	transport.emit(remote)
	assertNotificationID(t, first.send, 2)
	assertNotificationID(t, second.send, 2)
}

func assertNotificationID(t *testing.T, messages <-chan []byte, id int) {
	t.Helper()
	select {
	case data := <-messages:
		var message models.WebSocketNotification
		if err := json.Unmarshal(data, &message); err != nil {
			t.Fatalf("unmarshal notification: %v", err)
		}
		if message.Notification.ID != id {
			t.Fatalf("notification ID = %d, want %d", message.Notification.ID, id)
		}
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for notification %d", id)
	}
}
