package kafka

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/erp-system/fm-service/internal/business/domain"
	"github.com/erp-system/fm-service/internal/business/service"
	"github.com/erp-system/fm-service/internal/data/memory"
)

type mockEventPublisher struct{}

func (p *mockEventPublisher) Publish(ctx context.Context, topic string, key string, payload interface{}) error {
	return nil
}

func TestKafkaConsumer_Idempotency(t *testing.T) {
	// Initialize memory repos
	accounts := memory.NewMemoryChartOfAccountsRepo()
	entries := memory.NewMemoryUniversalJournalEntryRepo()
	invoices := memory.NewMemoryArInvoiceRepo()
	payments := memory.NewMemoryPaymentRepo()
	statements := memory.NewMemoryBankStatementRepo()
	bills := memory.NewMemoryApVendorBillRepo()
	outbox := memory.NewMemoryTransactionalOutboxRepo()
	inbox := memory.NewMemoryKafkaEventInboxRepo()

	tmGL := memory.NewMemoryTransactionManager(accounts, entries, outbox)
	glSvc := service.NewGeneralLedgerService(accounts, entries, outbox, tmGL)

	tmAR := memory.NewMemoryTransactionManager(invoices, outbox)
	arSvc := service.NewAccountsReceivableService(invoices, outbox, tmAR)

	tmAP := memory.NewMemoryTransactionManager(bills, outbox)
	apSvc := service.NewAccountsPayableService(bills, outbox, tmAP)

	tmCM := memory.NewMemoryTransactionManager(payments, invoices, outbox)
	cmSvc := service.NewCashManagementService(payments, invoices, statements, outbox, tmCM)

	budgetSvc := service.NewBudgetingService(memory.NewMemoryBudgetRepo(), accounts, entries, outbox, tmGL)

	publisher := &mockEventPublisher{}

	consumer := NewKafkaConsumer(
		[]string{"localhost:9092"},
		"test-group",
		publisher,
		glSvc,
		apSvc,
		arSvc,
		cmSvc,
		budgetSvc,
		inbox,
	)

	ctx := context.Background()

	// 1. Create a customer event (to check processing logic)
	// CRM customer created event
	custEvent := map[string]interface{}{
		"event_id":      "evt_cust_123",
		"customer_id":   "cust_12345678",
		"customer_name": "Acme Corp",
		"email":         "acme@example.com",
		"timestamp":     time.Now().Format(time.RFC3339),
	}
	payloadBytes, _ := json.Marshal(custEvent)

	// Process the event first time
	err := consumer.handleMessage(ctx, domain.TopicCrmCustomerCreated, payloadBytes)
	if err != nil {
		t.Fatalf("failed to process customer created event first time: %v", err)
	}

	// Verify account was created
	accList, err := accounts.List(ctx)
	if err != nil || len(accList) != 1 {
		t.Fatalf("expected 1 account to be created, got %d, err: %v", len(accList), err)
	}

	// Process the event second time (should be deduplicated/skipped)
	err = consumer.handleMessage(ctx, domain.TopicCrmCustomerCreated, payloadBytes)
	if err != nil {
		t.Fatalf("failed to skip customer created event second time: %v", err)
	}

	// Verify that NO duplicate account was created
	accList, err = accounts.List(ctx)
	if err != nil || len(accList) != 1 {
		t.Errorf("expected still 1 account (skipped duplicate), but got %d", len(accList))
	}

	// Verify the inbox has the successful record
	inboxRec, err := inbox.GetByID(ctx, "evt_cust_123")
	if err != nil || inboxRec.ProcessingStatus != domain.EventProcessingStatusSUCCESS {
		t.Errorf("expected successful inbox record, got status %s, err: %v", inboxRec.ProcessingStatus, err)
	}
}
