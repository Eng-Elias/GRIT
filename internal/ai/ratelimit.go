package ai

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	chatRateLimit  = 10               // requests per minute per IP
	chatBurstLimit = 10               // burst size matches rate
	cleanupInterval = 5 * time.Minute // evict stale entries
	staleDuration   = 10 * time.Minute
)

// RateLimiter enforces per-IP rate limiting for chat requests.
type RateLimiter struct {
	limiters sync.Map // map[string]*limiterEntry
	stop     chan struct{}
}

type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimiter creates a rate limiter and starts a background cleanup goroutine.
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{stop: make(chan struct{})}
	go rl.cleanup()
	return rl
}

// Allow returns true if the request from the given IP is within the rate limit.
func (rl *RateLimiter) Allow(ip string) bool {
	entry, _ := rl.limiters.LoadOrStore(ip, &limiterEntry{
		limiter:  rate.NewLimiter(rate.Every(time.Minute/chatRateLimit), chatBurstLimit),
		lastSeen: time.Now(),
	})
	le := entry.(*limiterEntry)
	le.lastSeen = time.Now()
	return le.limiter.Allow()
}

// Stop terminates the background cleanup goroutine.
func (rl *RateLimiter) Stop() {
	close(rl.stop)
}

// cleanup evicts limiter entries that haven't been seen for staleDuration.
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-rl.stop:
			return
		case <-ticker.C:
			cutoff := time.Now().Add(-staleDuration)
			rl.limiters.Range(func(key, value any) bool {
				le := value.(*limiterEntry)
				if le.lastSeen.Before(cutoff) {
					rl.limiters.Delete(key)
				}
				return true
			})
		}
	}
}
