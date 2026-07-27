package objectstore

import (
	"strings"
	"testing"
)

func TestFromLookup(t *testing.T) {
	config, err := FromLookup(mapLookup(map[string]string{
		"OBJECT_STORAGE_ENDPOINT":        "media:9000",
		"OBJECT_STORAGE_ACCESS_KEY":      "local-user",
		"OBJECT_STORAGE_SECRET_KEY":      "local-password",
		"OBJECT_STORAGE_BUCKET":          "social-network-media",
		"OBJECT_STORAGE_USE_TLS":         "false",
		"OBJECT_STORAGE_PUBLIC_BASE_URL": "/media/social-network-media/",
	}))
	if err != nil {
		t.Fatalf("FromLookup returned an error: %v", err)
	}
	if config.UseTLS || config.PublicBaseURL != "/media/social-network-media" {
		t.Fatalf("unexpected config: %+v", config)
	}
}

func TestFromLookupRejectsEndpointURL(t *testing.T) {
	_, err := FromLookup(mapLookup(map[string]string{
		"OBJECT_STORAGE_ENDPOINT":        "http://media:9000",
		"OBJECT_STORAGE_ACCESS_KEY":      "local-user",
		"OBJECT_STORAGE_SECRET_KEY":      "local-password",
		"OBJECT_STORAGE_BUCKET":          "social-network-media",
		"OBJECT_STORAGE_PUBLIC_BASE_URL": "/media/social-network-media",
	}))
	if err == nil || !strings.Contains(err.Error(), "without a URL scheme") {
		t.Fatalf("expected endpoint validation error, got %v", err)
	}
}

func TestURLRoundTrip(t *testing.T) {
	store := &S3Store{publicBaseURL: "/media/social-network-media"}
	key := "chat/users/7/opaque-name.png"
	objectURL := store.URL(key)
	if objectURL != "/media/social-network-media/chat/users/7/opaque-name.png" {
		t.Fatalf("unexpected URL: %q", objectURL)
	}
	roundTrip, err := store.KeyFromURL(objectURL)
	if err != nil {
		t.Fatalf("KeyFromURL returned an error: %v", err)
	}
	if roundTrip != key {
		t.Fatalf("round trip key = %q, want %q", roundTrip, key)
	}
}

func TestKeyFromURLRejectsDifferentPrefix(t *testing.T) {
	store := &S3Store{publicBaseURL: "/media/social-network-media"}
	if _, err := store.KeyFromURL("/media/other-bucket/chat/image.png"); err == nil {
		t.Fatal("expected an error for a different media prefix")
	}
}

func TestKeyFromURLRejectsAbsoluteURLForRelativeBase(t *testing.T) {
	store := &S3Store{publicBaseURL: "/media/social-network-media"}
	if _, err := store.KeyFromURL("https://evil.example/media/social-network-media/chat/image.png"); err == nil {
		t.Fatal("expected an error for an absolute URL with a relative public base")
	}
}

func mapLookup(values map[string]string) func(string) (string, bool) {
	return func(name string) (string, bool) {
		value, ok := values[name]
		return value, ok
	}
}
