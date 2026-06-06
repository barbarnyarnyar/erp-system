package service

import (
	"context"
	"fmt"
	"time"

	"github.com/erp-system/pm-service/internal/business/domain"
)

type ResourceManagementService struct {
	allocRepo domain.ResourceAllocationRepository
	publisher domain.EventPublisher
}

func NewResourceManagementService(allocRepo domain.ResourceAllocationRepository, publisher domain.EventPublisher) *ResourceManagementService {
	return &ResourceManagementService{
		allocRepo: allocRepo,
		publisher: publisher,
	}
}

func (s *ResourceManagementService) AllocateResource(ctx context.Context, projectID, userID, role string, allocationPct int, startDate time.Time, endDate *time.Time) (*domain.ResourceAllocation, error) {
	id := fmt.Sprintf("alloc_%d", time.Now().UnixNano())
	alloc := &domain.ResourceAllocation{
		ID:                   id,
		ProjectID:            projectID,
		UserID:               userID,
		Role:                 role,
		AllocationPercentage: allocationPct,
		StartDate:            startDate,
		EndDate:              endDate,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	err := s.allocRepo.Create(ctx, alloc)
	if err != nil {
		return nil, err
	}

	// Publish Event
	_ = s.publisher.Publish(ctx, domain.TopicPrjResourceAllocated, id, domain.ResourceAllocatedEvent{
		AllocationID: id,
		ProjectID:    projectID,
		UserID:       userID,
		Role:         role,
		Timestamp:    time.Now(),
	})

	return alloc, nil
}

func (s *ResourceManagementService) ListAllocations(ctx context.Context, projectID string) ([]domain.ResourceAllocation, error) {
	return s.allocRepo.ListByProject(ctx, projectID)
}
