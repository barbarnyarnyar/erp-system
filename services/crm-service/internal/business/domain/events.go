package domain

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

type EventPublisher interface {
	Publish(ctx context.Context, topic string, key string, payload interface{}) error
}

// Published Event Payloads

type CustomerCreatedEvent struct {
	CustomerID  string    `json:"customer_id"`
	CompanyName string    `json:"company_name"`
	ContactName string    `json:"contact_name"`
	Email       string    `json:"email"`
	Timestamp   time.Time `json:"timestamp"`
}

type CustomerUpdatedEvent struct {
	CustomerID  string    `json:"customer_id"`
	CompanyName string    `json:"company_name"`
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
}

type CustomerActivatedEvent struct {
	CustomerID string    `json:"customer_id"`
	Timestamp  time.Time `json:"timestamp"`
}

type CustomerDeactivatedEvent struct {
	CustomerID string    `json:"customer_id"`
	Timestamp  time.Time `json:"timestamp"`
}

type LeadCreatedEvent struct {
	LeadID    string    `json:"lead_id"`
	Company   string    `json:"company"`
	Email     string    `json:"email"`
	Timestamp time.Time `json:"timestamp"`
}

type LeadQualifiedEvent struct {
	LeadID    string    `json:"lead_id"`
	Score     int       `json:"score"`
	Timestamp time.Time `json:"timestamp"`
}

type LeadConvertedEvent struct {
	LeadID        string    `json:"lead_id"`
	CustomerID    string    `json:"customer_id"`
	OpportunityID string    `json:"opportunity_id"`
	Timestamp     time.Time `json:"timestamp"`
}

type LeadLostEvent struct {
	LeadID    string    `json:"lead_id"`
	Timestamp time.Time `json:"timestamp"`
}

type OpportunityCreatedEvent struct {
	OpportunityID string          `json:"opportunity_id"`
	CustomerID    string          `json:"customer_id"`
	Title         string          `json:"title"`
	Value         decimal.Decimal `json:"value"`
	Timestamp     time.Time       `json:"timestamp"`
}

type OpportunityUpdatedEvent struct {
	OpportunityID string          `json:"opportunity_id"`
	Status        string          `json:"status"`
	Stage         string          `json:"stage"`
	Value         decimal.Decimal `json:"value"`
	Timestamp     time.Time       `json:"timestamp"`
}

type OpportunityWonEvent struct {
	OpportunityID string          `json:"opportunity_id"`
	CustomerID    string          `json:"customer_id"`
	Value         decimal.Decimal `json:"value"`
	Timestamp     time.Time       `json:"timestamp"`
}

type OpportunityLostEvent struct {
	OpportunityID string    `json:"opportunity_id"`
	Timestamp     time.Time `json:"timestamp"`
}

type SalesOrderCreatedEvent struct {
	SalesOrderID string          `json:"sales_order_id"`
	CustomerID   string          `json:"customer_id"`
	TotalAmount  decimal.Decimal `json:"total_amount"`
	Timestamp    time.Time       `json:"timestamp"`
}

type SalesOrderUpdatedEvent struct {
	SalesOrderID string          `json:"sales_order_id"`
	Status       string          `json:"status"`
	TotalAmount  decimal.Decimal `json:"total_amount"`
	Timestamp    time.Time       `json:"timestamp"`
}

type SalesOrderConfirmedEvent struct {
	SalesOrderID string          `json:"sales_order_id"`
	CustomerID   string          `json:"customer_id"`
	TotalAmount  decimal.Decimal `json:"total_amount"`
	Timestamp    time.Time       `json:"timestamp"`
}

type SalesOrderCancelledEvent struct {
	SalesOrderID string    `json:"sales_order_id"`
	Reason       string    `json:"reason"`
	Timestamp    time.Time `json:"timestamp"`
}

type SalesOrderShippedEvent struct {
	SalesOrderID string    `json:"sales_order_id"`
	Timestamp    time.Time `json:"timestamp"`
}

type SalesOrderDeliveredEvent struct {
	SalesOrderID string    `json:"sales_order_id"`
	Timestamp    time.Time `json:"timestamp"`
}

type ServiceTicketCreatedEvent struct {
	TicketID   string    `json:"ticket_id"`
	CustomerID string    `json:"customer_id"`
	Title      string    `json:"title"`
	Priority   string    `json:"priority"`
	Timestamp  time.Time `json:"timestamp"`
}

type ServiceTicketUpdatedEvent struct {
	TicketID  string    `json:"ticket_id"`
	Status    string    `json:"status"`
	Priority  string    `json:"priority"`
	Timestamp time.Time `json:"timestamp"`
}

type ServiceTicketResolvedEvent struct {
	TicketID  string    `json:"ticket_id"`
	Timestamp time.Time `json:"timestamp"`
}

type ServiceTicketEscalatedEvent struct {
	TicketID  string    `json:"ticket_id"`
	Timestamp time.Time `json:"timestamp"`
}

type CampaignLaunchedEvent struct {
	CampaignID string    `json:"campaign_id"`
	Name       string    `json:"name"`
	Timestamp  time.Time `json:"timestamp"`
}

type CampaignCompletedEvent struct {
	CampaignID string    `json:"campaign_id"`
	Timestamp  time.Time `json:"timestamp"`
}

type EmailSentEvent struct {
	EmailID    string    `json:"email_id"`
	CampaignID string    `json:"campaign_id"`
	Recipient  string    `json:"recipient"`
	Timestamp  time.Time `json:"timestamp"`
}

type EmailOpenedEvent struct {
	EmailID   string    `json:"email_id"`
	Timestamp time.Time `json:"timestamp"`
}

type EmailClickedEvent struct {
	EmailID   string    `json:"email_id"`
	URL       string    `json:"url"`
	Timestamp time.Time `json:"timestamp"`
}

type SalesOrderReceivedEvent struct {
	SalesOrderID string          `json:"sales_order_id"`
	CustomerID   string          `json:"customer_id"`
	TotalAmount  decimal.Decimal `json:"total_amount"`
	Timestamp    time.Time       `json:"timestamp"`
}

type CustomerInteractionLoggedEvent struct {
	InteractionID   string    `json:"interaction_id"`
	CustomerID      string    `json:"customer_id"`
	Type            string    `json:"type"`
	Subject         string    `json:"subject"`
	InteractionDate time.Time `json:"interaction_date"`
	CreatedBy       string    `json:"created_by"`
	Timestamp       time.Time `json:"timestamp"`
}

// Consumed Event Payloads

type InventoryAvailableEvent struct {
	ProductID      string          `json:"product_id"`
	QuantityOnHand decimal.Decimal `json:"quantity_on_hand"`
	Timestamp      time.Time       `json:"timestamp"`
}

type ShipmentDeliveredEvent struct {
	ShipmentID   string    `json:"shipment_id"`
	SalesOrderID string    `json:"sales_order_id"`
	Timestamp    time.Time `json:"timestamp"`
}

type PaymentReceivedEvent struct {
	InvoiceID   string          `json:"invoice_id"`
	ReferenceID string          `json:"reference_id"`
	Amount      decimal.Decimal `json:"amount"`
	Timestamp   time.Time       `json:"timestamp"`
}

type CreditCheckCompletedEvent struct {
	CustomerID   string    `json:"customer_id"`
	CreditStatus string    `json:"credit_status"` // APPROVED, DENIED
	Timestamp    time.Time `json:"timestamp"`
}

type ProductionCompletedEvent struct {
	ProductionOrderID string    `json:"production_order_id"`
	ProductID         string    `json:"product_id"`
	Quantity          int       `json:"quantity"`
	Timestamp         time.Time `json:"timestamp"`
}

type ProjectCompletedEvent struct {
	ProjectID string    `json:"project_id"`
	Timestamp time.Time `json:"timestamp"`
}

type EmployeePerformanceEvent struct {
	EmployeeID string          `json:"employee_id"`
	Rating     decimal.Decimal `json:"rating"`
	Timestamp  time.Time       `json:"timestamp"`
}
