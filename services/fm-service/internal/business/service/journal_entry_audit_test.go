package service_test

import (
	sharedtesting "erp-system/shared/testing"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/erp-system/fm-service/internal/business/domain"
	"github.com/erp-system/fm-service/internal/business/service"
	"github.com/erp-system/fm-service/internal/data/memory"
	"github.com/shopspring/decimal"
)

func newPostedEntry(t *testing.T, entryRepo *memory.MemoryJournalEntryRepo, accountRepo *memory.MemoryAccountRepo) string {
	t.Helper()
	ctx := context.Background()
	acc := &domain.Account{
		ID: "acc_a", AccountNumber: "1000", Name: "Cash", Type: domain.AccountTypeAsset,
		Balance: decimal.Zero, Currency: "USD", IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	_ = accountRepo.Create(ctx, acc)
	now := time.Now()
	entry := &domain.JournalEntry{
		ID: "je_p", Reference: "REF-P", Date: now, Description: "Posted", Status: string(domain.JournalEntryStatusPosted),
		CreatedBy: "system", PostedBy: "system", PostedAt: &now, CreatedAt: now,
	}
	lines := []domain.JournalEntryLine{
		{ID: "l1", EntryID: "je_p", AccountID: "acc_a", DebitAmount: decimal.NewFromInt(100), CreditAmount: decimal.Zero},
		{ID: "l2", EntryID: "je_p", AccountID: "acc_a", DebitAmount: decimal.Zero, CreditAmount: decimal.NewFromInt(100)},
	}
	if err := entryRepo.Create(ctx, entry, lines); err != nil {
		t.Fatalf("seed entry: %v", err)
	}
	return entry.ID
}

func newPendingEntry(t *testing.T, entryRepo *memory.MemoryJournalEntryRepo, accountRepo *memory.MemoryAccountRepo) string {
	t.Helper()
	ctx := context.Background()
	acc := &domain.Account{
		ID: "acc_b", AccountNumber: "2000", Name: "Bank", Type: domain.AccountTypeAsset,
		Balance: decimal.Zero, Currency: "USD", IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	_ = accountRepo.Create(ctx, acc)
	now := time.Now()
	entry := &domain.JournalEntry{
		ID: "je_d", Reference: "REF-D", Date: now, Description: "Draft", Status: string(domain.JournalEntryStatusPending),
		CreatedBy: "user1", CreatedAt: now,
	}
	lines := []domain.JournalEntryLine{
		{ID: "l3", EntryID: "je_d", AccountID: "acc_b", DebitAmount: decimal.NewFromInt(50), CreditAmount: decimal.Zero},
		{ID: "l4", EntryID: "je_d", AccountID: "acc_b", DebitAmount: decimal.Zero, CreditAmount: decimal.NewFromInt(50)},
	}
	if err := entryRepo.Create(ctx, entry, lines); err != nil {
		t.Fatalf("seed entry: %v", err)
	}
	return entry.ID
}

func TestJournalEntry_Create_SetsPostedByAndPostedAt(t *testing.T) {
	accounts := memory.NewMemoryAccountRepo()
	entries := memory.NewMemoryJournalEntryRepo()
	publisher := &sharedtesting.MockPublisher{}
	svc := service.NewGeneralLedgerService(accounts, entries, publisher)

	accA, _ := svc.CreateAccount(context.Background(), "1000", "Cash", "ASSET", "", "USD")
	accB, _ := svc.CreateAccount(context.Background(), "2000", "Revenue", "REVENUE", "", "USD")

	lines := []domain.JournalEntryLine{
		{AccountID: accA.ID, DebitAmount: decimal.NewFromInt(100), CreditAmount: decimal.Zero},
		{AccountID: accB.ID, DebitAmount: decimal.Zero, CreditAmount: decimal.NewFromInt(100)},
	}

	entry, err := svc.CreateJournalEntry(context.Background(), "JE-001", "Sale", lines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.PostedBy != "system" {
		t.Errorf("PostedBy = %q, want %q", entry.PostedBy, "system")
	}
	if entry.PostedAt == nil {
		t.Errorf("PostedAt should be set")
	}
	if entry.Status != string(domain.JournalEntryStatusPosted) {
		t.Errorf("Status = %q, want %q", entry.Status, domain.JournalEntryStatusPosted)
	}
}

func TestJournalEntry_Update_BlocksPosted(t *testing.T) {
	accounts := memory.NewMemoryAccountRepo()
	entries := memory.NewMemoryJournalEntryRepo()
	publisher := &sharedtesting.MockPublisher{}
	svc := service.NewGeneralLedgerService(accounts, entries, publisher)

	id := newPostedEntry(t, entries, accounts)

	lines := []domain.JournalEntryLine{
		{AccountID: "acc_a", DebitAmount: decimal.NewFromInt(200), CreditAmount: decimal.Zero},
		{AccountID: "acc_a", DebitAmount: decimal.Zero, CreditAmount: decimal.NewFromInt(200)},
	}
	_, err := svc.UpdateJournalEntry(context.Background(), id, "REF-P-2", "Edited", lines)
	if !errors.Is(err, domain.ErrJournalEntryNotMutable) {
		t.Errorf("err = %v, want ErrJournalEntryNotMutable", err)
	}
}

func TestJournalEntry_Update_BlocksReversed(t *testing.T) {
	accounts := memory.NewMemoryAccountRepo()
	entries := memory.NewMemoryJournalEntryRepo()
	publisher := &sharedtesting.MockPublisher{}
	svc := service.NewGeneralLedgerService(accounts, entries, publisher)

	id := newPostedEntry(t, entries, accounts)
	entry, _, _ := entries.GetByID(context.Background(), id)
	entry.Status = string(domain.JournalEntryStatusReversed)
	_ = entries.Update(context.Background(), entry, nil)

	lines := []domain.JournalEntryLine{
		{AccountID: "acc_a", DebitAmount: decimal.NewFromInt(200), CreditAmount: decimal.Zero},
		{AccountID: "acc_a", DebitAmount: decimal.Zero, CreditAmount: decimal.NewFromInt(200)},
	}
	_, err := svc.UpdateJournalEntry(context.Background(), id, "REF-R-2", "Edited", lines)
	if !errors.Is(err, domain.ErrJournalEntryNotMutable) {
		t.Errorf("err = %v, want ErrJournalEntryNotMutable", err)
	}
}

func TestJournalEntry_Update_AllowsPending(t *testing.T) {
	accounts := memory.NewMemoryAccountRepo()
	entries := memory.NewMemoryJournalEntryRepo()
	publisher := &sharedtesting.MockPublisher{}
	svc := service.NewGeneralLedgerService(accounts, entries, publisher)

	id := newPendingEntry(t, entries, accounts)

	lines := []domain.JournalEntryLine{
		{AccountID: "acc_b", DebitAmount: decimal.NewFromInt(75), CreditAmount: decimal.Zero},
		{AccountID: "acc_b", DebitAmount: decimal.Zero, CreditAmount: decimal.NewFromInt(75)},
	}
	updated, err := svc.UpdateJournalEntry(context.Background(), id, "REF-D-2", "Draft edited", lines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Reference != "REF-D-2" {
		t.Errorf("Reference = %q, want REF-D-2", updated.Reference)
	}
}
