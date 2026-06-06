package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/erp-system/crm-service/internal/business/domain"
	"github.com/erp-system/crm-service/internal/business/service"
	"github.com/erp-system/crm-service/internal/data/memory"
)

func TestCustomerInteraction_Create_AndList(t *testing.T) {
	repo := memory.NewCustomerInteractionRepository()
	pub := &MockPublisher{}
	svc := service.NewCustomerInteractionService(repo, pub)

	ctx := context.Background()
	ci, err := svc.CreateCustomerInteraction(ctx, "cust_1", "CALL", "Follow-up", "Discussed renewal", time.Now(), "alice")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if ci.Type != "CALL" {
		t.Errorf("Type = %q, want CALL", ci.Type)
	}
	if ci.Subject != "Follow-up" {
		t.Errorf("Subject = %q, want Follow-up", ci.Subject)
	}
	if ci.CreatedBy != "alice" {
		t.Errorf("CreatedBy = %q, want alice", ci.CreatedBy)
	}

	list, _ := svc.ListCustomerInteractions(ctx, "cust_1")
	if len(list) != 1 {
		t.Fatalf("expected 1 interaction, got %d", len(list))
	}

	found := false
	for _, ev := range pub.Events {
		if ev.Topic == domain.TopicCrmCustomerInteractionLogged {
			found = true
		}
	}
	if !found {
		t.Errorf("expected %s event", domain.TopicCrmCustomerInteractionLogged)
	}
}

func TestCustomerInteraction_RequiresCustomerIDAndType(t *testing.T) {
	repo := memory.NewCustomerInteractionRepository()
	svc := service.NewCustomerInteractionService(repo, &MockPublisher{})

	_, err := svc.CreateCustomerInteraction(context.Background(), "", "CALL", "x", "y", time.Now(), "alice")
	if err == nil {
		t.Errorf("expected error for empty customer_id")
	}
	_, err = svc.CreateCustomerInteraction(context.Background(), "cust_1", "", "x", "y", time.Now(), "alice")
	if err == nil {
		t.Errorf("expected error for empty type")
	}
}

func TestCustomerInteraction_Delete(t *testing.T) {
	repo := memory.NewCustomerInteractionRepository()
	svc := service.NewCustomerInteractionService(repo, &MockPublisher{})

	ctx := context.Background()
	ci, _ := svc.CreateCustomerInteraction(ctx, "cust_1", "EMAIL", "Quote sent", "Attached proposal", time.Now(), "bob")
	if err := svc.DeleteCustomerInteraction(ctx, ci.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := svc.GetCustomerInteraction(ctx, ci.ID); err == nil {
		t.Errorf("expected error fetching deleted interaction")
	}
}
