package main

import (
	"os"

	"fm-service/docs"
	"fm-service/utils"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Financial Management Service API
// @version 1.0
// @description This is the Financial Management Service for ERP System
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8001
// @BasePath /api/fm
// @schemes http

func main() {
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}

	serviceName := "fm-service"
	
	// Setup logger and response helper
	logger := utils.SetupLogger(serviceName)
	responseHelper := utils.NewResponseHelper(serviceName)

	// Create Gin router
	r := gin.New()
	
	// Add middleware
	r.Use(utils.RequestIDMiddleware(serviceName))
	r.Use(logger.GinLogger())
	r.Use(gin.Recovery())

	// Swagger docs
	docs.SwaggerInfo.Host = "localhost:" + port
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check endpoint
	// @Summary Health Check
	// @Description Check if the service is healthy
	// @Tags health
	// @Accept json
	// @Produce json
	// @Success 200 {object} utils.HealthResponse
	// @Router /health [get]
	r.GET("/health", func(c *gin.Context) {
		responseHelper.Health(c, port)
	})

	// API group
	api := r.Group("/api/fm")
	{
		// Hello World endpoint
		// @Summary Hello World
		// @Description Returns a hello world message
		// @Tags general
		// @Accept json
		// @Produce json
		// @Success 200 {object} utils.StandardResponse
		// @Router /hello [get]
		api.GET("/hello", func(c *gin.Context) {
			responseHelper.Success(c, "Hello World from Financial Management Service!", gin.H{
				"service": serviceName,
				"port":    port,
			})
		})
	}

	// Root endpoint
	// @Summary Root
	// @Description Returns service information and available endpoints
	// @Tags general
	// @Accept json
	// @Produce json
	// @Success 200 {object} utils.StandardResponse
	// @Router / [get]
	r.GET("/", func(c *gin.Context) {
		responseHelper.Success(c, "Financial Management Service is running", gin.H{
			"service": serviceName,
			"port":    port,
			"endpoints": []string{
				"GET /health - Health check",
				"GET /api/fm/hello - Hello world endpoint",
				"GET /swagger/index.html - API Documentation",
			},
		})
	})

	logger.Info("Financial Management Service starting on port %s", port)
	logger.Info("Swagger documentation available at: http://localhost:%s/swagger/index.html", port)
	
	// Start server
	if err := r.Run(":" + port); err != nil {
		logger.Error("Failed to start server: %s", err.Error())
	}
}