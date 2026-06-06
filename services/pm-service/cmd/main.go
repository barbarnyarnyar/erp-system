package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/erp-system/pm-service/internal/api/handlers"
	"github.com/erp-system/pm-service/internal/api/routes"
	"github.com/erp-system/pm-service/internal/business/service"
	"github.com/erp-system/pm-service/internal/config"
	"github.com/erp-system/pm-service/internal/data/kafka"
	"github.com/erp-system/pm-service/internal/data/memory"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

func main() {
	// 1. Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Starting pm-service in %s environment on port %s...", cfg.Server.Env, cfg.Server.Port)

	// 2. Initialize in-memory repositories
	portfolioRepo := memory.NewPortfolioRepository()
	projectRepo := memory.NewProjectRepository()
	taskRepo := memory.NewTaskRepository()
	depRepo := memory.NewTaskDependencyRepository()
	allocRepo := memory.NewResourceAllocationRepository()
	timeRepo := memory.NewTimeEntryRepository()
	expenseRepo := memory.NewProjectExpenseRepository()
	docRepo := memory.NewProjectDocumentRepository()
	issueRepo := memory.NewProjectIssueRepository()
	changeRepo := memory.NewChangeRequestRepository()

	// 3. Initialize Kafka publisher
	kafkaPub := kafka.NewKafkaPublisher(cfg.Kafka.Brokers)
	defer kafkaPub.Close()

	// 4. Initialize subdivided services (matching CDD components)
	planningSvc := service.NewProjectPlanningService(portfolioRepo, projectRepo, kafkaPub)
	taskSvc := service.NewTaskManagementService(taskRepo, depRepo, kafkaPub)
	resourceSvc := service.NewResourceManagementService(allocRepo, kafkaPub)
	timeSvc := service.NewTimeExpenseService(timeRepo, expenseRepo, kafkaPub)
	collabSvc := service.NewCollaborationService(docRepo, issueRepo, changeRepo, kafkaPub)
	analyticsSvc := service.NewPortfolioAnalyticsService(projectRepo, taskRepo, timeRepo, expenseRepo)

	// 5. Seed initial mock data
	seedMockData(planningSvc, taskSvc, resourceSvc, timeSvc, collabSvc)

	// 6. Start Kafka consumer in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kafkaSub := kafka.NewKafkaConsumer(cfg.Kafka.Brokers, cfg.Kafka.GroupID, planningSvc, taskSvc)
	go kafkaSub.Start(ctx)
	defer kafkaSub.Close()

	// 7. Setup Gin routing
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "pm-service",
			"status":  "healthy",
			"port":    cfg.Server.Port,
		})
	})

	// Hello World endpoint
	r.GET("/api/v1/projects/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello World from Project Management Service!",
			"service": "pm-service",
			"port":    cfg.Server.Port,
		})
	})

	// Root endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Project Management Service is running",
			"service": "pm-service",
			"port":    cfg.Server.Port,
		})
	})

	handler := handlers.NewProjectHandler(planningSvc, taskSvc, resourceSvc, timeSvc, collabSvc, analyticsSvc)
	routes.SetupPMRoutes(r, handler)

	// 8. Start HTTP server with graceful shutdown
	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}

	go func() {
		log.Printf("PM HTTP Server listening on port %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down pm-service...")

	// Gracefully shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("pm-service stopped gracefully.")
}

func seedMockData(
	planningSvc *service.ProjectPlanningService,
	taskSvc *service.TaskManagementService,
	resourceSvc *service.ResourceManagementService,
	timeSvc *service.TimeExpenseService,
	collabSvc *service.CollaborationService,
) {
	ctx := context.Background()
	log.Println("Seeding Project Management mock data...")

	// Seed portfolio
	portfolio, err := planningSvc.CreatePortfolio(ctx, "Enterprise Digital Transformation", "Strategic portfolio for digitizing company operations", "emp_manager1")
	if err != nil {
		log.Printf("Failed to seed portfolio: %v", err)
		return
	}

	// Seed project
	startDate := time.Now().AddDate(0, -1, 0) // Started 1 month ago
	endDate := startDate.AddDate(0, 6, 0)     // Ends in 5 months
	proj, err := planningSvc.CreateProject(ctx, "Warehouse Automation System", "Design and roll out new automated warehouse storage and picking workflows", startDate, &endDate, portfolio.ID)
	if err != nil {
		log.Printf("Failed to seed project: %v", err)
		return
	}

	// Seed resource allocation
	_, _ = resourceSvc.AllocateResource(ctx, proj.ID, "emp_lead_eng", "Lead Software Engineer", 100, startDate, &endDate)
	_, _ = resourceSvc.AllocateResource(ctx, proj.ID, "emp_pm", "Project Manager", 50, startDate, &endDate)

	// Seed tasks (WBS)
	t1, _ := taskSvc.CreateTask(ctx, proj.ID, "", "Requirement Gathering & Architecture Design", "Collect warehouse hardware specs and sketch overall software architecture", "", &startDate, nil)
	t2, _ := taskSvc.CreateTask(ctx, proj.ID, t1.ID, "Database Schema definition", "Draft PostgreSQL schema tables for storage levels and layout", "", &startDate, nil)

	// Update task progress
	_, _ = taskSvc.UpdateTaskProgress(ctx, t1.ID, 100, decimal.NewFromFloat(40.0), "DONE")
	_, _ = taskSvc.UpdateTaskProgress(ctx, t2.ID, 60, decimal.NewFromFloat(12.5), "IN_PROGRESS")

	// Seed task dependency
	_, _ = taskSvc.AddTaskDependency(ctx, t2.ID, t1.ID, "FS")

	// Seed logged time
	timeEntry, err := timeSvc.LogTime(ctx, proj.ID, t1.ID, "emp_lead_eng", time.Now().AddDate(0, 0, -5), decimal.NewFromFloat(8.0), "Designed system core architecture")
	if err == nil && timeEntry != nil {
		_, _ = timeSvc.ApproveTime(ctx, timeEntry.ID, "emp_pm")
	}

	// Seed logged expense
	expEntry, err := timeSvc.LogExpense(ctx, proj.ID, t1.ID, "emp_pm", decimal.NewFromFloat(150.00), "USD", time.Now().AddDate(0, 0, -10), "Software License", "Architecture sketching tool license")
	if err == nil && expEntry != nil {
		_, _ = timeSvc.ApproveExpense(ctx, expEntry.ID, "emp_pm")
	}

	// Seed document
	_, _ = collabSvc.UploadDocument(ctx, proj.ID, "System Architecture Draft.pdf", "/uploads/docs/sys_arch_draft.pdf", 2048576, "emp_lead_eng")

	// Seed issue
	issue, _ := collabSvc.LogIssue(ctx, proj.ID, "Delay in controller hardware shipment", "SCM vendor reports 2-week delay in sensor delivery", "HIGH", "emp_pm")
	_, _ = collabSvc.ResolveIssue(ctx, issue.ID, "emp_pm")

	// Seed change request
	_, _ = collabSvc.CreateChangeRequest(ctx, proj.ID, "Support Multi-Language UI", "Required by regional warehouse teams in Europe", "User requests internationalization", "UI team 3 days extra effort", "emp_pm")

	log.Println("Project Management mock data seeded successfully.")
}