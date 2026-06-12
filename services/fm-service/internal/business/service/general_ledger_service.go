package service

import (
	"context"
	"erp-system/shared/utils"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/erp-system/fm-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type GeneralLedgerService struct {
	accounts domain.ChartOfAccountsRepository
	entries  domain.UniversalJournalEntryRepository
	outbox   domain.TransactionalOutboxRepository
	tm       domain.TransactionManager
}

func NewGeneralLedgerService(
	accounts domain.ChartOfAccountsRepository,
	entries domain.UniversalJournalEntryRepository,
	outbox domain.TransactionalOutboxRepository,
	tm domain.TransactionManager,
) *GeneralLedgerService {
	return &GeneralLedgerService{
		accounts: accounts,
		entries:  entries,
		outbox:   outbox,
		tm:       tm,
	}
}

func (s *GeneralLedgerService) calculateAccountBalance(ctx context.Context, accountID string) (decimal.Decimal, error) {
	entries, err := s.entries.List(ctx)
	if err != nil {
		return decimal.Zero, err
	}

	balance := decimal.Zero
	for _, entry := range entries {
		if entry.Status != domain.LedgerStatePOSTED && entry.Status != domain.LedgerStateREVERSED {
			continue
		}
		_, lines, err := s.entries.GetByID(ctx, entry.ID)
		if err != nil {
			return decimal.Zero, err
		}
		for _, line := range lines {
			if line.AccountID == accountID {
				balance = balance.Add(line.AmountFunctional)
			}
		}
	}
	return balance, nil
}

func (s *GeneralLedgerService) ListAccounts(ctx context.Context) ([]domain.ChartOfAccounts, error) {
	return s.accounts.List(ctx)
}

func (s *GeneralLedgerService) CreateAccount(ctx context.Context, legalEntityID, accountCode, name, accType string) (*domain.ChartOfAccounts, error) {
	if legalEntityID == "" || accountCode == "" || name == "" || accType == "" {
		return nil, errors.New("legal entity ID, account code, name, and type are required")
	}

	typeEnum := domain.AccountType(accType)
	if !typeEnum.IsValid() {
		return nil, fmt.Errorf("invalid account type: %s", accType)
	}

	id := utils.NewID("acc")
	acc := &domain.ChartOfAccounts{
		ID:            id,
		LegalEntityID: legalEntityID,
		AccountCode:   accountCode,
		AccountName:   name,
		Type:          typeEnum,
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		err := s.accounts.Create(txCtx, acc)
		if err != nil {
			return err
		}

		// Write to outbox
		outboxRec := &domain.TransactionalOutbox{
			ID:          utils.NewID("outbox"),
			EventType:   string(domain.TopicFmAccountCreated),
			AggregateID: acc.ID,
			Payload: domain.AccountEventPayload{
				ID:            acc.ID,
				AccountNumber: acc.AccountCode,
				Name:          acc.AccountName,
				Type:          string(acc.Type),
				Balance:       decimal.Zero,
				Timestamp:     time.Now(),
			},
			Status:    domain.OutboxStatusPENDING,
			CreatedAt: time.Now(),
		}
		return s.outbox.Create(txCtx, outboxRec)
	})

	if err != nil {
		return nil, err
	}

	return acc, nil
}

func (s *GeneralLedgerService) GetAccount(ctx context.Context, id string) (*domain.ChartOfAccounts, error) {
	return s.accounts.GetByID(ctx, id)
}

func (s *GeneralLedgerService) GetAccountByCode(ctx context.Context, legalEntityID, accountCode string) (*domain.ChartOfAccounts, error) {
	return s.accounts.GetByCode(ctx, legalEntityID, accountCode)
}

func (s *GeneralLedgerService) UpdateAccount(ctx context.Context, id, name, accType string, isActive bool) (*domain.ChartOfAccounts, error) {
	typeEnum := domain.AccountType(accType)
	if !typeEnum.IsValid() {
		return nil, fmt.Errorf("invalid account type: %s", accType)
	}

	acc, err := s.accounts.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	acc.AccountName = name
	acc.Type = typeEnum
	acc.IsActive = isActive
	acc.UpdatedAt = time.Now()

	err = s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		err = s.accounts.Update(txCtx, acc)
		if err != nil {
			return err
		}

		// Write to outbox
		outboxRec := &domain.TransactionalOutbox{
			ID:          utils.NewID("outbox"),
			EventType:   string(domain.TopicFmAccountUpdated),
			AggregateID: acc.ID,
			Payload: domain.AccountEventPayload{
				ID:            acc.ID,
				AccountNumber: acc.AccountCode,
				Name:          acc.AccountName,
				Type:          string(acc.Type),
				Balance:       decimal.Zero,
				Timestamp:     time.Now(),
			},
			Status:    domain.OutboxStatusPENDING,
			CreatedAt: time.Now(),
		}
		return s.outbox.Create(txCtx, outboxRec)
	})

	if err != nil {
		return nil, err
	}

	return acc, nil
}

func (s *GeneralLedgerService) DeleteAccount(ctx context.Context, id string) error {
	return s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		return s.accounts.Delete(txCtx, id)
	})
}

func (s *GeneralLedgerService) GetAccountBalance(ctx context.Context, id string) (decimal.Decimal, error) {
	acc, err := s.accounts.GetByID(ctx, id)
	if err != nil {
		return decimal.Zero, err
	}

	balance, err := s.calculateAccountBalance(ctx, acc.ID)
	if err != nil {
		return decimal.Zero, err
	}

	if acc.Type == domain.AccountTypeLIABILITY || acc.Type == domain.AccountTypeEQUITY || acc.Type == domain.AccountTypeREVENUE {
		balance = balance.Neg()
	}
	return balance, nil
}

func (s *GeneralLedgerService) ListJournalEntries(ctx context.Context) ([]domain.UniversalJournalEntry, error) {
	return s.entries.List(ctx)
}

func (s *GeneralLedgerService) CreateJournalEntry(ctx context.Context, legalEntityID, sourceModule, sourceDocID string, postingDate time.Time, lines []domain.UniversalJournalLine) (*domain.UniversalJournalEntry, error) {
	if len(lines) < 2 {
		return nil, errors.New("a journal entry must have at least 2 lines")
	}

	sum := decimal.Zero
	for _, l := range lines {
		sum = sum.Add(l.AmountFunctional)
	}
	if !sum.Equal(decimal.Zero) {
		return nil, fmt.Errorf("journal entry is unbalanced: functional sum=%s", sum)
	}

	id := utils.NewID("je")
	entry := &domain.UniversalJournalEntry{
		ID:               id,
		LegalEntityID:    legalEntityID,
		SourceModule:     sourceModule,
		SourceDocumentID: sourceDocID,
		PostingDate:      postingDate,
		FinancialPeriod:  postingDate.Format("2006-01"),
		Status:           domain.LedgerStatePOSTED,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	err := s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		for i := range lines {
			lines[i].ID = utils.NewID("jel")
			lines[i].JournalEntryID = id
		}

		err := s.entries.Create(txCtx, entry, lines)
		if err != nil {
			return err
		}

		for _, l := range lines {
			acc, err := s.accounts.GetByID(txCtx, l.AccountID)
			if err != nil {
				return fmt.Errorf("account not found: %s", l.AccountID)
			}
			outboxRec := &domain.TransactionalOutbox{
				ID:          utils.NewID("outbox"),
				EventType:   string(domain.TopicFmAccountBalanceChanged),
				AggregateID: acc.ID,
				Payload: domain.AccountEventPayload{
					ID:            acc.ID,
					AccountNumber: acc.AccountCode,
					Name:          acc.AccountName,
					Type:          string(acc.Type),
					Timestamp:     time.Now(),
				},
				Status:    domain.OutboxStatusPENDING,
				CreatedAt: time.Now(),
			}
			if err := s.outbox.Create(txCtx, outboxRec); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return entry, nil
}

func (s *GeneralLedgerService) GetJournalEntry(ctx context.Context, id string) (*domain.UniversalJournalEntry, []domain.UniversalJournalLine, error) {
	return s.entries.GetByID(ctx, id)
}

func (s *GeneralLedgerService) ReverseJournalEntry(ctx context.Context, id string) (*domain.UniversalJournalEntry, error) {
	var revEntry *domain.UniversalJournalEntry
	err := s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		entry, lines, err := s.entries.GetByID(txCtx, id)
		if err != nil {
			return err
		}

		if entry.Status == domain.LedgerStateREVERSED {
			return errors.New("journal entry is already reversed")
		}

		revLines := make([]domain.UniversalJournalLine, len(lines))
		for i, l := range lines {
			revLines[i] = domain.UniversalJournalLine{
				AccountID:             l.AccountID,
				AmountFunctional:      l.AmountFunctional.Neg(),
				AmountTransactional:   l.AmountTransactional.Neg(),
				CurrencyTransactional: l.CurrencyTransactional,
				TrackingDimensions:    l.TrackingDimensions,
			}
		}

		revEntry, err = s.CreateJournalEntry(txCtx, entry.LegalEntityID, entry.SourceModule, entry.SourceDocumentID, time.Now(), revLines)
		if err != nil {
			return fmt.Errorf("failed to create reversing entry: %w", err)
		}

		entry.Status = domain.LedgerStateREVERSED
		err = s.entries.Update(txCtx, entry, lines)
		if err != nil {
			return fmt.Errorf("failed to update original entry status: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
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
		rawBal, err := s.calculateAccountBalance(ctx, a.ID)
		if err != nil {
			return nil, err
		}

		isDebitType := a.Type == domain.AccountTypeASSET || a.Type == domain.AccountTypeEXPENSE
		if isDebitType {
			totalDebits = totalDebits.Add(rawBal)
			if _, ok := balances[a.AccountName]; !ok {
				balances[a.AccountName] = make(map[string]decimal.Decimal)
			}
			balances[a.AccountName]["debit"] = rawBal
		} else {
			reportedBal := rawBal.Neg()
			totalCredits = totalCredits.Add(reportedBal)
			if _, ok := balances[a.AccountName]; !ok {
				balances[a.AccountName] = make(map[string]decimal.Decimal)
			}
			balances[a.AccountName]["credit"] = reportedBal
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
		rawBal, err := s.calculateAccountBalance(ctx, a.ID)
		if err != nil {
			return nil, err
		}

		switch a.Type {
		case domain.AccountTypeASSET:
			assets[a.AccountName] = rawBal
			totalAssets = totalAssets.Add(rawBal)
		case domain.AccountTypeLIABILITY:
			reportedBal := rawBal.Neg()
			liabilities[a.AccountName] = reportedBal
			totalLiabilities = totalLiabilities.Add(reportedBal)
		case domain.AccountTypeEQUITY:
			reportedBal := rawBal.Neg()
			equity[a.AccountName] = reportedBal
			totalEquity = totalEquity.Add(reportedBal)
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

func (s *GeneralLedgerService) UpdateJournalEntry(ctx context.Context, id string, legalEntityID, sourceModule, sourceDocID string, postingDate time.Time, lines []domain.UniversalJournalLine) (*domain.UniversalJournalEntry, error) {
	if len(lines) < 2 {
		return nil, errors.New("a journal entry must have at least 2 lines")
	}

	sum := decimal.Zero
	for _, l := range lines {
		sum = sum.Add(l.AmountFunctional)
	}
	if !sum.Equal(decimal.Zero) {
		return nil, fmt.Errorf("journal entry is unbalanced: functional sum=%s", sum)
	}

	var entry *domain.UniversalJournalEntry
	err := s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		var err error
		entry, _, err = s.entries.GetByID(txCtx, id)
		if err != nil {
			return err
		}

		if entry.Status != domain.LedgerStateDRAFT {
			return domain.ErrJournalEntryNotMutable
		}

		for i := range lines {
			lines[i].ID = utils.NewID("jel")
			lines[i].JournalEntryID = id
		}

		entry.LegalEntityID = legalEntityID
		entry.SourceModule = sourceModule
		entry.SourceDocumentID = sourceDocID
		entry.PostingDate = postingDate
		entry.FinancialPeriod = postingDate.Format("2006-01")
		entry.UpdatedAt = time.Now()

		err = s.entries.Update(txCtx, entry, lines)
		if err != nil {
			return err
		}

		for _, l := range lines {
			acc, err := s.accounts.GetByID(txCtx, l.AccountID)
			if err != nil {
				return fmt.Errorf("account not found: %s", l.AccountID)
			}
			outboxRec := &domain.TransactionalOutbox{
				ID:          utils.NewID("outbox"),
				EventType:   string(domain.TopicFmAccountBalanceChanged),
				AggregateID: acc.ID,
				Payload: domain.AccountEventPayload{
					ID:            acc.ID,
					AccountNumber: acc.AccountCode,
					Name:          acc.AccountName,
					Type:          string(acc.Type),
					Timestamp:     time.Now(),
				},
				Status:    domain.OutboxStatusPENDING,
				CreatedAt: time.Now(),
			}
			if err := s.outbox.Create(txCtx, outboxRec); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return entry, nil
}

func (s *GeneralLedgerService) DeleteJournalEntry(ctx context.Context, id string) error {
	return s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		return s.entries.Delete(txCtx, id)
	})
}

func (s *GeneralLedgerService) GetIncomeStatement(ctx context.Context) (map[string]interface{}, error) {
	accs, err := s.accounts.List(ctx)
	if err != nil {
		return nil, err
	}

	revenues := make(map[string]decimal.Decimal)
	expenses := make(map[string]decimal.Decimal)

	var totalRevenue, totalExpense decimal.Decimal

	for _, a := range accs {
		if !a.IsActive {
			continue
		}
		rawBal, err := s.calculateAccountBalance(ctx, a.ID)
		if err != nil {
			return nil, err
		}

		switch a.Type {
		case domain.AccountTypeREVENUE:
			reportedBal := rawBal.Neg()
			revenues[a.AccountName] = reportedBal
			totalRevenue = totalRevenue.Add(reportedBal)
		case domain.AccountTypeEXPENSE:
			expenses[a.AccountName] = rawBal
			totalExpense = totalExpense.Add(rawBal)
		}
	}

	netIncome := totalRevenue.Sub(totalExpense)

	return map[string]interface{}{
		"revenues":      revenues,
		"total_revenue": totalRevenue,
		"expenses":      expenses,
		"total_expense": totalExpense,
		"net_income":    netIncome,
	}, nil
}

func (s *GeneralLedgerService) GetCashFlow(ctx context.Context) (map[string]interface{}, error) {
	accs, err := s.accounts.List(ctx)
	if err != nil {
		return nil, err
	}

	inflows := make(map[string]decimal.Decimal)
	outflows := make(map[string]decimal.Decimal)

	var totalInflow, totalOutflow decimal.Decimal

	for _, a := range accs {
		if !a.IsActive {
			continue
		}
		nameLower := strings.ToLower(a.AccountName)
		if a.Type == domain.AccountTypeASSET && (strings.Contains(nameLower, "cash") || strings.Contains(nameLower, "bank")) {
			rawBal, err := s.calculateAccountBalance(ctx, a.ID)
			if err != nil {
				return nil, err
			}
			if rawBal.IsPositive() {
				inflows[a.AccountName] = rawBal
				totalInflow = totalInflow.Add(rawBal)
			} else if rawBal.IsNegative() {
				absBal := rawBal.Abs()
				outflows[a.AccountName] = absBal
				totalOutflow = totalOutflow.Add(absBal)
			}
		}
	}

	netCashFlow := totalInflow.Sub(totalOutflow)

	return map[string]interface{}{
		"operating_inflows":  inflows,
		"total_inflows":      totalInflow,
		"operating_outflows": outflows,
		"total_outflows":     totalOutflow,
		"net_cash_flow":      netCashFlow,
	}, nil
}
