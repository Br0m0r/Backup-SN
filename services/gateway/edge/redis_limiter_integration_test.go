package edge

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"social-network/services/common/redisstore"
)

func TestRedisLimiterSharesBucketAcrossInstances(t *testing.T) {
	url := os.Getenv("REDIS_TEST_URL")
	if url == "" {
		t.Skip("REDIS_TEST_URL is not set")
	}
	store, err := redisstore.Open(context.Background(), redisstore.Config{
		URL:              url,
		Namespace:        fmt.Sprintf("social-network-rate-test-%d", time.Now().UnixNano()),
		OperationTimeout: time.Second,
	})
	if err != nil {
		t.Fatalf("redisstore.Open() error = %v", err)
	}
	defer store.Close()

	config := Config{RatePerSecond: 0.001, Burst: 2, MaxBodyBytes: 1024}
	first := NewRedisLimiter(store, config)
	second := NewRedisLimiter(store, config)
	for attempt, testCase := range []struct {
		limiter *RedisLimiter
		want    bool
	}{
		{limiter: first, want: true},
		{limiter: second, want: true},
		{limiter: first, want: false},
	} {
		got, err := testCase.limiter.Allow(context.Background(), "192.0.2.1")
		if err != nil {
			t.Fatalf("attempt %d Allow() error = %v", attempt+1, err)
		}
		if got != testCase.want {
			t.Fatalf("attempt %d Allow() = %v, want %v", attempt+1, got, testCase.want)
		}
	}
}
