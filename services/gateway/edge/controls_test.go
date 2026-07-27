package edge

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type stubLimiter struct {
	allowed bool
	err     error
}

func (s stubLimiter) Allow(context.Context, string) (bool, error) {
	return s.allowed, s.err
}

func TestControlsRateLimitAPIByClientIP(t *testing.T) {
	controls := New(Config{RatePerSecond: 0.001, Burst: 2, MaxBodyBytes: 1024})
	fixedTime := time.Unix(100, 0)
	controls.now = func() time.Time { return fixedTime }
	handler := controls.Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for attempt, want := range []int{http.StatusNoContent, http.StatusNoContent, http.StatusTooManyRequests} {
		request := httptest.NewRequest(http.MethodGet, "/api/posts/posts/feed", nil)
		request.RemoteAddr = "192.0.2.1:1234"
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != want {
			t.Fatalf("attempt %d status = %d, want %d", attempt+1, response.Code, want)
		}
	}

	request := httptest.NewRequest(http.MethodGet, "/api/posts/posts/feed", nil)
	request.RemoteAddr = "192.0.2.2:5678"
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("independent client status = %d", response.Code)
	}
}

func TestControlsRejectOversizedKnownBody(t *testing.T) {
	controls := New(Config{RatePerSecond: 10, Burst: 10, MaxBodyBytes: 4})
	handler := controls.Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodPost, "/api/posts/posts", bytes.NewBufferString("12345"))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)
	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusRequestEntityTooLarge)
	}
}

func TestControlsUsesDistributedLimiter(t *testing.T) {
	controls := NewDistributed(
		Config{RatePerSecond: 10, Burst: 10, MaxBodyBytes: 1024},
		stubLimiter{allowed: false},
	)
	handler := controls.Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/api/posts/posts/feed", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)
	if response.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusTooManyRequests)
	}
}

func TestControlsFallsBackLocallyWhenRedisFails(t *testing.T) {
	controls := NewDistributed(
		Config{RatePerSecond: 0.001, Burst: 1, MaxBodyBytes: 1024},
		stubLimiter{err: errors.New("redis unavailable")},
	)
	controls.now = func() time.Time { return time.Unix(100, 0) }
	handler := controls.Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for attempt, want := range []int{http.StatusNoContent, http.StatusTooManyRequests} {
		request := httptest.NewRequest(http.MethodGet, "/api/posts/posts/feed", nil)
		request.RemoteAddr = "192.0.2.10:1234"
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != want {
			t.Fatalf("attempt %d status = %d, want %d", attempt+1, response.Code, want)
		}
	}
}

func TestControlsSetSecurityHeadersOnFrontend(t *testing.T) {
	controls := New(DefaultConfig())
	handler := controls.Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)
	for _, header := range []string{"Content-Security-Policy", "X-Content-Type-Options", "X-Frame-Options", "Referrer-Policy", "Permissions-Policy"} {
		if response.Header().Get(header) == "" {
			t.Fatalf("missing security header %s", header)
		}
	}
}

func TestConfigFromEnvironmentRejectsInvalidValues(t *testing.T) {
	t.Setenv("GATEWAY_RATE_LIMIT_RPS", "0")
	if _, err := ConfigFromEnvironment(); err == nil {
		t.Fatal("expected zero rate to be rejected")
	}
}
