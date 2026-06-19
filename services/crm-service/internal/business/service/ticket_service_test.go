package service_test

import (
	sharedtesting "erp-system/shared/testing"
	"context"
	"testing"

	"github.com/erp-system/crm-service/internal/business/domain"
	"github.com/erp-system/crm-service/internal/business/service"
	"github.com/erp-system/crm-service/internal/data/memory"
)

func TestServiceTicketService_All(t *testing.T) {
	repo := memory.NewServiceTicketRepository()
	pub := &sharedtesting.MockPublisher{}
	svc := service.NewServiceTicketService(repo, pub)

	ctx := context.Background()

	// 1. Create Ticket
	ticket, err := svc.CreateServiceTicket(ctx, "cust_1", "UI Bug", "Login page is slow", "HIGH")
	if err != nil {
		t.Fatalf("failed to create ticket: %v", err)
	}
	if ticket.Title != "UI Bug" {
		t.Errorf("expected title 'UI Bug', got %q", ticket.Title)
	}
	if ticket.Status != "OPEN" {
		t.Errorf("expected status 'OPEN', got %q", ticket.Status)
	}

	// Verify event
	foundCreated := false
	for _, ev := range pub.Events {
		if ev.Topic == domain.TopicCrmServiceTicketCreated {
			foundCreated = true
		}
	}
	if !foundCreated {
		t.Errorf("expected ticket created event to be published")
	}

	// 2. Get Ticket
	fetched, err := svc.GetServiceTicket(ctx, ticket.ID)
	if err != nil {
		t.Fatalf("failed to get ticket: %v", err)
	}
	if fetched.ID != ticket.ID {
		t.Errorf("expected ticket ID %q, got %q", ticket.ID, fetched.ID)
	}

	// 3. List Tickets
	list, err := svc.ListServiceTickets(ctx)
	if err != nil {
		t.Fatalf("failed to list tickets: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected list length 1, got %d", len(list))
	}

	// 4. Update Ticket -> ESCALATED
	pub.Events = nil
	updated, err := svc.UpdateServiceTicket(ctx, ticket.ID, "ESCALATED", "CRITICAL")
	if err != nil {
		t.Fatalf("failed to update ticket status to ESCALATED: %v", err)
	}
	if updated.Status != "ESCALATED" {
		t.Errorf("expected status ESCALATED, got %q", updated.Status)
	}
	if updated.Priority != "CRITICAL" {
		t.Errorf("expected priority CRITICAL, got %q", updated.Priority)
	}
	foundEscalated := false
	for _, ev := range pub.Events {
		if ev.Topic == domain.TopicCrmServiceTicketEscalated {
			foundEscalated = true
		}
	}
	if !foundEscalated {
		t.Errorf("expected ticket escalated event to be published")
	}

	// 5. Update Ticket -> RESOLVED
	pub.Events = nil
	updated, err = svc.UpdateServiceTicket(ctx, ticket.ID, "RESOLVED", "CRITICAL")
	if err != nil {
		t.Fatalf("failed to update ticket status to RESOLVED: %v", err)
	}
	if updated.Status != "RESOLVED" {
		t.Errorf("expected status RESOLVED, got %q", updated.Status)
	}
	foundResolved := false
	for _, ev := range pub.Events {
		if ev.Topic == domain.TopicCrmServiceTicketResolved {
			foundResolved = true
		}
	}
	if !foundResolved {
		t.Errorf("expected ticket resolved event to be published")
	}

	// 6. Delete Ticket
	err = svc.DeleteServiceTicket(ctx, ticket.ID)
	if err != nil {
		t.Fatalf("failed to delete ticket: %v", err)
	}

	// Verify deletion
	_, err = svc.GetServiceTicket(ctx, ticket.ID)
	if err == nil {
		t.Errorf("expected error when getting deleted ticket, got nil")
	}
}

func TestServiceTicketService_Errors(t *testing.T) {
	repo := memory.NewServiceTicketRepository()
	pub := &sharedtesting.MockPublisher{}
	svc := service.NewServiceTicketService(repo, pub)

	ctx := context.Background()

	_, err := svc.UpdateServiceTicket(ctx, "non-existent", "RESOLVED", "LOW")
	if err == nil {
		t.Errorf("expected error updating non-existent ticket, got nil")
	}
}
