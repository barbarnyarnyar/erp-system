package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/erp-system/qms-service/internal/business/domain"
	"github.com/erp-system/qms-service/internal/business/service"
	"github.com/erp-system/qms-service/internal/data/sql"
	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestQmsService(t *testing.T) {
	// Initialize an in-memory SQLite database for test stability and isolation
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite database: %v", err)
	}

	// AutoMigrate all domain models matching QMS
	err = db.AutoMigrate(
		&sql.InspectionPlan{},
		&sql.InspectionMetricDefinition{},
		&sql.QualityInspection{},
		&sql.InspectionResultLine{},
		&sql.NonConformanceLog{},
		&sql.TransactionalOutbox{},
		&sql.KafkaEventInbox{},
	)
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Instantiate repositories wrapping the GORM connection
	planRepo := sql.NewSQLInspectionPlanRepository(db)
	metricRepo := sql.NewSQLInspectionMetricDefinitionRepository(db)
	qiRepo := sql.NewSQLQualityInspectionRepository(db)
	resRepo := sql.NewSQLInspectionResultLineRepository(db)
	ncRepo := sql.NewSQLNonConformanceLogRepository(db)
	inboxRepo := sql.NewSQLKafkaEventInboxRepository(db)
	outboxRepo := sql.NewSQLTransactionalOutboxRepository(db)

	reliableSvc := service.NewReliableMessagingService(db, inboxRepo, outboxRepo)
	planSvc := service.NewInspectionPlanService(db, planRepo, metricRepo)
	ncSvc := service.NewNonConformanceService(db, ncRepo, planRepo, qiRepo, reliableSvc)
	execSvc := service.NewInspectionExecutionService(db, qiRepo, resRepo, planRepo, ncSvc, reliableSvc)
	analySvc := service.NewQualityAnalyticsService(db, resRepo)

	ctx := context.Background()

	// 1. Test ConfigurePlan
	plan, err := planSvc.ConfigurePlan(ctx, "tenant-1", "mat-100", "SCM Inbound Check")
	if err != nil {
		t.Fatalf("failed to configure plan: %v", err)
	}
	if plan.PlanName != "SCM Inbound Check" {
		t.Errorf("expected SCM Inbound Check, got %s", plan.PlanName)
	}

	// 2. Test RegisterPlanMetric
	minVal := decimal.NewFromFloat(5.0)
	maxVal := decimal.NewFromFloat(15.0)
	metric, err := planSvc.RegisterPlanMetric(ctx, plan.ID, "length_mm", "Material Length", domain.MetricDataTypeNUMERIC, &minVal, &maxVal)
	if err != nil {
		t.Fatalf("failed to register metric: %v", err)
	}
	if metric.MetricKey != "length_mm" {
		t.Errorf("expected length_mm, got %s", metric.MetricKey)
	}

	// 3. Test StageInspection
	qi, err := execSvc.StageInspection(ctx, "tenant-1", plan.ID, domain.InspectionTriggerTypeINBOUND_RECEIPT, "scm-po-101")
	if err != nil {
		t.Fatalf("failed to stage inspection: %v", err)
	}
	if qi.Status != domain.InspectionStatusPENDING {
		t.Errorf("expected PENDING status, got %v", qi.Status)
	}

	// 4. Test AssignInspector
	qi, err = execSvc.AssignInspector(ctx, qi.ID, "emp-303")
	if err != nil {
		t.Fatalf("failed to assign inspector: %v", err)
	}
	if qi.Status != domain.InspectionStatusIN_PROGRESS {
		t.Errorf("expected IN_PROGRESS, got %v", qi.Status)
	}

	// 5. Test RecordBulkMeasurements
	samples := []domain.MetricSubmissionInput{
		{
			MetricDefinitionID: metric.ID,
			SampleSequence:     1,
			NumericValue:       &minVal, // Compliant
			IsCompliant:        true,
		},
	}
	err = execSvc.RecordBulkMeasurements(ctx, qi.ID, samples)
	if err != nil {
		t.Fatalf("failed to record measurements: %v", err)
	}

	// 6. Test LogFailureIncident
	nc, err := ncSvc.LogFailureIncident(ctx, "tenant-1", qi.ID, "Failed length tolerance check", decimal.NewFromInt(1), true)
	if err != nil {
		t.Fatalf("failed to log failure incident: %v", err)
	}
	if !nc.IsQuarantined {
		t.Errorf("expected non-conformance item to be quarantined")
	}

	// 7. Test ExecuteDisposition
	nc, err = ncSvc.ExecuteDisposition(ctx, nc.ID, domain.DispositionActionREWORK, "Send for reprocessing", "emp-404")
	if err != nil {
		t.Fatalf("failed to execute disposition: %v", err)
	}
	if nc.IsQuarantined {
		t.Errorf("expected quarantine release on disposition")
	}

	// 8. Test ComputeSpcDistribution
	// Create a window that includes the current timestamp
	tr := domain.TimeRange{
		StartDate: time.Now().Add(-24 * time.Hour),
		EndDate:   time.Now().Add(24 * time.Hour),
	}
	analytics, err := analySvc.ComputeSpcDistribution(ctx, plan.ID, metric.ID, tr)
	if err != nil {
		t.Fatalf("failed to compute analytics: %v", err)
	}
	if analytics.PlanID != plan.ID {
		t.Errorf("expected plan %s, got %s", plan.ID, analytics.PlanID)
	}
	if analytics.SampleSize != 1 {
		t.Errorf("expected sample size 1, got %d", analytics.SampleSize)
	}
}
