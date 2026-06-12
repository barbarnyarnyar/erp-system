package service

import (
	"context"
	"erp-system/shared/utils"
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
	id := utils.NewID("alloc")
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
	if err := s.publisher.Publish(ctx, domain.TopicPrjResourceAllocated, id, domain.ResourceAllocatedEvent{
		AllocationID: id,
		ProjectID:    projectID,
		UserID:       userID,
		Role:         role,
		Timestamp:    time.Now(),
	}); err != nil {
		utils.LogPublishErr("pm-service", domain.TopicPrjResourceAllocated, err)
	}

	return alloc, nil
}

func (s *ResourceManagementService) ListAllocations(ctx context.Context, projectID string) ([]domain.ResourceAllocation, error) {
	return s.allocRepo.ListByProject(ctx, projectID)
}

func (s *ResourceManagementService) ReleaseResource(ctx context.Context, allocationID string) error {
	alloc, err := s.allocRepo.GetByID(ctx, allocationID)
	if err != nil {
		return err
	}

	err = s.allocRepo.Delete(ctx, allocationID)
	if err != nil {
		return err
	}

	if err := s.publisher.Publish(ctx, domain.TopicPrjResourceReleased, allocationID, domain.ResourceReleasedEvent{
		AllocationID: allocationID,
		ProjectID:    alloc.ProjectID,
		UserID:       alloc.UserID,
		Timestamp:    time.Now(),
	}); err != nil {
		utils.LogPublishErr("pm-service", domain.TopicPrjResourceReleased, err)
	}

	return nil
}

func (s *ResourceManagementService) CheckResourceOverallocation(ctx context.Context, userID, projectID string, totalCapacity int) error {
	if err := s.publisher.Publish(ctx, domain.TopicPrjResourceOverallocated, userID, domain.ResourceOverallocatedEvent{
		UserID:        userID,
		ProjectID:     projectID,
		TotalCapacity: totalCapacity,
		Timestamp:     time.Now(),
	}); err != nil {
		utils.LogPublishErr("pm-service", domain.TopicPrjResourceOverallocated, err)
		return err
	}
	return nil
}
