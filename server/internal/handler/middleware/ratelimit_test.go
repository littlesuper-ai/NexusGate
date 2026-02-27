package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRateLimiter_AllowsWithinBurst(t *testing.T) {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     1.0, // 1 req/sec
		burst:    5,
	}

	for i := 0; i < 5; i++ {
		if !rl.allow("192.168.1.1") {
			t.Errorf("request %d should be allowed within burst", i+1)
		}
	}
}

func TestRateLimiter_DeniesOverBurst(t *testing.T) {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     0.001, // very slow refill
		burst:    3,
	}

	// Exhaust burst
	for i := 0; i < 3; i++ {
		rl.allow("192.168.1.1")
	}

	// Next request should be denied
	if rl.allow("192.168.1.1") {
		t.Error("request should be denied after burst exhausted")
	}
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     0.001,
		burst:    2,
	}

	// Exhaust IP 1
	rl.allow("1.1.1.1")
	rl.allow("1.1.1.1")
	if rl.allow("1.1.1.1") {
		t.Error("IP 1 should be denied")
	}

	// IP 2 should still be allowed
	if !rl.allow("2.2.2.2") {
		t.Error("IP 2 should be allowed (independent bucket)")
	}
}

func TestRateLimiter_Middleware_Returns429(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     0.001,
		burst:    1,
	}

	r := gin.New()
	r.Use(rl.Middleware())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// First request allowed
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("first request: status = %d, want %d", w.Code, http.StatusOK)
	}

	// Second request should be rate limited
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req)

	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("second request: status = %d, want %d", w2.Code, http.StatusTooManyRequests)
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     1.0,
		burst:    5,
	}

	// Add a visitor
	rl.allow("1.1.1.1")
	if len(rl.visitors) != 1 {
		t.Fatalf("visitors count = %d, want 1", len(rl.visitors))
	}

	// Simulate stale entry (manually set lastSeen far in the past)
	rl.mu.Lock()
	for _, v := range rl.visitors {
		v.lastSeen = v.lastSeen.Add(-20 * 60 * 1e9) // 20 minutes ago
	}
	rl.mu.Unlock()

	rl.cleanup()

	rl.mu.Lock()
	count := len(rl.visitors)
	rl.mu.Unlock()

	if count != 0 {
		t.Errorf("visitors count after cleanup = %d, want 0", count)
	}
}
