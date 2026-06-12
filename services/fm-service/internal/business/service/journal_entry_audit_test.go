package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/erp-system/fm-service/internal/business/domain"
	"github.com/erp-system/fm-service/internal/business/service"
	"github.com/erp-system/fm-service/internal/data/memory"
	"github.com/shopspring/decimal"
)

func newPostedEntry(t *testing.T, entryRepo *memory.MemoryUniversalJournalEntryRepo, accountRepo *memory.MemoryChartOfAccountsRepo) string {
	t.Helper()
	ctx := context.Background()
	acc := &domain.ChartOfAccounts{
		ID: "acc_a", AccountCode: "1000", AccountName: "Cash", Type: domain.AccountTypeASSET,
		IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	_ = accountRepo.Create(ctx, acc)
	now := time.Now()
	entry := &domain.UniversalJournalEntry{
		ID: "je_p", LegalEntityID: "legal_123", SourceModule: "FM", SourceDocumentID: "doc_p",
		PostingDate: now, FinancialPeriod: now.Format("2006-01"), Status: domain.LedgerStatePOSTED,
		CreatedAt: now, UpdatedAt: now,
	}
	lines := []domain.UniversalJournalLine{
		{ID: "l1", JournalEntryID: "je_p", AccountID: "acc_a", AmountFunctional: decimal.NewFromInt(100), AmountTransactional: decimal.NewFromInt(100), CurrencyTransactional: "USD"},
		{ID: "l2", JournalEntryID: "je_p", AccountID: "acc_a", AmountFunctional: decimal.NewFromInt(-100), AmountTransactional: decimal.NewFromInt(-100), CurrencyTransactional: "USD"},
	}
	if err := entryRepo.Create(ctx, entry, lines); err != nil {
		t.Fatalf("seed entry: %v", err)
	}
	return entry.ID
}

func newDraftEntry(t *testing.T, entryRepo *memory.MemoryUniversalJournalEntryRepo, accountRepo *memory.MemoryChartOfAccountsRepo) string {
	t.Helper()
	ctx := context.Background()
	acc := &domain.ChartOfAccounts{
		ID: "acc_b", AccountCode: "2000", AccountName: "Bank", Type: domain.AccountTypeASSET,
		IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	_ = accountRepo.Create(ctx, acc)
	now := time.Now()
	entry := &domain.UniversalJournalEntry{
		ID: "je_d", LegalEntityID: "legal_123", SourceModule: "FM", SourceDocumentID: "doc_d",
		PostingDate: now, FinancialPeriod: now.Format("2006-01"), Status: domain.LedgerStateDRAFT,
		CreatedAt: now, UpdatedAt: now,
	}
	lines := []domain.UniversalJournalLine{
		{ID: "l3", JournalEntryID: "je_d", AccountID: "acc_b", AmountFunctional: decimal.NewFromInt(50), AmountTransactional: decimal.NewFromInt(50), CurrencyTransactional: "USD"},
		{ID: "l4", JournalEntryID: "je_d", AccountID: "acc_b", AmountFunctional: decimal.NewFromInt(-50), AmountTransactional: decimal.NewFromInt(-50), CurrencyTransactional: "USD"},
	}
	if err := entryRepo.Create(ctx, entry, lines); err != nil {
		t.Fatalf("seed entry: %v", err)
	}
	return entry.ID
}

func TestJournalEntry_Create_SetsPostedStatus(t *testing.T) {
	accounts := memory.NewMemoryChartOfAccountsRepo()
	entries := memory.NewMemoryUniversalJournalEntryRepo()
	outbox := memory.NewMemoryTransactionalOutboxRepo()
	tm := memory.NewMemoryTransactionManager(accounts, entries, outbox)
	svc := service.NewGeneralLedgerService(accounts, entries, outbox, tm)

	accA, _ := svc.CreateAccount(context.Background(), "legal_123", "1000", "Cash", "ASSET")
	accB, _ := svc.CreateAccount(context.Background(), "legal_123", "2000", "Revenue", "REVENUE")

	lines := []domain.UniversalJournalLine{
		{AccountID: accA.ID, AmountFunctional: decimal.NewFromInt(100), AmountTransactional: decimal.NewFromInt(100), CurrencyTransactional: "USD"},
		{AccountID: accB.ID, AmountFunctional: decimal.NewFromInt(-100), AmountTransactional: decimal.NewFromInt(-100), CurrencyTransactional: "USD"},
	}

	entry, err := svc.CreateJournalEntry(context.Background(), "legal_123", "FM", "JE-001", time.Now(), lines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Status != domain.LedgerStatePOSTED {
		t.Errorf("Status = %q, want %q", entry.Status, domain.LedgerStatePOSTED)
	}
}

func TestJournalEntry_Update_BlocksPosted(t *testing.T) {
	accounts := memory.NewMemoryChartOfAccountsRepo()
	entries := memory.NewMemoryUniversalJournalEntryRepo()
	outbox := memory.NewMemoryTransactionalOutboxRepo()
	tm := memory.NewMemoryTransactionManager(accounts, entries, outbox)
	svc := service.NewGeneralLedgerService(accounts, entries, outbox, tm)

	id := newPostedEntry(t, entries, accounts)

	lines := []domain.UniversalJournalLine{
		{AccountID: "acc_a", AmountFunctional: decimal.NewFromInt(200), AmountTransactional: decimal.NewFromInt(200), CurrencyTransactional: "USD"},
		{AccountID: "acc_a", AmountFunctional: decimal.NewFromInt(-200), AmountTransactional: decimal.NewFromInt(-200), CurrencyTransactional: "USD"},
	}
	_, err := svc.UpdateJournalEntry(context.Background(), id, "legal_123", "FM", "doc_p_2", time.Now(), lines)
	if !errors.Is(err, domain.ErrJournalEntryNotMutable) {
		t.Errorf("err = %v, want ErrJournalEntryNotMutable", err)
	}
}

func TestJournalEntry_Update_BlocksReversed(t *testing.T) {
	accounts := memory.NewMemoryChartOfAccountsRepo()
	entries := memory.NewMemoryUniversalJournalEntryRepo()
	outbox := memory.NewMemoryTransactionalOutboxRepo()
	tm := memory.NewMemoryTransactionManager(accounts, entries, outbox)
	svc := service.NewGeneralLedgerService(accounts, entries, outbox, tm)

	id := newPostedEntry(t, entries, accounts)
	entry, lines, _ := entries.GetByID(context.Background(), id)
	entry.Status = domain.LedgerStateREVERSED
	_ = entries.Update(context.Background(), entry, lines)

	lines2 := []domain.UniversalJournalLine{
		{AccountID: "acc_a", AmountFunctional: decimal.NewFromInt(200), AmountTransactional: decimal.NewFromInt(200), CurrencyTransactional: "USD"},
		{AccountID: "acc_a", AmountFunctional: decimal.NewFromInt(-200), AmountTransactional: decimal.NewFromInt(-200), CurrencyTransactional: "USD"},
	}
	_, err := svc.UpdateJournalEntry(context.Background(), id, "legal_123", "FM", "doc_r_2", time.Now(), lines2)
	if !errors.Is(err, domain.ErrJournalEntryNotMutable) {
		t.Errorf("err = %v, want ErrJournalEntryNotMutable", err)
	}
}

func TestJournalEntry_Update_AllowsDraft(t *testing.T) {
	accounts := memory.NewMemoryChartOfAccountsRepo()
	entries := memory.NewMemoryUniversalJournalEntryRepo()
	outbox := memory.NewMemoryTransactionalOutboxRepo()
	tm := memory.NewMemoryTransactionManager(accounts, entries, outbox)
	svc := service.NewGeneralLedgerService(accounts, entries, outbox, tm)

	id := newDraftEntry(t, entries, accounts)

	lines := []domain.UniversalJournalLine{
		{AccountID: "acc_b", AmountFunctional: decimal.NewFromInt(75), AmountTransactional: decimal.NewFromInt(75), CurrencyTransactional: "USD"},
		{AccountID: "acc_b", AmountFunctional: decimal.NewFromInt(-75), AmountTransactional: decimal.NewFromInt(-75), CurrencyTransactional: "USD"},
	}
	updated, err := svc.UpdateJournalEntry(context.Background(), id, "legal_123", "FM", "doc_d_2", time.Now(), lines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.SourceDocumentID != "doc_d_2" {
		t.Errorf("SourceDocumentID = %q, want doc_d_2", updated.SourceDocumentID)
	}
}
