package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/erp-system/hr-service/internal/business/domain"
	"github.com/erp-system/hr-service/internal/business/service"
	"github.com/erp-system/hr-service/internal/data/memory"
	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Define mock wrappers
type mockDeptRepo struct {
	*memory.MemoryDepartmentRepo
	getByIDError   error
	getByCodeError error
	listError      error
	createError    error
	updateError    error
}

func (m *mockDeptRepo) GetByID(ctx context.Context, id string) (*domain.Department, error) {
	if m.getByIDError != nil {
		return nil, m.getByIDError
	}
	return m.MemoryDepartmentRepo.GetByID(ctx, id)
}

func (m *mockDeptRepo) GetByCode(ctx context.Context, legalEntityID, code string) (*domain.Department, error) {
	if m.getByCodeError != nil {
		return nil, m.getByCodeError
	}
	return m.MemoryDepartmentRepo.GetByCode(ctx, legalEntityID, code)
}

func (m *mockDeptRepo) List(ctx context.Context) ([]domain.Department, error) {
	if m.listError != nil {
		return nil, m.listError
	}
	return m.MemoryDepartmentRepo.List(ctx)
}

func (m *mockDeptRepo) Create(ctx context.Context, dept *domain.Department) error {
	if m.createError != nil {
		return m.createError
	}
	return m.MemoryDepartmentRepo.Create(ctx, dept)
}

func (m *mockDeptRepo) Update(ctx context.Context, dept *domain.Department) error {
	if m.updateError != nil {
		return m.updateError
	}
	return m.MemoryDepartmentRepo.Update(ctx, dept)
}

type mockEmpRepo struct {
	*memory.MemoryEmployeeMasterRepo
	getByIDError     error
	getByNumberError error
	getByEmailError  error
	listError        error
	createError      error
	updateError      error
	deleteError      error
}

func (m *mockEmpRepo) GetByID(ctx context.Context, id string) (*domain.EmployeeMaster, error) {
	if m.getByIDError != nil {
		return nil, m.getByIDError
	}
	return m.MemoryEmployeeMasterRepo.GetByID(ctx, id)
}

func (m *mockEmpRepo) GetByNumber(ctx context.Context, legalEntityID, number string) (*domain.EmployeeMaster, error) {
	if m.getByNumberError != nil {
		return nil, m.getByNumberError
	}
	return m.MemoryEmployeeMasterRepo.GetByNumber(ctx, legalEntityID, number)
}

func (m *mockEmpRepo) GetByEmail(ctx context.Context, email string) (*domain.EmployeeMaster, error) {
	if m.getByEmailError != nil {
		return nil, m.getByEmailError
	}
	return m.MemoryEmployeeMasterRepo.GetByEmail(ctx, email)
}

func (m *mockEmpRepo) List(ctx context.Context) ([]domain.EmployeeMaster, error) {
	if m.listError != nil {
		return nil, m.listError
	}
	return m.MemoryEmployeeMasterRepo.List(ctx)
}

func (m *mockEmpRepo) Create(ctx context.Context, emp *domain.EmployeeMaster) error {
	if m.createError != nil {
		return m.createError
	}
	return m.MemoryEmployeeMasterRepo.Create(ctx, emp)
}

func (m *mockEmpRepo) Update(ctx context.Context, emp *domain.EmployeeMaster) error {
	if m.updateError != nil {
		return m.updateError
	}
	return m.MemoryEmployeeMasterRepo.Update(ctx, emp)
}

func (m *mockEmpRepo) Delete(ctx context.Context, id string) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	return m.MemoryEmployeeMasterRepo.Delete(ctx, id)
}

type mockOutboxRepo struct {
	*memory.MemoryTransactionalOutboxRepo
	createError    error
	getByIDError   error
	getUnsentError error
	updateError    error
}

func (m *mockOutboxRepo) Create(ctx context.Context, msg *domain.TransactionalOutbox) error {
	if m.createError != nil {
		return m.createError
	}
	return m.MemoryTransactionalOutboxRepo.Create(ctx, msg)
}

func (m *mockOutboxRepo) GetByID(ctx context.Context, id string) (*domain.TransactionalOutbox, error) {
	if m.getByIDError != nil {
		return nil, m.getByIDError
	}
	return m.MemoryTransactionalOutboxRepo.GetByID(ctx, id)
}

func (m *mockOutboxRepo) GetUnsent(ctx context.Context, limit int) ([]domain.TransactionalOutbox, error) {
	if m.getUnsentError != nil {
		return nil, m.getUnsentError
	}
	return m.MemoryTransactionalOutboxRepo.GetUnsent(ctx, limit)
}

func (m *mockOutboxRepo) Update(ctx context.Context, msg *domain.TransactionalOutbox) error {
	if m.updateError != nil {
		return m.updateError
	}
	return m.MemoryTransactionalOutboxRepo.Update(ctx, msg)
}

type mockPayrollRepo struct {
	*memory.MemoryPayrollRunRepo
	createError      error
	getByIDError     error
	getByPeriodError error
	listError        error
	updateError      error
}

func (m *mockPayrollRepo) Create(ctx context.Context, run *domain.PayrollRun) error {
	if m.createError != nil {
		return m.createError
	}
	return m.MemoryPayrollRunRepo.Create(ctx, run)
}

func (m *mockPayrollRepo) GetByID(ctx context.Context, id string) (*domain.PayrollRun, error) {
	if m.getByIDError != nil {
		return nil, m.getByIDError
	}
	return m.MemoryPayrollRunRepo.GetByID(ctx, id)
}

func (m *mockPayrollRepo) GetByPeriod(ctx context.Context, legalEntityID string, year, period int) (*domain.PayrollRun, error) {
	if m.getByPeriodError != nil {
		return nil, m.getByPeriodError
	}
	return m.MemoryPayrollRunRepo.GetByPeriod(ctx, legalEntityID, year, period)
}

func (m *mockPayrollRepo) List(ctx context.Context) ([]domain.PayrollRun, error) {
	if m.listError != nil {
		return nil, m.listError
	}
	return m.MemoryPayrollRunRepo.List(ctx)
}

func (m *mockPayrollRepo) Update(ctx context.Context, run *domain.PayrollRun) error {
	if m.updateError != nil {
		return m.updateError
	}
	return m.MemoryPayrollRunRepo.Update(ctx, run)
}

type mockExpenseClaimRepo struct {
	*memory.MemoryExpenseClaimRepo
	createError      error
	getByIDError     error
	getByNumberError error
	listError        error
	updateError      error
}

func (m *mockExpenseClaimRepo) Create(ctx context.Context, claim *domain.ExpenseClaim) error {
	if m.createError != nil {
		return m.createError
	}
	return m.MemoryExpenseClaimRepo.Create(ctx, claim)
}

func (m *mockExpenseClaimRepo) GetByID(ctx context.Context, id string) (*domain.ExpenseClaim, error) {
	if m.getByIDError != nil {
		return nil, m.getByIDError
	}
	return m.MemoryExpenseClaimRepo.GetByID(ctx, id)
}

func (m *mockExpenseClaimRepo) GetByNumber(ctx context.Context, legalEntityID, number string) (*domain.ExpenseClaim, error) {
	if m.getByNumberError != nil {
		return nil, m.getByNumberError
	}
	return m.MemoryExpenseClaimRepo.GetByNumber(ctx, legalEntityID, number)
}

func (m *mockExpenseClaimRepo) List(ctx context.Context) ([]domain.ExpenseClaim, error) {
	if m.listError != nil {
		return nil, m.listError
	}
	return m.MemoryExpenseClaimRepo.List(ctx)
}

func (m *mockExpenseClaimRepo) Update(ctx context.Context, claim *domain.ExpenseClaim) error {
	if m.updateError != nil {
		return m.updateError
	}
	return m.MemoryExpenseClaimRepo.Update(ctx, claim)
}

type mockExpenseClaimLineRepo struct {
	*memory.MemoryExpenseClaimLineRepo
	createError        error
	getByIDError       error
	listByClaimIDError error
}

func (m *mockExpenseClaimLineRepo) Create(ctx context.Context, line *domain.ExpenseClaimLine) error {
	if m.createError != nil {
		return m.createError
	}
	return m.MemoryExpenseClaimLineRepo.Create(ctx, line)
}

func (m *mockExpenseClaimLineRepo) GetByID(ctx context.Context, id string) (*domain.ExpenseClaimLine, error) {
	if m.getByIDError != nil {
		return nil, m.getByIDError
	}
	return m.MemoryExpenseClaimLineRepo.GetByID(ctx, id)
}

func (m *mockExpenseClaimLineRepo) ListByClaimID(ctx context.Context, claimID string) ([]domain.ExpenseClaimLine, error) {
	if m.listByClaimIDError != nil {
		return nil, m.listByClaimIDError
	}
	return m.MemoryExpenseClaimLineRepo.ListByClaimID(ctx, claimID)
}

type mockInboxRepo struct {
	*memory.MemoryKafkaEventInboxRepo
	createError  error
	getByIDError error
	updateError  error
}

func (m *mockInboxRepo) Create(ctx context.Context, msg *domain.KafkaEventInbox) error {
	if m.createError != nil {
		return m.createError
	}
	return m.MemoryKafkaEventInboxRepo.Create(ctx, msg)
}

func (m *mockInboxRepo) GetByID(ctx context.Context, eventID string) (*domain.KafkaEventInbox, error) {
	if m.getByIDError != nil {
		return nil, m.getByIDError
	}
	return m.MemoryKafkaEventInboxRepo.GetByID(ctx, eventID)
}

func (m *mockInboxRepo) Update(ctx context.Context, msg *domain.KafkaEventInbox) error {
	if m.updateError != nil {
		return m.updateError
	}
	return m.MemoryKafkaEventInboxRepo.Update(ctx, msg)
}

func setupTest(t *testing.T) (*gorm.DB, *mockEmpRepo, *mockDeptRepo, *mockOutboxRepo, *mockPayrollRepo, *mockExpenseClaimRepo, *mockExpenseClaimLineRepo, *mockInboxRepo) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open sqlite database: %v", err)
	}

	empRepo := &mockEmpRepo{MemoryEmployeeMasterRepo: memory.NewMemoryEmployeeMasterRepo()}
	deptRepo := &mockDeptRepo{MemoryDepartmentRepo: memory.NewMemoryDepartmentRepo()}
	outboxRepo := &mockOutboxRepo{MemoryTransactionalOutboxRepo: memory.NewMemoryTransactionalOutboxRepo()}
	payrollRepo := &mockPayrollRepo{MemoryPayrollRunRepo: memory.NewMemoryPayrollRunRepo()}
	claimRepo := &mockExpenseClaimRepo{MemoryExpenseClaimRepo: memory.NewMemoryExpenseClaimRepo()}
	lineRepo := &mockExpenseClaimLineRepo{MemoryExpenseClaimLineRepo: memory.NewMemoryExpenseClaimLineRepo()}
	inboxRepo := &mockInboxRepo{MemoryKafkaEventInboxRepo: memory.NewMemoryKafkaEventInboxRepo()}

	return db, empRepo, deptRepo, outboxRepo, payrollRepo, claimRepo, lineRepo, inboxRepo
}

func TestEmployeeService_HireEmployee_Success(t *testing.T) {
	db, empRepo, deptRepo, outboxRepo, _, _, _, _ := setupTest(t)

	// Create department
	dept := &domain.Department{
		ID:            "dept-1",
		LegalEntityID: "tenant-1",
		IsActive:      true,
	}
	_ = deptRepo.Create(context.Background(), dept)

	// Test hire without manager
	svc := service.NewEmployeeService(db, empRepo, deptRepo, outboxRepo)
	salary := decimal.NewFromFloat(5000)
	emp, err := svc.HireEmployee(context.Background(), "tenant-1", "dept-1", nil, "EMP001", "John", "Doe", "john@example.com", salary, domain.EmploymentTypeFULL_TIME)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if emp.OrgDepthLevel != 0 {
		t.Errorf("Expected OrgDepthLevel 0, got %d", emp.OrgDepthLevel)
	}

	// Verify outbox entry
	unsent, err := outboxRepo.GetUnsent(context.Background(), 10)
	if err != nil || len(unsent) != 1 {
		t.Fatalf("Expected 1 unsent outbox msg, got: %d (err: %v)", len(unsent), err)
	}
	if unsent[0].AggregateID != emp.ID {
		t.Errorf("Expected outbox aggregate ID %s, got %s", emp.ID, unsent[0].AggregateID)
	}
}

func TestEmployeeService_HireEmployee_WithManager_Success(t *testing.T) {
	db, empRepo, deptRepo, outboxRepo, _, _, _, _ := setupTest(t)

	// Create department
	dept := &domain.Department{
		ID:            "dept-1",
		LegalEntityID: "tenant-1",
		IsActive:      true,
	}
	_ = deptRepo.Create(context.Background(), dept)

	// Create manager
	manager := &domain.EmployeeMaster{
		ID:            "mgr-1",
		OrgDepthLevel: 2,
	}
	_ = empRepo.Create(context.Background(), manager)

	// Test hire with manager
	svc := service.NewEmployeeService(db, empRepo, deptRepo, outboxRepo)
	salary := decimal.NewFromFloat(5000)
	mgrID := "mgr-1"
	emp, err := svc.HireEmployee(context.Background(), "tenant-1", "dept-1", &mgrID, "EMP002", "Jane", "Doe", "jane@example.com", salary, domain.EmploymentTypeFULL_TIME)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if emp.OrgDepthLevel != 3 {
		t.Errorf("Expected OrgDepthLevel 3, got %d", emp.OrgDepthLevel)
	}
}

func TestEmployeeService_HireEmployee_DepartmentNotFound(t *testing.T) {
	db, empRepo, deptRepo, outboxRepo, _, _, _, _ := setupTest(t)

	// Test hire with non-existent department
	svc := service.NewEmployeeService(db, empRepo, deptRepo, outboxRepo)
	salary := decimal.NewFromFloat(5000)
	_, err := svc.HireEmployee(context.Background(), "tenant-1", "dept-1", nil, "EMP001", "John", "Doe", "john@example.com", salary, domain.EmploymentTypeFULL_TIME)
	if err == nil {
		t.Fatal("Expected error department not found, got nil")
	}
}

func TestEmployeeService_HireEmployee_ManagerNotFound(t *testing.T) {
	db, empRepo, deptRepo, outboxRepo, _, _, _, _ := setupTest(t)

	// Create department
	dept := &domain.Department{
		ID:            "dept-1",
		LegalEntityID: "tenant-1",
		IsActive:      true,
	}
	_ = deptRepo.Create(context.Background(), dept)

	svc := service.NewEmployeeService(db, empRepo, deptRepo, outboxRepo)
	salary := decimal.NewFromFloat(5000)
	mgrID := "mgr-1"
	_, err := svc.HireEmployee(context.Background(), "tenant-1", "dept-1", &mgrID, "EMP001", "John", "Doe", "john@example.com", salary, domain.EmploymentTypeFULL_TIME)
	if err == nil {
		t.Fatal("Expected error manager not found, got nil")
	}
}

func TestEmployeeService_HireEmployee_CreateEmpFails(t *testing.T) {
	db, empRepo, deptRepo, outboxRepo, _, _, _, _ := setupTest(t)

	dept := &domain.Department{
		ID:            "dept-1",
		LegalEntityID: "tenant-1",
		IsActive:      true,
	}
	_ = deptRepo.Create(context.Background(), dept)

	empRepo.createError = errors.New("db create emp error")

	svc := service.NewEmployeeService(db, empRepo, deptRepo, outboxRepo)
	salary := decimal.NewFromFloat(5000)
	_, err := svc.HireEmployee(context.Background(), "tenant-1", "dept-1", nil, "EMP001", "John", "Doe", "john@example.com", salary, domain.EmploymentTypeFULL_TIME)
	if err == nil || err.Error() != "db create emp error" {
		t.Fatalf("Expected 'db create emp error', got %v", err)
	}
}

func TestEmployeeService_HireEmployee_CreateOutboxFails(t *testing.T) {
	db, empRepo, deptRepo, outboxRepo, _, _, _, _ := setupTest(t)

	dept := &domain.Department{
		ID:            "dept-1",
		LegalEntityID: "tenant-1",
		IsActive:      true,
	}
	_ = deptRepo.Create(context.Background(), dept)

	outboxRepo.createError = errors.New("db create outbox error")

	svc := service.NewEmployeeService(db, empRepo, deptRepo, outboxRepo)
	salary := decimal.NewFromFloat(5000)
	_, err := svc.HireEmployee(context.Background(), "tenant-1", "dept-1", nil, "EMP001", "John", "Doe", "john@example.com", salary, domain.EmploymentTypeFULL_TIME)
	if err == nil || err.Error() != "db create outbox error" {
		t.Fatalf("Expected 'db create outbox error', got %v", err)
	}
}

func TestEmployeeService_TerminateEmployee_Success(t *testing.T) {
	db, empRepo, _, outboxRepo, _, _, _, _ := setupTest(t)

	emp := &domain.EmployeeMaster{
		ID:             "emp-1",
		Status:         domain.EmployeeStatusACTIVE,
		EmployeeNumber: "EMP001",
	}
	_ = empRepo.Create(context.Background(), emp)

	svc := service.NewEmployeeService(db, empRepo, nil, outboxRepo)
	res, err := svc.TerminateEmployee(context.Background(), "emp-1", time.Now())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if res.Status != domain.EmployeeStatusTERMINATED {
		t.Errorf("Expected status TERMINATED, got %s", res.Status)
	}
	if res.DeletedAt == nil {
		t.Error("Expected DeletedAt to be set")
	}

	unsent, err := outboxRepo.GetUnsent(context.Background(), 10)
	if err != nil || len(unsent) != 1 {
		t.Fatalf("Expected 1 unsent outbox msg, got: %d (err: %v)", len(unsent), err)
	}
}

func TestEmployeeService_TerminateEmployee_NotFound(t *testing.T) {
	db, empRepo, _, outboxRepo, _, _, _, _ := setupTest(t)

	svc := service.NewEmployeeService(db, empRepo, nil, outboxRepo)
	_, err := svc.TerminateEmployee(context.Background(), "emp-1", time.Now())
	if err == nil {
		t.Fatal("Expected error employee not found, got nil")
	}
}

func TestEmployeeService_TerminateEmployee_UpdateFails(t *testing.T) {
	db, empRepo, _, outboxRepo, _, _, _, _ := setupTest(t)

	emp := &domain.EmployeeMaster{
		ID: "emp-1",
	}
	_ = empRepo.Create(context.Background(), emp)

	empRepo.updateError = errors.New("db update emp error")

	svc := service.NewEmployeeService(db, empRepo, nil, outboxRepo)
	_, err := svc.TerminateEmployee(context.Background(), "emp-1", time.Now())
	if err == nil || err.Error() != "db update emp error" {
		t.Fatalf("Expected 'db update emp error', got %v", err)
	}
}

func TestEmployeeService_TerminateEmployee_CreateOutboxFails(t *testing.T) {
	db, empRepo, _, outboxRepo, _, _, _, _ := setupTest(t)

	emp := &domain.EmployeeMaster{
		ID: "emp-1",
	}
	_ = empRepo.Create(context.Background(), emp)

	outboxRepo.createError = errors.New("db create outbox error")

	svc := service.NewEmployeeService(db, empRepo, nil, outboxRepo)
	_, err := svc.TerminateEmployee(context.Background(), "emp-1", time.Now())
	if err == nil || err.Error() != "db create outbox error" {
		t.Fatalf("Expected 'db create outbox error', got %v", err)
	}
}

func TestEmployeeService_AdjustCompensation_Success(t *testing.T) {
	db, empRepo, _, _, _, _, _, _ := setupTest(t)

	emp := &domain.EmployeeMaster{
		ID:         "emp-1",
		BaseSalary: decimal.NewFromFloat(5000),
	}
	_ = empRepo.Create(context.Background(), emp)

	svc := service.NewEmployeeService(db, empRepo, nil, nil)
	res, err := svc.AdjustCompensation(context.Background(), "emp-1", decimal.NewFromFloat(6000))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !res.BaseSalary.Equal(decimal.NewFromFloat(6000)) {
		t.Errorf("Expected salary 6000, got %s", res.BaseSalary)
	}
}

func TestEmployeeService_AdjustCompensation_NotFound(t *testing.T) {
	db, empRepo, _, _, _, _, _, _ := setupTest(t)

	svc := service.NewEmployeeService(db, empRepo, nil, nil)
	_, err := svc.AdjustCompensation(context.Background(), "emp-1", decimal.NewFromFloat(6000))
	if err == nil {
		t.Fatal("Expected error employee not found, got nil")
	}
}

func TestEmployeeService_AdjustCompensation_UpdateFails(t *testing.T) {
	db, empRepo, _, _, _, _, _, _ := setupTest(t)

	emp := &domain.EmployeeMaster{
		ID:         "emp-1",
		BaseSalary: decimal.NewFromFloat(5000),
	}
	_ = empRepo.Create(context.Background(), emp)

	empRepo.updateError = errors.New("db update error")

	svc := service.NewEmployeeService(db, empRepo, nil, nil)
	_, err := svc.AdjustCompensation(context.Background(), "emp-1", decimal.NewFromFloat(6000))
	if err == nil || err.Error() != "db update error" {
		t.Fatalf("Expected 'db update error', got %v", err)
	}
}

func TestEmployeeService_FetchManagementChain_Success(t *testing.T) {
	db, empRepo, _, _, _, _, _, _ := setupTest(t)

	mgr2ID := "mgr-2"
	mgr1ID := "mgr-1"

	emp := domain.EmployeeMaster{
		ID:          "emp-1",
		ManagerHrID: &mgr1ID,
	}
	mgr1 := domain.EmployeeMaster{
		ID:          "mgr-1",
		ManagerHrID: &mgr2ID,
	}
	mgr2 := domain.EmployeeMaster{
		ID:          "mgr-2",
		ManagerHrID: nil,
	}

	_ = empRepo.Create(context.Background(), &emp)
	_ = empRepo.Create(context.Background(), &mgr1)
	_ = empRepo.Create(context.Background(), &mgr2)

	svc := service.NewEmployeeService(db, empRepo, nil, nil)
	chain, err := svc.FetchManagementChain(context.Background(), "emp-1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(chain) != 3 {
		t.Fatalf("Expected chain length 3, got %d", len(chain))
	}
	if chain[0].ID != "emp-1" || chain[1].ID != "mgr-1" || chain[2].ID != "mgr-2" {
		t.Errorf("Unexpected chain order or elements: %v", chain)
	}
}

func TestEmployeeService_FetchManagementChain_SelfReferencing(t *testing.T) {
	db, empRepo, _, _, _, _, _, _ := setupTest(t)

	empID := "emp-1"
	emp := domain.EmployeeMaster{
		ID:          "emp-1",
		ManagerHrID: &empID,
	}
	_ = empRepo.Create(context.Background(), &emp)

	svc := service.NewEmployeeService(db, empRepo, nil, nil)
	chain, err := svc.FetchManagementChain(context.Background(), "emp-1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(chain) != 1 {
		t.Fatalf("Expected chain length 1, got %d", len(chain))
	}
	if chain[0].ID != "emp-1" {
		t.Errorf("Unexpected chain: %v", chain)
	}
}

func TestEmployeeService_FetchManagementChain_EmployeeNotFound(t *testing.T) {
	db, empRepo, _, _, _, _, _, _ := setupTest(t)

	svc := service.NewEmployeeService(db, empRepo, nil, nil)
	chain, err := svc.FetchManagementChain(context.Background(), "emp-1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(chain) != 0 {
		t.Fatalf("Expected empty chain, got %v", chain)
	}
}

func TestPayrollService_InitiatePeriodRun_Success(t *testing.T) {
	db, empRepo, _, outboxRepo, payrollRepo, _, _, _ := setupTest(t)

	svc := service.NewPayrollService(db, payrollRepo, empRepo, outboxRepo)
	run, err := svc.InitiatePeriodRun(context.Background(), "tenant-1", 2026, 6)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if run.Status != domain.PayrollStatusDRAFT {
		t.Errorf("Expected DRAFT, got %s", run.Status)
	}
}

func TestPayrollService_InitiatePeriodRun_AlreadyExists(t *testing.T) {
	db, empRepo, _, outboxRepo, payrollRepo, _, _, _ := setupTest(t)

	existing := &domain.PayrollRun{
		ID:            "pay-existing",
		LegalEntityID: "tenant-1",
		FiscalYear:    2026,
		PeriodNumber:  6,
	}
	_ = payrollRepo.Create(context.Background(), existing)

	svc := service.NewPayrollService(db, payrollRepo, empRepo, outboxRepo)
	_, err := svc.InitiatePeriodRun(context.Background(), "tenant-1", 2026, 6)
	if err == nil || err.Error() != "payroll run for this period already exists" {
		t.Fatalf("Expected 'payroll run for this period already exists', got %v", err)
	}
}

func TestPayrollService_InitiatePeriodRun_CreateFails(t *testing.T) {
	db, empRepo, _, outboxRepo, payrollRepo, _, _, _ := setupTest(t)

	payrollRepo.createError = errors.New("db create error")

	svc := service.NewPayrollService(db, payrollRepo, empRepo, outboxRepo)
	_, err := svc.InitiatePeriodRun(context.Background(), "tenant-1", 2026, 6)
	if err == nil || err.Error() != "db create error" {
		t.Fatalf("Expected 'db create error', got %v", err)
	}
}

func TestPayrollService_ExecuteCalculations_Success(t *testing.T) {
	db, empRepo, _, outboxRepo, payrollRepo, _, _, _ := setupTest(t)

	run := &domain.PayrollRun{
		ID:            "pay-1",
		LegalEntityID: "tenant-1",
		Status:        domain.PayrollStatusDRAFT,
	}
	_ = payrollRepo.Create(context.Background(), run)

	emp1 := &domain.EmployeeMaster{
		ID:            "emp-1",
		LegalEntityID: "tenant-1",
		Status:        domain.EmployeeStatusACTIVE,
		BaseSalary:    decimal.NewFromFloat(5000),
	}
	emp2 := &domain.EmployeeMaster{
		ID:            "emp-2",
		LegalEntityID: "tenant-1",
		Status:        domain.EmployeeStatusTERMINATED, // Should be ignored
		BaseSalary:    decimal.NewFromFloat(3000),
	}
	emp3 := &domain.EmployeeMaster{
		ID:            "emp-3",
		LegalEntityID: "tenant-2", // Different tenant, should be ignored
		Status:        domain.EmployeeStatusACTIVE,
		BaseSalary:    decimal.NewFromFloat(4000),
	}
	_ = empRepo.Create(context.Background(), emp1)
	_ = empRepo.Create(context.Background(), emp2)
	_ = empRepo.Create(context.Background(), emp3)

	svc := service.NewPayrollService(db, payrollRepo, empRepo, outboxRepo)
	res, err := svc.ExecuteCalculations(context.Background(), "pay-1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedGross := decimal.NewFromFloat(5000)
	expectedDeductions := decimal.NewFromFloat(500)
	expectedNet := decimal.NewFromFloat(4500)

	if !res.TotalGrossPay.Equal(expectedGross) {
		t.Errorf("Expected gross %s, got %s", expectedGross, res.TotalGrossPay)
	}
	if !res.TotalDeductions.Equal(expectedDeductions) {
		t.Errorf("Expected deductions %s, got %s", expectedDeductions, res.TotalDeductions)
	}
	if !res.TotalNetPay.Equal(expectedNet) {
		t.Errorf("Expected net %s, got %s", expectedNet, res.TotalNetPay)
	}
}

func TestPayrollService_ExecuteCalculations_NotFound(t *testing.T) {
	db, empRepo, _, outboxRepo, payrollRepo, _, _, _ := setupTest(t)

	svc := service.NewPayrollService(db, payrollRepo, empRepo, outboxRepo)
	_, err := svc.ExecuteCalculations(context.Background(), "pay-1")
	if err == nil {
		t.Fatal("Expected error payroll run not found, got nil")
	}
}

func TestPayrollService_ExecuteCalculations_NotDraft(t *testing.T) {
	db, empRepo, _, outboxRepo, payrollRepo, _, _, _ := setupTest(t)

	run := &domain.PayrollRun{
		ID:     "pay-1",
		Status: domain.PayrollStatusAPPROVED,
	}
	_ = payrollRepo.Create(context.Background(), run)

	svc := service.NewPayrollService(db, payrollRepo, empRepo, outboxRepo)
	_, err := svc.ExecuteCalculations(context.Background(), "pay-1")
	if err == nil || err.Error() != "calculations can only be executed on DRAFT payroll runs" {
		t.Fatalf("Expected 'calculations can only be executed on DRAFT payroll runs', got %v", err)
	}
}

func TestPayrollService_ExecuteCalculations_EmployeeListFails(t *testing.T) {
	db, empRepo, _, outboxRepo, payrollRepo, _, _, _ := setupTest(t)

	run := &domain.PayrollRun{
		ID:     "pay-1",
		Status: domain.PayrollStatusDRAFT,
	}
	_ = payrollRepo.Create(context.Background(), run)

	empRepo.listError = errors.New("employee list error")

	svc := service.NewPayrollService(db, payrollRepo, empRepo, outboxRepo)
	_, err := svc.ExecuteCalculations(context.Background(), "pay-1")
	if err == nil || err.Error() != "employee list error" {
		t.Fatalf("Expected 'employee list error', got %v", err)
	}
}

func TestPayrollService_ExecuteCalculations_UpdateFails(t *testing.T) {
	db, empRepo, _, outboxRepo, payrollRepo, _, _, _ := setupTest(t)

	run := &domain.PayrollRun{
		ID:     "pay-1",
		Status: domain.PayrollStatusDRAFT,
	}
	_ = payrollRepo.Create(context.Background(), run)

	payrollRepo.updateError = errors.New("db update error")

	svc := service.NewPayrollService(db, payrollRepo, empRepo, outboxRepo)
	_, err := svc.ExecuteCalculations(context.Background(), "pay-1")
	if err == nil || err.Error() != "db update error" {
		t.Fatalf("Expected 'db update error', got %v", err)
	}
}

func TestPayrollService_CloseAndApprovePayroll_Success(t *testing.T) {
	db, empRepo, _, outboxRepo, payrollRepo, _, _, _ := setupTest(t)

	run := &domain.PayrollRun{
		ID:            "pay-1",
		LegalEntityID: "tenant-1",
		Status:        domain.PayrollStatusDRAFT,
	}
	_ = payrollRepo.Create(context.Background(), run)

	svc := service.NewPayrollService(db, payrollRepo, empRepo, outboxRepo)
	res, err := svc.CloseAndApprovePayroll(context.Background(), "pay-1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if res.Status != domain.PayrollStatusAPPROVED {
		t.Errorf("Expected APPROVED, got %s", res.Status)
	}

	unsent, err := outboxRepo.GetUnsent(context.Background(), 10)
	if err != nil || len(unsent) != 1 {
		t.Fatalf("Expected 1 unsent outbox, got %d (err: %v)", len(unsent), err)
	}
}

func TestPayrollService_CloseAndApprovePayroll_NotFound(t *testing.T) {
	db, empRepo, _, outboxRepo, payrollRepo, _, _, _ := setupTest(t)

	svc := service.NewPayrollService(db, payrollRepo, empRepo, outboxRepo)
	_, err := svc.CloseAndApprovePayroll(context.Background(), "pay-1")
	if err == nil {
		t.Fatal("Expected error payroll not found, got nil")
	}
}

func TestPayrollService_CloseAndApprovePayroll_NotDraft(t *testing.T) {
	db, empRepo, _, outboxRepo, payrollRepo, _, _, _ := setupTest(t)

	run := &domain.PayrollRun{
		ID:     "pay-1",
		Status: domain.PayrollStatusAPPROVED,
	}
	_ = payrollRepo.Create(context.Background(), run)

	svc := service.NewPayrollService(db, payrollRepo, empRepo, outboxRepo)
	_, err := svc.CloseAndApprovePayroll(context.Background(), "pay-1")
	if err == nil || err.Error() != "only DRAFT payroll runs can be approved" {
		t.Fatalf("Expected 'only DRAFT payroll runs can be approved', got %v", err)
	}
}

func TestPayrollService_CloseAndApprovePayroll_UpdateFails(t *testing.T) {
	db, empRepo, _, outboxRepo, payrollRepo, _, _, _ := setupTest(t)

	run := &domain.PayrollRun{
		ID:     "pay-1",
		Status: domain.PayrollStatusDRAFT,
	}
	_ = payrollRepo.Create(context.Background(), run)

	payrollRepo.updateError = errors.New("db update error")

	svc := service.NewPayrollService(db, payrollRepo, empRepo, outboxRepo)
	_, err := svc.CloseAndApprovePayroll(context.Background(), "pay-1")
	if err == nil || err.Error() != "db update error" {
		t.Fatalf("Expected 'db update error', got %v", err)
	}
}

func TestPayrollService_CloseAndApprovePayroll_OutboxFails(t *testing.T) {
	db, empRepo, _, outboxRepo, payrollRepo, _, _, _ := setupTest(t)

	run := &domain.PayrollRun{
		ID:     "pay-1",
		Status: domain.PayrollStatusDRAFT,
	}
	_ = payrollRepo.Create(context.Background(), run)

	outboxRepo.createError = errors.New("db outbox create error")

	svc := service.NewPayrollService(db, payrollRepo, empRepo, outboxRepo)
	_, err := svc.CloseAndApprovePayroll(context.Background(), "pay-1")
	if err == nil || err.Error() != "db outbox create error" {
		t.Fatalf("Expected 'db outbox create error', got %v", err)
	}
}

func TestExpenseService_SubmitClaim_Success(t *testing.T) {
	db, empRepo, _, _, _, claimRepo, lineRepo, _ := setupTest(t)

	emp := &domain.EmployeeMaster{
		ID: "emp-1",
	}
	_ = empRepo.Create(context.Background(), emp)

	lines := []domain.ExpenseClaimLineInput{
		{Description: "line 1", Amount: decimal.NewFromFloat(100)},
		{Description: "line 2", Amount: decimal.NewFromFloat(150)},
	}

	svc := service.NewExpenseService(db, claimRepo, lineRepo, empRepo, nil)
	claim, err := svc.SubmitClaim(context.Background(), "tenant-1", "emp-1", "CLAIM001", "Business trip", "CC01", lines)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !claim.TotalAmount.Equal(decimal.NewFromFloat(250)) {
		t.Errorf("Expected total amount 250, got %s", claim.TotalAmount)
	}
	if claim.Status != domain.ExpenseStatusSUBMITTED {
		t.Errorf("Expected status SUBMITTED, got %s", claim.Status)
	}

	// Verify lines pre-saved
	savedLines, err := lineRepo.ListByClaimID(context.Background(), claim.ID)
	if err != nil || len(savedLines) != 2 {
		t.Fatalf("Expected 2 saved lines, got %d (err: %v)", len(savedLines), err)
	}
}

func TestExpenseService_SubmitClaim_EmployeeNotFound(t *testing.T) {
	db, empRepo, _, _, _, claimRepo, lineRepo, _ := setupTest(t)

	svc := service.NewExpenseService(db, claimRepo, lineRepo, empRepo, nil)
	_, err := svc.SubmitClaim(context.Background(), "tenant-1", "emp-1", "CLAIM001", "Business trip", "CC01", nil)
	if err == nil {
		t.Fatal("Expected error employee not found, got nil")
	}
}

func TestExpenseService_SubmitClaim_ClaimCreateFails(t *testing.T) {
	db, empRepo, _, _, _, claimRepo, lineRepo, _ := setupTest(t)

	emp := &domain.EmployeeMaster{
		ID: "emp-1",
	}
	_ = empRepo.Create(context.Background(), emp)

	claimRepo.createError = errors.New("claim create error")

	svc := service.NewExpenseService(db, claimRepo, lineRepo, empRepo, nil)
	_, err := svc.SubmitClaim(context.Background(), "tenant-1", "emp-1", "CLAIM001", "Business trip", "CC01", nil)
	if err == nil || err.Error() != "claim create error" {
		t.Fatalf("Expected 'claim create error', got %v", err)
	}
}

func TestExpenseService_SubmitClaim_LineCreateFails(t *testing.T) {
	db, empRepo, _, _, _, claimRepo, lineRepo, _ := setupTest(t)

	emp := &domain.EmployeeMaster{
		ID: "emp-1",
	}
	_ = empRepo.Create(context.Background(), emp)

	lineRepo.createError = errors.New("line create error")

	lines := []domain.ExpenseClaimLineInput{
		{Description: "line 1", Amount: decimal.NewFromFloat(100)},
	}

	svc := service.NewExpenseService(db, claimRepo, lineRepo, empRepo, nil)
	_, err := svc.SubmitClaim(context.Background(), "tenant-1", "emp-1", "CLAIM001", "Business trip", "CC01", lines)
	if err == nil || err.Error() != "line create error" {
		t.Fatalf("Expected 'line create error', got %v", err)
	}
}

func TestExpenseService_VerifyAndApproveClaim_Success(t *testing.T) {
	db, empRepo, _, outboxRepo, _, claimRepo, lineRepo, _ := setupTest(t)

	reviewer := &domain.EmployeeMaster{
		ID: "reviewer-1",
	}
	_ = empRepo.Create(context.Background(), reviewer)

	claim := &domain.ExpenseClaim{
		ID:            "claim-1",
		Status:        domain.ExpenseStatusSUBMITTED,
		LegalEntityID: "tenant-1",
	}
	_ = claimRepo.Create(context.Background(), claim)

	svc := service.NewExpenseService(db, claimRepo, lineRepo, empRepo, outboxRepo)
	res, err := svc.VerifyAndApproveClaim(context.Background(), "claim-1", "reviewer-1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if res.Status != domain.ExpenseStatusAPPROVED {
		t.Errorf("Expected APPROVED status, got %s", res.Status)
	}

	unsent, err := outboxRepo.GetUnsent(context.Background(), 10)
	if err != nil || len(unsent) != 1 {
		t.Fatalf("Expected 1 unsent outbox, got %d (err: %v)", len(unsent), err)
	}
}

func TestExpenseService_VerifyAndApproveClaim_ClaimNotFound(t *testing.T) {
	db, empRepo, _, outboxRepo, _, claimRepo, lineRepo, _ := setupTest(t)

	svc := service.NewExpenseService(db, claimRepo, lineRepo, empRepo, outboxRepo)
	_, err := svc.VerifyAndApproveClaim(context.Background(), "claim-1", "reviewer-1")
	if err == nil {
		t.Fatal("Expected error claim not found, got nil")
	}
}

func TestExpenseService_VerifyAndApproveClaim_NotSubmitted(t *testing.T) {
	db, empRepo, _, outboxRepo, _, claimRepo, lineRepo, _ := setupTest(t)

	claim := &domain.ExpenseClaim{
		ID:     "claim-1",
		Status: domain.ExpenseStatusAPPROVED,
	}
	_ = claimRepo.Create(context.Background(), claim)

	svc := service.NewExpenseService(db, claimRepo, lineRepo, empRepo, outboxRepo)
	_, err := svc.VerifyAndApproveClaim(context.Background(), "claim-1", "reviewer-1")
	if err == nil || err.Error() != "only SUBMITTED claims can be approved" {
		t.Fatalf("Expected 'only SUBMITTED claims can be approved', got %v", err)
	}
}

func TestExpenseService_VerifyAndApproveClaim_ReviewerNotFound(t *testing.T) {
	db, empRepo, _, outboxRepo, _, claimRepo, lineRepo, _ := setupTest(t)

	claim := &domain.ExpenseClaim{
		ID:     "claim-1",
		Status: domain.ExpenseStatusSUBMITTED,
	}
	_ = claimRepo.Create(context.Background(), claim)

	svc := service.NewExpenseService(db, claimRepo, lineRepo, empRepo, outboxRepo)
	_, err := svc.VerifyAndApproveClaim(context.Background(), "claim-1", "reviewer-1")
	if err == nil {
		t.Fatal("Expected error reviewer not found, got nil")
	}
}

func TestExpenseService_VerifyAndApproveClaim_UpdateFails(t *testing.T) {
	db, empRepo, _, outboxRepo, _, claimRepo, lineRepo, _ := setupTest(t)

	reviewer := &domain.EmployeeMaster{
		ID: "reviewer-1",
	}
	_ = empRepo.Create(context.Background(), reviewer)

	claim := &domain.ExpenseClaim{
		ID:     "claim-1",
		Status: domain.ExpenseStatusSUBMITTED,
	}
	_ = claimRepo.Create(context.Background(), claim)

	claimRepo.updateError = errors.New("claim update error")

	svc := service.NewExpenseService(db, claimRepo, lineRepo, empRepo, outboxRepo)
	_, err := svc.VerifyAndApproveClaim(context.Background(), "claim-1", "reviewer-1")
	if err == nil || err.Error() != "claim update error" {
		t.Fatalf("Expected 'claim update error', got %v", err)
	}
}

func TestExpenseService_VerifyAndApproveClaim_OutboxFails(t *testing.T) {
	db, empRepo, _, outboxRepo, _, claimRepo, lineRepo, _ := setupTest(t)

	reviewer := &domain.EmployeeMaster{
		ID: "reviewer-1",
	}
	_ = empRepo.Create(context.Background(), reviewer)

	claim := &domain.ExpenseClaim{
		ID:     "claim-1",
		Status: domain.ExpenseStatusSUBMITTED,
	}
	_ = claimRepo.Create(context.Background(), claim)

	outboxRepo.createError = errors.New("outbox create error")

	svc := service.NewExpenseService(db, claimRepo, lineRepo, empRepo, outboxRepo)
	_, err := svc.VerifyAndApproveClaim(context.Background(), "claim-1", "reviewer-1")
	if err == nil || err.Error() != "outbox create error" {
		t.Fatalf("Expected 'outbox create error', got %v", err)
	}
}

func TestExpenseService_ClearClaimForPayment_Success(t *testing.T) {
	db, _, _, _, _, claimRepo, lineRepo, _ := setupTest(t)

	claim := &domain.ExpenseClaim{
		ID:     "claim-1",
		Status: domain.ExpenseStatusAPPROVED,
	}
	_ = claimRepo.Create(context.Background(), claim)

	svc := service.NewExpenseService(db, claimRepo, lineRepo, nil, nil)
	err := svc.ClearClaimForPayment(context.Background(), "claim-1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	updated, _ := claimRepo.GetByID(context.Background(), "claim-1")
	if updated.Status != domain.ExpenseStatusPAID {
		t.Errorf("Expected status PAID, got %s", updated.Status)
	}
}

func TestExpenseService_ClearClaimForPayment_NotFound(t *testing.T) {
	db, _, _, _, _, claimRepo, lineRepo, _ := setupTest(t)

	svc := service.NewExpenseService(db, claimRepo, lineRepo, nil, nil)
	err := svc.ClearClaimForPayment(context.Background(), "claim-1")
	if err == nil {
		t.Fatal("Expected error claim not found, got nil")
	}
}

func TestExpenseService_ClearClaimForPayment_UpdateFails(t *testing.T) {
	db, _, _, _, _, claimRepo, lineRepo, _ := setupTest(t)

	claim := &domain.ExpenseClaim{
		ID:     "claim-1",
		Status: domain.ExpenseStatusAPPROVED,
	}
	_ = claimRepo.Create(context.Background(), claim)

	claimRepo.updateError = errors.New("db error")

	svc := service.NewExpenseService(db, claimRepo, lineRepo, nil, nil)
	err := svc.ClearClaimForPayment(context.Background(), "claim-1")
	if err == nil || err.Error() != "db error" {
		t.Fatalf("Expected 'db error', got %v", err)
	}
}

func TestOutboxRelayWorker_GetUnsentMessages_Success(t *testing.T) {
	_, _, _, outboxRepo, _, _, _, _ := setupTest(t)

	msg1 := &domain.TransactionalOutbox{
		ID:        "msg-1",
		Status:    domain.OutboxStatusPENDING,
		CreatedAt: time.Now().Add(-10 * time.Minute),
	}
	msg2 := &domain.TransactionalOutbox{
		ID:        "msg-2",
		Status:    domain.OutboxStatusPENDING,
		CreatedAt: time.Now().Add(-5 * time.Minute),
	}
	msg3 := &domain.TransactionalOutbox{
		ID:        "msg-3",
		Status:    domain.OutboxStatusSENT,
		CreatedAt: time.Now().Add(-2 * time.Minute),
	}
	_ = outboxRepo.Create(context.Background(), msg1)
	_ = outboxRepo.Create(context.Background(), msg2)
	_ = outboxRepo.Create(context.Background(), msg3)

	worker := service.NewOutboxRelayWorker(outboxRepo)
	msgs, err := worker.GetUnsentMessages(context.Background(), 1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("Expected 1 msg, got %d", len(msgs))
	}
	if msgs[0].ID != "msg-1" {
		t.Errorf("Expected older message msg-1, got %s", msgs[0].ID)
	}
}

func TestOutboxRelayWorker_GetUnsentMessages_Error(t *testing.T) {
	_, _, _, outboxRepo, _, _, _, _ := setupTest(t)

	outboxRepo.getUnsentError = errors.New("db error")

	worker := service.NewOutboxRelayWorker(outboxRepo)
	_, err := worker.GetUnsentMessages(context.Background(), 5)
	if err == nil || err.Error() != "db error" {
		t.Fatalf("Expected 'db error', got %v", err)
	}
}

func TestOutboxRelayWorker_LogProcessingAttempt_UnderLimit(t *testing.T) {
	_, _, _, outboxRepo, _, _, _, _ := setupTest(t)

	msg := &domain.TransactionalOutbox{
		ID:     "msg-1",
		Status: domain.OutboxStatusPENDING,
	}
	_ = outboxRepo.Create(context.Background(), msg)

	worker := service.NewOutboxRelayWorker(outboxRepo)
	err := worker.LogProcessingAttempt(context.Background(), "msg-1", 3, "failed to connect")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	updated, _ := outboxRepo.GetByID(context.Background(), "msg-1")
	if updated.Status != domain.OutboxStatusPENDING {
		t.Errorf("Expected status to remain PENDING, got %s", updated.Status)
	}
}

func TestOutboxRelayWorker_LogProcessingAttempt_OverLimit(t *testing.T) {
	_, _, _, outboxRepo, _, _, _, _ := setupTest(t)

	msg := &domain.TransactionalOutbox{
		ID:     "msg-1",
		Status: domain.OutboxStatusPENDING,
	}
	_ = outboxRepo.Create(context.Background(), msg)

	worker := service.NewOutboxRelayWorker(outboxRepo)
	err := worker.LogProcessingAttempt(context.Background(), "msg-1", 5, "failed to connect")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	updated, _ := outboxRepo.GetByID(context.Background(), "msg-1")
	if updated.Status != domain.OutboxStatusFAILED {
		t.Errorf("Expected status to be FAILED, got %s", updated.Status)
	}
}

func TestOutboxRelayWorker_LogProcessingAttempt_NotFound(t *testing.T) {
	_, _, _, outboxRepo, _, _, _, _ := setupTest(t)

	worker := service.NewOutboxRelayWorker(outboxRepo)
	err := worker.LogProcessingAttempt(context.Background(), "msg-1", 3, "failed to connect")
	if err == nil {
		t.Fatal("Expected error msg not found, got nil")
	}
}

func TestOutboxRelayWorker_LogProcessingAttempt_UpdateFails(t *testing.T) {
	_, _, _, outboxRepo, _, _, _, _ := setupTest(t)

	msg := &domain.TransactionalOutbox{
		ID:     "msg-1",
		Status: domain.OutboxStatusPENDING,
	}
	_ = outboxRepo.Create(context.Background(), msg)

	outboxRepo.updateError = errors.New("update error")

	worker := service.NewOutboxRelayWorker(outboxRepo)
	err := worker.LogProcessingAttempt(context.Background(), "msg-1", 5, "error")
	if err == nil || err.Error() != "update error" {
		t.Fatalf("Expected 'update error', got %v", err)
	}
}

func TestOutboxRelayWorker_UpdateOutboxStatus_Success(t *testing.T) {
	_, _, _, outboxRepo, _, _, _, _ := setupTest(t)

	msg := &domain.TransactionalOutbox{
		ID:     "msg-1",
		Status: domain.OutboxStatusPENDING,
	}
	_ = outboxRepo.Create(context.Background(), msg)

	worker := service.NewOutboxRelayWorker(outboxRepo)
	err := worker.UpdateOutboxStatus(context.Background(), "msg-1", domain.OutboxStatusSENT)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	updated, _ := outboxRepo.GetByID(context.Background(), "msg-1")
	if updated.Status != domain.OutboxStatusSENT {
		t.Errorf("Expected status SENT, got %s", updated.Status)
	}
}

func TestOutboxRelayWorker_UpdateOutboxStatus_NotFound(t *testing.T) {
	_, _, _, outboxRepo, _, _, _, _ := setupTest(t)

	worker := service.NewOutboxRelayWorker(outboxRepo)
	err := worker.UpdateOutboxStatus(context.Background(), "msg-1", domain.OutboxStatusSENT)
	if err == nil {
		t.Fatal("Expected error msg not found, got nil")
	}
}

func TestOutboxRelayWorker_UpdateOutboxStatus_UpdateFails(t *testing.T) {
	_, _, _, outboxRepo, _, _, _, _ := setupTest(t)

	msg := &domain.TransactionalOutbox{
		ID:     "msg-1",
		Status: domain.OutboxStatusPENDING,
	}
	_ = outboxRepo.Create(context.Background(), msg)

	outboxRepo.updateError = errors.New("update error")

	worker := service.NewOutboxRelayWorker(outboxRepo)
	err := worker.UpdateOutboxStatus(context.Background(), "msg-1", domain.OutboxStatusSENT)
	if err == nil || err.Error() != "update error" {
		t.Fatalf("Expected 'update error', got %v", err)
	}
}

func TestReliableMessagingService_IsEventProcessed_Success(t *testing.T) {
	db, _, _, _, _, _, _, inboxRepo := setupTest(t)

	entry := &domain.KafkaEventInbox{
		EventID:          "evt-1",
		ProcessingStatus: domain.EventProcessingStatusSUCCESS,
	}
	_ = inboxRepo.Create(context.Background(), entry)

	svc := service.NewReliableMessagingService(db, inboxRepo)
	processed, err := svc.IsEventProcessed(context.Background(), "evt-1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !processed {
		t.Error("Expected processed true, got false")
	}
}

func TestReliableMessagingService_IsEventProcessed_Failed(t *testing.T) {
	db, _, _, _, _, _, _, inboxRepo := setupTest(t)

	entry := &domain.KafkaEventInbox{
		EventID:          "evt-1",
		ProcessingStatus: domain.EventProcessingStatusFAILED,
	}
	_ = inboxRepo.Create(context.Background(), entry)

	svc := service.NewReliableMessagingService(db, inboxRepo)
	processed, err := svc.IsEventProcessed(context.Background(), "evt-1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if processed {
		t.Error("Expected processed false, got true")
	}
}

func TestReliableMessagingService_IsEventProcessed_NotFound(t *testing.T) {
	db, _, _, _, _, _, _, inboxRepo := setupTest(t)

	svc := service.NewReliableMessagingService(db, inboxRepo)
	processed, err := svc.IsEventProcessed(context.Background(), "evt-not-found")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if processed {
		t.Error("Expected processed false, got true")
	}
}

func TestReliableMessagingService_IsEventProcessed_Error(t *testing.T) {
	db, _, _, _, _, _, _, inboxRepo := setupTest(t)

	inboxRepo.getByIDError = errors.New("db error")

	svc := service.NewReliableMessagingService(db, inboxRepo)
	processed, err := svc.IsEventProcessed(context.Background(), "evt-1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if processed {
		t.Error("Expected processed false, got true")
	}
}

func TestReliableMessagingService_ExecuteIdempotentTransaction_AlreadyProcessed(t *testing.T) {
	db, _, _, _, _, _, _, inboxRepo := setupTest(t)

	entry := &domain.KafkaEventInbox{
		EventID:          "evt-1",
		ProcessingStatus: domain.EventProcessingStatusSUCCESS,
	}
	_ = inboxRepo.Create(context.Background(), entry)

	svc := service.NewReliableMessagingService(db, inboxRepo)
	called := false
	err := svc.ExecuteIdempotentTransaction(context.Background(), "evt-1", "topic.type", nil, func(ctx context.Context) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if called {
		t.Error("Expected business routine NOT to be called")
	}
}

func TestReliableMessagingService_ExecuteIdempotentTransaction_IsEventProcessedError(t *testing.T) {
	db, _, _, _, _, _, _, inboxRepo := setupTest(t)

	inboxRepo.getByIDError = errors.New("db error")

	svc := service.NewReliableMessagingService(db, inboxRepo)
	called := false
	err := svc.ExecuteIdempotentTransaction(context.Background(), "evt-1", "topic.type", "payload", func(ctx context.Context) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !called {
		t.Error("Expected business routine to be called")
	}
}

func TestReliableMessagingService_ExecuteIdempotentTransaction_BusinessRoutineFails(t *testing.T) {
	db, _, _, _, _, _, _, inboxRepo := setupTest(t)

	svc := service.NewReliableMessagingService(db, inboxRepo)
	err := svc.ExecuteIdempotentTransaction(context.Background(), "evt-1", "topic.type", "payload", func(ctx context.Context) error {
		return errors.New("business logic error")
	})
	if err == nil || err.Error() != "business logic error" {
		t.Fatalf("Expected 'business logic error', got %v", err)
	}

	// Verify inbox entry was created with FAILED status
	entry, err := inboxRepo.GetByID(context.Background(), "evt-1")
	if err != nil {
		t.Fatalf("Expected to find inbox entry, got: %v", err)
	}
	if entry.ProcessingStatus != domain.EventProcessingStatusFAILED {
		t.Errorf("Expected status FAILED, got %s", entry.ProcessingStatus)
	}
}

func TestReliableMessagingService_ExecuteIdempotentTransaction_Success(t *testing.T) {
	db, _, _, _, _, _, _, inboxRepo := setupTest(t)

	svc := service.NewReliableMessagingService(db, inboxRepo)
	called := false
	err := svc.ExecuteIdempotentTransaction(context.Background(), "evt-1", "topic.type", "payload", func(ctx context.Context) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !called {
		t.Error("Expected business routine to be called")
	}

	// Verify inbox entry was created with SUCCESS status
	entry, err := inboxRepo.GetByID(context.Background(), "evt-1")
	if err != nil {
		t.Fatalf("Expected to find inbox entry, got: %v", err)
	}
	if entry.ProcessingStatus != domain.EventProcessingStatusSUCCESS {
		t.Errorf("Expected status SUCCESS, got %s", entry.ProcessingStatus)
	}
}

func TestReliableMessagingService_ExecuteIdempotentTransaction_Success_InboxCreateFails(t *testing.T) {
	db, _, _, _, _, _, _, inboxRepo := setupTest(t)

	inboxRepo.createError = errors.New("inbox create error")

	svc := service.NewReliableMessagingService(db, inboxRepo)
	err := svc.ExecuteIdempotentTransaction(context.Background(), "evt-1", "topic.type", "payload", func(ctx context.Context) error {
		return nil
	})
	if err == nil || err.Error() != "inbox create error" {
		t.Fatalf("Expected 'inbox create error', got %v", err)
	}
}

func TestGetDB(t *testing.T) {
	defaultDB, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	
	// Case 1: txKey not in context
	ctx := context.Background()
	db := service.GetDB(ctx, defaultDB)
	if db == nil {
		t.Error("Expected non-nil db")
	}

	// Case 2: txKey in context
	txDB := defaultDB.Session(&gorm.Session{})
	ctxWithTx := context.WithValue(ctx, "gorm_tx", txDB)
	dbWithTx := service.GetDB(ctxWithTx, defaultDB)
	if dbWithTx == nil {
		t.Error("Expected non-nil db from context")
	}
}
