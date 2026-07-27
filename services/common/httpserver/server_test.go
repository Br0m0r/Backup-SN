package httpserver

import (
	"net/http"
	"testing"
	"time"
)

func TestAddress(t *testing.T) {
	t.Setenv("PORT", "")
	if got := Address("8080"); got != ":8080" {
		t.Fatalf("default address = %q", got)
	}

	t.Setenv("PORT", "9090")
	if got := Address("8080"); got != ":9090" {
		t.Fatalf("configured address = %q", got)
	}
}

func TestNewSetsOperationalTimeouts(t *testing.T) {
	server := New(":8080", http.NewServeMux())
	if server.ReadHeaderTimeout != 5*time.Second || server.ReadTimeout != 15*time.Second {
		t.Fatalf("unexpected read timeouts: header=%v read=%v", server.ReadHeaderTimeout, server.ReadTimeout)
	}
	if server.WriteTimeout != 30*time.Second || server.IdleTimeout != 120*time.Second {
		t.Fatalf("unexpected write/idle timeouts: write=%v idle=%v", server.WriteTimeout, server.IdleTimeout)
	}
	if server.MaxHeaderBytes != 1<<20 {
		t.Fatalf("unexpected max header bytes: %d", server.MaxHeaderBytes)
	}
}
