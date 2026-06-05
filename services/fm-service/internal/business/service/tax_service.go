package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/erp-system/fm-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type TaxService struct{}

func NewTaxService() *TaxService {
	return &TaxService{}
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
	
	// Normally we would persist this via a repository
	return taxRate, nil
}
