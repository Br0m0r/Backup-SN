// Command copy-sqlite copies the legacy shared-SQLite notification rows into
// the service-owned PostgreSQL database. Stop notification writes while it runs.
package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"social-network/services/common/postgres"
	"social-network/services/notifications/migrations"

	_ "github.com/mattn/go-sqlite3"
)

type notificationRecord struct {
	ID        int64
	UserID    int64
	Type      string
	RelatedID sql.NullInt64
	Content   string
	IsRead    bool
	CreatedAt time.Time
}

type queryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func main() {
	sqlitePath := flag.String("sqlite-path", "./social_network.db", "path to the legacy SQLite database")
	flag.Parse()
	if err := run(*sqlitePath); err != nil {
		log.Fatal(err)
	}
}

func run(sqlitePath string) error {
	if _, err := os.Stat(sqlitePath); err != nil {
		return fmt.Errorf("legacy SQLite database is unavailable: %w", err)
	}

	config, err := postgres.FromEnvironment()
	if err != nil {
		return fmt.Errorf("invalid notification database configuration: %w", err)
	}
	target, err := postgres.Open(context.Background(), config)
	if err != nil {
		return fmt.Errorf("open notification PostgreSQL database: %w", err)
	}
	defer target.Close()
	if err := migrations.Apply(context.Background(), target); err != nil {
		return fmt.Errorf("migrate target notification database: %w", err)
	}

	source, err := sql.Open("sqlite3", sqlitePath)
	if err != nil {
		return fmt.Errorf("open legacy SQLite database: %w", err)
	}
	defer source.Close()
	if err := source.Ping(); err != nil {
		return fmt.Errorf("connect to legacy SQLite database: %w", err)
	}

	ctx := context.Background()
	sourceRecords, err := readRecords(ctx, source, `
		SELECT id, user_id, type, related_id, content, is_read, created_at
		FROM notifications
		ORDER BY id
	`)
	if err != nil {
		return fmt.Errorf("read legacy notifications: %w", err)
	}

	transaction, err := target.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("start notification copy transaction: %w", err)
	}
	defer transaction.Rollback()

	for _, record := range sourceRecords {
		if _, err := transaction.ExecContext(ctx, `
			INSERT INTO notifications
				(id, user_id, type, related_id, content, is_read, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (id) DO UPDATE SET
				user_id = EXCLUDED.user_id,
				type = EXCLUDED.type,
				related_id = EXCLUDED.related_id,
				content = EXCLUDED.content,
				is_read = EXCLUDED.is_read,
				created_at = EXCLUDED.created_at
		`, record.ID, record.UserID, record.Type, record.RelatedID, record.Content, record.IsRead, record.CreatedAt); err != nil {
			return fmt.Errorf("copy notification %d: %w", record.ID, err)
		}
	}

	targetRecords, err := readRecords(ctx, transaction, `
		SELECT id, user_id, type, related_id, content, is_read, created_at
		FROM notifications
		ORDER BY id
	`)
	if err != nil {
		return fmt.Errorf("verify copied notifications: %w", err)
	}
	sourceChecksum, err := checksum(sourceRecords)
	if err != nil {
		return fmt.Errorf("checksum legacy notifications: %w", err)
	}
	targetChecksum, err := checksum(targetRecords)
	if err != nil {
		return fmt.Errorf("checksum copied notifications: %w", err)
	}
	if len(sourceRecords) != len(targetRecords) || sourceChecksum != targetChecksum {
		return fmt.Errorf(
			"notification verification failed: source rows=%d checksum=%s, target rows=%d checksum=%s; target transaction rolled back",
			len(sourceRecords), sourceChecksum, len(targetRecords), targetChecksum,
		)
	}

	if _, err := transaction.ExecContext(ctx, `
		SELECT setval(
			pg_get_serial_sequence('notifications', 'id')::regclass,
			COALESCE(MAX(id), 1),
			COUNT(*) > 0
		)
		FROM notifications
	`); err != nil {
		return fmt.Errorf("advance notification identity sequence: %w", err)
	}
	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit notification copy: %w", err)
	}

	log.Printf("Copied and verified %d notifications (checksum %s)", len(sourceRecords), sourceChecksum)
	return nil
}

func readRecords(ctx context.Context, database queryer, query string) ([]notificationRecord, error) {
	rows, err := database.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]notificationRecord, 0)
	for rows.Next() {
		var record notificationRecord
		if err := rows.Scan(
			&record.ID,
			&record.UserID,
			&record.Type,
			&record.RelatedID,
			&record.Content,
			&record.IsRead,
			&record.CreatedAt,
		); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

func checksum(records []notificationRecord) (string, error) {
	hash := sha256.New()
	encoder := json.NewEncoder(hash)
	for _, record := range records {
		relatedID := any(nil)
		if record.RelatedID.Valid {
			relatedID = record.RelatedID.Int64
		}
		fingerprint := struct {
			ID        int64  `json:"id"`
			UserID    int64  `json:"user_id"`
			Type      string `json:"type"`
			RelatedID any    `json:"related_id"`
			Content   string `json:"content"`
			IsRead    bool   `json:"is_read"`
			CreatedAt string `json:"created_at"`
		}{
			ID:        record.ID,
			UserID:    record.UserID,
			Type:      record.Type,
			RelatedID: relatedID,
			Content:   record.Content,
			IsRead:    record.IsRead,
			CreatedAt: record.CreatedAt.UTC().Format(time.RFC3339Nano),
		}
		if err := encoder.Encode(fingerprint); err != nil {
			return "", fmt.Errorf("encode notification %d: %w", record.ID, err)
		}
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
