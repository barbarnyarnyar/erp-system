package main

import (
	"erp-system/shared/utils"
	"context"
	sharedkafka "erp-system/shared/kafka"
	"log"
	"net/http"

	"github.com/erp-system/fm-service/internal/api/handlers"
	"github.com/erp-system/fm-service/internal/api/routes"
	"github.com/erp-system/fm-service/internal/business/service"
	"github.com/erp-system/fm-service/internal/config"
	kafkaData "github.com/erp-system/fm-service/internal/data/kafka"
	"github.com/erp-system/fm-service/internal/data/memory"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	utils.InitLogger("fm-service")
	responseHelper := utils.NewResponseHelper("fm-service")


	// Initialize Kafka Publisher
	kafkaPublisher := sharedkafka.NewPublisher(cfg.Kafka.Brokers)
	defer kafkaPublisher.Close()

	// Initialize in-memory repositories
	accountRepo := memory.NewMemoryAccountRepo()
	entryRepo := memory.NewMemoryJournalEntryRepo()
	invoiceRepo := memory.NewMemoryInvoiceRepo()
	paymentRepo := memory.NewMemoryPaymentRepo()
	budgetRepo := memory.NewMemoryBudgetRepo()
	vendorBillRepo := memory.NewMemoryVendorBillRepo()

	// Register new memory repository instances (or SQL if DB connection is active, but here we inject memory)
	currencyRateRepo := memory.NewMemoryCurrencyRateRepo()
	fiscalYearRepo := memory.NewMemoryFiscalYearRepo()
	costCenterRepo := memory.NewMemoryCostCenterRepo()
	bankAccountRepo := memory.NewMemoryBankAccountRepo()
	customerCreditRepo := memory.NewMemoryCustomerCreditRepo()
	bankStatementRepo := memory.NewMemoryBankStatementRepo()
	transactionRepo := memory.NewMemoryTransactionRepo()

	// Suppress unused variables since they might not be added to service constructors yet
	_ = currencyRateRepo
	_ = fiscalYearRepo
	_ = costCenterRepo
	_ = bankAccountRepo
	_ = customerCreditRepo
	_ = bankStatementRepo
	_ = transactionRepo


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
		bankStatementRepo,
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
		kafkaPublisher,
		generalLedgerSvc,
		accountsPayableSvc,
		accountsReceivableSvc,
		cashManagementSvc,
		budgetingSvc,
	)
	go kafkaConsumer.Start(ctx)
	defer kafkaConsumer.Close()

	// Initialize handlers
	accHandler := handlers.NewAccountHandler(generalLedgerSvc, responseHelper)
	txHandler := handlers.NewTransactionHandler(generalLedgerSvc, responseHelper)
	repHandler := handlers.NewReportHandler(generalLedgerSvc, responseHelper)
	invHandler := handlers.NewInvoiceHandler(accountsReceivableSvc, responseHelper)
	payHandler := handlers.NewPaymentHandler(cashManagementSvc, responseHelper)
	billHandler := handlers.NewVendorBillHandler(accountsPayableSvc, responseHelper)

	// Initialize Gin router
	router := gin.Default()

	// Setup routes
	routes.SetupRoutes(router, cfg, accHandler, txHandler, repHandler, invHandler, payHandler, billHandler)

	// Start server
	log.Printf("Financial Management Service starting on port %s", cfg.Server.Port)
	if err := http.ListenAndServe(":"+cfg.Server.Port, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
