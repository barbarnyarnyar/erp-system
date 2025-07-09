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
		port = "8005"
	}

	// Create Gin router
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "pm-service",
			"status":  "healthy",
			"port":    port,
		})
	})

	// Hello World endpoint
	r.GET("/api/projects/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello World from Project Management Service!",
			"service": "pm-service",
			"port":    port,
		})
	})

	// Root endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Project Management Service is running",
			"service": "pm-service",
			"port":    port,
		})
	})

	// Start server
	r.Run(":" + port)}