package domain

import "context"

type DepartmentRepository interface {
	Create(ctx context.Context, dept *Department) error
	GetByID(ctx context.Context, id string) (*Department, error)
	List(ctx context.Context) ([]Department, error)
	Update(ctx context.Context, dept *Department) error
}

type PositionRepository interface {
	Create(ctx context.Context, pos *Position) error
	GetByID(ctx context.Context, id string) (*Position, error)
	List(ctx context.Context) ([]Position, error)
	Update(ctx context.Context, pos *Position) error
}

type EmployeeRepository interface {
	Create(ctx context.Context, emp *Employee) error
	GetByID(ctx context.Context, id string) (*Employee, error)
	List(ctx context.Context) ([]Employee, error)
	Update(ctx context.Context, emp *Employee) error
	Delete(ctx context.Context, id string) error
}

type PayrollRecordRepository interface {
	Create(ctx context.Context, pr *PayrollRecord) error
	GetByID(ctx context.Context, id string) (*PayrollRecord, error)
	List(ctx context.Context) ([]PayrollRecord, error)
	Update(ctx context.Context, pr *PayrollRecord) error
	GetByEmployeeID(ctx context.Context, empID string) ([]PayrollRecord, error)
}

type TimeEntryRepository interface {
	Create(ctx context.Context, te *TimeEntry) error
	GetByID(ctx context.Context, id string) (*TimeEntry, error)
	List(ctx context.Context) ([]TimeEntry, error)
	Update(ctx context.Context, te *TimeEntry) error
}

type LeaveRequestRepository interface {
	Create(ctx context.Context, lr *LeaveRequest) error
	GetByID(ctx context.Context, id string) (*LeaveRequest, error)
	List(ctx context.Context) ([]LeaveRequest, error)
	Update(ctx context.Context, lr *LeaveRequest) error
}

type JobPostingRepository interface {
	Create(ctx context.Context, jp *JobPosting) error
	GetByID(ctx context.Context, id string) (*JobPosting, error)
	List(ctx context.Context) ([]JobPosting, error)
	Update(ctx context.Context, jp *JobPosting) error
	Delete(ctx context.Context, id string) error
}

type JobApplicationRepository interface {
	Create(ctx context.Context, ja *JobApplication) error
	GetByID(ctx context.Context, id string) (*JobApplication, error)
	List(ctx context.Context) ([]JobApplication, error)
	Update(ctx context.Context, ja *JobApplication) error
}

type PerformanceReviewRepository interface {
	Create(ctx context.Context, pr *PerformanceReview) error
	GetByID(ctx context.Context, id string) (*PerformanceReview, error)
	List(ctx context.Context) ([]PerformanceReview, error)
	Update(ctx context.Context, pr *PerformanceReview) error
}

type TrainingProgramRepository interface {
	Create(ctx context.Context, tp *TrainingProgram) error
	GetByID(ctx context.Context, id string) (*TrainingProgram, error)
	List(ctx context.Context) ([]TrainingProgram, error)
	Update(ctx context.Context, tp *TrainingProgram) error
}

type EmployeeDocumentRepository interface {
	Create(ctx context.Context, doc *EmployeeDocument) error
	GetByID(ctx context.Context, id string) (*EmployeeDocument, error)
	ListByEmployeeID(ctx context.Context, empID string) ([]EmployeeDocument, error)
	Delete(ctx context.Context, id string) error
}

type PayrollDeductionRepository interface {
	Create(ctx context.Context, pd *PayrollDeduction) error
	GetByID(ctx context.Context, id string) (*PayrollDeduction, error)
	ListByPayrollID(ctx context.Context, payrollID string) ([]PayrollDeduction, error)
}

type LeaveBalanceRepository interface {
	Create(ctx context.Context, lb *LeaveBalance) error
	GetByID(ctx context.Context, id string) (*LeaveBalance, error)
	GetByEmployeeAndTypeAndYear(ctx context.Context, empID string, leaveType string, year int) (*LeaveBalance, error)
	GetByEmployeeID(ctx context.Context, empID string) ([]LeaveBalance, error)
	Update(ctx context.Context, lb *LeaveBalance) error
	List(ctx context.Context) ([]LeaveBalance, error)
}

type TrainingEnrollmentRepository interface {
	Create(ctx context.Context, te *TrainingEnrollment) error
	GetByID(ctx context.Context, id string) (*TrainingEnrollment, error)
	GetByTrainingAndEmployee(ctx context.Context, trainingID string, empID string) (*TrainingEnrollment, error)
	Update(ctx context.Context, te *TrainingEnrollment) error
	List(ctx context.Context) ([]TrainingEnrollment, error)
}

type ExpenseClaimRepository interface {
	Create(ctx context.Context, ec *ExpenseClaim) error
	GetByID(ctx context.Context, id string) (*ExpenseClaim, error)
	List(ctx context.Context) ([]ExpenseClaim, error)
	Update(ctx context.Context, ec *ExpenseClaim) error
}

type ExpenseClaimLineRepository interface {
	Create(ctx context.Context, ecl *ExpenseClaimLine) error
	ListByClaimID(ctx context.Context, claimID string) ([]ExpenseClaimLine, error)
}


