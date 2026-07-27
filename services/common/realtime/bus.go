package realtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"social-network/services/common/redisstore"

	"github.com/redis/go-redis/v9"
)

const (
	defaultPresenceTTL     = 45 * time.Second
	defaultPresenceRefresh = 15 * time.Second
)

var (
	markOnlineScript = redis.NewScript(`
local now = redis.call("TIME")
local now_ms = (now[1] * 1000) + math.floor(now[2] / 1000)
local expires_at = now_ms + tonumber(ARGV[2])
redis.call("ZADD", KEYS[1], expires_at, ARGV[1])
redis.call("PEXPIRE", KEYS[1], tonumber(ARGV[2]) * 2)
return expires_at
`)
	isOnlineScript = redis.NewScript(`
local now = redis.call("TIME")
local now_ms = (now[1] * 1000) + math.floor(now[2] / 1000)
redis.call("ZREMRANGEBYSCORE", KEYS[1], "-inf", now_ms)
return redis.call("ZCARD", KEYS[1])
`)
)

// Transport is the realtime contract consumed by WebSocket hubs.
type Transport interface {
	Publish(context.Context, []byte) error
	Subscribe(context.Context, func([]byte)) error
	MarkOnline(context.Context, int) error
	MarkOffline(context.Context, int) error
	IsOnline(context.Context, int) (bool, error)
	PresenceRefreshInterval() time.Duration
}

// Bus provides service-scoped Redis Pub/Sub and per-instance user presence.
type Bus struct {
	store           *redisstore.Store
	service         string
	instanceID      string
	channel         string
	presenceTTL     time.Duration
	presenceRefresh time.Duration
}

type envelope struct {
	Origin  string          `json:"origin"`
	Payload json.RawMessage `json:"payload"`
}

// New creates a service-scoped realtime bus.
func New(store *redisstore.Store, service string) (*Bus, error) {
	service = strings.Trim(strings.TrimSpace(service), ":")
	if service == "" || strings.ContainsAny(service, " \t\r\n") {
		return nil, errors.New("realtime service name must be non-empty and contain no whitespace")
	}
	presenceTTL := defaultPresenceTTL
	if value := strings.TrimSpace(os.Getenv("REDIS_PRESENCE_TTL")); value != "" {
		parsed, err := time.ParseDuration(value)
		if err != nil || parsed <= 0 {
			return nil, errors.New("REDIS_PRESENCE_TTL must be a positive duration")
		}
		presenceTTL = parsed
	}
	presenceRefresh := defaultPresenceRefresh
	if value := strings.TrimSpace(os.Getenv("REDIS_PRESENCE_REFRESH")); value != "" {
		parsed, err := time.ParseDuration(value)
		if err != nil || parsed <= 0 {
			return nil, errors.New("REDIS_PRESENCE_REFRESH must be a positive duration")
		}
		presenceRefresh = parsed
	}
	if presenceRefresh >= presenceTTL {
		return nil, errors.New("REDIS_PRESENCE_REFRESH must be shorter than REDIS_PRESENCE_TTL")
	}
	hostname, err := os.Hostname()
	if err != nil || strings.TrimSpace(hostname) == "" {
		hostname = "unknown-host"
	}
	instanceID := strings.TrimSpace(os.Getenv("SERVICE_INSTANCE_ID"))
	if instanceID == "" {
		instanceID = fmt.Sprintf("%s-%d", hostname, os.Getpid())
	}
	return &Bus{
		store:           store,
		service:         service,
		instanceID:      instanceID,
		channel:         store.Key("realtime", service),
		presenceTTL:     presenceTTL,
		presenceRefresh: presenceRefresh,
	}, nil
}

func (b *Bus) Publish(parent context.Context, payload []byte) error {
	if !json.Valid(payload) {
		return errors.New("realtime payload must be valid JSON")
	}
	data, err := json.Marshal(envelope{Origin: b.instanceID, Payload: json.RawMessage(payload)})
	if err != nil {
		return fmt.Errorf("marshal realtime envelope: %w", err)
	}
	ctx, cancel := b.store.Context(parent)
	defer cancel()
	return b.store.Client().Publish(ctx, b.channel, data).Err()
}

// Subscribe consumes messages published by other instances of this service.
// Redis Pub/Sub is intentionally best-effort; persisted state remains the
// recovery source after disconnects or subscriber failures.
func (b *Bus) Subscribe(ctx context.Context, handle func([]byte)) error {
	pubsub := b.store.Client().Subscribe(ctx, b.channel)
	defer pubsub.Close()
	if _, err := pubsub.Receive(ctx); err != nil {
		return fmt.Errorf("subscribe to %s: %w", b.channel, err)
	}
	messages := pubsub.Channel(redis.WithChannelSize(256))
	for {
		select {
		case <-ctx.Done():
			return nil
		case message, ok := <-messages:
			if !ok {
				if ctx.Err() != nil {
					return nil
				}
				return fmt.Errorf("subscription channel %s closed", b.channel)
			}
			var item envelope
			if err := json.Unmarshal([]byte(message.Payload), &item); err != nil {
				log.Printf("Ignoring malformed realtime envelope on %s: %v", b.channel, err)
				continue
			}
			if item.Origin == b.instanceID || !json.Valid(item.Payload) {
				continue
			}
			handle(append([]byte(nil), item.Payload...))
		}
	}
}

func (b *Bus) MarkOnline(parent context.Context, userID int) error {
	ctx, cancel := b.store.Context(parent)
	defer cancel()
	return markOnlineScript.Run(
		ctx,
		b.store.Client(),
		[]string{b.presenceKey(userID)},
		b.instanceID,
		b.presenceTTL.Milliseconds(),
	).Err()
}

func (b *Bus) MarkOffline(parent context.Context, userID int) error {
	ctx, cancel := b.store.Context(parent)
	defer cancel()
	return b.store.Client().ZRem(ctx, b.presenceKey(userID), b.instanceID).Err()
}

func (b *Bus) IsOnline(parent context.Context, userID int) (bool, error) {
	ctx, cancel := b.store.Context(parent)
	defer cancel()
	count, err := isOnlineScript.Run(ctx, b.store.Client(), []string{b.presenceKey(userID)}).Int64()
	return count > 0, err
}

func (b *Bus) PresenceRefreshInterval() time.Duration {
	return b.presenceRefresh
}

func (b *Bus) presenceKey(userID int) string {
	return b.store.IntKey([]string{"presence", b.service}, userID)
}

// Compile-time check for accidental Transport drift.
var _ Transport = (*Bus)(nil)
