// File: shared/middleware/auth_client.go (for individual services)
package middleware

import (
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
)

// JWTMiddleware for individual services (validates tokens via headers from gateway)
func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user context from headers set by API Gateway
		userIDStr := c.GetHeader("X-User-ID")
		username := c.GetHeader("X-Username")
		
		if userIDStr == "" || username == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing user context"})
			c.Abort()
			return
		}
		
		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
			c.Abort()
			return
		}
		
		// Set user context
		c.Set("user_id", uint(userID))
		c.Set("username", username)
		
		c.Next()
	}
}
