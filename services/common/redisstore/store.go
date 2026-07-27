package redisstore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	defaultNamespace        = "social-network"
	defaultOperationTimeout = 2 * time.Second
)

// Config contains the shared Redis connection and keyspace settings.
type Config struct {
	URL              string
	Namespace        string
	OperationTimeout time.Duration
}

// FromEnvironment loads the Redis configuration shared by the Gateway, Chat,
// and Notifications services.
func FromEnvironment() (Config, error) {
	config := Config{
		URL:              strings.TrimSpace(os.Getenv("REDIS_URL")),
		Namespace:        strings.TrimSpace(os.Getenv("REDIS_NAMESPACE")),
		OperationTimeout: defaultOperationTimeout,
	}
	if config.URL == "" {
		return Config{}, errors.New("REDIS_URL is required")
	}
	if config.Namespace == "" {
		config.Namespace = defaultNamespace
	}
	if strings.ContainsAny(config.Namespace, " \t\r\n") {
		return Config{}, errors.New("REDIS_NAMESPACE must not contain whitespace")
	}
	if value := strings.TrimSpace(os.Getenv("REDIS_OPERATION_TIMEOUT")); value != "" {
		timeout, err := time.ParseDuration(value)
		if err != nil || timeout <= 0 {
			return Config{}, errors.New("REDIS_OPERATION_TIMEOUT must be a positive duration")
		}
		config.OperationTimeout = timeout
	}
	return config, nil
}

// Store is a connected Redis client with a namespaced key builder and bounded
// operation contexts.
type Store struct {
	client           *redis.Client
	namespace        string
	operationTimeout time.Duration
}

// Open connects to Redis and verifies the connection before returning.
func Open(parent context.Context, config Config) (*Store, error) {
	options, err := redis.ParseURL(config.URL)
	if err != nil {
		return nil, fmt.Errorf("parse REDIS_URL: %w", err)
	}
	client := redis.NewClient(options)
	ctx, cancel := context.WithTimeout(parent, config.OperationTimeout)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("connect to Redis: %w", err)
	}
	return &Store{
		client:           client,
		namespace:        config.Namespace,
		operationTimeout: config.OperationTimeout,
	}, nil
}

func (s *Store) Close() error {
	return s.client.Close()
}

func (s *Store) Client() *redis.Client {
	return s.client
}

func (s *Store) Context(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, s.operationTimeout)
}

func (s *Store) Key(parts ...string) string {
	cleaned := make([]string, 0, len(parts)+1)
	cleaned = append(cleaned, s.namespace)
	for _, part := range parts {
		cleaned = append(cleaned, strings.Trim(strings.TrimSpace(part), ":"))
	}
	return strings.Join(cleaned, ":")
}

func (s *Store) IntKey(parts []string, value int) string {
	return s.Key(append(parts, strconv.Itoa(value))...)
}
