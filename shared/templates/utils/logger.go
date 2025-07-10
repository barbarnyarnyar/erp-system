package utils

import (
	"fmt"
	"log"
	"time"
	"github.com/gin-gonic/gin"
)

// Logger is a simple logger implementation
type Logger struct {
	serviceName string
}

// NewLogger creates a new logger instance
func NewLogger(serviceName string) *Logger {
	return &Logger{
		serviceName: serviceName,
	}
}

// Info logs info messages
func (l *Logger) Info(message string, args ...interface{}) {
	l.log("INFO", message, args...)
}

// Error logs error messages
func (l *Logger) Error(message string, args ...interface{}) {
	l.log("ERROR", message, args...)
}

// Debug logs debug messages
func (l *Logger) Debug(message string, args ...interface{}) {
	l.log("DEBUG", message, args...)
}

// Warn logs warning messages
func (l *Logger) Warn(message string, args ...interface{}) {
	l.log("WARN", message, args...)
}

// log is the internal logging method
func (l *Logger) log(level, message string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	formattedMessage := fmt.Sprintf(message, args...)
	logMessage := fmt.Sprintf("[%s] [%s] [%s] %s", timestamp, level, l.serviceName, formattedMessage)
	log.Println(logMessage)
}

// GinLogger returns a Gin middleware for logging HTTP requests
func (l *Logger) GinLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] [HTTP] [%s] %s %s %d %s \"%s\"\n",
			param.TimeStamp.Format("2006-01-02 15:04:05"),
			l.serviceName,
			param.ClientIP,
			param.Method,
			param.StatusCode,
			param.Latency,
			param.Path,
		)
	})
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware(serviceName string) gin.HandlerFunc {
	logger := NewLogger(serviceName)
	return func(c *gin.Context) {
		requestID := generateRequestID()
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Set("logger", logger)
		c.Next()
	}
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// GetLogger gets logger from Gin context
func GetLogger(c *gin.Context) *Logger {
	if logger, exists := c.Get("logger"); exists {
		return logger.(*Logger)
	}
	return NewLogger("unknown")
}

// SetupLogger configures the default logger
func SetupLogger(serviceName string) *Logger {
	return NewLogger(serviceName)
}