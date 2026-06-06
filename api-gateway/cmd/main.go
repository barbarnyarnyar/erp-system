// api-gateway/cmd/main.go
package main

import (
    "log"

    "api-gateway/internal/config"
    "api-gateway/internal/server"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    log.Printf("Starting API Gateway on port %s...", cfg.Port)
    srv := server.New(cfg)
    if err := srv.Start(); err != nil {
        log.Fatalf("Failed to start API Gateway: %v", err)
    }
}