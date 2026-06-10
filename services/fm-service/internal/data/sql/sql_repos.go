package sql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/erp-system/fm-service/internal/business/domain"
)

type SQLCurrencyRateRepo struct {
	db *sql.DB
}

func NewSQLCurrencyRateRepo(db *sql.DB) *SQLCurrencyRateRepo {
	return &SQLCurrencyRateRepo{db: db}
}

func (r *SQLCurrencyRateRepo) Create(ctx context.Context, rate *domain.CurrencyRate) error {
	// Not fully implemented for mock
	return nil
}

func (r *SQLCurrencyRateRepo) GetByID(ctx context.Context, id string) (*domain.CurrencyRate, error) {
	return nil, errors.New("not implemented")
}

func (r *SQLCurrencyRateRepo) List(ctx context.Context) ([]domain.CurrencyRate, error) {
	return nil, nil
}

type SQLFiscalYearRepo struct {
	db *sql.DB
}

func NewSQLFiscalYearRepo(db *sql.DB) *SQLFiscalYearRepo {
	return &SQLFiscalYearRepo{db: db}
}

func (r *SQLFiscalYearRepo) Create(ctx context.Context, fy *domain.FiscalYear) error {
	return nil
}

func (r *SQLFiscalYearRepo) GetByID(ctx context.Context, id string) (*domain.FiscalYear, error) {
	return nil, errors.New("not implemented")
}

func (r *SQLFiscalYearRepo) Update(ctx context.Context, fy *domain.FiscalYear) error {
	return nil
}

func (r *SQLFiscalYearRepo) List(ctx context.Context) ([]domain.FiscalYear, error) {
	return nil, nil
}

type SQLCostCenterRepo struct {
	db *sql.DB
}

func NewSQLCostCenterRepo(db *sql.DB) *SQLCostCenterRepo {
	return &SQLCostCenterRepo{db: db}
}

func (r *SQLCostCenterRepo) Create(ctx context.Context, cc *domain.CostCenter) error {
	return nil
}

func (r *SQLCostCenterRepo) GetByID(ctx context.Context, id string) (*domain.CostCenter, error) {
	return nil, errors.New("not implemented")
}

func (r *SQLCostCenterRepo) Update(ctx context.Context, cc *domain.CostCenter) error {
	return nil
}

func (r *SQLCostCenterRepo) Delete(ctx context.Context, id string) error {
	return nil
}

func (r *SQLCostCenterRepo) List(ctx context.Context) ([]domain.CostCenter, error) {
	return nil, nil
}

type SQLBankAccountRepo struct {
	db *sql.DB
}

func NewSQLBankAccountRepo(db *sql.DB) *SQLBankAccountRepo {
	return &SQLBankAccountRepo{db: db}
}

func (r *SQLBankAccountRepo) Create(ctx context.Context, ba *domain.BankAccount) error {
	return nil
}

func (r *SQLBankAccountRepo) GetByID(ctx context.Context, id string) (*domain.BankAccount, error) {
	return nil, errors.New("not implemented")
}

func (r *SQLBankAccountRepo) Update(ctx context.Context, ba *domain.BankAccount) error {
	return nil
}

func (r *SQLBankAccountRepo) List(ctx context.Context) ([]domain.BankAccount, error) {
	return nil, nil
}

type SQLCustomerCreditRepo struct {
	db *sql.DB
}

func NewSQLCustomerCreditRepo(db *sql.DB) *SQLCustomerCreditRepo {
	return &SQLCustomerCreditRepo{db: db}
}

func (r *SQLCustomerCreditRepo) Create(ctx context.Context, cc *domain.CustomerCredit) error {
	return nil
}

func (r *SQLCustomerCreditRepo) GetByID(ctx context.Context, id string) (*domain.CustomerCredit, error) {
	return nil, errors.New("not implemented")
}

func (r *SQLCustomerCreditRepo) Update(ctx context.Context, cc *domain.CustomerCredit) error {
	return nil
}

func (r *SQLCustomerCreditRepo) List(ctx context.Context) ([]domain.CustomerCredit, error) {
	return nil, nil
}

type SQLBankStatementRepo struct {
	db *sql.DB
}

func NewSQLBankStatementRepo(db *sql.DB) *SQLBankStatementRepo {
	return &SQLBankStatementRepo{db: db}
}

func (r *SQLBankStatementRepo) Create(ctx context.Context, bs *domain.BankStatement, lines []domain.BankStatementLine) error {
	return nil
}

func (r *SQLBankStatementRepo) GetByID(ctx context.Context, id string) (*domain.BankStatement, []domain.BankStatementLine, error) {
	return nil, nil, errors.New("not implemented")
}

func (r *SQLBankStatementRepo) Update(ctx context.Context, bs *domain.BankStatement, lines []domain.BankStatementLine) error {
	return nil
}

func (r *SQLBankStatementRepo) List(ctx context.Context) ([]domain.BankStatement, error) {
	return nil, nil
}

type SQLTransactionRepo struct {
	db *sql.DB
}

func NewSQLTransactionRepo(db *sql.DB) *SQLTransactionRepo {
	return &SQLTransactionRepo{db: db}
}

func (r *SQLTransactionRepo) Create(ctx context.Context, tx *domain.Transaction, lines []domain.TransactionLine) error {
	return nil
}

func (r *SQLTransactionRepo) GetByID(ctx context.Context, id string) (*domain.Transaction, []domain.TransactionLine, error) {
	return nil, nil, errors.New("not implemented")
}

func (r *SQLTransactionRepo) Update(ctx context.Context, tx *domain.Transaction, lines []domain.TransactionLine) error {
	return nil
}

func (r *SQLTransactionRepo) List(ctx context.Context) ([]domain.Transaction, error) {
	return nil, nil
}
