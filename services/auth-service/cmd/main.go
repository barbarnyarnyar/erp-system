package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"auth-service/utils"
)

func main() {
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	serviceName := "auth-service"
	
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
	r.GET("/api/auth/hello", func(c *gin.Context) {
		responseHelper.Success(c, "Hello World from Auth Service!", gin.H{
			"endpoints": []string{
				"POST /api/v1/auth/login - User Login",
				"POST /api/v1/auth/register - User Registration",
				"POST /api/v1/auth/refresh - Refresh Token",
				"POST /api/v1/auth/logout - User Logout",
			},
		})
	})

	// Root endpoint
	r.GET("/", func(c *gin.Context) {
		responseHelper.Success(c, "Auth Service is running", gin.H{
			"service": serviceName,
			"port":    port,
			"endpoints": []string{
				"GET /health - Health check",
				"GET /api/auth/hello - Hello world endpoint",
			},
		})
	})

	r.GET("/api/v1/auth/login", func(c *gin.Context) {
		responseHelper.Success(c, "Hello from Login!", gin.H{
			"module": "User Authentication",
			"features": []string{"JWT Tokens", "Session Management", "Password Validation"},
		})
	})

	r.GET("/api/v1/auth/users", func(c *gin.Context) {
		responseHelper.Success(c, "Hello from User Management!", gin.H{
			"module": "User Management",
			"features": []string{"User Profiles", "Role Management", "Permissions"},
		})
	})

	logger.Info("Auth Service starting on port %s", port)
	
	// Start server
	if err := r.Run(":" + port); err != nil {
		logger.Error("Failed to start server: %s", err.Error())
	}
}

// To use this updated version:
// 1. Update auth-service/go.mod to include shared dependency
// 2. Replace main.go content with this file content
// 3. Remove this file