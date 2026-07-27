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

func TestCopyLegacyMediaMigratesPostsAndComments(t *testing.T) {
	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open SQLite: %v", err)
	}
	defer database.Close()
	database.SetMaxOpenConns(1)
	if _, err := database.Exec(`
		CREATE TABLE posts (id INTEGER PRIMARY KEY, user_id INTEGER NOT NULL, image_path TEXT);
		CREATE TABLE comments (id INTEGER PRIMARY KEY, user_id INTEGER NOT NULL, image_path TEXT);
		INSERT INTO posts (id, user_id, image_path) VALUES (1, 7, 'uploads/posts/post.png');
		INSERT INTO comments (id, user_id, image_path) VALUES (2, 8, '/uploads/posts/comment.png');
	`); err != nil {
		t.Fatalf("create fixture: %v", err)
	}

	uploadDirectory := t.TempDir()
	png := append([]byte("\x89PNG\r\n\x1a\n"), make([]byte, 32)...)
	for _, name := range []string{"post.png", "comment.png"} {
		if err := os.WriteFile(filepath.Join(uploadDirectory, name), png, 0600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	store := &fakeObjectStore{}

	count, err := copyLegacyMedia(context.Background(), database, store, uploadDirectory)
	if err != nil {
		t.Fatalf("copyLegacyMedia: %v", err)
	}
	if count != 2 || len(store.putKeys) != 2 {
		t.Fatalf("count = %d, uploaded keys = %v", count, store.putKeys)
	}

	for _, check := range []struct {
		table  string
		id     int
		prefix string
	}{
		{table: "posts", id: 1, prefix: "/media/social-network-media/posts/users/7/"},
		{table: "comments", id: 2, prefix: "/media/social-network-media/posts/users/8/"},
	} {
		var imagePath string
		query := "SELECT image_path FROM " + check.table + " WHERE id = ?"
		if err := database.QueryRow(query, check.id).Scan(&imagePath); err != nil {
			t.Fatalf("read migrated %s: %v", check.table, err)
		}
		if !strings.HasPrefix(imagePath, check.prefix) || !strings.HasSuffix(imagePath, ".png") {
			t.Fatalf("%s image_path = %q", check.table, imagePath)
		}
	}

	count, err = copyLegacyMedia(context.Background(), database, store, uploadDirectory)
	if err != nil || count != 0 {
		t.Fatalf("repeat copy = %d, %v; want 0, nil", count, err)
	}
}
