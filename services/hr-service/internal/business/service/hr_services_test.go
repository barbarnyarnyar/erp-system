package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/erp-system/hr-service/internal/business/domain"
	"github.com/erp-system/hr-service/internal/business/service"
	"github.com/erp-system/hr-service/internal/data/sql"
	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) (*gorm.DB, domain.EmployeeMasterRepository, domain.DepartmentRepository, domain.PayrollRunRepository, domain.ExpenseClaimRepository, domain.ExpenseClaimLineRepository, domain.TransactionalOutboxRepository) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		t.Fatalf("Failed to open sqlite database: %v", err)
	}

	err = db.AutoMigrate(
		&sql.Department{},
		&sql.EmployeeMaster{},
		&sql.PayrollRun{},
		&sql.ExpenseClaim{},
		&sql.ExpenseClaimLine{},
		&sql.TransactionalOutbox{},
		&sql.KafkaEventInbox{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	empRepo := sql.NewSQLEmployeeMasterRepository(db)
	deptRepo := sql.NewSQLDepartmentRepository(db)
	payrollRepo := sql.NewSQLPayrollRunRepository(db)
	expenseClaimRepo := sql.NewSQLExpenseClaimRepository(db)
	expenseClaimLineRepo := sql.NewSQLExpenseClaimLineRepository(db)
	outboxRepo := sql.NewSQLTransactionalOutboxRepository(db)

	return db, empRepo, deptRepo, payrollRepo, expenseClaimRepo, expenseClaimLineRepo, outboxRepo
}

func TestEmployeeService_HireEmployee(t *testing.T) {
	db, empRepo, deptRepo, _, _, _, outboxRepo := setupTestDB(t)

	dept := &domain.Department{
		ID:             "dept-1",
		LegalEntityID:  "tenant-1",
		DepartmentCode: "DEPT_ENG",
		Name:           "Engineering",
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := deptRepo.Create(context.Background(), dept); err != nil {
		t.Fatalf("Failed to create department: %v", err)
	}

	empSvc := service.NewEmployeeService(db, empRepo, deptRepo, outboxRepo)

	salary := decimal.NewFromFloat(5000)
	emp, err := empSvc.HireEmployee(
		context.Background(),
		"tenant-1",
		"dept-1",
		nil,
		"EMP001",
		"John",
		"Doe",
		"john.doe@example.com",
		salary,
		domain.EmploymentTypeFULL_TIME,
	)

	if err != nil {
		t.Fatalf("Expected no error hiring employee, got: %v", err)
	}

	if emp.EmployeeNumber != "EMP001" {
		t.Errorf("Expected employee number EMP001, got: %s", emp.EmployeeNumber)
	}

	unsent, err := outboxRepo.GetUnsent(context.Background(), 10)
	if err != nil {
		t.Fatalf("Failed to get unsent messages: %v", err)
	}
	if len(unsent) != 1 {
		t.Errorf("Expected 1 outbox message, got: %d", len(unsent))
	}
}

func TestPayrollService_InitiateAndExecute(t *testing.T) {
	db, empRepo, deptRepo, payrollRepo, _, _, outboxRepo := setupTestDB(t)

	dept := &domain.Department{
		ID:             "dept-1",
		LegalEntityID:  "tenant-1",
		DepartmentCode: "DEPT_ENG",
		Name:           "Engineering",
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_ = deptRepo.Create(context.Background(), dept)

	emp := &domain.EmployeeMaster{
		ID:             "emp-1",
		LegalEntityID:  "tenant-1",
		DepartmentID:   "dept-1",
		EmployeeNumber: "EMP001",
		FirstName:      "John",
		LastName:       "Doe",
		Email:          "john@example.com",
		Status:         domain.EmployeeStatusACTIVE,
		Type:           domain.EmploymentTypeFULL_TIME,
		BaseSalary:     decimal.NewFromFloat(6000),
		Version:        1,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_ = empRepo.Create(context.Background(), emp)

	payrollSvc := service.NewPayrollService(db, payrollRepo, empRepo, outboxRepo)

	run, err := payrollSvc.InitiatePeriodRun(context.Background(), "tenant-1", 2026, 6)
	if err != nil {
		t.Fatalf("Failed to initiate payroll run: %v", err)
	}

	if run.Status != domain.PayrollStatusDRAFT {
		t.Errorf("Expected DRAFT status, got %s", run.Status)
	}

	run, err = payrollSvc.ExecuteCalculations(context.Background(), run.ID)
	if err != nil {
		t.Fatalf("Failed to execute calculations: %v", err)
	}

	expectedGross := decimal.NewFromFloat(6000)
	if !run.TotalGrossPay.Equal(expectedGross) {
		t.Errorf("Expected gross %s, got %s", expectedGross, run.TotalGrossPay)
	}

	run, err = payrollSvc.CloseAndApprovePayroll(context.Background(), run.ID)
	if err != nil {
		t.Fatalf("Failed to approve payroll: %v", err)
	}

	if run.Status != domain.PayrollStatusAPPROVED {
		t.Errorf("Expected APPROVED status, got %s", run.Status)
	}
}

func TestExpenseService_SubmitAndApprove(t *testing.T) {
	db, empRepo, deptRepo, _, claimRepo, lineRepo, outboxRepo := setupTestDB(t)

	dept := &domain.Department{
		ID:             "dept-1",
		LegalEntityID:  "tenant-1",
		DepartmentCode: "DEPT_ENG",
		Name:           "Engineering",
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_ = deptRepo.Create(context.Background(), dept)

	emp := &domain.EmployeeMaster{
		ID:             "emp-1",
		LegalEntityID:  "tenant-1",
		DepartmentID:   "dept-1",
		EmployeeNumber: "EMP001",
		FirstName:      "John",
		LastName:       "Doe",
		Email:          "john@example.com",
		Status:         domain.EmployeeStatusACTIVE,
		Type:           domain.EmploymentTypeFULL_TIME,
		BaseSalary:     decimal.NewFromFloat(6000),
		Version:        1,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_ = empRepo.Create(context.Background(), emp)

	expenseSvc := service.NewExpenseService(db, claimRepo, lineRepo, empRepo, outboxRepo)

	lines := []domain.ExpenseClaimLineInput{
		{Description: "Flight Ticket", Amount: decimal.NewFromFloat(500)},
		{Description: "Hotel stay", Amount: decimal.NewFromFloat(300)},
	}

	claim, err := expenseSvc.SubmitClaim(
		context.Background(),
		"tenant-1",
		"emp-1",
		"CLAIM-101",
		"Business trip to SF",
		"COST-CENTER-RD",
		lines,
	)

	if err != nil {
		t.Fatalf("Failed to submit claim: %v", err)
	}

	expectedTotal := decimal.NewFromFloat(800)
	if !claim.TotalAmount.Equal(expectedTotal) {
		t.Errorf("Expected total amount %s, got %s", expectedTotal, claim.TotalAmount)
	}

	claim, err = expenseSvc.VerifyAndApproveClaim(context.Background(), claim.ID, "emp-1")
	if err != nil {
		t.Fatalf("Failed to approve claim: %v", err)
	}

	if claim.Status != domain.ExpenseStatusAPPROVED {
		t.Errorf("Expected APPROVED status, got %s", claim.Status)
	}
}
