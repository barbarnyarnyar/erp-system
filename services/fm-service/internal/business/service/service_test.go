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

func TestGeneralLedgerService_CreateAccount_PublishesEvent(t *testing.T) {
	accounts := memory.NewMemoryChartOfAccountsRepo()
	entries := memory.NewMemoryUniversalJournalEntryRepo()
	outbox := memory.NewMemoryTransactionalOutboxRepo()
	tm := memory.NewMemoryTransactionManager(accounts, entries, outbox)

	svc := service.NewGeneralLedgerService(accounts, entries, outbox, tm)

	acc, err := svc.CreateAccount(context.Background(), "legal_123", "1000", "Cash", "ASSET")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if acc.AccountName != "Cash" {
		t.Errorf("expected account name Cash, got %s", acc.AccountName)
	}

	pending, err := outbox.GetPending(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(pending) != 1 {
		t.Fatalf("expected 1 published event in outbox, got %d", len(pending))
	}

	ev := pending[0]
	if ev.EventType != string(domain.TopicFmAccountCreated) {
		t.Errorf("expected topic %s, got %s", domain.TopicFmAccountCreated, ev.EventType)
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
	invoices := memory.NewMemoryArInvoiceRepo()
	credits := memory.NewMemoryCustomerCreditRepo()
	outbox := memory.NewMemoryTransactionalOutboxRepo()
	tm := memory.NewMemoryTransactionManager(invoices, outbox)

	svc := service.NewAccountsReceivableService(invoices, credits, outbox, tm)

	inv, err := svc.CreateInvoice(context.Background(), "legal_123", "cust_123", "so_123", decimal.NewFromInt(750), decimal.NewFromInt(50), time.Now().AddDate(0, 0, 30))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !inv.TotalAmount.Equal(decimal.NewFromInt(750)) {
		t.Errorf("expected total amount 750, got %s", inv.TotalAmount)
	}

	pending, err := outbox.GetPending(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(pending) != 1 {
		t.Fatalf("expected 1 published event in outbox, got %d", len(pending))
	}

	ev := pending[0]
	if ev.EventType != string(domain.TopicFmInvoiceCreated) {
		t.Errorf("expected topic %s, got %s", domain.TopicFmInvoiceCreated, ev.EventType)
	}
}

func TestGeneralLedgerService_UpdateJournalEntry(t *testing.T) {
	accounts := memory.NewMemoryChartOfAccountsRepo()
	entries := memory.NewMemoryUniversalJournalEntryRepo()
	outbox := memory.NewMemoryTransactionalOutboxRepo()
	tm := memory.NewMemoryTransactionManager(accounts, entries, outbox)

	svc := service.NewGeneralLedgerService(accounts, entries, outbox, tm)

	ctx := context.Background()

	// 1. Create two accounts
	accA, err := svc.CreateAccount(ctx, "legal_123", "1000", "Cash", "ASSET")
	if err != nil {
		t.Fatalf("failed to create Account A: %v", err)
	}
	accB, err := svc.CreateAccount(ctx, "legal_123", "4000", "Revenue", "REVENUE")
	if err != nil {
		t.Fatalf("failed to create Account B: %v", err)
	}

	// Clear outbox from account creations to isolate journal entry events
	*outbox = *memory.NewMemoryTransactionalOutboxRepo()

	// 2. Create draft journal entry in repo
	draftEntry := &domain.UniversalJournalEntry{
		ID:               "je_draft",
		LegalEntityID:    "legal_123",
		SourceModule:     "FM",
		SourceDocumentID: "doc_123",
		PostingDate:      time.Now(),
		FinancialPeriod:  time.Now().Format("2006-01"),
		Status:           domain.LedgerStateDRAFT,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	draftLines := []domain.UniversalJournalLine{
		{
			ID:                    "jel_1",
			JournalEntryID:        draftEntry.ID,
			AccountID:             accA.ID,
			AmountFunctional:      decimal.NewFromInt(100),
			AmountTransactional:   decimal.NewFromInt(100),
			CurrencyTransactional: "USD",
		},
		{
			ID:                    "jel_2",
			JournalEntryID:        draftEntry.ID,
			AccountID:             accB.ID,
			AmountFunctional:      decimal.NewFromInt(-100),
			AmountTransactional:   decimal.NewFromInt(-100),
			CurrencyTransactional: "USD",
		},
	}
	_ = entries.Create(ctx, draftEntry, draftLines)

	// Verify dynamic balances (DRAFT entry shouldn't affect dynamic balance)
	balA, _ := svc.GetAccountBalance(ctx, accA.ID)
	balB, _ := svc.GetAccountBalance(ctx, accB.ID)
	if !balA.Equal(decimal.Zero) {
		t.Errorf("expected Account A balance 0, got %s", balA)
	}
	if !balB.Equal(decimal.Zero) {
		t.Errorf("expected Account B balance 0, got %s", balB)
	}

	// 3. Update with new balanced lines (Debit A 150, Credit B 150)
	newLines := []domain.UniversalJournalLine{
		{
			AccountID:             accA.ID,
			AmountFunctional:      decimal.NewFromInt(150),
			AmountTransactional:   decimal.NewFromInt(150),
			CurrencyTransactional: "USD",
		},
		{
			AccountID:             accB.ID,
			AmountFunctional:      decimal.NewFromInt(-150),
			AmountTransactional:   decimal.NewFromInt(-150),
			CurrencyTransactional: "USD",
		},
	}

	_, err = svc.UpdateJournalEntry(ctx, draftEntry.ID, "legal_123", "FM", "doc_123", time.Now(), newLines)
	if err != nil {
		t.Fatalf("unexpected error updating entry: %v", err)
	}

	// 4. Try updating with unbalanced lines (should fail)
	unbalancedLines := []domain.UniversalJournalLine{
		{
			AccountID:             accA.ID,
			AmountFunctional:      decimal.NewFromInt(200),
			AmountTransactional:   decimal.NewFromInt(200),
			CurrencyTransactional: "USD",
		},
		{
			AccountID:             accB.ID,
			AmountFunctional:      decimal.NewFromInt(-150),
			AmountTransactional:   decimal.NewFromInt(-150),
			CurrencyTransactional: "USD",
		},
	}

	_, err = svc.UpdateJournalEntry(ctx, draftEntry.ID, "legal_123", "FM", "doc_123", time.Now(), unbalancedLines)
	if err == nil {
		t.Errorf("expected error when updating with unbalanced lines, got nil")
	}

	// 5. Try updating with non-existent account ID to trigger rollback
	invalidLines := []domain.UniversalJournalLine{
		{
			AccountID:             "non_existent",
			AmountFunctional:      decimal.NewFromInt(200),
			AmountTransactional:   decimal.NewFromInt(200),
			CurrencyTransactional: "USD",
		},
		{
			AccountID:             accB.ID,
			AmountFunctional:      decimal.NewFromInt(-200),
			AmountTransactional:   decimal.NewFromInt(-200),
			CurrencyTransactional: "USD",
		},
	}

	_, err = svc.UpdateJournalEntry(ctx, draftEntry.ID, "legal_123", "FM", "doc_123", time.Now(), invalidLines)
	if err == nil {
		t.Errorf("expected error when updating with invalid account, got nil")
	}
}
