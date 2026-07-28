package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"social-network/services/auth/migrations"
	"social-network/services/common/postgres"

	_ "github.com/mattn/go-sqlite3"
)

type account struct {
	ID                            int64
	Username, Email, PasswordHash string
	CreatedAt                     time.Time
}
type session struct {
	ID, AccountID int64
	Token         string
	CreatedAt     time.Time
	ExpiresAt     time.Time
}
type records struct {
	Accounts []account
	Sessions []session
}
type queryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func main() {
	path := flag.String("sqlite-path", "./social_network.db", "legacy SQLite path")
	flag.Parse()
	if err := run(*path); err != nil {
		log.Fatal(err)
	}
}

func run(path string) error {
	if _, err := os.Stat(path); err != nil {
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
	source, err := sql.Open("sqlite3", "file:"+path+"?mode=ro")
	if err != nil {
		return err
	}
	defer source.Close()
	ctx := context.Background()
	sourceRecords, err := readAll(ctx, source, true)
	if err != nil {
		return err
	}
	tx, err := target.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, row := range sourceRecords.Accounts {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO accounts(id,username,email,password_hash,created_at)
			VALUES($1,$2,$3,$4,$5)
			ON CONFLICT(id) DO UPDATE SET username=EXCLUDED.username,
				email=EXCLUDED.email,password_hash=EXCLUDED.password_hash,
				created_at=EXCLUDED.created_at`,
			row.ID, row.Username, row.Email, row.PasswordHash, row.CreatedAt); err != nil {
			return fmt.Errorf("copy account %d: %w", row.ID, err)
		}
	}
	for _, row := range sourceRecords.Sessions {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO sessions(id,account_id,token,created_at,expires_at)
			VALUES($1,$2,$3,$4,$5)
			ON CONFLICT(id) DO UPDATE SET account_id=EXCLUDED.account_id,
				token=EXCLUDED.token,created_at=EXCLUDED.created_at,
				expires_at=EXCLUDED.expires_at`,
			row.ID, row.AccountID, row.Token, row.CreatedAt, row.ExpiresAt); err != nil {
			return fmt.Errorf("copy session %d: %w", row.ID, err)
		}
	}
	targetRecords, err := readAll(ctx, tx, false)
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
	if len(sourceRecords.Accounts) != len(targetRecords.Accounts) ||
		len(sourceRecords.Sessions) != len(targetRecords.Sessions) ||
		sourceChecksum != targetChecksum {
		return errors.New("Auth verification failed; target transaction rolled back")
	}
	for _, table := range []string{"accounts", "sessions"} {
		if _, err := tx.ExecContext(ctx, fmt.Sprintf(`
			SELECT setval(pg_get_serial_sequence('%s','id')::regclass,
				COALESCE(MAX(id),1),COUNT(*) > 0) FROM %s`, table, table)); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	log.Printf("Copied %d accounts and %d sessions (checksum %s)",
		len(sourceRecords.Accounts), len(sourceRecords.Sessions), sourceChecksum)
	return nil
}

func readAll(ctx context.Context, database queryer, legacy bool) (records, error) {
	var result records
	accountTable := "accounts"
	sessionOwnerColumn := "account_id"
	if legacy {
		accountTable = "users"
		sessionOwnerColumn = "user_id"
	}
	rows, err := database.QueryContext(ctx, `SELECT id,username,email,password_hash,created_at FROM `+accountTable+` ORDER BY id`)
	if err != nil {
		return result, err
	}
	for rows.Next() {
		var row account
		if err := rows.Scan(&row.ID, &row.Username, &row.Email, &row.PasswordHash, &row.CreatedAt); err != nil {
			rows.Close()
			return result, err
		}
		result.Accounts = append(result.Accounts, row)
	}
	if err := rows.Close(); err != nil {
		return result, err
	}
	rows, err = database.QueryContext(ctx, `SELECT id,`+sessionOwnerColumn+`,token,created_at,expires_at FROM sessions ORDER BY id`)
	if err != nil {
		return result, err
	}
	for rows.Next() {
		var row session
		if err := rows.Scan(&row.ID, &row.AccountID, &row.Token, &row.CreatedAt, &row.ExpiresAt); err != nil {
			rows.Close()
			return result, err
		}
		result.Sessions = append(result.Sessions, row)
	}
	return result, rows.Close()
}

func checksum(data records) (string, error) {
	for index := range data.Accounts {
		data.Accounts[index].CreatedAt = data.Accounts[index].CreatedAt.UTC()
	}
	for index := range data.Sessions {
		data.Sessions[index].CreatedAt = data.Sessions[index].CreatedAt.UTC()
		data.Sessions[index].ExpiresAt = data.Sessions[index].ExpiresAt.UTC()
	}
	hash := sha256.New()
	if err := json.NewEncoder(hash).Encode(data); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
