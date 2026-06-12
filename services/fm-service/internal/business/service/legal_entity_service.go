package service

import (
	"context"
	"errors"
	"time"

	"erp-system/shared/utils"
	"github.com/erp-system/fm-service/internal/business/domain"
)

type LegalEntityService struct {
	repo domain.LegalEntityRepository
	tm   domain.TransactionManager
}

func NewLegalEntityService(repo domain.LegalEntityRepository, tm domain.TransactionManager) *LegalEntityService {
	return &LegalEntityService{
		repo: repo,
		tm:   tm,
	}
}

func (s *LegalEntityService) CreateLegalEntity(ctx context.Context, companyCode, companyName, functionalCurrency, taxRegistrationNumber string) (*domain.LegalEntity, error) {
	if companyCode == "" || companyName == "" || functionalCurrency == "" || taxRegistrationNumber == "" {
		return nil, errors.New("company code, company name, functional currency, and tax registration number are required")
	}

	if len(functionalCurrency) != 3 {
		return nil, errors.New("functional currency must be a valid 3-letter ISO 4217 code")
	}

	le := &domain.LegalEntity{
		ID:                    utils.NewID("le"),
		CompanyCode:           companyCode,
		CompanyName:           companyName,
		FunctionalCurrency:    functionalCurrency,
		TaxRegistrationNumber: taxRegistrationNumber,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	err := s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		return s.repo.Create(txCtx, le)
	})
	if err != nil {
		return nil, err
	}

	return le, nil
}

func (s *LegalEntityService) GetLegalEntity(ctx context.Context, id string) (*domain.LegalEntity, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *LegalEntityService) GetLegalEntityByCode(ctx context.Context, code string) (*domain.LegalEntity, error) {
	return s.repo.GetByCode(ctx, code)
}

func (s *LegalEntityService) ListLegalEntities(ctx context.Context) ([]domain.LegalEntity, error) {
	return s.repo.List(ctx)
}
