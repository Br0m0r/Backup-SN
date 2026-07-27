package edge

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"social-network/services/common/redisstore"

	"github.com/redis/go-redis/v9"
)

var tokenBucketScript = redis.NewScript(`
local now = redis.call("TIME")
local now_ms = (now[1] * 1000) + math.floor(now[2] / 1000)
local values = redis.call("HMGET", KEYS[1], "tokens", "updated_at")
local burst = tonumber(ARGV[1])
local rate = tonumber(ARGV[2])
local tokens = tonumber(values[1])
local updated_at = tonumber(values[2])

if tokens == nil then
  tokens = burst
  updated_at = now_ms
end

local elapsed = math.max(0, now_ms - updated_at) / 1000
tokens = math.min(burst, tokens + (elapsed * rate))
local allowed = 0
if tokens >= 1 then
  tokens = tokens - 1
  allowed = 1
end

redis.call("HSET", KEYS[1], "tokens", tokens, "updated_at", now_ms)
local ttl_ms = math.max(1000, math.ceil((burst / rate) * 2000))
redis.call("PEXPIRE", KEYS[1], ttl_ms)
return allowed
`)

// RedisLimiter applies one atomic token bucket shared by every Gateway replica.
type RedisLimiter struct {
	store *redisstore.Store
	rate  float64
	burst int
}

func NewRedisLimiter(store *redisstore.Store, config Config) *RedisLimiter {
	return &RedisLimiter{store: store, rate: config.RatePerSecond, burst: config.Burst}
}

func (l *RedisLimiter) Allow(parent context.Context, identity string) (bool, error) {
	hash := sha256.Sum256([]byte(identity))
	key := l.store.Key("rate-limit", "gateway", hex.EncodeToString(hash[:]))
	ctx, cancel := l.store.Context(parent)
	defer cancel()
	allowed, err := tokenBucketScript.Run(
		ctx,
		l.store.Client(),
		[]string{key},
		l.burst,
		l.rate,
	).Int()
	if err != nil {
		return false, fmt.Errorf("evaluate Redis token bucket: %w", err)
	}
	return allowed == 1, nil
}

var _ Limiter = (*RedisLimiter)(nil)
