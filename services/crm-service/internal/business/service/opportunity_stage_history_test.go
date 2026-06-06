package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/erp-system/crm-service/internal/business/service"
	"github.com/erp-system/crm-service/internal/data/memory"
	"github.com/shopspring/decimal"
)

func TestOpportunity_Update_RecordsStageHistory(t *testing.T) {
	oppRepo := memory.NewOpportunityRepository()
	historyRepo := memory.NewOpportunityStageHistoryRepository()
	pub := &MockPublisher{}
	svc := service.NewOpportunityService(oppRepo, historyRepo, pub)

	ctx := context.Background()
	opp, err := svc.CreateOpportunity(ctx, "cust_1", "Big Deal", decimal.NewFromInt(1000), "DISCOVERY")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	histories, _ := historyRepo.ListByOpportunityID(ctx, opp.ID)
	if len(histories) != 1 {
		t.Fatalf("expected 1 history entry after create, got %d", len(histories))
	}
	if histories[0].Stage != "DISCOVERY" {
		t.Errorf("seeded stage = %q, want DISCOVERY", histories[0].Stage)
	}
	if histories[0].ChangedBy != "system" {
		t.Errorf("seeded ChangedBy = %q, want system", histories[0].ChangedBy)
	}

	_, err = svc.UpdateOpportunity(ctx, opp.ID, "Big Deal v2", decimal.NewFromInt(1500), "QUALIFIED", "NEGOTIATION", decimal.NewFromFloat(0.5), "alice")
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	histories, _ = historyRepo.ListByOpportunityID(ctx, opp.ID)
	if len(histories) != 2 {
		t.Fatalf("expected 2 history entries after stage change, got %d", len(histories))
	}
	if histories[1].Stage != "NEGOTIATION" {
		t.Errorf("new stage = %q, want NEGOTIATION", histories[1].Stage)
	}
	if histories[1].ChangedBy != "alice" {
		t.Errorf("ChangedBy = %q, want alice", histories[1].ChangedBy)
	}
}

func TestOpportunity_Update_SameStage_NoNewHistory(t *testing.T) {
	oppRepo := memory.NewOpportunityRepository()
	historyRepo := memory.NewOpportunityStageHistoryRepository()
	pub := &MockPublisher{}
	svc := service.NewOpportunityService(oppRepo, historyRepo, pub)

	ctx := context.Background()
	opp, _ := svc.CreateOpportunity(ctx, "cust_1", "Big Deal", decimal.NewFromInt(1000), "DISCOVERY")

	_, err := svc.UpdateOpportunity(ctx, opp.ID, "Big Deal v2", decimal.NewFromInt(1500), "NEW", "DISCOVERY", decimal.NewFromFloat(0.2), "alice")
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	histories, _ := historyRepo.ListByOpportunityID(ctx, opp.ID)
	if len(histories) != 1 {
		t.Errorf("expected 1 history entry (no new one for same stage), got %d", len(histories))
	}
}

func TestOpportunityStageHistory_TimeOrder(t *testing.T) {
	oppRepo := memory.NewOpportunityRepository()
	historyRepo := memory.NewOpportunityStageHistoryRepository()
	pub := &MockPublisher{}
	svc := service.NewOpportunityService(oppRepo, historyRepo, pub)

	ctx := context.Background()
	opp, _ := svc.CreateOpportunity(ctx, "cust_1", "Big Deal", decimal.NewFromInt(1000), "DISCOVERY")

	time.Sleep(time.Millisecond)
	_, _ = svc.UpdateOpportunity(ctx, opp.ID, "v2", decimal.NewFromInt(1500), "QUALIFIED", "NEGOTIATION", decimal.NewFromFloat(0.5), "bob")
	time.Sleep(time.Millisecond)
	_, _ = svc.UpdateOpportunity(ctx, opp.ID, "v3", decimal.NewFromInt(2000), "PROPOSAL", "NEGOTIATION", decimal.NewFromFloat(0.7), "carol")

	histories, _ := historyRepo.ListByOpportunityID(ctx, opp.ID)
	if len(histories) != 2 {
		t.Fatalf("expected 2 history entries, got %d", len(histories))
	}
	if !histories[0].ChangedAt.Before(histories[1].ChangedAt) {
		t.Errorf("expected chronological order, got %v then %v", histories[0].ChangedAt, histories[1].ChangedAt)
	}
}
