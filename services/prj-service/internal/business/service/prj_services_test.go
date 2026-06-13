package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/erp-system/pm-service/internal/business/domain"
	"github.com/erp-system/pm-service/internal/business/service"
	"github.com/erp-system/pm-service/internal/data/sql"
	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) (*gorm.DB, domain.ProjectRepository, domain.WbsNodeRepository, domain.TimeLogRepository, domain.TransactionalOutboxRepository) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		t.Fatalf("Failed to open sqlite database: %v", err)
	}

	err = db.AutoMigrate(
		&sql.Project{},
		&sql.WbsNode{},
		&sql.TimeLog{},
		&sql.TransactionalOutbox{},
		&sql.KafkaEventInbox{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	projRepo := sql.NewSQLProjectRepository(db)
	wbsRepo := sql.NewSQLWbsNodeRepository(db)
	timeRepo := sql.NewSQLTimeLogRepository(db)
	outboxRepo := sql.NewSQLTransactionalOutboxRepository(db)

	return db, projRepo, wbsRepo, timeRepo, outboxRepo
}

func TestProjectTrackingService_InitializeAndTransition(t *testing.T) {
	db, projRepo, _, _, _ := setupTestDB(t)

	svc := service.NewProjectTrackingService(db, projRepo)

	proj, err := svc.InitializeProject(
		context.Background(),
		"tenant-1",
		"cust-1",
		"PRJ-001",
		"Implementation Project",
		domain.BillingMethodTIME_AND_MATERIALS,
		time.Now(),
	)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if proj.Status != domain.ProjectStatusDRAFT {
		t.Errorf("Expected DRAFT, got: %s", proj.Status)
	}

	proj, err = svc.TransitionProjectStatus(context.Background(), proj.ID, domain.ProjectStatusACTIVE)
	if err != nil {
		t.Fatalf("Expected no error transitioning status, got: %v", err)
	}

	if proj.Status != domain.ProjectStatusACTIVE {
		t.Errorf("Expected ACTIVE status, got: %s", proj.Status)
	}
}

func TestWbsStructureService_AppendAndCompleteMilestone(t *testing.T) {
	db, projRepo, wbsRepo, _, outboxRepo := setupTestDB(t)

	projSvc := service.NewProjectTrackingService(db, projRepo)
	wbsSvc := service.NewWbsStructureService(db, projRepo, wbsRepo, outboxRepo)

	proj, _ := projSvc.InitializeProject(
		context.Background(),
		"tenant-1",
		"cust-1",
		"PRJ-001",
		"Test Project",
		domain.BillingMethodFIXED_PRICE,
		time.Now(),
	)

	// Append a Milestone WBS Node
	revAmt := decimal.NewFromFloat(5000)
	node, err := wbsSvc.AppendWbsNode(
		context.Background(),
		proj.ID,
		nil,
		"1.0",
		"Kick-off milestone",
		domain.WbsNodeTypeMILESTONE,
		decimal.NewFromFloat(0),
	)
	if err != nil {
		t.Fatalf("Failed to append WBS node: %v", err)
	}

	// Update node to assign functional budget revenue
	node.BudgetRevenueFunctional = &revAmt
	_ = wbsRepo.Update(context.Background(), node)

	// Declare Milestone completed
	node, err = wbsSvc.DeclareNodeCompletion(context.Background(), node.ID, "emp-pm-1")
	if err != nil {
		t.Fatalf("Failed to complete milestone node: %v", err)
	}

	if !node.IsCompleted {
		t.Errorf("Expected node to be completed")
	}

	// Verify transactional outbox message was written
	unsent, err := outboxRepo.GetUnsent(context.Background(), 10)
	if err != nil {
		t.Fatalf("Failed to read outbox: %v", err)
	}
	if len(unsent) != 1 {
		t.Fatalf("Expected 1 outbox message, got: %d", len(unsent))
	}
	if unsent[0].EventType != domain.TopicPrjMilestoneAchieved {
		t.Errorf("Expected event type prj.milestone.achieved, got: %s", unsent[0].EventType)
	}
}

func TestTimeTrackingService_LogAndApproveTimeLogs(t *testing.T) {
	db, projRepo, wbsRepo, timeRepo, outboxRepo := setupTestDB(t)

	projSvc := service.NewProjectTrackingService(db, projRepo)
	wbsSvc := service.NewWbsStructureService(db, projRepo, wbsRepo, outboxRepo)
	timeSvc := service.NewTimeTrackingService(db, projRepo, wbsRepo, timeRepo, outboxRepo)

	proj, _ := projSvc.InitializeProject(
		context.Background(),
		"tenant-1",
		"cust-1",
		"PRJ-001",
		"T&M Project",
		domain.BillingMethodTIME_AND_MATERIALS,
		time.Now(),
	)

	node, _ := wbsSvc.AppendWbsNode(
		context.Background(),
		proj.ID,
		nil,
		"1.1",
		"Development task",
		domain.WbsNodeTypeTASK,
		decimal.NewFromFloat(100),
	)

	logs := []domain.TimeLogSubmissionInput{
		{
			WbsNodeID:        node.ID,
			WorkDate:         time.Now(),
			HoursSpent:       decimal.NewFromFloat(8),
			InternalCostRate: decimal.NewFromFloat(50),
			BillingRate:      decimal.NewFromFloat(150),
			IsBillable:       true,
		},
	}

	err := timeSvc.LogOperationalHoursBulk(context.Background(), "tenant-1", "emp-dev-1", logs)
	if err != nil {
		t.Fatalf("Expected no error logging hours, got: %v", err)
	}

	// Fetch all logs to get the ID
	allLogs, err := timeRepo.List(context.Background())
	if err != nil {
		t.Fatalf("Failed to list logs: %v", err)
	}
	if len(allLogs) != 1 {
		t.Fatalf("Expected 1 log, got %d", len(allLogs))
	}

	logID := allLogs[0].ID

	// Approve logs
	err = timeSvc.ProcessTimesheetApproval(context.Background(), []string{logID}, "emp-pm-1")
	if err != nil {
		t.Fatalf("Failed to process approval: %v", err)
	}

	// Verify approval status in DB
	approvedLog, err := timeRepo.GetByID(context.Background(), logID)
	if err != nil {
		t.Fatalf("Failed to fetch log: %v", err)
	}
	if !approvedLog.IsApproved {
		t.Errorf("Expected log to be approved")
	}

	// Verify transactional outbox message was written for prj.time.logged
	unsent, err := outboxRepo.GetUnsent(context.Background(), 10)
	if err != nil {
		t.Fatalf("Failed to read outbox: %v", err)
	}
	if len(unsent) != 1 {
		t.Fatalf("Expected 1 outbox message, got: %d", len(unsent))
	}
	if unsent[0].EventType != domain.TopicPrjTimeLogged {
		t.Errorf("Expected event type prj.time.logged, got: %s", unsent[0].EventType)
	}
}
