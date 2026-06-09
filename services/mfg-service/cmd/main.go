package main

import (
	"context"
	sharedkafka "erp-system/shared/kafka"
	"log"
	"net/http"

	"github.com/erp-system/m-service/internal/api/handlers"
	"github.com/erp-system/m-service/internal/api/routes"
	"github.com/erp-system/m-service/internal/business/domain"
	"github.com/erp-system/m-service/internal/business/service"
	"github.com/erp-system/m-service/internal/config"
	"github.com/erp-system/m-service/internal/data/kafka"
	"github.com/erp-system/m-service/internal/data/memory"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Initialize Event Publisher (Kafka)
	publisher := sharedkafka.NewPublisher(cfg.Kafka.Brokers)
	defer func() {
		if err := publisher.Close(); err != nil {
			log.Printf("Failed to close Kafka publisher: %v", err)
		}
	}()

	// 3. Initialize Memory Repositories
	bomRepo := memory.NewMemoryBillOfMaterialsRepo()
	compRepo := memory.NewMemoryBOMComponentRepo()
	wcRepo := memory.NewMemoryWorkCenterRepo()
	routingRepo := memory.NewMemoryRoutingOperationRepo()
	poRepo := memory.NewMemoryProductionOrderRepo()
	woRepo := memory.NewMemoryWorkOrderRepo()
	laborRepo := memory.NewMemoryLaborReportRepo()
	machineRepo := memory.NewMemoryMachineLogRepo()
	qualityRepo := memory.NewMemoryQualityInspectionRepo()
	nonConfRepo := memory.NewMemoryNonConformanceRepo()
	equipRepo := memory.NewMemoryEquipmentRepo()
	maintRepo := memory.NewMemoryMaintenanceOrderRepo()
	costRepo := memory.NewMemoryCostingRecordRepo()

	// Seed some initial data for testing
	ctx := context.Background()
	// Seed a default BOM
	_ = bomRepo.Create(ctx, &domain.BillOfMaterials{
		ID:          "bom_default",
		ProductID:   "prod_default",
		Version:     "V1.0",
		Status:      "ACTIVE",
		Description: "Default BOM for Auto-Scheduled Production",
	})

	// 4. Initialize Services (Split Components)
	bomSvc := service.NewBOMService(bomRepo, compRepo, wcRepo, routingRepo, publisher)
	prodSvc := service.NewProductionService(poRepo, woRepo, bomRepo, compRepo, routingRepo, wcRepo, laborRepo, costRepo, publisher)
	qualitySvc := service.NewQualityService(qualityRepo, nonConfRepo, woRepo, publisher)
	qualitySvc.SetProductionService(prodSvc)
	maintSvc := service.NewMaintenanceService(machineRepo, equipRepo, maintRepo, publisher)
	costingSvc := service.NewCostingService(costRepo, poRepo, compRepo, publisher)

	prodSvc.SetMaintenanceService(maintSvc)
	prodSvc.SetQualityService(qualitySvc)
	prodSvc.SetCostingService(costingSvc)

	// 5. Initialize Handlers
	bomHandler := handlers.NewBOMHandler(bomSvc)
	prodHandler := handlers.NewProductionHandler(prodSvc)
	qualityHandler := handlers.NewQualityHandler(qualitySvc)
	maintHandler := handlers.NewMaintenanceHandler(maintSvc)
	costingHandler := handlers.NewCostingHandler(costingSvc)

	// 5b. Initialize Event Consumer (Kafka)
	ctxCancel, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer := kafka.NewKafkaConsumer(cfg.Kafka.Brokers, cfg.Kafka.GroupID, publisher, prodSvc)
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
			"service": "m-service",
			"status":  "healthy",
			"port":    cfg.Server.Port,
		})
	})

	// Hello World API
	r.GET("/api/manufacturing/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello World from Manufacturing Service!",
			"service": "m-service",
			"port":    cfg.Server.Port,
		})
	})

	// Register API Routes
	routes.RegisterRoutes(r, bomHandler, prodHandler, qualityHandler, maintHandler, costingHandler)

	// Start Server
	log.Printf("Starting Manufacturing Service on port %s...", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
