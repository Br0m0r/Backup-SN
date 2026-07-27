package redisstore

import (
	"context"
	"testing"
	"time"
)

func TestFromEnvironmentRequiresURL(t *testing.T) {
	t.Setenv("REDIS_URL", "")
	if _, err := FromEnvironment(); err == nil {
		t.Fatal("expected REDIS_URL to be required")
	}
}

func TestFromEnvironmentLoadsDefaultsAndOverrides(t *testing.T) {
	t.Setenv("REDIS_URL", "redis://localhost:6379/2")
	t.Setenv("REDIS_NAMESPACE", "test-network")
	t.Setenv("REDIS_OPERATION_TIMEOUT", "750ms")

	config, err := FromEnvironment()
	if err != nil {
		t.Fatalf("FromEnvironment() error = %v", err)
	}
	if config.Namespace != "test-network" {
		t.Fatalf("Namespace = %q", config.Namespace)
	}
	if config.OperationTimeout != 750*time.Millisecond {
		t.Fatalf("OperationTimeout = %s", config.OperationTimeout)
	}
}

func TestKeyNormalizesSeparators(t *testing.T) {
	store := &Store{namespace: "social-network", operationTimeout: time.Second}
	if got := store.Key(":presence:", " chat ", "42"); got != "social-network:presence:chat:42" {
		t.Fatalf("Key() = %q", got)
	}
}

func TestContextUsesConfiguredTimeout(t *testing.T) {
	store := &Store{operationTimeout: 10 * time.Millisecond}
	ctx, cancel := store.Context(context.Background())
	defer cancel()
	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("context did not expire")
	}
}
