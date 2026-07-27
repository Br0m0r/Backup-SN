// Package migrations owns the Chat service's PostgreSQL schema.
package migrations

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

const advisoryLockName = "social-network-chat-migrations"

//go:embed *.up.sql
var migrationFiles embed.FS

type Definition struct {
	Version  string
	SQL      string
	Checksum string
}

func Definitions() ([]Definition, error) {
	names, err := fs.Glob(migrationFiles, "*.up.sql")
	if err != nil {
		return nil, fmt.Errorf("list Chat migrations: %w", err)
	}
	sort.Strings(names)
	definitions := make([]Definition, 0, len(names))
	for _, name := range names {
		contents, err := migrationFiles.ReadFile(name)
		if err != nil {
			return nil, fmt.Errorf("read Chat migration %s: %w", name, err)
		}
		digest := sha256.Sum256(contents)
		definitions = append(definitions, Definition{
			Version: strings.TrimSuffix(name, ".up.sql"), SQL: string(contents),
			Checksum: hex.EncodeToString(digest[:]),
		})
	}
	return definitions, nil
}

func Apply(ctx context.Context, database *sql.DB) error {
	definitions, err := Definitions()
	if err != nil {
		return err
	}
	connection, err := database.Conn(ctx)
	if err != nil {
		return fmt.Errorf("reserve Chat migration connection: %w", err)
	}
	defer connection.Close()
	if _, err := connection.ExecContext(ctx, "SELECT pg_advisory_lock(hashtext($1))", advisoryLockName); err != nil {
		return fmt.Errorf("acquire Chat migration lock: %w", err)
	}
	defer connection.ExecContext(context.Background(), "SELECT pg_advisory_unlock(hashtext($1))", advisoryLockName)
	if _, err := connection.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			checksum TEXT NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		return fmt.Errorf("create Chat migration history: %w", err)
	}
	for _, definition := range definitions {
		if err := applyOne(ctx, connection, definition); err != nil {
			return err
		}
	}
	return nil
}

func applyOne(ctx context.Context, connection *sql.Conn, definition Definition) error {
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin Chat migration %s: %w", definition.Version, err)
	}
	defer transaction.Rollback()
	var checksum string
	err = transaction.QueryRowContext(ctx,
		"SELECT checksum FROM schema_migrations WHERE version = $1", definition.Version,
	).Scan(&checksum)
	switch {
	case err == nil:
		if checksum != definition.Checksum {
			return fmt.Errorf("Chat migration %s checksum changed after application", definition.Version)
		}
		return transaction.Commit()
	case !errors.Is(err, sql.ErrNoRows):
		return fmt.Errorf("read Chat migration %s state: %w", definition.Version, err)
	}
	if _, err := transaction.ExecContext(ctx, definition.SQL); err != nil {
		return fmt.Errorf("apply Chat migration %s: %w", definition.Version, err)
	}
	if _, err := transaction.ExecContext(ctx,
		"INSERT INTO schema_migrations (version, checksum) VALUES ($1, $2)",
		definition.Version, definition.Checksum,
	); err != nil {
		return fmt.Errorf("record Chat migration %s: %w", definition.Version, err)
	}
	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit Chat migration %s: %w", definition.Version, err)
	}
	return nil
}
