package edge

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultRatePerSecond = 20.0
	defaultBurst         = 40
	defaultMaxBodyBytes  = int64(12 << 20)
)

// Config controls the gateway protection layer.
type Config struct {
	RatePerSecond float64
	Burst         int
	MaxBodyBytes  int64
}

func DefaultConfig() Config {
	return Config{
		RatePerSecond: defaultRatePerSecond,
		Burst:         defaultBurst,
		MaxBodyBytes:  defaultMaxBodyBytes,
	}
}

func ConfigFromEnvironment() (Config, error) {
	config := DefaultConfig()
	var err error
	if value := strings.TrimSpace(os.Getenv("GATEWAY_RATE_LIMIT_RPS")); value != "" {
		config.RatePerSecond, err = strconv.ParseFloat(value, 64)
		if err != nil {
			return Config{}, errors.New("GATEWAY_RATE_LIMIT_RPS must be a positive number")
		}
	}
	if value := strings.TrimSpace(os.Getenv("GATEWAY_RATE_LIMIT_BURST")); value != "" {
		config.Burst, err = strconv.Atoi(value)
		if err != nil {
			return Config{}, errors.New("GATEWAY_RATE_LIMIT_BURST must be a positive integer")
		}
	}
	if value := strings.TrimSpace(os.Getenv("GATEWAY_MAX_BODY_BYTES")); value != "" {
		config.MaxBodyBytes, err = strconv.ParseInt(value, 10, 64)
		if err != nil {
			return Config{}, errors.New("GATEWAY_MAX_BODY_BYTES must be a positive integer")
		}
	}
	if config.RatePerSecond <= 0 || config.Burst <= 0 || config.MaxBodyBytes <= 0 {
		return Config{}, errors.New("gateway rate, burst, and body-size limits must be positive")
	}
	return config, nil
}

type visitor struct {
	tokens   float64
	lastSeen time.Time
}

// Limiter is the distributed rate-limit contract used by the Gateway.
type Limiter interface {
	Allow(context.Context, string) (bool, error)
}

// Controls applies security headers, body limits, and a distributed token
// bucket with a replica-local degraded fallback.
type Controls struct {
	config      Config
	limiter     Limiter
	mu          sync.Mutex
	visitors    map[string]*visitor
	now         func() time.Time
	lastCleanup time.Time
	lastErrorLog time.Time
}

func New(config Config) *Controls {
	return NewDistributed(config, nil)
}

func NewDistributed(config Config, limiter Limiter) *Controls {
	now := time.Now()
	return &Controls{
		config:      config,
		limiter:     limiter,
		visitors:    make(map[string]*visitor),
		now:         time.Now,
		lastCleanup: now,
	}
}

func (c *Controls) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setSecurityHeaders(w.Header())

		if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/api" {
			if !c.allow(r.Context(), clientIP(r.RemoteAddr)) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Cache-Control", "no-store")
				w.Header().Set("Retry-After", strconv.Itoa(max(1, int(math.Ceil(1/c.config.RatePerSecond)))))
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(map[string]any{"success": false, "error": "Rate limit exceeded"})
				return
			}

			if requestHasBody(r) {
				if r.ContentLength > c.config.MaxBodyBytes {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusRequestEntityTooLarge)
					_ = json.NewEncoder(w).Encode(map[string]any{"success": false, "error": "Request body too large"})
					return
				}
				r.Body = http.MaxBytesReader(w, r.Body, c.config.MaxBodyBytes)
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (c *Controls) allow(ctx context.Context, key string) bool {
	if c.limiter != nil {
		allowed, err := c.limiter.Allow(ctx, key)
		if err == nil {
			return allowed
		}
		c.logLimiterFailure(err)
	}
	return c.allowLocal(key)
}

func (c *Controls) logLimiterFailure(err error) {
	c.mu.Lock()
	if time.Since(c.lastErrorLog) < 10*time.Second {
		c.mu.Unlock()
		return
	}
	c.lastErrorLog = time.Now()
	c.mu.Unlock()
	log.Printf("Distributed Gateway rate limiter unavailable; using replica-local fallback: %v", err)
}

func (c *Controls) allowLocal(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.now()
	if now.Sub(c.lastCleanup) >= time.Minute {
		for visitorKey, entry := range c.visitors {
			if now.Sub(entry.lastSeen) > 10*time.Minute {
				delete(c.visitors, visitorKey)
			}
		}
		c.lastCleanup = now
	}

	entry, exists := c.visitors[key]
	if !exists {
		entry = &visitor{tokens: float64(c.config.Burst), lastSeen: now}
		c.visitors[key] = entry
	}
	elapsed := now.Sub(entry.lastSeen).Seconds()
	entry.tokens = math.Min(float64(c.config.Burst), entry.tokens+elapsed*c.config.RatePerSecond)
	entry.lastSeen = now
	if entry.tokens < 1 {
		return false
	}
	entry.tokens--
	return true
}

func clientIP(remoteAddress string) string {
	host, _, err := net.SplitHostPort(remoteAddress)
	if err == nil {
		return host
	}
	return remoteAddress
}

func requestHasBody(r *http.Request) bool {
	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		return true
	default:
		return false
	}
}

func setSecurityHeaders(header http.Header) {
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("X-Frame-Options", "DENY")
	header.Set("Referrer-Policy", "strict-origin-when-cross-origin")
	header.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
	header.Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' data: https://fonts.gstatic.com; img-src 'self' data: https:; connect-src 'self' ws: wss:; object-src 'none'; base-uri 'self'; frame-ancestors 'none'")
}
