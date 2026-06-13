package memory

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/erp-system/hr-service/internal/business/domain"
)

// MemoryDepartmentRepo implements domain.DepartmentRepository
type MemoryDepartmentRepo struct {
	mu    sync.RWMutex
	depts map[string]domain.Department
}

func NewMemoryDepartmentRepo() *MemoryDepartmentRepo {
	return &MemoryDepartmentRepo{depts: make(map[string]domain.Department)}
}

func (r *MemoryDepartmentRepo) Create(ctx context.Context, dept *domain.Department) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.depts[dept.ID] = *dept
	return nil
}

func (r *MemoryDepartmentRepo) GetByID(ctx context.Context, id string) (*domain.Department, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	dept, ok := r.depts[id]
	if !ok {
		return nil, errors.New("department not found")
	}
	return &dept, nil
}

func (r *MemoryDepartmentRepo) GetByCode(ctx context.Context, legalEntityID, code string) (*domain.Department, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, d := range r.depts {
		if d.LegalEntityID == legalEntityID && d.DepartmentCode == code {
			return &d, nil
		}
	}
	return nil, errors.New("department not found by code")
}

func (r *MemoryDepartmentRepo) List(ctx context.Context) ([]domain.Department, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.Department, 0, len(r.depts))
	for _, d := range r.depts {
		list = append(list, d)
	}
	return list, nil
}

func (r *MemoryDepartmentRepo) Update(ctx context.Context, dept *domain.Department) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.depts[dept.ID] = *dept
	return nil
}

// MemoryEmployeeMasterRepo implements domain.EmployeeMasterRepository
type MemoryEmployeeMasterRepo struct {
	mu   sync.RWMutex
	emps map[string]domain.EmployeeMaster
}

func NewMemoryEmployeeMasterRepo() *MemoryEmployeeMasterRepo {
	return &MemoryEmployeeMasterRepo{emps: make(map[string]domain.EmployeeMaster)}
}

func (r *MemoryEmployeeMasterRepo) Create(ctx context.Context, emp *domain.EmployeeMaster) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.emps[emp.ID] = *emp
	return nil
}

func (r *MemoryEmployeeMasterRepo) GetByID(ctx context.Context, id string) (*domain.EmployeeMaster, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	emp, ok := r.emps[id]
	if !ok || emp.DeletedAt != nil {
		return nil, errors.New("employee not found")
	}
	return &emp, nil
}

func (r *MemoryEmployeeMasterRepo) GetByNumber(ctx context.Context, legalEntityID, number string) (*domain.EmployeeMaster, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, e := range r.emps {
		if e.DeletedAt == nil && e.LegalEntityID == legalEntityID && e.EmployeeNumber == number {
			return &e, nil
		}
	}
	return nil, errors.New("employee not found by number")
}

func (r *MemoryEmployeeMasterRepo) GetByEmail(ctx context.Context, email string) (*domain.EmployeeMaster, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, e := range r.emps {
		if e.DeletedAt == nil && e.Email == email {
			return &e, nil
		}
	}
	return nil, errors.New("employee not found by email")
}

func (r *MemoryEmployeeMasterRepo) List(ctx context.Context) ([]domain.EmployeeMaster, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.EmployeeMaster, 0)
	for _, e := range r.emps {
		if e.DeletedAt == nil {
			list = append(list, e)
		}
	}
	return list, nil
}

func (r *MemoryEmployeeMasterRepo) Update(ctx context.Context, emp *domain.EmployeeMaster) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.emps[emp.ID] = *emp
	return nil
}

func (r *MemoryEmployeeMasterRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.emps[id]; ok {
		// standard delete
		delete(r.emps, id)
	}
	return nil
}

// MemoryPayrollRunRepo implements domain.PayrollRunRepository
type MemoryPayrollRunRepo struct {
	mu   sync.RWMutex
	runs map[string]domain.PayrollRun
}

func NewMemoryPayrollRunRepo() *MemoryPayrollRunRepo {
	return &MemoryPayrollRunRepo{runs: make(map[string]domain.PayrollRun)}
}

func (r *MemoryPayrollRunRepo) Create(ctx context.Context, run *domain.PayrollRun) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.runs[run.ID] = *run
	return nil
}

func (r *MemoryPayrollRunRepo) GetByID(ctx context.Context, id string) (*domain.PayrollRun, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	run, ok := r.runs[id]
	if !ok {
		return nil, errors.New("payroll run not found")
	}
	return &run, nil
}

func (r *MemoryPayrollRunRepo) GetByPeriod(ctx context.Context, legalEntityID string, year, period int) (*domain.PayrollRun, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, run := range r.runs {
		if run.LegalEntityID == legalEntityID && run.FiscalYear == year && run.PeriodNumber == period {
			return &run, nil
		}
	}
	return nil, errors.New("payroll run not found by period")
}

func (r *MemoryPayrollRunRepo) List(ctx context.Context) ([]domain.PayrollRun, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.PayrollRun, 0, len(r.runs))
	for _, run := range r.runs {
		list = append(list, run)
	}
	return list, nil
}

func (r *MemoryPayrollRunRepo) Update(ctx context.Context, run *domain.PayrollRun) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.runs[run.ID] = *run
	return nil
}

// MemoryExpenseClaimRepo implements domain.ExpenseClaimRepository
type MemoryExpenseClaimRepo struct {
	mu     sync.RWMutex
	claims map[string]domain.ExpenseClaim
}

func NewMemoryExpenseClaimRepo() *MemoryExpenseClaimRepo {
	return &MemoryExpenseClaimRepo{claims: make(map[string]domain.ExpenseClaim)}
}

func (r *MemoryExpenseClaimRepo) Create(ctx context.Context, claim *domain.ExpenseClaim) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.claims[claim.ID] = *claim
	return nil
}

func (r *MemoryExpenseClaimRepo) GetByID(ctx context.Context, id string) (*domain.ExpenseClaim, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	claim, ok := r.claims[id]
	if !ok {
		return nil, errors.New("expense claim not found")
	}
	return &claim, nil
}

func (r *MemoryExpenseClaimRepo) GetByNumber(ctx context.Context, legalEntityID, number string) (*domain.ExpenseClaim, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, claim := range r.claims {
		if claim.LegalEntityID == legalEntityID && claim.ClaimNumber == number {
			return &claim, nil
		}
	}
	return nil, errors.New("expense claim not found by number")
}

func (r *MemoryExpenseClaimRepo) List(ctx context.Context) ([]domain.ExpenseClaim, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.ExpenseClaim, 0, len(r.claims))
	for _, claim := range r.claims {
		list = append(list, claim)
	}
	return list, nil
}

func (r *MemoryExpenseClaimRepo) Update(ctx context.Context, claim *domain.ExpenseClaim) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.claims[claim.ID] = *claim
	return nil
}

// MemoryExpenseClaimLineRepo implements domain.ExpenseClaimLineRepository
type MemoryExpenseClaimLineRepo struct {
	mu    sync.RWMutex
	lines map[string]domain.ExpenseClaimLine
}

func NewMemoryExpenseClaimLineRepo() *MemoryExpenseClaimLineRepo {
	return &MemoryExpenseClaimLineRepo{lines: make(map[string]domain.ExpenseClaimLine)}
}

func (r *MemoryExpenseClaimLineRepo) Create(ctx context.Context, line *domain.ExpenseClaimLine) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lines[line.ID] = *line
	return nil
}

func (r *MemoryExpenseClaimLineRepo) GetByID(ctx context.Context, id string) (*domain.ExpenseClaimLine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	line, ok := r.lines[id]
	if !ok {
		return nil, errors.New("expense claim line not found")
	}
	return &line, nil
}

func (r *MemoryExpenseClaimLineRepo) ListByClaimID(ctx context.Context, claimID string) ([]domain.ExpenseClaimLine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.ExpenseClaimLine, 0)
	for _, line := range r.lines {
		if line.ExpenseClaimID == claimID {
			list = append(list, line)
		}
	}
	return list, nil
}

// MemoryTransactionalOutboxRepo implements domain.TransactionalOutboxRepository
type MemoryTransactionalOutboxRepo struct {
	mu   sync.RWMutex
	msgs map[string]domain.TransactionalOutbox
}

func NewMemoryTransactionalOutboxRepo() *MemoryTransactionalOutboxRepo {
	return &MemoryTransactionalOutboxRepo{msgs: make(map[string]domain.TransactionalOutbox)}
}

func (r *MemoryTransactionalOutboxRepo) Create(ctx context.Context, msg *domain.TransactionalOutbox) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.msgs[msg.ID] = *msg
	return nil
}

func (r *MemoryTransactionalOutboxRepo) GetByID(ctx context.Context, id string) (*domain.TransactionalOutbox, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	msg, ok := r.msgs[id]
	if !ok {
		return nil, errors.New("outbox message not found")
	}
	return &msg, nil
}

func (r *MemoryTransactionalOutboxRepo) GetUnsent(ctx context.Context, limit int) ([]domain.TransactionalOutbox, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	unsent := make([]domain.TransactionalOutbox, 0)
	for _, msg := range r.msgs {
		if msg.Status == domain.OutboxStatusPENDING {
			unsent = append(unsent, msg)
		}
	}
	sort.Slice(unsent, func(i, j int) bool {
		return unsent[i].CreatedAt.Before(unsent[j].CreatedAt)
	})
	if len(unsent) > limit {
		unsent = unsent[:limit]
	}
	return unsent, nil
}

func (r *MemoryTransactionalOutboxRepo) Update(ctx context.Context, msg *domain.TransactionalOutbox) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.msgs[msg.ID] = *msg
	return nil
}

// MemoryKafkaEventInboxRepo implements domain.KafkaEventInboxRepository
type MemoryKafkaEventInboxRepo struct {
	mu    sync.RWMutex
	inbox map[string]domain.KafkaEventInbox
}

func NewMemoryKafkaEventInboxRepo() *MemoryKafkaEventInboxRepo {
	return &MemoryKafkaEventInboxRepo{inbox: make(map[string]domain.KafkaEventInbox)}
}

func (r *MemoryKafkaEventInboxRepo) Create(ctx context.Context, msg *domain.KafkaEventInbox) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.inbox[msg.EventID] = *msg
	return nil
}

func (r *MemoryKafkaEventInboxRepo) GetByID(ctx context.Context, eventID string) (*domain.KafkaEventInbox, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	msg, ok := r.inbox[eventID]
	if !ok {
		return nil, errors.New("inbox message not found")
	}
	return &msg, nil
}

func (r *MemoryKafkaEventInboxRepo) Update(ctx context.Context, msg *domain.KafkaEventInbox) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.inbox[msg.EventID] = *msg
	return nil
}
