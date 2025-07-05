// shared/middleware/logging.go
package middleware

import (
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/sirupsen/logrus"
)

func RequestLogger(logger *logrus.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        requestID := uuid.New().String()
        c.Set("request_id", requestID)
        
        logger.WithFields(logrus.Fields{
            "request_id": requestID,
            "method":     c.Request.Method,
            "path":       c.Request.URL.Path,
            "ip":         c.ClientIP(),
        }).Info("Request started")
        
        c.Next()
        
        duration := time.Since(start)
        logger.WithFields(logrus.Fields{
            "request_id": requestID,
            "method":     c.Request.Method,
            "path":       c.Request.URL.Path,
            "status":     c.Writer.Status(),
            "duration":   duration,
        }).Info("Request completed")
    }
}
