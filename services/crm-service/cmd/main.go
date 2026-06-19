package main

import (
	"erp-system/shared/utils"
	"context"
	sharedkafka "erp-system/shared/kafka"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/erp-system/crm-service/internal/api/handlers"
	"github.com/erp-system/crm-service/internal/api/routes"
	"github.com/erp-system/crm-service/internal/business/service"
	"github.com/erp-system/crm-service/internal/config"
	"github.com/erp-system/crm-service/internal/data/kafka"
	"github.com/erp-system/crm-service/internal/data/sql"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

func main() {
	// 1. Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	utils.InitLogger("crm-service")
	responseHelper := utils.NewResponseHelper("crm-service")


	log.Printf("Starting crm-service in %s environment...", cfg.Server.Env)

	// 2. Initialize GORM Database & SQL Repositories
	db, err := sql.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	custRepo := sql.NewSQLCustomerRepository(db)
	leadRepo := sql.NewSQLLeadRepository(db)
	oppRepo := sql.NewSQLOpportunityRepository(db)
	oppStageHistoryRepo := sql.NewSQLOpportunityStageHistoryRepository(db)
	orderRepo := sql.NewSQLSalesOrderRepository(db)
	orderItemRepo := sql.NewSQLSalesOrderLineRepository(db)
	quoteRepo := sql.NewSQLQuoteRepository(db)
	quoteItemRepo := sql.NewSQLQuoteLineItemRepository(db)
	priceListRepo := sql.NewSQLPriceBookHeaderRepository(db)
	priceListItemRepo := sql.NewSQLPriceBookEntryRepository(db)
	ticketRepo := sql.NewSQLServiceTicketRepository(db)
	campaignRepo := sql.NewSQLCampaignRepository(db)
	custInteractionRepo := sql.NewSQLCustomerInteractionRepository(db)

	// 3. Initialize Kafka publisher
	kafkaPub := sharedkafka.NewPublisher(cfg.Kafka.Brokers)
	defer kafkaPub.Close()

	// 4. Initialize subdivided business services
	custSvc := service.NewCustomerService(custRepo, kafkaPub)
	oppSvc := service.NewOpportunityService(oppRepo, oppStageHistoryRepo, kafkaPub)
	leadSvc := service.NewLeadService(leadRepo, custSvc, oppSvc, kafkaPub)
	orderSvc := service.NewSalesOrderService(orderRepo, orderItemRepo, custRepo, kafkaPub)
	quoteSvc := service.NewQuoteService(quoteRepo, quoteItemRepo, kafkaPub)
	ticketSvc := service.NewServiceTicketService(ticketRepo, kafkaPub)
	campSvc := service.NewCampaignService(campaignRepo, kafkaPub)
	custInteractionSvc := service.NewCustomerInteractionService(custInteractionRepo, kafkaPub)
	plSvc := service.NewPriceListService(priceListRepo, priceListItemRepo)

	// 5. Seed initial mock data
	seedMockData(custSvc, leadSvc, oppSvc)

	// 6. Start Kafka consumer in a background thread
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kafkaSub := kafka.NewKafkaConsumer(cfg.Kafka.Brokers, cfg.Kafka.GroupID, kafkaPub, orderSvc, leadSvc, oppSvc, custInteractionSvc)
	go kafkaSub.Start(ctx)
	defer kafkaSub.Close()

	// 7. Setup Gin routing
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	r.Use(utils.TracingMiddleware("crm-service"))

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "crm-service",
			"status":  "healthy",
			"port":    cfg.Server.Port,
		})
	})

	custLeadHandler := handlers.NewCustomerLeadHandler(custSvc, leadSvc, responseHelper)
	salesOppHandler := handlers.NewSalesOpportunityHandler(oppSvc, orderSvc, quoteSvc, ticketSvc, campSvc, plSvc, responseHelper)
	custInteractionHandler := handlers.NewCustomerInteractionHandler(custInteractionSvc, responseHelper)

	routes.SetupCRMRoutes(r, custLeadHandler, salesOppHandler, custInteractionHandler)

	// 8. Start HTTP server with graceful shutdown
	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}

	go func() {
		log.Printf("CRM HTTP Server listening on port %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down crm-service...")

	// Gracefully shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("crm-service stopped gracefully.")
}

func seedMockData(custSvc *service.CustomerService, leadSvc *service.LeadService, oppSvc *service.OpportunityService) {
	ctx := context.Background()
	log.Println("Seeding CRM mock data...")

	// Seed customer
	cust, err := custSvc.CreateCustomer(ctx, "Acme Corporation", "John Doe", "john@acme.com", "+1-555-0199", "WHOLESALE", "")
	if err != nil {
		log.Printf("Failed to seed customer: %v", err)
		return
	}

	// Seed leads
	_, _ = leadSvc.CreateLead(ctx, "Alice", "Smith", "Initech", "alice@initech.com", "+1-555-0100", "WEBSITE")
	_, _ = leadSvc.CreateLead(ctx, "Bob", "Johnson", "Umbrella Corp", "bob@umbrella.com", "+1-555-0120", "CAMPAIGN")

	// Seed opportunity
	_, _ = oppSvc.CreateOpportunity(ctx, cust.ID, "Upgrade Database Infrastructure", decimal.NewFromFloat(45000.00), "NEGOTIATION")

	log.Println("CRM mock data seeded successfully.")
}
