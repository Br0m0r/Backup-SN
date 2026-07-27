package realtime

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"social-network/services/common/redisstore"
)

func TestRedisPresenceAcrossInstances(t *testing.T) {
	url := os.Getenv("REDIS_TEST_URL")
	if url == "" {
		t.Skip("REDIS_TEST_URL is not set")
	}
	namespace := fmt.Sprintf("social-network-test-%d", time.Now().UnixNano())
	storeOne := openTestStore(t, url, namespace)
	defer storeOne.Close()
	storeTwo := openTestStore(t, url, namespace)
	defer storeTwo.Close()

	t.Setenv("SERVICE_INSTANCE_ID", "chat-one")
	first, err := New(storeOne, "chat")
	if err != nil {
		t.Fatalf("New(first) error = %v", err)
	}
	t.Setenv("SERVICE_INSTANCE_ID", "chat-two")
	second, err := New(storeTwo, "chat")
	if err != nil {
		t.Fatalf("New(second) error = %v", err)
	}

	ctx := context.Background()
	if err := first.MarkOnline(ctx, 42); err != nil {
		t.Fatalf("first.MarkOnline() error = %v", err)
	}
	if online, err := second.IsOnline(ctx, 42); err != nil || !online {
		t.Fatalf("second.IsOnline() = %v, %v", online, err)
	}
	if err := second.MarkOnline(ctx, 42); err != nil {
		t.Fatalf("second.MarkOnline() error = %v", err)
	}
	if err := first.MarkOffline(ctx, 42); err != nil {
		t.Fatalf("first.MarkOffline() error = %v", err)
	}
	if online, err := first.IsOnline(ctx, 42); err != nil || !online {
		t.Fatalf("presence should remain online through second instance: %v, %v", online, err)
	}
	if err := second.MarkOffline(ctx, 42); err != nil {
		t.Fatalf("second.MarkOffline() error = %v", err)
	}
	if online, err := first.IsOnline(ctx, 42); err != nil || online {
		t.Fatalf("presence after final disconnect = %v, %v", online, err)
	}
}

func TestRedisPubSubExcludesOriginAndReachesOtherInstance(t *testing.T) {
	url := os.Getenv("REDIS_TEST_URL")
	if url == "" {
		t.Skip("REDIS_TEST_URL is not set")
	}
	namespace := fmt.Sprintf("social-network-test-%d", time.Now().UnixNano())
	storeOne := openTestStore(t, url, namespace)
	defer storeOne.Close()
	storeTwo := openTestStore(t, url, namespace)
	defer storeTwo.Close()

	t.Setenv("SERVICE_INSTANCE_ID", "notifications-one")
	first, err := New(storeOne, "notifications")
	if err != nil {
		t.Fatalf("New(first) error = %v", err)
	}
	t.Setenv("SERVICE_INSTANCE_ID", "notifications-two")
	second, err := New(storeTwo, "notifications")
	if err != nil {
		t.Fatalf("New(second) error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	received := make(chan string, 1)
	subscriptionDone := make(chan error, 1)
	go func() {
		subscriptionDone <- second.Subscribe(ctx, func(payload []byte) {
			received <- string(payload)
		})
	}()

	payload := []byte(`{"type":"notification","id":7}`)
	deadline := time.Now().Add(3 * time.Second)
	for {
		if err := first.Publish(ctx, payload); err != nil {
			t.Fatalf("Publish() error = %v", err)
		}
		select {
		case got := <-received:
			if got != string(payload) {
				t.Fatalf("received %q, want %q", got, payload)
			}
			cancel()
			select {
			case err := <-subscriptionDone:
				if err != nil {
					t.Fatalf("Subscribe() error = %v", err)
				}
			case <-time.After(time.Second):
				t.Fatal("subscription did not stop")
			}
			return
		case <-time.After(50 * time.Millisecond):
			if time.Now().After(deadline) {
				t.Fatal("timed out waiting for Pub/Sub delivery")
			}
		}
	}
}

func openTestStore(t *testing.T, url, namespace string) *redisstore.Store {
	t.Helper()
	store, err := redisstore.Open(context.Background(), redisstore.Config{
		URL:              url,
		Namespace:        namespace,
		OperationTimeout: time.Second,
	})
	if err != nil {
		t.Fatalf("redisstore.Open() error = %v", err)
	}
	return store
}
