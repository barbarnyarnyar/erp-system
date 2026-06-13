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
// PRODUCER EVENTS PAYLOADS (prj.cdd)
// ============================================================================

type PrjTimeLoggedEvent struct {
	EventID               string           `json:"event_id"`
	LegalEntityID         string           `json:"legal_entity_id"`
	ProjectID             string           `json:"project_id"`
	CustomerID            string           `json:"customer_id"`
	TotalAccumulatedHours decimal.Decimal  `json:"total_accumulated_hours"`
	Details               []TimeLogPayload `json:"details"`
	Timestamp             time.Time        `json:"timestamp"`
}

type PrjMilestoneAchievedEvent struct {
	EventID       string          `json:"event_id"`
	LegalEntityID string          `json:"legal_entity_id"`
	ProjectID     string          `json:"project_id"`
	CustomerID    string          `json:"customer_id"`
	WbsNodeID     string          `json:"wbs_node_id"`
	RevenueAmount decimal.Decimal `json:"revenue_amount"`
	Timestamp     time.Time       `json:"timestamp"`
}

// ============================================================================
// CONSUMER EVENTS PAYLOADS (prj.cdd)
// ============================================================================

type HrEmployeeCreatedEvent struct {
	EventID       string    `json:"event_id"`
	LegalEntityID string    `json:"legal_entity_id"`
	EmployeeID    string    `json:"employee_id"`
	ExplicitRole  string    `json:"explicit_role"`
	Timestamp     time.Time `json:"timestamp"`
}

type HrEmployeeTerminatedEvent struct {
	EventID       string    `json:"event_id"`
	LegalEntityID string    `json:"legal_entity_id"`
	EmployeeID    string    `json:"employee_id"`
	Timestamp     time.Time `json:"timestamp"`
}

type CrmSalesOrderConfirmedEvent struct {
	EventID          string      `json:"event_id"`
	LegalEntityID    string      `json:"legal_entity_id"`
	SalesOrderID     string      `json:"sales_order_id"`
	CustomerID       string      `json:"customer_id"`
	ContractMetadata interface{} `json:"contract_metadata"`
	Timestamp        time.Time   `json:"timestamp"`
}
