package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/erp-system/m-service/internal/business/domain"
	"github.com/erp-system/m-service/internal/business/service"
	"github.com/erp-system/m-service/internal/data/sql"
	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupMfgTestDB(t *testing.T) (*gorm.DB, domain.WorkOrderRepository, domain.WorkOrderRoutingStateRepository, domain.RoutingStationRepository, domain.TransactionalOutboxRepository) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open sqlite database: %v", err)
	}

	err = db.AutoMigrate(
		&sql.WorkCenter{},
		&sql.RoutingStation{},
		&sql.WorkOrder{},
		&sql.WorkOrderRoutingState{},
		&sql.MaterialConsumptionLog{},
		&sql.ProductionYieldLog{},
		&sql.TransactionalOutbox{},
		&sql.KafkaEventInbox{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	woRepo := sql.NewSQLWorkOrderRepository(db)
	stateRepo := sql.NewSQLWorkOrderRoutingStateRepository(db)
	stationRepo := sql.NewSQLRoutingStationRepository(db)
	outboxRepo := sql.NewSQLTransactionalOutboxRepository(db)

	return db, woRepo, stateRepo, stationRepo, outboxRepo
}

func TestWorkOrderExecutionService_FreezeObsoleteWorkOrders(t *testing.T) {
	db, woRepo, stateRepo, stationRepo, outboxRepo := setupMfgTestDB(t)

	execSvc := service.NewWorkOrderExecutionService(db, woRepo, stateRepo, stationRepo, outboxRepo)

	materialID := "material-steel-123"
	oldBom := "bom-rev-1"
	newBom := "bom-rev-2"

	wo1 := &domain.WorkOrder{
		ID:               "wo-1",
		LegalEntityID:    "tenant-1",
		MaterialID:       materialID,
		BomHeaderID:      oldBom,
		WorkOrderNumber:  "WO-001",
		QuantityTarget:   decimal.NewFromFloat(100),
		QuantityProduced: decimal.Zero,
		Status:           domain.WorkOrderStateSTAGED,
		Version:          1,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	wo2 := &domain.WorkOrder{
		ID:               "wo-2",
		LegalEntityID:    "tenant-1",
		MaterialID:       materialID,
		BomHeaderID:      oldBom,
		WorkOrderNumber:  "WO-002",
		QuantityTarget:   decimal.NewFromFloat(100),
		QuantityProduced: decimal.Zero,
		Status:           domain.WorkOrderStateIN_PROGRESS,
		Version:          1,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	wo3 := &domain.WorkOrder{
		ID:               "wo-3",
		LegalEntityID:    "tenant-1",
		MaterialID:       materialID,
		BomHeaderID:      newBom,
		WorkOrderNumber:  "WO-003",
		QuantityTarget:   decimal.NewFromFloat(100),
		QuantityProduced: decimal.Zero,
		Status:           domain.WorkOrderStateSTAGED,
		Version:          1,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	_ = woRepo.Create(context.Background(), wo1)
	_ = woRepo.Create(context.Background(), wo2)
	_ = woRepo.Create(context.Background(), wo3)

	err := execSvc.FreezeObsoleteWorkOrders(context.Background(), materialID, newBom)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	res1, _ := woRepo.GetByID(context.Background(), "wo-1")
	if res1.Status != domain.WorkOrderStateON_HOLD {
		t.Errorf("Expected wo-1 to be ON_HOLD, got %s", res1.Status)
	}

	res2, _ := woRepo.GetByID(context.Background(), "wo-2")
	if res2.Status != domain.WorkOrderStateON_HOLD {
		t.Errorf("Expected wo-2 to be ON_HOLD, got %s", res2.Status)
	}

	res3, _ := woRepo.GetByID(context.Background(), "wo-3")
	if res3.Status != domain.WorkOrderStateSTAGED {
		t.Errorf("Expected wo-3 to remain STAGED, got %s", res3.Status)
	}
}
