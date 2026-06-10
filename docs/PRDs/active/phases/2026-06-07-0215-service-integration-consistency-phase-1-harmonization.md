# Service Integration Consistency - Phase 1: Harmonization

**Source PRD**: docs/PRDs/active/2026-06-07-0215-service-integration-consistency.md
**PRD ID**: PRD-2026-06-07-0215
**Phase**: 1 of 3
**Status**: Ready
**Created**: June 10, 2026
**Author**: Jules

---

## Objective

This phase focuses on harmonizing the implementation across microservices by aligning contracts, schemas, and API gateway routing. It involves implementing missing SQL-backed repository layers for Financial Management (FM), adding standalone HTTP endpoints in `fm-service` and `scm-service` routers, registering the new endpoints in the API Gateway with proper RBAC controls, and officially registering the extracted `MaintenanceService` in the CDD file.

## Scope

### In Scope

- Implement SQL-backed repository instances for 7 FM entities.
- Add and wire missing HTTP endpoints in `fm-service` and `scm-service`.
- Map new endpoints in the API Gateway (`api-gateway/internal/server/server.go`).
- Define `MaintenanceService` domain structs and interfaces in `services/mfg-service/contracts/mfg.cdd`.

### Out of Scope

- Implementing Event Handlers (Target of Phase 2).
- Adding completely new functional domains outside the specified entities.

---

## Inputs

| Input | Source | Notes |
| ----- | ------ | ----- |
| FM Schema Migrations | `services/fm-service/internal/data/migrations/schema.sql` | 7 tables are already present. |
| FM Repository Interface | `services/fm-service/internal/business/domain/repository.go` | Interfaces defined. |

## Dependencies

| Dependency | Type | Required Before | Notes |
| ---------- | ---- | --------------- | ----- |
| N/A | N/A | N/A | Foundational phase. |

---

## Implementation Tasks

### Task 1: Create SQL Repository Implementations for FM Service

- [ ] Create `services/fm-service/internal/data/sql/sql_repos.go`.
- [ ] Implement `domain.CurrencyRateRepository`.
- [ ] Implement `domain.FiscalYearRepository`.
- [ ] Implement `domain.CostCenterRepository`.
- [ ] Implement `domain.BankAccountRepository`.
- [ ] Implement `domain.CustomerCreditRepository`.
- [ ] Implement `domain.BankStatementRepository`.
- [ ] Implement `domain.TransactionRepository`.
- [ ] Update `services/fm-service/cmd/main.go` to inject these SQL repositories when a database connection is active (or replace memory versions as appropriate based on existing logic).

**Acceptance Criteria:**

- `sql_repos.go` exists and compiles.
- `cmd/main.go` wires these repositories up to `GeneralLedgerService` and others without compile errors.

**Files / Areas:**

- `services/fm-service/internal/data/sql/sql_repos.go` - New file with 7 struct implementations.
- `services/fm-service/cmd/main.go` - Wiring.

### Task 2: API Gateway and Routing Alignments

- [ ] Add standalone HTTP endpoints for FM inner line entities (`InvoiceLine`, `VendorBillLine`, `BankStatementLine`) in `fm-service` router. Note: Create basic placeholder/passthrough handlers in `services/fm-service/internal/api/handlers` if they don't exist.
- [ ] Add standalone HTTP endpoints for SCM lines (`PurchaseOrderLine`, `ReceiptLine`, `ShipmentLine`) in `scm-service` router (`services/scm-service/internal/api/routes/routes.go`).
- [ ] Map newly exposed paths (e.g., `/api/fm/invoice-lines`, `/api/scm/po-lines`) inside the dynamic router `api-gateway/internal/server/server.go` with JWT and RBAC checks.

**Acceptance Criteria:**

- `fm-service` router exposes line endpoints.
- `scm-service` router exposes line endpoints.
- API gateway forwards requests to those endpoints with RBAC checking.

**Files / Areas:**

- `services/fm-service/internal/api/routes/routes.go` - New routes.
- `services/fm-service/internal/api/handlers/` - New handlers for the lines.
- `services/scm-service/internal/api/routes/routes.go` - New routes.
- `services/scm-service/internal/api/handlers/` - Handlers for the lines.
- `api-gateway/internal/server/server.go` - Proxy setup.

### Task 3: Manufacturing Extraction Alignment

- [ ] Align `services/mfg-service/contracts/mfg.cdd` by explicitly defining `MaintenanceService` as a distinct interface.
- [ ] Move methods such as `CreateMaintenanceOrder`, `CompleteMaintenanceOrder`, `LogMachineStatus`, `CreateEquipment`, `ScheduleMaintenance`, `ListMaintenanceSchedules`, `GetMaintenanceSchedule`, `UpdateMaintenanceSchedule` into `interface MaintenanceService { ... }` in the `.cdd` file.

**Acceptance Criteria:**

- `mfg.cdd` has a clear `MaintenanceService` interface definition.

**Files / Areas:**

- `services/mfg-service/contracts/mfg.cdd` - `.cdd` modification.

---

## Verification

### Automated

```bash
cd services/fm-service && go build ./... && go test ./...
cd services/scm-service && go build ./... && go test ./...
cd api-gateway && go build ./... && go test ./...
```

### Manual

1. Run API Gateway and ensure routing configuration compiles.
2. Check `mfg.cdd` syntax validity if there's a specific parser script, otherwise just visually inspect.

---

## Risks

| Risk | Likelihood | Mitigation |
| ---- | ---------- | ---------- |
| Compilations failures on new repos | Low | Rely on Go type-checking and existing memory repos as reference. |

## Open Questions

- Should the SQL Repositories fully replace Memory Repositories in `main.go`, or should they be toggled by a config variable? (Assuming replacement for a real DB-backed application).

## Definition of Done

- [ ] All implementation tasks completed
- [ ] Acceptance criteria verified
- [ ] Automated checks passing
- [ ] Manual verification completed
- [ ] No unresolved blockers remain

---

## Handoff Notes

These steps cover the specific subtasks outlined in Phase 1 of the integration consistency PRD.
