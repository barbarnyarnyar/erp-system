package service

import (
	"log"
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
	stageEnum := domain.OpportunityStage(stage)
	if !stageEnum.IsValid() {
		return nil, fmt.Errorf("invalid opportunity stage: %s", stage)
	}

	id := fmt.Sprintf("opp_%d", time.Now().UnixNano())
	opp := &domain.Opportunity{
		ID:          id,
		CustomerID:  customerID,
		Title:       title,
		Value:       value,
		Status:      "NEW",
		Stage:       stageEnum,
		Probability: decimal.NewFromFloat(0.10),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := s.oppRepo.Create(ctx, opp)
	if err != nil {
		return nil, err
	}

	if err := s.publisher.Publish(ctx, domain.TopicCrmOpportunityCreated, id, domain.OpportunityCreatedEvent{
		OpportunityID: id,
		CustomerID:    customerID,
		Title:         title,
		Value:         value,
		Timestamp:     time.Now(),
	}); err != nil {
		log.Printf("ERROR: failed to publish event %s: %v", domain.TopicCrmOpportunityCreated, err)
	}

	return opp, nil
}

func (s *OpportunityService) GetOpportunity(ctx context.Context, id string) (*domain.Opportunity, error) {
	return s.oppRepo.GetByID(ctx, id)
}

func (s *OpportunityService) ListOpportunities(ctx context.Context) ([]domain.Opportunity, error) {
	return s.oppRepo.List(ctx)
}

func (s *OpportunityService) UpdateOpportunity(ctx context.Context, id string, title string, value decimal.Decimal, status, stage string, probability decimal.Decimal) (*domain.Opportunity, error) {
	stageEnum := domain.OpportunityStage(stage)
	if !stageEnum.IsValid() {
		return nil, fmt.Errorf("invalid opportunity stage: %s", stage)
	}

	opp, err := s.oppRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	oldStatus := opp.Status
	opp.Title = title
	opp.Value = value
	opp.Status = status
	opp.Stage = stageEnum
	opp.Probability = probability
	opp.UpdatedAt = time.Now()

	err = s.oppRepo.Update(ctx, opp)
	if err != nil {
		return nil, err
	}

	if err := s.publisher.Publish(ctx, domain.TopicCrmOpportunityUpdated, id, domain.OpportunityUpdatedEvent{
		OpportunityID: id,
		Status:        status,
		Stage:         stage,
		Value:         value,
		Timestamp:     time.Now(),
	}); err != nil {
		log.Printf("ERROR: failed to publish event %s: %v", domain.TopicCrmOpportunityUpdated, err)
	}

	if oldStatus != status {
		if status == "WON" {
			if err := s.publisher.Publish(ctx, domain.TopicCrmOpportunityWon, id, domain.OpportunityWonEvent{
				OpportunityID: id,
				CustomerID:    opp.CustomerID,
				Value:         value,
				Timestamp:     time.Now(),
			}); err != nil {
				log.Printf("ERROR: failed to publish event %s: %v", domain.TopicCrmOpportunityWon, err)
			}
		} else if status == "LOST" {
			if err := s.publisher.Publish(ctx, domain.TopicCrmOpportunityLost, id, domain.OpportunityLostEvent{
				OpportunityID: id,
				Timestamp:     time.Now(),
			}); err != nil {
				log.Printf("ERROR: failed to publish event %s: %v", domain.TopicCrmOpportunityLost, err)
			}
		}
	}

	return opp, nil
}

func (s *OpportunityService) DeleteOpportunity(ctx context.Context, id string) error {
	return s.oppRepo.Delete(ctx, id)
}
