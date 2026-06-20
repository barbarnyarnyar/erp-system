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

type MockReceiptRepo struct {
	domain.ReceiptRepository
	createErr error
	getErr    error
	updateErr error
}

func (m *MockReceiptRepo) Create(ctx context.Context, rec *domain.Receipt) error {
	if m.createErr != nil {
		return m.createErr
	}
	return m.ReceiptRepository.Create(ctx, rec)
}

func (m *MockReceiptRepo) GetByID(ctx context.Context, id string) (*domain.Receipt, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.ReceiptRepository.GetByID(ctx, id)
}

func (m *MockReceiptRepo) Update(ctx context.Context, rec *domain.Receipt) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	return m.ReceiptRepository.Update(ctx, rec)
}

type MockShipmentRepo struct {
	domain.ShipmentRepository
	createErr error
	getErr    error
	updateErr error
}

func (m *MockShipmentRepo) Create(ctx context.Context, s *domain.Shipment) error {
	if m.createErr != nil {
		return m.createErr
	}
	return m.ShipmentRepository.Create(ctx, s)
}

func (m *MockShipmentRepo) GetByID(ctx context.Context, id string) (*domain.Shipment, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.ShipmentRepository.GetByID(ctx, id)
}

func (m *MockShipmentRepo) Update(ctx context.Context, s *domain.Shipment) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	return m.ShipmentRepository.Update(ctx, s)
}

type MockReceiptLineRepo struct {
	domain.ReceiptLineRepository
	createErr error
	listErr   error
}

func (m *MockReceiptLineRepo) Create(ctx context.Context, rl *domain.ReceiptLine) error {
	if m.createErr != nil {
		return m.createErr
	}
	return m.ReceiptLineRepository.Create(ctx, rl)
}

func (m *MockReceiptLineRepo) ListByReceiptID(ctx context.Context, receiptID string) ([]domain.ReceiptLine, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.ReceiptLineRepository.ListByReceiptID(ctx, receiptID)
}

type MockShipmentLineRepo struct {
	domain.ShipmentLineRepository
	createErr error
	listErr   error
}

func (m *MockShipmentLineRepo) Create(ctx context.Context, sl *domain.ShipmentLine) error {
	if m.createErr != nil {
		return m.createErr
	}
	return m.ShipmentLineRepository.Create(ctx, sl)
}

func (m *MockShipmentLineRepo) ListByShipmentID(ctx context.Context, shipmentID string) ([]domain.ShipmentLine, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.ShipmentLineRepository.ListByShipmentID(ctx, shipmentID)
}

func TestWarehouseService_Receipts(t *testing.T) {
	ctx := context.Background()

	setupService := func() (*WarehouseService, *memory.MemoryReceiptRepo, *memory.MemoryReceiptLineRepo, *memory.MemoryPurchaseOrderRepo, *memory.MemoryPurchaseOrderLineRepo, *InventoryService) {
		recRepo := memory.NewMemoryReceiptRepo()
		recLRepo := memory.NewMemoryReceiptLineRepo()
		shipRepo := memory.NewMemoryShipmentRepo()
		shipLRepo := memory.NewMemoryShipmentLineRepo()
		poRepo := memory.NewMemoryPurchaseOrderRepo()
		poLRepo := memory.NewMemoryPurchaseOrderLineRepo()
		invRepo := memory.NewMemoryStockBalanceRepo()
		invMovRepo := memory.NewMemoryInventoryMovementRepo()
		stRepo := memory.NewMemoryStockTransferRepo()
		tm := memory.NewMemoryTransactionManager()
		pub := &MockPublisher{}

		invSvc := NewInventoryService(invRepo, invMovRepo, stRepo, pub, tm)
		ws := NewWarehouseService(recRepo, recLRepo, shipRepo, shipLRepo, poRepo, poLRepo, invSvc, pub, tm)

		return ws, recRepo, recLRepo, poRepo, poLRepo, invSvc
	}

	t.Run("CreateReceipt with PO (Partial and Full Delivery)", func(t *testing.T) {
		ws, _, _, poRepo, poLRepo, invSvc := setupService()

		// Setup Inventory Item
		iiCreated, _ := invSvc.CreateStockBalance(ctx, "prod-1", "loc-1", decimal.NewFromInt(10))

		// Setup PO
		po := &domain.PurchaseOrder{
			ID:       "po-1",
			PoNumber: "PO-100",
			Status:   "APPROVED",
		}
		_ = poRepo.Create(ctx, po)

		pol := &domain.PurchaseOrderLine{
			ID:               "pol-1",
			PurchaseOrderID:  "po-1",
			MaterialID:       "prod-1",
			QuantityOrdered:  decimal.NewFromInt(50),
			QuantityReceived: decimal.NewFromInt(0),
		}
		_ = poLRepo.Create(ctx, pol)

		// 1. Partial delivery
		input := []ReceiptLineInput{
			{ProductID: "prod-1", QuantityReceived: 20, LocationID: "loc-1"},
		}
		res, err := ws.CreateReceipt(ctx, "po-1", "Partial notes", input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Status != "RECEIVED" {
			t.Errorf("expected RECEIVED, got %s", res.Status)
		}

		// Verify PO Status is PARTIALLY_DELIVERED
		poGot, _ := poRepo.GetByID(ctx, "po-1")
		if poGot.Status != "PARTIALLY_DELIVERED" {
			t.Errorf("expected PARTIALLY_DELIVERED, got %s", poGot.Status)
		}

		// Verify inventory incremented (10 original + 20 received = 30)
		ii, _ := invSvc.GetStockBalance(ctx, iiCreated.ID)
		if !ii.QuantityOnHand.Equal(decimal.NewFromInt(30)) {
			t.Errorf("expected 30 on hand, got %s", ii.QuantityOnHand)
		}

		// 2. Full delivery
		input2 := []ReceiptLineInput{
			{ProductID: "prod-1", QuantityReceived: 30, LocationID: "loc-1"},
		}
		res2, err := ws.CreateReceipt(ctx, "po-1", "Full delivery notes", input2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		poGot2, _ := poRepo.GetByID(ctx, "po-1")
		if poGot2.Status != "DELIVERED" {
			t.Errorf("expected DELIVERED, got %s", poGot2.Status)
		}
		if len(res2.Lines) != 1 {
			t.Errorf("expected 1 line, got %d", len(res2.Lines))
		}
	})

	t.Run("CreateReceipt without PO and Default Location", func(t *testing.T) {
		ws, _, _, _, _, invSvc := setupService()
		// Setup inventory for default location
		iiCreated, _ := invSvc.CreateStockBalance(ctx, "prod-1", "loc_default", decimal.NewFromInt(5))

		input := []ReceiptLineInput{
			{ProductID: "prod-1", QuantityReceived: 10, LocationID: ""},
		}

		res, err := ws.CreateReceipt(ctx, "", "No PO notes", input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.PurchaseOrderID != "" {
			t.Errorf("expected empty PO ID, got %s", res.PurchaseOrderID)
		}

		ii, _ := invSvc.GetStockBalance(ctx, iiCreated.ID)
		if !ii.QuantityOnHand.Equal(decimal.NewFromInt(15)) {
			t.Errorf("expected 15 on hand, got %s", ii.QuantityOnHand)
		}
	})

	t.Run("CreateReceipt - database error on receipt create", func(t *testing.T) {
		ws, recRepo, _, _, _, _ := setupService()
		mockRecRepo := &MockReceiptRepo{
			ReceiptRepository: recRepo,
			createErr:         errors.New("db insert failed"),
		}
		ws.recRepo = mockRecRepo

		input := []ReceiptLineInput{
			{ProductID: "prod-1", QuantityReceived: 10, LocationID: "loc-1"},
		}
		_, err := ws.CreateReceipt(ctx, "", "notes", input)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("CreateReceipt - database error on receipt line create", func(t *testing.T) {
		ws, _, recLRepo, _, _, _ := setupService()
		mockRecLRepo := &MockReceiptLineRepo{
			ReceiptLineRepository: recLRepo,
			createErr:             errors.New("line insert failed"),
		}
		ws.recLRepo = mockRecLRepo

		input := []ReceiptLineInput{
			{ProductID: "prod-1", QuantityReceived: 10, LocationID: "loc-1"},
		}
		_, err := ws.CreateReceipt(ctx, "", "notes", input)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("CreateReceipt - inventory adjustment fails", func(t *testing.T) {
		ws, _, _, _, _, _ := setupService()
		mockInvRepo := &MockStockBalanceRepo{
			StockBalanceRepository: memory.NewMemoryStockBalanceRepo(),
			createErr:              errors.New("adjust failed"),
		}
		ws.invService = NewInventoryService(mockInvRepo, memory.NewMemoryInventoryMovementRepo(), memory.NewMemoryStockTransferRepo(), &MockPublisher{}, memory.NewMemoryTransactionManager())

		input := []ReceiptLineInput{
			{ProductID: "prod-not-found", QuantityReceived: 10, LocationID: "loc-1"},
		}
		_, err := ws.CreateReceipt(ctx, "", "notes", input)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("List and Get Receipts", func(t *testing.T) {
		ws, _, _, _, _, invSvc := setupService()
		_, _ = invSvc.CreateStockBalance(ctx, "prod-1", "loc-default", decimal.Zero)
		input := []ReceiptLineInput{
			{ProductID: "prod-1", QuantityReceived: 5, LocationID: "loc-default"},
		}
		res, _ := ws.CreateReceipt(ctx, "", "notes", input)

		// List
		list, err := ws.ListReceipts(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("expected 1 receipt, got %d", len(list))
		}

		// Get
		got, err := ws.GetReceipt(ctx, res.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != res.ID {
			t.Errorf("expected ID %s, got %s", res.ID, got.ID)
		}
		if len(got.Lines) != 1 {
			t.Errorf("expected 1 line, got %d", len(got.Lines))
		}

		// Get Nonexistent
		_, err = ws.GetReceipt(ctx, "nonexistent")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Get Receipt Line List Error", func(t *testing.T) {
		ws, _, _, _, _, _ := setupService()
		ws.recLRepo = &MockReceiptLineRepo{
			ReceiptLineRepository: ws.recLRepo,
			listErr:               errors.New("list lines failed"),
		}
		_, err := ws.GetReceipt(ctx, "some-id")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Update Receipt", func(t *testing.T) {
		ws, _, _, _, _, _ := setupService()
		res, _ := ws.CreateReceipt(ctx, "", "notes", nil)

		updated, err := ws.UpdateReceipt(ctx, res.ID, "APPROVED", "updated notes")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Status != "APPROVED" {
			t.Errorf("unexpected updated receipt: %+v", updated)
		}

		// Update nonexistent
		_, err = ws.UpdateReceipt(ctx, "nonexistent", "APPROVED", "")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("ListReceiptLines", func(t *testing.T) {
		ws, _, _, _, _, _ := setupService()
		lines, err := ws.ListReceiptLines(ctx, "rec-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(lines) != 0 {
			t.Errorf("expected 0 lines, got %d", len(lines))
		}
	})
}

type MockStockBalanceRepo struct {
	domain.StockBalanceRepository
	createErr           error
	getErr              error
	updateErr           error
	getByMaterialLocErr error
}

func (m *MockStockBalanceRepo) Create(ctx context.Context, sb *domain.StockBalance) error {
	if m.createErr != nil {
		return m.createErr
	}
	return m.StockBalanceRepository.Create(ctx, sb)
}

func (m *MockStockBalanceRepo) Update(ctx context.Context, sb *domain.StockBalance) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	return m.StockBalanceRepository.Update(ctx, sb)
}

func TestWarehouseService_Shipments(t *testing.T) {
	ctx := context.Background()

	setupService := func() (*WarehouseService, *memory.MemoryShipmentRepo, *memory.MemoryShipmentLineRepo, *InventoryService) {
		recRepo := memory.NewMemoryReceiptRepo()
		recLRepo := memory.NewMemoryReceiptLineRepo()
		shipRepo := memory.NewMemoryShipmentRepo()
		shipLRepo := memory.NewMemoryShipmentLineRepo()
		poRepo := memory.NewMemoryPurchaseOrderRepo()
		poLRepo := memory.NewMemoryPurchaseOrderLineRepo()
		invRepo := memory.NewMemoryStockBalanceRepo()
		invMovRepo := memory.NewMemoryInventoryMovementRepo()
		stRepo := memory.NewMemoryStockTransferRepo()
		tm := memory.NewMemoryTransactionManager()
		pub := &MockPublisher{}

		invSvc := NewInventoryService(invRepo, invMovRepo, stRepo, pub, tm)
		ws := NewWarehouseService(recRepo, recLRepo, shipRepo, shipLRepo, poRepo, poLRepo, invSvc, pub, tm)

		return ws, shipRepo, shipLRepo, invSvc
	}

	t.Run("CreateShipment Success", func(t *testing.T) {
		ws, _, _, invSvc := setupService()

		// Setup Inventory
		iiCreated, _ := invSvc.CreateStockBalance(ctx, "prod-1", "loc-1", decimal.NewFromInt(50))

		lines := []ShipmentLineInput{
			{ProductID: "prod-1", QuantityShipped: 20, LocationID: "loc-1"},
		}

		res, err := ws.CreateShipment(ctx, "DHL", "TRK123", time.Now().Add(48*time.Hour), "ship notes", lines)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if res.Status != "SHIPPED" {
			t.Errorf("expected status SHIPPED, got %s", res.Status)
		}

		// Verify stock deducted (50 - 20 = 30)
		ii, _ := invSvc.GetStockBalance(ctx, iiCreated.ID)
		if !ii.QuantityOnHand.Equal(decimal.NewFromInt(30)) {
			t.Errorf("expected 30 on hand, got %s", ii.QuantityOnHand)
		}
	})

	t.Run("CreateShipment Default Location", func(t *testing.T) {
		ws, _, _, invSvc := setupService()
		iiCreated, _ := invSvc.CreateStockBalance(ctx, "prod-1", "loc_default", decimal.NewFromInt(50))

		lines := []ShipmentLineInput{
			{ProductID: "prod-1", QuantityShipped: 20, LocationID: ""},
		}

		_, err := ws.CreateShipment(ctx, "DHL", "TRK123", time.Now().Add(48*time.Hour), "ship notes", lines)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ii, _ := invSvc.GetStockBalance(ctx, iiCreated.ID)
		if !ii.QuantityOnHand.Equal(decimal.NewFromInt(30)) {
			t.Errorf("expected 30 on hand, got %s", ii.QuantityOnHand)
		}
	})

	t.Run("CreateShipment - database error on shipment create", func(t *testing.T) {
		ws, shipRepo, _, _ := setupService()
		ws.shipRepo = &MockShipmentRepo{
			ShipmentRepository: shipRepo,
			createErr:          errors.New("db create error"),
		}

		lines := []ShipmentLineInput{
			{ProductID: "prod-1", QuantityShipped: 20, LocationID: "loc-1"},
		}
		_, err := ws.CreateShipment(ctx, "DHL", "TRK123", time.Now(), "", lines)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("CreateShipment - database error on shipment line create", func(t *testing.T) {
		ws, _, shipLRepo, _ := setupService()
		ws.shipLRepo = &MockShipmentLineRepo{
			ShipmentLineRepository: shipLRepo,
			createErr:              errors.New("db create line error"),
		}

		lines := []ShipmentLineInput{
			{ProductID: "prod-1", QuantityShipped: 20, LocationID: "loc-1"},
		}
		_, err := ws.CreateShipment(ctx, "DHL", "TRK123", time.Now(), "", lines)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("CreateShipment - inventory adjustment fails", func(t *testing.T) {
		ws, _, _, _ := setupService()
		lines := []ShipmentLineInput{
			{ProductID: "prod-not-found", QuantityShipped: 20, LocationID: "loc-1"},
		}
		_, err := ws.CreateShipment(ctx, "DHL", "TRK123", time.Now(), "", lines)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("List and Get Shipments", func(t *testing.T) {
		ws, _, _, invSvc := setupService()
		_, _ = invSvc.CreateStockBalance(ctx, "prod-1", "loc-1", decimal.NewFromInt(50))
		lines := []ShipmentLineInput{
			{ProductID: "prod-1", QuantityShipped: 10, LocationID: "loc-1"},
		}
		res, _ := ws.CreateShipment(ctx, "DHL", "TRK123", time.Now(), "", lines)

		// List
		list, err := ws.ListShipments(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("expected 1 shipment, got %d", len(list))
		}

		// Get
		got, err := ws.GetShipment(ctx, res.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != res.ID {
			t.Errorf("expected ID %s, got %s", res.ID, got.ID)
		}

		// Get Nonexistent
		_, err = ws.GetShipment(ctx, "nonexistent")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Get Shipment Line List Error", func(t *testing.T) {
		ws, _, _, _ := setupService()
		ws.shipLRepo = &MockShipmentLineRepo{
			ShipmentLineRepository: ws.shipLRepo,
			listErr:                errors.New("list lines failed"),
		}
		_, err := ws.GetShipment(ctx, "some-id")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Update Shipment - Status Transitions", func(t *testing.T) {
		ws, _, _, invSvc := setupService()
		_, _ = invSvc.CreateStockBalance(ctx, "prod-1", "loc-1", decimal.NewFromInt(50))
		lines := []ShipmentLineInput{
			{ProductID: "prod-1", QuantityShipped: 10, LocationID: "loc-1"},
		}
		res, _ := ws.CreateShipment(ctx, "DHL", "TRK123", time.Now(), "", lines)

		// Transition to DELAYED
		updated, err := ws.UpdateShipment(ctx, res.ID, "DELAYED", "weather issues")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Status != "DELAYED" {
			t.Errorf("expected DELAYED, got %s", updated.Status)
		}

		// Transition to DELIVERED
		updated, err = ws.UpdateShipment(ctx, res.ID, "DELIVERED", "delivered safely")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Status != "DELIVERED" {
			t.Errorf("expected DELIVERED, got %s", updated.Status)
		}

		// Update nonexistent
		_, err = ws.UpdateShipment(ctx, "nonexistent", "DELIVERED", "")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("ListShipmentLines", func(t *testing.T) {
		ws, _, _, _ := setupService()
		lines, err := ws.ListShipmentLines(ctx, "ship-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(lines) != 0 {
			t.Errorf("expected 0 lines, got %d", len(lines))
		}
	})
}

func TestWarehouseService_TriggerTrainingRequired(t *testing.T) {
	ws := NewWarehouseService(nil, nil, nil, nil, nil, nil, nil, &MockPublisher{}, nil)
	err := ws.TriggerTrainingRequired(context.Background(), "dept-1", "Forklift safety", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
