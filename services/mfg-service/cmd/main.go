package main

import (
	"context"
	"log"
	"net/http"

	"erp-system/shared/utils"
	sharedkafka "erp-system/shared/kafka"
	"github.com/erp-system/m-service/internal/api/handlers"
	"github.com/erp-system/m-service/internal/api/routes"
	"github.com/erp-system/m-service/internal/business/service"
	"github.com/erp-system/m-service/internal/config"
	"github.com/erp-system/m-service/internal/data/kafka"
	"github.com/erp-system/m-service/internal/data/sql"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	utils.InitLogger("mfg-service")

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

	wcRepo := sql.NewSQLWorkCenterRepository(db)
	stationRepo := sql.NewSQLRoutingStationRepository(db)
	woRepo := sql.NewSQLWorkOrderRepository(db)
	stateRepo := sql.NewSQLWorkOrderRoutingStateRepository(db)
	consumeRepo := sql.NewSQLMaterialConsumptionLogRepository(db)
	yieldRepo := sql.NewSQLProductionYieldLogRepository(db)
	outboxRepo := sql.NewSQLTransactionalOutboxRepository(db)
	inboxRepo := sql.NewSQLKafkaEventInboxRepository(db)

	// 4. Initialize Services (Split Components)
	floorSvc := service.NewFloorConfigurationService(wcRepo, stationRepo)
	execSvc := service.NewWorkOrderExecutionService(db, woRepo, stateRepo, stationRepo, outboxRepo)
	teleSvc := service.NewShopFloorTelemetryService(db, woRepo, stationRepo, consumeRepo, yieldRepo, outboxRepo)
	reliableSvc := service.NewReliableMessagingService(db, inboxRepo)

	// 5. Initialize Handlers
	mfgHandler := handlers.NewMfgHandler(floorSvc, execSvc, teleSvc)

	// 5b. Initialize Event Consumer (Kafka)
	ctxCancel, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer := kafka.NewKafkaConsumer(cfg.Kafka.Brokers, cfg.Kafka.GroupID, publisher, reliableSvc, execSvc)
	go consumer.Start(ctxCancel)
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Printf("Failed to close Kafka consumer: %v", err)
		}
	}()

	// 6. Setup Gin Engine
	r := gin.Default()
	r.Use(utils.TracingMiddleware("mfg-service"))

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "mfg-service",
			"status":  "healthy",
			"port":    cfg.Server.Port,
		})
	})

	// Hello World API
	r.GET("/api/manufacturing/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello World from Manufacturing Service!",
			"service": "mfg-service",
			"port":    cfg.Server.Port,
		})
	})

	// Register API Routes
	routes.RegisterRoutes(r, mfgHandler)

	// Start Server
	log.Printf("Starting Manufacturing Service on port %s...", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
