package domain

import "context"

type EventPublisher interface {
	Publish(ctx context.Context, topic string, key string, payload interface{}) error
}

type MaterialMasterRepository interface {
	Create(ctx context.Context, m *MaterialMaster) error
	GetByID(ctx context.Context, id string) (*MaterialMaster, error)
	GetBySKU(ctx context.Context, legalEntityId string, sku string) (*MaterialMaster, error)
	List(ctx context.Context) ([]MaterialMaster, error)
	Update(ctx context.Context, m *MaterialMaster) error
}

type BomHeaderRepository interface {
	Create(ctx context.Context, bh *BomHeader) error
	GetByID(ctx context.Context, id string) (*BomHeader, error)
	List(ctx context.Context) ([]BomHeader, error)
	Update(ctx context.Context, bh *BomHeader) error
}

type BomLineRepository interface {
	Create(ctx context.Context, bl *BomLine) error
	GetByID(ctx context.Context, id string) (*BomLine, error)
	ListByHeaderID(ctx context.Context, headerID string) ([]BomLine, error)
	DeleteByHeaderID(ctx context.Context, headerID string) error
}

type EngineeringChangeOrderRepository interface {
	Create(ctx context.Context, eco *EngineeringChangeOrder) error
	GetByID(ctx context.Context, id string) (*EngineeringChangeOrder, error)
	List(ctx context.Context) ([]EngineeringChangeOrder, error)
	Update(ctx context.Context, eco *EngineeringChangeOrder) error
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
