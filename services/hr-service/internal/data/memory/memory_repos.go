package memory

import (
	"context"
	"errors"
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

// MemoryPositionRepo implements domain.PositionRepository
type MemoryPositionRepo struct {
	mu   sync.RWMutex
	poss map[string]domain.Position
}

func NewMemoryPositionRepo() *MemoryPositionRepo {
	return &MemoryPositionRepo{poss: make(map[string]domain.Position)}
}

func (r *MemoryPositionRepo) Create(ctx context.Context, pos *domain.Position) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.poss[pos.ID] = *pos
	return nil
}

func (r *MemoryPositionRepo) GetByID(ctx context.Context, id string) (*domain.Position, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	pos, ok := r.poss[id]
	if !ok {
		return nil, errors.New("position not found")
	}
	return &pos, nil
}

func (r *MemoryPositionRepo) List(ctx context.Context) ([]domain.Position, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.Position, 0, len(r.poss))
	for _, p := range r.poss {
		list = append(list, p)
	}
	return list, nil
}

func (r *MemoryPositionRepo) Update(ctx context.Context, pos *domain.Position) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.poss[pos.ID] = *pos
	return nil
}

// MemoryEmployeeRepo implements domain.EmployeeRepository
type MemoryEmployeeRepo struct {
	mu   sync.RWMutex
	emps map[string]domain.Employee
}

func NewMemoryEmployeeRepo() *MemoryEmployeeRepo {
	return &MemoryEmployeeRepo{emps: make(map[string]domain.Employee)}
}

func (r *MemoryEmployeeRepo) Create(ctx context.Context, emp *domain.Employee) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.emps[emp.ID] = *emp
	return nil
}

func (r *MemoryEmployeeRepo) GetByID(ctx context.Context, id string) (*domain.Employee, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	emp, ok := r.emps[id]
	if !ok {
		return nil, errors.New("employee not found")
	}
	return &emp, nil
}

func (r *MemoryEmployeeRepo) List(ctx context.Context) ([]domain.Employee, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.Employee, 0, len(r.emps))
	for _, e := range r.emps {
		list = append(list, e)
	}
	return list, nil
}

func (r *MemoryEmployeeRepo) Update(ctx context.Context, emp *domain.Employee) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.emps[emp.ID] = *emp
	return nil
}

func (r *MemoryEmployeeRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.emps, id)
	return nil
}

// MemoryPayrollRecordRepo implements domain.PayrollRecordRepository
type MemoryPayrollRecordRepo struct {
	mu  sync.RWMutex
	prs map[string]domain.PayrollRecord
}

func NewMemoryPayrollRecordRepo() *MemoryPayrollRecordRepo {
	return &MemoryPayrollRecordRepo{prs: make(map[string]domain.PayrollRecord)}
}

func (r *MemoryPayrollRecordRepo) Create(ctx context.Context, pr *domain.PayrollRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.prs[pr.ID] = *pr
	return nil
}

func (r *MemoryPayrollRecordRepo) GetByID(ctx context.Context, id string) (*domain.PayrollRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	pr, ok := r.prs[id]
	if !ok {
		return nil, errors.New("payroll record not found")
	}
	return &pr, nil
}

func (r *MemoryPayrollRecordRepo) List(ctx context.Context) ([]domain.PayrollRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.PayrollRecord, 0, len(r.prs))
	for _, p := range r.prs {
		list = append(list, p)
	}
	return list, nil
}

func (r *MemoryPayrollRecordRepo) Update(ctx context.Context, pr *domain.PayrollRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.prs[pr.ID] = *pr
	return nil
}

func (r *MemoryPayrollRecordRepo) GetByEmployeeID(ctx context.Context, empID string) ([]domain.PayrollRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.PayrollRecord
	for _, pr := range r.prs {
		if pr.EmployeeID == empID {
			list = append(list, pr)
		}
	}
	return list, nil
}

// MemoryTimeEntryRepo implements domain.TimeEntryRepository
type MemoryTimeEntryRepo struct {
	mu  sync.RWMutex
	tes map[string]domain.TimeEntry
}

func NewMemoryTimeEntryRepo() *MemoryTimeEntryRepo {
	return &MemoryTimeEntryRepo{tes: make(map[string]domain.TimeEntry)}
}

func (r *MemoryTimeEntryRepo) Create(ctx context.Context, te *domain.TimeEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tes[te.ID] = *te
	return nil
}

func (r *MemoryTimeEntryRepo) GetByID(ctx context.Context, id string) (*domain.TimeEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	te, ok := r.tes[id]
	if !ok {
		return nil, errors.New("time entry not found")
	}
	return &te, nil
}

func (r *MemoryTimeEntryRepo) List(ctx context.Context) ([]domain.TimeEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.TimeEntry, 0, len(r.tes))
	for _, t := range r.tes {
		list = append(list, t)
	}
	return list, nil
}

func (r *MemoryTimeEntryRepo) Update(ctx context.Context, te *domain.TimeEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tes[te.ID] = *te
	return nil
}

// MemoryLeaveRequestRepo implements domain.LeaveRequestRepository
type MemoryLeaveRequestRepo struct {
	mu  sync.RWMutex
	lrs map[string]domain.LeaveRequest
}

func NewMemoryLeaveRequestRepo() *MemoryLeaveRequestRepo {
	return &MemoryLeaveRequestRepo{lrs: make(map[string]domain.LeaveRequest)}
}

func (r *MemoryLeaveRequestRepo) Create(ctx context.Context, lr *domain.LeaveRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lrs[lr.ID] = *lr
	return nil
}

func (r *MemoryLeaveRequestRepo) GetByID(ctx context.Context, id string) (*domain.LeaveRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	lr, ok := r.lrs[id]
	if !ok {
		return nil, errors.New("leave request not found")
	}
	return &lr, nil
}

func (r *MemoryLeaveRequestRepo) List(ctx context.Context) ([]domain.LeaveRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.LeaveRequest, 0, len(r.lrs))
	for _, l := range r.lrs {
		list = append(list, l)
	}
	return list, nil
}

func (r *MemoryLeaveRequestRepo) Update(ctx context.Context, lr *domain.LeaveRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lrs[lr.ID] = *lr
	return nil
}

// MemoryJobPostingRepo implements domain.JobPostingRepository
type MemoryJobPostingRepo struct {
	mu  sync.RWMutex
	jps map[string]domain.JobPosting
}

func NewMemoryJobPostingRepo() *MemoryJobPostingRepo {
	return &MemoryJobPostingRepo{jps: make(map[string]domain.JobPosting)}
}

func (r *MemoryJobPostingRepo) Create(ctx context.Context, jp *domain.JobPosting) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.jps[jp.ID] = *jp
	return nil
}

func (r *MemoryJobPostingRepo) GetByID(ctx context.Context, id string) (*domain.JobPosting, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	jp, ok := r.jps[id]
	if !ok {
		return nil, errors.New("job posting not found")
	}
	return &jp, nil
}

func (r *MemoryJobPostingRepo) List(ctx context.Context) ([]domain.JobPosting, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.JobPosting, 0, len(r.jps))
	for _, j := range r.jps {
		list = append(list, j)
	}
	return list, nil
}

func (r *MemoryJobPostingRepo) Update(ctx context.Context, jp *domain.JobPosting) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.jps[jp.ID] = *jp
	return nil
}

func (r *MemoryJobPostingRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.jps, id)
	return nil
}

// MemoryJobApplicationRepo implements domain.JobApplicationRepository
type MemoryJobApplicationRepo struct {
	mu  sync.RWMutex
	jas map[string]domain.JobApplication
}

func NewMemoryJobApplicationRepo() *MemoryJobApplicationRepo {
	return &MemoryJobApplicationRepo{jas: make(map[string]domain.JobApplication)}
}

func (r *MemoryJobApplicationRepo) Create(ctx context.Context, ja *domain.JobApplication) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.jas[ja.ID] = *ja
	return nil
}

func (r *MemoryJobApplicationRepo) GetByID(ctx context.Context, id string) (*domain.JobApplication, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ja, ok := r.jas[id]
	if !ok {
		return nil, errors.New("job application not found")
	}
	return &ja, nil
}

func (r *MemoryJobApplicationRepo) List(ctx context.Context) ([]domain.JobApplication, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.JobApplication, 0, len(r.jas))
	for _, j := range r.jas {
		list = append(list, j)
	}
	return list, nil
}

func (r *MemoryJobApplicationRepo) Update(ctx context.Context, ja *domain.JobApplication) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.jas[ja.ID] = *ja
	return nil
}

// MemoryPerformanceReviewRepo implements domain.PerformanceReviewRepository
type MemoryPerformanceReviewRepo struct {
	mu  sync.RWMutex
	prs map[string]domain.PerformanceReview
}

func NewMemoryPerformanceReviewRepo() *MemoryPerformanceReviewRepo {
	return &MemoryPerformanceReviewRepo{prs: make(map[string]domain.PerformanceReview)}
}

func (r *MemoryPerformanceReviewRepo) Create(ctx context.Context, pr *domain.PerformanceReview) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.prs[pr.ID] = *pr
	return nil
}

func (r *MemoryPerformanceReviewRepo) GetByID(ctx context.Context, id string) (*domain.PerformanceReview, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	pr, ok := r.prs[id]
	if !ok {
		return nil, errors.New("performance review not found")
	}
	return &pr, nil
}

func (r *MemoryPerformanceReviewRepo) List(ctx context.Context) ([]domain.PerformanceReview, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.PerformanceReview, 0, len(r.prs))
	for _, p := range r.prs {
		list = append(list, p)
	}
	return list, nil
}

func (r *MemoryPerformanceReviewRepo) Update(ctx context.Context, pr *domain.PerformanceReview) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.prs[pr.ID] = *pr
	return nil
}

// MemoryTrainingProgramRepo implements domain.TrainingProgramRepository
type MemoryTrainingProgramRepo struct {
	mu  sync.RWMutex
	tps map[string]domain.TrainingProgram
}

func NewMemoryTrainingProgramRepo() *MemoryTrainingProgramRepo {
	return &MemoryTrainingProgramRepo{tps: make(map[string]domain.TrainingProgram)}
}

func (r *MemoryTrainingProgramRepo) Create(ctx context.Context, tp *domain.TrainingProgram) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tps[tp.ID] = *tp
	return nil
}

func (r *MemoryTrainingProgramRepo) GetByID(ctx context.Context, id string) (*domain.TrainingProgram, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tp, ok := r.tps[id]
	if !ok {
		return nil, errors.New("training program not found")
	}
	return &tp, nil
}

func (r *MemoryTrainingProgramRepo) List(ctx context.Context) ([]domain.TrainingProgram, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.TrainingProgram, 0, len(r.tps))
	for _, t := range r.tps {
		list = append(list, t)
	}
	return list, nil
}

func (r *MemoryTrainingProgramRepo) Update(ctx context.Context, tp *domain.TrainingProgram) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tps[tp.ID] = *tp
	return nil
}

// MemoryEmployeeDocumentRepo implements domain.EmployeeDocumentRepository
type MemoryEmployeeDocumentRepo struct {
	mu   sync.RWMutex
	docs map[string]domain.EmployeeDocument
}

func NewMemoryEmployeeDocumentRepo() *MemoryEmployeeDocumentRepo {
	return &MemoryEmployeeDocumentRepo{docs: make(map[string]domain.EmployeeDocument)}
}

func (r *MemoryEmployeeDocumentRepo) Create(ctx context.Context, doc *domain.EmployeeDocument) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.docs[doc.ID] = *doc
	return nil
}

func (r *MemoryEmployeeDocumentRepo) GetByID(ctx context.Context, id string) (*domain.EmployeeDocument, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	doc, ok := r.docs[id]
	if !ok {
		return nil, errors.New("employee document not found")
	}
	return &doc, nil
}

func (r *MemoryEmployeeDocumentRepo) ListByEmployeeID(ctx context.Context, empID string) ([]domain.EmployeeDocument, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.EmployeeDocument
	for _, d := range r.docs {
		if d.EmployeeID == empID {
			list = append(list, d)
		}
	}
	return list, nil
}

func (r *MemoryEmployeeDocumentRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.docs, id)
	return nil
}

// MemoryPayrollDeductionRepo implements domain.PayrollDeductionRepository
type MemoryPayrollDeductionRepo struct {
	mu   sync.RWMutex
	deds map[string]domain.PayrollDeduction
}

func NewMemoryPayrollDeductionRepo() *MemoryPayrollDeductionRepo {
	return &MemoryPayrollDeductionRepo{deds: make(map[string]domain.PayrollDeduction)}
}

func (r *MemoryPayrollDeductionRepo) Create(ctx context.Context, pd *domain.PayrollDeduction) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.deds[pd.ID] = *pd
	return nil
}

func (r *MemoryPayrollDeductionRepo) GetByID(ctx context.Context, id string) (*domain.PayrollDeduction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	pd, ok := r.deds[id]
	if !ok {
		return nil, errors.New("payroll deduction not found")
	}
	return &pd, nil
}

func (r *MemoryPayrollDeductionRepo) ListByPayrollID(ctx context.Context, payrollID string) ([]domain.PayrollDeduction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.PayrollDeduction
	for _, pd := range r.deds {
		if pd.PayrollID == payrollID {
			list = append(list, pd)
		}
	}
	return list, nil
}

// MemoryLeaveBalanceRepo implements domain.LeaveBalanceRepository
type MemoryLeaveBalanceRepo struct {
	mu   sync.RWMutex
	bals map[string]domain.LeaveBalance
}

func NewMemoryLeaveBalanceRepo() *MemoryLeaveBalanceRepo {
	return &MemoryLeaveBalanceRepo{bals: make(map[string]domain.LeaveBalance)}
}

func (r *MemoryLeaveBalanceRepo) Create(ctx context.Context, lb *domain.LeaveBalance) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.bals[lb.ID] = *lb
	return nil
}

func (r *MemoryLeaveBalanceRepo) GetByID(ctx context.Context, id string) (*domain.LeaveBalance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	lb, ok := r.bals[id]
	if !ok {
		return nil, errors.New("leave balance not found")
	}
	return &lb, nil
}

func (r *MemoryLeaveBalanceRepo) GetByEmployeeAndTypeAndYear(ctx context.Context, empID string, leaveType string, year int) (*domain.LeaveBalance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, lb := range r.bals {
		if lb.EmployeeID == empID && lb.LeaveType == leaveType && lb.Year == year {
			return &lb, nil
		}
	}
	return nil, errors.New("leave balance not found")
}

func (r *MemoryLeaveBalanceRepo) GetByEmployeeID(ctx context.Context, empID string) ([]domain.LeaveBalance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.LeaveBalance
	for _, lb := range r.bals {
		if lb.EmployeeID == empID {
			list = append(list, lb)
		}
	}
	return list, nil
}

func (r *MemoryLeaveBalanceRepo) Update(ctx context.Context, lb *domain.LeaveBalance) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.bals[lb.ID] = *lb
	return nil
}

func (r *MemoryLeaveBalanceRepo) List(ctx context.Context) ([]domain.LeaveBalance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.LeaveBalance, 0, len(r.bals))
	for _, lb := range r.bals {
		list = append(list, lb)
	}
	return list, nil
}

// MemoryTrainingEnrollmentRepo implements domain.TrainingEnrollmentRepository
type MemoryTrainingEnrollmentRepo struct {
	mu     sync.RWMutex
	enrols map[string]domain.TrainingEnrollment
}

func NewMemoryTrainingEnrollmentRepo() *MemoryTrainingEnrollmentRepo {
	return &MemoryTrainingEnrollmentRepo{enrols: make(map[string]domain.TrainingEnrollment)}
}

func (r *MemoryTrainingEnrollmentRepo) Create(ctx context.Context, te *domain.TrainingEnrollment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.enrols[te.ID] = *te
	return nil
}

func (r *MemoryTrainingEnrollmentRepo) GetByID(ctx context.Context, id string) (*domain.TrainingEnrollment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	te, ok := r.enrols[id]
	if !ok {
		return nil, errors.New("training enrollment not found")
	}
	return &te, nil
}

func (r *MemoryTrainingEnrollmentRepo) GetByTrainingAndEmployee(ctx context.Context, trainingID string, empID string) (*domain.TrainingEnrollment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, te := range r.enrols {
		if te.TrainingID == trainingID && te.EmployeeID == empID {
			return &te, nil
		}
	}
	return nil, errors.New("training enrollment not found")
}

func (r *MemoryTrainingEnrollmentRepo) Update(ctx context.Context, te *domain.TrainingEnrollment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.enrols[te.ID] = *te
	return nil
}

func (r *MemoryTrainingEnrollmentRepo) List(ctx context.Context) ([]domain.TrainingEnrollment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.TrainingEnrollment, 0, len(r.enrols))
	for _, te := range r.enrols {
		list = append(list, te)
	}
	return list, nil
}

// MemoryExpenseClaimRepo implements domain.ExpenseClaimRepository
type MemoryExpenseClaimRepo struct {
	mu     sync.RWMutex
	claims map[string]domain.ExpenseClaim
}

func NewMemoryExpenseClaimRepo() *MemoryExpenseClaimRepo {
	return &MemoryExpenseClaimRepo{claims: make(map[string]domain.ExpenseClaim)}
}

func (r *MemoryExpenseClaimRepo) Create(ctx context.Context, ec *domain.ExpenseClaim) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.claims[ec.ID] = *ec
	return nil
}

func (r *MemoryExpenseClaimRepo) GetByID(ctx context.Context, id string) (*domain.ExpenseClaim, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ec, ok := r.claims[id]
	if !ok {
		return nil, errors.New("expense claim not found")
	}
	return &ec, nil
}

func (r *MemoryExpenseClaimRepo) List(ctx context.Context) ([]domain.ExpenseClaim, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.ExpenseClaim, 0, len(r.claims))
	for _, ec := range r.claims {
		list = append(list, ec)
	}
	return list, nil
}

func (r *MemoryExpenseClaimRepo) Update(ctx context.Context, ec *domain.ExpenseClaim) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.claims[ec.ID] = *ec
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

func (r *MemoryExpenseClaimLineRepo) Create(ctx context.Context, ecl *domain.ExpenseClaimLine) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lines[ecl.ID] = *ecl
	return nil
}

func (r *MemoryExpenseClaimLineRepo) ListByClaimID(ctx context.Context, claimID string) ([]domain.ExpenseClaimLine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.ExpenseClaimLine
	for _, ecl := range r.lines {
		if ecl.ClaimID == claimID {
			list = append(list, ecl)
		}
	}
	return list, nil
}


