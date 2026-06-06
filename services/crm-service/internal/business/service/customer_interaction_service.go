package service

import (
	"context"
	"erp-system/shared/utils"
	"fmt"
	"time"

	"github.com/erp-system/crm-service/internal/business/domain"
)

type CustomerInteractionService struct {
	repo      domain.CustomerInteractionRepository
	publisher domain.EventPublisher
}

func NewCustomerInteractionService(repo domain.CustomerInteractionRepository, publisher domain.EventPublisher) *CustomerInteractionService {
	return &CustomerInteractionService{repo: repo, publisher: publisher}
}

func (s *CustomerInteractionService) CreateCustomerInteraction(ctx context.Context, customerID, typ, subject, description string, interactionDate time.Time, createdBy string) (*domain.CustomerInteraction, error) {
	if customerID == "" || typ == "" {
		return nil, fmt.Errorf("customer_id and type are required")
	}
	id := utils.NewID("ci")
	ci := &domain.CustomerInteraction{
		ID:              id,
		CustomerID:      customerID,
		Type:            typ,
		Subject:         subject,
		Description:     description,
		InteractionDate: interactionDate,
		CreatedBy:       createdBy,
		CreatedAt:       time.Now(),
	}
	if err := s.repo.Create(ctx, ci); err != nil {
		return nil, err
	}
	if err := s.publisher.Publish(ctx, domain.TopicCrmCustomerInteractionLogged, ci.ID, domain.CustomerInteractionLoggedEvent{
		InteractionID:   ci.ID,
		CustomerID:      ci.CustomerID,
		Type:            ci.Type,
		Subject:         ci.Subject,
		InteractionDate: ci.InteractionDate,
		CreatedBy:       ci.CreatedBy,
		Timestamp:       time.Now(),
	}); err != nil {
		utils.LogPublishErr("crm-service", domain.TopicCrmCustomerInteractionLogged, err)
	}
	return ci, nil
}

func (s *CustomerInteractionService) GetCustomerInteraction(ctx context.Context, id string) (*domain.CustomerInteraction, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *CustomerInteractionService) ListCustomerInteractions(ctx context.Context, customerID string) ([]domain.CustomerInteraction, error) {
	return s.repo.ListByCustomerID(ctx, customerID)
}

func (s *CustomerInteractionService) DeleteCustomerInteraction(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
