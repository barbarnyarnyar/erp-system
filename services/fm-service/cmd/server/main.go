package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/erp-system/fm-service/internal/api/handlers"
	"github.com/erp-system/fm-service/internal/api/routes"
	"github.com/erp-system/fm-service/internal/business/service"
	"github.com/erp-system/fm-service/internal/config"
	"github.com/erp-system/fm-service/internal/data/memory"
	kafkaData "github.com/erp-system/fm-service/internal/data/kafka"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Kafka Publisher
	kafkaPublisher := kafkaData.NewKafkaPublisher(cfg.Kafka.Brokers)
	defer kafkaPublisher.Close()

	// Initialize in-memory repositories
	accountRepo := memory.NewMemoryAccountRepo()
	entryRepo := memory.NewMemoryJournalEntryRepo()
	invoiceRepo := memory.NewMemoryInvoiceRepo()
	paymentRepo := memory.NewMemoryPaymentRepo()
	budgetRepo := memory.NewMemoryBudgetRepo()
	vendorBillRepo := memory.NewMemoryVendorBillRepo()

	// Initialize application services
	generalLedgerSvc := service.NewGeneralLedgerService(
		accountRepo,
		entryRepo,
		kafkaPublisher,
	)
	accountsReceivableSvc := service.NewAccountsReceivableService(
		invoiceRepo,
		kafkaPublisher,
	)
	cashManagementSvc := service.NewCashManagementService(
		paymentRepo,
		invoiceRepo,
		kafkaPublisher,
	)
	accountsPayableSvc := service.NewAccountsPayableService(
		vendorBillRepo,
		kafkaPublisher,
	)
	budgetingSvc := service.NewBudgetingService(
		budgetRepo,
		accountRepo,
		kafkaPublisher,
	)

	// Initialize and start Kafka Consumer in the background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kafkaConsumer := kafkaData.NewKafkaConsumer(
		cfg.Kafka.Brokers,
		cfg.Kafka.GroupID,
		generalLedgerSvc,
		accountsPayableSvc,
		accountsReceivableSvc,
		cashManagementSvc,
		budgetingSvc,
	)
	go kafkaConsumer.Start(ctx)
	defer kafkaConsumer.Close()

	// Initialize handlers
	accHandler := handlers.NewAccountHandler(generalLedgerSvc)
	txHandler := handlers.NewTransactionHandler(generalLedgerSvc)
	repHandler := handlers.NewReportHandler(generalLedgerSvc)
	invHandler := handlers.NewInvoiceHandler(accountsReceivableSvc)
	payHandler := handlers.NewPaymentHandler(cashManagementSvc)

	// Initialize Gin router
	router := gin.Default()

	// Setup routes
	routes.SetupRoutes(router, cfg, accHandler, txHandler, repHandler, invHandler, payHandler)

	// Start server
	log.Printf("Financial Management Service starting on port %s", cfg.Server.Port)
	if err := http.ListenAndServe(":"+cfg.Server.Port, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}