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
// Published Event Payloads (PM Module)
// ============================================================================

// Project Events
type ProjectCreatedEvent struct {
	ProjectID   string    `json:"project_id"`
	ProjectName string    `json:"project_name"`
	ManagerID   string    `json:"manager_id"`
	Timestamp   time.Time `json:"timestamp"`
}

type ProjectUpdatedEvent struct {
	ProjectID string    `json:"project_id"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

type ProjectStartedEvent struct {
	ProjectID string    `json:"project_id"`
	Timestamp time.Time `json:"timestamp"`
}

type ProjectCompletedEvent struct {
	ProjectID string    `json:"project_id"`
	Timestamp time.Time `json:"timestamp"`
}

type ProjectCancelledEvent struct {
	ProjectID string    `json:"project_id"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

type ProjectDelayedEvent struct {
	ProjectID string    `json:"project_id"`
	DelayDays int       `json:"delay_days"`
	Timestamp time.Time `json:"timestamp"`
}

// Task Events
type TaskCreatedEvent struct {
	TaskID    string    `json:"task_id"`
	ProjectID string    `json:"project_id"`
	Title     string    `json:"title"`
	Timestamp time.Time `json:"timestamp"`
}

type TaskAssignedEvent struct {
	TaskID      string    `json:"task_id"`
	ProjectID   string    `json:"project_id"`
	EmployeeID  string    `json:"employee_id"`
	Workload    int       `json:"workload"` // expected workload in hours
	Timestamp   time.Time `json:"timestamp"`
}

type TaskStartedEvent struct {
	TaskID    string    `json:"task_id"`
	ProjectID string    `json:"project_id"`
	Timestamp time.Time `json:"timestamp"`
}

type TaskCompletedEvent struct {
	TaskID    string    `json:"task_id"`
	ProjectID string    `json:"project_id"`
	Timestamp time.Time `json:"timestamp"`
}

type TaskOverdueEvent struct {
	TaskID    string    `json:"task_id"`
	ProjectID string    `json:"project_id"`
	DueDate   time.Time `json:"due_date"`
	Timestamp time.Time `json:"timestamp"`
}

// Resource Events
type ResourceAllocatedEvent struct {
	AllocationID string    `json:"allocation_id"`
	ProjectID    string    `json:"project_id"`
	UserID       string    `json:"user_id"`
	Role         string    `json:"role"`
	Timestamp    time.Time `json:"timestamp"`
}

type ResourceReleasedEvent struct {
	AllocationID string    `json:"allocation_id"`
	ProjectID    string    `json:"project_id"`
	UserID       string    `json:"user_id"`
	Timestamp    time.Time `json:"timestamp"`
}

type ResourceOverallocatedEvent struct {
	UserID        string    `json:"user_id"`
	ProjectID     string    `json:"project_id"`
	TotalCapacity int       `json:"total_capacity"`
	Timestamp     time.Time `json:"timestamp"`
}

// Time Events
type TimeLoggedEvent struct {
	TimeLogID    string          `json:"time_log_id"`
	ProjectID    string          `json:"project_id"`
	EmployeeID   string          `json:"employee_id"`
	HoursLogged  decimal.Decimal `json:"hours_logged"`
	BillableRate decimal.Decimal `json:"billable_rate"`
	Timestamp    time.Time       `json:"timestamp"`
}

type TimeApprovedEvent struct {
	TimeLogID  string    `json:"time_log_id"`
	ApprovedBy string    `json:"approved_by"`
	Timestamp  time.Time `json:"timestamp"`
}

type TimeRejectedEvent struct {
	TimeLogID  string    `json:"time_log_id"`
	RejectedBy string    `json:"rejected_by"`
	Reason     string    `json:"reason"`
	Timestamp  time.Time `json:"timestamp"`
}

// Expense Events
type ExpenseSubmittedEvent struct {
	ExpenseID string          `json:"expense_id"`
	ProjectID string          `json:"project_id"`
	UserID    string          `json:"user_id"`
	Amount    decimal.Decimal `json:"amount"`
	Currency  string          `json:"currency"`
	Timestamp time.Time       `json:"timestamp"`
}

type ExpenseApprovedEvent struct {
	ExpenseID  string    `json:"expense_id"`
	ApprovedBy string    `json:"approved_by"`
	Timestamp  time.Time `json:"timestamp"`
}

type ExpenseRejectedEvent struct {
	ExpenseID  string    `json:"expense_id"`
	RejectedBy string    `json:"rejected_by"`
	Reason     string    `json:"reason"`
	Timestamp  time.Time `json:"timestamp"`
}

// Milestone Events
type MilestoneAchievedEvent struct {
	ProjectID   string    `json:"project_id"`
	MilestoneID string    `json:"milestone_id"`
	Name        string    `json:"name"`
	Timestamp   time.Time `json:"timestamp"`
}

type MilestoneDelayedEvent struct {
	ProjectID   string    `json:"project_id"`
	MilestoneID string    `json:"milestone_id"`
	Name        string    `json:"name"`
	TargetDate  time.Time `json:"target_date"`
	Timestamp   time.Time `json:"timestamp"`
}

// ============================================================================
// Consumed Event Payloads
// ============================================================================

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

type BudgetApprovedEvent struct {
	ProjectID    string          `json:"project_id"`
	TotalBudget  decimal.Decimal `json:"total_budget"`
	ApprovedDate time.Time       `json:"approved_date"`
	Timestamp    time.Time       `json:"timestamp"`
}

type PaymentReceivedEvent struct {
	ProjectID   string          `json:"project_id"`
	InvoiceID   string          `json:"invoice_id"`
	AmountPaid  decimal.Decimal `json:"amount_paid"`
	Timestamp   time.Time       `json:"timestamp"`
}

type SalesOrderReceivedEvent struct {
	SalesOrderID string          `json:"sales_order_id"`
	CustomerID   string          `json:"customer_id"`
	TotalAmount  decimal.Decimal `json:"total_amount"`
	Timestamp    time.Time       `json:"timestamp"`
}

type MaterialDeliveredEvent struct {
	ProjectID    string    `json:"project_id"`
	TaskID       string    `json:"task_id"`
	ShipmentID   string    `json:"shipment_id"`
	DeliveryDate time.Time `json:"delivery_date"`
	Timestamp    time.Time `json:"timestamp"`
}

type CustomProductionCompletedEvent struct {
	ProjectID         string    `json:"project_id"`
	ProductionOrderID string    `json:"production_order_id"`
	CustomItemID      string    `json:"custom_item_id"`
	Quantity          int       `json:"quantity"`
	Timestamp         time.Time `json:"timestamp"`
}

// ============================================================================
// Integration Compatibility Events (Required by other ERP services)
// ============================================================================

type PrjCustomOrderCreatedEvent struct {
	ProjectID    string    `json:"project_id"`
	CustomItemID string    `json:"custom_item_id"`
	Quantity     int       `json:"quantity"`
	RequiredBy   time.Time `json:"required_by"`
	Timestamp    time.Time `json:"timestamp"`
}

type MaterialRequestedEvent struct {
	ProjectID   string    `json:"project_id"`
	TaskID      string    `json:"task_id"`
	ProductID   string    `json:"product_id"`
	QtyRequired int       `json:"qty_required"`
	Timestamp   time.Time `json:"timestamp"`
}

type ProjectExpenseIncurredEvent struct {
	ExpenseID   string          `json:"expense_id"`
	ProjectID   string          `json:"project_id"`
	Description string          `json:"description"`
	Amount      decimal.Decimal `json:"amount"`
	Timestamp   time.Time       `json:"timestamp"`
}

