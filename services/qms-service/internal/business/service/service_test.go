package service_test

import (
	"context"
	"testing"

	"github.com/erp-system/qms-service/internal/business/domain"
	"github.com/erp-system/qms-service/internal/business/service"
	"github.com/erp-system/qms-service/internal/data/memory"
	"github.com/shopspring/decimal"
)

type MockPublisher struct{}

func (m *MockPublisher) Publish(ctx context.Context, topic string, key string, payload interface{}) error {
	return nil
}

func TestQmsService(t *testing.T) {
	planRepo := memory.NewMemoryInspectionPlanRepo()
	metricRepo := memory.NewMemoryInspectionMetricDefinitionRepo()
	qiRepo := memory.NewMemoryQualityInspectionRepo()
	resRepo := memory.NewMemoryInspectionResultLineRepo()
	ncRepo := memory.NewMemoryNonConformanceLogRepo()
	publisher := &MockPublisher{}

	planSvc := service.NewInspectionPlanService(planRepo, metricRepo)
	ncSvc := service.NewNonConformanceService(ncRepo, planRepo, qiRepo, publisher)
	execSvc := service.NewInspectionExecutionService(qiRepo, resRepo, planRepo, ncSvc, publisher)
	analySvc := service.NewQualityAnalyticsService(resRepo)

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
	analytics, err := analySvc.ComputeSpcDistribution(ctx, plan.ID, metric.ID, domain.TimeRange{})
	if err != nil {
		t.Fatalf("failed to compute analytics: %v", err)
	}
	if analytics.PlanID != plan.ID {
		t.Errorf("expected plan %s, got %s", plan.ID, analytics.PlanID)
	}
}
