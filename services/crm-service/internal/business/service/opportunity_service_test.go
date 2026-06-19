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

func TestOpportunityService_All(t *testing.T) {
	oppRepo := memory.NewOpportunityRepository()
	historyRepo := memory.NewOpportunityStageHistoryRepository()
	pub := &sharedtesting.MockPublisher{}
	svc := service.NewOpportunityService(oppRepo, historyRepo, pub)

	ctx := context.Background()

	// 1. Create Opportunity -> success
	opp, err := svc.CreateOpportunity(ctx, "cust_1", "Acme Deal", decimal.NewFromInt(5000), "DISCOVERY")
	if err != nil {
		t.Fatalf("failed to create opportunity: %v", err)
	}
	if opp.Title != "Acme Deal" {
		t.Errorf("expected title 'Acme Deal', got %s", opp.Title)
	}

	// Verify create event
	foundCreated := false
	for _, ev := range pub.Events {
		if ev.Topic == domain.TopicCrmOpportunityCreated {
			foundCreated = true
		}
	}
	if !foundCreated {
		t.Errorf("expected opportunity created event to be published")
	}

	// 2. Get Opportunity
	fetched, err := svc.GetOpportunity(ctx, opp.ID)
	if err != nil {
		t.Fatalf("failed to get opportunity: %v", err)
	}
	if fetched.ID != opp.ID {
		t.Errorf("expected opportunity ID %q, got %q", opp.ID, fetched.ID)
	}

	// 3. List Opportunities
	list, err := svc.ListOpportunities(ctx)
	if err != nil {
		t.Fatalf("failed to list opportunities: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected list length 1, got %d", len(list))
	}

	// 4. Update Opportunity Status -> WON
	pub.Events = nil
	updated, err := svc.UpdateOpportunity(ctx, opp.ID, "Acme Deal Updated", decimal.NewFromInt(6000), "WON", "CLOSED_WON", decimal.NewFromFloat(1.0), "alice")
	if err != nil {
		t.Fatalf("failed to update opportunity: %v", err)
	}
	if updated.Status != "WON" {
		t.Errorf("expected status WON, got %s", updated.Status)
	}
	foundWon := false
	for _, ev := range pub.Events {
		if ev.Topic == domain.TopicCrmOpportunityWon {
			foundWon = true
		}
	}
	if !foundWon {
		t.Errorf("expected opportunity won event to be published")
	}

	// 5. Update Opportunity Status -> LOST
	pub.Events = nil
	updated, err = svc.UpdateOpportunity(ctx, opp.ID, "Acme Deal Updated", decimal.NewFromInt(6000), "LOST", "CLOSED_LOST", decimal.NewFromFloat(0.0), "alice")
	if err != nil {
		t.Fatalf("failed to update opportunity: %v", err)
	}
	if updated.Status != "LOST" {
		t.Errorf("expected status LOST, got %s", updated.Status)
	}
	foundLost := false
	for _, ev := range pub.Events {
		if ev.Topic == domain.TopicCrmOpportunityLost {
			foundLost = true
		}
	}
	if !foundLost {
		t.Errorf("expected opportunity lost event to be published")
	}

	// 6. Delete Opportunity
	err = svc.DeleteOpportunity(ctx, opp.ID)
	if err != nil {
		t.Fatalf("failed to delete opportunity: %v", err)
	}
	_, err = svc.GetOpportunity(ctx, opp.ID)
	if err == nil {
		t.Errorf("expected error when getting deleted opportunity, got nil")
	}
}

func TestOpportunityService_ValidationErrors(t *testing.T) {
	oppRepo := memory.NewOpportunityRepository()
	historyRepo := memory.NewOpportunityStageHistoryRepository()
	pub := &sharedtesting.MockPublisher{}
	svc := service.NewOpportunityService(oppRepo, historyRepo, pub)

	ctx := context.Background()

	// 1. Create with invalid stage
	_, err := svc.CreateOpportunity(ctx, "cust_1", "Deal", decimal.NewFromInt(100), "INVALID_STAGE")
	if err == nil {
		t.Errorf("expected error creating opportunity with invalid stage, got nil")
	}

	// Seed one valid opp
	opp, _ := svc.CreateOpportunity(ctx, "cust_1", "Deal", decimal.NewFromInt(100), "DISCOVERY")

	// 2. Update with invalid stage
	_, err = svc.UpdateOpportunity(ctx, opp.ID, "Deal", decimal.NewFromInt(100), "NEW", "INVALID_STAGE", decimal.NewFromFloat(0.5), "rep")
	if err == nil {
		t.Errorf("expected error updating opportunity with invalid stage, got nil")
	}

	// 3. Update non-existent opp
	_, err = svc.UpdateOpportunity(ctx, "non-existent", "Deal", decimal.NewFromInt(100), "NEW", "DISCOVERY", decimal.NewFromFloat(0.5), "rep")
	if err == nil {
		t.Errorf("expected error updating non-existent opportunity, got nil")
	}
}
