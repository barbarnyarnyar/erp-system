package domain

import (
	"context"
)

// ChartOfAccountsRepository defines operations for chart of accounts
type ChartOfAccountsRepository interface {
	Create(ctx context.Context, coa *ChartOfAccounts) error
	GetByID(ctx context.Context, id string) (*ChartOfAccounts, error)
	GetByCode(ctx context.Context, legalEntityID, accountCode string) (*ChartOfAccounts, error)
	Update(ctx context.Context, coa *ChartOfAccounts) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]ChartOfAccounts, error)
}

// UniversalJournalEntryRepository defines operations for universal journal entries
type UniversalJournalEntryRepository interface {
	Create(ctx context.Context, entry *UniversalJournalEntry, lines []UniversalJournalLine) error
	GetByID(ctx context.Context, id string) (*UniversalJournalEntry, []UniversalJournalLine, error)
	Update(ctx context.Context, entry *UniversalJournalEntry, lines []UniversalJournalLine) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]UniversalJournalEntry, error)
}

// ArInvoiceRepository defines operations for customer invoices (Accounts Receivable)
type ArInvoiceRepository interface {
	Create(ctx context.Context, invoice *ArInvoice) error
	GetByID(ctx context.Context, id string) (*ArInvoice, error)
	GetByNumber(ctx context.Context, invoiceNumber string) (*ArInvoice, error)
	Update(ctx context.Context, invoice *ArInvoice) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]ArInvoice, error)
}

// PaymentRepository defines operations for payments
type PaymentRepository interface {
	Create(ctx context.Context, payment *Payment) error
	GetByID(ctx context.Context, id string) (*Payment, error)
	List(ctx context.Context) ([]Payment, error)
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

// ApVendorBillRepository defines operations for vendor bills (Accounts Payable)
type ApVendorBillRepository interface {
	Create(ctx context.Context, bill *ApVendorBill) error
	GetByID(ctx context.Context, id string) (*ApVendorBill, error)
	GetByNumber(ctx context.Context, billNumber string) (*ApVendorBill, error)
	Update(ctx context.Context, bill *ApVendorBill) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]ApVendorBill, error)
}

// TaxRateRepository defines operations for tax rates
type TaxRateRepository interface {
	Create(ctx context.Context, tr *TaxRate) error
	GetByID(ctx context.Context, id string) (*TaxRate, error)
	List(ctx context.Context) ([]TaxRate, error)
}

// CurrencyRateRepository defines operations for currency rates
type CurrencyRateRepository interface {
	Create(ctx context.Context, rate *CurrencyRate) error
	GetByID(ctx context.Context, id string) (*CurrencyRate, error)
	List(ctx context.Context) ([]CurrencyRate, error)
}

// FiscalYearRepository defines operations for fiscal years
type FiscalYearRepository interface {
	Create(ctx context.Context, fy *FiscalYear) error
	GetByID(ctx context.Context, id string) (*FiscalYear, error)
	Update(ctx context.Context, fy *FiscalYear) error
	List(ctx context.Context) ([]FiscalYear, error)
}

// CostCenterRepository defines operations for cost centers
type CostCenterRepository interface {
	Create(ctx context.Context, cc *CostCenter) error
	GetByID(ctx context.Context, id string) (*CostCenter, error)
	Update(ctx context.Context, cc *CostCenter) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]CostCenter, error)
}

// BankAccountRepository defines operations for bank accounts
type BankAccountRepository interface {
	Create(ctx context.Context, ba *BankAccount) error
	GetByID(ctx context.Context, id string) (*BankAccount, error)
	Update(ctx context.Context, ba *BankAccount) error
	List(ctx context.Context) ([]BankAccount, error)
}

// CustomerCreditRepository defines operations for customer credits
type CustomerCreditRepository interface {
	Create(ctx context.Context, cc *CustomerCredit) error
	GetByID(ctx context.Context, id string) (*CustomerCredit, error)
	Update(ctx context.Context, cc *CustomerCredit) error
	List(ctx context.Context) ([]CustomerCredit, error)
}

// BankStatementRepository defines operations for bank statements
type BankStatementRepository interface {
	Create(ctx context.Context, bs *BankStatement, lines []BankStatementLine) error
	GetByID(ctx context.Context, id string) (*BankStatement, []BankStatementLine, error)
	Update(ctx context.Context, bs *BankStatement, lines []BankStatementLine) error
	List(ctx context.Context) ([]BankStatement, error)
}

// TransactionManager defines an interface for running operations within a database transaction
type TransactionManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// TransactionalOutboxRepository defines operations for the transactional outbox
type TransactionalOutboxRepository interface {
	Create(ctx context.Context, record *TransactionalOutbox) error
	GetPending(ctx context.Context, limit int) ([]TransactionalOutbox, error)
	UpdateStatus(ctx context.Context, id string, status OutboxStatus) error
}

