package sql

import (
	"encoding/json"
	"time"

	"github.com/erp-system/pm-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

// 1. Project
type Project struct {
	ID            string    `gorm:"primaryKey;type:uuid"`
	LegalEntityID string    `gorm:"type:uuid;not null;uniqueIndex:idx_prj_project_code_tenant"`
	CustomerID    string    `gorm:"type:uuid;not null"`
	ProjectCode   string    `gorm:"type:varchar(100);not null;uniqueIndex:idx_prj_project_code_tenant"`
	Name          string    `gorm:"type:varchar(255);not null"`
	Status        string    `gorm:"type:varchar(50);not null"`
	BillingMethod string    `gorm:"type:varchar(50);not null"`
	StartDate     time.Time `gorm:"type:date;not null"`
	EndDate       *time.Time `gorm:"type:date;default:null"`
	Version       int       `gorm:"type:integer;not null;default:1"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`
}

func (Project) TableName() string {
	return "prj_projects"
}

func ToProjectDomain(p *Project) *domain.Project {
	if p == nil {
		return nil
	}
	return &domain.Project{
		ID:            p.ID,
		LegalEntityID: p.LegalEntityID,
		CustomerID:    p.CustomerID,
		ProjectCode:   p.ProjectCode,
		Name:          p.Name,
		Status:        domain.ProjectStatus(p.Status),
		BillingMethod: domain.BillingMethod(p.BillingMethod),
		StartDate:     p.StartDate,
		EndDate:       p.EndDate,
		Version:       p.Version,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

func FromProjectDomain(p *domain.Project) *Project {
	if p == nil {
		return nil
	}
	return &Project{
		ID:            p.ID,
		LegalEntityID: p.LegalEntityID,
		CustomerID:    p.CustomerID,
		ProjectCode:   p.ProjectCode,
		Name:          p.Name,
		Status:        string(p.Status),
		BillingMethod: string(p.BillingMethod),
		StartDate:     p.StartDate,
		EndDate:       p.EndDate,
		Version:       p.Version,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

// 2. WbsNode
type WbsNode struct {
	ID                      string          `gorm:"primaryKey;type:uuid"`
	ProjectID               string          `gorm:"type:uuid;not null;uniqueIndex:idx_prj_wbs_node_code"`
	ParentNodeID            *string         `gorm:"type:uuid;default:null"`
	WbsDepthLevel           int             `gorm:"type:integer;not null;default:0"`
	NodeCode                string          `gorm:"type:varchar(100);not null;uniqueIndex:idx_prj_wbs_node_code"`
	Title                   string          `gorm:"type:varchar(255);not null"`
	NodeType                string          `gorm:"type:varchar(50);not null"`
	EstimatedHours          decimal.Decimal `gorm:"type:numeric(8,2);not null;default:0"`
	BudgetRevenueFunctional *decimal.Decimal `gorm:"type:numeric(18,4);default:null"`
	IsCompleted             bool            `gorm:"type:boolean;not null;default:false"`
	Version                 int             `gorm:"type:integer;not null;default:1"`
	CreatedAt               time.Time       `gorm:"not null"`
	UpdatedAt               time.Time       `gorm:"not null"`
}

func (WbsNode) TableName() string {
	return "prj_wbs_nodes"
}

func ToWbsNodeDomain(w *WbsNode) *domain.WbsNode {
	if w == nil {
		return nil
	}
	return &domain.WbsNode{
		ID:                      w.ID,
		ProjectID:               w.ProjectID,
		ParentNodeID:            w.ParentNodeID,
		WbsDepthLevel:           w.WbsDepthLevel,
		NodeCode:                w.NodeCode,
		Title:                   w.Title,
		NodeType:                domain.WbsNodeType(w.NodeType),
		EstimatedHours:          w.EstimatedHours,
		BudgetRevenueFunctional: w.BudgetRevenueFunctional,
		IsCompleted:             w.IsCompleted,
		Version:                 w.Version,
		CreatedAt:               w.CreatedAt,
		UpdatedAt:               w.UpdatedAt,
	}
}

func FromWbsNodeDomain(w *domain.WbsNode) *WbsNode {
	if w == nil {
		return nil
	}
	return &WbsNode{
		ID:                      w.ID,
		ProjectID:               w.ProjectID,
		ParentNodeID:            w.ParentNodeID,
		WbsDepthLevel:           w.WbsDepthLevel,
		NodeCode:                w.NodeCode,
		Title:                   w.Title,
		NodeType:                string(w.NodeType),
		EstimatedHours:          w.EstimatedHours,
		BudgetRevenueFunctional: w.BudgetRevenueFunctional,
		IsCompleted:             w.IsCompleted,
		Version:                 w.Version,
		CreatedAt:               w.CreatedAt,
		UpdatedAt:               w.UpdatedAt,
	}
}

// 3. TimeLog
type TimeLog struct {
	ID               string          `gorm:"primaryKey;type:uuid"`
	LegalEntityID    string          `gorm:"type:uuid;not null"`
	WbsNodeID        string          `gorm:"type:uuid;not null;uniqueIndex:idx_prj_timelog_unique"`
	EmployeeID       string          `gorm:"type:uuid;not null;uniqueIndex:idx_prj_timelog_unique"`
	WorkDate         time.Time       `gorm:"type:date;not null;uniqueIndex:idx_prj_timelog_unique"`
	HoursSpent       decimal.Decimal `gorm:"type:numeric(6,2);not null"`
	InternalCostRate decimal.Decimal `gorm:"type:numeric(12,4);not null"`
	BillingRate      decimal.Decimal `gorm:"type:numeric(12,4);not null"`
	IsBillable       bool            `gorm:"type:boolean;not null"`
	IsApproved       bool            `gorm:"type:boolean;not null;default:false"`
	ApprovedByHrID   *string         `gorm:"type:uuid;default:null"`
	CreatedAt        time.Time       `gorm:"not null"`
}

func (TimeLog) TableName() string {
	return "prj_time_logs"
}

func ToTimeLogDomain(t *TimeLog) *domain.TimeLog {
	if t == nil {
		return nil
	}
	return &domain.TimeLog{
		ID:               t.ID,
		LegalEntityID:    t.LegalEntityID,
		WbsNodeID:        t.WbsNodeID,
		EmployeeID:       t.EmployeeID,
		WorkDate:         t.WorkDate,
		HoursSpent:       t.HoursSpent,
		InternalCostRate: t.InternalCostRate,
		BillingRate:      t.BillingRate,
		IsBillable:       t.IsBillable,
		IsApproved:       t.IsApproved,
		ApprovedByHrID:   t.ApprovedByHrID,
		CreatedAt:        t.CreatedAt,
	}
}

func FromTimeLogDomain(t *domain.TimeLog) *TimeLog {
	if t == nil {
		return nil
	}
	return &TimeLog{
		ID:               t.ID,
		LegalEntityID:    t.LegalEntityID,
		WbsNodeID:        t.WbsNodeID,
		EmployeeID:       t.EmployeeID,
		WorkDate:         t.WorkDate,
		HoursSpent:       t.HoursSpent,
		InternalCostRate: t.InternalCostRate,
		BillingRate:      t.BillingRate,
		IsBillable:       t.IsBillable,
		IsApproved:       t.IsApproved,
		ApprovedByHrID:   t.ApprovedByHrID,
		CreatedAt:        t.CreatedAt,
	}
}

// 4. TransactionalOutbox
type TransactionalOutbox struct {
	ID          string    `gorm:"primaryKey;type:uuid"`
	EventType   string    `gorm:"type:varchar(255);not null"`
	AggregateID string    `gorm:"type:uuid;not null"`
	Payload     []byte    `gorm:"type:jsonb;not null"`
	Status      string    `gorm:"type:varchar(50);not null;index:idx_prj_outbox_status_date"`
	CreatedAt   time.Time `gorm:"not null;index:idx_prj_outbox_status_date"`
}

func (TransactionalOutbox) TableName() string {
	return "prj_transactional_outbox"
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
	return &TransactionalOutbox{
		ID:          o.ID,
		EventType:   o.EventType,
		AggregateID: o.AggregateID,
		Payload:     payloadBytes,
		Status:      string(o.Status),
		CreatedAt:   o.CreatedAt,
	}
}

// 5. KafkaEventInbox
type KafkaEventInbox struct {
	EventID          string    `gorm:"primaryKey;type:uuid"`
	EventType        string    `gorm:"type:varchar(255);not null"`
	ProcessedAt      time.Time `gorm:"not null"`
	ProcessingStatus string    `gorm:"type:varchar(50);not null"`
	Payload          []byte    `gorm:"type:jsonb;not null"`
}

func (KafkaEventInbox) TableName() string {
	return "prj_kafka_event_inbox"
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
		EventID:          i.EventID,
		EventType:        i.EventType,
		ProcessedAt:      i.ProcessedAt,
		ProcessingStatus: string(i.ProcessingStatus),
		Payload:          payloadBytes,
	}
}
