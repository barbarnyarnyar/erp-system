package main

import (
	"context"
	"log"
	"net/http"

	"github.com/erp-system/hr-service/internal/api/handlers"
	"github.com/erp-system/hr-service/internal/api/routes"
	"github.com/erp-system/hr-service/internal/business/service"
	"github.com/erp-system/hr-service/internal/config"
	"github.com/erp-system/hr-service/internal/data/kafka"
	"github.com/erp-system/hr-service/internal/data/sql"
	sharedkafka "erp-system/shared/kafka"
	"erp-system/shared/utils"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	utils.InitLogger("hr-service")

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
		log.Fatalf("Failed to initialize GORM database: %v", err)
	}

	empRepo := sql.NewSQLEmployeeMasterRepository(db)
	deptRepo := sql.NewSQLDepartmentRepository(db)
	payrollRepo := sql.NewSQLPayrollRunRepository(db)
	expenseClaimRepo := sql.NewSQLExpenseClaimRepository(db)
	expenseClaimLineRepo := sql.NewSQLExpenseClaimLineRepository(db)
	outboxRepo := sql.NewSQLTransactionalOutboxRepository(db)
	inboxRepo := sql.NewSQLKafkaEventInboxRepository(db)

	// 4. Initialize Services
	empSvc := service.NewEmployeeService(db, empRepo, deptRepo, outboxRepo)
	payrollSvc := service.NewPayrollService(db, payrollRepo, empRepo, outboxRepo)
	expenseSvc := service.NewExpenseService(db, expenseClaimRepo, expenseClaimLineRepo, empRepo, outboxRepo)
	reliableSvc := service.NewReliableMessagingService(db, inboxRepo)

	// 5. Initialize Handlers
	hrHandler := handlers.NewHrHandler(
		empSvc,
		payrollSvc,
		expenseSvc,
		deptRepo,
		empRepo,
		payrollRepo,
		expenseClaimRepo,
		expenseClaimLineRepo,
	)

	// 5b. Initialize Event Consumer (Kafka)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer := kafka.NewKafkaConsumer(cfg.Kafka.Brokers, cfg.Kafka.GroupID, publisher, expenseSvc, reliableSvc)
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
			"service": "hr-service",
			"status":  "healthy",
			"port":    cfg.Server.Port,
		})
	})

	// Register API Routes
	routes.RegisterRoutes(r, hrHandler)

	// 7. Start Server
	log.Printf("Starting hr-service on port %s in %s mode", cfg.Server.Port, cfg.Server.Env)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
