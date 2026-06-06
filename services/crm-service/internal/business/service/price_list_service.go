package service

import (
	"context"
	"erp-system/shared/utils"
	"time"

	"github.com/erp-system/crm-service/internal/business/domain"
)

type PriceListService struct {
	priceListRepo     domain.PriceListRepository
	priceListItemRepo domain.PriceListItemRepository
}

func NewPriceListService(priceListRepo domain.PriceListRepository, priceListItemRepo domain.PriceListItemRepository) *PriceListService {
	return &PriceListService{
		priceListRepo:     priceListRepo,
		priceListItemRepo: priceListItemRepo,
	}
}

func (s *PriceListService) CreatePriceList(ctx context.Context, name, description string, isActive bool) (*domain.PriceList, error) {
	id := utils.NewID("pl")
	pl := &domain.PriceList{
		ID:          id,
		Name:        name,
		Description: description,
		IsActive:    isActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := s.priceListRepo.Create(ctx, pl)
	if err != nil {
		return nil, err
	}

	return pl, nil
}

func (s *PriceListService) GetPriceList(ctx context.Context, id string) (*domain.PriceList, error) {
	return s.priceListRepo.GetByID(ctx, id)
}

func (s *PriceListService) ListPriceLists(ctx context.Context) ([]domain.PriceList, error) {
	return s.priceListRepo.List(ctx)
}

func (s *PriceListService) UpdatePriceList(ctx context.Context, id string, name, description string, isActive bool) (*domain.PriceList, error) {
	pl, err := s.priceListRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	pl.Name = name
	pl.Description = description
	pl.IsActive = isActive
	pl.UpdatedAt = time.Now()

	err = s.priceListRepo.Update(ctx, pl)
	if err != nil {
		return nil, err
	}

	return pl, nil
}

func (s *PriceListService) DeletePriceList(ctx context.Context, id string) error {
	return s.priceListRepo.Delete(ctx, id)
}
