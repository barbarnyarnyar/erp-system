package sql

import (
	"context"
	"fmt"

	"github.com/erp-system/qms-service/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const TxKey = "gorm_tx"

func GetDB(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(TxKey).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return defaultDB.WithContext(ctx)
}

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

	// AutoMigrate QMS entities
	err = db.AutoMigrate(
		&InspectionPlan{},
		&InspectionMetricDefinition{},
		&QualityInspection{},
		&InspectionResultLine{},
		&NonConformanceLog{},
		&TransactionalOutbox{},
		&KafkaEventInbox{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to auto migrate database: %w", err)
	}

	return db, nil
}
