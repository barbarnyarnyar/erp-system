package main

import (
	"context"
	"log"
	"net/http"

	"github.com/erp-system/pm-service/internal/api/handlers"
	"github.com/erp-system/pm-service/internal/api/routes"
	"github.com/erp-system/pm-service/internal/business/service"
	"github.com/erp-system/pm-service/internal/config"
	"github.com/erp-system/pm-service/internal/data/kafka"
	"github.com/erp-system/pm-service/internal/data/sql"
	sharedkafka "erp-system/shared/kafka"
	"erp-system/shared/utils"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	utils.InitLogger("prj-service")

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
		log.Fatalf("Failed to initialize GORM database: %v", err)
	}

	projRepo := sql.NewSQLProjectRepository(db)
	wbsRepo := sql.NewSQLWbsNodeRepository(db)
	timeRepo := sql.NewSQLTimeLogRepository(db)
	outboxRepo := sql.NewSQLTransactionalOutboxRepository(db)
	inboxRepo := sql.NewSQLKafkaEventInboxRepository(db)

	// 4. Initialize Services
	projTrackingSvc := service.NewProjectTrackingService(db, projRepo)
	wbsSvc := service.NewWbsStructureService(db, projRepo, wbsRepo, outboxRepo)
	timeSvc := service.NewTimeTrackingService(db, projRepo, wbsRepo, timeRepo, outboxRepo)
	reliableSvc := service.NewReliableMessagingService(db, inboxRepo)

	// 5. Initialize Handlers
	handler := handlers.NewPrjHandler(
		projTrackingSvc,
		wbsSvc,
		timeSvc,
		projRepo,
		wbsRepo,
		timeRepo,
	)

	// 6. Initialize Event Consumer (Kafka)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer := kafka.NewKafkaConsumer(cfg.Kafka.Brokers, cfg.Kafka.GroupID, publisher, reliableSvc, projTrackingSvc)
	go consumer.Start(ctx)
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Printf("Failed to close Kafka consumer: %v", err)
		}
	}()

	// 7. Setup Gin Engine
	r := gin.Default()
	r.Use(utils.TracingMiddleware("prj-service"))

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "pm-service",
			"status":  "healthy",
			"port":    cfg.Server.Port,
		})
	})

	// Register API Routes
	routes.SetupPMRoutes(r, handler)

	// 8. Start Server
	log.Printf("Starting pm-service on port %s in %s mode", cfg.Server.Port, cfg.Server.Env)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
