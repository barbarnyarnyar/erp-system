package main

import (
	"log"
	"net/http"

	"api-gateway-bff/internal/config"
	"api-gateway-bff/internal/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	r := gin.Default()

	// Global middleware
	r.Use(gin.Recovery())

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	// Dashboard Handler setup
	dashHandler := handlers.NewDashboardHandler(cfg)

	v1 := r.Group("/api/v1")
	{
		ui := v1.Group("/ui")
		{
			ui.GET("/sales-dashboard/:order_id", dashHandler.GetSalesDashboard)
		}
	}

	log.Printf("Starting api-gateway-bff on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to run api-gateway-bff: %v", err)
	}
}
