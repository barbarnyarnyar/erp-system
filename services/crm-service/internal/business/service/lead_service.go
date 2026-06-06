package service

import (
	"log"
	"context"
	"fmt"
	"time"

	"github.com/erp-system/crm-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type LeadService struct {
	leadRepo   domain.LeadRepository
	custSvc    *CustomerService
	oppSvc     *OpportunityService
	publisher  domain.EventPublisher
}

func NewLeadService(
	leadRepo domain.LeadRepository,
	custSvc *CustomerService,
	oppSvc *OpportunityService,
	publisher domain.EventPublisher,
) *LeadService {
	return &LeadService{
		leadRepo:  leadRepo,
		custSvc:   custSvc,
		oppSvc:    oppSvc,
		publisher: publisher,
	}
}

func (s *LeadService) CreateLead(ctx context.Context, firstName, lastName, company, email, phone, source string) (*domain.Lead, error) {
	id := fmt.Sprintf("lead_%d", time.Now().UnixNano())
	lead := &domain.Lead{
		ID:        id,
		FirstName: firstName,
		LastName:  lastName,
		Company:   company,
		Email:     email,
		Phone:     phone,
		Status:    "NEW",
		Score:     10,
		Source:    source,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := s.leadRepo.Create(ctx, lead)
	if err != nil {
		return nil, err
	}

	if err := s.publisher.Publish(ctx, domain.TopicCrmLeadCreated, id, domain.LeadCreatedEvent{
		LeadID:    id,
		Company:   company,
		Email:     email,
		Timestamp: time.Now(),
	}); err != nil {
		log.Printf("ERROR: failed to publish event %s: %v", domain.TopicCrmLeadCreated, err)
	}

	return lead, nil
}

func (s *LeadService) GetLead(ctx context.Context, id string) (*domain.Lead, error) {
	return s.leadRepo.GetByID(ctx, id)
}

func (s *LeadService) ListLeads(ctx context.Context) ([]domain.Lead, error) {
	return s.leadRepo.List(ctx)
}

func (s *LeadService) UpdateLead(ctx context.Context, id string, firstName, lastName, company, status string, score int) (*domain.Lead, error) {
	lead, err := s.leadRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	oldStatus := lead.Status
	lead.FirstName = firstName
	lead.LastName = lastName
	lead.Company = company
	lead.Status = status
	lead.Score = score
	lead.UpdatedAt = time.Now()

	err = s.leadRepo.Update(ctx, lead)
	if err != nil {
		return nil, err
	}

	if oldStatus != status {
		if status == "QUALIFIED" {
			if err := s.publisher.Publish(ctx, domain.TopicCrmLeadQualified, id, domain.LeadQualifiedEvent{
				LeadID:    id,
				Score:     score,
				Timestamp: time.Now(),
			}); err != nil {
				log.Printf("ERROR: failed to publish event %s: %v", domain.TopicCrmLeadQualified, err)
			}
		} else if status == "LOST" {
			if err := s.publisher.Publish(ctx, domain.TopicCrmLeadLost, id, domain.LeadLostEvent{
				LeadID:    id,
				Timestamp: time.Now(),
			}); err != nil {
				log.Printf("ERROR: failed to publish event %s: %v", domain.TopicCrmLeadLost, err)
			}
		}
	}

	return lead, nil
}

func (s *LeadService) DeleteLead(ctx context.Context, id string) error {
	return s.leadRepo.Delete(ctx, id)
}

func (s *LeadService) ConvertLead(ctx context.Context, id string) (*domain.Opportunity, error) {
	lead, err := s.leadRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	lead.Status = "CONVERTED"
	lead.UpdatedAt = time.Now()
	_ = s.leadRepo.Update(ctx, lead)

	// Create Customer via CustomerService
	custName := fmt.Sprintf("%s %s", lead.FirstName, lead.LastName)
	cust, err := s.custSvc.CreateCustomer(ctx, lead.Company, custName, lead.Email, lead.Phone, "RETAIL", "")
	if err != nil {
		return nil, err
	}

	// Create Opportunity via OpportunityService
	opp, err := s.oppSvc.CreateOpportunity(ctx, cust.ID, "Opportunity from Lead "+lead.Company, decimal.NewFromFloat(5000.00), "QUALIFIED")
	if err != nil {
		return nil, err
	}

	if err := s.publisher.Publish(ctx, domain.TopicCrmLeadConverted, id, domain.LeadConvertedEvent{
		LeadID:        id,
		CustomerID:    cust.ID,
		OpportunityID: opp.ID,
		Timestamp:     time.Now(),
	}); err != nil {
		log.Printf("ERROR: failed to publish event %s: %v", domain.TopicCrmLeadConverted, err)
	}

	return opp, nil
}
