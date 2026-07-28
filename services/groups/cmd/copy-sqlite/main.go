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
	"social-network/services/groups/migrations"

	_ "github.com/mattn/go-sqlite3"
)

type groupRecord struct {
	ID, CreatorID         int64
	Name                  string
	Description, ImageURL sql.NullString
	CreatedAt             time.Time
}

type memberRecord struct {
	ID, GroupID, UserID int64
	Role, Status        string
	JoinedAt            time.Time
}

type eventRecord struct {
	ID, GroupID          int64
	CreatorID            sql.NullInt64
	Title                string
	Description          sql.NullString
	EventTime, CreatedAt time.Time
}

type responseRecord struct {
	ID, EventID, UserID int64
	Response            string
	CreatedAt           time.Time
}

type records struct {
	Groups    []groupRecord
	Members   []memberRecord
	Events    []eventRecord
	Responses []responseRecord
}

type queryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func main() {
	sqlitePath := flag.String("sqlite-path", "./social_network.db", "legacy SQLite path")
	flag.Parse()
	if err := run(*sqlitePath); err != nil {
		log.Fatal(err)
	}
}

func run(sqlitePath string) error {
	if _, err := os.Stat(sqlitePath); err != nil {
		return err
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
	sourceRecords, err := readAll(ctx, source)
	if err != nil {
		return err
	}
	transaction, err := target.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer transaction.Rollback()
	if err := copyAll(ctx, transaction, sourceRecords); err != nil {
		return err
	}
	targetRecords, err := readAll(ctx, transaction)
	if err != nil {
		return err
	}
	sourceChecksum, err := checksum(sourceRecords)
	if err != nil {
		return err
	}
	targetChecksum, err := checksum(targetRecords)
	if err != nil {
		return err
	}
	if len(sourceRecords.Groups) != len(targetRecords.Groups) ||
		len(sourceRecords.Members) != len(targetRecords.Members) ||
		len(sourceRecords.Events) != len(targetRecords.Events) ||
		len(sourceRecords.Responses) != len(targetRecords.Responses) ||
		sourceChecksum != targetChecksum {
		return fmt.Errorf("Groups verification failed; target transaction rolled back")
	}
	for _, table := range []string{"groups", "group_members", "events", "event_responses"} {
		if _, err := transaction.ExecContext(ctx, fmt.Sprintf(`
			SELECT setval(pg_get_serial_sequence('%s','id')::regclass,
				COALESCE(MAX(id),1),COUNT(*) > 0) FROM %s`, table, table)); err != nil {
			return err
		}
	}
	if err := transaction.Commit(); err != nil {
		return err
	}
	log.Printf("Copied %d groups, %d members, %d events, and %d responses (checksum %s)",
		len(sourceRecords.Groups), len(sourceRecords.Members), len(sourceRecords.Events),
		len(sourceRecords.Responses), sourceChecksum)
	return nil
}

func copyAll(ctx context.Context, transaction *sql.Tx, data records) error {
	for _, row := range data.Groups {
		if _, err := transaction.ExecContext(ctx, `
			INSERT INTO groups (id,name,description,image_url,creator_id,created_at)
			VALUES ($1,$2,$3,$4,$5,$6)
			ON CONFLICT (id) DO UPDATE SET name=EXCLUDED.name,
				description=EXCLUDED.description,image_url=EXCLUDED.image_url,
				creator_id=EXCLUDED.creator_id,created_at=EXCLUDED.created_at`,
			row.ID, row.Name, row.Description, row.ImageURL, row.CreatorID, row.CreatedAt); err != nil {
			return fmt.Errorf("copy group %d: %w", row.ID, err)
		}
	}
	for _, row := range data.Members {
		if _, err := transaction.ExecContext(ctx, `
			INSERT INTO group_members (id,group_id,user_id,role,status,joined_at)
			VALUES ($1,$2,$3,$4,$5,$6)
			ON CONFLICT (id) DO UPDATE SET group_id=EXCLUDED.group_id,
				user_id=EXCLUDED.user_id,role=EXCLUDED.role,status=EXCLUDED.status,
				joined_at=EXCLUDED.joined_at`,
			row.ID, row.GroupID, row.UserID, row.Role, row.Status, row.JoinedAt); err != nil {
			return fmt.Errorf("copy group member %d: %w", row.ID, err)
		}
	}
	for _, row := range data.Events {
		if _, err := transaction.ExecContext(ctx, `
			INSERT INTO events (id,group_id,creator_id,title,description,event_time,created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7)
			ON CONFLICT (id) DO UPDATE SET group_id=EXCLUDED.group_id,
				creator_id=EXCLUDED.creator_id,title=EXCLUDED.title,
				description=EXCLUDED.description,event_time=EXCLUDED.event_time,
				created_at=EXCLUDED.created_at`,
			row.ID, row.GroupID, row.CreatorID, row.Title, row.Description, row.EventTime, row.CreatedAt); err != nil {
			return fmt.Errorf("copy event %d: %w", row.ID, err)
		}
	}
	for _, row := range data.Responses {
		if _, err := transaction.ExecContext(ctx, `
			INSERT INTO event_responses (id,event_id,user_id,response,created_at)
			VALUES ($1,$2,$3,$4,$5)
			ON CONFLICT (id) DO UPDATE SET event_id=EXCLUDED.event_id,
				user_id=EXCLUDED.user_id,response=EXCLUDED.response,
				created_at=EXCLUDED.created_at`,
			row.ID, row.EventID, row.UserID, row.Response, row.CreatedAt); err != nil {
			return fmt.Errorf("copy event response %d: %w", row.ID, err)
		}
	}
	return nil
}

func readAll(ctx context.Context, database queryer) (records, error) {
	var result records
	rows, err := database.QueryContext(ctx, `SELECT id,name,description,image_url,creator_id,created_at FROM groups ORDER BY id`)
	if err != nil {
		return result, err
	}
	for rows.Next() {
		var row groupRecord
		if err := rows.Scan(&row.ID, &row.Name, &row.Description, &row.ImageURL, &row.CreatorID, &row.CreatedAt); err != nil {
			rows.Close()
			return result, err
		}
		result.Groups = append(result.Groups, row)
	}
	if err := rows.Close(); err != nil {
		return result, err
	}
	rows, err = database.QueryContext(ctx, `SELECT id,group_id,user_id,role,status,joined_at FROM group_members ORDER BY id`)
	if err != nil {
		return result, err
	}
	for rows.Next() {
		var row memberRecord
		if err := rows.Scan(&row.ID, &row.GroupID, &row.UserID, &row.Role, &row.Status, &row.JoinedAt); err != nil {
			rows.Close()
			return result, err
		}
		result.Members = append(result.Members, row)
	}
	if err := rows.Close(); err != nil {
		return result, err
	}
	rows, err = database.QueryContext(ctx, `SELECT id,group_id,creator_id,title,description,event_time,created_at FROM events ORDER BY id`)
	if err != nil {
		return result, err
	}
	for rows.Next() {
		var row eventRecord
		if err := rows.Scan(&row.ID, &row.GroupID, &row.CreatorID, &row.Title, &row.Description, &row.EventTime, &row.CreatedAt); err != nil {
			rows.Close()
			return result, err
		}
		result.Events = append(result.Events, row)
	}
	if err := rows.Close(); err != nil {
		return result, err
	}
	rows, err = database.QueryContext(ctx, `SELECT id,event_id,user_id,response,created_at FROM event_responses ORDER BY id`)
	if err != nil {
		return result, err
	}
	for rows.Next() {
		var row responseRecord
		if err := rows.Scan(&row.ID, &row.EventID, &row.UserID, &row.Response, &row.CreatedAt); err != nil {
			rows.Close()
			return result, err
		}
		result.Responses = append(result.Responses, row)
	}
	return result, rows.Close()
}

func checksum(data records) (string, error) {
	for index := range data.Groups {
		data.Groups[index].CreatedAt = data.Groups[index].CreatedAt.UTC()
	}
	for index := range data.Members {
		data.Members[index].JoinedAt = data.Members[index].JoinedAt.UTC()
	}
	for index := range data.Events {
		data.Events[index].EventTime = data.Events[index].EventTime.UTC()
		data.Events[index].CreatedAt = data.Events[index].CreatedAt.UTC()
	}
	for index := range data.Responses {
		data.Responses[index].CreatedAt = data.Responses[index].CreatedAt.UTC()
	}
	hash := sha256.New()
	if err := json.NewEncoder(hash).Encode(data); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
