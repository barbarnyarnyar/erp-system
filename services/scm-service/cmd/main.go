package main

import (
	"context"
	"log"
	"net/http"

	"github.com/erp-system/scm-service/internal/api/handlers"
	"github.com/erp-system/scm-service/internal/api/routes"
	"github.com/erp-system/scm-service/internal/business/domain"
	"github.com/erp-system/scm-service/internal/business/service"
	"github.com/erp-system/scm-service/internal/config"
	"github.com/erp-system/scm-service/internal/data/kafka"
	"github.com/erp-system/scm-service/internal/data/memory"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Initialize Event Publisher (Kafka)
	publisher := kafka.NewKafkaPublisher(cfg.Kafka.Brokers)
	defer func() {
		if err := publisher.Close(); err != nil {
			log.Printf("Failed to close Kafka publisher: %v", err)
		}
	}()

	// 3. Initialize Memory Repositories
	prodRepo := memory.NewMemoryProductRepo()
	catRepo := memory.NewMemoryProductCategoryRepo()
	locRepo := memory.NewMemoryLocationRepo()
	supRepo := memory.NewMemorySupplierRepo()
	contRepo := memory.NewMemoryVendorContractRepo()
	invRepo := memory.NewMemoryInventoryItemRepo()
	moveRepo := memory.NewMemoryInventoryMovementRepo()
	poRepo := memory.NewMemoryPurchaseOrderRepo()
	lineRepo := memory.NewMemoryPurchaseOrderLineRepo()
	reqRepo := memory.NewMemoryPurchaseRequisitionRepo()
	reqLineRepo := memory.NewMemoryPurchaseRequisitionLineRepo()
	recRepo := memory.NewMemoryReceiptRepo()
	recLRepo := memory.NewMemoryReceiptLineRepo()
	shipRepo := memory.NewMemoryShipmentRepo()
	shipLRepo := memory.NewMemoryShipmentLineRepo()
	forecastRepo := memory.NewMemoryDemandForecastRepo()

	// Seed default warehouse location
	_ = locRepo.Create(context.Background(), &domain.Location{
		ID:           "loc_default",
		LocationCode: "WH-MAIN",
		LocationName: "Main Distribution Center",
		LocationType: "WAREHOUSE",
		IsActive:     true,
	})

	// 4. Initialize Services
	prodSvc := service.NewProductManagementService(prodRepo, catRepo)
	supSvc := service.NewSupplierManagementService(supRepo, contRepo)
	poSvc := service.NewPurchaseOrderService(poRepo, lineRepo, reqRepo, reqLineRepo, publisher)
	invSvc := service.NewInventoryService(invRepo, moveRepo, publisher)
	whSvc := service.NewWarehouseService(recRepo, recLRepo, shipRepo, shipLRepo, poRepo, lineRepo, invSvc)
	demandSvc := service.NewDemandPlanningService(forecastRepo)
	reportSvc := service.NewReportService(prodRepo, invRepo, supRepo, poRepo, moveRepo, forecastRepo)

	// 5. Initialize Handlers
	prodHandler := handlers.NewProductHandler(prodSvc)
	vendorHandler := handlers.NewVendorHandler(supSvc)
	poHandler := handlers.NewPurchaseOrderHandler(poSvc)
	invHandler := handlers.NewInventoryHandler(invSvc)
	whHandler := handlers.NewWarehouseHandler(whSvc)
	demandHandler := handlers.NewDemandForecastHandler(demandSvc)
	reportHandler := handlers.NewReportHandler(reportSvc)

	// 5b. Start Event Consumer (Kafka)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	consumer := kafka.NewKafkaConsumer(cfg.Kafka.Brokers, cfg.Kafka.GroupID, poSvc, invSvc, demandSvc)
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
			"service": "scm-service",
			"status":  "healthy",
			"port":    cfg.Server.Port,
		})
	})

	// Register Routes
	routes.RegisterRoutes(
		r,
		prodHandler,
		vendorHandler,
		poHandler,
		invHandler,
		whHandler,
		demandHandler,
		reportHandler,
	)

	// 7. Start Server
	log.Printf("Starting scm-service on port %s in %s mode", cfg.Server.Port, cfg.Server.Env)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}