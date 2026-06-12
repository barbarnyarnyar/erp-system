package service

import (
	"context"
	"erp-system/shared/utils"
	"errors"
	"time"

	"github.com/erp-system/fm-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type BudgetingService struct {
	budgets  domain.BudgetRepository
	accounts domain.ChartOfAccountsRepository
	entries  domain.UniversalJournalEntryRepository
	outbox   domain.TransactionalOutboxRepository
	tm       domain.TransactionManager
}

func NewBudgetingService(
	budgets domain.BudgetRepository,
	accounts domain.ChartOfAccountsRepository,
	entries domain.UniversalJournalEntryRepository,
	outbox domain.TransactionalOutboxRepository,
	tm domain.TransactionManager,
) *BudgetingService {
	return &BudgetingService{
		budgets:  budgets,
		accounts: accounts,
		entries:  entries,
		outbox:   outbox,
		tm:       tm,
	}
}

func (s *BudgetingService) calculateAccountBalance(ctx context.Context, accountID string) (decimal.Decimal, error) {
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

func (s *BudgetingService) ListBudgets(ctx context.Context) ([]domain.Budget, error) {
	return s.budgets.List(ctx)
}

func (s *BudgetingService) CreateBudget(ctx context.Context, accountID, costCenterID string, fiscalYear, period int, allocatedAmount decimal.Decimal) (*domain.Budget, error) {
	if accountID == "" || fiscalYear <= 0 || period < 1 || period > 12 {
		return nil, errors.New("invalid budget inputs: account ID, year, and valid month period are required")
	}

	id := utils.NewID("bud")
	budget := &domain.Budget{
		ID:              id,
		AccountID:       accountID,
		FiscalYear:      fiscalYear,
		Period:          period,
		AllocatedAmount: allocatedAmount,
		SpentAmount:     decimal.Zero,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if costCenterID != "" {
		budget.CostCenterID = &costCenterID
	}

	err := s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		err := s.budgets.Create(txCtx, budget)
		if err != nil {
			return err
		}

		// Write to outbox
		outboxRec := &domain.TransactionalOutbox{
			ID:          utils.NewID("outbox"),
			EventType:   string(domain.TopicFmBudgetCreated),
			AggregateID: budget.ID,
			Payload: domain.BudgetEventPayload{
				AccountID:       budget.AccountID,
				CostCenterID:    budget.CostCenterID,
				FiscalYear:      budget.FiscalYear,
				Period:          budget.Period,
				AllocatedAmount: budget.AllocatedAmount,
				SpentAmount:     budget.SpentAmount,
				Timestamp:       time.Now(),
			},
			Status:    domain.OutboxStatusPENDING,
			CreatedAt: time.Now(),
		}
		return s.outbox.Create(txCtx, outboxRec)
	})

	if err != nil {
		return nil, err
	}

	return budget, nil
}

func (s *BudgetingService) GetBudgetVsActualReport(ctx context.Context, accountID string, fiscalYear int) (map[string]interface{}, error) {
	acc, err := s.accounts.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	buds, err := s.budgets.List(ctx)
	if err != nil {
		return nil, err
	}

	var totalBudget decimal.Decimal
	for _, b := range buds {
		if b.AccountID == accountID && b.FiscalYear == fiscalYear {
			totalBudget = totalBudget.Add(b.AllocatedAmount)
		}
	}

	actualSpent, err := s.calculateAccountBalance(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if acc.Type == domain.AccountTypeLIABILITY || acc.Type == domain.AccountTypeEQUITY || acc.Type == domain.AccountTypeREVENUE {
		actualSpent = actualSpent.Neg()
	}
	variance := totalBudget.Sub(actualSpent)

	return map[string]interface{}{
		"account_number": acc.AccountCode,
		"account_name":   acc.AccountName,
		"fiscal_year":    fiscalYear,
		"budget_amount":  totalBudget,
		"actual_spent":   actualSpent,
		"variance":       variance,
	}, nil
}

func (s *BudgetingService) CheckAndTrackBudgetExpense(ctx context.Context, accountID string, amount decimal.Decimal, year, period int) error {
	return s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		bud, err := s.budgets.GetByAccountAndPeriod(txCtx, accountID, year, period)
		if err != nil {
			return nil
		}

		newSpent := bud.SpentAmount.Add(amount)
		bud.SpentAmount = newSpent
		bud.UpdatedAt = time.Now()

		err = s.budgets.Update(txCtx, bud)
		if err != nil {
			return err
		}

		// Write budget updated to outbox
		outboxRec := &domain.TransactionalOutbox{
			ID:          utils.NewID("outbox"),
			EventType:   string(domain.TopicFmBudgetUpdated),
			AggregateID: bud.ID,
			Payload: domain.BudgetEventPayload{
				AccountID:       bud.AccountID,
				CostCenterID:    bud.CostCenterID,
				FiscalYear:      bud.FiscalYear,
				Period:          bud.Period,
				AllocatedAmount: bud.AllocatedAmount,
				SpentAmount:     bud.SpentAmount,
				Timestamp:       time.Now(),
			},
			Status:    domain.OutboxStatusPENDING,
			CreatedAt: time.Now(),
		}
		if err := s.outbox.Create(txCtx, outboxRec); err != nil {
			return err
		}

		if newSpent.GreaterThan(bud.AllocatedAmount) {
			// Write budget exceeded to outbox
			outboxRecExceeded := &domain.TransactionalOutbox{
				ID:          utils.NewID("outbox"),
				EventType:   string(domain.TopicFmBudgetExceeded),
				AggregateID: bud.ID,
				Payload: domain.BudgetEventPayload{
					AccountID:       bud.AccountID,
					CostCenterID:    bud.CostCenterID,
					FiscalYear:      bud.FiscalYear,
					Period:          bud.Period,
					AllocatedAmount: bud.AllocatedAmount,
					SpentAmount:     bud.SpentAmount,
					Timestamp:       time.Now(),
				},
				Status:    domain.OutboxStatusPENDING,
				CreatedAt: time.Now(),
			}
			if err := s.outbox.Create(txCtx, outboxRecExceeded); err != nil {
				return err
			}
		}

		return nil
	})
}
