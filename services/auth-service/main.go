// File: services/auth-service/main.go
package main

import (
	"log"
	"auth-service/internal/config"
	"auth-service/internal/server"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Start server
	srv := server.New(cfg)
	if err := srv.Start(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
