package sql

import (
	"context"
	"fmt"

	"github.com/erp-system/fm-service/internal/business/domain"
	"github.com/erp-system/fm-service/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type contextKey string

const txKey contextKey = "gorm_tx"

// GetDB retrieves the database connection or active transaction from context
func GetDB(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return defaultDB.WithContext(ctx)
}

// GORMTransactionManager implements domain.TransactionManager using GORM
type GORMTransactionManager struct {
	db *gorm.DB
}

func NewGORMTransactionManager(db *gorm.DB) domain.TransactionManager {
	return &GORMTransactionManager{db: db}
}

func (tm *GORMTransactionManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := ctx.Value(txKey).(*gorm.DB); ok {
		return fn(ctx)
	}

	return tm.db.Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)
		return fn(txCtx)
	})
}

// InitDB initializes PostgreSQL connection, runs AutoMigrate, and returns DB instance
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Database,
		cfg.Database.Port,
		cfg.Database.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Run AutoMigrate for all models in correct order of referential dependency
	err = db.AutoMigrate(
		&LegalEntity{},
		&CostCenter{},
		&TaxRate{},
		&CurrencyRate{},
		&FiscalYear{},
		&BankAccount{},
		&CustomerCredit{},
		&Payment{},
		&Budget{},
		&BankStatement{},
		&BankStatementLine{},
		&ChartOfAccounts{},
		&UniversalJournalEntry{},
		&UniversalJournalLine{},
		&CapitalAsset{},
		&DepreciationScheduleLine{},
		&KafkaEventInbox{},
		&TransactionalOutbox{},
		&ApVendorBill{},
		&ArInvoice{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to auto migrate database: %w", err)
	}

	return db, nil
}
