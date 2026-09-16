package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

const (
	requestsPerMinute = 10
	burstSize         = 10
	visitorTTL        = 3 * time.Minute
	cleanupInterval   = time.Minute
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	mu       sync.Mutex
	visitors = make(map[string]*visitor)
)

func getVisitor(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	v, exists := visitors[ip]

	if !exists {
		limiter := rate.NewLimiter(
			rate.Every(time.Minute/requestsPerMinute),
			burstSize,
		)

		visitors[ip] = &visitor{
			limiter:  limiter,
			lastSeen: time.Now(),
		}

		return limiter
	}

	v.lastSeen = time.Now()

	return v.limiter
}

func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		limiter := getVisitor(ip)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests, please try again later",
			})

			c.Abort()
			return
		}

		c.Next()
	}
}

func cleanupVisitors() {
	for {
		time.Sleep(cleanupInterval)

		mu.Lock()

		for ip, v := range visitors {
			if time.Since(v.lastSeen) > visitorTTL {
				delete(visitors, ip)
			}
		}

		mu.Unlock()
	}
}

func init() {
	go cleanupVisitors()
}