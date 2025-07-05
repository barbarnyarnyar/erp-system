package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"fm-service/utils"
)

func main() {
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}

	serviceName := "fm-service"
	
	// Setup logger
	logger := utils.SetupLogger(serviceName)
	
	// Setup response helper
	responseHelper := utils.NewResponseHelper(serviceName)

	// Create Gin router
	r := gin.New()
	
	// Add middleware
	r.Use(utils.RequestIDMiddleware(serviceName))
	r.Use(logger.GinLogger())
	r.Use(gin.Recovery())

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		responseHelper.Health(c, port)
	})

	// Hello World endpoint
	r.GET("/api/fm/hello", func(c *gin.Context) {
		responseHelper.Success(c, "Hello World from Financial Management Service!", gin.H{
			"service": serviceName,
			"port":    port,
		})
	})

	// Root endpoint
	r.GET("/", func(c *gin.Context) {
		responseHelper.Success(c, "Financial Management Service is running", gin.H{
			"service": serviceName,
			"port":    port,
			"endpoints": []string{
				"GET /health - Health check",
				"GET /api/fm/hello - Hello world endpoint",
			},
		})
	})

	// Example error endpoint to demonstrate error handling
	r.GET("/api/fm/error", func(c *gin.Context) {
		responseHelper.InternalServerError(c, "This is a test error", nil)
	})

	logger.Info("Financial Management Service starting on port %s", port)
	
	// Start server
	if err := r.Run(":" + port); err != nil {
		logger.Error("Failed to start server: %s", err.Error())
	}
}