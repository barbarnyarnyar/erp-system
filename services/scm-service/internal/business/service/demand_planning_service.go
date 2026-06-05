package service

import (
	"context"
	"fmt"
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

func (s *DemandPlanningService) CreateForecast(ctx context.Context, productID string, forecastDate time.Time, qty int, confidence decimal.Decimal, notes string) (*domain.DemandForecast, error) {
	id := fmt.Sprintf("fore_%d", time.Now().UnixNano())

	df := &domain.DemandForecast{
		ID:               id,
		ProductID:        productID,
		ForecastDate:     forecastDate,
		ForecastQuantity: qty,
		ConfidenceLevel:  confidence,
		Notes:            notes,
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

func (s *DemandPlanningService) UpdateForecast(ctx context.Context, id string, forecastDate time.Time, qty int, confidence decimal.Decimal, notes string) (*domain.DemandForecast, error) {
	df, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	df.ForecastDate = forecastDate
	df.ForecastQuantity = qty
	df.ConfidenceLevel = confidence
	df.Notes = notes
	df.UpdatedAt = time.Now()

	err = s.repo.Update(ctx, df)
	if err != nil {
		return nil, err
	}

	return df, nil
}
