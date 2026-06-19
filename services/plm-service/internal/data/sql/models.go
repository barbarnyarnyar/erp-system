package sql

import (
	"time"

	"github.com/erp-system/plm-service/internal/business/domain"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type MaterialMaster struct {
	ID                      string         `gorm:"primaryKey;type:varchar(255)"`
	LegalEntityID           string         `gorm:"type:varchar(255);index"`
	Sku                     string         `gorm:"type:varchar(255);uniqueIndex:idx_mat_sku"`
	Description             string         `gorm:"type:text"`
	Uom                     string         `gorm:"column:uom;type:varchar(50)"`
	ProcurementType         string         `gorm:"type:varchar(50)"`
	Status                  string         `gorm:"type:varchar(50)"`
	TechnicalSpecifications string         `gorm:"type:jsonb"`
	Version                 int            `gorm:"type:int;default:1"`
	CreatedAt               time.Time      `gorm:"index"`
	UpdatedAt               time.Time
	DeletedAt               gorm.DeletedAt `gorm:"index"`
}

func (MaterialMaster) TableName() string {
	return "plm_materials"
}

func ToMaterialMasterDomain(m *MaterialMaster) *domain.MaterialMaster {
	if m == nil {
		return nil
	}
	var delAt *time.Time
	if m.DeletedAt.Valid {
		delAt = &m.DeletedAt.Time
	}
	return &domain.MaterialMaster{
		ID:                      m.ID,
		LegalEntityID:           m.LegalEntityID,
		Sku:                     m.Sku,
		Description:             m.Description,
		Uom:                     m.Uom,
		ProcurementType:         domain.ProcurementType(m.ProcurementType),
		Status:                  domain.MaterialStatus(m.Status),
		TechnicalSpecifications: m.TechnicalSpecifications,
		Version:                 m.Version,
		CreatedAt:               m.CreatedAt,
		UpdatedAt:               m.UpdatedAt,
		DeletedAt:               delAt,
	}
}

func FromMaterialMasterDomain(m *domain.MaterialMaster) *MaterialMaster {
	if m == nil {
		return nil
	}
	var delAt gorm.DeletedAt
	if m.DeletedAt != nil {
		delAt = gorm.DeletedAt{Time: *m.DeletedAt, Valid: true}
	}
	var specs string
	if m.TechnicalSpecifications != nil {
		if s, ok := m.TechnicalSpecifications.(string); ok {
			specs = s
		}
	}
	return &MaterialMaster{
		ID:                      m.ID,
		LegalEntityID:           m.LegalEntityID,
		Sku:                     m.Sku,
		Description:             m.Description,
		Uom:                     m.Uom,
		ProcurementType:         string(m.ProcurementType),
		Status:                  string(m.Status),
		TechnicalSpecifications: specs,
		Version:                 m.Version,
		CreatedAt:               m.CreatedAt,
		UpdatedAt:               m.UpdatedAt,
		DeletedAt:               delAt,
	}
}

type BomHeader struct {
	ID            string    `gorm:"primaryKey;type:varchar(255)"`
	LegalEntityID string    `gorm:"type:varchar(255);index"`
	MaterialID    string    `gorm:"type:varchar(255);index"`
	EcoID         *string   `gorm:"type:varchar(255);index"`
	VersionString string    `gorm:"type:varchar(50)"`
	Status        string    `gorm:"type:varchar(50)"`
	CreatedAt     time.Time `gorm:"index"`
	UpdatedAt     time.Time
}

func (BomHeader) TableName() string {
	return "plm_bom_headers"
}

func ToBomHeaderDomain(b *BomHeader) *domain.BomHeader {
	if b == nil {
		return nil
	}
	return &domain.BomHeader{
		ID:            b.ID,
		LegalEntityID: b.LegalEntityID,
		MaterialID:    b.MaterialID,
		EcoID:         b.EcoID,
		VersionString: b.VersionString,
		Status:        domain.BomStatus(b.Status),
		CreatedAt:     b.CreatedAt,
		UpdatedAt:     b.UpdatedAt,
	}
}

func FromBomHeaderDomain(b *domain.BomHeader) *BomHeader {
	if b == nil {
		return nil
	}
	return &BomHeader{
		ID:            b.ID,
		LegalEntityID: b.LegalEntityID,
		MaterialID:    b.MaterialID,
		EcoID:         b.EcoID,
		VersionString: b.VersionString,
		Status:        string(b.Status),
		CreatedAt:     b.CreatedAt,
		UpdatedAt:     b.UpdatedAt,
	}
}

type BomLine struct {
	ID                  string          `gorm:"primaryKey;type:varchar(255)"`
	BomHeaderID         string          `gorm:"type:varchar(255);index"`
	ComponentMaterialID string          `gorm:"type:varchar(255);index"`
	SequenceNumber      int             `gorm:"type:int"`
	QuantityRequired    decimal.Decimal `gorm:"type:numeric(14,4)"`
	Uom                 string          `gorm:"column:uom;type:varchar(50)"`
	ScrapPercentage     decimal.Decimal `gorm:"type:numeric(5,4)"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func (BomLine) TableName() string {
	return "plm_bom_lines"
}

func ToBomLineDomain(l *BomLine) *domain.BomLine {
	if l == nil {
		return nil
	}
	return &domain.BomLine{
		ID:                  l.ID,
		BomHeaderID:         l.BomHeaderID,
		ComponentMaterialID: l.ComponentMaterialID,
		SequenceNumber:     l.SequenceNumber,
		QuantityRequired:   l.QuantityRequired,
		Uom:                l.Uom,
		ScrapPercentage:    l.ScrapPercentage,
		CreatedAt:          l.CreatedAt,
		UpdatedAt:          l.UpdatedAt,
	}
}

func FromBomLineDomain(l *domain.BomLine) *BomLine {
	if l == nil {
		return nil
	}
	return &BomLine{
		ID:                  l.ID,
		BomHeaderID:         l.BomHeaderID,
		ComponentMaterialID: l.ComponentMaterialID,
		SequenceNumber:     l.SequenceNumber,
		QuantityRequired:   l.QuantityRequired,
		Uom:                l.Uom,
		ScrapPercentage:    l.ScrapPercentage,
		CreatedAt:          l.CreatedAt,
		UpdatedAt:          l.UpdatedAt,
	}
}

type EngineeringChangeOrder struct {
	ID               string    `gorm:"primaryKey;type:varchar(255)"`
	LegalEntityID    string    `gorm:"type:varchar(255);index"`
	TargetMaterialID string    `gorm:"type:varchar(255);index"`
	EcoNumber        string    `gorm:"type:varchar(255);uniqueIndex:idx_eco_num"`
	Title            string    `gorm:"type:varchar(255)"`
	Description      string    `gorm:"type:text"`
	Status           string    `gorm:"type:varchar(50)"`
	RequestedByHrID  string    `gorm:"type:varchar(255);index"`
	ApprovedByHrID   *string   `gorm:"type:varchar(255);index"`
	Version          int       `gorm:"type:int;default:1"`
	CreatedAt        time.Time `gorm:"index"`
	UpdatedAt        time.Time
}

func (EngineeringChangeOrder) TableName() string {
	return "plm_engineering_change_orders"
}

func ToEcoDomain(eco *EngineeringChangeOrder) *domain.EngineeringChangeOrder {
	if eco == nil {
		return nil
	}
	return &domain.EngineeringChangeOrder{
		ID:               eco.ID,
		LegalEntityID:    eco.LegalEntityID,
		TargetMaterialID: eco.TargetMaterialID,
		EcoNumber:        eco.EcoNumber,
		Title:            eco.Title,
		Description:      eco.Description,
		Status:           domain.EcoStatus(eco.Status),
		RequestedByHrID:  eco.RequestedByHrID,
		ApprovedByHrID:   eco.ApprovedByHrID,
		Version:          eco.Version,
		CreatedAt:        eco.CreatedAt,
		UpdatedAt:        eco.UpdatedAt,
	}
}

func FromEcoDomain(eco *domain.EngineeringChangeOrder) *EngineeringChangeOrder {
	if eco == nil {
		return nil
	}
	return &EngineeringChangeOrder{
		ID:               eco.ID,
		LegalEntityID:    eco.LegalEntityID,
		TargetMaterialID: eco.TargetMaterialID,
		EcoNumber:        eco.EcoNumber,
		Title:            eco.Title,
		Description:      eco.Description,
		Status:           string(eco.Status),
		RequestedByHrID:  eco.RequestedByHrID,
		ApprovedByHrID:   eco.ApprovedByHrID,
		Version:          eco.Version,
		CreatedAt:        eco.CreatedAt,
		UpdatedAt:        eco.UpdatedAt,
	}
}

type TransactionalOutbox struct {
	ID          string    `gorm:"primaryKey;type:varchar(255)"`
	EventType   string    `gorm:"type:varchar(255);index"`
	AggregateID string    `gorm:"type:varchar(255);index"`
	Payload     string    `gorm:"type:jsonb"`
	Status      string    `gorm:"type:varchar(50);index"`
	CreatedAt   time.Time `gorm:"index"`
}

func (TransactionalOutbox) TableName() string {
	return "plm_transactional_outbox"
}

func ToOutboxDomain(o *TransactionalOutbox) *domain.TransactionalOutbox {
	if o == nil {
		return nil
	}
	return &domain.TransactionalOutbox{
		ID:          o.ID,
		EventType:   o.EventType,
		AggregateID: o.AggregateID,
		Payload:     o.Payload,
		Status:      domain.OutboxStatus(o.Status),
		CreatedAt:   o.CreatedAt,
	}
}

func FromOutboxDomain(o *domain.TransactionalOutbox) *TransactionalOutbox {
	if o == nil {
		return nil
	}
	var payloadStr string
	if o.Payload != nil {
		if s, ok := o.Payload.(string); ok {
			payloadStr = s
		}
	}
	return &TransactionalOutbox{
		ID:          o.ID,
		EventType:   o.EventType,
		AggregateID: o.AggregateID,
		Payload:     payloadStr,
		Status:      string(o.Status),
		CreatedAt:   o.CreatedAt,
	}
}

type KafkaEventInbox struct {
	AttemptCount     int       `gorm:"type:integer;default:0;not null"`
	EventID          string    `gorm:"primaryKey;type:varchar(255)"`
	EventType        string    `gorm:"type:varchar(255)"`
	ProcessedAt      time.Time `gorm:"index"`
	ProcessingStatus string    `gorm:"type:varchar(50)"`
	Payload          string    `gorm:"type:jsonb"`
}

func (KafkaEventInbox) TableName() string {
	return "plm_kafka_event_inbox"
}

func ToInboxDomain(i *KafkaEventInbox) *domain.KafkaEventInbox {
	if i == nil {
		return nil
	}
	return &domain.KafkaEventInbox{
		AttemptCount:     i.AttemptCount,
		EventID:          i.EventID,
		EventType:        i.EventType,
		ProcessedAt:      i.ProcessedAt,
		ProcessingStatus: domain.EventProcessingStatus(i.ProcessingStatus),
		Payload:          i.Payload,
	}
}

func FromInboxDomain(i *domain.KafkaEventInbox) *KafkaEventInbox {
	if i == nil {
		return nil
	}
	var payloadStr string
	if i.Payload != nil {
		if s, ok := i.Payload.(string); ok {
			payloadStr = s
		}
	}
	return &KafkaEventInbox{
		AttemptCount:     i.AttemptCount,
		EventID:          i.EventID,
		EventType:        i.EventType,
		ProcessedAt:      i.ProcessedAt,
		ProcessingStatus: string(i.ProcessingStatus),
		Payload:          payloadStr,
	}
}
