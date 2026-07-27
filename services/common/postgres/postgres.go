// Package postgres provides shared PostgreSQL connection configuration.
// It intentionally contains infrastructure concerns only; domain repositories
// and migrations remain owned by each service.
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	defaultMaxOpenConns    = 25
	defaultMaxIdleConns    = 5
	defaultConnMaxLifetime = 30 * time.Minute
	defaultConnMaxIdleTime = 5 * time.Minute
	defaultConnectTimeout  = 10 * time.Second
)

// Config contains connection and pool settings for a service-owned database.
type Config struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	ConnectTimeout  time.Duration
}

// FromEnvironment reads PostgreSQL configuration from the process environment.
func FromEnvironment() (Config, error) {
	return FromLookup(os.LookupEnv)
}

// FromLookup builds Config using an environment-like lookup function.
func FromLookup(lookup func(string) (string, bool)) (Config, error) {
	databaseURL, ok := lookup("DATABASE_URL")
	if !ok || strings.TrimSpace(databaseURL) == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	maxOpen, err := positiveInt(lookup, "DB_MAX_OPEN_CONNS", defaultMaxOpenConns)
	if err != nil {
		return Config{}, err
	}
	maxIdle, err := nonNegativeInt(lookup, "DB_MAX_IDLE_CONNS", defaultMaxIdleConns)
	if err != nil {
		return Config{}, err
	}
	if maxIdle > maxOpen {
		return Config{}, fmt.Errorf("DB_MAX_IDLE_CONNS must not exceed DB_MAX_OPEN_CONNS")
	}

	maxLifetime, err := positiveDuration(lookup, "DB_CONN_MAX_LIFETIME", defaultConnMaxLifetime)
	if err != nil {
		return Config{}, err
	}
	maxIdleTime, err := positiveDuration(lookup, "DB_CONN_MAX_IDLE_TIME", defaultConnMaxIdleTime)
	if err != nil {
		return Config{}, err
	}
	connectTimeout, err := positiveDuration(lookup, "DB_CONNECT_TIMEOUT", defaultConnectTimeout)
	if err != nil {
		return Config{}, err
	}

	return Config{
		URL:             databaseURL,
		MaxOpenConns:    maxOpen,
		MaxIdleConns:    maxIdle,
		ConnMaxLifetime: maxLifetime,
		ConnMaxIdleTime: maxIdleTime,
		ConnectTimeout:  connectTimeout,
	}, nil
}

// Open connects to PostgreSQL, verifies connectivity, and configures the pool.
func Open(parent context.Context, config Config) (*sql.DB, error) {
	database, err := sql.Open("pgx", config.URL)
	if err != nil {
		return nil, fmt.Errorf("open PostgreSQL: %w", err)
	}

	database.SetMaxOpenConns(config.MaxOpenConns)
	database.SetMaxIdleConns(config.MaxIdleConns)
	database.SetConnMaxLifetime(config.ConnMaxLifetime)
	database.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	ctx, cancel := context.WithTimeout(parent, config.ConnectTimeout)
	defer cancel()
	if err := database.PingContext(ctx); err != nil {
		database.Close()
		return nil, fmt.Errorf("connect to PostgreSQL: %w", err)
	}

	return database, nil
}

// Description returns a credential-free database description suitable for logs.
func Description(databaseURL string) string {
	parsed, err := url.Parse(databaseURL)
	if err != nil || parsed.Host == "" {
		return "configured PostgreSQL database"
	}
	databaseName := strings.TrimPrefix(parsed.Path, "/")
	if databaseName == "" {
		databaseName = "(default)"
	}
	return fmt.Sprintf("PostgreSQL host=%s database=%s", parsed.Host, databaseName)
}

func positiveInt(lookup func(string) (string, bool), name string, fallback int) (int, error) {
	value, ok := lookup(name)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", name)
	}
	return parsed, nil
}

func nonNegativeInt(lookup func(string) (string, bool), name string, fallback int) (int, error) {
	value, ok := lookup(name)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return 0, fmt.Errorf("%s must be a non-negative integer", name)
	}
	return parsed, nil
}

func positiveDuration(lookup func(string) (string, bool), name string, fallback time.Duration) (time.Duration, error) {
	value, ok := lookup(name)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive Go duration", name)
	}
	return parsed, nil
}
