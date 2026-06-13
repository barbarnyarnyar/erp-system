package main

import (
	"context"
	"log"
	"net/http"

	sharedkafka "erp-system/shared/kafka"
	"erp-system/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/erp-system/qms-service/internal/api/handlers"
	"github.com/erp-system/qms-service/internal/api/routes"
	"github.com/erp-system/qms-service/internal/business/domain"
	"github.com/erp-system/qms-service/internal/business/service"
	"github.com/erp-system/qms-service/internal/config"
	"github.com/erp-system/qms-service/internal/data/kafka"
	"github.com/erp-system/qms-service/internal/data/sql"
)

func main() {
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	utils.InitLogger("qms-service")
	responseHelper := utils.NewResponseHelper("qms-service")

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

	planRepo := sql.NewSQLInspectionPlanRepository(db)
	metricRepo := sql.NewSQLInspectionMetricDefinitionRepository(db)
	qiRepo := sql.NewSQLQualityInspectionRepository(db)
	resRepo := sql.NewSQLInspectionResultLineRepository(db)
	ncRepo := sql.NewSQLNonConformanceLogRepository(db)
	outboxRepo := sql.NewSQLTransactionalOutboxRepository(db)
	inboxRepo := sql.NewSQLKafkaEventInboxRepository(db)

	// Seed default inspection plan if it doesn't exist
	ctx := context.Background()
	if _, err := planRepo.GetByID(ctx, "plan_default"); err != nil {
		_ = planRepo.Create(ctx, &domain.InspectionPlan{
			ID:            "plan_default",
			LegalEntityID: "tenant_default",
			MaterialID:    "mat_default",
			PlanName:      "Default Receiving Plan",
			IsActive:      true,
			Version:       1,
		})
	}

	// 4. Initialize Services
	reliableSvc := service.NewReliableMessagingService(db, inboxRepo, outboxRepo)
	planSvc := service.NewInspectionPlanService(db, planRepo, metricRepo)
	ncSvc := service.NewNonConformanceService(db, ncRepo, planRepo, qiRepo, reliableSvc)
	execSvc := service.NewInspectionExecutionService(db, qiRepo, resRepo, planRepo, ncSvc, reliableSvc)
	analySvc := service.NewQualityAnalyticsService(db, resRepo)

	// 5. Initialize Handlers
	qmsHandler := handlers.NewQmsHandler(planSvc, execSvc, ncSvc, analySvc, responseHelper)

	// 5b. Start Event Consumer (Kafka)
	ctxCancel, cancel := context.WithCancel(context.Background())
	defer cancel()
	consumer := kafka.NewKafkaConsumer(cfg.Kafka.Brokers, cfg.Kafka.GroupID, publisher, reliableSvc, planSvc, execSvc)
	go consumer.Start(ctxCancel)
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
			"service": "qms-service",
			"status":  "healthy",
			"port":    cfg.Server.Port,
		})
	})

	// Register Routes
	routes.RegisterRoutes(r, qmsHandler)

	// 7. Start Server
	log.Printf("Starting qms-service on port %s in %s mode", cfg.Server.Port, cfg.Server.Env)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
