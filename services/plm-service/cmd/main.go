package main

import (
	"context"
	sharedkafka "erp-system/shared/kafka"
	"erp-system/shared/utils"
	"log"
	"net/http"

	"github.com/erp-system/plm-service/internal/api/handlers"
	"github.com/erp-system/plm-service/internal/api/routes"
	"github.com/erp-system/plm-service/internal/business/service"
	"github.com/erp-system/plm-service/internal/config"
	"github.com/erp-system/plm-service/internal/data/kafka"
	"github.com/erp-system/plm-service/internal/data/memory"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	utils.InitLogger("plm-service")
	responseHelper := utils.NewResponseHelper("plm-service")

	// 2. Initialize Event Publisher (Kafka)
	publisher := sharedkafka.NewPublisher(cfg.Kafka.Brokers)
	defer func() {
		if err := publisher.Close(); err != nil {
			log.Printf("Failed to close Kafka publisher: %v", err)
		}
	}()

	// 3. Initialize Memory Repositories
	matRepo := memory.NewMemoryMaterialMasterRepo()
	hdrRepo := memory.NewMemoryBomHeaderRepo()
	lineRepo := memory.NewMemoryBomLineRepo()
	ecoRepo := memory.NewMemoryEngineeringChangeOrderRepo()

	// 4. Initialize Services
	matSvc := service.NewMaterialService(matRepo, publisher)
	bomSvc := service.NewBomService(hdrRepo, lineRepo, matRepo, publisher)
	changeSvc := service.NewEngineeringChangeService(ecoRepo, matRepo, publisher)

	// 5. Initialize Handlers
	plmHandler := handlers.NewPlmHandler(matSvc, bomSvc, changeSvc, responseHelper)

	// 5b. Start Event Consumer (Kafka)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	consumer := kafka.NewKafkaConsumer(cfg.Kafka.Brokers, cfg.Kafka.GroupID, publisher, matSvc, bomSvc)
	go consumer.Start(ctx)
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Printf("Failed to close Kafka consumer: %v", err)
		}
	}()

	// 6. Setup Gin Engine
	r := gin.Default()

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "plm-service",
			"status":  "healthy",
			"port":    cfg.Server.Port,
		})
	})

	// Register Routes
	routes.RegisterRoutes(r, plmHandler)

	// 7. Start Server
	log.Printf("Starting plm-service on port %s in %s mode", cfg.Server.Port, cfg.Server.Env)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
