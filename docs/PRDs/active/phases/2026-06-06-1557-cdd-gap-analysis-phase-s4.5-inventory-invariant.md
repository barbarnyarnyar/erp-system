# ERP System CDD Gap Analysis — Phase S4.5: InventoryItem Invariant Enforcement

**Source PRD**: docs/PRDs/active/2026-06-06-1557-cdd-gap-analysis.md
**PRD ID**: PRD-2026-06-06-1557
**Phase**: S4.5 of 6
**Status**: Active
**Created**: June 06, 2026

---

## Objective

Enforce the critical SCM inventory invariant:

$$\text{quantity\_available} = \text{quantity\_on\_hand} - \text{quantity\_reserved}$$

This invariant is currently maintained by hand at every mutation site, with **no validation** and **no test**. The `AdjustInventory` path has a latent bug: it mutates both `QuantityOnHand` and `QuantityAvailable` by the same delta, which silently breaks the invariant when `QuantityReserved > 0`.

## Rationale

- Concrete bug in `AdjustInventory` (line 117-124 of `inventory_service.go`)
- No `assertInvariant()` function — invariant only maintained by developer discipline
- No `CHECK` constraint (using in-memory repo; DB constraint not feasible)
- Zero tests
- 6 mutation sites in `inventory_service.go`:
  1. `CreateInventoryItem` (line 50) — sets all 3 quantities
  2. `UpdateInventoryItem` (line 81) — sets all 3 quantities
  3. `AdjustInventory` (line 105) — **mutates both on_hand + available by delta (BUG)**
  4. `ReserveStock` (line 216) — increments reserved, recomputes available
  5. `ReleaseReservation` (line 247) — decrements reserved, recomputes available
  6. `ExecuteStockTransfer` (line 325) — moves qty between from/to locations

## Scope

### In Scope
- Add `assertInvariant(ii *domain.InventoryItem) error` helper that validates `available == on_hand - reserved` and all three are `>= 0`
- Fix `AdjustInventory` to mutate only `QuantityOnHand`, then recompute `QuantityAvailable` from formula
- Call `assertInvariant` after every mutation (at end of each method, before publishValuation)
- Return error from helper if invariant violated — caller can decide whether to fail the write or log
- Add a small unit test file `inventory_invariant_test.go` with at least 3 cases:
  - Happy path: reserve stock, available = on_hand - reserved
  - Bounded: cannot go below 0
  - Bug regression: AdjustInventory with reserved > 0 maintains invariant

### Out of Scope
- Database `CHECK` constraint (in-memory repo; no DB)
- Per-location reservations (current design is global)
- Negative inventory policies (still rejected)

---

## Implementation Tasks

### Task 1: Add `assertInventoryInvariant` helper
- File: `services/scm-service/internal/business/service/inventory_service.go`
- Add private method `assertInventoryInvariant(ii *domain.InventoryItem) error`
- Returns nil if valid; returns `fmt.Errorf` describing the violation
- Validates: `on_hand >= 0`, `reserved >= 0`, `available >= 0`, `available == on_hand - reserved`

### Task 2: Fix `AdjustInventory` bug
- Lines 117-124: stop mutating `QuantityAvailable` by delta
- Replace with: `ii.QuantityOnHand += qty` (RECEIPT/ADJUSTMENT_ADD) or `ii.QuantityOnHand -= qty` (ISSUE/ADJUSTMENT_SUB)
- Recompute: `ii.QuantityAvailable = ii.QuantityOnHand - ii.QuantityReserved`

### Task 3: Call `assertInventoryInvariant` at end of each mutation method
- `CreateInventoryItem`: after `s.invRepo.Create`
- `UpdateInventoryItem`: after `s.invRepo.Update`
- `AdjustInventory`: after `s.invRepo.Update` (line 130)
- `ReserveStock`: after `s.invRepo.Update` (line 230)
- `ReleaseReservation`: after `s.invRepo.Update` (line 269)
- `ExecuteStockTransfer`: after BOTH `fromItem` and `toItem` updates (lines 347, 377)

If invariant fails, return error (do NOT persist broken state — but the in-memory store already updated, so the practical mitigation is: log + return error + rely on tests).

### Task 4: Add unit test
- File: `services/scm-service/internal/business/service/inventory_invariant_test.go`
- Test cases:
  1. `TestAdjustInventory_MaintainsInvariant_WithReservations` — pre-reserve some stock, then adjust, assert available = on_hand - reserved
  2. `TestReserveStock_AvailableEqualsOnHandMinusReserved`
  3. `TestReleaseReservation_AvailableEqualsOnHandMinusReserved`
  4. `TestExecuteStockTransfer_InvariantOnBothSides`

---

## Verification

```bash
# Build
cd services/scm-service && go build ./...

# Run new test
cd services/scm-service && go test ./internal/business/service/ -run TestInventory -v

# Confirm no regression in other services (no other service uses InventoryItem)
for svc in auth fm hr m crm pm; do
  cd services/$svc-service && go build ./... && cd ../..
done
```

---

## Risks

| Risk | Likelihood | Mitigation |
| ---- | ---------- | ---------- |
| Invariant error causes in-memory state to drift from repo | Medium | Tests cover both happy path and invariant violation; in-memory state can be reset on test failure |
| Other call sites I missed | Low | grep confirmed only 6 sites; also `warehouse_service.go` for transfers |
| Performance — invariant check on every write | Negligible | Single integer comparison + 3 int comparisons; microsecond cost |

## Definition of Done

- [x] `assertInventoryInvariant` helper added
- [x] `AdjustInventory` bug fixed (no longer mutates QuantityAvailable by delta)
- [x] All 6 mutation sites call `assertInventoryInvariant`
- [x] Unit tests added (≥3 cases)
- [x] `go build ./...` passes for all 7 services
- [x] `go test ./...` passes in scm-service
- [x] Master PRD 2.7 DoD checkbox marked complete
