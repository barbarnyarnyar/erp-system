package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/erp-system/scm-service/internal/business/domain"
	"github.com/erp-system/scm-service/internal/data/memory"
	"github.com/shopspring/decimal"
)

type MockSupplierRepo struct {
	domain.SupplierRepository
	createErr error
	getErr    error
	updateErr error
}

func (m *MockSupplierRepo) Create(ctx context.Context, s *domain.Supplier) error {
	if m.createErr != nil {
		return m.createErr
	}
	return m.SupplierRepository.Create(ctx, s)
}

func (m *MockSupplierRepo) GetByID(ctx context.Context, id string) (*domain.Supplier, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.SupplierRepository.GetByID(ctx, id)
}

func (m *MockSupplierRepo) Update(ctx context.Context, s *domain.Supplier) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	return m.SupplierRepository.Update(ctx, s)
}

type MockVendorContractRepo struct {
	domain.VendorContractRepository
	createErr error
	getErr    error
	updateErr error
}

func (m *MockVendorContractRepo) Create(ctx context.Context, vc *domain.VendorContract) error {
	if m.createErr != nil {
		return m.createErr
	}
	return m.VendorContractRepository.Create(ctx, vc)
}

func (m *MockVendorContractRepo) GetByID(ctx context.Context, id string) (*domain.VendorContract, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.VendorContractRepository.GetByID(ctx, id)
}

func (m *MockVendorContractRepo) Update(ctx context.Context, vc *domain.VendorContract) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	return m.VendorContractRepository.Update(ctx, vc)
}

func TestSupplierManagementService_Suppliers(t *testing.T) {
	ctx := context.Background()

	t.Run("Create and List Suppliers", func(t *testing.T) {
		repo := memory.NewMemorySupplierRepo()
		contRepo := memory.NewMemoryVendorContractRepo()
		pub := &MockPublisher{}
		svc := NewSupplierManagementService(repo, contRepo, pub)

		s, err := svc.CreateSupplier(ctx, "SUP001", "Supplier 1", "John", "john@test.com", "12345")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s.ID == "" {
			t.Error("expected generated ID")
		}

		list, err := svc.ListSuppliers(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("expected 1 supplier, got %d", len(list))
		}
	})

	t.Run("Create Supplier Repo Error", func(t *testing.T) {
		repo := &MockSupplierRepo{
			SupplierRepository: memory.NewMemorySupplierRepo(),
			createErr:          errors.New("db create error"),
		}
		svc := NewSupplierManagementService(repo, nil, &MockPublisher{})
		_, err := svc.CreateSupplier(ctx, "SUP001", "Supplier 1", "John", "john@test.com", "12345")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Create Supplier Publisher Error (logs but succeeds)", func(t *testing.T) {
		repo := memory.NewMemorySupplierRepo()
		pub := &MockPublisher{
			PublishFunc: func(ctx context.Context, topic string, key string, event interface{}) error {
				return errors.New("pub error")
			},
		}
		svc := NewSupplierManagementService(repo, nil, pub)
		_, err := svc.CreateSupplier(ctx, "SUP001", "Supplier 1", "John", "john@test.com", "12345")
		if err != nil {
			t.Fatalf("expected success, got err: %v", err)
		}
	})

	t.Run("Get Supplier", func(t *testing.T) {
		repo := memory.NewMemorySupplierRepo()
		svc := NewSupplierManagementService(repo, nil, &MockPublisher{})
		s, _ := svc.CreateSupplier(ctx, "SUP001", "Supplier 1", "John", "john@test.com", "12345")

		got, err := svc.GetSupplier(ctx, s.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.SupplierName != "Supplier 1" {
			t.Errorf("expected Supplier 1, got %s", got.SupplierName)
		}

		_, err = svc.GetSupplier(ctx, "nonexistent")
		if err == nil {
			t.Error("expected error for nonexistent supplier, got nil")
		}
	})

	t.Run("Update Supplier Success", func(t *testing.T) {
		repo := memory.NewMemorySupplierRepo()
		svc := NewSupplierManagementService(repo, nil, &MockPublisher{})
		s, _ := svc.CreateSupplier(ctx, "SUP001", "Supplier 1", "John", "john@test.com", "12345")

		updated, err := svc.UpdateSupplier(ctx, s.ID, "SUP001-Updated", "Supplier 1 Updated", "Jane", "jane@test.com", "67890", false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.SupplierName != "Supplier 1 Updated" || updated.IsActive != false {
			t.Errorf("unexpected updated values: %+v", updated)
		}
	})

	t.Run("Update Supplier Get Error", func(t *testing.T) {
		repo := &MockSupplierRepo{
			SupplierRepository: memory.NewMemorySupplierRepo(),
			getErr:             errors.New("not found"),
		}
		svc := NewSupplierManagementService(repo, nil, &MockPublisher{})
		_, err := svc.UpdateSupplier(ctx, "nonexistent", "SUP001", "Supplier 1", "John", "john@test.com", "12345", true)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Update Supplier Update Error", func(t *testing.T) {
		repo := memory.NewMemorySupplierRepo()
		mockRepo := &MockSupplierRepo{
			SupplierRepository: repo,
			updateErr:          errors.New("db update error"),
		}
		svc := NewSupplierManagementService(mockRepo, nil, &MockPublisher{})
		// Seed
		seedSvc := NewSupplierManagementService(repo, nil, &MockPublisher{})
		s, _ := seedSvc.CreateSupplier(ctx, "SUP001", "Supplier 1", "John", "john@test.com", "12345")

		_, err := svc.UpdateSupplier(ctx, s.ID, "SUP001", "Supplier 1", "John", "john@test.com", "12345", true)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Delete Supplier", func(t *testing.T) {
		repo := memory.NewMemorySupplierRepo()
		svc := NewSupplierManagementService(repo, nil, &MockPublisher{})
		s, _ := svc.CreateSupplier(ctx, "SUP001", "Supplier 1", "John", "john@test.com", "12345")

		err := svc.DeleteSupplier(ctx, s.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = svc.GetSupplier(ctx, s.ID)
		if err == nil {
			t.Error("expected error for deleted supplier")
		}
	})

	t.Run("Evaluate Performance", func(t *testing.T) {
		svc := NewSupplierManagementService(nil, nil, &MockPublisher{})
		err := svc.EvaluatePerformance(ctx, "supp-1", decimal.NewFromFloat(0.95), decimal.NewFromFloat(15000.0), decimal.NewFromFloat(4.5))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("Evaluate Performance Publisher Error (logs but succeeds)", func(t *testing.T) {
		pub := &MockPublisher{
			PublishFunc: func(ctx context.Context, topic string, key string, event interface{}) error {
				return errors.New("pub error")
			},
		}
		svc := NewSupplierManagementService(nil, nil, pub)
		err := svc.EvaluatePerformance(ctx, "supp-1", decimal.NewFromFloat(0.95), decimal.NewFromFloat(15000.0), decimal.NewFromFloat(4.5))
		if err != nil {
			t.Fatalf("expected success, got err: %v", err)
		}
	})
}

func TestSupplierManagementService_Contracts(t *testing.T) {
	ctx := context.Background()

	t.Run("Create and List Contracts", func(t *testing.T) {
		contRepo := memory.NewMemoryVendorContractRepo()
		svc := NewSupplierManagementService(nil, contRepo, &MockPublisher{})

		startDate := time.Now()
		endDate := startDate.Add(365 * 24 * time.Hour)

		vc, err := svc.CreateContract(ctx, "CON001", "supp-1", startDate, endDate, "Net 30")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if vc.ID == "" {
			t.Error("expected generated ID")
		}

		list, err := svc.ListContracts(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("expected 1 contract, got %d", len(list))
		}
	})

	t.Run("Create Contract Error", func(t *testing.T) {
		contRepo := &MockVendorContractRepo{
			VendorContractRepository: memory.NewMemoryVendorContractRepo(),
			createErr:                errors.New("db create error"),
		}
		svc := NewSupplierManagementService(nil, contRepo, &MockPublisher{})
		_, err := svc.CreateContract(ctx, "CON001", "supp-1", time.Now(), time.Now(), "Net 30")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Get Contract", func(t *testing.T) {
		contRepo := memory.NewMemoryVendorContractRepo()
		svc := NewSupplierManagementService(nil, contRepo, &MockPublisher{})
		vc, _ := svc.CreateContract(ctx, "CON001", "supp-1", time.Now(), time.Now(), "Net 30")

		got, err := svc.GetContract(ctx, vc.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ContractNumber != "CON001" {
			t.Errorf("expected CON001, got %s", got.ContractNumber)
		}

		_, err = svc.GetContract(ctx, "nonexistent")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Update Contract Success", func(t *testing.T) {
		contRepo := memory.NewMemoryVendorContractRepo()
		svc := NewSupplierManagementService(nil, contRepo, &MockPublisher{})
		vc, _ := svc.CreateContract(ctx, "CON001", "supp-1", time.Now(), time.Now(), "Net 30")

		updated, err := svc.UpdateContract(ctx, vc.ID, "CON001-Updated", "supp-2", time.Now(), time.Now(), "Net 60", "EXPIRED")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.ContractNumber != "CON001-Updated" || updated.Status != "EXPIRED" {
			t.Errorf("unexpected updated values: %+v", updated)
		}
	})

	t.Run("Update Contract Get Error", func(t *testing.T) {
		contRepo := &MockVendorContractRepo{
			VendorContractRepository: memory.NewMemoryVendorContractRepo(),
			getErr:                   errors.New("not found"),
		}
		svc := NewSupplierManagementService(nil, contRepo, &MockPublisher{})
		_, err := svc.UpdateContract(ctx, "nonexistent", "CON001", "supp-1", time.Now(), time.Now(), "Net 30", "ACTIVE")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Update Contract Update Error", func(t *testing.T) {
		contRepo := memory.NewMemoryVendorContractRepo()
		mockContRepo := &MockVendorContractRepo{
			VendorContractRepository: contRepo,
			updateErr:                errors.New("db update error"),
		}
		svc := NewSupplierManagementService(nil, mockContRepo, &MockPublisher{})
		// Seed
		seedSvc := NewSupplierManagementService(nil, contRepo, &MockPublisher{})
		vc, _ := seedSvc.CreateContract(ctx, "CON001", "supp-1", time.Now(), time.Now(), "Net 30")

		_, err := svc.UpdateContract(ctx, vc.ID, "CON001", "supp-1", time.Now(), time.Now(), "Net 30", "ACTIVE")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Delete Contract", func(t *testing.T) {
		contRepo := memory.NewMemoryVendorContractRepo()
		svc := NewSupplierManagementService(nil, contRepo, &MockPublisher{})
		vc, _ := svc.CreateContract(ctx, "CON001", "supp-1", time.Now(), time.Now(), "Net 30")

		err := svc.DeleteContract(ctx, vc.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = svc.GetContract(ctx, vc.ID)
		if err == nil {
			t.Error("expected error for deleted contract")
		}
	})
}
