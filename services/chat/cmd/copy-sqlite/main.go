// Command copy-sqlite copies legacy Chat messages into Chat PostgreSQL.
// Stop all Chat writes while it runs.
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

	"social-network/services/chat/migrations"
	"social-network/services/common/postgres"

	_ "github.com/mattn/go-sqlite3"
)

type directMessage struct {
	ID, SenderID, RecipientID int64
	Content                   string
	IsRead                    bool
	CreatedAt                 time.Time
	ImagePath                 sql.NullString
}

type groupMessage struct {
	ID, GroupID, SenderID int64
	Content               string
	CreatedAt             time.Time
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
		return err
	}
	target, err := postgres.Open(context.Background(), config)
	if err != nil {
		return err
	}
	defer target.Close()
	if err := migrations.Apply(context.Background(), target); err != nil {
		return err
	}
	source, err := sql.Open("sqlite3", "file:"+sqlitePath+"?mode=ro")
	if err != nil {
		return err
	}
	defer source.Close()

	ctx := context.Background()
	direct, err := readDirect(ctx, source)
	if err != nil {
		return fmt.Errorf("read direct messages: %w", err)
	}
	group, err := readGroup(ctx, source)
	if err != nil {
		return fmt.Errorf("read group messages: %w", err)
	}

	transaction, err := target.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer transaction.Rollback()
	for _, message := range direct {
		if _, err := transaction.ExecContext(ctx, `
			INSERT INTO messages (id, sender_id, recipient_id, content, is_read, created_at, image_path)
			VALUES ($1,$2,$3,$4,$5,$6,$7)
			ON CONFLICT (id) DO UPDATE SET sender_id=EXCLUDED.sender_id,
				recipient_id=EXCLUDED.recipient_id, content=EXCLUDED.content,
				is_read=EXCLUDED.is_read, created_at=EXCLUDED.created_at,
				image_path=EXCLUDED.image_path
		`, message.ID, message.SenderID, message.RecipientID, message.Content,
			message.IsRead, message.CreatedAt, message.ImagePath); err != nil {
			return fmt.Errorf("copy direct message %d: %w", message.ID, err)
		}
	}
	for _, message := range group {
		if _, err := transaction.ExecContext(ctx, `
			INSERT INTO group_messages (id, group_id, sender_id, content, created_at)
			VALUES ($1,$2,$3,$4,$5)
			ON CONFLICT (id) DO UPDATE SET group_id=EXCLUDED.group_id,
				sender_id=EXCLUDED.sender_id, content=EXCLUDED.content,
				created_at=EXCLUDED.created_at
		`, message.ID, message.GroupID, message.SenderID, message.Content, message.CreatedAt); err != nil {
			return fmt.Errorf("copy group message %d: %w", message.ID, err)
		}
	}
	targetDirect, err := readDirect(ctx, transaction)
	if err != nil {
		return err
	}
	targetGroup, err := readGroup(ctx, transaction)
	if err != nil {
		return err
	}
	sourceChecksum, err := checksum(direct, group)
	if err != nil {
		return err
	}
	targetChecksum, err := checksum(targetDirect, targetGroup)
	if err != nil {
		return err
	}
	if len(direct) != len(targetDirect) || len(group) != len(targetGroup) || sourceChecksum != targetChecksum {
		return fmt.Errorf("Chat verification failed; target transaction rolled back")
	}
	for _, table := range []string{"messages", "group_messages"} {
		if _, err := transaction.ExecContext(ctx, fmt.Sprintf(`
			SELECT setval(pg_get_serial_sequence('%s','id')::regclass,
				COALESCE(MAX(id),1), COUNT(*) > 0) FROM %s
		`, table, table)); err != nil {
			return err
		}
	}
	if err := transaction.Commit(); err != nil {
		return err
	}
	log.Printf("Copied and verified %d direct and %d group messages (checksum %s)", len(direct), len(group), sourceChecksum)
	return nil
}

type queryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func readDirect(ctx context.Context, database queryer) ([]directMessage, error) {
	rows, err := database.QueryContext(ctx, `
		SELECT id, sender_id, recipient_id, content, is_read, created_at, image_path
		FROM messages ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	records := make([]directMessage, 0)
	for rows.Next() {
		var record directMessage
		if err := rows.Scan(&record.ID, &record.SenderID, &record.RecipientID,
			&record.Content, &record.IsRead, &record.CreatedAt, &record.ImagePath); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

func readGroup(ctx context.Context, database queryer) ([]groupMessage, error) {
	rows, err := database.QueryContext(ctx, `
		SELECT id, group_id, sender_id, content, created_at
		FROM group_messages ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	records := make([]groupMessage, 0)
	for rows.Next() {
		var record groupMessage
		if err := rows.Scan(&record.ID, &record.GroupID, &record.SenderID,
			&record.Content, &record.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

func checksum(direct []directMessage, group []groupMessage) (string, error) {
	hash := sha256.New()
	encoder := json.NewEncoder(hash)
	for _, record := range direct {
		imagePath := any(nil)
		if record.ImagePath.Valid {
			imagePath = record.ImagePath.String
		}
		if err := encoder.Encode(struct {
			ID, SenderID, RecipientID int64
			Content                   string
			IsRead                    bool
			CreatedAt                 string
			ImagePath                 any
		}{
			record.ID, record.SenderID, record.RecipientID, record.Content,
			record.IsRead, record.CreatedAt.UTC().Format(time.RFC3339Nano), imagePath,
		}); err != nil {
			return "", err
		}
	}
	for _, record := range group {
		if err := encoder.Encode(struct {
			ID, GroupID, SenderID int64
			Content               string
			CreatedAt             string
		}{
			record.ID, record.GroupID, record.SenderID, record.Content,
			record.CreatedAt.UTC().Format(time.RFC3339Nano),
		}); err != nil {
			return "", err
		}
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
