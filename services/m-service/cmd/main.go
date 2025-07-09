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
		port = "8004"
	}

	// Create Gin router
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "m-service",
			"status":  "healthy",
			"port":    port,
		})
	})

	// Hello World endpoint
	r.GET("/api/manufacturing/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello World from Manufacturing Service!",
			"service": "m-service",
			"port":    port,
		})
	})

	// Root endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Manufacturing Service is running",
			"service": "m-service",
			"port":    port,
		})
	})

	// Start server
	r.Run(":" + port)}