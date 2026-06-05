package service

import (
	"context"
	"fmt"
	"time"

	"github.com/erp-system/hr-service/internal/business/domain"
)

type LeaveManagementService struct {
	repo      domain.LeaveRequestRepository
	publisher domain.EventPublisher
}

func NewLeaveManagementService(repo domain.LeaveRequestRepository, publisher domain.EventPublisher) *LeaveManagementService {
	return &LeaveManagementService{
		repo:      repo,
		publisher: publisher,
	}
}

func (s *LeaveManagementService) ListLeaveRequests(ctx context.Context) ([]domain.LeaveRequest, error) {
	return s.repo.List(ctx)
}

func (s *LeaveManagementService) CreateLeaveRequest(ctx context.Context, employeeID string, leaveType string, start, end time.Time, reason string) (*domain.LeaveRequest, error) {
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

	err := s.repo.Create(ctx, lr)
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

func (s *LeaveManagementService) ApproveLeaveRequest(ctx context.Context, id string, approvedBy string) (*domain.LeaveRequest, error) {
	lr, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	lr.Status = "APPROVED"
	lr.ApprovedBy = &approvedBy
	lr.UpdatedAt = time.Now()

	err = s.repo.Update(ctx, lr)
	if err != nil {
		return nil, err
	}

	// Publish leave approved event
	_ = s.publisher.Publish(ctx, domain.TopicHrLeaveApproved, lr.ID, domain.LeaveApprovedEvent{
		LeaveRequestID: lr.ID,
		EmployeeID:     lr.EmployeeID,
		ApprovedBy:     approvedBy,
		Timestamp:      time.Now(),
	})

	return lr, nil
}

func (s *LeaveManagementService) RejectLeaveRequest(ctx context.Context, id string, rejectedBy string) (*domain.LeaveRequest, error) {
	lr, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	lr.Status = "REJECTED"
	lr.ApprovedBy = &rejectedBy
	lr.UpdatedAt = time.Now()

	err = s.repo.Update(ctx, lr)
	if err != nil {
		return nil, err
	}

	// Publish leave rejected event
	_ = s.publisher.Publish(ctx, domain.TopicHrLeaveRejected, lr.ID, domain.LeaveRejectedEvent{
		LeaveRequestID: lr.ID,
		EmployeeID:     lr.EmployeeID,
		RejectedBy:     rejectedBy,
		Reason:         "Rejected by manager",
		Timestamp:      time.Now(),
	})

	return lr, nil
}

