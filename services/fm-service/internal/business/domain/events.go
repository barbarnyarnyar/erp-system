package domain

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// EventPublisher defines the interface for publishing domain events to Kafka
type EventPublisher interface {
	Publish(ctx context.Context, topic string, key string, payload interface{}) error
}

// -----------------------------------------------------------------
// PUBLISHED EVENTS PAYLOADS
// -----------------------------------------------------------------

type InvoiceEventPayload struct {
	ID            string          `json:"id"`
	CustomerID    string          `json:"customer_id"`
	InvoiceNumber string          `json:"invoice_number"`
	TotalAmount   decimal.Decimal `json:"total_amount"`
	Status        string          `json:"status"`
	Timestamp     time.Time       `json:"timestamp"`
}

type PaymentEventPayload struct {
	ID            string          `json:"id"`
	InvoiceID     *string         `json:"invoice_id,omitempty"`
	BillID        *string         `json:"bill_id,omitempty"`
	PaymentNumber string          `json:"payment_number"`
	Amount        decimal.Decimal `json:"amount"`
	PaymentMethod string          `json:"payment_method"`
	Status        string          `json:"status"`
	Timestamp     time.Time       `json:"timestamp"`
}

type BudgetEventPayload struct {
	AccountID       string          `json:"account_id"`
	CostCenterID    *string         `json:"cost_center_id,omitempty"`
	FiscalYear      int             `json:"fiscal_year"`
	Period          int             `json:"period"`
	AllocatedAmount decimal.Decimal `json:"allocated_amount"`
	SpentAmount     decimal.Decimal `json:"spent_amount"`
	Timestamp       time.Time       `json:"timestamp"`
}

type AccountEventPayload struct {
	ID            string          `json:"id"`
	AccountNumber string          `json:"account_number"`
	Name          string          `json:"name"`
	Type          string          `json:"type"`
	Balance       decimal.Decimal `json:"balance"`
	Currency      string          `json:"currency"`
	Timestamp     time.Time       `json:"timestamp"`
}

type VendorEventPayload struct {
	ID         string    `json:"id"`
	VendorCode string    `json:"vendor_code"`
	VendorName string    `json:"vendor_name"`
	Email      string    `json:"email"`
	Timestamp  time.Time `json:"timestamp"`
}

type VendorBillEventPayload struct {
	ID         string          `json:"id"`
	VendorID   string          `json:"vendor_id"`
	BillNumber string          `json:"bill_number"`
	Amount     decimal.Decimal `json:"amount"`
	DueDate    time.Time       `json:"due_date"`
	Timestamp  time.Time       `json:"timestamp"`
}

type BudgetApprovedEvent struct {
	ProjectID    string          `json:"project_id"`
	TotalBudget  decimal.Decimal `json:"total_budget"`
	ApprovedDate time.Time       `json:"approved_date"`
	Timestamp    time.Time       `json:"timestamp"`
}

// -----------------------------------------------------------------
// CONSUMED EVENTS PAYLOADS
// -----------------------------------------------------------------

// EmployeeCreatedEvent from HR
type EmployeeCreatedEvent struct {
	EmployeeID   string          `json:"employee_id"`
	FirstName    string          `json:"first_name"`
	LastName     string          `json:"last_name"`
	DepartmentID string          `json:"department_id"`
	Salary       decimal.Decimal `json:"salary"`
	Timestamp    time.Time       `json:"timestamp"`
}

// PurchaseOrderCreatedEvent from SCM
type PurchaseOrderCreatedEvent struct {
	PurchaseOrderID string          `json:"purchase_order_id"`
	PONumber        string          `json:"po_number"`
	SupplierID      string          `json:"supplier_id"`
	TotalAmount     decimal.Decimal `json:"total_amount"`
	Timestamp       time.Time       `json:"timestamp"`
}

// SalesOrderConfirmedEvent from CRM
type SalesOrderConfirmedEvent struct {
	CustomerID   string          `json:"customer_id"`
	SalesOrderID string          `json:"sales_order_id"`
	TotalAmount  decimal.Decimal `json:"total_amount"`
	Timestamp    time.Time       `json:"timestamp"`
}

// MaterialConsumedEvent from Manufacturing
type MaterialConsumedEvent struct {
	ProductionOrderID string          `json:"production_order_id"`
	ProductID         string          `json:"product_id"`
	Quantity          int             `json:"quantity"`
	UnitCost          decimal.Decimal `json:"unit_cost"`
	TotalCost         decimal.Decimal `json:"total_cost"`
	Timestamp         time.Time       `json:"timestamp"`
}

// PayrollProcessedEvent from HR
type PayrollProcessedEvent struct {
	PayrollID   string          `json:"payroll_id"`
	PeriodStart time.Time       `json:"period_start"`
	PeriodEnd   time.Time       `json:"period_end"`
	TotalGross  decimal.Decimal `json:"total_gross"`
	TotalNet    decimal.Decimal `json:"total_net"`
	Timestamp   time.Time       `json:"timestamp"`
}

// ExpenseSubmittedEvent from HR
type ExpenseSubmittedEvent struct {
	ExpenseID   string          `json:"expense_id"`
	EmployeeID  string          `json:"employee_id"`
	Description string          `json:"description"`
	Amount      decimal.Decimal `json:"amount"`
	Timestamp   time.Time       `json:"timestamp"`
}

// InvoiceReceivedEvent from SCM
type InvoiceReceivedEvent struct {
	VendorID    string          `json:"vendor_id"`
	InvoiceNo   string          `json:"invoice_no"`
	POID        string          `json:"po_id"`
	TotalAmount decimal.Decimal `json:"total_amount"`
	DueDate     time.Time       `json:"due_date"`
	Timestamp   time.Time       `json:"timestamp"`
}

// InventoryValuedEvent from SCM
type InventoryValuedEvent struct {
	LocationID    string          `json:"location_id"`
	ValuationDate time.Time       `json:"valuation_date"`
	TotalValue    decimal.Decimal `json:"total_value"`
	Timestamp     time.Time       `json:"timestamp"`
}

// CustomerCreatedEvent from CRM
type CustomerCreatedEvent struct {
	CustomerID   string    `json:"customer_id"`
	CustomerName string    `json:"customer_name"`
	Email        string    `json:"email"`
	Timestamp    time.Time `json:"timestamp"`
}

// ProductionCompletedEvent from Manufacturing
type ProductionCompletedEvent struct {
	ProductionOrderID string          `json:"production_order_id"`
	ProductID         string          `json:"product_id"`
	QuantityCompleted int             `json:"quantity_completed"`
	TotalValuation    decimal.Decimal `json:"total_valuation"`
	Timestamp         time.Time       `json:"timestamp"`
}

// ProjectCreatedEvent from Project Module
type ProjectCreatedEvent struct {
	ProjectID   string    `json:"project_id"`
	ProjectName string    `json:"project_name"`
	ManagerID   string    `json:"manager_id"`
	Timestamp   time.Time `json:"timestamp"`
}

// TimeLoggedEvent from Project Module
type TimeLoggedEvent struct {
	TimeLogID    string          `json:"time_log_id"`
	ProjectID    string          `json:"project_id"`
	EmployeeID   string          `json:"employee_id"`
	HoursLogged  decimal.Decimal `json:"hours_logged"`
	BillableRate decimal.Decimal `json:"billable_rate"`
	Timestamp    time.Time       `json:"timestamp"`
}

// ProjectExpenseIncurredEvent from Project Module
type ProjectExpenseIncurredEvent struct {
	ExpenseID   string          `json:"expense_id"`
	ProjectID   string          `json:"project_id"`
	Description string          `json:"description"`
	Amount      decimal.Decimal `json:"amount"`
	Timestamp   time.Time       `json:"timestamp"`
}
