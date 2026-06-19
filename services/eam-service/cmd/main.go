package main

import (
	"context"
	"log"
	"net/http"

	sharedkafka "erp-system/shared/kafka"
	"erp-system/shared/utils"
	"github.com/erp-system/eam-service/internal/api/handlers"
	"github.com/erp-system/eam-service/internal/api/routes"
	"github.com/erp-system/eam-service/internal/business/domain"
	"github.com/erp-system/eam-service/internal/business/service"
	"github.com/erp-system/eam-service/internal/config"
	"github.com/erp-system/eam-service/internal/data/kafka"
	"github.com/erp-system/eam-service/internal/data/sql"
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

	// 3. Initialize GORM Database & SQL Repositories
	db, err := sql.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	facRepo := sql.NewSQLFacilityRepository(db)
	eqRepo := sql.NewSQLEquipmentRepository(db)
	woRepo := sql.NewSQLMaintenanceWorkOrderRepository(db)
	schRepo := sql.NewSQLPreventativeScheduleRepository(db)
	bufRepo := sql.NewSQLTelemetryIngestBufferRepository(db)
	outboxRepo := sql.NewSQLTransactionalOutboxRepository(db)
	inboxRepo := sql.NewSQLKafkaEventInboxRepository(db)

	// Seed default facility if it doesn't exist
	ctx := context.Background()
	if _, err := facRepo.GetByID(ctx, "fac_default"); err != nil {
		_ = facRepo.Create(ctx, &domain.Facility{
			ID:              "fac_default",
			LegalEntityID:   "tenant_default",
			Name:            "Default Facility",
			PhysicalAddress: "123 Main St",
			IsActive:        true,
		})
	}

	// 4. Initialize Services
	reliableSvc := service.NewReliableMessagingService(db, inboxRepo, outboxRepo)
	eqSvc := service.NewEquipmentService(db, facRepo, eqRepo, reliableSvc)
	maintSvc := service.NewMaintenanceService(db, woRepo, eqRepo, schRepo, reliableSvc)
	telSvc := service.NewTelemetryIngestionService(db, bufRepo)

	// 5. Initialize Handlers
	eamHandler := handlers.NewEamHandler(eqSvc, maintSvc, telSvc, responseHelper)

	// 5b. Start Event Consumer (Kafka)
	ctxCancel, cancel := context.WithCancel(context.Background())
	defer cancel()
	consumer := kafka.NewKafkaConsumer(cfg.Kafka.Brokers, cfg.Kafka.GroupID, publisher, reliableSvc, eqSvc, maintSvc)
	go consumer.Start(ctxCancel)
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Printf("Failed to close Kafka consumer: %v", err)
		}
	}()

	// 6. Setup Gin Engine
	r := gin.Default()
	r.Use(utils.TracingMiddleware("eam-service"))

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
