package service_test

import (
	sharedtesting "erp-system/shared/testing"
	"context"
	"testing"
	"time"

	"github.com/erp-system/fm-service/internal/business/domain"
	"github.com/erp-system/fm-service/internal/business/service"
	"github.com/erp-system/fm-service/internal/data/memory"
	"github.com/shopspring/decimal"
)

type MockEvent = sharedtesting.MockEvent

func TestGeneralLedgerService_CreateAccount_PublishesEvent(t *testing.T) {
	accounts := memory.NewMemoryAccountRepo()
	entries := memory.NewMemoryJournalEntryRepo()
	publisher := &sharedtesting.MockPublisher{}

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
	if ev.Topic != domain.TopicFmAccountCreated {
		t.Errorf("expected topic %s, got %s", domain.TopicFmAccountCreated, ev.Topic)
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
	publisher := &sharedtesting.MockPublisher{}

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
	if ev.Topic != domain.TopicFmInvoiceCreated {
		t.Errorf("expected topic %s, got %s", domain.TopicFmInvoiceCreated, ev.Topic)
	}
}

func TestGeneralLedgerService_UpdateJournalEntry(t *testing.T) {
	accounts := memory.NewMemoryAccountRepo()
	entries := memory.NewMemoryJournalEntryRepo()
	publisher := &sharedtesting.MockPublisher{}

	svc := service.NewGeneralLedgerService(accounts, entries, publisher)

	ctx := context.Background()

	// 1. Create two accounts
	accA, err := svc.CreateAccount(ctx, "1000", "Cash", "ASSET", "", "USD")
	if err != nil {
		t.Fatalf("failed to create Account A: %v", accA)
	}
	accB, err := svc.CreateAccount(ctx, "4000", "Revenue", "REVENUE", "", "USD")
	if err != nil {
		t.Fatalf("failed to create Account B: %v", accB)
	}

	// 2. Create balanced journal entry
	lines := []domain.JournalEntryLine{
		{
			AccountID:    accA.ID,
			DebitAmount:  decimal.NewFromInt(100),
			CreditAmount: decimal.Zero,
		},
		{
			AccountID:    accB.ID,
			DebitAmount:  decimal.Zero,
			CreditAmount: decimal.NewFromInt(100),
		},
	}

	entry, err := svc.CreateJournalEntry(ctx, "REF-001", "Initial sales", lines)
	if err != nil {
		t.Fatalf("unexpected error creating entry: %v", err)
	}

	storedEntry, storedLines, _ := entries.GetByID(ctx, entry.ID)
	storedEntry.Status = string(domain.JournalEntryStatusPending)
	if err := entries.Update(ctx, storedEntry, storedLines); err != nil {
		t.Fatalf("failed to flip entry to PENDING: %v", err)
	}

	// Verify initial balances
	accA, _ = accounts.GetByID(ctx, accA.ID)
	accB, _ = accounts.GetByID(ctx, accB.ID)
	if !accA.Balance.Equal(decimal.NewFromInt(100)) {
		t.Errorf("expected Account A balance 100, got %s", accA.Balance)
	}
	if !accB.Balance.Equal(decimal.NewFromInt(100)) {
		t.Errorf("expected Account B balance 100, got %s", accB.Balance)
	}

	// 3. Update with new balanced lines (Debit A 150, Credit B 150)
	newLines := []domain.JournalEntryLine{
		{
			AccountID:    accA.ID,
			DebitAmount:  decimal.NewFromInt(150),
			CreditAmount: decimal.Zero,
		},
		{
			AccountID:    accB.ID,
			DebitAmount:  decimal.Zero,
			CreditAmount: decimal.NewFromInt(150),
		},
	}

	_, err = svc.UpdateJournalEntry(ctx, entry.ID, "REF-001-REV", "Revised sales", newLines)
	if err != nil {
		t.Fatalf("unexpected error updating entry: %v", err)
	}

	// Verify updated balances (old reversed, new applied)
	accA, _ = accounts.GetByID(ctx, accA.ID)
	accB, _ = accounts.GetByID(ctx, accB.ID)
	if !accA.Balance.Equal(decimal.NewFromInt(150)) {
		t.Errorf("expected Account A updated balance 150, got %s", accA.Balance)
	}
	if !accB.Balance.Equal(decimal.NewFromInt(150)) {
		t.Errorf("expected Account B updated balance 150, got %s", accB.Balance)
	}

	// 4. Try updating with unbalanced lines (should fail, balances should remain 150/150)
	unbalancedLines := []domain.JournalEntryLine{
		{
			AccountID:    accA.ID,
			DebitAmount:  decimal.NewFromInt(200),
			CreditAmount: decimal.Zero,
		},
		{
			AccountID:    accB.ID,
			DebitAmount:  decimal.Zero,
			CreditAmount: decimal.NewFromInt(150),
		},
	}

	_, err = svc.UpdateJournalEntry(ctx, entry.ID, "REF-001-REV2", "Unbalanced update", unbalancedLines)
	if err == nil {
		t.Errorf("expected error when updating with unbalanced lines, got nil")
	}

	// Verify balances did not change
	accA, _ = accounts.GetByID(ctx, accA.ID)
	accB, _ = accounts.GetByID(ctx, accB.ID)
	if !accA.Balance.Equal(decimal.NewFromInt(150)) {
		t.Errorf("expected Account A balance to stay 150, got %s", accA.Balance)
	}
	if !accB.Balance.Equal(decimal.NewFromInt(150)) {
		t.Errorf("expected Account B balance to stay 150, got %s", accB.Balance)
	}

	// 5. Try updating with non-existent account ID to trigger rollback
	invalidLines := []domain.JournalEntryLine{
		{
			AccountID:    "non_existent",
			DebitAmount:  decimal.NewFromInt(200),
			CreditAmount: decimal.Zero,
		},
		{
			AccountID:    accB.ID,
			DebitAmount:  decimal.Zero,
			CreditAmount: decimal.NewFromInt(200),
		},
	}

	_, err = svc.UpdateJournalEntry(ctx, entry.ID, "REF-001-REV3", "Invalid account update", invalidLines)
	if err == nil {
		t.Errorf("expected error when updating with invalid account, got nil")
	}

	// Verify balances did not change and were rolled back
	accA, _ = accounts.GetByID(ctx, accA.ID)
	accB, _ = accounts.GetByID(ctx, accB.ID)
	if !accA.Balance.Equal(decimal.NewFromInt(150)) {
		t.Errorf("expected Account A balance rolled back to 150, got %s", accA.Balance)
	}
	if !accB.Balance.Equal(decimal.NewFromInt(150)) {
		t.Errorf("expected Account B balance rolled back to 150, got %s", accB.Balance)
	}
}
