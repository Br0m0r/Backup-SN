package observability

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPLoggingPreservesRequestIDAndLogsResponse(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))

	handler := HTTPLoggingWithLogger("test-service", logger, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID, ok := RequestIDFromContext(r.Context())
		if !ok || requestID != "client-request-123" {
			t.Fatalf("unexpected request ID in context: %q, %v", requestID, ok)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	}))

	request := httptest.NewRequest(http.MethodPost, "/resources?secret=excluded", nil)
	request.Header.Set(RequestIDHeader, "client-request-123")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if got := response.Header().Get(RequestIDHeader); got != "client-request-123" {
		t.Fatalf("response request ID = %q", got)
	}

	var entry map[string]any
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("decode structured log: %v", err)
	}
	if entry["service"] != "test-service" || entry["path"] != "/resources" {
		t.Fatalf("unexpected structured log: %#v", entry)
	}
	if entry["status"] != float64(http.StatusCreated) || entry["bytes"] != float64(len("created")) {
		t.Fatalf("unexpected response fields: %#v", entry)
	}
}

func TestHTTPLoggingReplacesUnsafeRequestID(t *testing.T) {
	handler := HTTPLoggingWithLogger("test-service", slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil)), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set(RequestIDHeader, "unsafe request id\nvalue")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	requestID := response.Header().Get(RequestIDHeader)
	if requestID == "" || requestID == "unsafe request id\nvalue" {
		t.Fatalf("unsafe request ID was not replaced: %q", requestID)
	}
}
