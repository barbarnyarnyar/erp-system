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
// PRODUCER EVENTS PAYLOADS (hr.cdd)
// ============================================================================

type EmployeeCreatedEvent struct {
	EventID        string          `json:"event_id"`
	LegalEntityID  string          `json:"legal_entity_id"`
	EmployeeID     string          `json:"employee_id"`
	ManagerHrID    *string         `json:"manager_hr_id,omitempty"`
	EmployeeNumber string          `json:"employee_number"`
	Email          string          `json:"email"`
	BaseSalary     decimal.Decimal `json:"base_salary"`
	Type           string          `json:"type"`
	Timestamp      time.Time       `json:"timestamp"`
}

type EmployeeTerminatedEvent struct {
	EventID        string    `json:"event_id"`
	LegalEntityID  string    `json:"legal_entity_id"`
	EmployeeID     string    `json:"employee_id"`
	EmployeeNumber string    `json:"employee_number"`
	Email          string    `json:"email"`
	Timestamp      time.Time `json:"timestamp"`
}

type PayrollProcessedEvent struct {
	EventID       string          `json:"event_id"`
	LegalEntityID string          `json:"legal_entity_id"`
	PayrollRunID  string          `json:"payroll_run_id"`
	FiscalYear    int             `json:"fiscal_year"`
	PeriodNumber  int             `json:"period_number"`
	TotalNetPay   decimal.Decimal `json:"total_net_pay"`
	TotalGrossPay decimal.Decimal `json:"total_gross_pay"`
	Timestamp     time.Time       `json:"timestamp"`
}

type ExpenseApprovedEvent struct {
	EventID       string          `json:"event_id"`
	LegalEntityID string          `json:"legal_entity_id"`
	ClaimID       string          `json:"claim_id"`
	EmployeeID    string          `json:"employee_id"`
	TotalAmount   decimal.Decimal `json:"total_amount"`
	CostCenter    string          `json:"cost_center"`
	Timestamp     time.Time       `json:"timestamp"`
}

// ============================================================================
// CONSUMER EVENTS PAYLOADS (hr.cdd)
// ============================================================================

type TimeLogPayload struct {
	TimeLogID   string          `json:"time_log_id"`
	WbsNodeID   string          `json:"wbs_node_id"`
	EmployeeID  string          `json:"employee_id"`
	HoursSpent  decimal.Decimal `json:"hours_spent"`
	BillingRate decimal.Decimal `json:"billing_rate"`
}

type PrjTimeLoggedEvent struct {
	EventID               string           `json:"event_id"`
	LegalEntityID         string           `json:"legal_entity_id"`
	ProjectID             string           `json:"project_id"`
	CustomerID            string           `json:"customer_id"`
	TotalAccumulatedHours decimal.Decimal  `json:"total_accumulated_hours"`
	Details               []TimeLogPayload `json:"details"`
	Timestamp             time.Time        `json:"timestamp"`
}

type FmVendorPaidEvent struct {
	EventID         string    `json:"event_id"`
	BillID          string    `json:"bill_id"`
	LegalEntityID   string    `json:"legal_entity_id"`
	TargetDocumentID string    `json:"target_document_id"`
	Timestamp       time.Time `json:"timestamp"`
}
