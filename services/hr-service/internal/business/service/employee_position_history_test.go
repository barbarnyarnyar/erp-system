package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/erp-system/hr-service/internal/business/service"
	"github.com/erp-system/hr-service/internal/data/memory"
	"github.com/shopspring/decimal"
)

type silentPub struct{}

func (silentPub) Publish(ctx context.Context, topic string, key string, payload interface{}) error { return nil }

func TestEmployee_Update_RecordsPositionAndDepartmentHistory(t *testing.T) {
	empRepo := memory.NewMemoryEmployeeRepo()
	posRepo := memory.NewMemoryPositionRepo()
	deptRepo := memory.NewMemoryDepartmentRepo()
	posHistRepo := memory.NewMemoryPositionHistoryRepo()
	deptHistRepo := memory.NewMemoryDepartmentHistoryRepo()
	compHistRepo := memory.NewMemoryEmployeeCompensationHistoryRepo()
	expClaimRepo := memory.NewMemoryExpenseClaimRepo()
	expClaimLineRepo := memory.NewMemoryExpenseClaimLineRepo()

	svc := service.NewEmployeeManagementService(empRepo, expClaimRepo, expClaimLineRepo, compHistRepo, posHistRepo, deptHistRepo, deptRepo, posRepo, silentPub{})

	ctx := context.Background()
	emp, err := svc.CreateEmployee(ctx, "Alice", "Smith", "alice@x.com", "dept_1", "pos_1", decimal.NewFromInt(5000))
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	_, err = svc.UpdateEmployee(ctx, emp.ID, "Alice", "Smith", "alice@x.com", "dept_2", "pos_2", "ACTIVE")
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	posHist, _ := svc.ListPositionHistory(ctx, emp.ID)
	if len(posHist) != 1 {
		t.Fatalf("expected 1 position history entry, got %d", len(posHist))
	}
	if posHist[0].PositionID != "pos_2" {
		t.Errorf("position history PositionID = %q, want pos_2", posHist[0].PositionID)
	}

	deptHist, _ := svc.ListDepartmentHistory(ctx, emp.ID)
	if len(deptHist) != 1 {
		t.Fatalf("expected 1 department history entry, got %d", len(deptHist))
	}
	if deptHist[0].DepartmentID != "dept_2" {
		t.Errorf("department history DepartmentID = %q, want dept_2", deptHist[0].DepartmentID)
	}
}

func TestEmployee_Update_SamePositionNoNewHistory(t *testing.T) {
	empRepo := memory.NewMemoryEmployeeRepo()
	posRepo := memory.NewMemoryPositionRepo()
	deptRepo := memory.NewMemoryDepartmentRepo()
	posHistRepo := memory.NewMemoryPositionHistoryRepo()
	deptHistRepo := memory.NewMemoryDepartmentHistoryRepo()
	compHistRepo := memory.NewMemoryEmployeeCompensationHistoryRepo()
	expClaimRepo := memory.NewMemoryExpenseClaimRepo()
	expClaimLineRepo := memory.NewMemoryExpenseClaimLineRepo()

	svc := service.NewEmployeeManagementService(empRepo, expClaimRepo, expClaimLineRepo, compHistRepo, posHistRepo, deptHistRepo, deptRepo, posRepo, silentPub{})

	ctx := context.Background()
	emp, _ := svc.CreateEmployee(ctx, "Bob", "Jones", "bob@x.com", "dept_1", "pos_1", decimal.NewFromInt(5000))

	_, _ = svc.UpdateEmployee(ctx, emp.ID, "Bob", "Jones", "bob@x.com", "dept_1", "pos_1", "ACTIVE")

	posHist, _ := svc.ListPositionHistory(ctx, emp.ID)
	if len(posHist) != 0 {
		t.Errorf("expected 0 position history entries (no change), got %d", len(posHist))
	}
	deptHist, _ := svc.ListDepartmentHistory(ctx, emp.ID)
	if len(deptHist) != 0 {
		t.Errorf("expected 0 department history entries (no change), got %d", len(deptHist))
	}
}

func TestEmployee_MultipleUpdates_MultipleHistoryEntries(t *testing.T) {
	empRepo := memory.NewMemoryEmployeeRepo()
	posRepo := memory.NewMemoryPositionRepo()
	deptRepo := memory.NewMemoryDepartmentRepo()
	posHistRepo := memory.NewMemoryPositionHistoryRepo()
	deptHistRepo := memory.NewMemoryDepartmentHistoryRepo()
	compHistRepo := memory.NewMemoryEmployeeCompensationHistoryRepo()
	expClaimRepo := memory.NewMemoryExpenseClaimRepo()
	expClaimLineRepo := memory.NewMemoryExpenseClaimLineRepo()

	svc := service.NewEmployeeManagementService(empRepo, expClaimRepo, expClaimLineRepo, compHistRepo, posHistRepo, deptHistRepo, deptRepo, posRepo, silentPub{})

	ctx := context.Background()
	emp, _ := svc.CreateEmployee(ctx, "Carol", "Lee", "carol@x.com", "dept_1", "pos_1", decimal.NewFromInt(5000))

	time.Sleep(time.Millisecond)
	_, _ = svc.UpdateEmployee(ctx, emp.ID, "Carol", "Lee", "carol@x.com", "dept_2", "pos_1", "ACTIVE")
	time.Sleep(time.Millisecond)
	_, _ = svc.UpdateEmployee(ctx, emp.ID, "Carol", "Lee", "carol@x.com", "dept_2", "pos_2", "ACTIVE")

	posHist, _ := svc.ListPositionHistory(ctx, emp.ID)
	if len(posHist) != 1 {
		t.Errorf("expected 1 position history (1 change), got %d", len(posHist))
	}
	deptHist, _ := svc.ListDepartmentHistory(ctx, emp.ID)
	if len(deptHist) != 1 {
		t.Errorf("expected 1 department history (1 change), got %d", len(deptHist))
	}
}
