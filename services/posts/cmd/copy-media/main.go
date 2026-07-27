// Command copy-media migrates legacy Post and Comment image files to object
// storage and rewrites their Posts-owned SQLite references. Stop writes first.
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

const maxLegacyPostImageSize = 5 << 20

type legacyMediaReference struct {
	Table     string
	ID        int64
	UserID    int64
	ImagePath string
}

func main() {
	sqlitePath := flag.String("sqlite-path", "./social_network.db", "path to the legacy shared SQLite database")
	uploadsDirectory := flag.String("uploads-dir", "./services/posts/uploads/posts", "directory containing legacy Post images")
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
		return fmt.Errorf("legacy Post upload directory is unavailable: %w", err)
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
	log.Printf("Copied %d legacy Post/Comment images and updated their references", count)
	return nil
}

func copyLegacyMedia(ctx context.Context, database *sql.DB, store objectstore.Store, uploadsDirectory string) (int, error) {
	rows, err := database.QueryContext(ctx, `
		SELECT 'posts', id, user_id, image_path
		FROM posts
		WHERE image_path IS NOT NULL AND image_path != ''
		  AND (image_path LIKE '/uploads/posts/%' OR image_path LIKE 'uploads/posts/%')
		UNION ALL
		SELECT 'comments', id, user_id, image_path
		FROM comments
		WHERE image_path IS NOT NULL AND image_path != ''
		  AND (image_path LIKE '/uploads/posts/%' OR image_path LIKE 'uploads/posts/%')
		ORDER BY 1, 2
	`)
	if err != nil {
		return 0, fmt.Errorf("query legacy Post media: %w", err)
	}

	references := make([]legacyMediaReference, 0)
	for rows.Next() {
		var reference legacyMediaReference
		if err := rows.Scan(&reference.Table, &reference.ID, &reference.UserID, &reference.ImagePath); err != nil {
			rows.Close()
			return 0, fmt.Errorf("scan legacy Post media: %w", err)
		}
		references = append(references, reference)
	}
	if err := rows.Close(); err != nil {
		return 0, fmt.Errorf("close legacy Post media rows: %w", err)
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("iterate legacy Post media: %w", err)
	}

	copied := 0
	for _, reference := range references {
		contents, contentType, extension, err := readLegacyImage(uploadsDirectory, reference.ImagePath)
		if err != nil {
			return copied, fmt.Errorf("%s %d: %w", reference.Table, reference.ID, err)
		}
		name, err := opaqueName()
		if err != nil {
			return copied, fmt.Errorf("%s %d: generate object key: %w", reference.Table, reference.ID, err)
		}
		key := fmt.Sprintf("posts/users/%d/%s%s", reference.UserID, name, extension)
		if err := store.Put(ctx, key, bytes.NewReader(contents), int64(len(contents)), contentType); err != nil {
			return copied, fmt.Errorf("%s %d: upload legacy image: %w", reference.Table, reference.ID, err)
		}

		publicURL := store.URL(key)
		query, err := conditionalUpdateQuery(reference.Table)
		if err != nil {
			_ = store.Delete(ctx, key)
			return copied, err
		}
		result, err := database.ExecContext(ctx, query, publicURL, reference.ID, reference.ImagePath)
		if err != nil {
			_ = store.Delete(ctx, key)
			return copied, fmt.Errorf("%s %d: update image reference: %w", reference.Table, reference.ID, err)
		}
		affected, err := result.RowsAffected()
		if err != nil || affected != 1 {
			_ = store.Delete(ctx, key)
			return copied, fmt.Errorf("%s %d: image reference changed during migration", reference.Table, reference.ID)
		}
		copied++
	}
	return copied, nil
}

func conditionalUpdateQuery(table string) (string, error) {
	switch table {
	case "posts":
		return "UPDATE posts SET image_path = ? WHERE id = ? AND image_path = ?", nil
	case "comments":
		return "UPDATE comments SET image_path = ? WHERE id = ? AND image_path = ?", nil
	default:
		return "", fmt.Errorf("unsupported media reference table %q", table)
	}
}

func readLegacyImage(directory, imagePath string) ([]byte, string, string, error) {
	filename := filepath.Base(imagePath)
	filePath := filepath.Join(directory, filename)
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, "", "", fmt.Errorf("read %q: %w", filename, err)
	}
	if !info.Mode().IsRegular() || info.Size() <= 0 || info.Size() > maxLegacyPostImageSize {
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
