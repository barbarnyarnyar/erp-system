package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/erp-system/fm-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type TaxService struct {
	repo domain.TaxRateRepository
}

func NewTaxService(repo domain.TaxRateRepository) *TaxService {
	return &TaxService{repo: repo}
}

func (s *TaxService) CreateTaxRate(ctx context.Context, code, name string, rate decimal.Decimal) (*domain.TaxRate, error) {
	if code == "" || name == "" {
		return nil, errors.New("tax code and name are required")
	}

	id := fmt.Sprintf("tax_%s", code)
	taxRate := &domain.TaxRate{
		ID:       id,
		Code:     code,
		Name:     name,
		Rate:     rate,
		IsActive: true,
	}
	
	err := s.repo.Create(ctx, taxRate)
	if err != nil {
		return nil, err
	}

	return taxRate, nil
}

func (s *TaxService) ListTaxRates(ctx context.Context) ([]domain.TaxRate, error) {
	return s.repo.List(ctx)
}

func (s *TaxService) GetTaxRate(ctx context.Context, id string) (*domain.TaxRate, error) {
	return s.repo.GetByID(ctx, id)
}

