package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/erp-system/fm-service/internal/api/routes"
	"github.com/erp-system/fm-service/internal/config"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Gin router
	router := gin.Default()

	// Setup routes
	routes.SetupRoutes(router, cfg)

	// Start server
	log.Printf("Financial Management Service starting on port %s", cfg.Server.Port)
	if err := http.ListenAndServe(":"+cfg.Server.Port, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}