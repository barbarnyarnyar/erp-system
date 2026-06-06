package domain

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

type EventPublisher interface {
	Publish(ctx context.Context, topic string, key string, payload interface{}) error
}

type PurchaseOrderCreatedEvent struct {
	PurchaseOrderID string          `json:"purchase_order_id"`
	PONumber        string          `json:"po_number"`
	SupplierID      string          `json:"supplier_id"`
	TotalAmount     decimal.Decimal `json:"total_amount"`
	Timestamp       time.Time       `json:"timestamp"`
}

type InventoryValuedEvent struct {
	InventoryItemID string          `json:"inventory_item_id"`
	ProductID       string          `json:"product_id"`
	LocationID      string          `json:"location_id"`
	QuantityOnHand  int             `json:"quantity_on_hand"`
	UnitCost        decimal.Decimal `json:"unit_cost"`
	TotalValuation  decimal.Decimal `json:"total_valuation"`
	Timestamp       time.Time       `json:"timestamp"`
}

type SCMTrainingRequiredEvent struct {
	DepartmentID string    `json:"department_id"`
	Topic        string    `json:"topic"`
	Deadline     time.Time `json:"deadline"`
	Timestamp    time.Time `json:"timestamp"`
}

type ProductCreatedEvent struct {
	ProductID   string    `json:"product_id"`
	ProductCode string    `json:"product_code"`
	ProductName string    `json:"product_name"`
	ProductType string    `json:"product_type"`
	Timestamp   time.Time `json:"timestamp"`
}

type ProductUpdatedEvent struct {
	ProductID   string    `json:"product_id"`
	ProductCode string    `json:"product_code"`
	ProductName string    `json:"product_name"`
	IsActive    bool      `json:"is_active"`
	Timestamp   time.Time `json:"timestamp"`
}

type ProductDiscontinuedEvent struct {
	ProductID string    `json:"product_id"`
	Timestamp time.Time `json:"timestamp"`
}

type InventoryReceivedEvent struct {
	InventoryItemID string    `json:"inventory_item_id"`
	ProductID       string    `json:"product_id"`
	LocationID      string    `json:"location_id"`
	Quantity        int       `json:"quantity"`
	Timestamp       time.Time `json:"timestamp"`
}

type InventoryShippedEvent struct {
	InventoryItemID string    `json:"inventory_item_id"`
	ProductID       string    `json:"product_id"`
	LocationID      string    `json:"location_id"`
	Quantity        int       `json:"quantity"`
	Timestamp       time.Time `json:"timestamp"`
}

type InventoryAdjustedEvent struct {
	InventoryItemID string          `json:"inventory_item_id"`
	ProductID       string          `json:"product_id"`
	LocationID      string          `json:"location_id"`
	QuantityChange  int             `json:"quantity_change"`
	NewQuantity     int             `json:"new_quantity"`
	Reason          string          `json:"reason"`
	Timestamp       time.Time       `json:"timestamp"`
}

type InventoryLowStockEvent struct {
	ProductID      string    `json:"product_id"`
	LocationID     string    `json:"location_id"`
	QuantityOnHand int       `json:"quantity_on_hand"`
	ReorderPoint   int       `json:"reorder_point"`
	Timestamp      time.Time `json:"timestamp"`
}

type InventoryOutOfStockEvent struct {
	ProductID  string    `json:"product_id"`
	LocationID string    `json:"location_id"`
	Timestamp  time.Time `json:"timestamp"`
}

type PurchaseOrderSentEvent struct {
	PurchaseOrderID string    `json:"purchase_order_id"`
	PONumber        string    `json:"po_number"`
	SupplierID      string    `json:"supplier_id"`
	Timestamp       time.Time `json:"timestamp"`
}

type PurchaseOrderReceivedEvent struct {
	PurchaseOrderID string    `json:"purchase_order_id"`
	PONumber        string    `json:"po_number"`
	ReceivedDate    time.Time `json:"received_date"`
	Timestamp       time.Time `json:"timestamp"`
}

type PurchaseOrderCancelledEvent struct {
	PurchaseOrderID string    `json:"purchase_order_id"`
	PONumber        string    `json:"po_number"`
	Reason          string    `json:"reason"`
	Timestamp       time.Time `json:"timestamp"`
}

type VendorCreatedEvent struct {
	VendorID   string    `json:"vendor_id"`
	VendorCode string    `json:"vendor_code"`
	VendorName string    `json:"vendor_name"`
	Timestamp  time.Time `json:"timestamp"`
}

type VendorUpdatedEvent struct {
	VendorID   string    `json:"vendor_id"`
	VendorCode string    `json:"vendor_code"`
	VendorName string    `json:"vendor_name"`
	IsActive   bool      `json:"is_active"`
	Timestamp  time.Time `json:"timestamp"`
}

type VendorPerformanceEvaluatedEvent struct {
	VendorID       string          `json:"vendor_id"`
	CompletionRate decimal.Decimal `json:"completion_rate"`
	TotalSpend     decimal.Decimal `json:"total_spend"`
	Score          decimal.Decimal `json:"score"`
	Timestamp      time.Time       `json:"timestamp"`
}

type ShipmentCreatedEvent struct {
	ShipmentID     string    `json:"shipment_id"`
	ShipmentNumber string    `json:"shipment_number"`
	Carrier        string    `json:"carrier"`
	TrackingNumber string    `json:"tracking_number"`
	Timestamp      time.Time `json:"timestamp"`
}

type ShipmentDispatchedEvent struct {
	ShipmentID   string    `json:"shipment_id"`
	DispatchedAt time.Time `json:"dispatched_at"`
	Timestamp    time.Time `json:"timestamp"`
}

type ShipmentDeliveredEvent struct {
	ShipmentID  string    `json:"shipment_id"`
	DeliveredAt time.Time `json:"delivered_at"`
	Timestamp   time.Time `json:"timestamp"`
}

type ShipmentDelayedEvent struct {
	ShipmentID       string    `json:"shipment_id"`
	NewEstimatedDeliv time.Time `json:"new_estimated_delivery"`
	Reason           string    `json:"reason"`
	Timestamp        time.Time `json:"timestamp"`
}

type MaterialDeliveredEvent struct {
	ProjectID    string    `json:"project_id"`
	TaskID       string    `json:"task_id"`
	ShipmentID   string    `json:"shipment_id"`
	DeliveryDate time.Time `json:"delivery_date"`
	Timestamp    time.Time `json:"timestamp"`
}

// Consumer Events payloads

type SalesOrderCreatedEvent struct {
	SalesOrderID string    `json:"sales_order_id"`
	OrderNumber  string    `json:"order_number"`
	CustomerID   string    `json:"customer_id"`
	Timestamp    time.Time `json:"timestamp"`
}

type CustomerDemandForecastEvent struct {
	ProductID        string          `json:"product_id"`
	ForecastDate     time.Time       `json:"forecast_date"`
	ForecastQuantity int             `json:"forecast_quantity"`
	ConfidenceLevel  decimal.Decimal `json:"confidence_level"`
	Timestamp        time.Time `json:"timestamp"`
}

type MaterialRequiredEvent struct {
	MaterialID string    `json:"material_id"`
	Quantity   int       `json:"quantity"`
	RequiredBy time.Time `json:"required_by"`
	Timestamp  time.Time `json:"timestamp"`
}

type ProductionCompletedEvent struct {
	ProductionOrderID string    `json:"production_order_id"`
	ProductID         string    `json:"product_id"`
	QuantityProduced  int       `json:"quantity_produced"`
	Timestamp         time.Time `json:"timestamp"`
}

type VendorPaymentProcessedEvent struct {
	PaymentID  string          `json:"payment_id"`
	VendorID   string          `json:"vendor_id"`
	AmountPaid decimal.Decimal `json:"amount_paid"`
	Status     string          `json:"status"`
	Timestamp  time.Time       `json:"timestamp"`
}

type MaterialRequestedEvent struct {
	ProjectID   string    `json:"project_id"`
	TaskID      string    `json:"task_id"`
	ProductID   string    `json:"product_id"`
	QtyRequired int       `json:"qty_required"`
	Timestamp   time.Time `json:"timestamp"`
}

type MaterialConsumedEvent struct {
	ProductionOrderID string          `json:"production_order_id"`
	ProductID         string          `json:"product_id"`
	Quantity          decimal.Decimal `json:"quantity"`
	Timestamp         time.Time       `json:"timestamp"`
}

