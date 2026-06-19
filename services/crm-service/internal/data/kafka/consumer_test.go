package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"erp-system/shared/utils"
	"github.com/erp-system/crm-service/internal/business/domain"
	"github.com/erp-system/crm-service/internal/business/service"
	"github.com/erp-system/crm-service/internal/data/memory"
	"github.com/shopspring/decimal"
)

func init() {
	utils.InitLogger("crm-consumer-test")
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
	consumer       *KafkaConsumer
	orderSvc       *service.SalesOrderService
	leadSvc        *service.LeadService
	oppSvc         *service.OpportunityService
	interactionSvc *service.CustomerInteractionService
	oppRepo        *memory.OpportunityRepository
	orderRepo      *memory.SalesOrderRepository
	interactRepo   *memory.CustomerInteractionRepository
}

func setupTestEnv() *testEnv {
	custRepo := memory.NewCustomerRepository()
	leadRepo := memory.NewLeadRepository()
	oppRepo := memory.NewOpportunityRepository()
	orderRepo := memory.NewSalesOrderRepository()
	orderLineRepo := memory.NewSalesOrderLineRepository()
	historyRepo := memory.NewOpportunityStageHistoryRepository()
	interactRepo := memory.NewCustomerInteractionRepository()

	publisher := &mockPublisher{}

	custSvc := service.NewCustomerService(custRepo, publisher)
	oppSvc := service.NewOpportunityService(oppRepo, historyRepo, publisher)
	leadSvc := service.NewLeadService(leadRepo, custSvc, oppSvc, publisher)
	orderSvc := service.NewSalesOrderService(orderRepo, orderLineRepo, custRepo, publisher)
	interactionSvc := service.NewCustomerInteractionService(interactRepo, publisher)

	consumer := NewKafkaConsumer([]string{"localhost:9092"}, "crm-group", publisher, orderSvc, leadSvc, oppSvc, interactionSvc)

	return &testEnv{
		consumer:       consumer,
		orderSvc:       orderSvc,
		leadSvc:        leadSvc,
		oppSvc:         oppSvc,
		interactionSvc: interactionSvc,
		oppRepo:        oppRepo,
		orderRepo:      orderRepo,
		interactRepo:   interactRepo,
	}
}

func TestConsumer_AllEvents(t *testing.T) {
	env := setupTestEnv()
	ctx := context.Background()

	// Seed data
	// 1. Seed opportunity for InventoryAvailable
	opp := &domain.Opportunity{
		ID:          "opp-1",
		CustomerID:  "cust-1",
		Title:       "New Sale",
		Value:       decimal.NewFromFloat(1000),
		Status:      "NEW",
		Stage:       "QUALIFICATION",
		Probability: decimal.NewFromFloat(0.1),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	_ = env.oppRepo.Create(ctx, opp)

	// 2. Seed sales orders
	order1 := &domain.SalesOrder{
		ID:         "order-1",
		CustomerID: "cust-1",
		Status:     "DRAFT",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	_ = env.orderRepo.Create(ctx, order1)

	order2 := &domain.SalesOrder{
		ID:         "order-2",
		CustomerID: "cust-1",
		Status:     "DRAFT",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	_ = env.orderRepo.Create(ctx, order2)

	// A. TopicScmInventoryAvailable
	evInv := domain.InventoryAvailableEvent{
		ProductID:      "prod-123",
		QuantityOnHand: decimal.NewFromFloat(100),
		Timestamp:      time.Now(),
	}
	valInv, _ := json.Marshal(evInv)
	if err := env.consumer.handleMessage(ctx, domain.TopicScmInventoryAvailable, valInv); err != nil {
		t.Fatalf("failed to process InventoryAvailable: %v", err)
	}

	// B. TopicScmShipmentDelivered
	evShip := domain.ShipmentDeliveredEvent{
		ShipmentID:   "ship-1",
		SalesOrderID: "order-1",
		Timestamp:    time.Now(),
	}
	valShip, _ := json.Marshal(evShip)
	if err := env.consumer.handleMessage(ctx, domain.TopicScmShipmentDelivered, valShip); err != nil {
		t.Fatalf("failed to process ShipmentDelivered: %v", err)
	}

	// Verify order-1 status is DELIVERED
	o1, _ := env.orderRepo.GetByID(ctx, "order-1")
	if o1.Status != "DELIVERED" {
		t.Errorf("expected order-1 status to be DELIVERED, got %s", o1.Status)
	}

	// C. TopicFmPaymentReceived
	evPay := domain.PaymentReceivedEvent{
		InvoiceID:   "inv-1",
		ReferenceID: "ref-1",
		Amount:      decimal.NewFromFloat(500),
		Timestamp:   time.Now(),
	}
	valPay, _ := json.Marshal(evPay)
	if err := env.consumer.handleMessage(ctx, domain.TopicFmPaymentReceived, valPay); err != nil {
		t.Fatalf("failed to process PaymentReceived: %v", err)
	}

	// D. TopicFmCreditCheckCompleted
	evCredit := domain.CreditCheckCompletedEvent{
		CustomerID:   "cust-1",
		CreditStatus: "APPROVED",
		Timestamp:    time.Now(),
	}
	valCredit, _ := json.Marshal(evCredit)
	if err := env.consumer.handleMessage(ctx, domain.TopicFmCreditCheckCompleted, valCredit); err != nil {
		t.Fatalf("failed to process CreditCheckCompleted: %v", err)
	}

	// Verify order-2 status is CONFIRMED (since it was DRAFT and for cust-1)
	o2, _ := env.orderRepo.GetByID(ctx, "order-2")
	if o2.Status != "CONFIRMED" {
		t.Errorf("expected order-2 status to be CONFIRMED, got %s", o2.Status)
	}

	// E. TopicMfgProductionCompleted
	evProd := domain.ProductionCompletedEvent{
		ProductionOrderID: "mfg-order-1",
		ProductID:         "prod-123",
		Quantity:          50,
		Timestamp:         time.Now(),
	}
	valProd, _ := json.Marshal(evProd)
	if err := env.consumer.handleMessage(ctx, domain.TopicMfgProductionCompleted, valProd); err != nil {
		t.Fatalf("failed to process ProductionCompleted: %v", err)
	}

	// F. TopicPrjProjectCompleted
	evProj := domain.ProjectCompletedEvent{
		ProjectID: "proj-1",
		Timestamp: time.Now(),
	}
	valProj, _ := json.Marshal(evProj)
	if err := env.consumer.handleMessage(ctx, domain.TopicPrjProjectCompleted, valProj); err != nil {
		t.Fatalf("failed to process ProjectCompleted: %v", err)
	}

	// G. TopicHrEmployeePerformance
	evPerf := domain.EmployeePerformanceEvent{
		EmployeeID: "emp-1",
		Rating:     decimal.NewFromFloat(4.5),
		Timestamp:  time.Now(),
	}
	valPerf, _ := json.Marshal(evPerf)
	if err := env.consumer.handleMessage(ctx, domain.TopicHrEmployeePerformance, valPerf); err != nil {
		t.Fatalf("failed to process EmployeePerformance: %v", err)
	}

	// Test unmarshal failures for error paths
	topics := []string{
		domain.TopicScmInventoryAvailable,
		domain.TopicScmShipmentDelivered,
		domain.TopicFmPaymentReceived,
		domain.TopicFmCreditCheckCompleted,
		domain.TopicMfgProductionCompleted,
		domain.TopicPrjProjectCompleted,
		domain.TopicHrEmployeePerformance,
	}
	for _, topic := range topics {
		if err := env.consumer.handleMessage(ctx, topic, []byte("invalid-json")); err == nil {
			t.Errorf("expected error for invalid json on topic %s", topic)
		}
	}

	// Test unknown topic
	if err := env.consumer.handleMessage(ctx, "unknown-topic", []byte("{}")); err != nil {
		t.Fatalf("expected nil error for unknown topic, got %v", err)
	}

	// Test DLQ Publish happy path
	env.consumer.publishToDLQ(ctx, "test-topic", "test-key", []byte("test-val"), fmt.Errorf("some error"))

	// Test DLQ Publish error path
	failPub := &mockPublisher{failPublish: true}
	failConsumer := NewKafkaConsumer([]string{"localhost:9092"}, "crm-group", failPub, env.orderSvc, env.leadSvc, env.oppSvc, env.interactionSvc)
	failConsumer.publishToDLQ(ctx, "test-topic", "test-key", []byte("test-val"), fmt.Errorf("some error"))
}

func TestConsumer_StartAndClose(t *testing.T) {
	env := setupTestEnv()

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
