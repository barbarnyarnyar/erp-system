package domain

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// EventPublisher defines the publisher interface for event streaming
type EventPublisher interface {
	Publish(ctx context.Context, topic string, key string, payload interface{}) error
}

// ==========================================
// PRODUCER EVENTS (EMITTED BY MFG)
// ==========================================

// MfgProductionStartedEvent (mfg.production.started)
type MfgProductionStartedEvent struct {
	EventID       string    `json:"event_id"`
	LegalEntityID string    `json:"legal_entity_id"`
	WorkOrderID   string    `json:"work_order_id"`
	MaterialID    string    `json:"material_id"`
	Timestamp     time.Time `json:"timestamp"`
}

// MfgMaterialConsumedEvent (mfg.material.consumed)
type MfgMaterialConsumedEvent struct {
	EventID       string                `json:"event_id"`
	LegalEntityID string                `json:"legal_entity_id"`
	WorkOrderID   string                `json:"work_order_id"`
	Items         []ConsumedItemPayload `json:"items"`
	Timestamp     time.Time             `json:"timestamp"`
}

// MfgYieldProducedEvent (mfg.yield.produced)
type MfgYieldProducedEvent struct {
	EventID          string          `json:"event_id"`
	LegalEntityID    string          `json:"legal_entity_id"`
	WorkOrderID      string          `json:"work_order_id"`
	RoutingStationID string          `json:"routing_station_id"`
	QuantityGood     decimal.Decimal `json:"quantity_good"`
	QuantityScrap    decimal.Decimal `json:"quantity_scrap"`
	OperatorHrID     string          `json:"operator_hr_id"`
	Timestamp        time.Time       `json:"timestamp"`
}

// MfgWorkOrderCompletedEvent (mfg.work_order.completed)
type MfgWorkOrderCompletedEvent struct {
	EventID          string          `json:"event_id"`
	LegalEntityID    string          `json:"legal_entity_id"`
	WorkOrderID      string          `json:"work_order_id"`
	MaterialID       string          `json:"material_id"`
	QuantityProduced decimal.Decimal `json:"quantity_produced"`
	Timestamp        time.Time       `json:"timestamp"`
}

// ==========================================
// CONSUMER EVENTS (RECEIVED BY MFG)
// ==========================================

// PlmBomReleasedEvent (plm.bom.released)
type PlmBomReleasedEvent struct {
	EventID       string      `json:"event_id"`
	LegalEntityID string      `json:"legal_entity_id"`
	BomHeaderID   string      `json:"bom_header_id"`
	MaterialID    string      `json:"material_id"`
	VersionString string      `json:"version_string"`
	Components    interface{} `json:"components"` // JSONB payload
	Timestamp     time.Time   `json:"timestamp"`
}

// QmsInspectionPassedEvent (qms.inspection.passed)
type QmsInspectionPassedEvent struct {
	EventID          string    `json:"event_id"`
	LegalEntityID    string    `json:"legal_entity_id"`
	InspectionID     string    `json:"inspection_id"`
	TriggerSource    string    `json:"trigger_source"`
	SourceDocumentID string    `json:"source_document_id"`
	MaterialID       string    `json:"material_id"`
	Timestamp        time.Time `json:"timestamp"`
}

// QmsInspectionFailedEvent (qms.inspection.failed)
type QmsInspectionFailedEvent struct {
	EventID          string    `json:"event_id"`
	LegalEntityID    string    `json:"legal_entity_id"`
	InspectionID     string    `json:"inspection_id"`
	TriggerSource    string    `json:"trigger_source"`
	SourceDocumentID string    `json:"source_document_id"`
	MaterialID       string    `json:"material_id"`
	NonConformanceID string    `json:"non_conformance_id"`
	Timestamp        time.Time `json:"timestamp"`
}

// EamMachineOfflineEvent (eam.machine.offline)
type EamMachineOfflineEvent struct {
	EventID       string    `json:"event_id"`
	LegalEntityID string    `json:"legal_entity_id"`
	EquipmentID   string    `json:"equipment_id"`
	WorkOrderID   string    `json:"work_order_id"`
	Priority      string    `json:"priority"`
	Timestamp     time.Time `json:"timestamp"`
}
