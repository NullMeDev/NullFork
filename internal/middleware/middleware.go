package middleware

import (
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter controls request rate
func RateLimiter(rateLimit time.Duration) gin.HandlerFunc {
	ticker := time.NewTicker(rateLimit)
	return func(c *gin.Context) {
		select {
		case <-ticker.C:
			c.Next()
			return
		}
	}
}

// UserAgentRotator rotates User Agents from a list
func UserAgentRotator(userAgents []string) gin.HandlerFunc {
	random := rand.New(rand.NewSource(time.Now().Unix()))
	return func(c *gin.Context) {
		index := random.Intn(len(userAgents))
		c.Request.Header.Set("User-Agent", userAgents[index])
		c.Next()
	}
}
