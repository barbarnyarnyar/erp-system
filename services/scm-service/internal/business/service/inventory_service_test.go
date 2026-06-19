package service

import (
	"context"
	"errors"
	"testing"

	"github.com/erp-system/scm-service/internal/business/domain"
	"github.com/erp-system/scm-service/internal/data/memory"
	"github.com/shopspring/decimal"
)

type MockInventoryItemRepoEx struct {
	domain.InventoryItemRepository
	createErr           error
	getErr              error
	updateErr           error
	getByProductLocErr error
}

func (m *MockInventoryItemRepoEx) Create(ctx context.Context, ii *domain.InventoryItem) error {
	if m.createErr != nil {
		return m.createErr
	}
	return m.InventoryItemRepository.Create(ctx, ii)
}

func (m *MockInventoryItemRepoEx) GetByID(ctx context.Context, id string) (*domain.InventoryItem, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.InventoryItemRepository.GetByID(ctx, id)
}

func (m *MockInventoryItemRepoEx) Update(ctx context.Context, ii *domain.InventoryItem) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	return m.InventoryItemRepository.Update(ctx, ii)
}

func (m *MockInventoryItemRepoEx) GetByProductAndLocation(ctx context.Context, productID string, locationID string) (*domain.InventoryItem, error) {
	if m.getByProductLocErr != nil {
		return nil, m.getByProductLocErr
	}
	return m.InventoryItemRepository.GetByProductAndLocation(ctx, productID, locationID)
}

type MockInventoryMovementRepoEx struct {
	domain.InventoryMovementRepository
	createErr error
}

func (m *MockInventoryMovementRepoEx) Create(ctx context.Context, im *domain.InventoryMovement) error {
	if m.createErr != nil {
		return m.createErr
	}
	return m.InventoryMovementRepository.Create(ctx, im)
}

type MockStockTransferRepoEx struct {
	domain.StockTransferRepository
	createErr error
	getErr    error
	updateErr error
}

func (m *MockStockTransferRepoEx) Create(ctx context.Context, st *domain.StockTransfer) error {
	if m.createErr != nil {
		return m.createErr
	}
	return m.StockTransferRepository.Create(ctx, st)
}

func (m *MockStockTransferRepoEx) GetByID(ctx context.Context, id string) (*domain.StockTransfer, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.StockTransferRepository.GetByID(ctx, id)
}

func (m *MockStockTransferRepoEx) Update(ctx context.Context, st *domain.StockTransfer) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	return m.StockTransferRepository.Update(ctx, st)
}

func TestInventoryService_ListInventory(t *testing.T) {
	ctx := context.Background()
	svc := newInventoryService(t)

	_, _ = svc.CreateInventoryItem(ctx, "prod-1", "loc-1", 10, 5, 100, decimal.NewFromFloat(5.0))
	list, err := svc.ListInventory(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 item, got %d", len(list))
	}
}

func TestInventoryService_CreateInventoryItemInvariantFailure(t *testing.T) {
	ctx := context.Background()
	svc := newInventoryService(t)

	// Invariant violation: negative qtyOnHand
	_, err := svc.CreateInventoryItem(ctx, "prod-1", "loc-1", -5, 5, 100, decimal.NewFromFloat(5.0))
	if err == nil {
		t.Error("expected invariant error, got nil")
	}
}

func TestInventoryService_UpdateInventoryItem(t *testing.T) {
	ctx := context.Background()
	svc := newInventoryService(t)

	ii, err := svc.CreateInventoryItem(ctx, "prod-1", "loc-1", 10, 5, 100, decimal.NewFromFloat(5.0))
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	t.Run("success", func(t *testing.T) {
		updated, err := svc.UpdateInventoryItem(ctx, ii.ID, 20, 2, 8, 200, decimal.NewFromFloat(6.0))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.QuantityOnHand != 20 || updated.QuantityReserved != 2 || updated.QuantityAvailable != 18 {
			t.Errorf("unexpected updated state: %+v", updated)
		}
	})

	t.Run("nonexistent", func(t *testing.T) {
		_, err := svc.UpdateInventoryItem(ctx, "nonexistent", 20, 0, 5, 100, decimal.Zero)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("invariant failure", func(t *testing.T) {
		_, err := svc.UpdateInventoryItem(ctx, ii.ID, 10, 20, 5, 100, decimal.Zero) // reserved > on hand
		if err == nil {
			t.Error("expected invariant error, got nil")
		}
	})
}

func TestInventoryService_AdjustInventory_Branches(t *testing.T) {
	ctx := context.Background()

	t.Run("insufficient stock for issue", func(t *testing.T) {
		svc := newInventoryService(t)
		_, _ = svc.CreateInventoryItem(ctx, "prod-1", "loc-1", 10, 5, 100, decimal.NewFromFloat(5.0))

		_, err := svc.AdjustInventory(ctx, "prod-1", "loc-1", 15, "ISSUE", "")
		if err == nil {
			t.Error("expected error due to insufficient stock, got nil")
		}
	})

	t.Run("unknown movement type", func(t *testing.T) {
		svc := newInventoryService(t)
		_, _ = svc.CreateInventoryItem(ctx, "prod-1", "loc-1", 10, 5, 100, decimal.NewFromFloat(5.0))

		_, err := svc.AdjustInventory(ctx, "prod-1", "loc-1", 5, "UNKNOWN", "")
		if err == nil {
			t.Error("expected error for unknown movement type, got nil")
		}
	})

	t.Run("low stock trigger event", func(t *testing.T) {
		svc := newInventoryService(t)
		// reorderPoint = 15, onHand = 20
		_, _ = svc.CreateInventoryItem(ctx, "prod-1", "loc-1", 20, 15, 100, decimal.NewFromFloat(5.0))

		// Adjust sub 8 -> onHand = 12 (which is < 15, triggering low stock)
		_, err := svc.AdjustInventory(ctx, "prod-1", "loc-1", 8, "ADJUSTMENT_SUB", "low stock check")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("out of stock trigger event", func(t *testing.T) {
		svc := newInventoryService(t)
		_, _ = svc.CreateInventoryItem(ctx, "prod-1", "loc-1", 10, 5, 100, decimal.NewFromFloat(5.0))

		_, err := svc.AdjustInventory(ctx, "prod-1", "loc-1", 10, "ISSUE", "clear stock")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("db error inside transaction", func(t *testing.T) {
		invRepo := &MockInventoryItemRepoEx{
			InventoryItemRepository: memory.NewMemoryInventoryItemRepo(),
			updateErr:               errors.New("db update failed"),
		}
		svc := NewInventoryService(invRepo, memory.NewMemoryInventoryMovementRepo(), memory.NewMemoryStockTransferRepo(), &MockPublisher{}, memory.NewMemoryTransactionManager())
		_, _ = svc.CreateInventoryItem(ctx, "prod-1", "loc-1", 10, 5, 100, decimal.NewFromFloat(5.0))

		_, err := svc.AdjustInventory(ctx, "prod-1", "loc-1", 5, "RECEIPT", "")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestInventoryService_ReserveStock_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("nonexistent item", func(t *testing.T) {
		svc := newInventoryService(t)
		err := svc.ReserveStock(ctx, "nonexistent", "loc-1", 5, "ref-1")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("insufficient available", func(t *testing.T) {
		svc := newInventoryService(t)
		_, _ = svc.CreateInventoryItem(ctx, "prod-1", "loc-1", 10, 5, 100, decimal.NewFromFloat(5.0))

		err := svc.ReserveStock(ctx, "prod-1", "loc-1", 15, "ref-1")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestInventoryService_ReleaseReservation_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("reservation not found", func(t *testing.T) {
		svc := newInventoryService(t)
		err := svc.ReleaseReservation(ctx, "nonexistent-ref")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("inventory item not found for released reservation", func(t *testing.T) {
		invRepo := memory.NewMemoryInventoryItemRepo()
		svc := NewInventoryService(invRepo, memory.NewMemoryInventoryMovementRepo(), memory.NewMemoryStockTransferRepo(), &MockPublisher{}, memory.NewMemoryTransactionManager())

		// Create reservation manually to bypass checks
		svc.mu.Lock()
		svc.reservations["ref-invalid"] = stockReservation{
			productID:  "prod-1",
			locationID: "loc-1",
			quantity:   5,
		}
		svc.mu.Unlock()

		err := svc.ReleaseReservation(ctx, "ref-invalid")
		if err == nil {
			t.Error("expected error because inventory item does not exist, got nil")
		}
	})
}

func TestInventoryService_CreateStockTransfer_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("quantity <= 0", func(t *testing.T) {
		svc := newInventoryService(t)
		_, err := svc.CreateStockTransfer(ctx, "loc-1", "loc-2", "prod-1", 0)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("source inventory not found", func(t *testing.T) {
		svc := newInventoryService(t)
		_, err := svc.CreateStockTransfer(ctx, "loc-1", "loc-2", "prod-1", 5)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("insufficient source inventory available", func(t *testing.T) {
		svc := newInventoryService(t)
		_, _ = svc.CreateInventoryItem(ctx, "prod-1", "loc-1", 3, 0, 100, decimal.Zero)

		_, err := svc.CreateStockTransfer(ctx, "loc-1", "loc-2", "prod-1", 5)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("db create error", func(t *testing.T) {
		invRepo := memory.NewMemoryInventoryItemRepo()
		transferRepo := &MockStockTransferRepoEx{
			StockTransferRepository: memory.NewMemoryStockTransferRepo(),
			createErr:               errors.New("db create error"),
		}
		svc := NewInventoryService(invRepo, memory.NewMemoryInventoryMovementRepo(), transferRepo, &MockPublisher{}, memory.NewMemoryTransactionManager())
		_, _ = svc.CreateInventoryItem(ctx, "prod-1", "loc-1", 10, 0, 100, decimal.Zero)

		_, err := svc.CreateStockTransfer(ctx, "loc-1", "loc-2", "prod-1", 5)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestInventoryService_ExecuteStockTransfer_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("not pending", func(t *testing.T) {
		svc := newInventoryService(t)
		_, _ = svc.CreateInventoryItem(ctx, "prod-1", "loc-1", 10, 0, 100, decimal.Zero)
		st, _ := svc.CreateStockTransfer(ctx, "loc-1", "loc-2", "prod-1", 5)

		// Set status to TRANSFERRED first
		st.Status = "TRANSFERRED"
		_ = svc.transferRepo.Update(ctx, st)

		_, err := svc.ExecuteStockTransfer(ctx, st.ID)
		if err == nil {
			t.Error("expected error for already transferred, got nil")
		}
	})

	t.Run("destination item auto-creation success", func(t *testing.T) {
		svc := newInventoryService(t)
		_, _ = svc.CreateInventoryItem(ctx, "prod-1", "loc-1", 10, 0, 100, decimal.NewFromFloat(5.0))
		// Destination location loc-2 doesn't have inventory seeded, should auto-create it

		st, err := svc.CreateStockTransfer(ctx, "loc-1", "loc-2", "prod-1", 5)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		executed, err := svc.ExecuteStockTransfer(ctx, st.ID)
		if err != nil {
			t.Fatalf("unexpected execute error: %v", err)
		}
		if executed.Status != "TRANSFERRED" {
			t.Errorf("expected TRANSFERRED, got %s", executed.Status)
		}
	})
}

func TestInventoryService_ListMovementsAndGetStockTransfer(t *testing.T) {
	ctx := context.Background()
	svc := newInventoryService(t)

	_, _ = svc.CreateInventoryItem(ctx, "prod-1", "loc-1", 10, 0, 100, decimal.Zero)
	st, _ := svc.CreateStockTransfer(ctx, "loc-1", "loc-2", "prod-1", 5)

	got, err := svc.GetStockTransfer(ctx, st.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != st.ID {
		t.Errorf("expected ID %s, got %s", st.ID, got.ID)
	}

	transfers, err := svc.ListStockTransfers(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(transfers) != 1 {
		t.Errorf("expected 1 transfer, got %d", len(transfers))
	}

	movements, err := svc.ListMovements(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(movements) != 0 {
		t.Errorf("expected 0 movements manually logged, got %d", len(movements))
	}
}
