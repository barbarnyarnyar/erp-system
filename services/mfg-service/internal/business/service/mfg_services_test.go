package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/erp-system/m-service/internal/business/domain"
	"github.com/erp-system/m-service/internal/business/service"
	"github.com/erp-system/m-service/internal/data/memory"
	"github.com/erp-system/m-service/internal/data/sql"
	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ==========================================
// Test Helpers & Mocks
// ==========================================

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

type mockWCRepo struct {
	domain.WorkCenterRepository
	getByCodeFunc func(ctx context.Context, legalEntityID, code string) (*domain.WorkCenter, error)
	getByIDFunc   func(ctx context.Context, id string) (*domain.WorkCenter, error)
	createFunc    func(ctx context.Context, wc *domain.WorkCenter) error
}

func (m *mockWCRepo) GetByCode(ctx context.Context, legalEntityID, code string) (*domain.WorkCenter, error) {
	if m.getByCodeFunc != nil {
		return m.getByCodeFunc(ctx, legalEntityID, code)
	}
	return nil, errors.New("not found")
}

func (m *mockWCRepo) GetByID(ctx context.Context, id string) (*domain.WorkCenter, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *mockWCRepo) Create(ctx context.Context, wc *domain.WorkCenter) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, wc)
	}
	return nil
}

type mockStationRepo struct {
	domain.RoutingStationRepository
	getByIDFunc   func(ctx context.Context, id string) (*domain.RoutingStation, error)
	getByCodeFunc func(ctx context.Context, workCenterID, code string) (*domain.RoutingStation, error)
	createFunc    func(ctx context.Context, station *domain.RoutingStation) error
}

func (m *mockStationRepo) GetByID(ctx context.Context, id string) (*domain.RoutingStation, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *mockStationRepo) GetByCode(ctx context.Context, workCenterID, code string) (*domain.RoutingStation, error) {
	if m.getByCodeFunc != nil {
		return m.getByCodeFunc(ctx, workCenterID, code)
	}
	return nil, errors.New("not found")
}

func (m *mockStationRepo) Create(ctx context.Context, station *domain.RoutingStation) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, station)
	}
	return nil
}

type mockWORepo struct {
	domain.WorkOrderRepository
	createFunc  func(ctx context.Context, wo *domain.WorkOrder) error
	getByIDFunc func(ctx context.Context, id string) (*domain.WorkOrder, error)
	updateFunc  func(ctx context.Context, wo *domain.WorkOrder) error
	listFunc    func(ctx context.Context) ([]domain.WorkOrder, error)
}

func (m *mockWORepo) Create(ctx context.Context, wo *domain.WorkOrder) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, wo)
	}
	return nil
}

func (m *mockWORepo) GetByID(ctx context.Context, id string) (*domain.WorkOrder, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *mockWORepo) Update(ctx context.Context, wo *domain.WorkOrder) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, wo)
	}
	return nil
}

func (m *mockWORepo) List(ctx context.Context) ([]domain.WorkOrder, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx)
	}
	return nil, nil
}

type mockStateRepo struct {
	domain.WorkOrderRoutingStateRepository
	getActiveByWorkOrderIDFunc func(ctx context.Context, workOrderID string) (*domain.WorkOrderRoutingState, error)
	updateFunc                 func(ctx context.Context, state *domain.WorkOrderRoutingState) error
	createFunc                 func(ctx context.Context, state *domain.WorkOrderRoutingState) error
}

func (m *mockStateRepo) GetActiveByWorkOrderID(ctx context.Context, workOrderID string) (*domain.WorkOrderRoutingState, error) {
	if m.getActiveByWorkOrderIDFunc != nil {
		return m.getActiveByWorkOrderIDFunc(ctx, workOrderID)
	}
	return nil, errors.New("not found")
}

func (m *mockStateRepo) Update(ctx context.Context, state *domain.WorkOrderRoutingState) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, state)
	}
	return nil
}

func (m *mockStateRepo) Create(ctx context.Context, state *domain.WorkOrderRoutingState) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, state)
	}
	return nil
}

type mockOutboxRepo struct {
	domain.TransactionalOutboxRepository
	createFunc    func(ctx context.Context, msg *domain.TransactionalOutbox) error
	getUnsentFunc func(ctx context.Context, limit int) ([]domain.TransactionalOutbox, error)
	getByIDFunc   func(ctx context.Context, id string) (*domain.TransactionalOutbox, error)
	updateFunc    func(ctx context.Context, msg *domain.TransactionalOutbox) error
}

func (m *mockOutboxRepo) Create(ctx context.Context, msg *domain.TransactionalOutbox) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, msg)
	}
	return nil
}

func (m *mockOutboxRepo) GetUnsent(ctx context.Context, limit int) ([]domain.TransactionalOutbox, error) {
	if m.getUnsentFunc != nil {
		return m.getUnsentFunc(ctx, limit)
	}
	return nil, nil
}

func (m *mockOutboxRepo) GetByID(ctx context.Context, id string) (*domain.TransactionalOutbox, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *mockOutboxRepo) Update(ctx context.Context, msg *domain.TransactionalOutbox) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, msg)
	}
	return nil
}

type mockConsumeRepo struct {
	domain.MaterialConsumptionLogRepository
	createFunc func(ctx context.Context, log *domain.MaterialConsumptionLog) error
}

func (m *mockConsumeRepo) Create(ctx context.Context, log *domain.MaterialConsumptionLog) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, log)
	}
	return nil
}

type mockYieldRepo struct {
	domain.ProductionYieldLogRepository
	createFunc func(ctx context.Context, log *domain.ProductionYieldLog) error
}

func (m *mockYieldRepo) Create(ctx context.Context, log *domain.ProductionYieldLog) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, log)
	}
	return nil
}

type mockInboxRepo struct {
	domain.KafkaEventInboxRepository
	getByIDFunc func(ctx context.Context, eventID string) (*domain.KafkaEventInbox, error)
	createFunc  func(ctx context.Context, msg *domain.KafkaEventInbox) error
}

func (m *mockInboxRepo) GetByID(ctx context.Context, eventID string) (*domain.KafkaEventInbox, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, eventID)
	}
	return nil, errors.New("not found")
}

func (m *mockInboxRepo) Create(ctx context.Context, msg *domain.KafkaEventInbox) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, msg)
	}
	return nil
}

// ==========================================
// Tests
// ==========================================

func TestGetDB(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	ctx := context.Background()
	// Case 1: no transaction in context
	if res := service.GetDB(ctx, db); res == nil {
		t.Error("expected non-nil db")
	}

	// Case 2: transaction in context
	tx := db.Begin()
	defer tx.Rollback()
	ctxTx := context.WithValue(ctx, "gorm_tx", tx)
	if res := service.GetDB(ctxTx, db); res == nil {
		t.Error("expected non-nil db from transaction context")
	}
}

func TestFloorConfigurationService_EstablishWorkCenter(t *testing.T) {
	ctx := context.Background()

	// Case 1: WorkCenter exists
	wcRepo := &mockWCRepo{
		getByCodeFunc: func(ctx context.Context, legalEntityID, code string) (*domain.WorkCenter, error) {
			return &domain.WorkCenter{ID: "wc-1", WorkCenterCode: code}, nil
		},
	}
	svc := service.NewFloorConfigurationService(wcRepo, &mockStationRepo{})
	wc, err := svc.EstablishWorkCenter(ctx, "legal-1", "WC-1", "Name 1")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if wc.ID != "wc-1" {
		t.Errorf("expected wc-1, got %s", wc.ID)
	}

	// Case 2: WorkCenter does not exist, created successfully
	wcRepo = &mockWCRepo{
		getByCodeFunc: func(ctx context.Context, legalEntityID, code string) (*domain.WorkCenter, error) {
			return nil, errors.New("not found")
		},
		createFunc: func(ctx context.Context, wc *domain.WorkCenter) error {
			return nil
		},
	}
	svc = service.NewFloorConfigurationService(wcRepo, &mockStationRepo{})
	wc, err = svc.EstablishWorkCenter(ctx, "legal-1", "WC-1", "Name 1")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if wc.WorkCenterCode != "WC-1" {
		t.Errorf("expected WC-1, got %s", wc.WorkCenterCode)
	}

	// Case 3: Create fails
	wcRepo = &mockWCRepo{
		getByCodeFunc: func(ctx context.Context, legalEntityID, code string) (*domain.WorkCenter, error) {
			return nil, errors.New("not found")
		},
		createFunc: func(ctx context.Context, wc *domain.WorkCenter) error {
			return errors.New("create error")
		},
	}
	svc = service.NewFloorConfigurationService(wcRepo, &mockStationRepo{})
	_, err = svc.EstablishWorkCenter(ctx, "legal-1", "WC-1", "Name 1")
	if err == nil || err.Error() != "create error" {
		t.Errorf("expected create error, got %v", err)
	}
}

func TestFloorConfigurationService_AppendStationToCenter(t *testing.T) {
	ctx := context.Background()
	equipmentID := "equip-1"

	// Case 1: WorkCenter not found
	wcRepo := &mockWCRepo{
		getByIDFunc: func(ctx context.Context, id string) (*domain.WorkCenter, error) {
			return nil, errors.New("wc not found")
		},
	}
	svc := service.NewFloorConfigurationService(wcRepo, &mockStationRepo{})
	_, err := svc.AppendStationToCenter(ctx, "wc-1", "ST-1", domain.StationTypeASSEMBLY, &equipmentID, 10, 20)
	if err == nil {
		t.Error("expected error, got nil")
	}

	// Case 2: RoutingStation already exists
	wcRepo = &mockWCRepo{
		getByIDFunc: func(ctx context.Context, id string) (*domain.WorkCenter, error) {
			return &domain.WorkCenter{ID: id}, nil
		},
	}
	stationRepo := &mockStationRepo{
		getByCodeFunc: func(ctx context.Context, workCenterID, code string) (*domain.RoutingStation, error) {
			return &domain.RoutingStation{ID: "st-1"}, nil
		},
	}
	svc = service.NewFloorConfigurationService(wcRepo, stationRepo)
	_, err = svc.AppendStationToCenter(ctx, "wc-1", "ST-1", domain.StationTypeASSEMBLY, &equipmentID, 10, 20)
	if err == nil {
		t.Error("expected error, got nil")
	}

	// Case 3: Create succeeds
	stationRepo.getByCodeFunc = func(ctx context.Context, workCenterID, code string) (*domain.RoutingStation, error) {
		return nil, errors.New("not found")
	}
	stationRepo.createFunc = func(ctx context.Context, station *domain.RoutingStation) error {
		return nil
	}
	wc, err := svc.AppendStationToCenter(ctx, "wc-1", "ST-1", domain.StationTypeASSEMBLY, &equipmentID, 10, 20)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if wc.RoutingCode != "ST-1" {
		t.Errorf("expected ST-1, got %s", wc.RoutingCode)
	}

	// Case 4: Create fails
	stationRepo.createFunc = func(ctx context.Context, station *domain.RoutingStation) error {
		return errors.New("create failed")
	}
	_, err = svc.AppendStationToCenter(ctx, "wc-1", "ST-1", domain.StationTypeASSEMBLY, &equipmentID, 10, 20)
	if err == nil || err.Error() != "create failed" {
		t.Errorf("expected create failed error, got %v", err)
	}
}

func TestWorkOrderExecutionService_InstantiateWorkOrder(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	ctx := context.Background()

	// Happy path
	woRepo := &mockWORepo{
		createFunc: func(ctx context.Context, wo *domain.WorkOrder) error {
			return nil
		},
	}
	svc := service.NewWorkOrderExecutionService(db, woRepo, &mockStateRepo{}, &mockStationRepo{}, &mockOutboxRepo{})
	wo, err := svc.InstantiateWorkOrder(ctx, "tenant-1", "mat-1", "bom-1", decimal.NewFromInt(100), time.Now(), time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if wo.MaterialID != "mat-1" {
		t.Errorf("expected mat-1, got %s", wo.MaterialID)
	}

	// Create fails
	woRepo.createFunc = func(ctx context.Context, wo *domain.WorkOrder) error {
		return errors.New("create error")
	}
	_, err = svc.InstantiateWorkOrder(ctx, "tenant-1", "mat-1", "bom-1", decimal.NewFromInt(100), time.Now(), time.Now().Add(time.Hour))
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestWorkOrderExecutionService_TransitionWorkOrderState(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	ctx := context.Background()

	// Case 1: WorkOrder not found
	woRepo := &mockWORepo{
		getByIDFunc: func(ctx context.Context, id string) (*domain.WorkOrder, error) {
			return nil, errors.New("not found")
		},
	}
	svc := service.NewWorkOrderExecutionService(db, woRepo, &mockStateRepo{}, &mockStationRepo{}, &mockOutboxRepo{})
	_, err = svc.TransitionWorkOrderState(ctx, "wo-1", domain.WorkOrderStateSTAGED, domain.WorkOrderStateRELEASED)
	if err == nil {
		t.Error("expected error, got nil")
	}

	// Case 2: Current state mismatch
	woRepo.getByIDFunc = func(ctx context.Context, id string) (*domain.WorkOrder, error) {
		return &domain.WorkOrder{ID: id, Status: domain.WorkOrderStateRELEASED}, nil
	}
	_, err = svc.TransitionWorkOrderState(ctx, "wo-1", domain.WorkOrderStateSTAGED, domain.WorkOrderStateRELEASED)
	if err == nil {
		t.Error("expected error, got nil")
	}

	// Case 3: Update fails
	woRepo.getByIDFunc = func(ctx context.Context, id string) (*domain.WorkOrder, error) {
		return &domain.WorkOrder{ID: id, Status: domain.WorkOrderStateSTAGED}, nil
	}
	woRepo.updateFunc = func(ctx context.Context, wo *domain.WorkOrder) error {
		return errors.New("update error")
	}
	_, err = svc.TransitionWorkOrderState(ctx, "wo-1", domain.WorkOrderStateSTAGED, domain.WorkOrderStateRELEASED)
	if err == nil {
		t.Error("expected error, got nil")
	}

	// Case 4: Transition to RELEASED (no event)
	woRepo.updateFunc = func(ctx context.Context, wo *domain.WorkOrder) error {
		return nil
	}
	wo, err := svc.TransitionWorkOrderState(ctx, "wo-1", domain.WorkOrderStateSTAGED, domain.WorkOrderStateRELEASED)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if wo.Status != domain.WorkOrderStateRELEASED {
		t.Errorf("expected RELEASED, got %s", wo.Status)
	}

	// Case 5: Transition to IN_PROGRESS (emits event)
	woRepo.getByIDFunc = func(ctx context.Context, id string) (*domain.WorkOrder, error) {
		return &domain.WorkOrder{ID: id, Status: domain.WorkOrderStateRELEASED}, nil
	}
	outboxRepo := &mockOutboxRepo{
		createFunc: func(ctx context.Context, msg *domain.TransactionalOutbox) error {
			return nil
		},
	}
	svc = service.NewWorkOrderExecutionService(db, woRepo, &mockStateRepo{}, &mockStationRepo{}, outboxRepo)
	wo, err = svc.TransitionWorkOrderState(ctx, "wo-1", domain.WorkOrderStateRELEASED, domain.WorkOrderStateIN_PROGRESS)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if wo.Status != domain.WorkOrderStateIN_PROGRESS {
		t.Errorf("expected IN_PROGRESS, got %s", wo.Status)
	}

	// Case 6: Transition to COMPLETED (emits event)
	woRepo.getByIDFunc = func(ctx context.Context, id string) (*domain.WorkOrder, error) {
		return &domain.WorkOrder{ID: id, Status: domain.WorkOrderStateIN_PROGRESS}, nil
	}
	wo, err = svc.TransitionWorkOrderState(ctx, "wo-1", domain.WorkOrderStateIN_PROGRESS, domain.WorkOrderStateCOMPLETED)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if wo.Status != domain.WorkOrderStateCOMPLETED {
		t.Errorf("expected COMPLETED, got %s", wo.Status)
	}

	// Case 7: Emit event fails
	outboxRepo.createFunc = func(ctx context.Context, msg *domain.TransactionalOutbox) error {
		return errors.New("outbox create error")
	}
	woRepo.getByIDFunc = func(ctx context.Context, id string) (*domain.WorkOrder, error) {
		return &domain.WorkOrder{ID: id, Status: domain.WorkOrderStateIN_PROGRESS}, nil
	}
	_, err = svc.TransitionWorkOrderState(ctx, "wo-1", domain.WorkOrderStateIN_PROGRESS, domain.WorkOrderStateCOMPLETED)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestWorkOrderExecutionService_RerouteWorkOrderStation(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	ctx := context.Background()

	// Case 1: WorkOrder not found
	woRepo := &mockWORepo{
		getByIDFunc: func(ctx context.Context, id string) (*domain.WorkOrder, error) {
			return nil, errors.New("not found")
		},
	}
	svc := service.NewWorkOrderExecutionService(db, woRepo, &mockStateRepo{}, &mockStationRepo{}, &mockOutboxRepo{})
	err = svc.RerouteWorkOrderStation(ctx, "wo-1", "st-old", "st-new", false)
	if err == nil {
		t.Error("expected error, got nil")
	}

	// Case 2: Target station not found
	woRepo.getByIDFunc = func(ctx context.Context, id string) (*domain.WorkOrder, error) {
		return &domain.WorkOrder{ID: id}, nil
	}
	stationRepo := &mockStationRepo{
		getByIDFunc: func(ctx context.Context, id string) (*domain.RoutingStation, error) {
			return nil, errors.New("not found")
		},
	}
	svc = service.NewWorkOrderExecutionService(db, woRepo, &mockStateRepo{}, stationRepo, &mockOutboxRepo{})
	err = svc.RerouteWorkOrderStation(ctx, "wo-1", "st-old", "st-new", false)
	if err == nil {
		t.Error("expected error, got nil")
	}

	// Case 3: Active state exists, update active fails
	stationRepo.getByIDFunc = func(ctx context.Context, id string) (*domain.RoutingStation, error) {
		return &domain.RoutingStation{ID: id}, nil
	}
	stateRepo := &mockStateRepo{
		getActiveByWorkOrderIDFunc: func(ctx context.Context, workOrderID string) (*domain.WorkOrderRoutingState, error) {
			return &domain.WorkOrderRoutingState{ID: "wors-1"}, nil
		},
		updateFunc: func(ctx context.Context, state *domain.WorkOrderRoutingState) error {
			return errors.New("update state error")
		},
	}
	svc = service.NewWorkOrderExecutionService(db, woRepo, stateRepo, stationRepo, &mockOutboxRepo{})
	err = svc.RerouteWorkOrderStation(ctx, "wo-1", "st-old", "st-new", false)
	if err == nil {
		t.Error("expected error, got nil")
	}

	// Case 4: Active state exists, update active succeeds, create new state fails
	stateRepo.updateFunc = func(ctx context.Context, state *domain.WorkOrderRoutingState) error {
		return nil
	}
	stateRepo.createFunc = func(ctx context.Context, state *domain.WorkOrderRoutingState) error {
		return errors.New("create state error")
	}
	svc = service.NewWorkOrderExecutionService(db, woRepo, stateRepo, stationRepo, &mockOutboxRepo{})
	err = svc.RerouteWorkOrderStation(ctx, "wo-1", "st-old", "st-new", false)
	if err == nil {
		t.Error("expected error, got nil")
	}

	// Case 5: Active state does not exist (returns error), create new state succeeds
	stateRepo.getActiveByWorkOrderIDFunc = func(ctx context.Context, workOrderID string) (*domain.WorkOrderRoutingState, error) {
		return nil, errors.New("not found")
	}
	stateRepo.createFunc = func(ctx context.Context, state *domain.WorkOrderRoutingState) error {
		return nil
	}
	svc = service.NewWorkOrderExecutionService(db, woRepo, stateRepo, stationRepo, &mockOutboxRepo{})
	err = svc.RerouteWorkOrderStation(ctx, "wo-1", "st-old", "st-new", true)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	// Case 6: Active state exists, update active succeeds, create new state succeeds
	stateRepo.getActiveByWorkOrderIDFunc = func(ctx context.Context, workOrderID string) (*domain.WorkOrderRoutingState, error) {
		return &domain.WorkOrderRoutingState{ID: "wors-1"}, nil
	}
	err = svc.RerouteWorkOrderStation(ctx, "wo-1", "st-old", "st-new", true)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestWorkOrderExecutionService_FreezeObsoleteWorkOrders_Errors(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	ctx := context.Background()

	// Case 1: List fails
	woRepo := &mockWORepo{
		listFunc: func(ctx context.Context) ([]domain.WorkOrder, error) {
			return nil, errors.New("list error")
		},
	}
	svc := service.NewWorkOrderExecutionService(db, woRepo, &mockStateRepo{}, &mockStationRepo{}, &mockOutboxRepo{})
	err = svc.FreezeObsoleteWorkOrders(ctx, "mat-1", "bom-new")
	if err == nil {
		t.Error("expected error, got nil")
	}

	// Case 2: Update fails
	woRepo.listFunc = func(ctx context.Context) ([]domain.WorkOrder, error) {
		return []domain.WorkOrder{
			{ID: "wo-1", MaterialID: "mat-1", BomHeaderID: "bom-old", Status: domain.WorkOrderStateSTAGED},
		}, nil
	}
	woRepo.updateFunc = func(ctx context.Context, wo *domain.WorkOrder) error {
		return errors.New("update error")
	}
	err = svc.FreezeObsoleteWorkOrders(ctx, "mat-1", "bom-new")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestShopFloorTelemetryService_RecordBulkMaterialConsumption(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	ctx := context.Background()

	woRepo := &mockWORepo{}
	stationRepo := &mockStationRepo{}
	consumeRepo := &mockConsumeRepo{}
	yieldRepo := &mockYieldRepo{}
	outboxRepo := &mockOutboxRepo{}

	svc := service.NewShopFloorTelemetryService(db, woRepo, stationRepo, consumeRepo, yieldRepo, outboxRepo)

	lines := []domain.ConsumptionSubmissionInput{
		{MaterialID: "mat-1", RoutingStationID: "st-1", QuantityConsumed: decimal.NewFromInt(5), WarehouseID: "wh-1"},
	}

	// Case 1: WorkOrder not found
	woRepo.getByIDFunc = func(ctx context.Context, id string) (*domain.WorkOrder, error) {
		return nil, errors.New("not found")
	}
	err = svc.RecordBulkMaterialConsumption(ctx, "tenant-1", "wo-1", lines)
	if err == nil {
		t.Error("expected error, got nil")
	}

	// Case 2: Station not found
	woRepo.getByIDFunc = func(ctx context.Context, id string) (*domain.WorkOrder, error) {
		return &domain.WorkOrder{ID: id}, nil
	}
	stationRepo.getByIDFunc = func(ctx context.Context, id string) (*domain.RoutingStation, error) {
		return nil, errors.New("not found")
	}
	err = svc.RecordBulkMaterialConsumption(ctx, "tenant-1", "wo-1", lines)
	if err == nil {
		t.Error("expected error, got nil")
	}

	// Case 3: Consume repo create fails
	stationRepo.getByIDFunc = func(ctx context.Context, id string) (*domain.RoutingStation, error) {
		return &domain.RoutingStation{ID: id}, nil
	}
	consumeRepo.createFunc = func(ctx context.Context, log *domain.MaterialConsumptionLog) error {
		return errors.New("create error")
	}
	err = svc.RecordBulkMaterialConsumption(ctx, "tenant-1", "wo-1", lines)
	if err == nil {
		t.Error("expected error, got nil")
	}

	// Case 4: Event emit fails
	consumeRepo.createFunc = func(ctx context.Context, log *domain.MaterialConsumptionLog) error {
		return nil
	}
	outboxRepo.createFunc = func(ctx context.Context, msg *domain.TransactionalOutbox) error {
		return errors.New("create error")
	}
	err = svc.RecordBulkMaterialConsumption(ctx, "tenant-1", "wo-1", lines)
	if err == nil {
		t.Error("expected error, got nil")
	}

	// Case 5: Happy path
	outboxRepo.createFunc = func(ctx context.Context, msg *domain.TransactionalOutbox) error {
		return nil
	}
	err = svc.RecordBulkMaterialConsumption(ctx, "tenant-1", "wo-1", lines)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
}

func TestShopFloorTelemetryService_CommitProductionYield(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	ctx := context.Background()

	woRepo := &mockWORepo{}
	stationRepo := &mockStationRepo{}
	consumeRepo := &mockConsumeRepo{}
	yieldRepo := &mockYieldRepo{}
	outboxRepo := &mockOutboxRepo{}

	svc := service.NewShopFloorTelemetryService(db, woRepo, stationRepo, consumeRepo, yieldRepo, outboxRepo)

	// Case 1: WorkOrder not found
	woRepo.getByIDFunc = func(ctx context.Context, id string) (*domain.WorkOrder, error) {
		return nil, errors.New("not found")
	}
	err = svc.CommitProductionYield(ctx, "tenant-1", "wo-1", "st-1", decimal.NewFromInt(10), decimal.NewFromInt(1), "op-1")
	if err == nil {
		t.Error("expected error, got nil")
	}

	// Case 2: Station not found
	woRepo.getByIDFunc = func(ctx context.Context, id string) (*domain.WorkOrder, error) {
		return &domain.WorkOrder{ID: id, QuantityProduced: decimal.Zero}, nil
	}
	stationRepo.getByIDFunc = func(ctx context.Context, id string) (*domain.RoutingStation, error) {
		return nil, errors.New("not found")
	}
	err = svc.CommitProductionYield(ctx, "tenant-1", "wo-1", "st-1", decimal.NewFromInt(10), decimal.NewFromInt(1), "op-1")
	if err == nil {
		t.Error("expected error, got nil")
	}

	// Case 3: Create yield log fails
	stationRepo.getByIDFunc = func(ctx context.Context, id string) (*domain.RoutingStation, error) {
		return &domain.RoutingStation{ID: id}, nil
	}
	yieldRepo.createFunc = func(ctx context.Context, log *domain.ProductionYieldLog) error {
		return errors.New("create error")
	}
	err = svc.CommitProductionYield(ctx, "tenant-1", "wo-1", "st-1", decimal.NewFromInt(10), decimal.NewFromInt(1), "op-1")
	if err == nil {
		t.Error("expected error, got nil")
	}

	// Case 4: Update WorkOrder fails
	yieldRepo.createFunc = func(ctx context.Context, log *domain.ProductionYieldLog) error {
		return nil
	}
	woRepo.updateFunc = func(ctx context.Context, wo *domain.WorkOrder) error {
		return errors.New("update error")
	}
	err = svc.CommitProductionYield(ctx, "tenant-1", "wo-1", "st-1", decimal.NewFromInt(10), decimal.NewFromInt(1), "op-1")
	if err == nil {
		t.Error("expected error, got nil")
	}

	// Case 5: Outbox emit fails
	woRepo.updateFunc = func(ctx context.Context, wo *domain.WorkOrder) error {
		return nil
	}
	outboxRepo.createFunc = func(ctx context.Context, msg *domain.TransactionalOutbox) error {
		return errors.New("create outbox error")
	}
	err = svc.CommitProductionYield(ctx, "tenant-1", "wo-1", "st-1", decimal.NewFromInt(10), decimal.NewFromInt(1), "op-1")
	if err == nil {
		t.Error("expected error, got nil")
	}

	// Case 6: Happy path
	outboxRepo.createFunc = func(ctx context.Context, msg *domain.TransactionalOutbox) error {
		return nil
	}
	err = svc.CommitProductionYield(ctx, "tenant-1", "wo-1", "st-1", decimal.NewFromInt(10), decimal.NewFromInt(1), "op-1")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
}

func TestOutboxRelayWorker(t *testing.T) {
	ctx := context.Background()

	// GetUnsentMessages happy path
	outboxRepo := &mockOutboxRepo{
		getUnsentFunc: func(ctx context.Context, limit int) ([]domain.TransactionalOutbox, error) {
			return []domain.TransactionalOutbox{{ID: "msg-1"}}, nil
		},
	}
	svc := service.NewOutboxRelayWorker(outboxRepo)
	msgs, err := svc.GetUnsentMessages(ctx, 10)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if len(msgs) != 1 || msgs[0].ID != "msg-1" {
		t.Errorf("unexpected message array: %v", msgs)
	}

	// LogProcessingAttempt GetByID fails
	outboxRepo.getByIDFunc = func(ctx context.Context, id string) (*domain.TransactionalOutbox, error) {
		return nil, errors.New("not found")
	}
	err = svc.LogProcessingAttempt(ctx, "msg-1", 0, "error notes")
	if err == nil {
		t.Error("expected error, got nil")
	}

	// LogProcessingAttempt happy path (retries < 5)
	outboxRepo.getByIDFunc = func(ctx context.Context, id string) (*domain.TransactionalOutbox, error) {
		return &domain.TransactionalOutbox{ID: id, RetryCount: 0, Status: domain.OutboxStatusPENDING}, nil
	}
	outboxRepo.updateFunc = func(ctx context.Context, msg *domain.TransactionalOutbox) error {
		if msg.RetryCount != 1 {
			return fmt.Errorf("expected RetryCount 1, got %d", msg.RetryCount)
		}
		if msg.Status != domain.OutboxStatusPENDING {
			return fmt.Errorf("expected Status PENDING, got %s", msg.Status)
		}
		return nil
	}
	err = svc.LogProcessingAttempt(ctx, "msg-1", 0, "error notes")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	// LogProcessingAttempt becomes FAILED (retries >= 4, which increments to >= 5)
	outboxRepo.getByIDFunc = func(ctx context.Context, id string) (*domain.TransactionalOutbox, error) {
		return &domain.TransactionalOutbox{ID: id, RetryCount: 4, Status: domain.OutboxStatusPENDING}, nil
	}
	outboxRepo.updateFunc = func(ctx context.Context, msg *domain.TransactionalOutbox) error {
		if msg.RetryCount != 5 {
			return fmt.Errorf("expected RetryCount 5, got %d", msg.RetryCount)
		}
		if msg.Status != domain.OutboxStatusFAILED {
			return fmt.Errorf("expected Status FAILED, got %s", msg.Status)
		}
		return nil
	}
	err = svc.LogProcessingAttempt(ctx, "msg-1", 4, "error notes")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	// UpdateOutboxStatus GetByID fails
	outboxRepo.getByIDFunc = func(ctx context.Context, id string) (*domain.TransactionalOutbox, error) {
		return nil, errors.New("not found")
	}
	err = svc.UpdateOutboxStatus(ctx, "msg-1", domain.OutboxStatusSENT)
	if err == nil {
		t.Error("expected error, got nil")
	}

	// UpdateOutboxStatus happy path
	outboxRepo.getByIDFunc = func(ctx context.Context, id string) (*domain.TransactionalOutbox, error) {
		return &domain.TransactionalOutbox{ID: id, Status: domain.OutboxStatusPENDING}, nil
	}
	outboxRepo.updateFunc = func(ctx context.Context, msg *domain.TransactionalOutbox) error {
		if msg.Status != domain.OutboxStatusSENT {
			return fmt.Errorf("expected Status SENT, got %s", msg.Status)
		}
		return nil
	}
	err = svc.UpdateOutboxStatus(ctx, "msg-1", domain.OutboxStatusSENT)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
}

func TestReliableMessagingService(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	ctx := context.Background()

	inboxRepo := &mockInboxRepo{}
	svc := service.NewReliableMessagingService(db, inboxRepo)

	// IsEventProcessed Case 1: getByID returns error (returns false, nil in the implementation)
	inboxRepo.getByIDFunc = func(ctx context.Context, id string) (*domain.KafkaEventInbox, error) {
		return nil, errors.New("db error")
	}
	res, err := svc.IsEventProcessed(ctx, "evt-1")
	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
	if res {
		t.Error("expected false")
	}

	// IsEventProcessed Case 2: getByID not found (returns message not found error or similar, wait, if GetByID returns errors.New("not found") or nil message, does it return false, nil?)
	// Let's check IsEventProcessed implementation in mfg_services.go:
	// func (s *ReliableMessagingServiceImpl) IsEventProcessed(ctx context.Context, eventID string) (bool, error) {
	//     msg, err := s.inboxRepo.GetByID(ctx, eventID)
	//     if err == nil && msg != nil {
	//         return msg.ProcessingStatus == domain.EventProcessingStatusSUCCESS, nil
	//     }
	//     return false, nil
	// }
	// Ah! It only returns true if err == nil and msg != nil. If there is an error, it returns false, nil EXCEPT when err is something else? Wait, no! If err is not nil, the if condition fails, and it returns false, nil! Wait, no, earlier:
	// "inboxRepo.getByIDFunc = func(...) { return nil, errors.New("db error") }" -> IsEventProcessed returns false, nil! Wait, because err is not nil, the if condition fails, and it executes return false, nil.
	// Let's double check IsEventProcessed:
	// if err == nil && msg != nil { ... } return false, nil
	// This means IsEventProcessed NEVER returns a non-nil error, unless we change the repository GetByID? No, the code literally says:
	// func (s *ReliableMessagingServiceImpl) IsEventProcessed(ctx context.Context, eventID string) (bool, error) {
	// 	msg, err := s.inboxRepo.GetByID(ctx, eventID)
	// 	if err == nil && msg != nil {
	// 		return msg.ProcessingStatus == domain.EventProcessingStatusSUCCESS, nil
	// 	}
	// 	return false, nil
	// }
	// Indeed! The error returned from IsEventProcessed is always nil! That is interesting. So the test case of err returning from IsEventProcessed will actually return false, nil. Let's adapt our test assertions.

	inboxRepo.getByIDFunc = func(ctx context.Context, id string) (*domain.KafkaEventInbox, error) {
		return nil, errors.New("not found")
	}
	res, err = svc.IsEventProcessed(ctx, "evt-1")
	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
	if res {
		t.Error("expected false")
	}

	// IsEventProcessed Case 3: SUCCESS status
	inboxRepo.getByIDFunc = func(ctx context.Context, id string) (*domain.KafkaEventInbox, error) {
		return &domain.KafkaEventInbox{EventID: id, ProcessingStatus: domain.EventProcessingStatusSUCCESS}, nil
	}
	res, err = svc.IsEventProcessed(ctx, "evt-1")
	if err != nil || !res {
		t.Errorf("expected true, nil, got %v, %v", res, err)
	}

	// IsEventProcessed Case 4: FAILED status
	inboxRepo.getByIDFunc = func(ctx context.Context, id string) (*domain.KafkaEventInbox, error) {
		return &domain.KafkaEventInbox{EventID: id, ProcessingStatus: domain.EventProcessingStatusFAILED}, nil
	}
	res, err = svc.IsEventProcessed(ctx, "evt-1")
	if err != nil || res {
		t.Errorf("expected false, nil, got %v, %v", res, err)
	}

	// ExecuteIdempotentTransaction Case 1: Already processed
	inboxRepo.getByIDFunc = func(ctx context.Context, id string) (*domain.KafkaEventInbox, error) {
		return &domain.KafkaEventInbox{EventID: id, ProcessingStatus: domain.EventProcessingStatusSUCCESS}, nil
	}
	routineCalled := false
	err = svc.ExecuteIdempotentTransaction(ctx, "evt-1", "type-1", nil, func(ctx context.Context) error {
		routineCalled = true
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if routineCalled {
		t.Error("expected routine not to be called")
	}

	// ExecuteIdempotentTransaction Case 2: Not processed, businessRoutine fails
	inboxRepo.getByIDFunc = func(ctx context.Context, id string) (*domain.KafkaEventInbox, error) {
		return nil, errors.New("not found")
	}
	inboxRepo.createFunc = func(ctx context.Context, msg *domain.KafkaEventInbox) error {
		if msg.ProcessingStatus != domain.EventProcessingStatusFAILED {
			return fmt.Errorf("expected status FAILED, got %s", msg.ProcessingStatus)
		}
		return nil
	}
	err = svc.ExecuteIdempotentTransaction(ctx, "evt-1", "type-1", nil, func(ctx context.Context) error {
		return errors.New("routine error")
	})
	if err == nil || err.Error() != "routine error" {
		t.Errorf("expected routine error, got: %v", err)
	}

	// ExecuteIdempotentTransaction Case 3: Not processed, businessRoutine succeeds
	inboxRepo.createFunc = func(ctx context.Context, msg *domain.KafkaEventInbox) error {
		if msg.ProcessingStatus != domain.EventProcessingStatusSUCCESS {
			return fmt.Errorf("expected status SUCCESS, got %s", msg.ProcessingStatus)
		}
		return nil
	}
	err = svc.ExecuteIdempotentTransaction(ctx, "evt-1", "type-1", nil, func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
}

func TestWorkOrderExecutionService_FreezeObsoleteWorkOrders_Success(t *testing.T) {
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

func TestFloorConfigurationService_EstablishWorkCenter_MemoryRepo(t *testing.T) {
	ctx := context.Background()
	wcRepo := memory.NewMemoryWorkCenterRepo()
	stationRepo := memory.NewMemoryRoutingStationRepo()
	svc := service.NewFloorConfigurationService(wcRepo, stationRepo)

	// Test EstablishWorkCenter with MemoryWorkCenterRepo
	wc, err := svc.EstablishWorkCenter(ctx, "tenant-1", "WC-MEM-01", "Memory WC")
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if wc.WorkCenterCode != "WC-MEM-01" {
		t.Errorf("expected WC-MEM-01, got %s", wc.WorkCenterCode)
	}

	// Repeated call returns same
	wc2, err := svc.EstablishWorkCenter(ctx, "tenant-1", "WC-MEM-01", "Memory WC")
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if wc2.ID != wc.ID {
		t.Errorf("expected same ID")
	}
}

