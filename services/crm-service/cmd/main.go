package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8002"
	}

	// Create Gin router
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "crm-service",
			"status":  "healthy",
			"port":    port,
		})
	})

	// Hello World endpoint
	r.GET("/api/crm/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello World from Customer Relationship Management Service!",
			"service": "crm-service",
			"port":    port,
		})
	})

	// Root endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Customer Relationship Management Service is running",
			"service": "crm-service",
			"port":    port,
		})
	})

	// Start server
	r.Run(":" + port)
}
