// Command copy-media migrates legacy Chat image files to object storage and
// rewrites their Chat-owned message references. Stop Chat writes while it runs.
package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"social-network/services/common/objectstore"

	_ "github.com/mattn/go-sqlite3"
)

const maxLegacyImageSize = 5 << 20

type legacyMessage struct {
	ID        int64
	SenderID  int64
	ImagePath string
}

func main() {
	sqlitePath := flag.String("sqlite-path", "./social_network.db", "path to the legacy shared SQLite database")
	uploadsDirectory := flag.String("uploads-dir", "./services/chat/uploads/chat", "directory containing legacy Chat images")
	flag.Parse()

	if err := run(*sqlitePath, *uploadsDirectory); err != nil {
		log.Fatal(err)
	}
}

func run(sqlitePath, uploadsDirectory string) error {
	if _, err := os.Stat(sqlitePath); err != nil {
		return fmt.Errorf("legacy SQLite database is unavailable: %w", err)
	}
	if _, err := os.Stat(uploadsDirectory); err != nil {
		return fmt.Errorf("legacy Chat upload directory is unavailable: %w", err)
	}

	config, err := objectstore.FromEnvironment()
	if err != nil {
		return fmt.Errorf("invalid object storage configuration: %w", err)
	}
	store, err := objectstore.Open(context.Background(), config)
	if err != nil {
		return fmt.Errorf("connect to object storage: %w", err)
	}

	database, err := sql.Open("sqlite3", sqlitePath)
	if err != nil {
		return fmt.Errorf("open legacy SQLite database: %w", err)
	}
	defer database.Close()
	if err := database.Ping(); err != nil {
		return fmt.Errorf("connect to legacy SQLite database: %w", err)
	}

	count, err := copyLegacyMedia(context.Background(), database, store, uploadsDirectory)
	if err != nil {
		return err
	}
	log.Printf("Copied %d legacy Chat images and updated their message references", count)
	return nil
}

func copyLegacyMedia(ctx context.Context, database *sql.DB, store objectstore.Store, uploadsDirectory string) (int, error) {
	rows, err := database.QueryContext(ctx, `
		SELECT id, sender_id, image_path
		FROM messages
		WHERE image_path IS NOT NULL
		  AND image_path != ''
		  AND (image_path LIKE '/uploads/chat/%' OR image_path LIKE 'uploads/chat/%')
		ORDER BY id
	`)
	if err != nil {
		return 0, fmt.Errorf("query legacy Chat media: %w", err)
	}

	messages := make([]legacyMessage, 0)
	for rows.Next() {
		var message legacyMessage
		if err := rows.Scan(&message.ID, &message.SenderID, &message.ImagePath); err != nil {
			rows.Close()
			return 0, fmt.Errorf("scan legacy Chat media: %w", err)
		}
		messages = append(messages, message)
	}
	if err := rows.Close(); err != nil {
		return 0, fmt.Errorf("close legacy Chat media rows: %w", err)
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("iterate legacy Chat media: %w", err)
	}

	copied := 0
	for _, message := range messages {
		contents, contentType, extension, err := readLegacyImage(uploadsDirectory, message.ImagePath)
		if err != nil {
			return copied, fmt.Errorf("message %d: %w", message.ID, err)
		}
		name, err := opaqueName()
		if err != nil {
			return copied, fmt.Errorf("message %d: generate object key: %w", message.ID, err)
		}
		key := fmt.Sprintf("chat/users/%d/%s%s", message.SenderID, name, extension)
		if err := store.Put(ctx, key, bytes.NewReader(contents), int64(len(contents)), contentType); err != nil {
			return copied, fmt.Errorf("message %d: upload legacy image: %w", message.ID, err)
		}

		publicURL := store.URL(key)
		result, err := database.ExecContext(ctx,
			"UPDATE messages SET image_path = ? WHERE id = ? AND image_path = ?",
			publicURL, message.ID, message.ImagePath,
		)
		if err != nil {
			_ = store.Delete(ctx, key)
			return copied, fmt.Errorf("message %d: update image reference: %w", message.ID, err)
		}
		affected, err := result.RowsAffected()
		if err != nil || affected != 1 {
			_ = store.Delete(ctx, key)
			return copied, fmt.Errorf("message %d: image reference changed during migration", message.ID)
		}
		copied++
	}
	return copied, nil
}

func readLegacyImage(directory, imagePath string) ([]byte, string, string, error) {
	filename := filepath.Base(imagePath)
	filePath := filepath.Join(directory, filename)
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, "", "", fmt.Errorf("read %q: %w", filename, err)
	}
	if !info.Mode().IsRegular() || info.Size() <= 0 || info.Size() > maxLegacyImageSize {
		return nil, "", "", fmt.Errorf("legacy image %q has an invalid size", filename)
	}
	contents, err := os.ReadFile(filePath)
	if err != nil {
		return nil, "", "", fmt.Errorf("read %q: %w", filename, err)
	}
	contentType := http.DetectContentType(contents)
	extension, ok := extensionForContentType(contentType)
	if !ok {
		return nil, "", "", fmt.Errorf("legacy image %q has unsupported content type %q", filename, contentType)
	}
	return contents, contentType, extension, nil
}

func opaqueName() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(value[:]), nil
}

func extensionForContentType(contentType string) (string, bool) {
	switch contentType {
	case "image/jpeg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	case "image/gif":
		return ".gif", true
	default:
		return "", false
	}
}
