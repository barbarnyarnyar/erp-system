package sql

import (
	"encoding/json"
	"time"

	"github.com/erp-system/eam-service/internal/business/domain"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// 1. Facility
type Facility struct {
	ID              string    `gorm:"primaryKey;type:varchar(255)"`
	LegalEntityID   string    `gorm:"type:varchar(255);not null"`
	Name            string    `gorm:"type:varchar(255);not null"`
	PhysicalAddress string    `gorm:"type:varchar(255);not null"`
	IsActive        bool      `gorm:"type:boolean;not null"`
	CreatedAt       time.Time `gorm:"not null"`
	UpdatedAt       time.Time `gorm:"not null"`
}

func (Facility) TableName() string {
	return "eam_facilities"
}

func ToFacilityDomain(f *Facility) *domain.Facility {
	if f == nil {
		return nil
	}
	return &domain.Facility{
		ID:              f.ID,
		LegalEntityID:   f.LegalEntityID,
		Name:            f.Name,
		PhysicalAddress: f.PhysicalAddress,
		IsActive:        f.IsActive,
		CreatedAt:       f.CreatedAt,
		UpdatedAt:       f.UpdatedAt,
	}
}

func FromFacilityDomain(f *domain.Facility) *Facility {
	if f == nil {
		return nil
	}
	return &Facility{
		ID:              f.ID,
		LegalEntityID:   f.LegalEntityID,
		Name:            f.Name,
		PhysicalAddress: f.PhysicalAddress,
		IsActive:        f.IsActive,
		CreatedAt:       f.CreatedAt,
		UpdatedAt:       f.UpdatedAt,
	}
}

// 2. Equipment
type Equipment struct {
	ID                       string         `gorm:"primaryKey;type:varchar(255)"`
	LegalEntityID            string         `gorm:"type:varchar(255);not null"`
	FacilityID               string         `gorm:"type:varchar(255);not null"`
	AssetTag                 string         `gorm:"type:varchar(100);not null;uniqueIndex"`
	Name                     string         `gorm:"type:varchar(255);not null"`
	Manufacturer             string         `gorm:"type:varchar(255);not null"`
	SerialNumber             string         `gorm:"type:varchar(255);not null"`
	FinancialAssetID         *string        `gorm:"type:varchar(255);default:null"`
	Status                   string         `gorm:"type:varchar(50);not null"`
	InstallationDate         time.Time      `gorm:"type:date;not null"`
	TechnicalSpecifications  []byte         `gorm:"type:jsonb;not null"`
	CreatedAt                time.Time      `gorm:"not null"`
	UpdatedAt                time.Time      `gorm:"not null"`
	DeletedAt                gorm.DeletedAt `gorm:"index"`
}

func (Equipment) TableName() string {
	return "eam_equipment"
}

func ToEquipmentDomain(e *Equipment) *domain.Equipment {
	if e == nil {
		return nil
	}
	var techSpecs interface{}
	if len(e.TechnicalSpecifications) > 0 {
		_ = json.Unmarshal(e.TechnicalSpecifications, &techSpecs)
	}
	var delAt *time.Time
	if e.DeletedAt.Valid {
		delAt = &e.DeletedAt.Time
	}
	return &domain.Equipment{
		ID:                      e.ID,
		LegalEntityID:           e.LegalEntityID,
		FacilityID:              e.FacilityID,
		AssetTag:                e.AssetTag,
		Name:                    e.Name,
		Manufacturer:            e.Manufacturer,
		SerialNumber:            e.SerialNumber,
		FinancialAssetID:        e.FinancialAssetID,
		Status:                  domain.EquipmentStatus(e.Status),
		InstallationDate:        e.InstallationDate,
		TechnicalSpecifications: techSpecs,
		CreatedAt:               e.CreatedAt,
		UpdatedAt:               e.UpdatedAt,
		DeletedAt:               delAt,
	}
}

func FromEquipmentDomain(e *domain.Equipment) *Equipment {
	if e == nil {
		return nil
	}
	techSpecsBytes, _ := json.Marshal(e.TechnicalSpecifications)
	var delAt gorm.DeletedAt
	if e.DeletedAt != nil {
		delAt = gorm.DeletedAt{Time: *e.DeletedAt, Valid: true}
	}
	statusStr := string(e.Status)
	return &Equipment{
		ID:                      e.ID,
		LegalEntityID:           e.LegalEntityID,
		FacilityID:              e.FacilityID,
		AssetTag:                e.AssetTag,
		Name:                    e.Name,
		Manufacturer:            e.Manufacturer,
		SerialNumber:            e.SerialNumber,
		FinancialAssetID:        e.FinancialAssetID,
		Status:                  statusStr,
		InstallationDate:        e.InstallationDate,
		TechnicalSpecifications: techSpecsBytes,
		CreatedAt:               e.CreatedAt,
		UpdatedAt:               e.UpdatedAt,
		DeletedAt:               delAt,
	}
}

// 3. MaintenanceWorkOrder
type MaintenanceWorkOrder struct {
	ID               string     `gorm:"primaryKey;type:varchar(255)"`
	LegalEntityID    string     `gorm:"type:varchar(255);not null"`
	EquipmentID      string     `gorm:"type:varchar(255);not null"`
	TicketNumber     string     `gorm:"type:varchar(100);not null;uniqueIndex"`
	Title            string     `gorm:"type:varchar(255);not null"`
	Description      string     `gorm:"type:text;not null"`
	Category         string     `gorm:"type:varchar(50);not null"`
	Priority         string     `gorm:"type:varchar(50);not null"`
	Status           string     `gorm:"type:varchar(50);not null"`
	ReportedByHrID   string     `gorm:"type:varchar(255);not null"`
	AssignedTechHrID *string    `gorm:"type:varchar(255);default:null"`
	ReportedAt       time.Time  `gorm:"not null"`
	StartedAt        *time.Time `gorm:"default:null"`
	ResolvedAt       *time.Time `gorm:"default:null"`
	ResolutionNotes  *string    `gorm:"type:text;default:null"`
	CreatedAt        time.Time  `gorm:"not null"`
	UpdatedAt        time.Time  `gorm:"not null"`
}

func (MaintenanceWorkOrder) TableName() string {
	return "eam_work_orders"
}

func ToMaintenanceWorkOrderDomain(w *MaintenanceWorkOrder) *domain.MaintenanceWorkOrder {
	if w == nil {
		return nil
	}
	return &domain.MaintenanceWorkOrder{
		ID:               w.ID,
		LegalEntityID:    w.LegalEntityID,
		EquipmentID:      w.EquipmentID,
		TicketNumber:     w.TicketNumber,
		Title:            w.Title,
		Description:      w.Description,
		Category:         domain.MaintenanceCategory(w.Category),
		Priority:         domain.WorkOrderPriority(w.Priority),
		Status:           domain.WorkOrderStatus(w.Status),
		ReportedByHrID:   w.ReportedByHrID,
		AssignedTechHrID: w.AssignedTechHrID,
		ReportedAt:       w.ReportedAt,
		StartedAt:        w.StartedAt,
		ResolvedAt:       w.ResolvedAt,
		ResolutionNotes:  w.ResolutionNotes,
		CreatedAt:        w.CreatedAt,
		UpdatedAt:        w.UpdatedAt,
	}
}

func FromMaintenanceWorkOrderDomain(w *domain.MaintenanceWorkOrder) *MaintenanceWorkOrder {
	if w == nil {
		return nil
	}
	category := string(w.Category)
	priority := string(w.Priority)
	status := string(w.Status)
	return &MaintenanceWorkOrder{
		ID:               w.ID,
		LegalEntityID:    w.LegalEntityID,
		EquipmentID:      w.EquipmentID,
		TicketNumber:     w.TicketNumber,
		Title:            w.Title,
		Description:      w.Description,
		Category:         category,
		Priority:         priority,
		Status:           status,
		ReportedByHrID:   w.ReportedByHrID,
		AssignedTechHrID: w.AssignedTechHrID,
		ReportedAt:       w.ReportedAt,
		StartedAt:        w.StartedAt,
		ResolvedAt:       w.ResolvedAt,
		ResolutionNotes:  w.ResolutionNotes,
		CreatedAt:        w.CreatedAt,
		UpdatedAt:        w.UpdatedAt,
	}
}

// 4. PreventativeSchedule
type PreventativeSchedule struct {
	ID             string     `gorm:"primaryKey;type:varchar(255)"`
	LegalEntityID  string     `gorm:"type:varchar(255);not null"`
	EquipmentID    string     `gorm:"type:varchar(255);not null"`
	Title          string     `gorm:"type:varchar(255);not null"`
	InstructionSet string     `gorm:"type:text;not null"`
	IntervalDays   int        `gorm:"type:integer;not null"`
	LastExecutedAt *time.Time `gorm:"default:null"`
	NextDueDate    time.Time  `gorm:"type:date;not null"`
	IsActive       bool       `gorm:"type:boolean;not null"`
	CreatedAt      time.Time  `gorm:"not null"`
	UpdatedAt      time.Time  `gorm:"not null"`
}

func (PreventativeSchedule) TableName() string {
	return "eam_pm_schedules"
}

func ToPreventativeScheduleDomain(p *PreventativeSchedule) *domain.PreventativeSchedule {
	if p == nil {
		return nil
	}
	return &domain.PreventativeSchedule{
		ID:             p.ID,
		LegalEntityID:  p.LegalEntityID,
		EquipmentID:    p.EquipmentID,
		Title:          p.Title,
		InstructionSet: p.InstructionSet,
		IntervalDays:   p.IntervalDays,
		LastExecutedAt: p.LastExecutedAt,
		NextDueDate:    p.NextDueDate,
		IsActive:       p.IsActive,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}

func FromPreventativeScheduleDomain(p *domain.PreventativeSchedule) *PreventativeSchedule {
	if p == nil {
		return nil
	}
	return &PreventativeSchedule{
		ID:             p.ID,
		LegalEntityID:  p.LegalEntityID,
		EquipmentID:    p.EquipmentID,
		Title:          p.Title,
		InstructionSet: p.InstructionSet,
		IntervalDays:   p.IntervalDays,
		LastExecutedAt: p.LastExecutedAt,
		NextDueDate:    p.NextDueDate,
		IsActive:       p.IsActive,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}

// 5. TelemetryIngestBuffer
type TelemetryIngestBuffer struct {
	ID            string          `gorm:"primaryKey;type:varchar(255)"`
	LegalEntityID string          `gorm:"type:varchar(255);not null"`
	EquipmentID   string          `gorm:"type:varchar(255);not null"`
	SensorKey     string          `gorm:"type:varchar(255);not null"`
	ReadingValue  decimal.Decimal `gorm:"type:numeric(12,4);not null"`
	RecordedAt    time.Time       `gorm:"not null"`
}

func (TelemetryIngestBuffer) TableName() string {
	return "eam_telemetry_ingest_buffer"
}

func ToTelemetryIngestBufferDomain(t *TelemetryIngestBuffer) *domain.TelemetryIngestBuffer {
	if t == nil {
		return nil
	}
	return &domain.TelemetryIngestBuffer{
		ID:            t.ID,
		LegalEntityID: t.LegalEntityID,
		EquipmentID:   t.EquipmentID,
		SensorKey:     t.SensorKey,
		ReadingValue:  t.ReadingValue,
		RecordedAt:    t.RecordedAt,
	}
}

func FromTelemetryIngestBufferDomain(t *domain.TelemetryIngestBuffer) *TelemetryIngestBuffer {
	if t == nil {
		return nil
	}
	return &TelemetryIngestBuffer{
		ID:            t.ID,
		LegalEntityID: t.LegalEntityID,
		EquipmentID:   t.EquipmentID,
		SensorKey:     t.SensorKey,
		ReadingValue:  t.ReadingValue,
		RecordedAt:    t.RecordedAt,
	}
}

// 6. TransactionalOutbox
type TransactionalOutbox struct {
	ID          string    `gorm:"primaryKey;type:varchar(255)"`
	EventType   string    `gorm:"type:varchar(255);not null"`
	AggregateID string    `gorm:"type:varchar(255);not null"`
	Payload     []byte    `gorm:"type:jsonb;not null"`
	Status      string    `gorm:"type:varchar(50);not null;index:idx_eam_outbox_status_date"`
	CreatedAt   time.Time `gorm:"not null;index:idx_eam_outbox_status_date"`
}

func (TransactionalOutbox) TableName() string {
	return "eam_transactional_outbox"
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
		CreatedAt:   o.CreatedAt,
	}
}

func FromTransactionalOutboxDomain(o *domain.TransactionalOutbox) *TransactionalOutbox {
	if o == nil {
		return nil
	}
	payloadBytes, _ := json.Marshal(o.Payload)
	statusStr := string(o.Status)
	return &TransactionalOutbox{
		ID:          o.ID,
		EventType:   o.EventType,
		AggregateID: o.AggregateID,
		Payload:     payloadBytes,
		Status:      statusStr,
		CreatedAt:   o.CreatedAt,
	}
}

// 7. KafkaEventInbox
type KafkaEventInbox struct {
	AttemptCount     int       `gorm:"type:integer;default:0;not null"`
	EventID          string    `gorm:"primaryKey;type:varchar(255)"`
	EventType        string    `gorm:"type:varchar(255);not null"`
	ProcessedAt      time.Time `gorm:"not null"`
	ProcessingStatus string    `gorm:"type:varchar(50);not null"`
	Payload          []byte    `gorm:"type:jsonb;not null"`
}

func (KafkaEventInbox) TableName() string {
	return "eam_kafka_event_inbox"
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
	statusStr := string(i.ProcessingStatus)
	return &KafkaEventInbox{
		AttemptCount:     i.AttemptCount,
		EventID:          i.EventID,
		EventType:        i.EventType,
		ProcessedAt:      i.ProcessedAt,
		ProcessingStatus: statusStr,
		Payload:          payloadBytes,
	}
}
