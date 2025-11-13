package middleware

import (
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	visitors map[string]*visitor
	mutex    sync.RWMutex
}

type visitor struct {
	tokens     int
	lastSeen   time.Time
	maxTokens  int
	refillRate time.Duration
}

func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
	}
	go rl.cleanupVisitors()
	go rl.refillTokens()
	return rl
}

func (rl *RateLimiter) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		rl.mutex.Lock()
		v, exists := rl.visitors[ip]
		if !exists {
			v = &visitor{
				tokens:     10, // Start with 10 tokens
				maxTokens:  10, // Max 10 tokens
				lastSeen:   time.Now(),
				refillRate: time.Second, // Refill 1 token per second
			}
			rl.visitors[ip] = v
		}
		v.lastSeen = time.Now()

		if v.tokens > 0 {
			v.tokens--
			rl.mutex.Unlock()
			next.ServeHTTP(w, r)
		} else {
			rl.mutex.Unlock()
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		}
	})
}

func (rl *RateLimiter) refillTokens() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()
		for _, v := range rl.visitors {
			if v.tokens < v.maxTokens {
				v.tokens++
			}
		}
		rl.mutex.Unlock()
	}
}

func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mutex.Unlock()
	}
}
