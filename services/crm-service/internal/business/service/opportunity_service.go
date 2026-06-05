package service

import (
	"context"
	"fmt"
	"time"

	"github.com/erp-system/crm-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type OpportunityService struct {
	oppRepo   domain.OpportunityRepository
	publisher domain.EventPublisher
}

func NewOpportunityService(oppRepo domain.OpportunityRepository, publisher domain.EventPublisher) *OpportunityService {
	return &OpportunityService{
		oppRepo:   oppRepo,
		publisher: publisher,
	}
}

func (s *OpportunityService) CreateOpportunity(ctx context.Context, customerID, title string, value decimal.Decimal, stage string) (*domain.Opportunity, error) {
	id := fmt.Sprintf("opp_%d", time.Now().UnixNano())
	opp := &domain.Opportunity{
		ID:          id,
		CustomerID:  customerID,
		Title:       title,
		Value:       value,
		Status:      "NEW",
		Stage:       stage,
		Probability: decimal.NewFromFloat(0.10),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := s.oppRepo.Create(ctx, opp)
	if err != nil {
		return nil, err
	}

	_ = s.publisher.Publish(ctx, domain.TopicCrmOpportunityCreated, id, domain.OpportunityCreatedEvent{
		OpportunityID: id,
		CustomerID:    customerID,
		Title:         title,
		Value:         value,
		Timestamp:     time.Now(),
	})

	return opp, nil
}

func (s *OpportunityService) GetOpportunity(ctx context.Context, id string) (*domain.Opportunity, error) {
	return s.oppRepo.GetByID(ctx, id)
}

func (s *OpportunityService) ListOpportunities(ctx context.Context) ([]domain.Opportunity, error) {
	return s.oppRepo.List(ctx)
}

func (s *OpportunityService) UpdateOpportunity(ctx context.Context, id string, title string, value decimal.Decimal, status, stage string, probability decimal.Decimal) (*domain.Opportunity, error) {
	opp, err := s.oppRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	oldStatus := opp.Status
	opp.Title = title
	opp.Value = value
	opp.Status = status
	opp.Stage = stage
	opp.Probability = probability
	opp.UpdatedAt = time.Now()

	err = s.oppRepo.Update(ctx, opp)
	if err != nil {
		return nil, err
	}

	_ = s.publisher.Publish(ctx, domain.TopicCrmOpportunityUpdated, id, domain.OpportunityUpdatedEvent{
		OpportunityID: id,
		Status:        status,
		Stage:         stage,
		Value:         value,
		Timestamp:     time.Now(),
	})

	if oldStatus != status {
		if status == "WON" {
			_ = s.publisher.Publish(ctx, domain.TopicCrmOpportunityWon, id, domain.OpportunityWonEvent{
				OpportunityID: id,
				CustomerID:    opp.CustomerID,
				Value:         value,
				Timestamp:     time.Now(),
			})
		} else if status == "LOST" {
			_ = s.publisher.Publish(ctx, domain.TopicCrmOpportunityLost, id, domain.OpportunityLostEvent{
				OpportunityID: id,
				Timestamp:     time.Now(),
			})
		}
	}

	return opp, nil
}

func (s *OpportunityService) DeleteOpportunity(ctx context.Context, id string) error {
	return s.oppRepo.Delete(ctx, id)
}
