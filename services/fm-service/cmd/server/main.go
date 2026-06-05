package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/erp-system/fm-service/internal/api/handlers"
	"github.com/erp-system/fm-service/internal/api/routes"
	"github.com/erp-system/fm-service/internal/business/service"
	"github.com/erp-system/fm-service/internal/config"
	"github.com/erp-system/fm-service/internal/data/memory"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize in-memory repositories
	accountRepo := memory.NewMemoryAccountRepo()
	entryRepo := memory.NewMemoryJournalEntryRepo()
	invoiceRepo := memory.NewMemoryInvoiceRepo()
	paymentRepo := memory.NewMemoryPaymentRepo()
	vendorRepo := memory.NewMemoryVendorRepo()
	budgetRepo := memory.NewMemoryBudgetRepo()

	// Initialize application service
	financeSvc := service.NewFinanceService(
		accountRepo,
		entryRepo,
		invoiceRepo,
		paymentRepo,
		vendorRepo,
		budgetRepo,
	)

	// Initialize handlers
	accHandler := handlers.NewAccountHandler(financeSvc)
	txHandler := handlers.NewTransactionHandler(financeSvc)
	repHandler := handlers.NewReportHandler(financeSvc)

	// Initialize Gin router
	router := gin.Default()

	// Setup routes
	routes.SetupRoutes(router, cfg, accHandler, txHandler, repHandler)

	// Start server
	log.Printf("Financial Management Service starting on port %s", cfg.Server.Port)
	if err := http.ListenAndServe(":"+cfg.Server.Port, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}