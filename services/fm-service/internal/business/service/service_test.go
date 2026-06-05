package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/erp-system/fm-service/internal/business/domain"
	"github.com/erp-system/fm-service/internal/business/service"
	"github.com/erp-system/fm-service/internal/data/memory"
	"github.com/shopspring/decimal"
)

// MockPublisher tracks published events for assertion
type MockPublisher struct {
	Events []MockEvent
}

type MockEvent struct {
	Topic   string
	Key     string
	Payload interface{}
}

func (m *MockPublisher) Publish(ctx context.Context, topic string, key string, payload interface{}) error {
	m.Events = append(m.Events, MockEvent{
		Topic:   topic,
		Key:     key,
		Payload: payload,
	})
	return nil
}

func TestGeneralLedgerService_CreateAccount_PublishesEvent(t *testing.T) {
	accounts := memory.NewMemoryAccountRepo()
	entries := memory.NewMemoryJournalEntryRepo()
	publisher := &MockPublisher{}

	svc := service.NewGeneralLedgerService(accounts, entries, publisher)

	acc, err := svc.CreateAccount(context.Background(), "1000", "Cash", "ASSET", "", "USD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if acc.Name != "Cash" {
		t.Errorf("expected account name Cash, got %s", acc.Name)
	}

	if len(publisher.Events) != 1 {
		t.Fatalf("expected 1 published event, got %d", len(publisher.Events))
	}

	ev := publisher.Events[0]
	if ev.Topic != "fin.account.created" {
		t.Errorf("expected topic fin.account.created, got %s", ev.Topic)
	}

	payload, ok := ev.Payload.(domain.AccountEventPayload)
	if !ok {
		t.Fatalf("expected payload type domain.AccountEventPayload")
	}

	if payload.AccountNumber != "1000" {
		t.Errorf("expected payload AccountNumber 1000, got %s", payload.AccountNumber)
	}
}

func TestAccountsReceivableService_CreateInvoice_PublishesEvent(t *testing.T) {
	invoices := memory.NewMemoryInvoiceRepo()
	publisher := &MockPublisher{}

	svc := service.NewAccountsReceivableService(invoices, publisher)

	lines := []domain.InvoiceLine{
		{
			Description: "Consulting",
			Quantity:    5,
			UnitPrice:   decimal.NewFromInt(150),
			LineTotal:   decimal.NewFromInt(750),
		},
	}

	inv, err := svc.CreateInvoice(context.Background(), "cust_123", time.Now(), time.Now().AddDate(0, 0, 30), lines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !inv.TotalAmount.Equal(decimal.NewFromInt(750)) {
		t.Errorf("expected total amount 750, got %s", inv.TotalAmount)
	}

	if len(publisher.Events) != 1 {
		t.Fatalf("expected 1 published event, got %d", len(publisher.Events))
	}

	ev := publisher.Events[0]
	if ev.Topic != "fin.invoice.created" {
		t.Errorf("expected topic fin.invoice.created, got %s", ev.Topic)
	}
}

