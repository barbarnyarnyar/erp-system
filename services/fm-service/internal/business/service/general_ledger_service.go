package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/erp-system/fm-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type GeneralLedgerService struct {
	accounts  domain.AccountRepository
	entries   domain.JournalEntryRepository
	publisher domain.EventPublisher
}

func NewGeneralLedgerService(accounts domain.AccountRepository, entries domain.JournalEntryRepository, publisher domain.EventPublisher) *GeneralLedgerService {
	return &GeneralLedgerService{
		accounts:  accounts,
		entries:   entries,
		publisher: publisher,
	}
}

func (s *GeneralLedgerService) ListAccounts(ctx context.Context) ([]domain.Account, error) {
	return s.accounts.List(ctx)
}

func (s *GeneralLedgerService) CreateAccount(ctx context.Context, accNum, name, accType, parentID, currency string) (*domain.Account, error) {
	if accNum == "" || name == "" || accType == "" {
		return nil, errors.New("account number, name, and type are required")
	}

	id := fmt.Sprintf("acc_%d", time.Now().UnixNano())
	acc := &domain.Account{
		ID:            id,
		AccountNumber: accNum,
		Name:          name,
		Type:          accType,
		Balance:       decimal.Zero,
		Currency:      currency,
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if parentID != "" {
		acc.ParentID = &parentID
	}

	err := s.accounts.Create(ctx, acc)
	if err != nil {
		return nil, err
	}
	
	// Publish event
	_ = s.publisher.Publish(ctx, "fin.account.created", acc.ID, domain.AccountEventPayload{
		ID:            acc.ID,
		AccountNumber: acc.AccountNumber,
		Name:          acc.Name,
		Type:          string(acc.Type),
		Balance:       acc.Balance,
		Currency:      acc.Currency,
		Timestamp:     time.Now(),
	})
	
	return acc, nil
}

func (s *GeneralLedgerService) GetAccount(ctx context.Context, id string) (*domain.Account, error) {
	return s.accounts.GetByID(ctx, id)
}

func (s *GeneralLedgerService) GetAccountByNumber(ctx context.Context, accNum string) (*domain.Account, error) {
	return s.accounts.GetByNumber(ctx, accNum)
}

func (s *GeneralLedgerService) UpdateAccount(ctx context.Context, id, name, accType, parentID string, isActive bool) (*domain.Account, error) {
	acc, err := s.accounts.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	acc.Name = name
	acc.Type = accType
	acc.IsActive = isActive
	acc.UpdatedAt = time.Now()

	if parentID != "" {
		acc.ParentID = &parentID
	} else {
		acc.ParentID = nil
	}

	err = s.accounts.Update(ctx, acc)
	if err != nil {
		return nil, err
	}
	
	// Publish event
	_ = s.publisher.Publish(ctx, "fin.account.updated", acc.ID, domain.AccountEventPayload{
		ID:            acc.ID,
		AccountNumber: acc.AccountNumber,
		Name:          acc.Name,
		Type:          string(acc.Type),
		Balance:       acc.Balance,
		Currency:      acc.Currency,
		Timestamp:     time.Now(),
	})
	
	return acc, nil
}

func (s *GeneralLedgerService) DeleteAccount(ctx context.Context, id string) error {
	return s.accounts.Delete(ctx, id)
}

func (s *GeneralLedgerService) GetAccountBalance(ctx context.Context, id string) (decimal.Decimal, error) {
	acc, err := s.accounts.GetByID(ctx, id)
	if err != nil {
		return decimal.Zero, err
	}
	return acc.Balance, nil
}

func (s *GeneralLedgerService) ListJournalEntries(ctx context.Context) ([]domain.JournalEntry, error) {
	return s.entries.List(ctx)
}

func (s *GeneralLedgerService) CreateJournalEntry(ctx context.Context, ref, desc string, lines []domain.JournalEntryLine) (*domain.JournalEntry, error) {
	if len(lines) < 2 {
		return nil, errors.New("a journal entry must have at least 2 lines")
	}

	totalDebits := decimal.Zero
	totalCredits := decimal.Zero
	for _, l := range lines {
		totalDebits = totalDebits.Add(l.DebitAmount)
		totalCredits = totalCredits.Add(l.CreditAmount)
	}

	if !totalDebits.Equal(totalCredits) {
		return nil, fmt.Errorf("journal entry is unbalanced: debits=%s, credits=%s", totalDebits, totalCredits)
	}

	id := fmt.Sprintf("je_%d", time.Now().UnixNano())
	entry := &domain.JournalEntry{
		ID:          id,
		Reference:   ref,
		Date:        time.Now(),
		Description: desc,
		Status:      "POSTED",
		CreatedBy:   "system",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	for i, l := range lines {
		lines[i].ID = fmt.Sprintf("jel_%d_%d", time.Now().UnixNano(), i)
		lines[i].EntryID = id

		acc, err := s.accounts.GetByID(ctx, l.AccountID)
		if err != nil {
			return nil, fmt.Errorf("account not found: %s", l.AccountID)
		}

		isDebitIncreaseType := acc.Type == "ASSET" || acc.Type == "EXPENSE"
		if isDebitIncreaseType {
			acc.Balance = acc.Balance.Add(l.DebitAmount).Sub(l.CreditAmount)
		} else {
			acc.Balance = acc.Balance.Sub(l.DebitAmount).Add(l.CreditAmount)
		}

		err = s.accounts.Update(ctx, acc)
		if err != nil {
			return nil, fmt.Errorf("failed to update balance for account %s: %w", acc.ID, err)
		}

		// Publish event
		_ = s.publisher.Publish(ctx, "fin.account.balance.changed", acc.ID, domain.AccountEventPayload{
			ID:            acc.ID,
			AccountNumber: acc.AccountNumber,
			Name:          acc.Name,
			Type:          string(acc.Type),
			Balance:       acc.Balance,
			Currency:      acc.Currency,
			Timestamp:     time.Now(),
		})
	}

	err := s.entries.Create(ctx, entry, lines)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

func (s *GeneralLedgerService) GetJournalEntry(ctx context.Context, id string) (*domain.JournalEntry, []domain.JournalEntryLine, error) {
	return s.entries.GetByID(ctx, id)
}

func (s *GeneralLedgerService) ReverseJournalEntry(ctx context.Context, id string) (*domain.JournalEntry, error) {
	entry, lines, err := s.entries.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if entry.Status == "REVERSED" {
		return nil, errors.New("journal entry is already reversed")
	}

	// Create reverse lines
	revLines := make([]domain.JournalEntryLine, len(lines))
	for i, l := range lines {
		revLines[i] = domain.JournalEntryLine{
			AccountID:     l.AccountID,
			DebitAmount:   l.CreditAmount, // swap debits and credits
			CreditAmount:  l.DebitAmount,
			Description:   "Reversal of " + entry.Reference + ": " + l.Description,
			CostCenterID:  l.CostCenterID,
		}
	}

	revRef := fmt.Sprintf("REV-%s", entry.Reference)
	revDesc := fmt.Sprintf("Reversal of Journal Entry %s: %s", entry.Reference, entry.Description)
	
	// Create reversing journal entry (which will handle adjusting the GL account balances)
	revEntry, err := s.CreateJournalEntry(ctx, revRef, revDesc, revLines)
	if err != nil {
		return nil, fmt.Errorf("failed to create reversing entry: %w", err)
	}

	// Update original entry
	entry.Status = "REVERSED"
	entry.ReversedBy = &revEntry.ID
	entry.UpdatedAt = time.Now()

	err = s.entries.Update(ctx, entry, lines)
	if err != nil {
		return nil, fmt.Errorf("failed to update original entry status: %w", err)
	}

	return revEntry, nil
}


func (s *GeneralLedgerService) GetTrialBalance(ctx context.Context) (map[string]interface{}, error) {
	accs, err := s.accounts.List(ctx)
	if err != nil {
		return nil, err
	}

	var totalDebits, totalCredits decimal.Decimal
	balances := make(map[string]map[string]decimal.Decimal)

	for _, a := range accs {
		if !a.IsActive {
			continue
		}
		isDebitType := a.Type == "ASSET" || a.Type == "EXPENSE"
		if isDebitType {
			totalDebits = totalDebits.Add(a.Balance)
			if _, ok := balances[a.Name]; !ok {
				balances[a.Name] = make(map[string]decimal.Decimal)
			}
			balances[a.Name]["debit"] = a.Balance
		} else {
			totalCredits = totalCredits.Add(a.Balance)
			if _, ok := balances[a.Name]; !ok {
				balances[a.Name] = make(map[string]decimal.Decimal)
			}
			balances[a.Name]["credit"] = a.Balance
		}
	}

	return map[string]interface{}{
		"balances":      balances,
		"total_debits":  totalDebits,
		"total_credits": totalCredits,
	}, nil
}

func (s *GeneralLedgerService) GetBalanceSheet(ctx context.Context) (map[string]interface{}, error) {
	accs, err := s.accounts.List(ctx)
	if err != nil {
		return nil, err
	}

	assets := make(map[string]decimal.Decimal)
	liabilities := make(map[string]decimal.Decimal)
	equity := make(map[string]decimal.Decimal)
	
	var totalAssets, totalLiabilities, totalEquity decimal.Decimal

	for _, a := range accs {
		if !a.IsActive {
			continue
		}
		switch a.Type {
		case "ASSET":
			assets[a.Name] = a.Balance
			totalAssets = totalAssets.Add(a.Balance)
		case "LIABILITY":
			liabilities[a.Name] = a.Balance
			totalLiabilities = totalLiabilities.Add(a.Balance)
		case "EQUITY":
			equity[a.Name] = a.Balance
			totalEquity = totalEquity.Add(a.Balance)
		}
	}

	return map[string]interface{}{
		"assets":            assets,
		"total_assets":      totalAssets,
		"liabilities":       liabilities,
		"total_liabilities": totalLiabilities,
		"equity":            equity,
		"total_equity":      totalEquity,
	}, nil
}

func (s *GeneralLedgerService) UpdateJournalEntry(ctx context.Context, id string, ref, desc string, lines []domain.JournalEntryLine) (*domain.JournalEntry, error) {
	entry, _, err := s.entries.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	entry.Reference = ref
	entry.Description = desc
	entry.UpdatedAt = time.Now()

	err = s.entries.Update(ctx, entry, lines)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

func (s *GeneralLedgerService) DeleteJournalEntry(ctx context.Context, id string) error {
	return s.entries.Delete(ctx, id)
}

