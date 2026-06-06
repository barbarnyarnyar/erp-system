package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/erp-system/crm-service/internal/business/domain"
	"github.com/erp-system/crm-service/internal/business/service"
	"github.com/erp-system/crm-service/internal/data/memory"
)

// MockPublisher tracks events for testing
type MockPublisher struct {
	Events      []MockEvent
	FailPublish bool
}

type MockEvent struct {
	Topic   string
	Key     string
	Payload interface{}
}

func (m *MockPublisher) Publish(ctx context.Context, topic string, key string, payload interface{}) error {
	if m.FailPublish {
		return errors.New("failed to publish")
	}
	m.Events = append(m.Events, MockEvent{
		Topic:   topic,
		Key:     key,
		Payload: payload,
	})
	return nil
}

// MockOpportunityRepo that can fail on Create to test rollback
type MockFailOpportunityRepo struct {
	domain.OpportunityRepository
}

func (m *MockFailOpportunityRepo) Create(ctx context.Context, opp *domain.Opportunity) error {
	return errors.New("mock db failure during opportunity creation")
}

func TestLeadService_ConvertLead_Success(t *testing.T) {
	leadRepo := memory.NewLeadRepository()
	custRepo := memory.NewCustomerRepository()
	oppRepo := memory.NewOpportunityRepository()
	publisher := &MockPublisher{}

	custSvc := service.NewCustomerService(custRepo, publisher)
	oppSvc := service.NewOpportunityService(oppRepo, publisher)
	leadSvc := service.NewLeadService(leadRepo, custSvc, oppSvc, publisher)

	ctx := context.Background()

	// Seed lead
	lead := &domain.Lead{
		ID:        "lead_123",
		FirstName: "Jane",
		LastName:  "Doe",
		Company:   "TechCorp",
		Email:     "jane@techcorp.com",
		Phone:     "123456",
		Status:    "NEW",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = leadRepo.Create(ctx, lead)

	opp, err := leadSvc.ConvertLead(ctx, lead.ID)
	if err != nil {
		t.Fatalf("unexpected error converting lead: %v", err)
	}

	if opp.Title != "Opportunity from Lead TechCorp" {
		t.Errorf("unexpected opportunity title: %s", opp.Title)
	}

	// Verify Lead status is now CONVERTED
	dbLead, _ := leadRepo.GetByID(ctx, lead.ID)
	if dbLead.Status != "CONVERTED" {
		t.Errorf("expected lead status CONVERTED, got %s", dbLead.Status)
	}

	// Verify Customer was created
	customers, _ := custRepo.List(ctx)
	if len(customers) != 1 {
		t.Errorf("expected 1 customer, got %d", len(customers))
	}
}

func TestLeadService_ConvertLead_RollbackOnOpportunityFailure(t *testing.T) {
	leadRepo := memory.NewLeadRepository()
	custRepo := memory.NewCustomerRepository()
	// Use failing opportunity repo
	oppRepo := &MockFailOpportunityRepo{memory.NewOpportunityRepository()}
	publisher := &MockPublisher{}

	custSvc := service.NewCustomerService(custRepo, publisher)
	oppSvc := service.NewOpportunityService(oppRepo, publisher)
	leadSvc := service.NewLeadService(leadRepo, custSvc, oppSvc, publisher)

	ctx := context.Background()

	// Seed lead
	lead := &domain.Lead{
		ID:        "lead_123",
		FirstName: "Jane",
		LastName:  "Doe",
		Company:   "TechCorp",
		Email:     "jane@techcorp.com",
		Phone:     "123456",
		Status:    "NEW",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = leadRepo.Create(ctx, lead)

	_, err := leadSvc.ConvertLead(ctx, lead.ID)
	if err == nil {
		t.Fatalf("expected error converting lead, got nil")
	}

	// Verify Lead status was rolled back to NEW
	dbLead, _ := leadRepo.GetByID(ctx, lead.ID)
	if dbLead.Status != "NEW" {
		t.Errorf("expected lead status NEW, got %s", dbLead.Status)
	}

	// Verify Customer creation was rolled back (deleted)
	customers, _ := custRepo.List(ctx)
	if len(customers) != 0 {
		t.Errorf("expected 0 customers (rolled back), got %d", len(customers))
	}
}

func TestLeadService_ConvertLead_RollbackOnPublishFailure(t *testing.T) {
	leadRepo := memory.NewLeadRepository()
	custRepo := memory.NewCustomerRepository()
	oppRepo := memory.NewOpportunityRepository()
	publisher := &MockPublisher{FailPublish: true}

	custSvc := service.NewCustomerService(custRepo, publisher)
	oppSvc := service.NewOpportunityService(oppRepo, publisher)
	leadSvc := service.NewLeadService(leadRepo, custSvc, oppSvc, publisher)

	ctx := context.Background()

	// Seed lead
	lead := &domain.Lead{
		ID:        "lead_123",
		FirstName: "Jane",
		LastName:  "Doe",
		Company:   "TechCorp",
		Email:     "jane@techcorp.com",
		Phone:     "123456",
		Status:    "NEW",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = leadRepo.Create(ctx, lead)

	_, err := leadSvc.ConvertLead(ctx, lead.ID)
	if err == nil {
		t.Fatalf("expected error converting lead, got nil")
	}

	// Verify Lead status was rolled back to NEW
	dbLead, _ := leadRepo.GetByID(ctx, lead.ID)
	if dbLead.Status != "NEW" {
		t.Errorf("expected lead status NEW, got %s", dbLead.Status)
	}

	// Verify Customer creation was rolled back (deleted)
	customers, _ := custRepo.List(ctx)
	if len(customers) != 0 {
		t.Errorf("expected 0 customers (rolled back), got %d", len(customers))
	}

	// Verify Opportunity was rolled back (deleted)
	opps, _ := oppRepo.List(ctx)
	if len(opps) != 0 {
		t.Errorf("expected 0 opportunities, got %d", len(opps))
	}
}
