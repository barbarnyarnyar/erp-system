package service

import (
	"context"
	"erp-system/shared/utils"
	"time"

	"github.com/erp-system/scm-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type DemandPlanningService struct {
	repo domain.DemandForecastRepository
}

func NewDemandPlanningService(repo domain.DemandForecastRepository) *DemandPlanningService {
	return &DemandPlanningService{repo: repo}
}

func (s *DemandPlanningService) ListForecasts(ctx context.Context) ([]domain.DemandForecast, error) {
	return s.repo.List(ctx)
}

func (s *DemandPlanningService) CreateForecast(ctx context.Context, materialID string, forecastDate time.Time, qty decimal.Decimal, confidence decimal.Decimal, notes string) (*domain.DemandForecast, error) {
	id := utils.NewID("fore")

	df := &domain.DemandForecast{
		ID:               id,
		LegalEntityID:    "00000000-0000-0000-0000-000000000000",
		MaterialID:       materialID,
		ForecastDate:     forecastDate,
		ForecastQuantity: qty,
		ConfidenceLevel:  confidence,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	err := s.repo.Create(ctx, df)
	if err != nil {
		return nil, err
	}

	return df, nil
}

func (s *DemandPlanningService) GetForecast(ctx context.Context, id string) (*domain.DemandForecast, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *DemandPlanningService) UpdateForecast(ctx context.Context, id string, forecastDate time.Time, qty decimal.Decimal, confidence decimal.Decimal, notes string) (*domain.DemandForecast, error) {
	df, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	df.ForecastDate = forecastDate
	df.ForecastQuantity = qty
	df.ConfidenceLevel = confidence
	df.UpdatedAt = time.Now()

	err = s.repo.Update(ctx, df)
	if err != nil {
		return nil, err
	}

	return df, nil
}
