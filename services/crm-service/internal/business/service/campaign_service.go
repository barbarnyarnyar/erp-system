package service

import (
	"context"
	"erp-system/shared/utils"
	"time"

	"github.com/erp-system/crm-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type CampaignService struct {
	campaignRepo domain.CampaignRepository
	publisher    domain.EventPublisher
}

func NewCampaignService(campaignRepo domain.CampaignRepository, publisher domain.EventPublisher) *CampaignService {
	return &CampaignService{
		campaignRepo: campaignRepo,
		publisher:    publisher,
	}
}

func (s *CampaignService) CreateCampaign(ctx context.Context, name, campaignType string, budget decimal.Decimal) (*domain.Campaign, error) {
	id := utils.NewID("camp")
	camp := &domain.Campaign{
		ID:        id,
		Name:      name,
		Type:      campaignType,
		Status:    "DRAFT",
		Budget:    budget,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := s.campaignRepo.Create(ctx, camp)
	if err != nil {
		return nil, err
	}

	return camp, nil
}

func (s *CampaignService) GetCampaign(ctx context.Context, id string) (*domain.Campaign, error) {
	return s.campaignRepo.GetByID(ctx, id)
}

func (s *CampaignService) ListCampaigns(ctx context.Context) ([]domain.Campaign, error) {
	return s.campaignRepo.List(ctx)
}

func (s *CampaignService) UpdateCampaign(ctx context.Context, id string, status string, budget decimal.Decimal) (*domain.Campaign, error) {
	camp, err := s.campaignRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	oldStatus := camp.Status
	camp.Status = status
	camp.Budget = budget
	camp.UpdatedAt = time.Now()

	err = s.campaignRepo.Update(ctx, camp)
	if err != nil {
		return nil, err
	}

	if oldStatus != status {
		if status == "LAUNCHED" {
			if err := s.publisher.Publish(ctx, domain.TopicCrmCampaignLaunched, id, domain.CampaignLaunchedEvent{
				CampaignID: id,
				Name:       camp.Name,
				Timestamp:  time.Now(),
			}); err != nil {
				utils.LogPublishErr("crm-service", domain.TopicCrmCampaignLaunched, err)
			}
		} else if status == "COMPLETED" {
			if err := s.publisher.Publish(ctx, domain.TopicCrmCampaignCompleted, id, domain.CampaignCompletedEvent{
				CampaignID: id,
				Timestamp:  time.Now(),
			}); err != nil {
				utils.LogPublishErr("crm-service", domain.TopicCrmCampaignCompleted, err)
			}
		}
	}

	return camp, nil
}

func (s *CampaignService) DeleteCampaign(ctx context.Context, id string) error {
	return s.campaignRepo.Delete(ctx, id)
}
