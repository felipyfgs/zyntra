package middleware

import (
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/zyntra/backend/internal/api"
)

// RateLimiter provides rate limiting functionality
type RateLimiter struct {
	requests map[string]*rateLimitEntry
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

type rateLimitEntry struct {
	count     int
	expiresAt time.Time
}

// NewRateLimiter creates a new rate limiter
// limit: max requests per window
// window: time window duration
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*rateLimitEntry),
		limit:    limit,
		window:   window,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// Middleware returns an Echo middleware for rate limiting
func (rl *RateLimiter) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := rl.getKey(c)

			if !rl.allow(key) {
				c.Response().Header().Set("X-RateLimit-Limit", string(rune(rl.limit)))
				c.Response().Header().Set("X-RateLimit-Remaining", "0")
				return api.RateLimited(c)
			}

			remaining := rl.remaining(key)
			c.Response().Header().Set("X-RateLimit-Limit", string(rune(rl.limit)))
			c.Response().Header().Set("X-RateLimit-Remaining", string(rune(remaining)))

			return next(c)
		}
	}
}

// getKey returns the rate limit key for the request
func (rl *RateLimiter) getKey(c echo.Context) string {
	// Use API key ID if present
	if apiKey := GetAPIKey(c); apiKey != nil {
		return "apikey:" + apiKey.ID
	}

	// Use user ID if authenticated
	if user := GetUser(c); user != nil {
		return "user:" + user.UserID
	}

	// Fall back to IP address
	return "ip:" + c.RealIP()
}

// allow checks if the request should be allowed
func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	entry, exists := rl.requests[key]
	if !exists || now.After(entry.expiresAt) {
		rl.requests[key] = &rateLimitEntry{
			count:     1,
			expiresAt: now.Add(rl.window),
		}
		return true
	}

	if entry.count >= rl.limit {
		return false
	}

	entry.count++
	return true
}

// remaining returns the number of remaining requests
func (rl *RateLimiter) remaining(key string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	entry, exists := rl.requests[key]
	if !exists {
		return rl.limit
	}

	remaining := rl.limit - entry.count
	if remaining < 0 {
		return 0
	}
	return remaining
}

// cleanup periodically removes expired entries
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, entry := range rl.requests {
			if now.After(entry.expiresAt) {
				delete(rl.requests, key)
			}
		}
		rl.mu.Unlock()
	}
}

// DefaultRateLimiter returns a rate limiter with default settings
// 100 requests per minute for authenticated users
func DefaultRateLimiter() *RateLimiter {
	return NewRateLimiter(100, time.Minute)
}

// StrictRateLimiter returns a stricter rate limiter
// 20 requests per minute (for sensitive endpoints)
func StrictRateLimiter() *RateLimiter {
	return NewRateLimiter(20, time.Minute)
}

// APIKeyRateLimiter returns a rate limiter for API keys
// 1000 requests per minute
func APIKeyRateLimiter() *RateLimiter {
	return NewRateLimiter(1000, time.Minute)
}
