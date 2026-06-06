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

// BudgetRepository defines operations for budgets
type BudgetRepository interface {
	Create(ctx context.Context, budget *Budget) error
	GetByID(ctx context.Context, id string) (*Budget, error)
	Update(ctx context.Context, budget *Budget) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]Budget, error)
	GetByAccountAndPeriod(ctx context.Context, accountID string, fiscalYear int, period int) (*Budget, error)
}

// VendorBillRepository defines operations for vendor bills (Accounts Payable)
type VendorBillRepository interface {
	Create(ctx context.Context, bill *VendorBill, lines []VendorBillLine) error
	GetByID(ctx context.Context, id string) (*VendorBill, []VendorBillLine, error)
	Update(ctx context.Context, bill *VendorBill) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]VendorBill, error)
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

// TransactionRepository defines operations for transactions
type TransactionRepository interface {
	Create(ctx context.Context, tx *Transaction, lines []TransactionLine) error
	GetByID(ctx context.Context, id string) (*Transaction, []TransactionLine, error)
	Update(ctx context.Context, tx *Transaction, lines []TransactionLine) error
	List(ctx context.Context) ([]Transaction, error)
}
