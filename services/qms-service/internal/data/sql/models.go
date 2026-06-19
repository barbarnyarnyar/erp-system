package sql

import (
	"encoding/json"
	"time"

	"github.com/erp-system/qms-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

// 1. InspectionPlan
type InspectionPlan struct {
	ID            string    `gorm:"primaryKey;type:varchar(255)"`
	LegalEntityID string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_qms_plan_tenant_material"`
	MaterialID    string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_qms_plan_tenant_material"`
	PlanName      string    `gorm:"type:varchar(255);not null"`
	IsActive      bool      `gorm:"type:boolean;not null"`
	Version       int       `gorm:"type:integer;not null"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`
}

func (InspectionPlan) TableName() string {
	return "qms_inspection_plans"
}

func ToInspectionPlanDomain(p *InspectionPlan) *domain.InspectionPlan {
	if p == nil {
		return nil
	}
	return &domain.InspectionPlan{
		ID:            p.ID,
		LegalEntityID: p.LegalEntityID,
		MaterialID:    p.MaterialID,
		PlanName:      p.PlanName,
		IsActive:      p.IsActive,
		Version:       p.Version,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

func FromInspectionPlanDomain(p *domain.InspectionPlan) *InspectionPlan {
	if p == nil {
		return nil
	}
	return &InspectionPlan{
		ID:            p.ID,
		LegalEntityID: p.LegalEntityID,
		MaterialID:    p.MaterialID,
		PlanName:      p.PlanName,
		IsActive:      p.IsActive,
		Version:       p.Version,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

// 2. InspectionMetricDefinition
type InspectionMetricDefinition struct {
	ID                string           `gorm:"primaryKey;type:varchar(255)"`
	InspectionPlanID  string           `gorm:"type:varchar(255);not null;uniqueIndex:idx_qms_metric_plan_key"`
	MetricKey         string           `gorm:"type:varchar(255);not null;uniqueIndex:idx_qms_metric_plan_key"`
	DisplayName       string           `gorm:"type:varchar(255);not null"`
	DataType          string           `gorm:"type:varchar(50);not null"`
	MinToleranceLimit *decimal.Decimal `gorm:"type:numeric(12,4);default:null"`
	MaxToleranceLimit *decimal.Decimal `gorm:"type:numeric(12,4);default:null"`
	CreatedAt         time.Time        `gorm:"not null"`
	UpdatedAt         time.Time        `gorm:"not null"`
}

func (InspectionMetricDefinition) TableName() string {
	return "qms_inspection_metric_definitions"
}

func ToInspectionMetricDefinitionDomain(d *InspectionMetricDefinition) *domain.InspectionMetricDefinition {
	if d == nil {
		return nil
	}
	return &domain.InspectionMetricDefinition{
		ID:                d.ID,
		InspectionPlanID:  d.InspectionPlanID,
		MetricKey:         d.MetricKey,
		DisplayName:       d.DisplayName,
		DataType:          domain.MetricDataType(d.DataType),
		MinToleranceLimit: d.MinToleranceLimit,
		MaxToleranceLimit: d.MaxToleranceLimit,
		CreatedAt:         d.CreatedAt,
		UpdatedAt:         d.UpdatedAt,
	}
}

func FromInspectionMetricDefinitionDomain(d *domain.InspectionMetricDefinition) *InspectionMetricDefinition {
	if d == nil {
		return nil
	}
	dataTypeStr := ""
	if dt, ok := d.DataType.(domain.MetricDataType); ok {
		dataTypeStr = string(dt)
	} else if dt, ok := d.DataType.(string); ok {
		dataTypeStr = dt
	}
	return &InspectionMetricDefinition{
		ID:                d.ID,
		InspectionPlanID:  d.InspectionPlanID,
		MetricKey:         d.MetricKey,
		DisplayName:       d.DisplayName,
		DataType:          dataTypeStr,
		MinToleranceLimit: d.MinToleranceLimit,
		MaxToleranceLimit: d.MaxToleranceLimit,
		CreatedAt:         d.CreatedAt,
		UpdatedAt:         d.UpdatedAt,
	}
}

// 3. QualityInspection
type QualityInspection struct {
	ID               string    `gorm:"primaryKey;type:varchar(255)"`
	LegalEntityID    string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_qms_insp_tenant_num"`
	InspectionPlanID string    `gorm:"type:varchar(255);not null"`
	InspectionNumber string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_qms_insp_tenant_num"`
	TriggerSource    string    `gorm:"type:varchar(100);not null"`
	SourceDocumentID string    `gorm:"type:varchar(255);not null"`
	Status           string    `gorm:"type:varchar(100);not null"`
	InspectorHrID    *string   `gorm:"type:varchar(255);default:null"`
	Version          int       `gorm:"type:integer;not null"`
	CreatedAt        time.Time `gorm:"not null"`
	UpdatedAt        time.Time `gorm:"not null"`
}

func (QualityInspection) TableName() string {
	return "qms_quality_inspections"
}

func ToQualityInspectionDomain(i *QualityInspection) *domain.QualityInspection {
	if i == nil {
		return nil
	}
	return &domain.QualityInspection{
		ID:               i.ID,
		LegalEntityID:    i.LegalEntityID,
		InspectionPlanID: i.InspectionPlanID,
		InspectionNumber: i.InspectionNumber,
		TriggerSource:    domain.InspectionTriggerType(i.TriggerSource),
		SourceDocumentID: i.SourceDocumentID,
		Status:           domain.InspectionStatus(i.Status),
		InspectorHrID:    i.InspectorHrID,
		Version:          i.Version,
		CreatedAt:        i.CreatedAt,
		UpdatedAt:        i.UpdatedAt,
	}
}

func FromQualityInspectionDomain(i *domain.QualityInspection) *QualityInspection {
	if i == nil {
		return nil
	}
	triggerStr := ""
	if t, ok := i.TriggerSource.(domain.InspectionTriggerType); ok {
		triggerStr = string(t)
	} else if t, ok := i.TriggerSource.(string); ok {
		triggerStr = t
	}
	statusStr := ""
	if s, ok := i.Status.(domain.InspectionStatus); ok {
		statusStr = string(s)
	} else if s, ok := i.Status.(string); ok {
		statusStr = s
	}
	return &QualityInspection{
		ID:               i.ID,
		LegalEntityID:    i.LegalEntityID,
		InspectionPlanID: i.InspectionPlanID,
		InspectionNumber: i.InspectionNumber,
		TriggerSource:    triggerStr,
		SourceDocumentID: i.SourceDocumentID,
		Status:           statusStr,
		InspectorHrID:    i.InspectorHrID,
		Version:          i.Version,
		CreatedAt:        i.CreatedAt,
		UpdatedAt:        i.UpdatedAt,
	}
}

// 4. InspectionResultLine
type InspectionResultLine struct {
	ID                   string           `gorm:"primaryKey;type:varchar(255)"`
	InspectionID         string           `gorm:"type:varchar(255);not null;uniqueIndex:idx_qms_result_line"`
	MetricDefinitionID   string           `gorm:"type:varchar(255);not null;uniqueIndex:idx_qms_result_line"`
	SampleSequence       int              `gorm:"type:integer;not null;uniqueIndex:idx_qms_result_line"`
	MeasuredNumericValue *decimal.Decimal `gorm:"type:numeric(12,4);default:null"`
	MeasuredBooleanValue *bool            `gorm:"type:boolean;default:null"`
	IsCompliant          bool             `gorm:"type:boolean;not null"`
	CreatedAt            time.Time        `gorm:"not null;uniqueIndex:idx_qms_result_line"`
}

func (InspectionResultLine) TableName() string {
	return "qms_inspection_results"
}

func ToInspectionResultLineDomain(l *InspectionResultLine) *domain.InspectionResultLine {
	if l == nil {
		return nil
	}
	return &domain.InspectionResultLine{
		ID:                   l.ID,
		InspectionID:         l.InspectionID,
		MetricDefinitionID:   l.MetricDefinitionID,
		SampleSequence:       l.SampleSequence,
		MeasuredNumericValue: l.MeasuredNumericValue,
		MeasuredBooleanValue: l.MeasuredBooleanValue,
		IsCompliant:          l.IsCompliant,
		CreatedAt:            l.CreatedAt,
	}
}

func FromInspectionResultLineDomain(l *domain.InspectionResultLine) *InspectionResultLine {
	if l == nil {
		return nil
	}
	return &InspectionResultLine{
		ID:                   l.ID,
		InspectionID:         l.InspectionID,
		MetricDefinitionID:   l.MetricDefinitionID,
		SampleSequence:       l.SampleSequence,
		MeasuredNumericValue: l.MeasuredNumericValue,
		MeasuredBooleanValue: l.MeasuredBooleanValue,
		IsCompliant:          l.IsCompliant,
		CreatedAt:            l.CreatedAt,
	}
}

// 5. NonConformanceLog
type NonConformanceLog struct {
	ID                string          `gorm:"primaryKey;type:varchar(255)"`
	LegalEntityID     string          `gorm:"type:varchar(255);not null;uniqueIndex:idx_qms_nc_tenant_num"`
	InspectionID      string          `gorm:"type:varchar(255);not null"`
	NcNumber          string          `gorm:"type:varchar(255);not null;uniqueIndex:idx_qms_nc_tenant_num"`
	MaterialID        string          `gorm:"type:varchar(255);not null"`
	DefectDescription string          `gorm:"type:text;not null"`
	QuantityDefective decimal.Decimal `gorm:"type:numeric(14,4);not null"`
	IsQuarantined     bool            `gorm:"type:boolean;not null"`
	Disposition       *string         `gorm:"type:varchar(100);default:null"`
	DispositionNotes  *string         `gorm:"type:text;default:null"`
	ResolvedByHrID    *string         `gorm:"type:varchar(255);default:null"`
	ResolvedAt        *time.Time      `gorm:"default:null"`
	Version           int             `gorm:"type:integer;not null"`
	CreatedAt         time.Time       `gorm:"not null"`
	UpdatedAt         time.Time       `gorm:"not null"`
}

func (NonConformanceLog) TableName() string {
	return "qms_non_conformances"
}

func ToNonConformanceLogDomain(n *NonConformanceLog) *domain.NonConformanceLog {
	if n == nil {
		return nil
	}
	var disp interface{}
	if n.Disposition != nil {
		disp = domain.DispositionAction(*n.Disposition)
	}
	return &domain.NonConformanceLog{
		ID:                n.ID,
		LegalEntityID:     n.LegalEntityID,
		InspectionID:      n.InspectionID,
		NcNumber:          n.NcNumber,
		MaterialID:        n.MaterialID,
		DefectDescription: n.DefectDescription,
		QuantityDefective: n.QuantityDefective,
		IsQuarantined:     n.IsQuarantined,
		Disposition:       &disp,
		DispositionNotes:  n.DispositionNotes,
		ResolvedByHrID:    n.ResolvedByHrID,
		ResolvedAt:        n.ResolvedAt,
		Version:           n.Version,
		CreatedAt:         n.CreatedAt,
		UpdatedAt:         n.UpdatedAt,
	}
}

func FromNonConformanceLogDomain(n *domain.NonConformanceLog) *NonConformanceLog {
	if n == nil {
		return nil
	}
	var dispStr *string
	if n.Disposition != nil && *n.Disposition != nil {
		if val, ok := (*n.Disposition).(domain.DispositionAction); ok {
			str := string(val)
			dispStr = &str
		} else if val, ok := (*n.Disposition).(string); ok {
			dispStr = &val
		}
	}
	return &NonConformanceLog{
		ID:                n.ID,
		LegalEntityID:     n.LegalEntityID,
		InspectionID:      n.InspectionID,
		NcNumber:          n.NcNumber,
		MaterialID:        n.MaterialID,
		DefectDescription: n.DefectDescription,
		QuantityDefective: n.QuantityDefective,
		IsQuarantined:     n.IsQuarantined,
		Disposition:       dispStr,
		DispositionNotes:  n.DispositionNotes,
		ResolvedByHrID:    n.ResolvedByHrID,
		ResolvedAt:        n.ResolvedAt,
		Version:           n.Version,
		CreatedAt:         n.CreatedAt,
		UpdatedAt:         n.UpdatedAt,
	}
}

// 6. TransactionalOutbox
type TransactionalOutbox struct {
	ID          string    `gorm:"primaryKey;type:varchar(255)"`
	EventType   string    `gorm:"type:varchar(255);not null"`
	AggregateID string    `gorm:"type:varchar(255);not null"`
	Payload     []byte    `gorm:"type:jsonb;not null"`
	Status      string    `gorm:"type:varchar(50);not null;index:idx_qms_outbox_status_date"`
	CreatedAt   time.Time `gorm:"not null;index:idx_qms_outbox_status_date"`
}

func (TransactionalOutbox) TableName() string {
	return "qms_transactional_outbox"
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
	statusStr := "PENDING"
	if s, ok := o.Status.(domain.OutboxStatus); ok {
		statusStr = string(s)
	} else if s, ok := o.Status.(string); ok {
		statusStr = s
	}
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
	return "qms_kafka_event_inbox"
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
	statusStr := "SUCCESS"
	if s, ok := i.ProcessingStatus.(domain.EventProcessingStatus); ok {
		statusStr = string(s)
	} else if s, ok := i.ProcessingStatus.(string); ok {
		statusStr = s
	}
	return &KafkaEventInbox{
		AttemptCount:     i.AttemptCount,
		EventID:          i.EventID,
		EventType:        i.EventType,
		ProcessedAt:      i.ProcessedAt,
		ProcessingStatus: statusStr,
		Payload:          payloadBytes,
	}
}
