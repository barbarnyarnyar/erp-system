package main

import (
	"context"
	sharedkafka "erp-system/shared/kafka"
	"erp-system/shared/utils"
	"log"
	"net/http"
	"time"

	"github.com/erp-system/fm-service/internal/api/handlers"
	"github.com/erp-system/fm-service/internal/api/routes"
	"github.com/erp-system/fm-service/internal/business/service"
	"github.com/erp-system/fm-service/internal/config"
	kafkaData "github.com/erp-system/fm-service/internal/data/kafka"
	"github.com/erp-system/fm-service/internal/data/sql"
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

	// Initialize GORM Database
	db, err := sql.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize GORM Transaction Manager
	tm := sql.NewGORMTransactionManager(db)

	// Initialize SQL repositories
	accountRepo := sql.NewSQLChartOfAccountsRepo(db)
	entryRepo := sql.NewSQLUniversalJournalEntryRepo(db)
	invoiceRepo := sql.NewSQLArInvoiceRepo(db)
	paymentRepo := sql.NewSQLPaymentRepo(db)
	budgetRepo := sql.NewSQLBudgetRepo(db)
	vendorBillRepo := sql.NewSQLApVendorBillRepo(db)
	outboxRepo := sql.NewSQLTransactionalOutboxRepo(db)

	currencyRateRepo := sql.NewSQLCurrencyRateRepo(db)
	fiscalYearRepo := sql.NewSQLFiscalYearRepo(db)
	costCenterRepo := sql.NewSQLCostCenterRepo(db)
	bankAccountRepo := sql.NewSQLBankAccountRepo(db)
	customerCreditRepo := sql.NewSQLCustomerCreditRepo(db)
	bankStatementRepo := sql.NewSQLBankStatementRepo(db)

	legalEntityRepo := sql.NewSQLLegalEntityRepo(db)
	assetRepo := sql.NewSQLCapitalAssetRepo(db)
	lineRepo := sql.NewSQLDepreciationScheduleLineRepo(db)
	inboxRepo := sql.NewSQLKafkaEventInboxRepo(db)

	// Suppress unused variables to avoid compile errors
	_ = currencyRateRepo
	_ = fiscalYearRepo
	_ = costCenterRepo
	_ = bankAccountRepo
	_ = customerCreditRepo
	_ = bankStatementRepo

	// Initialize application services
	generalLedgerSvc := service.NewGeneralLedgerService(
		accountRepo,
		entryRepo,
		outboxRepo,
		tm,
	)
	accountsReceivableSvc := service.NewAccountsReceivableService(
		invoiceRepo,
		outboxRepo,
		tm,
	)
	cashManagementSvc := service.NewCashManagementService(
		paymentRepo,
		invoiceRepo,
		bankStatementRepo,
		outboxRepo,
		tm,
	)
	accountsPayableSvc := service.NewAccountsPayableService(
		vendorBillRepo,
		outboxRepo,
		tm,
	)
	budgetingSvc := service.NewBudgetingService(
		budgetRepo,
		accountRepo,
		entryRepo,
		outboxRepo,
		tm,
	)
	legalEntitySvc := service.NewLegalEntityService(
		legalEntityRepo,
		tm,
	)
	capitalAssetSvc := service.NewCapitalAssetService(
		assetRepo,
		lineRepo,
		accountRepo,
		entryRepo,
		outboxRepo,
		tm,
	)

	// Context for background processes
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize and start Outbox Relay Worker in the background
	outboxWorker := kafkaData.NewOutboxRelayWorker(outboxRepo, kafkaPublisher, 5*time.Second, 100)
	go outboxWorker.Start(ctx)

	// Initialize and start Kafka Consumer in the background
	kafkaConsumer := kafkaData.NewKafkaConsumer(
		cfg.Kafka.Brokers,
		cfg.Kafka.GroupID,
		kafkaPublisher,
		generalLedgerSvc,
		accountsPayableSvc,
		accountsReceivableSvc,
		cashManagementSvc,
		budgetingSvc,
		inboxRepo,
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
	leHandler := handlers.NewLegalEntityHandler(legalEntitySvc, responseHelper)
	assetHandler := handlers.NewAssetHandler(capitalAssetSvc, responseHelper)

	// Initialize Gin router
	router := gin.Default()

	// Setup routes
	routes.SetupRoutes(router, cfg, accHandler, txHandler, repHandler, invHandler, payHandler, billHandler, leHandler, assetHandler)

	// Start server
	log.Printf("Financial Management Service starting on port %s", cfg.Server.Port)
	if err := http.ListenAndServe(":"+cfg.Server.Port, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
