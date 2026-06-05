package service

import (
	"context"
	"fmt"
	"time"

	"github.com/erp-system/hr-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type LeaveManagementService struct {
	repo      domain.LeaveRequestRepository
	balances  domain.LeaveBalanceRepository
	publisher domain.EventPublisher
}

func NewLeaveManagementService(repo domain.LeaveRequestRepository, balances domain.LeaveBalanceRepository, publisher domain.EventPublisher) *LeaveManagementService {
	return &LeaveManagementService{
		repo:      repo,
		balances:  balances,
		publisher: publisher,
	}
}

func (s *LeaveManagementService) ListLeaveRequests(ctx context.Context) ([]domain.LeaveRequest, error) {
	return s.repo.List(ctx)
}

func (s *LeaveManagementService) CreateLeaveRequest(ctx context.Context, employeeID string, leaveType string, start, end time.Time, reason string) (*domain.LeaveRequest, error) {
	// Calculate leave duration in days
	durationDays := decimal.NewFromFloat(end.Sub(start).Hours()/24.0 + 1)

	// Validate leave balance
	year := start.Year()
	balance, err := s.balances.GetByEmployeeAndTypeAndYear(ctx, employeeID, leaveType, year)
	if err != nil {
		// Auto-initialize a default leave balance for testing/convenience
		balance = &domain.LeaveBalance{
			ID:           fmt.Sprintf("bal_%d", time.Now().UnixNano()),
			EmployeeID:   employeeID,
			LeaveType:    leaveType,
			EntitledDays: decimal.NewFromInt(15), // 15 entitled days by default
			UsedDays:     decimal.Zero,
			Year:         year,
		}
		_ = s.balances.Create(ctx, balance)
	}

	remaining := balance.EntitledDays.Sub(balance.UsedDays)
	if remaining.LessThan(durationDays) {
		return nil, fmt.Errorf("insufficient leave balance: requested %s days, only %s days remaining", durationDays.String(), remaining.String())
	}

	id := fmt.Sprintf("leave_%d", time.Now().UnixNano())

	lr := &domain.LeaveRequest{
		ID:          id,
		EmployeeID:  employeeID,
		LeaveType:   leaveType,
		StartDate:   start,
		EndDate:     end,
		Reason:      reason,
		Status:      "PENDING",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = s.repo.Create(ctx, lr)
	if err != nil {
		return nil, err
	}

	// Publish leave requested event
	_ = s.publisher.Publish(ctx, domain.TopicHrLeaveRequested, lr.ID, domain.LeaveRequestedEvent{
		LeaveRequestID: lr.ID,
		EmployeeID:     lr.EmployeeID,
		LeaveType:      lr.LeaveType,
		StartDate:      lr.StartDate,
		EndDate:        lr.EndDate,
		Timestamp:      time.Now(),
	})

	return lr, nil
}

func (s *LeaveManagementService) GetLeaveRequest(ctx context.Context, id string) (*domain.LeaveRequest, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *LeaveManagementService) UpdateLeaveRequest(ctx context.Context, id string, leaveType string, start, end time.Time, reason string) (*domain.LeaveRequest, error) {
	lr, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	lr.LeaveType = leaveType
	lr.StartDate = start
	lr.EndDate = end
	lr.Reason = reason
	lr.UpdatedAt = time.Now()

	err = s.repo.Update(ctx, lr)
	if err != nil {
		return nil, err
	}

	return lr, nil
}

func (s *LeaveManagementService) UpdateLeaveStatus(ctx context.Context, id string, approvedBy string, status string) (*domain.LeaveRequest, error) {
	lr, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	oldStatus := lr.Status
	lr.Status = status
	lr.ApprovedBy = &approvedBy
	lr.UpdatedAt = time.Now()

	err = s.repo.Update(ctx, lr)
	if err != nil {
		return nil, err
	}

	// If approved, update LeaveBalance
	if status == "APPROVED" && oldStatus != "APPROVED" {
		durationDays := decimal.NewFromFloat(lr.EndDate.Sub(lr.StartDate).Hours()/24.0 + 1)
		balance, err := s.balances.GetByEmployeeAndTypeAndYear(ctx, lr.EmployeeID, lr.LeaveType, lr.StartDate.Year())
		if err == nil {
			balance.UsedDays = balance.UsedDays.Add(durationDays)
			_ = s.balances.Update(ctx, balance)
		}
	}

	if status == "APPROVED" {
		// Publish leave approved event
		_ = s.publisher.Publish(ctx, domain.TopicHrLeaveApproved, lr.ID, domain.LeaveApprovedEvent{
			LeaveRequestID: lr.ID,
			EmployeeID:     lr.EmployeeID,
			ApprovedBy:     approvedBy,
			Timestamp:      time.Now(),
		})
	} else if status == "REJECTED" {
		// Publish leave rejected event
		_ = s.publisher.Publish(ctx, domain.TopicHrLeaveRejected, lr.ID, domain.LeaveRejectedEvent{
			LeaveRequestID: lr.ID,
			EmployeeID:     lr.EmployeeID,
			RejectedBy:     approvedBy,
			Reason:         "Rejected by manager",
			Timestamp:      time.Now(),
		})
	}

	return lr, nil
}

func (s *LeaveManagementService) ApproveLeaveRequest(ctx context.Context, id string, approvedBy string) (*domain.LeaveRequest, error) {
	return s.UpdateLeaveStatus(ctx, id, approvedBy, "APPROVED")
}

func (s *LeaveManagementService) RejectLeaveRequest(ctx context.Context, id string, rejectedBy string) (*domain.LeaveRequest, error) {
	return s.UpdateLeaveStatus(ctx, id, rejectedBy, "REJECTED")
}


