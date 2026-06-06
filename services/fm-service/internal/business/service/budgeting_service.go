package service

import (
	"log"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/erp-system/fm-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type BudgetingService struct {
	budgets   domain.BudgetRepository
	accounts  domain.AccountRepository
	publisher domain.EventPublisher
}

func NewBudgetingService(budgets domain.BudgetRepository, accounts domain.AccountRepository, publisher domain.EventPublisher) *BudgetingService {
	return &BudgetingService{
		budgets:   budgets,
		accounts:  accounts,
		publisher: publisher,
	}
}

func (s *BudgetingService) ListBudgets(ctx context.Context) ([]domain.Budget, error) {
	return s.budgets.List(ctx)
}

func (s *BudgetingService) CreateBudget(ctx context.Context, accountID, costCenterID string, fiscalYear, period int, allocatedAmount decimal.Decimal) (*domain.Budget, error) {
	if accountID == "" || fiscalYear <= 0 || period < 1 || period > 12 {
		return nil, errors.New("invalid budget inputs: account ID, year, and valid month period are required")
	}

	id := fmt.Sprintf("bud_%d", time.Now().UnixNano())
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

	err := s.budgets.Create(ctx, budget)
	if err != nil {
		return nil, err
	}

	// Publish event
	if err := s.publisher.Publish(ctx, domain.TopicFinBudgetCreated, budget.ID, domain.BudgetEventPayload{
		AccountID:       budget.AccountID,
		CostCenterID:    budget.CostCenterID,
		FiscalYear:      budget.FiscalYear,
		Period:          budget.Period,
		AllocatedAmount: budget.AllocatedAmount,
		SpentAmount:     budget.SpentAmount,
		Timestamp:       time.Now(),
	}); err != nil {
		log.Printf("ERROR: failed to publish event %s: %v", domain.TopicFinBudgetCreated, err)
	}

	return budget, nil
}

func (s *BudgetingService) GetBudgetVsActualReport(ctx context.Context, accountID string, fiscalYear int) (map[string]interface{}, error) {
	acc, err := s.accounts.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	// Sum all budgets for this account and year
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

	actualSpent := acc.Balance // Simplified in memory representation
	variance := totalBudget.Sub(actualSpent)

	return map[string]interface{}{
		"account_number": acc.AccountNumber,
		"account_name":   acc.Name,
		"fiscal_year":    fiscalYear,
		"budget_amount":  totalBudget,
		"actual_spent":   actualSpent,
		"variance":       variance,
	}, nil
}

func (s *BudgetingService) CheckAndTrackBudgetExpense(ctx context.Context, accountID string, amount decimal.Decimal, year, period int) error {
	bud, err := s.budgets.GetByAccountAndPeriod(ctx, accountID, year, period)
	if err != nil {
		// No budget set for this account/period, ignore
		return nil
	}

	newSpent := bud.SpentAmount.Add(amount)
	bud.SpentAmount = newSpent
	bud.UpdatedAt = time.Now()

	_ = s.budgets.Update(ctx, bud)

	// Publish budget updated event
	if err := s.publisher.Publish(ctx, domain.TopicFinBudgetUpdated, bud.ID, domain.BudgetEventPayload{
		AccountID:       bud.AccountID,
		CostCenterID:    bud.CostCenterID,
		FiscalYear:      bud.FiscalYear,
		Period:          bud.Period,
		AllocatedAmount: bud.AllocatedAmount,
		SpentAmount:     bud.SpentAmount,
		Timestamp:       time.Now(),
	}); err != nil {
		log.Printf("ERROR: failed to publish event %s: %v", domain.TopicFinBudgetUpdated, err)
	}

	if newSpent.GreaterThan(bud.AllocatedAmount) {
		// Publish budget exceeded event
		if err := s.publisher.Publish(ctx, domain.TopicFinBudgetExceeded, bud.ID, domain.BudgetEventPayload{
			AccountID:       bud.AccountID,
			CostCenterID:    bud.CostCenterID,
			FiscalYear:      bud.FiscalYear,
			Period:          bud.Period,
			AllocatedAmount: bud.AllocatedAmount,
			SpentAmount:     bud.SpentAmount,
			Timestamp:       time.Now(),
		}); err != nil {
			log.Printf("ERROR: failed to publish event %s: %v", domain.TopicFinBudgetExceeded, err)
		}
	}

	return nil
}
