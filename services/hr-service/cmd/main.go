package main

import (
	"erp-system/shared/utils"
	"context"
	sharedkafka "erp-system/shared/kafka"
	"log"
	"net/http"

	"github.com/erp-system/hr-service/internal/api/handlers"
	"github.com/erp-system/hr-service/internal/api/routes"
	"github.com/erp-system/hr-service/internal/business/service"
	"github.com/erp-system/hr-service/internal/config"
	"github.com/erp-system/hr-service/internal/data/kafka"
	"github.com/erp-system/hr-service/internal/data/memory"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	utils.InitLogger("hr-service")
	responseHelper := utils.NewResponseHelper("hr-service")


	// 2. Initialize Event Publisher (Kafka)
	publisher := sharedkafka.NewPublisher(cfg.Kafka.Brokers)
	defer func() {
		if err := publisher.Close(); err != nil {
			log.Printf("Failed to close Kafka publisher: %v", err)
		}
	}()

	// 3. Initialize Memory Repositories
	empRepo := memory.NewMemoryEmployeeRepo()
	deptRepo := memory.NewMemoryDepartmentRepo()
	posRepo := memory.NewMemoryPositionRepo()
	payrollRepo := memory.NewMemoryPayrollRecordRepo()
	deductionRepo := memory.NewMemoryPayrollDeductionRepo()
	timesheetRepo := memory.NewMemoryAttendanceEntryRepo()
	leaveRepo := memory.NewMemoryLeaveRequestRepo()
	leaveBalanceRepo := memory.NewMemoryLeaveBalanceRepo()
	jobPostingRepo := memory.NewMemoryJobPostingRepo()
	jobAppRepo := memory.NewMemoryJobApplicationRepo()
	perfRepo := memory.NewMemoryPerformanceReviewRepo()
	trainingRepo := memory.NewMemoryTrainingProgramRepo()
	trainingEnrollmentRepo := memory.NewMemoryTrainingEnrollmentRepo()
	docRepo := memory.NewMemoryEmployeeDocumentRepo()
	expenseClaimRepo := memory.NewMemoryExpenseClaimRepo()
	expenseClaimLineRepo := memory.NewMemoryExpenseClaimLineRepo()
	compHistoryRepo := memory.NewMemoryEmployeeCompensationHistoryRepo()
	posHistoryRepo := memory.NewMemoryPositionHistoryRepo()
	deptHistoryRepo := memory.NewMemoryDepartmentHistoryRepo()

	// 4. Initialize Services
	empSvc := service.NewEmployeeManagementService(empRepo, expenseClaimRepo, expenseClaimLineRepo, compHistoryRepo, posHistoryRepo, deptHistoryRepo, deptRepo, posRepo, publisher)
	payrollSvc := service.NewPayrollService(payrollRepo, deductionRepo, empRepo, publisher)
	timesheetSvc := service.NewTimeAttendanceService(timesheetRepo, publisher)
	leaveSvc := service.NewLeaveManagementService(leaveRepo, leaveBalanceRepo, publisher)
	recruitmentSvc := service.NewRecruitmentService(jobPostingRepo, jobAppRepo)
	perfSvc := service.NewPerformanceService(perfRepo, publisher)
	trainingSvc := service.NewTrainingService(trainingRepo, trainingEnrollmentRepo, publisher)
	docSvc := service.NewEmployeeDocumentService(docRepo)
	reportSvc := service.NewReportService(empRepo, payrollRepo, timesheetRepo)

	// 5. Initialize Handlers
	empHandler := handlers.NewEmployeeHandler(empSvc, responseHelper)
	payrollHandler := handlers.NewPayrollHandler(payrollSvc, responseHelper)
	timesheetHandler := handlers.NewTimesheetHandler(timesheetSvc, responseHelper)
	leaveHandler := handlers.NewLeaveHandler(leaveSvc, responseHelper)
	recruitmentHandler := handlers.NewRecruitmentHandler(recruitmentSvc, responseHelper)
	perfHandler := handlers.NewPerformanceHandler(perfSvc, responseHelper)
	trainingHandler := handlers.NewTrainingHandler(trainingSvc, responseHelper)
	docHandler := handlers.NewDocumentHandler(docSvc, responseHelper)
	reportHandler := handlers.NewReportHandler(reportSvc, responseHelper)

	// 5b. Initialize Event Consumer (Kafka)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer := kafka.NewKafkaConsumer(cfg.Kafka.Brokers, cfg.Kafka.GroupID, publisher, trainingSvc)
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
	routes.RegisterRoutes(
		r,
		empHandler,
		payrollHandler,
		timesheetHandler,
		leaveHandler,
		recruitmentHandler,
		perfHandler,
		trainingHandler,
		docHandler,
		reportHandler,
	)

	// 7. Start Server
	log.Printf("Starting hr-service on port %s in %s mode", cfg.Server.Port, cfg.Server.Env)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
