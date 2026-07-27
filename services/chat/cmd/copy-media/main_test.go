package main

import (
	"context"
	"database/sql"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

type fakeObjectStore struct {
	putKeys     []string
	deletedKeys []string
}

func (store *fakeObjectStore) Put(_ context.Context, key string, reader io.Reader, _ int64, contentType string) error {
	if contentType != "image/png" {
		return io.ErrUnexpectedEOF
	}
	if _, err := io.ReadAll(reader); err != nil {
		return err
	}
	store.putKeys = append(store.putKeys, key)
	return nil
}

func (store *fakeObjectStore) Delete(_ context.Context, key string) error {
	store.deletedKeys = append(store.deletedKeys, key)
	return nil
}

func (store *fakeObjectStore) URL(key string) string {
	return "/media/social-network-media/" + key
}

func (store *fakeObjectStore) KeyFromURL(string) (string, error) { return "", nil }

func TestCopyLegacyMediaUploadsAndRewritesMessage(t *testing.T) {
	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open SQLite: %v", err)
	}
	defer database.Close()
	database.SetMaxOpenConns(1)
	if _, err := database.Exec(`
		CREATE TABLE messages (id INTEGER PRIMARY KEY, sender_id INTEGER NOT NULL, image_path TEXT);
		INSERT INTO messages (id, sender_id, image_path) VALUES (1, 7, '/uploads/chat/legacy.png');
	`); err != nil {
		t.Fatalf("create fixture: %v", err)
	}

	uploadDirectory := t.TempDir()
	png := append([]byte("\x89PNG\r\n\x1a\n"), make([]byte, 32)...)
	if err := os.WriteFile(filepath.Join(uploadDirectory, "legacy.png"), png, 0600); err != nil {
		t.Fatalf("write legacy image: %v", err)
	}
	store := &fakeObjectStore{}

	count, err := copyLegacyMedia(context.Background(), database, store, uploadDirectory)
	if err != nil {
		t.Fatalf("copyLegacyMedia: %v", err)
	}
	if count != 1 || len(store.putKeys) != 1 {
		t.Fatalf("count = %d, uploaded keys = %v", count, store.putKeys)
	}
	if !strings.HasPrefix(store.putKeys[0], "chat/users/7/") || !strings.HasSuffix(store.putKeys[0], ".png") {
		t.Fatalf("unexpected uploaded key: %q", store.putKeys[0])
	}

	var imagePath string
	if err := database.QueryRow("SELECT image_path FROM messages WHERE id = 1").Scan(&imagePath); err != nil {
		t.Fatalf("read migrated message: %v", err)
	}
	if imagePath != "/media/social-network-media/"+store.putKeys[0] {
		t.Fatalf("image_path = %q", imagePath)
	}

	count, err = copyLegacyMedia(context.Background(), database, store, uploadDirectory)
	if err != nil || count != 0 {
		t.Fatalf("repeat copy = %d, %v; want 0, nil", count, err)
	}
}
