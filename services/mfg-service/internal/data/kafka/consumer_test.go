package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"erp-system/shared/utils"
	"github.com/erp-system/m-service/internal/business/domain"
	"github.com/erp-system/m-service/internal/business/service"
	"github.com/erp-system/m-service/internal/data/sql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	utils.InitLogger("mfg-consumer-test")
}

type mockPublisher struct {
	failPublish bool
}

func (m *mockPublisher) Publish(ctx context.Context, topic string, key string, payload interface{}) error {
	if m.failPublish {
		return fmt.Errorf("injected publish error")
	}
	return nil
}

type testEnv struct {
	db       *gorm.DB
	consumer *KafkaConsumer
	execSvc  service.WorkOrderExecutionService
}

func setupTestEnv(t *testing.T) *testEnv {
	// Use named in-memory sqlite unique to each test to isolate database instances
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite database: %v", err)
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
		t.Fatalf("failed to run migrations: %v", err)
	}

	wcRepo := sql.NewSQLWorkCenterRepository(db)
	stationRepo := sql.NewSQLRoutingStationRepository(db)
	woRepo := sql.NewSQLWorkOrderRepository(db)
	stateRepo := sql.NewSQLWorkOrderRoutingStateRepository(db)
	outboxRepo := sql.NewSQLTransactionalOutboxRepository(db)
	inboxRepo := sql.NewSQLKafkaEventInboxRepository(db)

	publisher := &mockPublisher{}
	execSvc := service.NewWorkOrderExecutionService(db, woRepo, stateRepo, stationRepo, outboxRepo)
	reliableSvc := service.NewReliableMessagingService(db, inboxRepo)

	// Seed some master floor data
	_ = wcRepo.Create(context.Background(), &domain.WorkCenter{
		ID:             "wc-1",
		LegalEntityID:  "tenant-1",
		WorkCenterCode: "WC-1",
		Name:           "Main WC",
		IsActive:       true,
	})

	_ = stationRepo.Create(context.Background(), &domain.RoutingStation{
		ID:                    "station-1",
		WorkCenterID:          "wc-1",
		RoutingCode:           "ST-1",
		StationType:           "MANUAL",
		StandardSetupTimeMins: 10,
		StandardRunTimeMins:   30,
	})

	consumer := NewKafkaConsumer([]string{"localhost:9092"}, "mfg-group", publisher, reliableSvc, execSvc)

	return &testEnv{
		db:       db,
		consumer: consumer,
		execSvc:  execSvc,
	}
}

func TestConsumer_AllEvents(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	// Seed work orders
	// wo1: for inspection passed test, material is prod-passed
	wo1 := &sql.WorkOrder{
		ID:              "wo-1",
		LegalEntityID:   "tenant-1",
		MaterialID:      "prod-passed",
		BomHeaderID:     "bom-123",
		WorkOrderNumber: "WO-001",
		Status:          string(domain.WorkOrderStateIN_PROGRESS),
		Version:         1,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	_ = env.db.Create(wo1).Error

	// wo2: for inspection failed test, material is prod-failed
	wo2 := &sql.WorkOrder{
		ID:              "wo-2",
		LegalEntityID:   "tenant-1",
		MaterialID:      "prod-failed",
		BomHeaderID:     "bom-123",
		WorkOrderNumber: "WO-002",
		Status:          string(domain.WorkOrderStateIN_PROGRESS),
		Version:         1,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	_ = env.db.Create(wo2).Error

	// wo3: for machine offline test, material is prod-offline
	wo3 := &sql.WorkOrder{
		ID:              "wo-3",
		LegalEntityID:   "tenant-1",
		MaterialID:      "prod-offline",
		BomHeaderID:     "bom-123",
		WorkOrderNumber: "WO-003",
		Status:          string(domain.WorkOrderStateIN_PROGRESS),
		Version:         1,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	_ = env.db.Create(wo3).Error

	// wo4: for obsolete BOM released test, material is prod-bom
	wo4 := &sql.WorkOrder{
		ID:              "wo-4",
		LegalEntityID:   "tenant-1",
		MaterialID:      "prod-bom",
		BomHeaderID:     "bom-123",
		WorkOrderNumber: "WO-004",
		Status:          string(domain.WorkOrderStateIN_PROGRESS),
		Version:         1,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	_ = env.db.Create(wo4).Error

	// 1. TopicPlmBomReleased
	evBom := domain.PlmBomReleasedEvent{
		EventID:       "evt-bom-1",
		LegalEntityID: "tenant-1",
		BomHeaderID:   "bom-header-1",
		MaterialID:    "prod-bom",
		VersionString: "v1.0",
		Timestamp:     time.Now(),
	}
	valBom, _ := json.Marshal(evBom)
	if err := env.consumer.handleMessage(ctx, domain.TopicPlmBomReleased, valBom); err != nil {
		t.Fatalf("failed to process PlmBomReleased: %v", err)
	}

	// Verify wo-4 is now ON_HOLD (frozen)
	var fetchedWO4 sql.WorkOrder
	if err := env.db.First(&fetchedWO4, "id = ?", "wo-4").Error; err == nil {
		if fetchedWO4.Status != string(domain.WorkOrderStateON_HOLD) {
			t.Errorf("expected wo-4 state to be ON_HOLD, got %s", fetchedWO4.Status)
		}
	}

	// 2. TopicQmsInspectionPassed
	evPassed := domain.QmsInspectionPassedEvent{
		EventID:          "evt-passed-1",
		LegalEntityID:    "tenant-1",
		InspectionID:     "insp-1",
		TriggerSource:    "WORK_ORDER",
		SourceDocumentID: "wo-1",
		MaterialID:       "prod-passed",
		Timestamp:        time.Now(),
	}
	valPassed, _ := json.Marshal(evPassed)
	if err := env.consumer.handleMessage(ctx, domain.TopicQmsInspectionPassed, valPassed); err != nil {
		t.Fatalf("failed to process QmsInspectionPassed: %v", err)
	}

	// Verify wo-1 state is now COMPLETED
	var fetchedWO1 sql.WorkOrder
	if err := env.db.First(&fetchedWO1, "id = ?", "wo-1").Error; err == nil {
		if fetchedWO1.Status != string(domain.WorkOrderStateCOMPLETED) {
			t.Errorf("expected wo-1 state to be COMPLETED, got %s", fetchedWO1.Status)
		}
	}

	// 3. TopicQmsInspectionFailed
	evFailed := domain.QmsInspectionFailedEvent{
		EventID:          "evt-failed-1",
		LegalEntityID:    "tenant-1",
		InspectionID:     "insp-2",
		TriggerSource:    "WORK_ORDER",
		SourceDocumentID: "wo-2",
		MaterialID:       "prod-failed",
		NonConformanceID: "nc-1",
		Timestamp:        time.Now(),
	}
	valFailed, _ := json.Marshal(evFailed)
	if err := env.consumer.handleMessage(ctx, domain.TopicQmsInspectionFailed, valFailed); err != nil {
		t.Fatalf("failed to process QmsInspectionFailed: %v", err)
	}

	// Verify wo-2 state is now ON_HOLD
	var fetchedWO2 sql.WorkOrder
	if err := env.db.First(&fetchedWO2, "id = ?", "wo-2").Error; err == nil {
		if fetchedWO2.Status != string(domain.WorkOrderStateON_HOLD) {
			t.Errorf("expected wo-2 state to be ON_HOLD, got %s", fetchedWO2.Status)
		}
	}

	// 4. TopicEamMachineOffline
	evOffline := domain.EamMachineOfflineEvent{
		EventID:       "evt-offline-1",
		LegalEntityID: "tenant-1",
		EquipmentID:   "eq-1",
		WorkOrderID:   "wo-3",
		Priority:      "HIGH",
		Timestamp:     time.Now(),
	}
	valOffline, _ := json.Marshal(evOffline)
	if err := env.consumer.handleMessage(ctx, domain.TopicEamMachineOffline, valOffline); err != nil {
		t.Fatalf("failed to process EamMachineOffline: %v", err)
	}

	// Verify wo-3 state is now ON_HOLD
	var fetchedWO3 sql.WorkOrder
	if err := env.db.First(&fetchedWO3, "id = ?", "wo-3").Error; err == nil {
		if fetchedWO3.Status != string(domain.WorkOrderStateON_HOLD) {
			t.Errorf("expected wo-3 state to be ON_HOLD, got %s", fetchedWO3.Status)
		}
	}

	// Test invalid JSON on TopicPlmBomReleased
	if err := env.consumer.handleMessage(ctx, domain.TopicPlmBomReleased, []byte("invalid-json")); err == nil {
		t.Fatalf("expected error for invalid json on TopicPlmBomReleased, got nil")
	}
	if err := env.consumer.handleMessage(ctx, domain.TopicQmsInspectionPassed, []byte("invalid-json")); err == nil {
		t.Fatalf("expected error for invalid json on TopicQmsInspectionPassed, got nil")
	}
	if err := env.consumer.handleMessage(ctx, domain.TopicQmsInspectionFailed, []byte("invalid-json")); err == nil {
		t.Fatalf("expected error for invalid json on TopicQmsInspectionFailed, got nil")
	}
	if err := env.consumer.handleMessage(ctx, domain.TopicEamMachineOffline, []byte("invalid-json")); err == nil {
		t.Fatalf("expected error for invalid json on TopicEamMachineOffline, got nil")
	}

	// Test unknown topic
	if err := env.consumer.handleMessage(ctx, "unknown-topic", []byte("{}")); err != nil {
		t.Fatalf("expected nil error for unknown topic, got %v", err)
	}

	// Test DLQ Publish happy path
	env.consumer.publishToDLQ(ctx, "test-topic", "test-key", []byte("test-val"), fmt.Errorf("test error"))

	// Test DLQ Publish error path
	failPub := &mockPublisher{failPublish: true}
	failConsumer := NewKafkaConsumer([]string{"localhost:9092"}, "mfg-group", failPub, service.NewReliableMessagingService(env.db, sql.NewSQLKafkaEventInboxRepository(env.db)), env.execSvc)
	failConsumer.publishToDLQ(ctx, "test-topic", "test-key", []byte("test-val"), fmt.Errorf("test error"))
}

func TestConsumer_StartAndClose(t *testing.T) {
	env := setupTestEnv(t)

	// Test Start with canceled context
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	env.consumer.Start(canceledCtx)

	// Test Start with delayed cancel
	ctx, cancel2 := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel2()
	}()
	env.consumer.Start(ctx)

	// Test Close
	_ = env.consumer.Close()
}
