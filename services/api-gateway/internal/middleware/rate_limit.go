// File: api-gateway/internal/middleware/rate_limit.go
package middleware

import (
	"net/http"
	"sync"
	"time"
	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	clients map[string]*ClientInfo
	mutex   sync.RWMutex
	limit   int
	window  time.Duration
}

type ClientInfo struct {
	requests  int
	lastReset time.Time
}

func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	return &RateLimiter{
		clients: make(map[string]*ClientInfo),
		limit:   requestsPerMinute,
		window:  time.Minute,
	}
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		
		rl.mutex.Lock()
		defer rl.mutex.Unlock()
		
		client, exists := rl.clients[clientIP]
		if !exists {
			client = &ClientInfo{
				requests:  0,
				lastReset: time.Now(),
			}
			rl.clients[clientIP] = client
		}
		
		// Reset counter if window has passed
		if time.Since(client.lastReset) > rl.window {
			client.requests = 0
			client.lastReset = time.Now()
		}
		
		// Check if limit exceeded
		if client.requests >= rl.limit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"retry_after": int(rl.window.Seconds()),
			})
			c.Abort()
			return
		}
		
		client.requests++
		c.Next()
	}
}