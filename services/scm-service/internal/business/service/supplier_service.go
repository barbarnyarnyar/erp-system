package service

import (
	"context"
	"fmt"
	"time"

	"github.com/erp-system/scm-service/internal/business/domain"
)

type SupplierManagementService struct {
	repo     domain.SupplierRepository
	contRepo domain.VendorContractRepository
}

func NewSupplierManagementService(repo domain.SupplierRepository, contRepo domain.VendorContractRepository) *SupplierManagementService {
	return &SupplierManagementService{
		repo:     repo,
		contRepo: contRepo,
	}
}

func (s *SupplierManagementService) ListSuppliers(ctx context.Context) ([]domain.Supplier, error) {
	return s.repo.List(ctx)
}

func (s *SupplierManagementService) CreateSupplier(ctx context.Context, code, name, contact, email, phone string) (*domain.Supplier, error) {
	id := fmt.Sprintf("supp_%d", time.Now().UnixNano())

	sup := &domain.Supplier{
		ID:           id,
		SupplierCode: code,
		SupplierName: name,
		ContactName:  contact,
		Email:        email,
		Phone:        phone,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := s.repo.Create(ctx, sup)
	if err != nil {
		return nil, err
	}

	return sup, nil
}

func (s *SupplierManagementService) GetSupplier(ctx context.Context, id string) (*domain.Supplier, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *SupplierManagementService) UpdateSupplier(ctx context.Context, id, code, name, contact, email, phone string, isActive bool) (*domain.Supplier, error) {
	sup, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	sup.SupplierCode = code
	sup.SupplierName = name
	sup.ContactName = contact
	sup.Email = email
	sup.Phone = phone
	sup.IsActive = isActive
	sup.UpdatedAt = time.Now()

	err = s.repo.Update(ctx, sup)
	if err != nil {
		return nil, err
	}

	return sup, nil
}

func (s *SupplierManagementService) DeleteSupplier(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// Vendor Contracts CRUD

func (s *SupplierManagementService) ListContracts(ctx context.Context) ([]domain.VendorContract, error) {
	return s.contRepo.List(ctx)
}

func (s *SupplierManagementService) CreateContract(ctx context.Context, contractNum, supplierID string, startDate, endDate time.Time, terms string) (*domain.VendorContract, error) {
	id := fmt.Sprintf("cont_%d", time.Now().UnixNano())
	vc := &domain.VendorContract{
		ID:             id,
		ContractNumber: contractNum,
		SupplierID:     supplierID,
		StartDate:      startDate,
		EndDate:        endDate,
		Terms:          terms,
		Status:         "ACTIVE",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err := s.contRepo.Create(ctx, vc)
	if err != nil {
		return nil, err
	}
	return vc, nil
}

func (s *SupplierManagementService) GetContract(ctx context.Context, id string) (*domain.VendorContract, error) {
	return s.contRepo.GetByID(ctx, id)
}

func (s *SupplierManagementService) UpdateContract(ctx context.Context, id, contractNum, supplierID string, startDate, endDate time.Time, terms, status string) (*domain.VendorContract, error) {
	vc, err := s.contRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	vc.ContractNumber = contractNum
	vc.SupplierID = supplierID
	vc.StartDate = startDate
	vc.EndDate = endDate
	vc.Terms = terms
	vc.Status = status
	vc.UpdatedAt = time.Now()

	err = s.contRepo.Update(ctx, vc)
	if err != nil {
		return nil, err
	}
	return vc, nil
}

func (s *SupplierManagementService) DeleteContract(ctx context.Context, id string) error {
	return s.contRepo.Delete(ctx, id)
}

