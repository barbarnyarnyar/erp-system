package main

import (
	"context"
	"erp-system/shared/utils"
	sharedkafka "erp-system/shared/kafka"
	"log"
	"net/http"
	"time"

	"github.com/erp-system/scm-service/internal/api/handlers"
	"github.com/erp-system/scm-service/internal/api/routes"
	"github.com/erp-system/scm-service/internal/business/domain"
	"github.com/erp-system/scm-service/internal/business/service"
	"github.com/erp-system/scm-service/internal/config"
	"github.com/erp-system/scm-service/internal/data/kafka"
	"github.com/erp-system/scm-service/internal/data/sql"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	utils.InitLogger("scm-service")
	responseHelper := utils.NewResponseHelper("scm-service")

	// 2. Initialize Event Publisher (Kafka)
	publisher := sharedkafka.NewPublisher(cfg.Kafka.Brokers)
	defer func() {
		if err := publisher.Close(); err != nil {
			log.Printf("Failed to close Kafka publisher: %v", err)
		}
	}()

	// 3. Initialize GORM Database
	db, err := sql.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize GORM Transaction Manager
	tm := sql.NewGORMTransactionManager(db)

	// 4. Initialize SQL Repositories
	prodRepo := sql.NewSQLProductRepo(db)
	catRepo := sql.NewSQLProductCategoryRepo(db)
	locRepo := sql.NewSQLLocationRepo(db)
	supRepo := sql.NewSQLSupplierRepo(db)
	contRepo := sql.NewSQLVendorContractRepo(db)
	invRepo := sql.NewSQLInventoryItemRepo(db)
	moveRepo := sql.NewSQLInventoryMovementRepo(db)
	poRepo := sql.NewSQLPurchaseOrderRepo(db)
	lineRepo := sql.NewSQLPurchaseOrderLineRepo(db)
	reqRepo := sql.NewSQLPurchaseRequisitionRepo(db)
	reqLineRepo := sql.NewSQLPurchaseRequisitionLineRepo(db)
	recRepo := sql.NewSQLReceiptRepo(db)
	recLRepo := sql.NewSQLReceiptLineRepo(db)
	shipRepo := sql.NewSQLShipmentRepo(db)
	shipLRepo := sql.NewSQLShipmentLineRepo(db)
	forecastRepo := sql.NewSQLDemandForecastRepo(db)
	transferRepo := sql.NewSQLStockTransferRepo(db)
	inboxRepo := sql.NewSQLKafkaEventInboxRepo(db)
	outboxRepo := sql.NewSQLTransactionalOutboxRepo(db)

	// Seed default warehouse location
	if _, err := locRepo.GetByID(context.Background(), "loc_default"); err != nil {
		_ = locRepo.Create(context.Background(), &domain.Location{
			ID:           "loc_default",
			LocationCode: "WH-MAIN",
			LocationName: "Main Distribution Center",
			LocationType: "WAREHOUSE",
			IsActive:     true,
		})
	}

	// 5. Initialize Services
	prodSvc := service.NewProductManagementService(prodRepo, catRepo, locRepo, publisher)
	supSvc := service.NewSupplierManagementService(supRepo, contRepo, publisher)
	poSvc := service.NewPurchaseOrderService(poRepo, lineRepo, reqRepo, reqLineRepo, publisher, tm)
	invSvc := service.NewInventoryService(invRepo, moveRepo, transferRepo, publisher, tm)
	whSvc := service.NewWarehouseService(recRepo, recLRepo, shipRepo, shipLRepo, poRepo, lineRepo, invSvc, publisher, tm)
	demandSvc := service.NewDemandPlanningService(forecastRepo)
	reportSvc := service.NewReportService(prodRepo, invRepo, supRepo, poRepo, moveRepo, forecastRepo)

	// 6. Initialize Handlers
	prodHandler := handlers.NewProductHandler(prodSvc, responseHelper)
	vendorHandler := handlers.NewVendorHandler(supSvc, responseHelper)
	poHandler := handlers.NewPurchaseOrderHandler(poSvc, responseHelper)
	invHandler := handlers.NewInventoryHandler(invSvc, responseHelper)
	whHandler := handlers.NewWarehouseHandler(whSvc, responseHelper)
	demandHandler := handlers.NewDemandForecastHandler(demandSvc, responseHelper)
	reportHandler := handlers.NewReportHandler(reportSvc, responseHelper)

	// 7. Start Event Consumer (Kafka) & Outbox Relay Worker
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	outboxWorker := kafka.NewOutboxRelayWorker(outboxRepo, publisher, 5*time.Second, 100)
	go outboxWorker.Start(ctx)

	consumer := kafka.NewKafkaConsumer(cfg.Kafka.Brokers, cfg.Kafka.GroupID, publisher, poSvc, invSvc, demandSvc, inboxRepo)
	go consumer.Start(ctx)
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Printf("Failed to close Kafka consumer: %v", err)
		}
	}()

	// 8. Setup Gin Engine
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

	// 9. Start Server
	log.Printf("Starting scm-service on port %s in %s mode", cfg.Server.Port, cfg.Server.Env)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
