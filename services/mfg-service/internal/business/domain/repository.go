package domain

import "context"

type WorkCenterRepository interface {
	Create(ctx context.Context, wc *WorkCenter) error
	GetByID(ctx context.Context, id string) (*WorkCenter, error)
	GetByCode(ctx context.Context, legalEntityID, code string) (*WorkCenter, error)
	List(ctx context.Context) ([]WorkCenter, error)
	Update(ctx context.Context, wc *WorkCenter) error
	Delete(ctx context.Context, id string) error
}

type RoutingStationRepository interface {
	Create(ctx context.Context, station *RoutingStation) error
	GetByID(ctx context.Context, id string) (*RoutingStation, error)
	GetByCode(ctx context.Context, workCenterID, code string) (*RoutingStation, error)
	ListByWorkCenterID(ctx context.Context, workCenterID string) ([]RoutingStation, error)
	Update(ctx context.Context, station *RoutingStation) error
	Delete(ctx context.Context, id string) error
}

type WorkOrderRepository interface {
	Create(ctx context.Context, wo *WorkOrder) error
	GetByID(ctx context.Context, id string) (*WorkOrder, error)
	GetByNumber(ctx context.Context, legalEntityID, number string) (*WorkOrder, error)
	List(ctx context.Context) ([]WorkOrder, error)
	Update(ctx context.Context, wo *WorkOrder) error
	Delete(ctx context.Context, id string) error
}

type WorkOrderRoutingStateRepository interface {
	Create(ctx context.Context, state *WorkOrderRoutingState) error
	GetByID(ctx context.Context, id string) (*WorkOrderRoutingState, error)
	GetActiveByWorkOrderID(ctx context.Context, workOrderID string) (*WorkOrderRoutingState, error)
	ListByWorkOrderID(ctx context.Context, workOrderID string) ([]WorkOrderRoutingState, error)
	Update(ctx context.Context, state *WorkOrderRoutingState) error
}

type MaterialConsumptionLogRepository interface {
	Create(ctx context.Context, log *MaterialConsumptionLog) error
	GetByID(ctx context.Context, id string) (*MaterialConsumptionLog, error)
	ListByWorkOrderID(ctx context.Context, workOrderID string) ([]MaterialConsumptionLog, error)
}

type ProductionYieldLogRepository interface {
	Create(ctx context.Context, log *ProductionYieldLog) error
	GetByID(ctx context.Context, id string) (*ProductionYieldLog, error)
	ListByWorkOrderID(ctx context.Context, workOrderID string) ([]ProductionYieldLog, error)
}

type TransactionalOutboxRepository interface {
	Create(ctx context.Context, msg *TransactionalOutbox) error
	GetByID(ctx context.Context, id string) (*TransactionalOutbox, error)
	GetUnsent(ctx context.Context, limit int) ([]TransactionalOutbox, error)
	Update(ctx context.Context, msg *TransactionalOutbox) error
}

type KafkaEventInboxRepository interface {
	Create(ctx context.Context, msg *KafkaEventInbox) error
	GetByID(ctx context.Context, eventID string) (*KafkaEventInbox, error)
	Update(ctx context.Context, msg *KafkaEventInbox) error
}
