package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"erp-system/shared/utils"
	"github.com/erp-system/scm-service/internal/business/domain"
	"github.com/erp-system/scm-service/internal/business/service"
	"github.com/erp-system/scm-service/internal/data/sql"
	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	utils.InitLogger("scm-consumer-test")
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
	db        *gorm.DB
	consumer  *KafkaConsumer
	poSvc     *service.PurchaseOrderService
	invSvc    *service.InventoryService
	demandSvc *service.DemandPlanningService
}

func setupTestEnv(t *testing.T) *testEnv {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite database: %v", err)
	}

	err = db.AutoMigrate(
		&sql.ProductCategory{},
		&sql.Product{},
		&sql.Location{},
		&sql.Supplier{},
		&sql.VendorContract{},
		&sql.StockBalance{},
		&sql.InventoryMovement{},
		&sql.StockTransfer{},
		&sql.PurchaseRequisition{},
		&sql.PurchaseRequisitionLine{},
		&sql.PurchaseOrder{},
		&sql.PurchaseOrderLine{},
		&sql.Receipt{},
		&sql.ReceiptLine{},
		&sql.Shipment{},
		&sql.ShipmentLine{},
		&sql.DemandForecast{},
		&sql.KafkaEventInbox{},
		&sql.TransactionalOutbox{},
	)
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// SQL Repositories
	prodRepo := sql.NewSQLProductRepo(db)
	locRepo := sql.NewSQLLocationRepo(db)
	invRepo := sql.NewSQLStockBalanceRepo(db)
	moveRepo := sql.NewSQLInventoryMovementRepo(db)
	poRepo := sql.NewSQLPurchaseOrderRepo(db)
	lineRepo := sql.NewSQLPurchaseOrderLineRepo(db)
	reqRepo := sql.NewSQLPurchaseRequisitionRepo(db)
	reqLineRepo := sql.NewSQLPurchaseRequisitionLineRepo(db)
	forecastRepo := sql.NewSQLDemandForecastRepo(db)
	transferRepo := sql.NewSQLStockTransferRepo(db)
	inboxRepo := sql.NewSQLKafkaEventInboxRepo(db)

	publisher := &mockPublisher{}
	tm := sql.NewGORMTransactionManager(db)

	// Seed default warehouse location
	_ = locRepo.Create(context.Background(), &domain.Location{
		ID:           "loc_default",
		LocationCode: "WH-MAIN",
		LocationName: "Main Distribution Center",
		LocationType: "WAREHOUSE",
		IsActive:     true,
	})

	// Seed product
	_ = prodRepo.Create(context.Background(), &domain.Product{
		ID:          "prod-123",
		ProductCode: "PROD-ABC",
		ProductName: "Test Product",
		IsActive:    true,
	})

	poSvc := service.NewPurchaseOrderService(poRepo, lineRepo, reqRepo, reqLineRepo, publisher, tm)
	invSvc := service.NewInventoryService(invRepo, moveRepo, transferRepo, publisher, tm)
	demandSvc := service.NewDemandPlanningService(forecastRepo)

	consumer := NewKafkaConsumer([]string{"localhost:9092"}, "scm-group", publisher, poSvc, invSvc, demandSvc, inboxRepo)

	return &testEnv{
		db:        db,
		consumer:  consumer,
		poSvc:     poSvc,
		invSvc:    invSvc,
		demandSvc: demandSvc,
	}
}

func TestConsumer_AllEvents(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	// Helper to marshal event with an event_id
	marshalEvent := func(eventID string, ev interface{}) []byte {
		b, _ := json.Marshal(ev)
		var m map[string]interface{}
		_ = json.Unmarshal(b, &m)
		m["event_id"] = eventID
		res, _ := json.Marshal(m)
		return res
	}

	// 1. TopicCrmSalesOrderCreated
	evSalesOrder := domain.SalesOrderCreatedEvent{
		SalesOrderID: "so-123",
		OrderNumber:  "SO-001",
		CustomerID:   "cust-123",
		Timestamp:    time.Now(),
	}
	valSO := marshalEvent("evt-so-123", evSalesOrder)
	if err := env.consumer.handleMessage(ctx, domain.TopicCrmSalesOrderCreated, valSO); err != nil {
		t.Fatalf("failed to process SalesOrderCreated: %v", err)
	}

	// 2. TopicCrmCustomerDemandForecast
	evForecast := domain.CustomerDemandForecastEvent{
		ProductID:        "prod-123",
		ForecastDate:     time.Now().AddDate(0, 1, 0),
		ForecastQuantity: 150,
		ConfidenceLevel:  decimal.NewFromFloat(0.85),
		Timestamp:        time.Now(),
	}
	valF := marshalEvent("evt-f-123", evForecast)
	if err := env.consumer.handleMessage(ctx, domain.TopicCrmCustomerDemandForecast, valF); err != nil {
		t.Fatalf("failed to process CustomerDemandForecast: %v", err)
	}

	// 3. TopicMfgMaterialRequired
	evMaterialReq := domain.MaterialRequiredEvent{
		MaterialID: "prod-123",
		Quantity:   200,
		RequiredBy: time.Now().AddDate(0, 0, 7),
		Timestamp:  time.Now(),
	}
	valMR := marshalEvent("evt-mr-123", evMaterialReq)
	if err := env.consumer.handleMessage(ctx, domain.TopicMfgMaterialRequired, valMR); err != nil {
		t.Fatalf("failed to process MaterialRequired: %v", err)
	}

	// 4. TopicMfgProductionCompleted (RECEIPT 100 units of prod-123)
	evProdCompleted := domain.ProductionCompletedEvent{
		ProductID:        "prod-123",
		QuantityProduced: 100,
		Timestamp:        time.Now(),
	}
	valPC := marshalEvent("evt-pc-123", evProdCompleted)
	if err := env.consumer.handleMessage(ctx, domain.TopicMfgProductionCompleted, valPC); err != nil {
		t.Fatalf("failed to process ProductionCompleted: %v", err)
	}

	// 5. TopicMfgMaterialConsumed (ISSUE 50 units)
	evMaterialConsumed := domain.MaterialConsumedEvent{
		ProductID:         "prod-123",
		Quantity:          decimal.NewFromFloat(50),
		ProductionOrderID: "po-123",
		Timestamp:         time.Now(),
	}
	valMC := marshalEvent("evt-mc-123", evMaterialConsumed)
	if err := env.consumer.handleMessage(ctx, domain.TopicMfgMaterialConsumed, valMC); err != nil {
		t.Fatalf("failed to process MaterialConsumed: %v", err)
	}

	// 6. TopicFinVendorPaymentProcessed
	evVendorPay := domain.VendorPaymentProcessedEvent{
		VendorID:   "vend-123",
		AmountPaid: decimal.NewFromFloat(1500.50),
		Status:     "PROCESSED",
		Timestamp:  time.Now(),
	}
	valVP := marshalEvent("evt-vp-123", evVendorPay)
	if err := env.consumer.handleMessage(ctx, domain.TopicFinVendorPaymentProcessed, valVP); err != nil {
		t.Fatalf("failed to process VendorPaymentProcessed: %v", err)
	}

	// 7. TopicPrjMaterialRequested (ISSUE 30 units)
	evPrjMaterial := domain.MaterialRequestedEvent{
		ProjectID:   "proj-123",
		TaskID:      "task-123",
		ProductID:   "prod-123",
		QtyRequired: 30,
		Timestamp:   time.Now(),
	}
	valPM := marshalEvent("evt-pm-123", evPrjMaterial)
	if err := env.consumer.handleMessage(ctx, domain.TopicPrjMaterialRequested, valPM); err != nil {
		t.Fatalf("failed to process MaterialRequested: %v", err)
	}

	// Idempotency: process TopicCrmSalesOrderCreated again (should skip and return nil)
	if err := env.consumer.handleMessage(ctx, domain.TopicCrmSalesOrderCreated, valSO); err != nil {
		t.Fatalf("idempotent event processing failed: %v", err)
	}

	// Test invalid JSON on TopicCrmSalesOrderCreated
	if err := env.consumer.handleMessage(ctx, domain.TopicCrmSalesOrderCreated, []byte("invalid-json")); err == nil {
		t.Fatalf("expected error for invalid json on TopicCrmSalesOrderCreated, got nil")
	}

	// Test unknown topic
	if err := env.consumer.handleMessage(ctx, "unknown-topic", []byte("{}")); err != nil {
		t.Fatalf("expected nil error for unknown topic, got %v", err)
	}

	// Test publishToDLQ happy path
	env.consumer.publishToDLQ(ctx, "test-topic", "test-key", []byte("test-val"), fmt.Errorf("some error"))

	// Test publishToDLQ fail path
	failPub := &mockPublisher{failPublish: true}
	failConsumer := NewKafkaConsumer([]string{"localhost:9092"}, "scm-group", failPub, env.poSvc, env.invSvc, env.demandSvc, sql.NewSQLKafkaEventInboxRepo(env.db))
	failConsumer.publishToDLQ(ctx, "test-topic", "test-key", []byte("test-val"), fmt.Errorf("some error"))
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

func TestOutboxRelay(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()
	outboxRepo := sql.NewSQLTransactionalOutboxRepo(env.db)

	// 1. NewOutboxRelayWorker defaults
	worker := NewOutboxRelayWorker(outboxRepo, &mockPublisher{}, 0, 0)
	if worker.interval != 5*time.Second {
		t.Errorf("expected default interval, got %v", worker.interval)
	}
	if worker.limit != 100 {
		t.Errorf("expected default limit, got %d", worker.limit)
	}

	// 2. Start with canceled context
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	worker.Start(canceledCtx)

	// 3. processPending - Happy Path
	rec1 := &domain.TransactionalOutbox{
		ID:          "out-1",
		EventType:   "test.event",
		AggregateID: "agg-1",
		Payload:     `{"hello":"world"}`,
		Status:      domain.OutboxStatusPENDING,
		RetryCount:  0,
		CreatedAt:   time.Now(),
	}
	if err := outboxRepo.Create(ctx, rec1); err != nil {
		t.Fatalf("failed to seed outbox: %v", err)
	}

	worker.processPending(ctx)

	// Verify in DB directly that status is SENT
	var dbRec sql.TransactionalOutbox
	if err := env.db.First(&dbRec, "id = ?", "out-1").Error; err != nil {
		t.Fatalf("failed to query db: %v", err)
	}
	if dbRec.Status != domain.OutboxStatusSENT {
		t.Errorf("expected SENT status, got %v", dbRec.Status)
	}

	// 4. processPending - Failure Path
	rec2 := &domain.TransactionalOutbox{
		ID:          "out-2",
		EventType:   "test.event",
		AggregateID: "agg-2",
		Payload:     `{"hello":"world"}`,
		Status:      domain.OutboxStatusPENDING,
		RetryCount:  0,
		CreatedAt:   time.Now(),
	}
	if err := outboxRepo.Create(ctx, rec2); err != nil {
		t.Fatalf("failed to seed outbox: %v", err)
	}

	failPublisher := &mockPublisher{failPublish: true}
	errWorker := NewOutboxRelayWorker(outboxRepo, failPublisher, 5*time.Second, 100)
	errWorker.processPending(ctx)

	// Verify status is FAILED in DB directly
	var dbRec2 sql.TransactionalOutbox
	if err := env.db.First(&dbRec2, "id = ?", "out-2").Error; err != nil {
		t.Fatalf("failed to query db: %v", err)
	}
	if dbRec2.Status != domain.OutboxStatusFAILED {
		t.Errorf("expected FAILED status, got %v", dbRec2.Status)
	}
	if dbRec2.RetryCount != 1 {
		t.Errorf("expected RetryCount to be 1, got %d", dbRec2.RetryCount)
	}

	// 5. processPending - Fetch error path (using canceled context)
	worker.processPending(canceledCtx)
}
