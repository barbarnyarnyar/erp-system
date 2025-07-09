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
		port = "8003"
	}

	// Create Gin router
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "hr-service",
			"status":  "healthy",
			"port":    port,
		})
	})

	// Hello World endpoint
	r.GET("/api/hr/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello World from Human Resources Service!",
			"service": "hr-service",
			"port":    port,
		})
	})

	// Root endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Human Resources Service is running",
			"service": "hr-service",
			"port":    port,
		})
	})

	// Start server
	r.Run(":" + port)}