package domain

import "context"

type EventPublisher interface {
	Publish(ctx context.Context, topic string, key string, payload interface{}) error
}

type InspectionPlanRepository interface {
	Create(ctx context.Context, ip *InspectionPlan) error
	GetByID(ctx context.Context, id string) (*InspectionPlan, error)
	GetByMaterial(ctx context.Context, legalEntityId string, materialId string) (*InspectionPlan, error)
	List(ctx context.Context) ([]InspectionPlan, error)
	Update(ctx context.Context, ip *InspectionPlan) error
}

type InspectionMetricDefinitionRepository interface {
	Create(ctx context.Context, imd *InspectionMetricDefinition) error
	GetByID(ctx context.Context, id string) (*InspectionMetricDefinition, error)
	ListByPlanID(ctx context.Context, planID string) ([]InspectionMetricDefinition, error)
}

type QualityInspectionRepository interface {
	Create(ctx context.Context, qi *QualityInspection) error
	GetByID(ctx context.Context, id string) (*QualityInspection, error)
	List(ctx context.Context) ([]QualityInspection, error)
	Update(ctx context.Context, qi *QualityInspection) error
}

type InspectionResultLineRepository interface {
	Create(ctx context.Context, irl *InspectionResultLine) error
	ListByInspectionID(ctx context.Context, inspectionID string) ([]InspectionResultLine, error)
}

type NonConformanceLogRepository interface {
	Create(ctx context.Context, ncl *NonConformanceLog) error
	GetByID(ctx context.Context, id string) (*NonConformanceLog, error)
	List(ctx context.Context) ([]NonConformanceLog, error)
	Update(ctx context.Context, ncl *NonConformanceLog) error
}

type TransactionalOutboxRepository interface {
	Create(ctx context.Context, outbox *TransactionalOutbox) error
	GetUnsent(ctx context.Context, limit int) ([]TransactionalOutbox, error)
	UpdateStatus(ctx context.Context, id string, status OutboxStatus) error
}

type KafkaEventInboxRepository interface {
	Create(ctx context.Context, inbox *KafkaEventInbox) error
	Exists(ctx context.Context, eventID string) (bool, error)
}
