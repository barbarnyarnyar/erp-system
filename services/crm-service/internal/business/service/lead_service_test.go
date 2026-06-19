package service_test

import (
	sharedtesting "erp-system/shared/testing"
	"context"
	"errors"
	"testing"

	"github.com/erp-system/crm-service/internal/business/domain"
	"github.com/erp-system/crm-service/internal/business/service"
	"github.com/erp-system/crm-service/internal/data/memory"
)

type MockFailCustomerRepo struct {
	domain.CustomerRepository
}

func (m *MockFailCustomerRepo) Create(ctx context.Context, customer *domain.CustomerProfile) error {
	return errors.New("mock db failure during customer creation")
}

func TestLeadService_All(t *testing.T) {
	leadRepo := memory.NewLeadRepository()
	custRepo := memory.NewCustomerRepository()
	oppRepo := memory.NewOpportunityRepository()
	pub := &sharedtesting.MockPublisher{}

	custSvc := service.NewCustomerService(custRepo, pub)
	oppSvc := service.NewOpportunityService(oppRepo, memory.NewOpportunityStageHistoryRepository(), pub)
	svc := service.NewLeadService(leadRepo, custSvc, oppSvc, pub)

	ctx := context.Background()

	// 1. Create Lead
	lead, err := svc.CreateLead(ctx, "Bob", "Smith", "Smith Co", "bob@smith.com", "123", "WEB")
	if err != nil {
		t.Fatalf("failed to create lead: %v", err)
	}
	if lead.FirstName != "Bob" {
		t.Errorf("expected Bob, got %s", lead.FirstName)
	}

	// Verify create event
	foundCreated := false
	for _, ev := range pub.Events {
		if ev.Topic == domain.TopicCrmLeadCreated {
			foundCreated = true
		}
	}
	if !foundCreated {
		t.Errorf("expected lead created event to be published")
	}

	// 2. Get Lead
	fetched, err := svc.GetLead(ctx, lead.ID)
	if err != nil {
		t.Fatalf("failed to get lead: %v", err)
	}
	if fetched.ID != lead.ID {
		t.Errorf("expected lead ID %q, got %q", lead.ID, fetched.ID)
	}

	// 3. List Leads
	list, err := svc.ListLeads(ctx)
	if err != nil {
		t.Fatalf("failed to list leads: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected list length 1, got %d", len(list))
	}

	// 4. Update Lead -> QUALIFIED
	pub.Events = nil
	_, err = svc.UpdateLead(ctx, lead.ID, "Bob", "Smith", "Smith Co", "QUALIFIED", 80)
	if err != nil {
		t.Fatalf("failed to update lead: %v", err)
	}
	foundQualified := false
	for _, ev := range pub.Events {
		if ev.Topic == domain.TopicCrmLeadQualified {
			foundQualified = true
		}
	}
	if !foundQualified {
		t.Errorf("expected lead qualified event to be published")
	}

	// 5. Update Lead -> LOST
	pub.Events = nil
	_, err = svc.UpdateLead(ctx, lead.ID, "Bob", "Smith", "Smith Co", "LOST", 10)
	if err != nil {
		t.Fatalf("failed to update lead: %v", err)
	}
	foundLost := false
	for _, ev := range pub.Events {
		if ev.Topic == domain.TopicCrmLeadLost {
			foundLost = true
		}
	}
	if !foundLost {
		t.Errorf("expected lead lost event to be published")
	}

	// 6. Delete Lead
	err = svc.DeleteLead(ctx, lead.ID)
	if err != nil {
		t.Fatalf("failed to delete lead: %v", err)
	}
	_, err = svc.GetLead(ctx, lead.ID)
	if err == nil {
		t.Errorf("expected error when getting deleted lead, got nil")
	}
}

func TestLeadService_ConvertLead_Errors(t *testing.T) {
	leadRepo := memory.NewLeadRepository()
	custRepo := memory.NewCustomerRepository()
	oppRepo := memory.NewOpportunityRepository()
	pub := &sharedtesting.MockPublisher{}

	custSvc := service.NewCustomerService(custRepo, pub)
	oppSvc := service.NewOpportunityService(oppRepo, memory.NewOpportunityStageHistoryRepository(), pub)
	svc := service.NewLeadService(leadRepo, custSvc, oppSvc, pub)

	ctx := context.Background()

	// 1. Convert non-existent lead
	_, err := svc.ConvertLead(ctx, "non-existent")
	if err == nil {
		t.Errorf("expected error converting non-existent lead, got nil")
	}

	// 2. Convert already converted lead
	lead := &domain.Lead{
		ID:     "lead_converted",
		Status: "CONVERTED",
	}
	_ = leadRepo.Create(ctx, lead)
	_, err = svc.ConvertLead(ctx, lead.ID)
	if err == nil {
		t.Errorf("expected error converting already converted lead, got nil")
	}
}

func TestLeadService_ConvertLead_RollbackOnCustomerFailure(t *testing.T) {
	leadRepo := memory.NewLeadRepository()
	// Use failing customer repo
	custRepo := &MockFailCustomerRepo{memory.NewCustomerRepository()}
	oppRepo := memory.NewOpportunityRepository()
	pub := &sharedtesting.MockPublisher{}

	custSvc := service.NewCustomerService(custRepo, pub)
	oppSvc := service.NewOpportunityService(oppRepo, memory.NewOpportunityStageHistoryRepository(), pub)
	svc := service.NewLeadService(leadRepo, custSvc, oppSvc, pub)

	ctx := context.Background()

	lead := &domain.Lead{
		ID:        "lead_123",
		FirstName: "Jane",
		LastName:  "Doe",
		Company:   "TechCorp",
		Email:     "jane@techcorp.com",
		Status:    "NEW",
	}
	_ = leadRepo.Create(ctx, lead)

	_, err := svc.ConvertLead(ctx, lead.ID)
	if err == nil {
		t.Fatalf("expected error converting lead, got nil")
	}

	// Verify Lead status was rolled back to NEW
	dbLead, _ := leadRepo.GetByID(ctx, lead.ID)
	if dbLead.Status != "NEW" {
		t.Errorf("expected lead status NEW, got %s", dbLead.Status)
	}
}

func TestLeadService_UpdateErrors(t *testing.T) {
	leadRepo := memory.NewLeadRepository()
	pub := &sharedtesting.MockPublisher{}
	svc := service.NewLeadService(leadRepo, nil, nil, pub)

	ctx := context.Background()
	_, err := svc.UpdateLead(ctx, "non-existent", "a", "b", "c", "LOST", 0)
	if err == nil {
		t.Errorf("expected error updating non-existent lead, got nil")
	}
}
