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
