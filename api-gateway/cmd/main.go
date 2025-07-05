// File: api-gateway/main.go
package main

import (
	"log"
	"api-gateway/internal/config"
	"api-gateway/internal/server"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Start gateway server
	srv := server.New(cfg)
	if err := srv.Start(); err != nil {
		log.Fatal("Failed to start gateway:", err)
	}
}
