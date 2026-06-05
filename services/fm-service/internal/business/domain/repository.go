package domain

import (
	"context"
)

// AccountRepository defines operations for GL accounts
type AccountRepository interface {
	Create(ctx context.Context, account *Account) error
	GetByID(ctx context.Context, id string) (*Account, error)
	GetByNumber(ctx context.Context, accountNumber string) (*Account, error)
	Update(ctx context.Context, account *Account) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]Account, error)
}

// JournalEntryRepository defines operations for journal entries
type JournalEntryRepository interface {
	Create(ctx context.Context, entry *JournalEntry, lines []JournalEntryLine) error
	GetByID(ctx context.Context, id string) (*JournalEntry, []JournalEntryLine, error)
	Update(ctx context.Context, entry *JournalEntry, lines []JournalEntryLine) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]JournalEntry, error)
}

// InvoiceRepository defines operations for customer invoices
type InvoiceRepository interface {
	Create(ctx context.Context, invoice *Invoice, lines []InvoiceLine) error
	GetByID(ctx context.Context, id string) (*Invoice, []InvoiceLine, error)
	Update(ctx context.Context, invoice *Invoice) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]Invoice, error)
}

// PaymentRepository defines operations for payments
type PaymentRepository interface {
	Create(ctx context.Context, payment *Payment) error
	GetByID(ctx context.Context, id string) (*Payment, error)
	List(ctx context.Context) ([]Payment, error)
}

// VendorRepository defines operations for vendors
type VendorRepository interface {
	Create(ctx context.Context, vendor *Vendor) error
	GetByID(ctx context.Context, id string) (*Vendor, error)
	Update(ctx context.Context, vendor *Vendor) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]Vendor, error)
}

// BudgetRepository defines operations for budgets
type BudgetRepository interface {
	Create(ctx context.Context, budget *Budget) error
	GetByID(ctx context.Context, id string) (*Budget, error)
	Update(ctx context.Context, budget *Budget) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]Budget, error)
	GetByAccountAndPeriod(ctx context.Context, accountID string, fiscalYear int, period int) (*Budget, error)
}
