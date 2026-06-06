package service

import (
	"context"
	"fmt"
	"time"

	"github.com/erp-system/pm-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type TimeExpenseService struct {
	timeRepo    domain.ProjectTimeEntryRepository
	expenseRepo domain.ProjectExpenseRepository
	publisher   domain.EventPublisher
}

func NewTimeExpenseService(
	timeRepo domain.ProjectTimeEntryRepository,
	expenseRepo domain.ProjectExpenseRepository,
	publisher domain.EventPublisher,
) *TimeExpenseService {
	return &TimeExpenseService{
		timeRepo:    timeRepo,
		expenseRepo: expenseRepo,
		publisher:   publisher,
	}
}

func (s *TimeExpenseService) LogTime(ctx context.Context, projectID, taskID, userID string, entryDate time.Time, hours decimal.Decimal, description string) (*domain.ProjectTimeEntry, error) {
	id := fmt.Sprintf("time_%d", time.Now().UnixNano())
	entry := &domain.ProjectTimeEntry{
		ID:          id,
		ProjectID:   projectID,
		TaskID:      taskID,
		UserID:      userID,
		EntryDate:   entryDate,
		Hours:       hours,
		Description: description,
		Status:      "SUBMITTED",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := s.timeRepo.Create(ctx, entry)
	if err != nil {
		return nil, err
	}

	_ = s.publisher.Publish(ctx, domain.TopicPrjTimeLogged, id, domain.TimeLoggedEvent{
		TimeLogID:    id,
		ProjectID:    projectID,
		EmployeeID:   userID,
		HoursLogged:  hours,
		BillableRate: decimal.NewFromFloat(75.00),
		Timestamp:    time.Now(),
	})

	return entry, nil
}

func (s *TimeExpenseService) ApproveTime(ctx context.Context, entryID string, approvedBy string) (*domain.ProjectTimeEntry, error) {
	entry, err := s.timeRepo.GetByID(ctx, entryID)
	if err != nil {
		return nil, err
	}

	entry.Status = "APPROVED"
	entry.ApprovedBy = &approvedBy
	entry.UpdatedAt = time.Now()

	err = s.timeRepo.Update(ctx, entry)
	if err != nil {
		return nil, err
	}

	// Publish Event
	_ = s.publisher.Publish(ctx, domain.TopicPrjTimeApproved, entryID, domain.TimeApprovedEvent{
		TimeLogID:  entryID,
		ApprovedBy: approvedBy,
		Timestamp:  time.Now(),
	})

	return entry, nil
}

func (s *TimeExpenseService) ListTimeEntries(ctx context.Context, projectID string) ([]domain.ProjectTimeEntry, error) {
	return s.timeRepo.ListByProject(ctx, projectID)
}

func (s *TimeExpenseService) LogExpense(ctx context.Context, projectID, taskID, userID string, amount decimal.Decimal, currency string, expenseDate time.Time, category, description string) (*domain.ProjectExpense, error) {
	id := fmt.Sprintf("exp_%d", time.Now().UnixNano())
	expense := &domain.ProjectExpense{
		ID:          id,
		ProjectID:   projectID,
		UserID:      userID,
		Amount:      amount,
		Currency:    currency,
		ExpenseDate: expenseDate,
		Category:    category,
		Description: description,
		Status:      "SUBMITTED",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if taskID != "" {
		expense.TaskID = &taskID
	}

	err := s.expenseRepo.Create(ctx, expense)
	if err != nil {
		return nil, err
	}

	// Publish new submit event
	_ = s.publisher.Publish(ctx, domain.TopicPrjExpenseSubmitted, id, domain.ExpenseSubmittedEvent{
		ExpenseID: id,
		ProjectID: projectID,
		UserID:    userID,
		Amount:    amount,
		Currency:  currency,
		Timestamp: time.Now(),
	})

	// Keep old incurrence event for FM service compatibility
	_ = s.publisher.Publish(ctx, domain.TopicPrjExpenseIncurred, id, domain.ProjectExpenseIncurredEvent{
		ExpenseID:   id,
		ProjectID:   projectID,
		Description: description,
		Amount:      amount,
		Timestamp:   time.Now(),
	})

	return expense, nil
}

func (s *TimeExpenseService) ApproveExpense(ctx context.Context, expenseID string, approvedBy string) (*domain.ProjectExpense, error) {
	exp, err := s.expenseRepo.GetByID(ctx, expenseID)
	if err != nil {
		return nil, err
	}

	exp.Status = "APPROVED"
	exp.ApprovedBy = &approvedBy
	exp.UpdatedAt = time.Now()

	err = s.expenseRepo.Update(ctx, exp)
	if err != nil {
		return nil, err
	}

	// Publish Event
	_ = s.publisher.Publish(ctx, domain.TopicPrjExpenseApproved, expenseID, domain.ExpenseApprovedEvent{
		ExpenseID:  expenseID,
		ApprovedBy: approvedBy,
		Timestamp:  time.Now(),
	})

	return exp, nil
}

func (s *TimeExpenseService) ListExpenses(ctx context.Context, projectID string) ([]domain.ProjectExpense, error) {
	return s.expenseRepo.ListByProject(ctx, projectID)
}
