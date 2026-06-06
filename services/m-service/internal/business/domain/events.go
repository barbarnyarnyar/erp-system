package domain

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

type EventPublisher interface {
	Publish(ctx context.Context, topic string, key string, payload interface{}) error
}

type ProductionScheduledEvent struct {
	ProductionOrderID string    `json:"production_order_id"`
	ProductID         string    `json:"product_id"`
	Quantity          int       `json:"quantity"`
	ScheduledDate     time.Time `json:"scheduled_date"`
	Timestamp         time.Time `json:"timestamp"`
}

type ProductionCompletedEvent struct {
	ProductionOrderID string    `json:"production_order_id"`
	ProductID         string    `json:"product_id"`
	Quantity          int       `json:"quantity"`
	Timestamp         time.Time `json:"timestamp"`
}

type MaterialConsumedEvent struct {
	ProductionOrderID string          `json:"production_order_id"`
	ProductID         string          `json:"product_id"`
	Quantity          decimal.Decimal `json:"quantity"`
	Timestamp         time.Time       `json:"timestamp"`
}

type MaterialRequiredEvent struct {
	ProductID  string          `json:"product_id"`
	Quantity   decimal.Decimal `json:"quantity"`
	RequiredBy time.Time       `json:"required_by"`
	Timestamp  time.Time       `json:"timestamp"`
}

type SalesOrderCreatedEvent struct {
	SalesOrderID string    `json:"sales_order_id"`
	CustomerID   string    `json:"customer_id"`
	ProductID    string    `json:"product_id"`
	Quantity     int       `json:"quantity"`
	Timestamp    time.Time `json:"timestamp"`
}

type DemandForecastEvent struct {
	ProductID        string    `json:"product_id"`
	ForecastQuantity int       `json:"forecast_quantity"`
	Period           string    `json:"period"`
	Timestamp        time.Time `json:"timestamp"`
}

type ProductionStartedEvent struct {
	ProductionOrderID string    `json:"production_order_id"`
	ProductID         string    `json:"product_id"`
	Timestamp         time.Time `json:"timestamp"`
}

type ProductionDelayedEvent struct {
	ProductionOrderID string    `json:"production_order_id"`
	Reason            string    `json:"reason"`
	NewScheduledDate  time.Time `json:"new_scheduled_date"`
	Timestamp         time.Time `json:"timestamp"`
}

type WorkOrderCreatedEvent struct {
	WorkOrderID       string    `json:"work_order_id"`
	ProductionOrderID string    `json:"production_order_id"`
	WorkCenterID      string    `json:"work_center_id"`
	Timestamp         time.Time `json:"timestamp"`
}

type WorkOrderStartedEvent struct {
	WorkOrderID string    `json:"work_order_id"`
	Timestamp   time.Time `json:"timestamp"`
}

type WorkOrderCompletedEvent struct {
	WorkOrderID string    `json:"work_order_id"`
	Timestamp   time.Time `json:"timestamp"`
}

type WorkOrderCancelledEvent struct {
	WorkOrderID string    `json:"work_order_id"`
	Reason      string    `json:"reason"`
	Timestamp   time.Time `json:"timestamp"`
}

type MaterialWastedEvent struct {
	ProductionOrderID string          `json:"production_order_id"`
	ProductID         string          `json:"product_id"`
	Quantity          decimal.Decimal `json:"quantity"`
	Reason            string          `json:"reason"`
	Timestamp         time.Time       `json:"timestamp"`
}

type QualityInspectionPassedEvent struct {
	InspectionID string    `json:"inspection_id"`
	WorkOrderID  string    `json:"work_order_id"`
	InspectorID  string    `json:"inspector_id"`
	Timestamp    time.Time `json:"timestamp"`
}

type QualityInspectionFailedEvent struct {
	InspectionID string    `json:"inspection_id"`
	WorkOrderID  string    `json:"work_order_id"`
	InspectorID  string    `json:"inspector_id"`
	Remarks      string    `json:"remarks"`
	Timestamp    time.Time `json:"timestamp"`
}

type QualityNonConformanceDetectedEvent struct {
	NonConformanceID string    `json:"non_conformance_id"`
	InspectionID     string    `json:"inspection_id"`
	Severity         string    `json:"severity"`
	Description      string    `json:"description"`
	Timestamp        time.Time `json:"timestamp"`
}

type MaintenanceScheduledEvent struct {
	MaintenanceOrderID string    `json:"maintenance_order_id"`
	EquipmentID        string    `json:"equipment_id"`
	ScheduledDate      time.Time `json:"scheduled_date"`
	Timestamp          time.Time `json:"timestamp"`
}

type MaintenanceCompletedEvent struct {
	MaintenanceOrderID string    `json:"maintenance_order_id"`
	EquipmentID        string    `json:"equipment_id"`
	Timestamp          time.Time `json:"timestamp"`
}

type EquipmentDownEvent struct {
	EquipmentID  string    `json:"equipment_id"`
	WorkCenterID string    `json:"work_center_id"`
	Reason       string    `json:"reason"`
	Timestamp    time.Time `json:"timestamp"`
}

type EquipmentUpEvent struct {
	EquipmentID  string    `json:"equipment_id"`
	WorkCenterID string    `json:"work_center_id"`
	Timestamp    time.Time `json:"timestamp"`
}

type CustomProductionCompletedEvent struct {
	ProjectID         string    `json:"project_id"`
	ProductionOrderID string    `json:"production_order_id"`
	CustomItemID      string    `json:"custom_item_id"`
	Quantity          int       `json:"quantity"`
	Timestamp         time.Time `json:"timestamp"`
}

type SCMMaterialReceivedEvent struct {
	PurchaseOrderID string          `json:"purchase_order_id"`
	ProductID       string          `json:"product_id"`
	Quantity        decimal.Decimal `json:"quantity"`
	Timestamp       time.Time       `json:"timestamp"`
}

type SCMInventoryUpdatedEvent struct {
	ProductID        string          `json:"product_id"`
	LocationID       string          `json:"location_id"`
	QuantityOnHand   decimal.Decimal `json:"quantity_on_hand"`
	ChangeType       string          `json:"change_type"` // e.g. ADJUSTED, RECEIVED, SHIPPED
	Timestamp        time.Time       `json:"timestamp"`
}

type FinCostBudgetAllocatedEvent struct {
	DepartmentID string          `json:"department_id"`
	ProjectID    string          `json:"project_id"`
	Amount       decimal.Decimal `json:"amount"`
	Timestamp    time.Time       `json:"timestamp"`
}

type HREmployeeScheduledEvent struct {
	EmployeeID   string    `json:"employee_id"`
	WorkCenterID string    `json:"work_center_id"`
	ShiftStart   time.Time `json:"shift_start"`
	ShiftEnd     time.Time `json:"shift_end"`
	Timestamp    time.Time `json:"timestamp"`
}

type PrjCustomOrderCreatedEvent struct {
	ProjectID    string    `json:"project_id"`
	CustomItemID string    `json:"custom_item_id"`
	Quantity     int       `json:"quantity"`
	RequiredBy   time.Time `json:"required_by"`
	Timestamp    time.Time `json:"timestamp"`
}
