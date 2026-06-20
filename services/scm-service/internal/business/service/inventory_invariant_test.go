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
		memory.NewMemoryStockBalanceRepo(),
		memory.NewMemoryInventoryMovementRepo(),
		memory.NewMemoryStockTransferRepo(),
		&sharedtesting.MockPublisher{},
		memory.NewMemoryTransactionManager(),
	)
}

func TestAdjustInventory_MaintainsInvariant_WithReservations(t *testing.T) {
	svc := newInventoryService(t)
	ctx := context.Background()

	sb, err := svc.CreateStockBalance(ctx, "prod_1", "loc_1", decimal.NewFromInt(100))
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Pre-reserve 30 units.
	if err := svc.ReserveStock(ctx, sb.MaterialID, sb.LocationID, decimal.NewFromInt(30), "ref_1"); err != nil {
		t.Fatalf("reserve: %v", err)
	}

	// Receive 50 more.
	sb, err = svc.AdjustInventory(ctx, sb.MaterialID, sb.LocationID, decimal.NewFromInt(50), "RECEIPT", "test receipt")
	if err != nil {
		t.Fatalf("adjust (receipt): %v", err)
	}
	if !sb.QuantityOnHand.Equal(decimal.NewFromInt(150)) {
		t.Errorf("expected on_hand=150, got %s", sb.QuantityOnHand)
	}
	if !sb.QuantityReserved.Equal(decimal.NewFromInt(30)) {
		t.Errorf("expected reserved=30, got %s", sb.QuantityReserved)
	}
	if !sb.QuantityAvailable.Equal(decimal.NewFromInt(120)) {
		t.Errorf("expected available=120 (150-30), got %s", sb.QuantityAvailable)
	}

	// Issue 40 more.
	sb, err = svc.AdjustInventory(ctx, sb.MaterialID, sb.LocationID, decimal.NewFromInt(40), "ISSUE", "test issue")
	if err != nil {
		t.Fatalf("adjust (issue): %v", err)
	}
	if !sb.QuantityOnHand.Equal(decimal.NewFromInt(110)) {
		t.Errorf("expected on_hand=110, got %s", sb.QuantityOnHand)
	}
	if !sb.QuantityAvailable.Equal(decimal.NewFromInt(80)) {
		t.Errorf("expected available=80 (110-30), got %s", sb.QuantityAvailable)
	}
}

func TestReserveStock_AvailableEqualsOnHandMinusReserved(t *testing.T) {
	svc := newInventoryService(t)
	ctx := context.Background()

	sb, err := svc.CreateStockBalance(ctx, "prod_2", "loc_2", decimal.NewFromInt(200))
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := svc.ReserveStock(ctx, sb.MaterialID, sb.LocationID, decimal.NewFromInt(75), "ref_a"); err != nil {
		t.Fatalf("reserve: %v", err)
	}

	got, err := svc.GetStockBalance(ctx, sb.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	expected := got.QuantityOnHand.Sub(got.QuantityReserved)
	if !got.QuantityAvailable.Equal(expected) {
		t.Errorf("invariant broken: on_hand=%s reserved=%s available=%s (expected available=%s)",
			got.QuantityOnHand, got.QuantityReserved, got.QuantityAvailable, expected)
	}
}

func TestReleaseReservation_AvailableEqualsOnHandMinusReserved(t *testing.T) {
	svc := newInventoryService(t)
	ctx := context.Background()

	sb, err := svc.CreateStockBalance(ctx, "prod_3", "loc_3", decimal.NewFromInt(80))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := svc.ReserveStock(ctx, sb.MaterialID, sb.LocationID, decimal.NewFromInt(50), "ref_b"); err != nil {
		t.Fatalf("reserve: %v", err)
	}
	if err := svc.ReleaseReservation(ctx, "ref_b"); err != nil {
		t.Fatalf("release: %v", err)
	}

	got, err := svc.GetStockBalance(ctx, sb.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !got.QuantityOnHand.Equal(decimal.NewFromInt(80)) || !got.QuantityReserved.Equal(decimal.Zero) || !got.QuantityAvailable.Equal(decimal.NewFromInt(80)) {
		t.Errorf("expected (80, 0, 80), got (%s, %s, %s)",
			got.QuantityOnHand, got.QuantityReserved, got.QuantityAvailable)
	}
}

func TestExecuteStockTransfer_InvariantOnBothSides(t *testing.T) {
	svc := newInventoryService(t)
	ctx := context.Background()

	src, err := svc.CreateStockBalance(ctx, "prod_4", "src_loc", decimal.NewFromInt(500))
	if err != nil {
		t.Fatalf("create src: %v", err)
	}
	dst, err := svc.CreateStockBalance(ctx, "prod_4", "dst_loc", decimal.NewFromInt(100))
	if err != nil {
		t.Fatalf("create dst: %v", err)
	}
	if err := svc.ReserveStock(ctx, src.MaterialID, src.LocationID, decimal.NewFromInt(25), "transfer_ref"); err != nil {
		t.Fatalf("reserve: %v", err)
	}

	st, err := svc.CreateStockTransfer(ctx, "src_loc", "dst_loc", "prod_4", 50)
	if err != nil {
		t.Fatalf("create transfer: %v", err)
	}
	if _, err := svc.ExecuteStockTransfer(ctx, st.ID); err != nil {
		t.Fatalf("execute transfer: %v", err)
	}

	from, _ := svc.GetStockBalance(ctx, src.ID)
	expectedFrom := from.QuantityOnHand.Sub(from.QuantityReserved)
	if !from.QuantityAvailable.Equal(expectedFrom) {
		t.Errorf("src invariant broken: on_hand=%s reserved=%s available=%s",
			from.QuantityOnHand, from.QuantityReserved, from.QuantityAvailable)
	}
	to, _ := svc.GetStockBalance(ctx, dst.ID)
	expectedTo := to.QuantityOnHand.Sub(to.QuantityReserved)
	if !to.QuantityAvailable.Equal(expectedTo) {
		t.Errorf("dst invariant broken: on_hand=%s reserved=%s available=%s",
			to.QuantityOnHand, to.QuantityReserved, to.QuantityAvailable)
	}
}

func TestAssertInventoryInvariant_CatchesViolation(t *testing.T) {
	cases := []struct {
		name string
		sb   *domain.StockBalance
	}{
		{"negative_on_hand", &domain.StockBalance{QuantityOnHand: decimal.NewFromInt(-1), QuantityReserved: decimal.Zero, QuantityAvailable: decimal.Zero}},
		{"negative_reserved", &domain.StockBalance{QuantityOnHand: decimal.NewFromInt(10), QuantityReserved: decimal.NewFromInt(-1), QuantityAvailable: decimal.NewFromInt(11)}},
		{"negative_available", &domain.StockBalance{QuantityOnHand: decimal.NewFromInt(10), QuantityReserved: decimal.Zero, QuantityAvailable: decimal.NewFromInt(-1)}},
		{"available_mismatch", &domain.StockBalance{QuantityOnHand: decimal.NewFromInt(10), QuantityReserved: decimal.NewFromInt(3), QuantityAvailable: decimal.NewFromInt(99)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := assertInventoryInvariant(tc.sb); err == nil {
				t.Errorf("expected invariant violation for %s, got nil", tc.name)
			}
		})
	}

	// And a happy path: invariant holds.
	good := &domain.StockBalance{QuantityOnHand: decimal.NewFromInt(10), QuantityReserved: decimal.NewFromInt(3), QuantityAvailable: decimal.NewFromInt(7)}
	if err := assertInventoryInvariant(good); err != nil {
		t.Errorf("expected nil error for valid invariant, got %v", err)
	}
}
