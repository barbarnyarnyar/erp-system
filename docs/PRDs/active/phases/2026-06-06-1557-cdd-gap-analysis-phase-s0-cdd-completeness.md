# ERP System CDD Gap Analysis — Phase S0: CDD Completeness (Code → CDD)

**Source PRD**: docs/PRDs/active/2026-06-06-1557-cdd-gap-analysis.md
**PRD ID**: PRD-2026-06-06-1557
**Phase**: S0 of S15 (P0 — Critical)
**Priority**: P0 — must be done first because CDD is the source of truth for all gap analysis
**Status**: Ready
**Created**: June 06, 2026

---

## Objective

Add `Transaction` and `TransactionLine` entities to `fm.cdd`. These are the only 2 Go structs in the entire codebase that exist in code but have no corresponding CDD definition. They have full repository interfaces and memory implementations.

## Rationale

CDD contracts are the authoritative source of truth for the architecture. If the CDD is incomplete, all downstream gap analysis (what's missing, what's extra, what needs routes) is based on an incomplete picture. Fixing this first ensures all subsequent phases operate against a complete contract.

## Scope

### In Scope

- Read `Transaction` and `TransactionLine` domain structs from Go code
- Read existing `fm.cdd` entity patterns for field format
- Add both entities to `fm.cdd` with matching fields
- The `Transaction` entity should include a `@reference` to related entities

### Out of Scope

- Removing `Transaction`/`TransactionLine` from code (they're legitimately used by services)
- Adding routes or handlers for them (already have repo+memory, handlers can wait for Phase S7/S8)
- Any other CDD changes

---

## Inputs

| Input | Source | Notes |
| ----- | ------ | ----- |
| Transaction struct | `services/fm-service/internal/business/domain/transaction.go:18` | 12 fields |
| TransactionLine struct | `services/fm-service/internal/business/domain/transaction.go:30` | 9 fields |
| TransactionRepository interface | `services/fm-service/internal/data/repositories/transaction_repository.go` | Methods: Create, GetByID, List, Update |
| Existing entity pattern in CDD | `services/fm-service/contracts/fm.cdd:1-20` | Format: `@entity EntityName` with typed fields |

---

## Implementation Tasks

### Task 1: Read Go structs

**Transaction** (from `transaction.go`):
```go
type Transaction struct {
    ID            string
    Reference     string
    Date          time.Time
    Description   string
    Status        string // DRAFT, POSTED, REVERSED
    TotalDebit    decimal.Decimal
    TotalCredit   decimal.Decimal
    CreatedBy     string
    ReversedBy    *string
    ReversalRef   *string
    CreatedAt     time.Time
    UpdatedAt     time.Time
}

type TransactionLine struct {
    ID            string
    TransactionID string
    AccountID     string
    DebitAmount   decimal.Decimal
    CreditAmount  decimal.Decimal
    Description   string
    CostCenterID  *string
    CreatedAt     time.Time
}
```

### Task 2: Add to `fm.cdd`

Insert before the `Account` entity (or after `JournalEntry` since Transaction is a related GL concept):

```
@entity Transaction
- id: uuid @primary
- reference: string @unique
- date: date
- description: string
- status: string  # DRAFT, POSTED, REVERSED
- total_debit: decimal
- total_credit: decimal
- created_by: string
- reversed_by: uuid @optional
- reversal_ref: string @optional
- created_at: timestamp
- updated_at: timestamp

@entity TransactionLine
- id: uuid @primary
- transaction_id: uuid @reference(Transaction.id)
- account_id: uuid @reference(Account.id)
- debit_amount: decimal
- credit_amount: decimal
- description: string
- cost_center_id: uuid @reference(CostCenter.id) @optional
- created_at: timestamp
```

**Note:** TransactionLine is a line-item entity (like JournalEntryLine, InvoiceLine, VendorBillLine) — bundled CRUD via TransactionRepository, not a standalone repo.

### Task 3: Verify completeness

```bash
# Verify both entities now in CDD
grep -c '@entity Transaction' services/fm-service/contracts/fm.cdd
grep -c '@entity TransactionLine' services/fm-service/contracts/fm.cdd
# Both should return >= 1

# Verify no other Go structs are missing from CDD
# (should be zero after this fix)
```

---

## Acceptance Criteria

- `fm.cdd` contains `@entity Transaction` with all fields matching the Go struct
- `fm.cdd` contains `@entity TransactionLine` with all fields matching the Go struct
- Zero remaining Go domain structs without a corresponding CDD entity

---

## Risks

| Risk | Likelihood | Mitigation |
| ---- | ---------- | ---------- |
| Transaction overlaps with JournalEntry | Medium — Transaction appears to be a legacy/alternate GL entry type | Keep both in CDD; note in comments that Transaction may be deprecated in favor of JournalEntry |
| CDD regeneration overwrites manual code | Low — CDD files are checked in manually, regeneration is a separate step |

## Definition of Done

- [ ] Task 1: Transaction + TransactionLine structs read and field-mapped
- [ ] Task 2: Both entities added to `fm.cdd` with correct types and references
- [ ] Task 3: Zero remaining Go-only domain structs (verified by scan)
- [ ] `grep '@entity Transaction' services/fm-service/contracts/fm.cdd` passes
