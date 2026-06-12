package main

import (
	"context"
	sharedkafka "erp-system/shared/kafka"
	"erp-system/shared/utils"
	"log"
	"net/http"

	"github.com/erp-system/qms-service/internal/api/handlers"
	"github.com/erp-system/qms-service/internal/api/routes"
	"github.com/erp-system/qms-service/internal/business/domain"
	"github.com/erp-system/qms-service/internal/business/service"
	"github.com/erp-system/qms-service/internal/config"
	"github.com/erp-system/qms-service/internal/data/kafka"
	"github.com/erp-system/qms-service/internal/data/memory"
	"github.com/gin-gonic/gin"
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

	// 3. Initialize Memory Repositories
	planRepo := memory.NewMemoryInspectionPlanRepo()
	metricRepo := memory.NewMemoryInspectionMetricDefinitionRepo()
	qiRepo := memory.NewMemoryQualityInspectionRepo()
	resRepo := memory.NewMemoryInspectionResultLineRepo()
	ncRepo := memory.NewMemoryNonConformanceLogRepo()

	// Seed default inspection plan
	_ = planRepo.Create(context.Background(), &domain.InspectionPlan{
		ID:            "plan_default",
		LegalEntityID: "tenant_default",
		MaterialID:    "mat_default",
		PlanName:      "Default Receiving Plan",
		IsActive:      true,
		Version:       1,
	})

	// 4. Initialize Services
	planSvc := service.NewInspectionPlanService(planRepo, metricRepo)
	ncSvc := service.NewNonConformanceService(ncRepo, planRepo, qiRepo, publisher)
	execSvc := service.NewInspectionExecutionService(qiRepo, resRepo, planRepo, ncSvc, publisher)
	analySvc := service.NewQualityAnalyticsService(resRepo)

	// 5. Initialize Handlers
	qmsHandler := handlers.NewQmsHandler(planSvc, execSvc, ncSvc, analySvc, responseHelper)

	// 5b. Start Event Consumer (Kafka)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	consumer := kafka.NewKafkaConsumer(cfg.Kafka.Brokers, cfg.Kafka.GroupID, publisher, planSvc, execSvc)
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
