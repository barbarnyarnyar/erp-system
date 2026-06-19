package service_test

import (
	sharedtesting "erp-system/shared/testing"
	"context"
	"testing"

	"github.com/erp-system/crm-service/internal/business/domain"
	"github.com/erp-system/crm-service/internal/business/service"
	"github.com/erp-system/crm-service/internal/data/memory"
	"github.com/shopspring/decimal"
)

func TestCampaignService_All(t *testing.T) {
	repo := memory.NewCampaignRepository()
	pub := &sharedtesting.MockPublisher{}
	svc := service.NewCampaignService(repo, pub)

	ctx := context.Background()

	// 1. Create campaign
	budget := decimal.NewFromInt(5000)
	camp, err := svc.CreateCampaign(ctx, "Summer Sale", "EMAIL", budget)
	if err != nil {
		t.Fatalf("failed to create campaign: %v", err)
	}
	if camp.Name != "Summer Sale" {
		t.Errorf("expected name 'Summer Sale', got %q", camp.Name)
	}
	if camp.Status != "DRAFT" {
		t.Errorf("expected status 'DRAFT', got %q", camp.Status)
	}

	// 2. Get campaign
	fetched, err := svc.GetCampaign(ctx, camp.ID)
	if err != nil {
		t.Fatalf("failed to get campaign: %v", err)
	}
	if fetched.ID != camp.ID {
		t.Errorf("expected campaign ID %q, got %q", camp.ID, fetched.ID)
	}

	// 3. List campaigns
	list, err := svc.ListCampaigns(ctx)
	if err != nil {
		t.Fatalf("failed to list campaigns: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected list length 1, got %d", len(list))
	}

	// 4. Update campaign status -> LAUNCHED (event should fire)
	newBudget := decimal.NewFromInt(6000)
	updated, err := svc.UpdateCampaign(ctx, camp.ID, "LAUNCHED", newBudget)
	if err != nil {
		t.Fatalf("failed to update campaign status to LAUNCHED: %v", err)
	}
	if updated.Status != "LAUNCHED" {
		t.Errorf("expected status 'LAUNCHED', got %q", updated.Status)
	}
	if !updated.Budget.Equal(newBudget) {
		t.Errorf("expected budget %s, got %s", newBudget, updated.Budget)
	}

	// Check launched event
	foundLaunched := false
	for _, ev := range pub.Events {
		if ev.Topic == domain.TopicCrmCampaignLaunched {
			foundLaunched = true
		}
	}
	if !foundLaunched {
		t.Errorf("expected campaign launched event to be published")
	}

	// 5. Update campaign status -> COMPLETED (event should fire)
	pub.Events = nil // clear events
	updated, err = svc.UpdateCampaign(ctx, camp.ID, "COMPLETED", newBudget)
	if err != nil {
		t.Fatalf("failed to update campaign status to COMPLETED: %v", err)
	}
	if updated.Status != "COMPLETED" {
		t.Errorf("expected status 'COMPLETED', got %q", updated.Status)
	}

	// Check completed event
	foundCompleted := false
	for _, ev := range pub.Events {
		if ev.Topic == domain.TopicCrmCampaignCompleted {
			foundCompleted = true
		}
	}
	if !foundCompleted {
		t.Errorf("expected campaign completed event to be published")
	}

	// 6. Delete campaign
	err = svc.DeleteCampaign(ctx, camp.ID)
	if err != nil {
		t.Fatalf("failed to delete campaign: %v", err)
	}

	// Verify deletion
	_, err = svc.GetCampaign(ctx, camp.ID)
	if err == nil {
		t.Errorf("expected error when getting deleted campaign, got nil")
	}
}

func TestCampaignService_UpdateNotFound(t *testing.T) {
	repo := memory.NewCampaignRepository()
	pub := &sharedtesting.MockPublisher{}
	svc := service.NewCampaignService(repo, pub)

	ctx := context.Background()
	_, err := svc.UpdateCampaign(ctx, "non-existent", "LAUNCHED", decimal.NewFromInt(1000))
	if err == nil {
		t.Errorf("expected error updating non-existent campaign, got nil")
	}
}
