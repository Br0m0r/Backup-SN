package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements a simple rate limiter
type RateLimiter struct {
	visitors map[string]*visitor
	mutex    sync.RWMutex
}

type visitor struct {
	limiter  *time.Ticker
	lastSeen time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
	}

	// Clean up old visitors every minute
	go rl.cleanupVisitors()

	return rl
}

// RateLimit middleware limits requests per IP
func (rl *RateLimiter) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		rl.mutex.Lock()
		v, exists := rl.visitors[ip]
		if !exists {
			// Allow 10 requests per minute per IP
			v = &visitor{
				limiter:  time.NewTicker(6 * time.Second), // 10 requests/minute = 1 every 6 seconds
				lastSeen: time.Now(),
			}
			rl.visitors[ip] = v
		}
		v.lastSeen = time.Now()
		rl.mutex.Unlock()

		select {
		case <-v.limiter.C:
			// Request allowed
			next.ServeHTTP(w, r)
		default:
			// Rate limit exceeded
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		}
	})
}

// cleanupVisitors removes old visitors to prevent memory leaks
func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				v.limiter.Stop()
				delete(rl.visitors, ip)
			}
		}
		rl.mutex.Unlock()
	}
}
