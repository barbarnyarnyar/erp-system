package service

import (
	sharedtesting "erp-system/shared/testing"
	"context"
	"testing"

	"github.com/erp-system/scm-service/internal/business/domain"
	"github.com/erp-system/scm-service/internal/data/memory"
	"github.com/shopspring/decimal"
)

func newInventoryService(t *testing.T) *InventoryService {
	t.Helper()
	return NewInventoryService(
		memory.NewMemoryInventoryItemRepo(),
		memory.NewMemoryInventoryMovementRepo(),
		memory.NewMemoryStockTransferRepo(),
		&sharedtesting.MockPublisher{},
	)
}

// TestAdjustInventory_MaintainsInvariant_WithReservations is the regression
// test for the bug fixed in Phase S4.5: previously AdjustInventory mutated
// both QuantityOnHand and QuantityAvailable by the same delta, which broke
// the invariant `available = on_hand - reserved` whenever `reserved > 0`.
func TestAdjustInventory_MaintainsInvariant_WithReservations(t *testing.T) {
	svc := newInventoryService(t)
	ctx := context.Background()

	ii, err := svc.CreateInventoryItem(ctx, "prod_1", "loc_1", 100, 10, 1000, decimal.NewFromInt(5))
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Pre-reserve 30 units.
	if err := svc.ReserveStock(ctx, ii.ProductID, ii.LocationID, 30, "ref_1"); err != nil {
		t.Fatalf("reserve: %v", err)
	}

	// Receive 50 more.
	ii, err = svc.AdjustInventory(ctx, ii.ProductID, ii.LocationID, 50, "RECEIPT", "test receipt")
	if err != nil {
		t.Fatalf("adjust (receipt): %v", err)
	}
	if ii.QuantityOnHand != 150 {
		t.Errorf("expected on_hand=150, got %d", ii.QuantityOnHand)
	}
	if ii.QuantityReserved != 30 {
		t.Errorf("expected reserved=30, got %d", ii.QuantityReserved)
	}
	if ii.QuantityAvailable != 120 {
		t.Errorf("expected available=120 (150-30), got %d", ii.QuantityAvailable)
	}

	// Issue 40 more.
	ii, err = svc.AdjustInventory(ctx, ii.ProductID, ii.LocationID, 40, "ISSUE", "test issue")
	if err != nil {
		t.Fatalf("adjust (issue): %v", err)
	}
	if ii.QuantityOnHand != 110 {
		t.Errorf("expected on_hand=110, got %d", ii.QuantityOnHand)
	}
	if ii.QuantityAvailable != 80 {
		t.Errorf("expected available=80 (110-30), got %d", ii.QuantityAvailable)
	}
}

func TestReserveStock_AvailableEqualsOnHandMinusReserved(t *testing.T) {
	svc := newInventoryService(t)
	ctx := context.Background()

	ii, err := svc.CreateInventoryItem(ctx, "prod_2", "loc_2", 200, 10, 1000, decimal.NewFromInt(2))
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := svc.ReserveStock(ctx, ii.ProductID, ii.LocationID, 75, "ref_a"); err != nil {
		t.Fatalf("reserve: %v", err)
	}

	got, err := svc.GetInventoryItem(ctx, ii.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.QuantityOnHand-got.QuantityReserved != got.QuantityAvailable {
		t.Errorf("invariant broken: on_hand=%d reserved=%d available=%d (expected available=%d)",
			got.QuantityOnHand, got.QuantityReserved, got.QuantityAvailable, got.QuantityOnHand-got.QuantityReserved)
	}
}

func TestReleaseReservation_AvailableEqualsOnHandMinusReserved(t *testing.T) {
	svc := newInventoryService(t)
	ctx := context.Background()

	ii, err := svc.CreateInventoryItem(ctx, "prod_3", "loc_3", 80, 10, 1000, decimal.NewFromInt(3))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := svc.ReserveStock(ctx, ii.ProductID, ii.LocationID, 50, "ref_b"); err != nil {
		t.Fatalf("reserve: %v", err)
	}
	if err := svc.ReleaseReservation(ctx, "ref_b"); err != nil {
		t.Fatalf("release: %v", err)
	}

	got, err := svc.GetInventoryItem(ctx, ii.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.QuantityOnHand != 80 || got.QuantityReserved != 0 || got.QuantityAvailable != 80 {
		t.Errorf("expected (80, 0, 80), got (%d, %d, %d)",
			got.QuantityOnHand, got.QuantityReserved, got.QuantityAvailable)
	}
}

func TestExecuteStockTransfer_InvariantOnBothSides(t *testing.T) {
	svc := newInventoryService(t)
	ctx := context.Background()

	src, err := svc.CreateInventoryItem(ctx, "prod_4", "src_loc", 500, 10, 1000, decimal.NewFromInt(4))
	if err != nil {
		t.Fatalf("create src: %v", err)
	}
	dst, err := svc.CreateInventoryItem(ctx, "prod_4", "dst_loc", 100, 10, 1000, decimal.NewFromInt(4))
	if err != nil {
		t.Fatalf("create dst: %v", err)
	}
	if err := svc.ReserveStock(ctx, src.ProductID, src.LocationID, 25, "transfer_ref"); err != nil {
		t.Fatalf("reserve: %v", err)
	}

	st, err := svc.CreateStockTransfer(ctx, "src_loc", "dst_loc", "prod_4", 50)
	if err != nil {
		t.Fatalf("create transfer: %v", err)
	}
	if _, err := svc.ExecuteStockTransfer(ctx, st.ID); err != nil {
		t.Fatalf("execute transfer: %v", err)
	}

	from, _ := svc.GetInventoryItem(ctx, src.ID)
	if from.QuantityAvailable != from.QuantityOnHand-from.QuantityReserved {
		t.Errorf("src invariant broken: on_hand=%d reserved=%d available=%d",
			from.QuantityOnHand, from.QuantityReserved, from.QuantityAvailable)
	}
	to, _ := svc.GetInventoryItem(ctx, dst.ID)
	if to.QuantityAvailable != to.QuantityOnHand-to.QuantityReserved {
		t.Errorf("dst invariant broken: on_hand=%d reserved=%d available=%d",
			to.QuantityOnHand, to.QuantityReserved, to.QuantityAvailable)
	}
}

// Sanity check that the helper itself catches a deliberate violation.
func TestAssertInventoryInvariant_CatchesViolation(t *testing.T) {
	cases := []struct {
		name string
		ii   *domain.InventoryItem
	}{
		{"negative_on_hand", &domain.InventoryItem{QuantityOnHand: -1, QuantityReserved: 0, QuantityAvailable: 0}},
		{"negative_reserved", &domain.InventoryItem{QuantityOnHand: 10, QuantityReserved: -1, QuantityAvailable: 11}},
		{"negative_available", &domain.InventoryItem{QuantityOnHand: 10, QuantityReserved: 0, QuantityAvailable: -1}},
		{"available_mismatch", &domain.InventoryItem{QuantityOnHand: 10, QuantityReserved: 3, QuantityAvailable: 99}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := assertInventoryInvariant(tc.ii); err == nil {
				t.Errorf("expected invariant violation for %s, got nil", tc.name)
			}
		})
	}

	// And a happy path: invariant holds.
	good := &domain.InventoryItem{QuantityOnHand: 10, QuantityReserved: 3, QuantityAvailable: 7}
	if err := assertInventoryInvariant(good); err != nil {
		t.Errorf("expected nil error for valid invariant, got %v", err)
	}
}
