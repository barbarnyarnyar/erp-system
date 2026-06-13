package domain

import (
	"context"
)

type ProjectRepository interface {
	Create(ctx context.Context, project *Project) error
	GetByID(ctx context.Context, id string) (*Project, error)
	List(ctx context.Context) ([]Project, error)
	Update(ctx context.Context, project *Project) error
	Delete(ctx context.Context, id string) error
}

type WbsNodeRepository interface {
	Create(ctx context.Context, node *WbsNode) error
	GetByID(ctx context.Context, id string) (*WbsNode, error)
	ListByProjectID(ctx context.Context, projectID string) ([]WbsNode, error)
	Update(ctx context.Context, node *WbsNode) error
	Delete(ctx context.Context, id string) error
}

type TimeLogRepository interface {
	Create(ctx context.Context, log *TimeLog) error
	GetByID(ctx context.Context, id string) (*TimeLog, error)
	List(ctx context.Context) ([]TimeLog, error)
	Update(ctx context.Context, log *TimeLog) error
	Delete(ctx context.Context, id string) error
	ApproveTimeLogs(ctx context.Context, ids []string, approverHrID string) error
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
