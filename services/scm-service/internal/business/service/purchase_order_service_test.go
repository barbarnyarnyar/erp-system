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

type MockPurchaseOrderRepo struct {
	domain.PurchaseOrderRepository
	createErr error
	getErr    error
	updateErr error
	deleteErr error
}

func (m *MockPurchaseOrderRepo) Create(ctx context.Context, po *domain.PurchaseOrder) error {
	if m.createErr != nil {
		return m.createErr
	}
	return m.PurchaseOrderRepository.Create(ctx, po)
}

func (m *MockPurchaseOrderRepo) GetByID(ctx context.Context, id string) (*domain.PurchaseOrder, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.PurchaseOrderRepository.GetByID(ctx, id)
}

func (m *MockPurchaseOrderRepo) Update(ctx context.Context, po *domain.PurchaseOrder) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	return m.PurchaseOrderRepository.Update(ctx, po)
}

func (m *MockPurchaseOrderRepo) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	return m.PurchaseOrderRepository.Delete(ctx, id)
}

type MockPurchaseOrderLineRepo struct {
	domain.PurchaseOrderLineRepository
	createErr error
	listErr   error
	deleteErr error
}

func (m *MockPurchaseOrderLineRepo) Create(ctx context.Context, pol *domain.PurchaseOrderLine) error {
	if m.createErr != nil {
		return m.createErr
	}
	return m.PurchaseOrderLineRepository.Create(ctx, pol)
}

func (m *MockPurchaseOrderLineRepo) ListByPOID(ctx context.Context, poID string) ([]domain.PurchaseOrderLine, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.PurchaseOrderLineRepository.ListByPOID(ctx, poID)
}

func (m *MockPurchaseOrderLineRepo) DeleteByPOID(ctx context.Context, poID string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	return m.PurchaseOrderLineRepository.DeleteByPOID(ctx, poID)
}

type MockPurchaseRequisitionRepo struct {
	domain.PurchaseRequisitionRepository
	createErr error
	getErr    error
	updateErr error
	deleteErr error
}

func (m *MockPurchaseRequisitionRepo) Create(ctx context.Context, pr *domain.PurchaseRequisition) error {
	if m.createErr != nil {
		return m.createErr
	}
	return m.PurchaseRequisitionRepository.Create(ctx, pr)
}

func (m *MockPurchaseRequisitionRepo) GetByID(ctx context.Context, id string) (*domain.PurchaseRequisition, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.PurchaseRequisitionRepository.GetByID(ctx, id)
}

func (m *MockPurchaseRequisitionRepo) Update(ctx context.Context, pr *domain.PurchaseRequisition) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	return m.PurchaseRequisitionRepository.Update(ctx, pr)
}

func (m *MockPurchaseRequisitionRepo) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	return m.PurchaseRequisitionRepository.Delete(ctx, id)
}

type MockPurchaseRequisitionLineRepo struct {
	domain.PurchaseRequisitionLineRepository
	createErr error
	listErr   error
	deleteErr error
}

func (m *MockPurchaseRequisitionLineRepo) Create(ctx context.Context, prl *domain.PurchaseRequisitionLine) error {
	if m.createErr != nil {
		return m.createErr
	}
	return m.PurchaseRequisitionLineRepository.Create(ctx, prl)
}

func (m *MockPurchaseRequisitionLineRepo) ListByRequisitionID(ctx context.Context, reqID string) ([]domain.PurchaseRequisitionLine, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.PurchaseRequisitionLineRepository.ListByRequisitionID(ctx, reqID)
}

func (m *MockPurchaseRequisitionLineRepo) DeleteByRequisitionID(ctx context.Context, reqID string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	return m.PurchaseRequisitionLineRepository.DeleteByRequisitionID(ctx, reqID)
}

func TestPurchaseOrderService_PurchaseOrders(t *testing.T) {
	ctx := context.Background()

	setupService := func() (*PurchaseOrderService, *memory.MemoryPurchaseOrderRepo, *memory.MemoryPurchaseOrderLineRepo) {
		poRepo := memory.NewMemoryPurchaseOrderRepo()
		lineRepo := memory.NewMemoryPurchaseOrderLineRepo()
		reqRepo := memory.NewMemoryPurchaseRequisitionRepo()
		reqLineRepo := memory.NewMemoryPurchaseRequisitionLineRepo()
		pub := &MockPublisher{}
		tm := memory.NewMemoryTransactionManager()

		svc := NewPurchaseOrderService(poRepo, lineRepo, reqRepo, reqLineRepo, pub, tm)
		return svc, poRepo, lineRepo
	}

	t.Run("Create and List POs", func(t *testing.T) {
		svc, _, _ := setupService()

		lines := []POLineInput{
			{ProductID: "prod-1", QuantityOrdered: 10, UnitPrice: decimal.NewFromFloat(15.0), Description: "Line 1"},
			{ProductID: "prod-2", QuantityOrdered: 5, UnitPrice: decimal.NewFromFloat(20.0), Description: "Line 2"},
		}

		po, err := svc.CreatePurchaseOrder(ctx, "supp-1", time.Now().Add(24*time.Hour), "notes", lines)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if po.ID == "" {
			t.Error("expected PO ID")
		}
		expectedAmount := decimal.NewFromFloat(250.0) // 10*15 + 5*20 = 250
		if !po.TotalAmount.Equal(expectedAmount) {
			t.Errorf("expected total amount %s, got %s", expectedAmount, po.TotalAmount)
		}
		if len(po.Lines) != 2 {
			t.Errorf("expected 2 lines, got %d", len(po.Lines))
		}

		list, err := svc.ListPurchaseOrders(ctx)
		if err != nil {
			t.Fatalf("list error: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("expected 1 PO, got %d", len(list))
		}
	})

	t.Run("Create PO - create po error", func(t *testing.T) {
		svc, poRepo, _ := setupService()
		svc.poRepo = &MockPurchaseOrderRepo{
			PurchaseOrderRepository: poRepo,
			createErr:               errors.New("db create error"),
		}

		lines := []POLineInput{{ProductID: "prod-1", QuantityOrdered: 10, UnitPrice: decimal.NewFromFloat(10.0)}}
		_, err := svc.CreatePurchaseOrder(ctx, "supp-1", time.Now(), "", lines)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Create PO - create line error", func(t *testing.T) {
		svc, _, lineRepo := setupService()
		svc.lineRepo = &MockPurchaseOrderLineRepo{
			PurchaseOrderLineRepository: lineRepo,
			createErr:                   errors.New("db line error"),
		}

		lines := []POLineInput{{ProductID: "prod-1", QuantityOrdered: 10, UnitPrice: decimal.NewFromFloat(10.0)}}
		_, err := svc.CreatePurchaseOrder(ctx, "supp-1", time.Now(), "", lines)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Get PO", func(t *testing.T) {
		svc, _, _ := setupService()
		lines := []POLineInput{{ProductID: "p-1", QuantityOrdered: 2, UnitPrice: decimal.NewFromFloat(5.0)}}
		res, _ := svc.CreatePurchaseOrder(ctx, "supp-1", time.Now(), "", lines)

		got, err := svc.GetPurchaseOrder(ctx, res.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != res.ID {
			t.Errorf("expected ID %s, got %s", res.ID, got.ID)
		}

		_, err = svc.GetPurchaseOrder(ctx, "nonexistent")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Get PO - lines list error", func(t *testing.T) {
		svc, _, lineRepo := setupService()
		svc.lineRepo = &MockPurchaseOrderLineRepo{
			PurchaseOrderLineRepository: lineRepo,
			listErr:                     errors.New("db list error"),
		}
		_, err := svc.GetPurchaseOrder(ctx, "some-id")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Update PO Status Transitions", func(t *testing.T) {
		svc, _, _ := setupService()
		res, _ := svc.CreatePurchaseOrder(ctx, "supp-1", time.Now(), "", nil)

		// Transition to DELIVERED
		updated, err := svc.UpdatePurchaseOrder(ctx, res.ID, time.Now(), "DELIVERED", "notes")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Status != "DELIVERED" {
			t.Errorf("expected status DELIVERED, got %s", updated.Status)
		}

		// Transition to CANCELLED
		updated, err = svc.UpdatePurchaseOrder(ctx, res.ID, time.Now(), "CANCELLED", "cancel reason")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Status != "CANCELLED" {
			t.Errorf("expected status CANCELLED, got %s", updated.Status)
		}

		// Nonexistent
		_, err = svc.UpdatePurchaseOrder(ctx, "nonexistent", time.Now(), "DELIVERED", "")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Update PO - update error", func(t *testing.T) {
		svc, poRepo, _ := setupService()
		res, _ := svc.CreatePurchaseOrder(ctx, "supp-1", time.Now(), "", nil)

		svc.poRepo = &MockPurchaseOrderRepo{
			PurchaseOrderRepository: poRepo,
			updateErr:               errors.New("db update error"),
		}
		_, err := svc.UpdatePurchaseOrder(ctx, res.ID, time.Now(), "DELIVERED", "")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Delete PO", func(t *testing.T) {
		svc, _, _ := setupService()
		res, _ := svc.CreatePurchaseOrder(ctx, "supp-1", time.Now(), "", nil)

		err := svc.DeletePurchaseOrder(ctx, res.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = svc.GetPurchaseOrder(ctx, res.ID)
		if err == nil {
			t.Error("expected error for deleted PO")
		}
	})

	t.Run("Delete PO - delete lines error", func(t *testing.T) {
		svc, _, lineRepo := setupService()
		svc.lineRepo = &MockPurchaseOrderLineRepo{
			PurchaseOrderLineRepository: lineRepo,
			deleteErr:                   errors.New("delete line error"),
		}
		err := svc.DeletePurchaseOrder(ctx, "some-id")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Send PO Success", func(t *testing.T) {
		svc, _, _ := setupService()
		res, _ := svc.CreatePurchaseOrder(ctx, "supp-1", time.Now(), "", nil)

		sent, err := svc.SendPurchaseOrder(ctx, res.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sent.Status != "SUBMITTED" {
			t.Errorf("expected status SUBMITTED, got %s", sent.Status)
		}
	})

	t.Run("Send PO - nonexistent", func(t *testing.T) {
		svc, _, _ := setupService()
		_, err := svc.SendPurchaseOrder(ctx, "nonexistent")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Send PO - update error", func(t *testing.T) {
		svc, poRepo, _ := setupService()
		res, _ := svc.CreatePurchaseOrder(ctx, "supp-1", time.Now(), "", nil)

		svc.poRepo = &MockPurchaseOrderRepo{
			PurchaseOrderRepository: poRepo,
			updateErr:               errors.New("db update error"),
		}
		_, err := svc.SendPurchaseOrder(ctx, res.ID)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("ListPurchaseOrderLines", func(t *testing.T) {
		svc, _, _ := setupService()
		lines, err := svc.ListPurchaseOrderLines(ctx, "po-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(lines) != 0 {
			t.Errorf("expected 0 lines, got %d", len(lines))
		}
	})
}

func TestPurchaseOrderService_PurchaseRequisitions(t *testing.T) {
	ctx := context.Background()

	setupService := func() (*PurchaseOrderService, *memory.MemoryPurchaseRequisitionRepo, *memory.MemoryPurchaseRequisitionLineRepo) {
		poRepo := memory.NewMemoryPurchaseOrderRepo()
		lineRepo := memory.NewMemoryPurchaseOrderLineRepo()
		reqRepo := memory.NewMemoryPurchaseRequisitionRepo()
		reqLineRepo := memory.NewMemoryPurchaseRequisitionLineRepo()
		pub := &MockPublisher{}
		tm := memory.NewMemoryTransactionManager()

		svc := NewPurchaseOrderService(poRepo, lineRepo, reqRepo, reqLineRepo, pub, tm)
		return svc, reqRepo, reqLineRepo
	}

	t.Run("Create and List Requisitions", func(t *testing.T) {
		svc, _, _ := setupService()

		lines := []RequisitionLineInput{
			{ProductID: "prod-1", QuantityRequested: 10, EstimatedUnitPrice: decimal.NewFromFloat(15.0)},
		}

		pr, err := svc.CreatePurchaseRequisition(ctx, "req-1", time.Now(), "notes", lines)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if pr.ID == "" {
			t.Error("expected generated ID")
		}
		if pr.Status != "DRAFT" {
			t.Errorf("expected status DRAFT, got %s", pr.Status)
		}

		list, err := svc.ListPurchaseRequisitions(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("expected 1 requisition, got %d", len(list))
		}
	})

	t.Run("Create Requisition - repo error", func(t *testing.T) {
		svc, reqRepo, _ := setupService()
		svc.reqRepo = &MockPurchaseRequisitionRepo{
			PurchaseRequisitionRepository: reqRepo,
			createErr:                     errors.New("db error"),
		}

		lines := []RequisitionLineInput{{ProductID: "prod-1", QuantityRequested: 10}}
		_, err := svc.CreatePurchaseRequisition(ctx, "req-1", time.Now(), "", lines)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Create Requisition - line repo error", func(t *testing.T) {
		svc, _, reqLineRepo := setupService()
		svc.reqLineRepo = &MockPurchaseRequisitionLineRepo{
			PurchaseRequisitionLineRepository: reqLineRepo,
			createErr:                         errors.New("db line error"),
		}

		lines := []RequisitionLineInput{{ProductID: "prod-1", QuantityRequested: 10}}
		_, err := svc.CreatePurchaseRequisition(ctx, "req-1", time.Now(), "", lines)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Get Requisition", func(t *testing.T) {
		svc, _, _ := setupService()
		res, _ := svc.CreatePurchaseRequisition(ctx, "req-1", time.Now(), "", nil)

		got, err := svc.GetPurchaseRequisition(ctx, res.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != res.ID {
			t.Errorf("expected ID %s, got %s", res.ID, got.ID)
		}

		_, err = svc.GetPurchaseRequisition(ctx, "nonexistent")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Get Requisition - line error", func(t *testing.T) {
		svc, _, reqLineRepo := setupService()
		svc.reqLineRepo = &MockPurchaseRequisitionLineRepo{
			PurchaseRequisitionLineRepository: reqLineRepo,
			listErr:                           errors.New("list error"),
		}
		_, err := svc.GetPurchaseRequisition(ctx, "some-id")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Update Requisition", func(t *testing.T) {
		svc, _, _ := setupService()
		res, _ := svc.CreatePurchaseRequisition(ctx, "req-1", time.Now(), "", nil)

		updated, err := svc.UpdatePurchaseRequisition(ctx, res.ID, time.Now(), "SUBMITTED", "new notes")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Status != "SUBMITTED" || updated.Notes != "new notes" {
			t.Errorf("unexpected updated requisition: %+v", updated)
		}

		// Nonexistent
		_, err = svc.UpdatePurchaseRequisition(ctx, "nonexistent", time.Now(), "APPROVED", "")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Update Requisition - update error", func(t *testing.T) {
		svc, reqRepo, _ := setupService()
		res, _ := svc.CreatePurchaseRequisition(ctx, "req-1", time.Now(), "", nil)

		svc.reqRepo = &MockPurchaseRequisitionRepo{
			PurchaseRequisitionRepository: reqRepo,
			updateErr:                     errors.New("db update error"),
		}
		_, err := svc.UpdatePurchaseRequisition(ctx, res.ID, time.Now(), "SUBMITTED", "")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Delete Requisition", func(t *testing.T) {
		svc, _, _ := setupService()
		res, _ := svc.CreatePurchaseRequisition(ctx, "req-1", time.Now(), "", nil)

		err := svc.DeletePurchaseRequisition(ctx, res.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = svc.GetPurchaseRequisition(ctx, res.ID)
		if err == nil {
			t.Error("expected error for deleted requisition")
		}
	})

	t.Run("Delete Requisition - lines delete error", func(t *testing.T) {
		svc, _, reqLineRepo := setupService()
		svc.reqLineRepo = &MockPurchaseRequisitionLineRepo{
			PurchaseRequisitionLineRepository: reqLineRepo,
			deleteErr:                         errors.New("delete error"),
		}
		err := svc.DeletePurchaseRequisition(ctx, "some-id")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Approve Requisition", func(t *testing.T) {
		svc, _, _ := setupService()
		res, _ := svc.CreatePurchaseRequisition(ctx, "req-1", time.Now(), "", nil)

		approved, err := svc.ApprovePurchaseRequisition(ctx, res.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if approved.Status != "APPROVED" {
			t.Errorf("expected APPROVED status, got %s", approved.Status)
		}

		// Nonexistent
		_, err = svc.ApprovePurchaseRequisition(ctx, "nonexistent")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Approve Requisition - update error", func(t *testing.T) {
		svc, reqRepo, _ := setupService()
		res, _ := svc.CreatePurchaseRequisition(ctx, "req-1", time.Now(), "", nil)

		svc.reqRepo = &MockPurchaseRequisitionRepo{
			PurchaseRequisitionRepository: reqRepo,
			updateErr:                     errors.New("db error"),
		}
		_, err := svc.ApprovePurchaseRequisition(ctx, res.ID)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Reject Requisition", func(t *testing.T) {
		svc, _, _ := setupService()
		res, _ := svc.CreatePurchaseRequisition(ctx, "req-1", time.Now(), "", nil)

		rejected, err := svc.RejectPurchaseRequisition(ctx, res.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rejected.Status != "REJECTED" {
			t.Errorf("expected REJECTED status, got %s", rejected.Status)
		}

		// Nonexistent
		_, err = svc.RejectPurchaseRequisition(ctx, "nonexistent")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Reject Requisition - update error", func(t *testing.T) {
		svc, reqRepo, _ := setupService()
		res, _ := svc.CreatePurchaseRequisition(ctx, "req-1", time.Now(), "", nil)

		svc.reqRepo = &MockPurchaseRequisitionRepo{
			PurchaseRequisitionRepository: reqRepo,
			updateErr:                     errors.New("db error"),
		}
		_, err := svc.RejectPurchaseRequisition(ctx, res.ID)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("ListPurchaseRequisitionLines", func(t *testing.T) {
		svc, _, _ := setupService()
		lines, err := svc.ListPurchaseRequisitionLines(ctx, "req-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(lines) != 0 {
			t.Errorf("expected 0 lines, got %d", len(lines))
		}
	})
}
