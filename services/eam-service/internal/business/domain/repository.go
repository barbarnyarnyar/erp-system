package domain

import "context"

type EventPublisher interface {
	Publish(ctx context.Context, topic string, key string, payload interface{}) error
}

type FacilityRepository interface {
	Create(ctx context.Context, f *Facility) error
	GetByID(ctx context.Context, id string) (*Facility, error)
	List(ctx context.Context) ([]Facility, error)
	Update(ctx context.Context, f *Facility) error
}

type EquipmentRepository interface {
	Create(ctx context.Context, eq *Equipment) error
	GetByID(ctx context.Context, id string) (*Equipment, error)
	List(ctx context.Context) ([]Equipment, error)
	Update(ctx context.Context, eq *Equipment) error
	ListByTenant(ctx context.Context, legalEntityId string) ([]Equipment, error)
}

type MaintenanceWorkOrderRepository interface {
	Create(ctx context.Context, wo *MaintenanceWorkOrder) error
	GetByID(ctx context.Context, id string) (*MaintenanceWorkOrder, error)
	List(ctx context.Context) ([]MaintenanceWorkOrder, error)
	Update(ctx context.Context, wo *MaintenanceWorkOrder) error
}

type PreventativeScheduleRepository interface {
	Create(ctx context.Context, ps *PreventativeSchedule) error
	GetByID(ctx context.Context, id string) (*PreventativeSchedule, error)
	List(ctx context.Context) ([]PreventativeSchedule, error)
	Update(ctx context.Context, ps *PreventativeSchedule) error
}

type TelemetryIngestBufferRepository interface {
	Create(ctx context.Context, tb *TelemetryIngestBuffer) error
	List(ctx context.Context) ([]TelemetryIngestBuffer, error)
	DeleteBatch(ctx context.Context, ids []string) error
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
