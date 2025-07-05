package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"m-service/utils" // Assuming utils package is in the same directory structure
)

func main() {
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}

	// Create Gin router
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "fm-service",
			"status":  "healthy",
			"port":    port,
		})
	})

	// Hello World endpoint
	r.GET("/api/fm/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello World from Financial Management Service!",
			"service": "fm-service",
			"port":    port,
		})
	})

	// Root endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Financial Management Service is running",
			"service": "fm-service",
			"port":    port,
		})
	})

	// Start server
	r.Run(":" + port)
}