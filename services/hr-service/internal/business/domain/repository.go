package domain

import "context"

type DepartmentRepository interface {
	Create(ctx context.Context, dept *Department) error
	GetByID(ctx context.Context, id string) (*Department, error)
	GetByCode(ctx context.Context, legalEntityID, code string) (*Department, error)
	List(ctx context.Context) ([]Department, error)
	Update(ctx context.Context, dept *Department) error
}

type EmployeeMasterRepository interface {
	Create(ctx context.Context, emp *EmployeeMaster) error
	GetByID(ctx context.Context, id string) (*EmployeeMaster, error)
	GetByNumber(ctx context.Context, legalEntityID, number string) (*EmployeeMaster, error)
	GetByEmail(ctx context.Context, email string) (*EmployeeMaster, error)
	List(ctx context.Context) ([]EmployeeMaster, error)
	Update(ctx context.Context, emp *EmployeeMaster) error
	Delete(ctx context.Context, id string) error
}

type PayrollRunRepository interface {
	Create(ctx context.Context, run *PayrollRun) error
	GetByID(ctx context.Context, id string) (*PayrollRun, error)
	GetByPeriod(ctx context.Context, legalEntityID string, year, period int) (*PayrollRun, error)
	List(ctx context.Context) ([]PayrollRun, error)
	Update(ctx context.Context, run *PayrollRun) error
}

type ExpenseClaimRepository interface {
	Create(ctx context.Context, claim *ExpenseClaim) error
	GetByID(ctx context.Context, id string) (*ExpenseClaim, error)
	GetByNumber(ctx context.Context, legalEntityID, number string) (*ExpenseClaim, error)
	List(ctx context.Context) ([]ExpenseClaim, error)
	Update(ctx context.Context, claim *ExpenseClaim) error
}

type ExpenseClaimLineRepository interface {
	Create(ctx context.Context, line *ExpenseClaimLine) error
	GetByID(ctx context.Context, id string) (*ExpenseClaimLine, error)
	ListByClaimID(ctx context.Context, claimID string) ([]ExpenseClaimLine, error)
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
