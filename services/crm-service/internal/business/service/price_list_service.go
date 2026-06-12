package service

import (
	"context"
	"erp-system/shared/utils"
	"time"

	"github.com/erp-system/crm-service/internal/business/domain"
)

type PriceListService struct {
	priceListRepo     domain.PriceBookHeaderRepository
	priceListItemRepo domain.PriceBookEntryRepository
}

func NewPriceListService(priceListRepo domain.PriceBookHeaderRepository, priceListItemRepo domain.PriceBookEntryRepository) *PriceListService {
	return &PriceListService{
		priceListRepo:     priceListRepo,
		priceListItemRepo: priceListItemRepo,
	}
}

func (s *PriceListService) CreatePriceList(ctx context.Context, name, description string, isActive bool) (*domain.PriceBookHeader, error) {
	id := utils.NewID("pl")
	pl := &domain.PriceBookHeader{
		ID:            id,
		LegalEntityID: "default_entity_id",
		PriceBookCode: "PB-" + id[:8],
		Name:          name,
		Type:          domain.PriceBookTypeSTANDARD,
		StartDate:     time.Now(),
		EndDate:       time.Now().AddDate(1, 0, 0),
		IsActive:      isActive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := s.priceListRepo.Create(ctx, pl)
	if err != nil {
		return nil, err
	}

	return pl, nil
}

func (s *PriceListService) GetPriceList(ctx context.Context, id string) (*domain.PriceBookHeader, error) {
	return s.priceListRepo.GetByID(ctx, id)
}

func (s *PriceListService) ListPriceLists(ctx context.Context) ([]domain.PriceBookHeader, error) {
	return s.priceListRepo.List(ctx)
}

func (s *PriceListService) UpdatePriceList(ctx context.Context, id string, name, description string, isActive bool) (*domain.PriceBookHeader, error) {
	pl, err := s.priceListRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	pl.Name = name
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
