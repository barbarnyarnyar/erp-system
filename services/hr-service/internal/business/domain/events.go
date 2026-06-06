package domain

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

type EventPublisher interface {
	Publish(ctx context.Context, topic string, key string, payload interface{}) error
}

// ============================================================================
// PRODUCER EVENTS PAYLOADS
// ============================================================================

type EmployeeCreatedEvent struct {
	EmployeeID   string          `json:"employee_id"`
	FirstName    string          `json:"first_name"`
	LastName     string          `json:"last_name"`
	DepartmentID string          `json:"department_id"`
	Salary       decimal.Decimal `json:"salary"`
	Timestamp    time.Time       `json:"timestamp"`
}

type EmployeeUpdatedEvent struct {
	EmployeeID   string          `json:"employee_id"`
	FirstName    string          `json:"first_name"`
	LastName     string          `json:"last_name"`
	DepartmentID string          `json:"department_id"`
	PositionID   string          `json:"position_id"`
	Salary       decimal.Decimal `json:"salary"`
	Status       string          `json:"status"`
	Timestamp    time.Time       `json:"timestamp"`
}

type EmployeeTerminatedEvent struct {
	EmployeeID string    `json:"employee_id"`
	TermDate   time.Time `json:"term_date"`
	Reason     string    `json:"reason"`
	Timestamp  time.Time `json:"timestamp"`
}

type EmployeePromotedEvent struct {
	EmployeeID    string          `json:"employee_id"`
	OldPositionID string          `json:"old_position_id"`
	NewPositionID string          `json:"new_position_id"`
	NewSalary     decimal.Decimal `json:"new_salary"`
	Timestamp     time.Time       `json:"timestamp"`
}

type PayrollProcessedEvent struct {
	PayrollID   string          `json:"payroll_id"`
	EmployeeID  string          `json:"employee_id"`
	PeriodStart time.Time       `json:"period_start"`
	PeriodEnd   time.Time       `json:"period_end"`
	TotalGross  decimal.Decimal `json:"total_gross"`
	TotalNet    decimal.Decimal `json:"total_net"`
	Timestamp   time.Time       `json:"timestamp"`
}

type PayrollFailedEvent struct {
	EmployeeID  string    `json:"employee_id"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	Reason      string    `json:"reason"`
	Timestamp   time.Time `json:"timestamp"`
}

type SalaryChangedEvent struct {
	EmployeeID string          `json:"employee_id"`
	OldSalary  decimal.Decimal `json:"old_salary"`
	NewSalary  decimal.Decimal `json:"new_salary"`
	Timestamp  time.Time       `json:"timestamp"`
}

type TimesheetSubmittedEvent struct {
	TimesheetID string          `json:"timesheet_id"`
	EmployeeID  string          `json:"employee_id"`
	EntryDate   time.Time       `json:"entry_date"`
	TotalHours  decimal.Decimal `json:"total_hours"`
	Timestamp   time.Time       `json:"timestamp"`
}

type TimesheetApprovedEvent struct {
	TimesheetID string    `json:"timesheet_id"`
	EmployeeID  string    `json:"employee_id"`
	ApprovedBy  string    `json:"approved_by"`
	Timestamp   time.Time `json:"timestamp"`
}

type OvertimeRecordedEvent struct {
	EmployeeID    string          `json:"employee_id"`
	EntryDate     time.Time       `json:"entry_date"`
	OvertimeHours decimal.Decimal `json:"overtime_hours"`
	Timestamp     time.Time       `json:"timestamp"`
}

type LeaveRequestedEvent struct {
	LeaveRequestID string    `json:"leave_request_id"`
	EmployeeID     string    `json:"employee_id"`
	LeaveType      string    `json:"leave_type"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	Timestamp      time.Time `json:"timestamp"`
}

type LeaveApprovedEvent struct {
	LeaveRequestID string    `json:"leave_request_id"`
	EmployeeID     string    `json:"employee_id"`
	ApprovedBy     string    `json:"approved_by"`
	Timestamp      time.Time `json:"timestamp"`
}

type LeaveRejectedEvent struct {
	LeaveRequestID string    `json:"leave_request_id"`
	EmployeeID     string    `json:"employee_id"`
	RejectedBy     string    `json:"rejected_by"`
	Reason         string    `json:"reason"`
	Timestamp      time.Time `json:"timestamp"`
}

type TrainingCompletedEvent struct {
	TrainingProgramID string    `json:"training_program_id"`
	EmployeeID        string    `json:"employee_id"`
	CompletionDate    time.Time `json:"completion_date"`
	Timestamp         time.Time `json:"timestamp"`
}

type CertificationEarnedEvent struct {
	EmployeeID        string    `json:"employee_id"`
	CertificationName string    `json:"certification_name"`
	ExpiryDate        time.Time `json:"expiry_date"`
	Timestamp         time.Time `json:"timestamp"`
}

type SkillAcquiredEvent struct {
	EmployeeID string    `json:"employee_id"`
	SkillName  string    `json:"skill_name"`
	Proficiency string   `json:"proficiency"`
	Timestamp  time.Time `json:"timestamp"`
}

type PerformanceReviewCompletedEvent struct {
	ReviewID   string    `json:"review_id"`
	EmployeeID string    `json:"employee_id"`
	ReviewerID string    `json:"reviewer_id"`
	Rating     int       `json:"rating"`
	Timestamp  time.Time `json:"timestamp"`
}

type GoalAchievedEvent struct {
	EmployeeID  string    `json:"employee_id"`
	GoalTitle   string    `json:"goal_title"`
	AchievedAt  time.Time `json:"achieved_at"`
	Timestamp   time.Time `json:"timestamp"`
}

type PerformanceImprovementNeededEvent struct {
	EmployeeID  string    `json:"employee_id"`
	ReviewID    string    `json:"review_id"`
	Details     string    `json:"details"`
	Timestamp   time.Time `json:"timestamp"`
}

type ExpenseSubmittedEvent struct {
	ExpenseID   string          `json:"expense_id"`
	EmployeeID  string          `json:"employee_id"`
	Description string          `json:"description"`
	Amount      decimal.Decimal `json:"amount"`
	Timestamp   time.Time       `json:"timestamp"`
}

type EmployeeAvailableEvent struct {
	EmployeeID string    `json:"employee_id"`
	Status     string    `json:"status"` // e.g. AVAILABLE, BUSY
	Timestamp  time.Time `json:"timestamp"`
}

type EmployeeSkillsUpdatedEvent struct {
	EmployeeID string   `json:"employee_id"`
	Skills     []string `json:"skills"`
	Timestamp  time.Time `json:"timestamp"`
}

// ============================================================================
// CONSUMER EVENTS PAYLOADS
// ============================================================================

type ProjectCreatedEvent struct {
	ProjectID   string    `json:"project_id"`
	Name        string    `json:"name"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Timestamp   time.Time `json:"timestamp"`
}

type TaskAssignedEvent struct {
	TaskID      string    `json:"task_id"`
	ProjectID   string    `json:"project_id"`
	EmployeeID  string    `json:"employee_id"`
	Workload    int       `json:"workload"` // e.g. expected hours
	Timestamp   time.Time `json:"timestamp"`
}

type BudgetAllocatedEvent struct {
	DepartmentID string          `json:"department_id"`
	Amount       decimal.Decimal `json:"amount"`
	Period       string          `json:"period"`
	Timestamp    time.Time       `json:"timestamp"`
}

type ProductionScheduledEvent struct {
	ScheduleID  string    `json:"schedule_id"`
	Workstation string    `json:"workstation"`
	RequiredStaff int     `json:"required_staff"`
	StartDate   time.Time `json:"start_date"`
	Timestamp   time.Time `json:"timestamp"`
}

type SCMTrainingRequiredEvent struct {
	DepartmentID string    `json:"department_id"`
	Topic        string    `json:"topic"`
	Deadline     time.Time `json:"deadline"`
	Timestamp    time.Time `json:"timestamp"`
}
