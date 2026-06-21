package sql

import (
	"context"
	"time"

	"github.com/erp-system/fm-service/internal/business/domain"
	"gorm.io/gorm"
)

// SQLChartOfAccountsRepo implements domain.ChartOfAccountsRepository
type SQLChartOfAccountsRepo struct {
	db *gorm.DB
}

func NewSQLChartOfAccountsRepo(db *gorm.DB) *SQLChartOfAccountsRepo {
	return &SQLChartOfAccountsRepo{db: db}
}

func (r *SQLChartOfAccountsRepo) Create(ctx context.Context, coa *domain.ChartOfAccounts) error {
	dbModel := FromDomainChartOfAccounts(coa)
	if err := GetDB(ctx, r.db).Create(dbModel).Error; err != nil {
		return err
	}
	coa.CreatedAt = dbModel.CreatedAt
	coa.UpdatedAt = dbModel.UpdatedAt
	return nil
}

func (r *SQLChartOfAccountsRepo) GetByID(ctx context.Context, id string) (*domain.ChartOfAccounts, error) {
	var dbModel ChartOfAccounts
	if err := GetDB(ctx, r.db).First(&dbModel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return ToDomainChartOfAccounts(&dbModel), nil
}

func (r *SQLChartOfAccountsRepo) GetByCode(ctx context.Context, legalEntityID, accountCode string) (*domain.ChartOfAccounts, error) {
	var dbModel ChartOfAccounts
	if err := GetDB(ctx, r.db).First(&dbModel, "legal_entity_id = ? AND account_code = ?", legalEntityID, accountCode).Error; err != nil {
		return nil, err
	}
	return ToDomainChartOfAccounts(&dbModel), nil
}

func (r *SQLChartOfAccountsRepo) Update(ctx context.Context, coa *domain.ChartOfAccounts) error {
	dbModel := FromDomainChartOfAccounts(coa)
	if err := GetDB(ctx, r.db).Save(dbModel).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLChartOfAccountsRepo) Delete(ctx context.Context, id string) error {
	if err := GetDB(ctx, r.db).Delete(&ChartOfAccounts{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLChartOfAccountsRepo) List(ctx context.Context) ([]domain.ChartOfAccounts, error) {
	var dbModels []ChartOfAccounts
	if err := GetDB(ctx, r.db).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.ChartOfAccounts, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainChartOfAccounts(&m)
	}
	return res, nil
}

// SQLUniversalJournalEntryRepo implements domain.UniversalJournalEntryRepository
type SQLUniversalJournalEntryRepo struct {
	db *gorm.DB
}

func NewSQLUniversalJournalEntryRepo(db *gorm.DB) *SQLUniversalJournalEntryRepo {
	return &SQLUniversalJournalEntryRepo{db: db}
}

func (r *SQLUniversalJournalEntryRepo) Create(ctx context.Context, entry *domain.UniversalJournalEntry, lines []domain.UniversalJournalLine) error {
	tx := GetDB(ctx, r.db)
	return tx.Transaction(func(txDb *gorm.DB) error {
		dbEntry := FromDomainUniversalJournalEntry(entry)
		if err := txDb.Create(dbEntry).Error; err != nil {
			return err
		}
		for i := range lines {
			dbLine := FromDomainUniversalJournalLine(&lines[i])
			dbLine.JournalEntryID = entry.ID
			if err := txDb.Create(dbLine).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *SQLUniversalJournalEntryRepo) GetByID(ctx context.Context, id string) (*domain.UniversalJournalEntry, []domain.UniversalJournalLine, error) {
	tx := GetDB(ctx, r.db)
	var dbEntry UniversalJournalEntry
	if err := tx.First(&dbEntry, "id = ?", id).Error; err != nil {
		return nil, nil, err
	}
	var dbLines []UniversalJournalLine
	if err := tx.Find(&dbLines, "journal_entry_id = ?", id).Error; err != nil {
		return nil, nil, err
	}
	lines := make([]domain.UniversalJournalLine, len(dbLines))
	for i, m := range dbLines {
		lines[i] = *ToDomainUniversalJournalLine(&m)
	}
	return ToDomainUniversalJournalEntry(&dbEntry), lines, nil
}

func (r *SQLUniversalJournalEntryRepo) Update(ctx context.Context, entry *domain.UniversalJournalEntry, lines []domain.UniversalJournalLine) error {
	tx := GetDB(ctx, r.db)
	return tx.Transaction(func(txDb *gorm.DB) error {
		dbEntry := FromDomainUniversalJournalEntry(entry)
		if err := txDb.Save(dbEntry).Error; err != nil {
			return err
		}
		if err := txDb.Delete(&UniversalJournalLine{}, "journal_entry_id = ?", entry.ID).Error; err != nil {
			return err
		}
		for i := range lines {
			dbLine := FromDomainUniversalJournalLine(&lines[i])
			dbLine.JournalEntryID = entry.ID
			if err := txDb.Create(dbLine).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *SQLUniversalJournalEntryRepo) Delete(ctx context.Context, id string) error {
	tx := GetDB(ctx, r.db)
	return tx.Transaction(func(txDb *gorm.DB) error {
		if err := txDb.Delete(&UniversalJournalLine{}, "journal_entry_id = ?", id).Error; err != nil {
			return err
		}
		if err := txDb.Delete(&UniversalJournalEntry{}, "id = ?", id).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *SQLUniversalJournalEntryRepo) List(ctx context.Context) ([]domain.UniversalJournalEntry, error) {
	var dbModels []UniversalJournalEntry
	if err := GetDB(ctx, r.db).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.UniversalJournalEntry, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainUniversalJournalEntry(&m)
	}
	return res, nil
}

// SQLArInvoiceRepo implements domain.ArInvoiceRepository
type SQLArInvoiceRepo struct {
	db *gorm.DB
}

func NewSQLArInvoiceRepo(db *gorm.DB) *SQLArInvoiceRepo {
	return &SQLArInvoiceRepo{db: db}
}

func (r *SQLArInvoiceRepo) Create(ctx context.Context, invoice *domain.ArInvoice) error {
	dbModel := FromDomainArInvoice(invoice)
	if err := GetDB(ctx, r.db).Create(dbModel).Error; err != nil {
		return err
	}
	invoice.CreatedAt = dbModel.CreatedAt
	invoice.UpdatedAt = dbModel.UpdatedAt
	return nil
}

func (r *SQLArInvoiceRepo) GetByID(ctx context.Context, id string) (*domain.ArInvoice, error) {
	var dbModel ArInvoice
	if err := GetDB(ctx, r.db).First(&dbModel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return ToDomainArInvoice(&dbModel), nil
}

func (r *SQLArInvoiceRepo) GetByNumber(ctx context.Context, invoiceNumber string) (*domain.ArInvoice, error) {
	var dbModel ArInvoice
	if err := GetDB(ctx, r.db).First(&dbModel, "invoice_number = ?", invoiceNumber).Error; err != nil {
		return nil, err
	}
	return ToDomainArInvoice(&dbModel), nil
}

func (r *SQLArInvoiceRepo) Update(ctx context.Context, invoice *domain.ArInvoice) error {
	dbModel := FromDomainArInvoice(invoice)
	if err := GetDB(ctx, r.db).Save(dbModel).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLArInvoiceRepo) Delete(ctx context.Context, id string) error {
	if err := GetDB(ctx, r.db).Delete(&ArInvoice{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLArInvoiceRepo) List(ctx context.Context) ([]domain.ArInvoice, error) {
	var dbModels []ArInvoice
	if err := GetDB(ctx, r.db).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.ArInvoice, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainArInvoice(&m)
	}
	return res, nil
}

// SQLPaymentRepo implements domain.PaymentRepository
type SQLPaymentRepo struct {
	db *gorm.DB
}

func NewSQLPaymentRepo(db *gorm.DB) *SQLPaymentRepo {
	return &SQLPaymentRepo{db: db}
}

func (r *SQLPaymentRepo) Create(ctx context.Context, payment *domain.Payment) error {
	dbModel := FromDomainPayment(payment)
	if err := GetDB(ctx, r.db).Create(dbModel).Error; err != nil {
		return err
	}
	payment.CreatedAt = dbModel.CreatedAt
	payment.UpdatedAt = dbModel.UpdatedAt
	return nil
}

func (r *SQLPaymentRepo) GetByID(ctx context.Context, id string) (*domain.Payment, error) {
	var dbModel Payment
	if err := GetDB(ctx, r.db).First(&dbModel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return ToDomainPayment(&dbModel), nil
}

func (r *SQLPaymentRepo) List(ctx context.Context) ([]domain.Payment, error) {
	var dbModels []Payment
	if err := GetDB(ctx, r.db).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.Payment, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainPayment(&m)
	}
	return res, nil
}

// SQLBudgetRepo implements domain.BudgetRepository
type SQLBudgetRepo struct {
	db *gorm.DB
}

func NewSQLBudgetRepo(db *gorm.DB) *SQLBudgetRepo {
	return &SQLBudgetRepo{db: db}
}

func (r *SQLBudgetRepo) Create(ctx context.Context, budget *domain.Budget) error {
	dbModel := FromDomainBudget(budget)
	if err := GetDB(ctx, r.db).Create(dbModel).Error; err != nil {
		return err
	}
	budget.CreatedAt = dbModel.CreatedAt
	budget.UpdatedAt = dbModel.UpdatedAt
	return nil
}

func (r *SQLBudgetRepo) GetByID(ctx context.Context, id string) (*domain.Budget, error) {
	var dbModel Budget
	if err := GetDB(ctx, r.db).First(&dbModel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return ToDomainBudget(&dbModel), nil
}

func (r *SQLBudgetRepo) Update(ctx context.Context, budget *domain.Budget) error {
	dbModel := FromDomainBudget(budget)
	if err := GetDB(ctx, r.db).Save(dbModel).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLBudgetRepo) Delete(ctx context.Context, id string) error {
	if err := GetDB(ctx, r.db).Delete(&Budget{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLBudgetRepo) List(ctx context.Context) ([]domain.Budget, error) {
	var dbModels []Budget
	if err := GetDB(ctx, r.db).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.Budget, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainBudget(&m)
	}
	return res, nil
}

func (r *SQLBudgetRepo) GetByAccountAndPeriod(ctx context.Context, accountID string, fiscalYear int, period int) (*domain.Budget, error) {
	var dbModel Budget
	if err := GetDB(ctx, r.db).First(&dbModel, "account_id = ? AND fiscal_year = ? AND period = ?", accountID, fiscalYear, period).Error; err != nil {
		return nil, err
	}
	return ToDomainBudget(&dbModel), nil
}

// SQLApVendorBillRepo implements domain.ApVendorBillRepository
type SQLApVendorBillRepo struct {
	db *gorm.DB
}

func NewSQLApVendorBillRepo(db *gorm.DB) *SQLApVendorBillRepo {
	return &SQLApVendorBillRepo{db: db}
}

func (r *SQLApVendorBillRepo) Create(ctx context.Context, bill *domain.ApVendorBill) error {
	dbModel := FromDomainApVendorBill(bill)
	if err := GetDB(ctx, r.db).Create(dbModel).Error; err != nil {
		return err
	}
	bill.CreatedAt = dbModel.CreatedAt
	bill.UpdatedAt = dbModel.UpdatedAt
	return nil
}

func (r *SQLApVendorBillRepo) GetByID(ctx context.Context, id string) (*domain.ApVendorBill, error) {
	var dbModel ApVendorBill
	if err := GetDB(ctx, r.db).First(&dbModel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return ToDomainApVendorBill(&dbModel), nil
}

func (r *SQLApVendorBillRepo) GetByNumber(ctx context.Context, billNumber string) (*domain.ApVendorBill, error) {
	var dbModel ApVendorBill
	if err := GetDB(ctx, r.db).First(&dbModel, "bill_number = ?", billNumber).Error; err != nil {
		return nil, err
	}
	return ToDomainApVendorBill(&dbModel), nil
}

func (r *SQLApVendorBillRepo) Update(ctx context.Context, bill *domain.ApVendorBill) error {
	dbModel := FromDomainApVendorBill(bill)
	if err := GetDB(ctx, r.db).Save(dbModel).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLApVendorBillRepo) Delete(ctx context.Context, id string) error {
	if err := GetDB(ctx, r.db).Delete(&ApVendorBill{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLApVendorBillRepo) List(ctx context.Context) ([]domain.ApVendorBill, error) {
	var dbModels []ApVendorBill
	if err := GetDB(ctx, r.db).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.ApVendorBill, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainApVendorBill(&m)
	}
	return res, nil
}

// SQLTaxRateRepo implements domain.TaxRateRepository
type SQLTaxRateRepo struct {
	db *gorm.DB
}

func NewSQLTaxRateRepo(db *gorm.DB) *SQLTaxRateRepo {
	return &SQLTaxRateRepo{db: db}
}

func (r *SQLTaxRateRepo) Create(ctx context.Context, tr *domain.TaxRate) error {
	dbModel := FromDomainTaxRate(tr)
	if err := GetDB(ctx, r.db).Create(dbModel).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLTaxRateRepo) GetByID(ctx context.Context, id string) (*domain.TaxRate, error) {
	var dbModel TaxRate
	if err := GetDB(ctx, r.db).First(&dbModel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return ToDomainTaxRate(&dbModel), nil
}

func (r *SQLTaxRateRepo) List(ctx context.Context) ([]domain.TaxRate, error) {
	var dbModels []TaxRate
	if err := GetDB(ctx, r.db).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.TaxRate, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainTaxRate(&m)
	}
	return res, nil
}

// SQLCurrencyRateRepo implements domain.CurrencyRateRepository
type SQLCurrencyRateRepo struct {
	db *gorm.DB
}

func NewSQLCurrencyRateRepo(db *gorm.DB) *SQLCurrencyRateRepo {
	return &SQLCurrencyRateRepo{db: db}
}

func (r *SQLCurrencyRateRepo) Create(ctx context.Context, rate *domain.CurrencyRate) error {
	dbModel := FromDomainCurrencyRate(rate)
	if err := GetDB(ctx, r.db).Create(dbModel).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLCurrencyRateRepo) GetByID(ctx context.Context, id string) (*domain.CurrencyRate, error) {
	var dbModel CurrencyRate
	if err := GetDB(ctx, r.db).First(&dbModel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return ToDomainCurrencyRate(&dbModel), nil
}

func (r *SQLCurrencyRateRepo) List(ctx context.Context) ([]domain.CurrencyRate, error) {
	var dbModels []CurrencyRate
	if err := GetDB(ctx, r.db).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.CurrencyRate, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainCurrencyRate(&m)
	}
	return res, nil
}

// SQLFiscalYearRepo implements domain.FiscalYearRepository
type SQLFiscalYearRepo struct {
	db *gorm.DB
}

func NewSQLFiscalYearRepo(db *gorm.DB) *SQLFiscalYearRepo {
	return &SQLFiscalYearRepo{db: db}
}

func (r *SQLFiscalYearRepo) Create(ctx context.Context, fy *domain.FiscalYear) error {
	dbModel := FromDomainFiscalYear(fy)
	if err := GetDB(ctx, r.db).Create(dbModel).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLFiscalYearRepo) GetByID(ctx context.Context, id string) (*domain.FiscalYear, error) {
	var dbModel FiscalYear
	if err := GetDB(ctx, r.db).First(&dbModel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return ToDomainFiscalYear(&dbModel), nil
}

func (r *SQLFiscalYearRepo) Update(ctx context.Context, fy *domain.FiscalYear) error {
	dbModel := FromDomainFiscalYear(fy)
	if err := GetDB(ctx, r.db).Save(dbModel).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLFiscalYearRepo) List(ctx context.Context) ([]domain.FiscalYear, error) {
	var dbModels []FiscalYear
	if err := GetDB(ctx, r.db).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.FiscalYear, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainFiscalYear(&m)
	}
	return res, nil
}

// SQLCostCenterRepo implements domain.CostCenterRepository
type SQLCostCenterRepo struct {
	db *gorm.DB
}

func NewSQLCostCenterRepo(db *gorm.DB) *SQLCostCenterRepo {
	return &SQLCostCenterRepo{db: db}
}

func (r *SQLCostCenterRepo) Create(ctx context.Context, cc *domain.CostCenter) error {
	dbModel := FromDomainCostCenter(cc)
	if err := GetDB(ctx, r.db).Create(dbModel).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLCostCenterRepo) GetByID(ctx context.Context, id string) (*domain.CostCenter, error) {
	var dbModel CostCenter
	if err := GetDB(ctx, r.db).First(&dbModel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return ToDomainCostCenter(&dbModel), nil
}

func (r *SQLCostCenterRepo) Update(ctx context.Context, cc *domain.CostCenter) error {
	dbModel := FromDomainCostCenter(cc)
	if err := GetDB(ctx, r.db).Save(dbModel).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLCostCenterRepo) Delete(ctx context.Context, id string) error {
	if err := GetDB(ctx, r.db).Delete(&CostCenter{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLCostCenterRepo) List(ctx context.Context) ([]domain.CostCenter, error) {
	var dbModels []CostCenter
	if err := GetDB(ctx, r.db).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.CostCenter, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainCostCenter(&m)
	}
	return res, nil
}

// SQLBankAccountRepo implements domain.BankAccountRepository
type SQLBankAccountRepo struct {
	db *gorm.DB
}

func NewSQLBankAccountRepo(db *gorm.DB) *SQLBankAccountRepo {
	return &SQLBankAccountRepo{db: db}
}

func (r *SQLBankAccountRepo) Create(ctx context.Context, ba *domain.BankAccount) error {
	dbModel := FromDomainBankAccount(ba)
	if err := GetDB(ctx, r.db).Create(dbModel).Error; err != nil {
		return err
	}
	ba.CreatedAt = dbModel.CreatedAt
	ba.UpdatedAt = dbModel.UpdatedAt
	return nil
}

func (r *SQLBankAccountRepo) GetByID(ctx context.Context, id string) (*domain.BankAccount, error) {
	var dbModel BankAccount
	if err := GetDB(ctx, r.db).First(&dbModel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return ToDomainBankAccount(&dbModel), nil
}

func (r *SQLBankAccountRepo) Update(ctx context.Context, ba *domain.BankAccount) error {
	tx := GetDB(ctx, r.db)

	var dbModel BankAccount
	if err := tx.First(&dbModel, "id = ?", ba.ID).Error; err != nil {
		return err
	}

	expectedVersion := dbModel.Version
	newVersion := expectedVersion + 1

	res := tx.Model(&BankAccount{}).
		Where("id = ? AND version = ?", ba.ID, expectedVersion).
		Updates(map[string]interface{}{
			"liquid_balance": ba.LiquidBalance,
			"updated_at":     time.Now(),
			"version":        newVersion,
		})

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return domain.ErrOptimisticLock
	}

	ba.Version = newVersion
	return nil
}

func (r *SQLBankAccountRepo) List(ctx context.Context) ([]domain.BankAccount, error) {
	var dbModels []BankAccount
	if err := GetDB(ctx, r.db).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.BankAccount, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainBankAccount(&m)
	}
	return res, nil
}

// SQLCustomerCreditRepo implements domain.CustomerCreditRepository
type SQLCustomerCreditRepo struct {
	db *gorm.DB
}

func NewSQLCustomerCreditRepo(db *gorm.DB) *SQLCustomerCreditRepo {
	return &SQLCustomerCreditRepo{db: db}
}

func (r *SQLCustomerCreditRepo) Create(ctx context.Context, cc *domain.CustomerCredit) error {
	dbModel := FromDomainCustomerCredit(cc)
	if err := GetDB(ctx, r.db).Create(dbModel).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLCustomerCreditRepo) GetByID(ctx context.Context, id string) (*domain.CustomerCredit, error) {
	var dbModel CustomerCredit
	if err := GetDB(ctx, r.db).First(&dbModel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return ToDomainCustomerCredit(&dbModel), nil
}

func (r *SQLCustomerCreditRepo) GetByCustomerID(ctx context.Context, customerID string) (*domain.CustomerCredit, error) {
	var dbModel CustomerCredit
	if err := GetDB(ctx, r.db).First(&dbModel, "customer_id = ?", customerID).Error; err != nil {
		return nil, err
	}
	return ToDomainCustomerCredit(&dbModel), nil
}

func (r *SQLCustomerCreditRepo) Update(ctx context.Context, cc *domain.CustomerCredit) error {
	tx := GetDB(ctx, r.db)

	var dbModel CustomerCredit
	if err := tx.First(&dbModel, "id = ?", cc.ID).Error; err != nil {
		return err
	}

	expectedVersion := dbModel.Version
	newVersion := expectedVersion + 1

	res := tx.Model(&CustomerCredit{}).
		Where("id = ? AND version = ?", cc.ID, expectedVersion).
		Updates(map[string]interface{}{
			"credit_limit":    cc.CreditLimit,
			"current_balance": cc.CurrentBalance,
			"is_on_hold":      cc.IsOnHold,
			"updated_at":      time.Now(),
			"version":         newVersion,
		})

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return domain.ErrOptimisticLock
	}

	cc.Version = newVersion
	return nil
}

func (r *SQLCustomerCreditRepo) List(ctx context.Context) ([]domain.CustomerCredit, error) {
	var dbModels []CustomerCredit
	if err := GetDB(ctx, r.db).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.CustomerCredit, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainCustomerCredit(&m)
	}
	return res, nil
}

// SQLBankStatementRepo implements domain.BankStatementRepository
type SQLBankStatementRepo struct {
	db *gorm.DB
}

func NewSQLBankStatementRepo(db *gorm.DB) *SQLBankStatementRepo {
	return &SQLBankStatementRepo{db: db}
}

func (r *SQLBankStatementRepo) Create(ctx context.Context, bs *domain.BankStatement, lines []domain.BankStatementLine) error {
	tx := GetDB(ctx, r.db)
	return tx.Transaction(func(txDb *gorm.DB) error {
		dbBs := FromDomainBankStatement(bs)
		if err := txDb.Create(dbBs).Error; err != nil {
			return err
		}
		for i := range lines {
			dbLine := FromDomainBankStatementLine(&lines[i])
			dbLine.StatementID = bs.ID
			if err := txDb.Create(dbLine).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *SQLBankStatementRepo) GetByID(ctx context.Context, id string) (*domain.BankStatement, []domain.BankStatementLine, error) {
	tx := GetDB(ctx, r.db)
	var dbBs BankStatement
	if err := tx.First(&dbBs, "id = ?", id).Error; err != nil {
		return nil, nil, err
	}
	var dbLines []BankStatementLine
	if err := tx.Find(&dbLines, "statement_id = ?", id).Error; err != nil {
		return nil, nil, err
	}
	lines := make([]domain.BankStatementLine, len(dbLines))
	for i, m := range dbLines {
		lines[i] = *ToDomainBankStatementLine(&m)
	}
	return ToDomainBankStatement(&dbBs), lines, nil
}

func (r *SQLBankStatementRepo) Update(ctx context.Context, bs *domain.BankStatement, lines []domain.BankStatementLine) error {
	tx := GetDB(ctx, r.db)
	return tx.Transaction(func(txDb *gorm.DB) error {
		dbBs := FromDomainBankStatement(bs)
		if err := txDb.Save(dbBs).Error; err != nil {
			return err
		}
		if err := txDb.Delete(&BankStatementLine{}, "statement_id = ?", bs.ID).Error; err != nil {
			return err
		}
		for i := range lines {
			dbLine := FromDomainBankStatementLine(&lines[i])
			dbLine.StatementID = bs.ID
			if err := txDb.Create(dbLine).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *SQLBankStatementRepo) List(ctx context.Context) ([]domain.BankStatement, error) {
	var dbModels []BankStatement
	if err := GetDB(ctx, r.db).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.BankStatement, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainBankStatement(&m)
	}
	return res, nil
}

// SQLTransactionalOutboxRepo implements domain.TransactionalOutboxRepository
type SQLTransactionalOutboxRepo struct {
	db *gorm.DB
}

func NewSQLTransactionalOutboxRepo(db *gorm.DB) *SQLTransactionalOutboxRepo {
	return &SQLTransactionalOutboxRepo{db: db}
}

func (r *SQLTransactionalOutboxRepo) Create(ctx context.Context, record *domain.TransactionalOutbox) error {
	dbModel := FromDomainTransactionalOutbox(record)
	if err := GetDB(ctx, r.db).Create(dbModel).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLTransactionalOutboxRepo) GetPending(ctx context.Context, limit int) ([]domain.TransactionalOutbox, error) {
	var dbModels []TransactionalOutbox
	if err := GetDB(ctx, r.db).Where("status IN ?", []domain.OutboxStatus{domain.OutboxStatusPENDING, domain.OutboxStatusFAILED}).Limit(limit).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.TransactionalOutbox, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainTransactionalOutbox(&m)
	}
	return res, nil
}

func (r *SQLTransactionalOutboxRepo) UpdateStatus(ctx context.Context, id string, status domain.OutboxStatus) error {
	if err := GetDB(ctx, r.db).Model(&TransactionalOutbox{}).Where("id = ?", id).Update("status", status).Error; err != nil {
		return err
	}
	return nil
}

// SQLLegalEntityRepo implements domain.LegalEntityRepository
type SQLLegalEntityRepo struct {
	db *gorm.DB
}

func NewSQLLegalEntityRepo(db *gorm.DB) *SQLLegalEntityRepo {
	return &SQLLegalEntityRepo{db: db}
}

func (r *SQLLegalEntityRepo) Create(ctx context.Context, le *domain.LegalEntity) error {
	dbModel := FromDomainLegalEntity(le)
	if err := GetDB(ctx, r.db).Create(dbModel).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLLegalEntityRepo) GetByID(ctx context.Context, id string) (*domain.LegalEntity, error) {
	var dbModel LegalEntity
	if err := GetDB(ctx, r.db).First(&dbModel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return ToDomainLegalEntity(&dbModel), nil
}

func (r *SQLLegalEntityRepo) GetByCode(ctx context.Context, code string) (*domain.LegalEntity, error) {
	var dbModel LegalEntity
	if err := GetDB(ctx, r.db).First(&dbModel, "company_code = ?", code).Error; err != nil {
		return nil, err
	}
	return ToDomainLegalEntity(&dbModel), nil
}

func (r *SQLLegalEntityRepo) List(ctx context.Context) ([]domain.LegalEntity, error) {
	var dbModels []LegalEntity
	if err := GetDB(ctx, r.db).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.LegalEntity, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainLegalEntity(&m)
	}
	return res, nil
}

// SQLCapitalAssetRepo implements domain.CapitalAssetRepository
type SQLCapitalAssetRepo struct {
	db *gorm.DB
}

func NewSQLCapitalAssetRepo(db *gorm.DB) *SQLCapitalAssetRepo {
	return &SQLCapitalAssetRepo{db: db}
}

func (r *SQLCapitalAssetRepo) Create(ctx context.Context, asset *domain.CapitalAsset) error {
	dbModel := FromDomainCapitalAsset(asset)
	if err := GetDB(ctx, r.db).Create(dbModel).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLCapitalAssetRepo) GetByID(ctx context.Context, id string) (*domain.CapitalAsset, error) {
	var dbModel CapitalAsset
	if err := GetDB(ctx, r.db).First(&dbModel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return ToDomainCapitalAsset(&dbModel), nil
}

func (r *SQLCapitalAssetRepo) Update(ctx context.Context, asset *domain.CapitalAsset) error {
	dbModel := FromDomainCapitalAsset(asset)
	if err := GetDB(ctx, r.db).Save(dbModel).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLCapitalAssetRepo) List(ctx context.Context) ([]domain.CapitalAsset, error) {
	var dbModels []CapitalAsset
	if err := GetDB(ctx, r.db).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.CapitalAsset, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainCapitalAsset(&m)
	}
	return res, nil
}

// SQLDepreciationScheduleLineRepo implements domain.DepreciationScheduleLineRepository
type SQLDepreciationScheduleLineRepo struct {
	db *gorm.DB
}

func NewSQLDepreciationScheduleLineRepo(db *gorm.DB) *SQLDepreciationScheduleLineRepo {
	return &SQLDepreciationScheduleLineRepo{db: db}
}

func (r *SQLDepreciationScheduleLineRepo) CreateMany(ctx context.Context, lines []domain.DepreciationScheduleLine) error {
	dbModels := make([]*DepreciationScheduleLine, len(lines))
	for i, l := range lines {
		dbModels[i] = FromDomainDepreciationScheduleLine(&l)
	}
	if err := GetDB(ctx, r.db).Create(&dbModels).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLDepreciationScheduleLineRepo) GetByAssetID(ctx context.Context, assetID string) ([]domain.DepreciationScheduleLine, error) {
	var dbModels []DepreciationScheduleLine
	if err := GetDB(ctx, r.db).Where("fixed_asset_id = ?", assetID).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.DepreciationScheduleLine, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainDepreciationScheduleLine(&m)
	}
	return res, nil
}

func (r *SQLDepreciationScheduleLineRepo) GetUnpostedByPeriod(ctx context.Context, fiscalYear, periodNumber int) ([]domain.DepreciationScheduleLine, error) {
	var dbModels []DepreciationScheduleLine
	if err := GetDB(ctx, r.db).Where("is_posted = ? AND fiscal_year = ? AND period_number = ?", false, fiscalYear, periodNumber).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.DepreciationScheduleLine, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainDepreciationScheduleLine(&m)
	}
	return res, nil
}

func (r *SQLDepreciationScheduleLineRepo) Update(ctx context.Context, line *domain.DepreciationScheduleLine) error {
	dbModel := FromDomainDepreciationScheduleLine(line)
	if err := GetDB(ctx, r.db).Save(dbModel).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLDepreciationScheduleLineRepo) List(ctx context.Context) ([]domain.DepreciationScheduleLine, error) {
	var dbModels []DepreciationScheduleLine
	if err := GetDB(ctx, r.db).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.DepreciationScheduleLine, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainDepreciationScheduleLine(&m)
	}
	return res, nil
}

// SQLKafkaEventInboxRepo implements domain.KafkaEventInboxRepository
type SQLKafkaEventInboxRepo struct {
	db *gorm.DB
}

func NewSQLKafkaEventInboxRepo(db *gorm.DB) *SQLKafkaEventInboxRepo {
	return &SQLKafkaEventInboxRepo{db: db}
}

func (r *SQLKafkaEventInboxRepo) Create(ctx context.Context, record *domain.KafkaEventInbox) error {
	dbModel := FromDomainKafkaEventInbox(record)
	if err := GetDB(ctx, r.db).Create(dbModel).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLKafkaEventInboxRepo) GetByID(ctx context.Context, eventID string) (*domain.KafkaEventInbox, error) {
	var dbModel KafkaEventInbox
	if err := GetDB(ctx, r.db).First(&dbModel, "event_id = ?", eventID).Error; err != nil {
		return nil, err
	}
	return ToDomainKafkaEventInbox(&dbModel), nil
}
