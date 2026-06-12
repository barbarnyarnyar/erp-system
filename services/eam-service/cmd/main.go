package main

import (
	"context"
	sharedkafka "erp-system/shared/kafka"
	"erp-system/shared/utils"
	"log"
	"net/http"

	"github.com/erp-system/eam-service/internal/api/handlers"
	"github.com/erp-system/eam-service/internal/api/routes"
	"github.com/erp-system/eam-service/internal/business/domain"
	"github.com/erp-system/eam-service/internal/business/service"
	"github.com/erp-system/eam-service/internal/config"
	"github.com/erp-system/eam-service/internal/data/kafka"
	"github.com/erp-system/eam-service/internal/data/memory"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	utils.InitLogger("eam-service")
	responseHelper := utils.NewResponseHelper("eam-service")

	// 2. Initialize Event Publisher (Kafka)
	publisher := sharedkafka.NewPublisher(cfg.Kafka.Brokers)
	defer func() {
		if err := publisher.Close(); err != nil {
			log.Printf("Failed to close Kafka publisher: %v", err)
		}
	}()

	// 3. Initialize Memory Repositories
	facRepo := memory.NewMemoryFacilityRepo()
	eqRepo := memory.NewMemoryEquipmentRepo()
	woRepo := memory.NewMemoryMaintenanceWorkOrderRepo()
	schRepo := memory.NewMemoryPreventativeScheduleRepo()
	bufRepo := memory.NewMemoryTelemetryIngestBufferRepo()

	// Seed default facility
	_ = facRepo.Create(context.Background(), &domain.Facility{
		ID:              "fac_default",
		LegalEntityID:   "tenant_default",
		Name:            "Default Facility",
		PhysicalAddress: "123 Main St",
		IsActive:        true,
	})

	// 4. Initialize Services
	eqSvc := service.NewEquipmentService(facRepo, eqRepo, publisher)
	maintSvc := service.NewMaintenanceService(woRepo, eqRepo, schRepo, publisher)
	telSvc := service.NewTelemetryIngestionService(bufRepo)

	// 5. Initialize Handlers
	eamHandler := handlers.NewEamHandler(eqSvc, maintSvc, telSvc, responseHelper)

	// 5b. Start Event Consumer (Kafka)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	consumer := kafka.NewKafkaConsumer(cfg.Kafka.Brokers, cfg.Kafka.GroupID, publisher, eqSvc, maintSvc)
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
			"service": "eam-service",
			"status":  "healthy",
			"port":    cfg.Server.Port,
		})
	})

	// Register Routes
	routes.RegisterRoutes(r, eamHandler)

	// 7. Start Server
	log.Printf("Starting eam-service on port %s in %s mode", cfg.Server.Port, cfg.Server.Env)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
