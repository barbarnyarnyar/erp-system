package service

import (
	"context"
	"testing"
	"time"

	"github.com/erp-system/hr-service/internal/business/domain"
	"github.com/erp-system/hr-service/internal/data/memory"
	"github.com/shopspring/decimal"
)

type MockAuditPublisher struct {
	Events []MockEvent
}

type MockEvent struct {
	Topic   string
	Key     string
	Payload interface{}
}

func (m *MockAuditPublisher) Publish(ctx context.Context, topic string, key string, payload interface{}) error {
	m.Events = append(m.Events, MockEvent{
		Topic:   topic,
		Key:     key,
		Payload: payload,
	})
	return nil
}

func TestEmployeeManagementService_AuditHistory(t *testing.T) {
	empRepo := memory.NewMemoryEmployeeRepo()
	claimsRepo := memory.NewMemoryExpenseClaimRepo()
	claimLinesRepo := memory.NewMemoryExpenseClaimLineRepo()
	compHistoryRepo := memory.NewMemoryEmployeeCompensationHistoryRepo()
	posHistoryRepo := memory.NewMemoryPositionHistoryRepo()
	deptHistoryRepo := memory.NewMemoryDepartmentHistoryRepo()
	deptRepo := memory.NewMemoryDepartmentRepo()
	posRepo := memory.NewMemoryPositionRepo()
	publisher := &MockAuditPublisher{}

	svc := NewEmployeeManagementService(
		empRepo,
		claimsRepo,
		claimLinesRepo,
		compHistoryRepo,
		posHistoryRepo,
		deptHistoryRepo,
		deptRepo,
		posRepo,
		publisher,
	)

	ctx := context.Background()

	// 1. Create an employee
	emp, err := svc.CreateEmployee(ctx, "John", "Doe", "john.doe@company.com", "dept_1", "pos_1", decimal.NewFromInt(5000))
	if err != nil {
		t.Fatalf("unexpected error creating employee: %v", err)
	}

	// Verify initial compensation history
	compHist, _ := svc.ListCompensationHistory(ctx, emp.ID)
	if len(compHist) != 1 {
		t.Fatalf("expected 1 compensation history record, got %d", len(compHist))
	}
	if !compHist[0].Salary.Equal(decimal.NewFromInt(5000)) {
		t.Errorf("expected salary 5000, got %s", compHist[0].Salary)
	}

	// 2. Update employee: change position and department, and try to bypass salary directly (direct salary update should be blocked)
	// UpdateEmployee signature: UpdateEmployee(ctx, id, firstName, lastName, email, deptID, posID, status)
	updatedEmp, err := svc.UpdateEmployee(ctx, emp.ID, "John", "Doe", "john.doe@company.com", "dept_2", "pos_2", "ACTIVE")
	if err != nil {
		t.Fatalf("unexpected error updating employee: %v", err)
	}

	// Verify salary is still 5000 (direct bypass blocked)
	if !updatedEmp.Salary.Equal(decimal.NewFromInt(5000)) {
		t.Errorf("expected salary to remain 5000, got %s", updatedEmp.Salary)
	}

	// Verify PositionHistory is recorded
	posHist, _ := svc.ListPositionHistory(ctx, emp.ID)
	if len(posHist) != 1 {
		t.Fatalf("expected 1 position history record, got %d", len(posHist))
	}
	if posHist[0].PositionID != "pos_2" {
		t.Errorf("expected position ID pos_2, got %s", posHist[0].PositionID)
	}

	// Verify DepartmentHistory is recorded
	deptHist, _ := svc.ListDepartmentHistory(ctx, emp.ID)
	if len(deptHist) != 1 {
		t.Fatalf("expected 1 department history record, got %d", len(deptHist))
	}
	if deptHist[0].DepartmentID != "dept_2" {
		t.Errorf("expected department ID dept_2, got %s", deptHist[0].DepartmentID)
	}

	// Verify TopicHrEmployeeTransferred event was published
	transferredEventFound := false
	for _, ev := range publisher.Events {
		if ev.Topic == domain.TopicHrEmployeeTransferred {
			transferredEventFound = true
			payload := ev.Payload.(domain.EmployeeTransferredEvent)
			if payload.OldDepartmentID != "dept_1" || payload.NewDepartmentID != "dept_2" {
				t.Errorf("unexpected old/new department: %s/%s", payload.OldDepartmentID, payload.NewDepartmentID)
			}
		}
	}
	if !transferredEventFound {
		t.Errorf("expected TopicHrEmployeeTransferred event to be published")
	}

	// 3. Update salary through dedicated UpdateCompensation
	updatedEmp, err = svc.UpdateCompensation(ctx, emp.ID, decimal.NewFromInt(6000), time.Now(), "manager_admin")
	if err != nil {
		t.Fatalf("unexpected error updating compensation: %v", err)
	}

	if !updatedEmp.Salary.Equal(decimal.NewFromInt(6000)) {
		t.Errorf("expected updated salary 6000, got %s", updatedEmp.Salary)
	}

	// Verify updated compensation history
	compHist, _ = svc.ListCompensationHistory(ctx, emp.ID)
	if len(compHist) != 2 {
		t.Fatalf("expected 2 compensation history records, got %d", len(compHist))
	}
	if !compHist[1].Salary.Equal(decimal.NewFromInt(6000)) || compHist[1].ChangedBy != "manager_admin" {
		t.Errorf("unexpected compensation history values: %+v", compHist[1])
	}
}
