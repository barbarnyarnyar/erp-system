package memory

import (
	"context"
	"errors"
	"sync"

	"github.com/erp-system/fm-service/internal/business/domain"
)

// MemoryAccountRepo implements domain.AccountRepository in-memory
type MemoryAccountRepo struct {
	mu       sync.RWMutex
	accounts map[string]domain.Account
}

func NewMemoryAccountRepo() *MemoryAccountRepo {
	return &MemoryAccountRepo{
		accounts: make(map[string]domain.Account),
	}
}

func (r *MemoryAccountRepo) Create(ctx context.Context, account *domain.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.accounts[account.ID]; ok {
		return errors.New("account already exists")
	}
	r.accounts[account.ID] = *account
	return nil
}

func (r *MemoryAccountRepo) GetByID(ctx context.Context, id string) (*domain.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	acc, ok := r.accounts[id]
	if !ok {
		return nil, errors.New("account not found")
	}
	return &acc, nil
}

func (r *MemoryAccountRepo) GetByNumber(ctx context.Context, accountNumber string) (*domain.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, acc := range r.accounts {
		if acc.AccountNumber == accountNumber {
			return &acc, nil
		}
	}
	return nil, errors.New("account not found")
}

func (r *MemoryAccountRepo) Update(ctx context.Context, account *domain.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.accounts[account.ID]; !ok {
		return errors.New("account not found")
	}
	r.accounts[account.ID] = *account
	return nil
}

func (r *MemoryAccountRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.accounts[id]; !ok {
		return errors.New("account not found")
	}
	delete(r.accounts, id)
	return nil
}

func (r *MemoryAccountRepo) List(ctx context.Context) ([]domain.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]domain.Account, 0, len(r.accounts))
	for _, acc := range r.accounts {
		list = append(list, acc)
	}
	return list, nil
}

// MemoryJournalEntryRepo implements domain.JournalEntryRepository in-memory
type MemoryJournalEntryRepo struct {
	mu      sync.RWMutex
	entries map[string]domain.JournalEntry
	lines   map[string][]domain.JournalEntryLine
}

func NewMemoryJournalEntryRepo() *MemoryJournalEntryRepo {
	return &MemoryJournalEntryRepo{
		entries: make(map[string]domain.JournalEntry),
		lines:   make(map[string][]domain.JournalEntryLine),
	}
}

func (r *MemoryJournalEntryRepo) Create(ctx context.Context, entry *domain.JournalEntry, lines []domain.JournalEntryLine) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.entries[entry.ID] = *entry
	r.lines[entry.ID] = lines
	return nil
}

func (r *MemoryJournalEntryRepo) GetByID(ctx context.Context, id string) (*domain.JournalEntry, []domain.JournalEntryLine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, ok := r.entries[id]
	if !ok {
		return nil, nil, errors.New("journal entry not found")
	}
	return &entry, r.lines[id], nil
}

func (r *MemoryJournalEntryRepo) Update(ctx context.Context, entry *domain.JournalEntry, lines []domain.JournalEntryLine) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.entries[entry.ID] = *entry
	r.lines[entry.ID] = lines
	return nil
}

func (r *MemoryJournalEntryRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.entries, id)
	delete(r.lines, id)
	return nil
}

func (r *MemoryJournalEntryRepo) List(ctx context.Context) ([]domain.JournalEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]domain.JournalEntry, 0, len(r.entries))
	for _, entry := range r.entries {
		list = append(list, entry)
	}
	return list, nil
}

// MemoryInvoiceRepo implements domain.InvoiceRepository
type MemoryInvoiceRepo struct {
	mu       sync.RWMutex
	invoices map[string]domain.Invoice
	lines    map[string][]domain.InvoiceLine
}

func NewMemoryInvoiceRepo() *MemoryInvoiceRepo {
	return &MemoryInvoiceRepo{
		invoices: make(map[string]domain.Invoice),
		lines:    make(map[string][]domain.InvoiceLine),
	}
}

func (r *MemoryInvoiceRepo) Create(ctx context.Context, invoice *domain.Invoice, lines []domain.InvoiceLine) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.invoices[invoice.ID] = *invoice
	r.lines[invoice.ID] = lines
	return nil
}

func (r *MemoryInvoiceRepo) GetByID(ctx context.Context, id string) (*domain.Invoice, []domain.InvoiceLine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	inv, ok := r.invoices[id]
	if !ok {
		return nil, nil, errors.New("invoice not found")
	}
	return &inv, r.lines[id], nil
}

func (r *MemoryInvoiceRepo) Update(ctx context.Context, invoice *domain.Invoice) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.invoices[invoice.ID] = *invoice
	return nil
}

func (r *MemoryInvoiceRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.invoices, id)
	delete(r.lines, id)
	return nil
}

func (r *MemoryInvoiceRepo) List(ctx context.Context) ([]domain.Invoice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]domain.Invoice, 0, len(r.invoices))
	for _, inv := range r.invoices {
		list = append(list, inv)
	}
	return list, nil
}

// MemoryPaymentRepo implements domain.PaymentRepository
type MemoryPaymentRepo struct {
	mu       sync.RWMutex
	payments map[string]domain.Payment
}

func NewMemoryPaymentRepo() *MemoryPaymentRepo {
	return &MemoryPaymentRepo{
		payments: make(map[string]domain.Payment),
	}
}

func (r *MemoryPaymentRepo) Create(ctx context.Context, payment *domain.Payment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.payments[payment.ID] = *payment
	return nil
}

func (r *MemoryPaymentRepo) GetByID(ctx context.Context, id string) (*domain.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	pay, ok := r.payments[id]
	if !ok {
		return nil, errors.New("payment not found")
	}
	return &pay, nil
}

func (r *MemoryPaymentRepo) List(ctx context.Context) ([]domain.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]domain.Payment, 0, len(r.payments))
	for _, pay := range r.payments {
		list = append(list, pay)
	}
	return list, nil
}

// MemoryBudgetRepo implements domain.BudgetRepository
type MemoryBudgetRepo struct {
	mu      sync.RWMutex
	budgets map[string]domain.Budget
}

func NewMemoryBudgetRepo() *MemoryBudgetRepo {
	return &MemoryBudgetRepo{
		budgets: make(map[string]domain.Budget),
	}
}

func (r *MemoryBudgetRepo) Create(ctx context.Context, budget *domain.Budget) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.budgets[budget.ID] = *budget
	return nil
}

func (r *MemoryBudgetRepo) GetByID(ctx context.Context, id string) (*domain.Budget, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	bud, ok := r.budgets[id]
	if !ok {
		return nil, errors.New("budget not found")
	}
	return &bud, nil
}

func (r *MemoryBudgetRepo) Update(ctx context.Context, budget *domain.Budget) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.budgets[budget.ID] = *budget
	return nil
}

func (r *MemoryBudgetRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.budgets, id)
	return nil
}

func (r *MemoryBudgetRepo) List(ctx context.Context) ([]domain.Budget, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]domain.Budget, 0, len(r.budgets))
	for _, bud := range r.budgets {
		list = append(list, bud)
	}
	return list, nil
}

func (r *MemoryBudgetRepo) GetByAccountAndPeriod(ctx context.Context, accountID string, fiscalYear int, period int) (*domain.Budget, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, bud := range r.budgets {
		if bud.AccountID == accountID && bud.FiscalYear == fiscalYear && bud.Period == period {
			return &bud, nil
		}
	}
	return nil, errors.New("budget not found")
}

// MemoryVendorBillRepo implements domain.VendorBillRepository in-memory
type MemoryVendorBillRepo struct {
	mu    sync.RWMutex
	bills map[string]domain.VendorBill
	lines map[string][]domain.VendorBillLine
}

func NewMemoryVendorBillRepo() *MemoryVendorBillRepo {
	return &MemoryVendorBillRepo{
		bills: make(map[string]domain.VendorBill),
		lines: make(map[string][]domain.VendorBillLine),
	}
}

func (r *MemoryVendorBillRepo) Create(ctx context.Context, bill *domain.VendorBill, lines []domain.VendorBillLine) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.bills[bill.ID] = *bill
	r.lines[bill.ID] = lines
	return nil
}

func (r *MemoryVendorBillRepo) GetByID(ctx context.Context, id string) (*domain.VendorBill, []domain.VendorBillLine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	bill, ok := r.bills[id]
	if !ok {
		return nil, nil, errors.New("vendor bill not found")
	}
	return &bill, r.lines[id], nil
}

func (r *MemoryVendorBillRepo) Update(ctx context.Context, bill *domain.VendorBill) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.bills[bill.ID] = *bill
	return nil
}

func (r *MemoryVendorBillRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.bills, id)
	delete(r.lines, id)
	return nil
}

func (r *MemoryVendorBillRepo) List(ctx context.Context) ([]domain.VendorBill, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]domain.VendorBill, 0, len(r.bills))
	for _, bill := range r.bills {
		list = append(list, bill)
	}
	return list, nil
}

// MemoryTaxRateRepo implements domain.TaxRateRepository
type MemoryTaxRateRepo struct {
	mu   sync.RWMutex
	data map[string]domain.TaxRate
}

func NewMemoryTaxRateRepo() *MemoryTaxRateRepo {
	return &MemoryTaxRateRepo{
		data: make(map[string]domain.TaxRate),
	}
}

func (r *MemoryTaxRateRepo) Create(ctx context.Context, tr *domain.TaxRate) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[tr.ID] = *tr
	return nil
}

func (r *MemoryTaxRateRepo) GetByID(ctx context.Context, id string) (*domain.TaxRate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tr, ok := r.data[id]
	if !ok {
		return nil, errors.New("tax rate not found")
	}
	return &tr, nil
}

func (r *MemoryTaxRateRepo) List(ctx context.Context) ([]domain.TaxRate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.TaxRate, 0, len(r.data))
	for _, tr := range r.data {
		list = append(list, tr)
	}
	return list, nil
}

// MemoryCurrencyRateRepo implements domain.CurrencyRateRepository in-memory
type MemoryCurrencyRateRepo struct {
	mu   sync.RWMutex
	data map[string]domain.CurrencyRate
}

func NewMemoryCurrencyRateRepo() *MemoryCurrencyRateRepo {
	return &MemoryCurrencyRateRepo{
		data: make(map[string]domain.CurrencyRate),
	}
}

func (r *MemoryCurrencyRateRepo) Create(ctx context.Context, rate *domain.CurrencyRate) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[rate.ID] = *rate
	return nil
}

func (r *MemoryCurrencyRateRepo) GetByID(ctx context.Context, id string) (*domain.CurrencyRate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rate, ok := r.data[id]
	if !ok {
		return nil, errors.New("currency rate not found")
	}
	return &rate, nil
}

func (r *MemoryCurrencyRateRepo) List(ctx context.Context) ([]domain.CurrencyRate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.CurrencyRate, 0, len(r.data))
	for _, rate := range r.data {
		list = append(list, rate)
	}
	return list, nil
}

// MemoryFiscalYearRepo implements domain.FiscalYearRepository in-memory
type MemoryFiscalYearRepo struct {
	mu   sync.RWMutex
	data map[string]domain.FiscalYear
}

func NewMemoryFiscalYearRepo() *MemoryFiscalYearRepo {
	return &MemoryFiscalYearRepo{
		data: make(map[string]domain.FiscalYear),
	}
}

func (r *MemoryFiscalYearRepo) Create(ctx context.Context, fy *domain.FiscalYear) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[fy.ID] = *fy
	return nil
}

func (r *MemoryFiscalYearRepo) GetByID(ctx context.Context, id string) (*domain.FiscalYear, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	fy, ok := r.data[id]
	if !ok {
		return nil, errors.New("fiscal year not found")
	}
	return &fy, nil
}

func (r *MemoryFiscalYearRepo) Update(ctx context.Context, fy *domain.FiscalYear) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[fy.ID] = *fy
	return nil
}

func (r *MemoryFiscalYearRepo) List(ctx context.Context) ([]domain.FiscalYear, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.FiscalYear, 0, len(r.data))
	for _, fy := range r.data {
		list = append(list, fy)
	}
	return list, nil
}

// MemoryCostCenterRepo implements domain.CostCenterRepository in-memory
type MemoryCostCenterRepo struct {
	mu   sync.RWMutex
	data map[string]domain.CostCenter
}

func NewMemoryCostCenterRepo() *MemoryCostCenterRepo {
	return &MemoryCostCenterRepo{
		data: make(map[string]domain.CostCenter),
	}
}

func (r *MemoryCostCenterRepo) Create(ctx context.Context, cc *domain.CostCenter) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[cc.ID] = *cc
	return nil
}

func (r *MemoryCostCenterRepo) GetByID(ctx context.Context, id string) (*domain.CostCenter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cc, ok := r.data[id]
	if !ok {
		return nil, errors.New("cost center not found")
	}
	return &cc, nil
}

func (r *MemoryCostCenterRepo) Update(ctx context.Context, cc *domain.CostCenter) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[cc.ID] = *cc
	return nil
}

func (r *MemoryCostCenterRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, id)
	return nil
}

func (r *MemoryCostCenterRepo) List(ctx context.Context) ([]domain.CostCenter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.CostCenter, 0, len(r.data))
	for _, cc := range r.data {
		list = append(list, cc)
	}
	return list, nil
}

// MemoryBankAccountRepo implements domain.BankAccountRepository in-memory
type MemoryBankAccountRepo struct {
	mu   sync.RWMutex
	data map[string]domain.BankAccount
}

func NewMemoryBankAccountRepo() *MemoryBankAccountRepo {
	return &MemoryBankAccountRepo{
		data: make(map[string]domain.BankAccount),
	}
}

func (r *MemoryBankAccountRepo) Create(ctx context.Context, ba *domain.BankAccount) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[ba.ID] = *ba
	return nil
}

func (r *MemoryBankAccountRepo) GetByID(ctx context.Context, id string) (*domain.BankAccount, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ba, ok := r.data[id]
	if !ok {
		return nil, errors.New("bank account not found")
	}
	return &ba, nil
}

func (r *MemoryBankAccountRepo) Update(ctx context.Context, ba *domain.BankAccount) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[ba.ID] = *ba
	return nil
}

func (r *MemoryBankAccountRepo) List(ctx context.Context) ([]domain.BankAccount, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.BankAccount, 0, len(r.data))
	for _, ba := range r.data {
		list = append(list, ba)
	}
	return list, nil
}

// MemoryCustomerCreditRepo implements domain.CustomerCreditRepository in-memory
type MemoryCustomerCreditRepo struct {
	mu   sync.RWMutex
	data map[string]domain.CustomerCredit
}

func NewMemoryCustomerCreditRepo() *MemoryCustomerCreditRepo {
	return &MemoryCustomerCreditRepo{
		data: make(map[string]domain.CustomerCredit),
	}
}

func (r *MemoryCustomerCreditRepo) Create(ctx context.Context, cc *domain.CustomerCredit) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[cc.ID] = *cc
	return nil
}

func (r *MemoryCustomerCreditRepo) GetByID(ctx context.Context, id string) (*domain.CustomerCredit, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cc, ok := r.data[id]
	if !ok {
		return nil, errors.New("customer credit not found")
	}
	return &cc, nil
}

func (r *MemoryCustomerCreditRepo) Update(ctx context.Context, cc *domain.CustomerCredit) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[cc.ID] = *cc
	return nil
}

func (r *MemoryCustomerCreditRepo) List(ctx context.Context) ([]domain.CustomerCredit, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.CustomerCredit, 0, len(r.data))
	for _, cc := range r.data {
		list = append(list, cc)
	}
	return list, nil
}

// MemoryBankStatementRepo implements domain.BankStatementRepository in-memory
type MemoryBankStatementRepo struct {
	mu    sync.RWMutex
	data  map[string]domain.BankStatement
	lines map[string][]domain.BankStatementLine
}

func NewMemoryBankStatementRepo() *MemoryBankStatementRepo {
	return &MemoryBankStatementRepo{
		data:  make(map[string]domain.BankStatement),
		lines: make(map[string][]domain.BankStatementLine),
	}
}

func (r *MemoryBankStatementRepo) Create(ctx context.Context, bs *domain.BankStatement, lines []domain.BankStatementLine) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[bs.ID] = *bs
	r.lines[bs.ID] = lines
	return nil
}

func (r *MemoryBankStatementRepo) GetByID(ctx context.Context, id string) (*domain.BankStatement, []domain.BankStatementLine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	bs, ok := r.data[id]
	if !ok {
		return nil, nil, errors.New("bank statement not found")
	}
	return &bs, r.lines[id], nil
}

func (r *MemoryBankStatementRepo) Update(ctx context.Context, bs *domain.BankStatement, lines []domain.BankStatementLine) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[bs.ID] = *bs
	r.lines[bs.ID] = lines
	return nil
}

func (r *MemoryBankStatementRepo) List(ctx context.Context) ([]domain.BankStatement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.BankStatement, 0, len(r.data))
	for _, bs := range r.data {
		list = append(list, bs)
	}
	return list, nil
}

// MemoryTransactionRepo implements domain.TransactionRepository in-memory
type MemoryTransactionRepo struct {
	mu    sync.RWMutex
	data  map[string]domain.Transaction
	lines map[string][]domain.TransactionLine
}

func NewMemoryTransactionRepo() *MemoryTransactionRepo {
	return &MemoryTransactionRepo{
		data:  make(map[string]domain.Transaction),
		lines: make(map[string][]domain.TransactionLine),
	}
}

func (r *MemoryTransactionRepo) Create(ctx context.Context, tx *domain.Transaction, lines []domain.TransactionLine) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[tx.ID] = *tx
	r.lines[tx.ID] = lines
	return nil
}

func (r *MemoryTransactionRepo) GetByID(ctx context.Context, id string) (*domain.Transaction, []domain.TransactionLine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tx, ok := r.data[id]
	if !ok {
		return nil, nil, errors.New("transaction not found")
	}
	return &tx, r.lines[id], nil
}

func (r *MemoryTransactionRepo) Update(ctx context.Context, tx *domain.Transaction, lines []domain.TransactionLine) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[tx.ID] = *tx
	r.lines[tx.ID] = lines
	return nil
}

func (r *MemoryTransactionRepo) List(ctx context.Context) ([]domain.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.Transaction, 0, len(r.data))
	for _, tx := range r.data {
		list = append(list, tx)
	}
	return list, nil
}
