package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type visitor struct {
	tokens    float64
	lastSeen  time.Time
}

// RateLimiter provides per-IP token-bucket rate limiting.
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     float64 // tokens per second
	burst    float64 // max tokens
}

// NewRateLimiter creates a rate limiter. rate is requests/second, burst is max burst size.
func NewRateLimiter(rate float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		burst:    float64(burst),
	}
	// Cleanup stale entries every 5 minutes
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			rl.cleanup()
		}
	}()
	return rl
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	cutoff := time.Now().Add(-10 * time.Minute)
	for ip, v := range rl.visitors {
		if v.lastSeen.Before(cutoff) {
			delete(rl.visitors, ip)
		}
	}
}

func (rl *RateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[ip]
	if !exists {
		rl.visitors[ip] = &visitor{tokens: rl.burst - 1, lastSeen: now}
		return true
	}

	// Refill tokens based on elapsed time
	elapsed := now.Sub(v.lastSeen).Seconds()
	v.tokens += elapsed * rl.rate
	if v.tokens > rl.burst {
		v.tokens = rl.burst
	}
	v.lastSeen = now

	if v.tokens < 1 {
		return false
	}
	v.tokens--
	return true
}

// Middleware returns a Gin middleware that rate-limits requests by client IP.
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !rl.allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests, please try again later",
			})
			return
		}
		c.Next()
	}
}
