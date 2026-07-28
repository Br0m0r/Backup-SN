package migrations

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
)

//go:embed 000001_initial.up.sql
var files embed.FS

func Apply(ctx context.Context, database *sql.DB) error {
	contents, err := files.ReadFile("000001_initial.up.sql")
	if err != nil {
		return err
	}
	digest := sha256.Sum256(contents)
	checksum := hex.EncodeToString(digest[:])
	connection, err := database.Conn(ctx)
	if err != nil {
		return err
	}
	defer connection.Close()
	if _, err := connection.ExecContext(ctx, "SELECT pg_advisory_lock(hashtext($1))", "social-network-auth-migrations"); err != nil {
		return err
	}
	defer connection.ExecContext(context.Background(), "SELECT pg_advisory_unlock(hashtext($1))", "social-network-auth-migrations")
	if _, err := connection.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY, checksum TEXT NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`); err != nil {
		return err
	}
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer transaction.Rollback()
	var existing string
	err = transaction.QueryRowContext(ctx, "SELECT checksum FROM schema_migrations WHERE version=$1", "000001_initial").Scan(&existing)
	if err == nil {
		if existing != checksum {
			return errors.New("Auth migration checksum changed after application")
		}
		return transaction.Commit()
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if _, err := transaction.ExecContext(ctx, string(contents)); err != nil {
		return fmt.Errorf("apply Auth migration: %w", err)
	}
	if _, err := transaction.ExecContext(ctx, "INSERT INTO schema_migrations(version,checksum) VALUES($1,$2)", "000001_initial", checksum); err != nil {
		return err
	}
	return transaction.Commit()
}
