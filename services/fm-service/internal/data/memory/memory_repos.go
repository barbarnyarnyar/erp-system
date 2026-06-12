package memory

import (
	"context"
	"errors"
	"sync"

	"github.com/erp-system/fm-service/internal/business/domain"
)

// Snapshotable defines the interface for repositories that support transactional snapshots.
type Snapshotable interface {
	TakeSnapshot()
	RollbackSnapshot()
	CommitSnapshot()
}

// MemoryChartOfAccountsRepo implements domain.ChartOfAccountsRepository in-memory
type MemoryChartOfAccountsRepo struct {
	mu        sync.RWMutex
	accounts  map[string]domain.ChartOfAccounts
	snapshots []map[string]domain.ChartOfAccounts
}

func NewMemoryChartOfAccountsRepo() *MemoryChartOfAccountsRepo {
	return &MemoryChartOfAccountsRepo{
		accounts: make(map[string]domain.ChartOfAccounts),
	}
}

func (r *MemoryChartOfAccountsRepo) TakeSnapshot() {
	r.mu.Lock()
	snap := make(map[string]domain.ChartOfAccounts, len(r.accounts))
	for k, v := range r.accounts {
		snap[k] = v
	}
	r.snapshots = append(r.snapshots, snap)
	r.mu.Unlock()
}

func (r *MemoryChartOfAccountsRepo) RollbackSnapshot() {
	r.mu.Lock()
	if len(r.snapshots) > 0 {
		r.accounts = r.snapshots[len(r.snapshots)-1]
		r.snapshots = r.snapshots[:len(r.snapshots)-1]
	}
	r.mu.Unlock()
}

func (r *MemoryChartOfAccountsRepo) CommitSnapshot() {
	r.mu.Lock()
	if len(r.snapshots) > 0 {
		r.snapshots = r.snapshots[:len(r.snapshots)-1]
	}
	r.mu.Unlock()
}

func (r *MemoryChartOfAccountsRepo) Create(ctx context.Context, coa *domain.ChartOfAccounts) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.accounts[coa.ID]; ok {
		return errors.New("chart of accounts already exists")
	}
	r.accounts[coa.ID] = *coa
	return nil
}

func (r *MemoryChartOfAccountsRepo) GetByID(ctx context.Context, id string) (*domain.ChartOfAccounts, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	coa, ok := r.accounts[id]
	if !ok {
		return nil, errors.New("chart of accounts not found")
	}
	return &coa, nil
}

func (r *MemoryChartOfAccountsRepo) GetByCode(ctx context.Context, legalEntityID, accountCode string) (*domain.ChartOfAccounts, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, coa := range r.accounts {
		if coa.LegalEntityID == legalEntityID && coa.AccountCode == accountCode {
			return &coa, nil
		}
	}
	return nil, errors.New("chart of accounts not found")
}

func (r *MemoryChartOfAccountsRepo) Update(ctx context.Context, coa *domain.ChartOfAccounts) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.accounts[coa.ID]; !ok {
		return errors.New("chart of accounts not found")
	}
	r.accounts[coa.ID] = *coa
	return nil
}

func (r *MemoryChartOfAccountsRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.accounts[id]; !ok {
		return errors.New("chart of accounts not found")
	}
	delete(r.accounts, id)
	return nil
}

func (r *MemoryChartOfAccountsRepo) List(ctx context.Context) ([]domain.ChartOfAccounts, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.ChartOfAccounts, 0, len(r.accounts))
	for _, coa := range r.accounts {
		list = append(list, coa)
	}
	return list, nil
}

type universalJournalEntryRepoSnapshot struct {
	entries map[string]domain.UniversalJournalEntry
	lines   map[string][]domain.UniversalJournalLine
}

// MemoryUniversalJournalEntryRepo implements domain.UniversalJournalEntryRepository in-memory
type MemoryUniversalJournalEntryRepo struct {
	mu        sync.RWMutex
	entries   map[string]domain.UniversalJournalEntry
	lines     map[string][]domain.UniversalJournalLine
	snapshots []universalJournalEntryRepoSnapshot
}

func NewMemoryUniversalJournalEntryRepo() *MemoryUniversalJournalEntryRepo {
	return &MemoryUniversalJournalEntryRepo{
		entries: make(map[string]domain.UniversalJournalEntry),
		lines:   make(map[string][]domain.UniversalJournalLine),
	}
}

func (r *MemoryUniversalJournalEntryRepo) TakeSnapshot() {
	r.mu.Lock()
	snapEntries := make(map[string]domain.UniversalJournalEntry, len(r.entries))
	for k, v := range r.entries {
		snapEntries[k] = v
	}
	snapLines := make(map[string][]domain.UniversalJournalLine, len(r.lines))
	for k, v := range r.lines {
		if v != nil {
			clonedLines := make([]domain.UniversalJournalLine, len(v))
			copy(clonedLines, v)
			snapLines[k] = clonedLines
		} else {
			snapLines[k] = nil
		}
	}
	r.snapshots = append(r.snapshots, universalJournalEntryRepoSnapshot{
		entries: snapEntries,
		lines:   snapLines,
	})
	r.mu.Unlock()
}

func (r *MemoryUniversalJournalEntryRepo) RollbackSnapshot() {
	r.mu.Lock()
	if len(r.snapshots) > 0 {
		snap := r.snapshots[len(r.snapshots)-1]
		r.entries = snap.entries
		r.lines = snap.lines
		r.snapshots = r.snapshots[:len(r.snapshots)-1]
	}
	r.mu.Unlock()
}

func (r *MemoryUniversalJournalEntryRepo) CommitSnapshot() {
	r.mu.Lock()
	if len(r.snapshots) > 0 {
		r.snapshots = r.snapshots[:len(r.snapshots)-1]
	}
	r.mu.Unlock()
}

func (r *MemoryUniversalJournalEntryRepo) Create(ctx context.Context, entry *domain.UniversalJournalEntry, lines []domain.UniversalJournalLine) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries[entry.ID] = *entry
	r.lines[entry.ID] = lines
	return nil
}

func (r *MemoryUniversalJournalEntryRepo) GetByID(ctx context.Context, id string) (*domain.UniversalJournalEntry, []domain.UniversalJournalLine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	entry, ok := r.entries[id]
	if !ok {
		return nil, nil, errors.New("universal journal entry not found")
	}
	return &entry, r.lines[id], nil
}

func (r *MemoryUniversalJournalEntryRepo) Update(ctx context.Context, entry *domain.UniversalJournalEntry, lines []domain.UniversalJournalLine) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries[entry.ID] = *entry
	r.lines[entry.ID] = lines
	return nil
}

func (r *MemoryUniversalJournalEntryRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.entries, id)
	delete(r.lines, id)
	return nil
}

func (r *MemoryUniversalJournalEntryRepo) List(ctx context.Context) ([]domain.UniversalJournalEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.UniversalJournalEntry, 0, len(r.entries))
	for _, entry := range r.entries {
		list = append(list, entry)
	}
	return list, nil
}

// MemoryArInvoiceRepo implements domain.ArInvoiceRepository
type MemoryArInvoiceRepo struct {
	mu        sync.RWMutex
	invoices  map[string]domain.ArInvoice
	snapshots []map[string]domain.ArInvoice
}

func NewMemoryArInvoiceRepo() *MemoryArInvoiceRepo {
	return &MemoryArInvoiceRepo{
		invoices: make(map[string]domain.ArInvoice),
	}
}

func (r *MemoryArInvoiceRepo) TakeSnapshot() {
	r.mu.Lock()
	snapInvoices := make(map[string]domain.ArInvoice, len(r.invoices))
	for k, v := range r.invoices {
		snapInvoices[k] = v
	}
	r.snapshots = append(r.snapshots, snapInvoices)
	r.mu.Unlock()
}

func (r *MemoryArInvoiceRepo) RollbackSnapshot() {
	r.mu.Lock()
	if len(r.snapshots) > 0 {
		r.invoices = r.snapshots[len(r.snapshots)-1]
		r.snapshots = r.snapshots[:len(r.snapshots)-1]
	}
	r.mu.Unlock()
}

func (r *MemoryArInvoiceRepo) CommitSnapshot() {
	r.mu.Lock()
	if len(r.snapshots) > 0 {
		r.snapshots = r.snapshots[:len(r.snapshots)-1]
	}
	r.mu.Unlock()
}

func (r *MemoryArInvoiceRepo) Create(ctx context.Context, invoice *domain.ArInvoice) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.invoices[invoice.ID] = *invoice
	return nil
}

func (r *MemoryArInvoiceRepo) GetByID(ctx context.Context, id string) (*domain.ArInvoice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	inv, ok := r.invoices[id]
	if !ok {
		return nil, errors.New("ar invoice not found")
	}
	return &inv, nil
}

func (r *MemoryArInvoiceRepo) GetByNumber(ctx context.Context, invoiceNumber string) (*domain.ArInvoice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, inv := range r.invoices {
		if inv.InvoiceNumber == invoiceNumber {
			return &inv, nil
		}
	}
	return nil, errors.New("ar invoice not found")
}

func (r *MemoryArInvoiceRepo) Update(ctx context.Context, invoice *domain.ArInvoice) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.invoices[invoice.ID] = *invoice
	return nil
}

func (r *MemoryArInvoiceRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.invoices, id)
	return nil
}

func (r *MemoryArInvoiceRepo) List(ctx context.Context) ([]domain.ArInvoice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.ArInvoice, 0, len(r.invoices))
	for _, inv := range r.invoices {
		list = append(list, inv)
	}
	return list, nil
}

// MemoryPaymentRepo implements domain.PaymentRepository
type MemoryPaymentRepo struct {
	mu        sync.RWMutex
	payments  map[string]domain.Payment
	snapshots []map[string]domain.Payment
}

func NewMemoryPaymentRepo() *MemoryPaymentRepo {
	return &MemoryPaymentRepo{
		payments: make(map[string]domain.Payment),
	}
}

func (r *MemoryPaymentRepo) TakeSnapshot() {
	r.mu.Lock()
	defer r.mu.Unlock()
	snap := make(map[string]domain.Payment, len(r.payments))
	for k, v := range r.payments {
		snap[k] = v
	}
	r.snapshots = append(r.snapshots, snap)
}

func (r *MemoryPaymentRepo) RollbackSnapshot() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.snapshots) == 0 {
		return
	}
	r.payments = r.snapshots[len(r.snapshots)-1]
	r.snapshots = r.snapshots[:len(r.snapshots)-1]
}

func (r *MemoryPaymentRepo) CommitSnapshot() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.snapshots) == 0 {
		return
	}
	r.snapshots = r.snapshots[:len(r.snapshots)-1]
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
	mu        sync.RWMutex
	budgets   map[string]domain.Budget
	snapshots []map[string]domain.Budget
}

func NewMemoryBudgetRepo() *MemoryBudgetRepo {
	return &MemoryBudgetRepo{
		budgets: make(map[string]domain.Budget),
	}
}

func (r *MemoryBudgetRepo) TakeSnapshot() {
	r.mu.Lock()
	defer r.mu.Unlock()
	snap := make(map[string]domain.Budget, len(r.budgets))
	for k, v := range r.budgets {
		snap[k] = v
	}
	r.snapshots = append(r.snapshots, snap)
}

func (r *MemoryBudgetRepo) RollbackSnapshot() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.snapshots) == 0 {
		return
	}
	r.budgets = r.snapshots[len(r.snapshots)-1]
	r.snapshots = r.snapshots[:len(r.snapshots)-1]
}

func (r *MemoryBudgetRepo) CommitSnapshot() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.snapshots) == 0 {
		return
	}
	r.snapshots = r.snapshots[:len(r.snapshots)-1]
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

// MemoryApVendorBillRepo implements domain.ApVendorBillRepository in-memory
type MemoryApVendorBillRepo struct {
	mu        sync.RWMutex
	bills     map[string]domain.ApVendorBill
	snapshots []map[string]domain.ApVendorBill
}

func NewMemoryApVendorBillRepo() *MemoryApVendorBillRepo {
	return &MemoryApVendorBillRepo{
		bills: make(map[string]domain.ApVendorBill),
	}
}

func (r *MemoryApVendorBillRepo) TakeSnapshot() {
	r.mu.Lock()
	snapBills := make(map[string]domain.ApVendorBill, len(r.bills))
	for k, v := range r.bills {
		snapBills[k] = v
	}
	r.snapshots = append(r.snapshots, snapBills)
	r.mu.Unlock()
}

func (r *MemoryApVendorBillRepo) RollbackSnapshot() {
	r.mu.Lock()
	if len(r.snapshots) > 0 {
		r.bills = r.snapshots[len(r.snapshots)-1]
		r.snapshots = r.snapshots[:len(r.snapshots)-1]
	}
	r.mu.Unlock()
}

func (r *MemoryApVendorBillRepo) CommitSnapshot() {
	r.mu.Lock()
	if len(r.snapshots) > 0 {
		r.snapshots = r.snapshots[:len(r.snapshots)-1]
	}
	r.mu.Unlock()
}

func (r *MemoryApVendorBillRepo) Create(ctx context.Context, bill *domain.ApVendorBill) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.bills[bill.ID] = *bill
	return nil
}

func (r *MemoryApVendorBillRepo) GetByID(ctx context.Context, id string) (*domain.ApVendorBill, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	bill, ok := r.bills[id]
	if !ok {
		return nil, errors.New("ap vendor bill not found")
	}
	return &bill, nil
}

func (r *MemoryApVendorBillRepo) GetByNumber(ctx context.Context, billNumber string) (*domain.ApVendorBill, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, bill := range r.bills {
		if bill.BillNumber == billNumber {
			return &bill, nil
		}
	}
	return nil, errors.New("ap vendor bill not found")
}

func (r *MemoryApVendorBillRepo) Update(ctx context.Context, bill *domain.ApVendorBill) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.bills[bill.ID] = *bill
	return nil
}

func (r *MemoryApVendorBillRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.bills, id)
	return nil
}

func (r *MemoryApVendorBillRepo) List(ctx context.Context) ([]domain.ApVendorBill, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.ApVendorBill, 0, len(r.bills))
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



// MemoryTransactionalOutboxRepo implements domain.TransactionalOutboxRepository in-memory
type MemoryTransactionalOutboxRepo struct {
	mu        sync.RWMutex
	records   map[string]domain.TransactionalOutbox
	snapshots []map[string]domain.TransactionalOutbox
}

func NewMemoryTransactionalOutboxRepo() *MemoryTransactionalOutboxRepo {
	return &MemoryTransactionalOutboxRepo{
		records: make(map[string]domain.TransactionalOutbox),
	}
}

func (r *MemoryTransactionalOutboxRepo) TakeSnapshot() {
	r.mu.Lock()
	defer r.mu.Unlock()
	snap := make(map[string]domain.TransactionalOutbox, len(r.records))
	for k, v := range r.records {
		snap[k] = v
	}
	r.snapshots = append(r.snapshots, snap)
}

func (r *MemoryTransactionalOutboxRepo) RollbackSnapshot() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.snapshots) == 0 {
		return
	}
	r.records = r.snapshots[len(r.snapshots)-1]
	r.snapshots = r.snapshots[:len(r.snapshots)-1]
}

func (r *MemoryTransactionalOutboxRepo) CommitSnapshot() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.snapshots) == 0 {
		return
	}
	r.snapshots = r.snapshots[:len(r.snapshots)-1]
}

func (r *MemoryTransactionalOutboxRepo) Create(ctx context.Context, record *domain.TransactionalOutbox) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records[record.ID] = *record
	return nil
}

func (r *MemoryTransactionalOutboxRepo) GetPending(ctx context.Context, limit int) ([]domain.TransactionalOutbox, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.TransactionalOutbox
	for _, rec := range r.records {
		if rec.Status == domain.OutboxStatusPENDING || rec.Status == domain.OutboxStatusFAILED {
			list = append(list, rec)
			if len(list) >= limit {
				break
			}
		}
	}
	return list, nil
}

func (r *MemoryTransactionalOutboxRepo) UpdateStatus(ctx context.Context, id string, status domain.OutboxStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	rec, ok := r.records[id]
	if !ok {
		return errors.New("outbox record not found")
	}
	rec.Status = status
	r.records[id] = rec
	return nil
}

// MemoryTransactionManager implements domain.TransactionManager in-memory
type MemoryTransactionManager struct {
	repos []Snapshotable
}

func NewMemoryTransactionManager(repos ...interface{}) domain.TransactionManager {
	var snapRepos []Snapshotable
	for _, repo := range repos {
		if snap, ok := repo.(Snapshotable); ok {
			snapRepos = append(snapRepos, snap)
		}
	}
	return &MemoryTransactionManager{
		repos: snapRepos,
	}
}

func (m *MemoryTransactionManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	for _, repo := range m.repos {
		repo.TakeSnapshot()
	}

	err := fn(ctx)
	if err != nil {
		for _, repo := range m.repos {
			repo.RollbackSnapshot()
		}
		return err
	}

	for _, repo := range m.repos {
		repo.CommitSnapshot()
	}
	return nil
}

