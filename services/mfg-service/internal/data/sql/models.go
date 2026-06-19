package sql

import (
	"encoding/json"
	"time"

	"github.com/erp-system/m-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

// 1. WorkCenter
type WorkCenter struct {
	ID             string    `gorm:"primaryKey;type:varchar(255)"`
	LegalEntityID  string    `gorm:"type:varchar(255);not null;index"`
	WorkCenterCode string    `gorm:"type:varchar(100);not null;uniqueIndex:idx_mfg_wc_code_tenant"`
	Name           string    `gorm:"type:varchar(255);not null"`
	IsActive       bool      `gorm:"type:boolean;default:true"`
	CreatedAt      time.Time `gorm:"not null"`
	UpdatedAt      time.Time `gorm:"not null"`
}

func (WorkCenter) TableName() string {
	return "mfg_work_centers"
}

func ToWorkCenterDomain(w *WorkCenter) *domain.WorkCenter {
	if w == nil {
		return nil
	}
	return &domain.WorkCenter{
		ID:             w.ID,
		LegalEntityID:  w.LegalEntityID,
		WorkCenterCode: w.WorkCenterCode,
		Name:           w.Name,
		IsActive:       w.IsActive,
		CreatedAt:      w.CreatedAt,
		UpdatedAt:      w.UpdatedAt,
	}
}

func FromWorkCenterDomain(w *domain.WorkCenter) *WorkCenter {
	if w == nil {
		return nil
	}
	return &WorkCenter{
		ID:             w.ID,
		LegalEntityID:  w.LegalEntityID,
		WorkCenterCode: w.WorkCenterCode,
		Name:           w.Name,
		IsActive:       w.IsActive,
		CreatedAt:      w.CreatedAt,
		UpdatedAt:      w.UpdatedAt,
	}
}

// 2. RoutingStation
type RoutingStation struct {
	ID                    string    `gorm:"primaryKey;type:varchar(255)"`
	WorkCenterID          string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_mfg_station_wc_code"`
	RoutingCode           string    `gorm:"type:varchar(100);not null;uniqueIndex:idx_mfg_station_wc_code"`
	StationType           string    `gorm:"type:varchar(50);not null"`
	EquipmentID           *string   `gorm:"type:varchar(255)"`
	StandardSetupTimeMins int       `gorm:"type:integer;not null;default:0"`
	StandardRunTimeMins   int       `gorm:"type:integer;not null;default:0"`
	CreatedAt             time.Time `gorm:"not null"`
	UpdatedAt             time.Time `gorm:"not null"`
}

func (RoutingStation) TableName() string {
	return "mfg_routing_stations"
}

func ToRoutingStationDomain(r *RoutingStation) *domain.RoutingStation {
	if r == nil {
		return nil
	}
	return &domain.RoutingStation{
		ID:                    r.ID,
		WorkCenterID:          r.WorkCenterID,
		RoutingCode:           r.RoutingCode,
		StationType:           domain.StationType(r.StationType),
		EquipmentID:           r.EquipmentID,
		StandardSetupTimeMins: r.StandardSetupTimeMins,
		StandardRunTimeMins:   r.StandardRunTimeMins,
		CreatedAt:             r.CreatedAt,
		UpdatedAt:             r.UpdatedAt,
	}
}

func FromRoutingStationDomain(r *domain.RoutingStation) *RoutingStation {
	if r == nil {
		return nil
	}
	return &RoutingStation{
		ID:                    r.ID,
		WorkCenterID:          r.WorkCenterID,
		RoutingCode:           r.RoutingCode,
		StationType:           string(r.StationType),
		EquipmentID:           r.EquipmentID,
		StandardSetupTimeMins: r.StandardSetupTimeMins,
		StandardRunTimeMins:   r.StandardRunTimeMins,
		CreatedAt:             r.CreatedAt,
		UpdatedAt:             r.UpdatedAt,
	}
}

// 3. WorkOrder
type WorkOrder struct {
	ID               string          `gorm:"primaryKey;type:varchar(255)"`
	LegalEntityID    string          `gorm:"type:varchar(255);not null;uniqueIndex:idx_mfg_wo_num_tenant"`
	MaterialID       string          `gorm:"type:varchar(255);not null"`
	BomHeaderID      string          `gorm:"type:varchar(255);not null"`
	WorkOrderNumber  string          `gorm:"type:varchar(100);not null;uniqueIndex:idx_mfg_wo_num_tenant"`
	QuantityTarget   decimal.Decimal `gorm:"type:numeric(14,4);not null"`
	QuantityProduced decimal.Decimal `gorm:"type:numeric(14,4);not null;default:0"`
	Status           string          `gorm:"type:varchar(50);not null"`
	ScheduledStart   time.Time       `gorm:"type:date;not null"`
	ScheduledEnd     time.Time       `gorm:"type:date;not null"`
	Version          int             `gorm:"type:integer;not null;default:1"`
	CreatedAt        time.Time       `gorm:"not null"`
	UpdatedAt        time.Time       `gorm:"not null"`
}

func (WorkOrder) TableName() string {
	return "mfg_work_orders"
}

func ToWorkOrderDomain(w *WorkOrder) *domain.WorkOrder {
	if w == nil {
		return nil
	}
	return &domain.WorkOrder{
		ID:               w.ID,
		LegalEntityID:    w.LegalEntityID,
		MaterialID:       w.MaterialID,
		BomHeaderID:      w.BomHeaderID,
		WorkOrderNumber:  w.WorkOrderNumber,
		QuantityTarget:   w.QuantityTarget,
		QuantityProduced: w.QuantityProduced,
		Status:           domain.WorkOrderState(w.Status),
		ScheduledStart:   w.ScheduledStart,
		ScheduledEnd:     w.ScheduledEnd,
		Version:          w.Version,
		CreatedAt:        w.CreatedAt,
		UpdatedAt:        w.UpdatedAt,
	}
}

func FromWorkOrderDomain(w *domain.WorkOrder) *WorkOrder {
	if w == nil {
		return nil
	}
	return &WorkOrder{
		ID:               w.ID,
		LegalEntityID:    w.LegalEntityID,
		MaterialID:       w.MaterialID,
		BomHeaderID:      w.BomHeaderID,
		WorkOrderNumber:  w.WorkOrderNumber,
		QuantityTarget:   w.QuantityTarget,
		QuantityProduced: w.QuantityProduced,
		Status:           string(w.Status),
		ScheduledStart:   w.ScheduledStart,
		ScheduledEnd:     w.ScheduledEnd,
		Version:          w.Version,
		CreatedAt:        w.CreatedAt,
		UpdatedAt:        w.UpdatedAt,
	}
}

// 4. WorkOrderRoutingState
type WorkOrderRoutingState struct {
	ID                     string     `gorm:"primaryKey;type:varchar(255)"`
	WorkOrderID            string     `gorm:"type:varchar(255);not null;uniqueIndex:idx_mfg_wo_routing_state"`
	CurrentStationID       string     `gorm:"type:varchar(255);not null;uniqueIndex:idx_mfg_wo_routing_state"`
	NextSuggestedStationID *string    `gorm:"type:varchar(255)"`
	IsReworkLoop           bool       `gorm:"type:boolean;not null;default:false"`
	EnteredAt              time.Time  `gorm:"not null"`
	ExitedAt               *time.Time `gorm:"default:null"`
}

func (WorkOrderRoutingState) TableName() string {
	return "mfg_work_order_routing_states"
}

func ToWorkOrderRoutingStateDomain(s *WorkOrderRoutingState) *domain.WorkOrderRoutingState {
	if s == nil {
		return nil
	}
	return &domain.WorkOrderRoutingState{
		ID:                     s.ID,
		WorkOrderID:            s.WorkOrderID,
		CurrentStationID:       s.CurrentStationID,
		NextSuggestedStationID: s.NextSuggestedStationID,
		IsReworkLoop:           s.IsReworkLoop,
		EnteredAt:              s.EnteredAt,
		ExitedAt:               s.ExitedAt,
	}
}

func FromWorkOrderRoutingStateDomain(s *domain.WorkOrderRoutingState) *WorkOrderRoutingState {
	if s == nil {
		return nil
	}
	return &WorkOrderRoutingState{
		ID:                     s.ID,
		WorkOrderID:            s.WorkOrderID,
		CurrentStationID:       s.CurrentStationID,
		NextSuggestedStationID: s.NextSuggestedStationID,
		IsReworkLoop:           s.IsReworkLoop,
		EnteredAt:              s.EnteredAt,
		ExitedAt:               s.ExitedAt,
	}
}

// 5. MaterialConsumptionLog
type MaterialConsumptionLog struct {
	ID               string          `gorm:"primaryKey;type:varchar(255)"`
	LegalEntityID    string          `gorm:"type:varchar(255);not null"`
	WorkOrderID      string          `gorm:"type:varchar(255);not null;uniqueIndex:idx_mfg_mat_log"`
	MaterialID       string          `gorm:"type:varchar(255);not null;uniqueIndex:idx_mfg_mat_log"`
	RoutingStationID string          `gorm:"type:varchar(255);not null"`
	QuantityConsumed decimal.Decimal `gorm:"type:numeric(14,4);not null"`
	WarehouseID      string          `gorm:"type:varchar(255);not null"`
	OperatorHrID     string          `gorm:"type:varchar(255);not null"`
	ConsumedAt       time.Time       `gorm:"type:timestamp;not null;uniqueIndex:idx_mfg_mat_log"`
}

func (MaterialConsumptionLog) TableName() string {
	return "mfg_material_consumption_logs"
}

func ToMaterialConsumptionLogDomain(l *MaterialConsumptionLog) *domain.MaterialConsumptionLog {
	if l == nil {
		return nil
	}
	return &domain.MaterialConsumptionLog{
		ID:               l.ID,
		LegalEntityID:    l.LegalEntityID,
		WorkOrderID:      l.WorkOrderID,
		MaterialID:       l.MaterialID,
		RoutingStationID: l.RoutingStationID,
		QuantityConsumed: l.QuantityConsumed,
		WarehouseID:      l.WarehouseID,
		OperatorHrID:     l.OperatorHrID,
		ConsumedAt:       l.ConsumedAt,
	}
}

func FromMaterialConsumptionLogDomain(l *domain.MaterialConsumptionLog) *MaterialConsumptionLog {
	if l == nil {
		return nil
	}
	return &MaterialConsumptionLog{
		ID:               l.ID,
		LegalEntityID:    l.LegalEntityID,
		WorkOrderID:      l.WorkOrderID,
		MaterialID:       l.MaterialID,
		RoutingStationID: l.RoutingStationID,
		QuantityConsumed: l.QuantityConsumed,
		WarehouseID:      l.WarehouseID,
		OperatorHrID:     l.OperatorHrID,
		ConsumedAt:       l.ConsumedAt,
	}
}

// 6. ProductionYieldLog
type ProductionYieldLog struct {
	ID               string          `gorm:"primaryKey;type:varchar(255)"`
	LegalEntityID    string          `gorm:"type:varchar(255);not null"`
	WorkOrderID      string          `gorm:"type:varchar(255);not null;uniqueIndex:idx_mfg_yield_log"`
	RoutingStationID string          `gorm:"type:varchar(255);not null"`
	QuantityGood     decimal.Decimal `gorm:"type:numeric(14,4);not null"`
	QuantityScrap    decimal.Decimal `gorm:"type:numeric(14,4);not null"`
	OperatorHrID     string          `gorm:"type:varchar(255);not null"`
	RecordedAt       time.Time       `gorm:"type:timestamp;not null;uniqueIndex:idx_mfg_yield_log"`
}

func (ProductionYieldLog) TableName() string {
	return "mfg_production_yield_logs"
}

func ToProductionYieldLogDomain(l *ProductionYieldLog) *domain.ProductionYieldLog {
	if l == nil {
		return nil
	}
	return &domain.ProductionYieldLog{
		ID:               l.ID,
		LegalEntityID:    l.LegalEntityID,
		WorkOrderID:      l.WorkOrderID,
		RoutingStationID: l.RoutingStationID,
		QuantityGood:     l.QuantityGood,
		QuantityScrap:    l.QuantityScrap,
		OperatorHrID:     l.OperatorHrID,
		RecordedAt:       l.RecordedAt,
	}
}

func FromProductionYieldLogDomain(l *domain.ProductionYieldLog) *ProductionYieldLog {
	if l == nil {
		return nil
	}
	return &ProductionYieldLog{
		ID:               l.ID,
		LegalEntityID:    l.LegalEntityID,
		WorkOrderID:      l.WorkOrderID,
		RoutingStationID: l.RoutingStationID,
		QuantityGood:     l.QuantityGood,
		QuantityScrap:    l.QuantityScrap,
		OperatorHrID:     l.OperatorHrID,
		RecordedAt:       l.RecordedAt,
	}
}

// 7. TransactionalOutbox
type TransactionalOutbox struct {
	ID          string    `gorm:"primaryKey;type:varchar(255)"`
	EventType   string    `gorm:"type:varchar(255);not null"`
	AggregateID string    `gorm:"type:varchar(255);not null"`
	Payload     []byte    `gorm:"type:jsonb;not null"`
	Status      string    `gorm:"type:varchar(50);not null;index:idx_mfg_outbox_status_date"`
	RetryCount  int       `gorm:"type:integer;not null;default:0"`
	CreatedAt   time.Time `gorm:"not null;index:idx_mfg_outbox_status_date"`
}

func (TransactionalOutbox) TableName() string {
	return "mfg_transactional_outbox"
}

func ToTransactionalOutboxDomain(o *TransactionalOutbox) *domain.TransactionalOutbox {
	if o == nil {
		return nil
	}
	var payload interface{}
	if len(o.Payload) > 0 {
		_ = json.Unmarshal(o.Payload, &payload)
	}
	return &domain.TransactionalOutbox{
		ID:          o.ID,
		EventType:   o.EventType,
		AggregateID: o.AggregateID,
		Payload:     payload,
		Status:      domain.OutboxStatus(o.Status),
		RetryCount:  o.RetryCount,
		CreatedAt:   o.CreatedAt,
	}
}

func FromTransactionalOutboxDomain(o *domain.TransactionalOutbox) *TransactionalOutbox {
	if o == nil {
		return nil
	}
	payloadBytes, _ := json.Marshal(o.Payload)
	return &TransactionalOutbox{
		ID:          o.ID,
		EventType:   o.EventType,
		AggregateID: o.AggregateID,
		Payload:     payloadBytes,
		Status:      string(o.Status),
		RetryCount:  o.RetryCount,
		CreatedAt:   o.CreatedAt,
	}
}

// 8. KafkaEventInbox
type KafkaEventInbox struct {
	AttemptCount     int       `gorm:"type:integer;default:0;not null"`
	EventID          string    `gorm:"primaryKey;type:varchar(255)"`
	EventType        string    `gorm:"type:varchar(255);not null"`
	ProcessedAt      time.Time `gorm:"not null"`
	ProcessingStatus string    `gorm:"type:varchar(50);not null"`
	Payload          []byte    `gorm:"type:jsonb;not null"`
}

func (KafkaEventInbox) TableName() string {
	return "mfg_kafka_event_inbox"
}

func ToKafkaEventInboxDomain(i *KafkaEventInbox) *domain.KafkaEventInbox {
	if i == nil {
		return nil
	}
	var payload interface{}
	if len(i.Payload) > 0 {
		_ = json.Unmarshal(i.Payload, &payload)
	}
	return &domain.KafkaEventInbox{
		AttemptCount:     i.AttemptCount,
		EventID:          i.EventID,
		EventType:        i.EventType,
		ProcessedAt:      i.ProcessedAt,
		ProcessingStatus: domain.EventProcessingStatus(i.ProcessingStatus),
		Payload:          payload,
	}
}

func FromKafkaEventInboxDomain(i *domain.KafkaEventInbox) *KafkaEventInbox {
	if i == nil {
		return nil
	}
	payloadBytes, _ := json.Marshal(i.Payload)
	return &KafkaEventInbox{
		AttemptCount:     i.AttemptCount,
		EventID:          i.EventID,
		EventType:        i.EventType,
		ProcessedAt:      i.ProcessedAt,
		ProcessingStatus: string(i.ProcessingStatus),
		Payload:          payloadBytes,
	}
}
