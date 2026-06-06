# ERP System CDD Gap Analysis — Phase 0: Trivial Gaps

**Source PRD**: docs/PRDs/active/2026-06-06-1557-cdd-gap-analysis.md
**PRD ID**: PRD-2026-06-06-1557
**Phase**: 0 of 6
**Status**: Ready
**Created**: June 06, 2026

---

## Objective

Close the simplest CDD-vs-code gaps: missing repository interfaces, missing service methods, and missing HTTP CRUD routes. These are pure additions with zero refactoring risk.

## Scope

### In Scope

- Create repository interfaces + memory implementations for 7 FM entities (CurrencyRate, FiscalYear, CostCenter, BankAccount, CustomerCredit, BankStatement, BankStatementLine)
- Implement 5 missing service methods: `GetIncomeStatement`, `GetCashFlow` (FM GL), `ListVendorBills` (FM AP), `ConsumeMaterials`, `ReceiveFinishedGoods` (M Production)
- Add HTTP CRUD routes for 27 entities missing them across Auth (5), FM (9), HR (3), SCM (6), M (4)
- Wire new repos, handlers, and services in each service's `main.go`

### Out of Scope

- Event integrity fixes (Phase 1)
- Architecture refactoring (Phase 2)
- Gateway consolidation (Phase 3)
- Security hardening (Phase 4)

---

## Inputs

| Input | Source | Notes |
| ----- | ------ | ----- |
| CDD contract files | `services/*/contracts/*.cdd` | Source of truth for entity definitions |
| Existing repository patterns | `services/*/internal/data/` | Follow existing repo interface + memory impl patterns |
| Existing route patterns | `services/*/internal/api/routes/` | Follow existing Gin route registration patterns |
| Existing service patterns | `services/*/internal/business/service/` | Follow existing method naming conventions |

## Dependencies

None — Phase 0 is purely additive.

---

## Implementation Tasks

### Task 1: Add FM repository implementations

**Description:** Create repository interfaces and in-memory implementations for 7 FM entities that currently only have domain structs.

**Entities:** CurrencyRate, FiscalYear, CostCenter, BankAccount, CustomerCredit, BankStatement, BankStatementLine

**Pattern to follow:** `services/fm-service/internal/data/repositories/account_repository.go` and `services/fm-service/internal/data/memory/account_repo.go`

**Note:** BankStatementLine is a line-item entity of BankStatement — bundle its CRUD inside BankStatementRepository (like JournalEntryLine is bundled in JournalEntryRepository), not as a standalone repo.

**Sub-tasks:**
- **1a**: CurrencyRate + FiscalYear repos and memory impls (2 entity types)
- **1b**: CostCenter + CustomerCredit repos and memory impls (2 entity types)
- **1c**: BankAccount repo and memory impl (1 entity type)
- **1d**: BankStatement + BankStatementLine repos and memory impls (bundled, 1 parent repo)
- **1e**: Wire all new repos into `main.go`, update service constructors that need them

**Acceptance Criteria:**
- All 6 entity types (7 entities, BankStatementLine bundled) have `*Repository` interfaces in `internal/data/repositories/`
- All have `Memory*Repo` implementations in `internal/data/memory/`
- `make build` passes for `fm-service`

**Files / Areas:**
- `services/fm-service/internal/data/repositories/`
- `services/fm-service/internal/data/memory/`
- `services/fm-service/cmd/server/main.go`

### Task 2: Implement missing service methods

**Description:** Add the 5 service methods defined in CDD but missing from Go code.

**Sub-tasks:**
- **2a**: FM GeneralLedgerService — `GetIncomeStatement`, `GetCashFlow` (follow pattern of `GetTrialBalance`: read all entries, compute report; wire into report handler if new endpoint needed)
- **2b**: FM AccountsPayableService — `ListVendorBills` (follow pattern of `ListInvoices` in AR service; returns list of vendor bills with basic filtering)
- **2c**: M ProductionService — `ConsumeMaterials` (reduce inventory quantities, create material consumed event)
- **2d**: M ProductionService — `ReceiveFinishedGoods` (create finished goods inventory records, create production completed event)

**Acceptance Criteria:**
- All 5 methods compile and return expected types
- Methods integrate with existing repository layer
- `make build` passes for both `fm-service` and `m-service`

**Files / Areas:**
- `services/fm-service/internal/business/service/general_ledger_service.go`
- `services/fm-service/internal/business/service/accounts_payable_service.go`
- `services/m-service/internal/business/service/production_service.go`

### Task 3: Add HTTP CRUD routes for missing entities (split by service)

**Description:** Add REST endpoints for entities that have domain structs and service methods but no HTTP routes. Each sub-task is a separate manageable unit.

**Sub-tasks:**

**3a — Auth routes (5 entities):** Session, Role, Permission, UserRole, RolePermission
- RBACService already has CreateRole/CreatePermission/AssignPermissionToRole/ValidatePermissions
- AuthService already has RefreshToken/RevokeToken (session endpoints)
- Endpoints to add: `GET/POST/PUT/DELETE /api/v1/roles`, `GET/POST/PUT/DELETE /api/v1/permissions`, `GET /api/v1/roles/:id/permissions`, `POST /api/v1/roles/:id/permissions`, `DELETE /api/v1/roles/:id/permissions/:permissionId`
- Create `services/auth-service/internal/api/handlers/rbac_handler.go`
- Update `services/auth-service/internal/api/routes/routes.go`

**3b — FM routes (9 entities):** Budget, CostCenter, BankAccount, CurrencyRate, FiscalYear, TaxRate, VendorBill, BankStatement, CustomerCredit
- BudgetingService already has ListBudgets/CreateBudget — wire into routes
- TaxService already has CreateTaxRate/ListTaxRates/GetTaxRate — wire into routes
- VendorBill: AccountsPayableService has CreateVendorBill/MatchPurchaseOrder, needs ListVendorBills (Task 2b first)
- BankStatement: CashManagementService has ReconcileBankStatement
- Rest: new services + handlers needed
- Follow pattern of `services/fm-service/internal/api/routes/routes.go`

**3c — HR routes (3 entities):** Department, Position, LeaveBalance
- Domain structs exist, repos exist, services exist via EmployeeManagement/LeaveManagement
- Endpoints: `GET/POST /api/v1/departments`, `GET/POST /api/v1/positions`, `GET /api/v1/leave-balances`
- Create handler methods in existing or new handler files

**3d — SCM routes (6 entities):** Location, InventoryMovement, PurchaseOrderLine, PurchaseRequisitionLine, ReceiptLine, ShipmentLine
- Location: add via ProductManagementService (it already has location patterns in domain/repo)
- Line items: serve as nested resources under parent (e.g., `GET /api/v1/purchase-orders/:id/lines`, `GET /api/v1/receipts/:id/lines`, `GET /api/v1/shipments/:id/lines`)
- InventoryMovement: add read-only list endpoint under inventory

**3e — M routes (4 entities):** BOMComponent, NonConformance, Equipment, CostingRecord
- BOMService already has AddBOMComponent/RemoveBOMComponent — add route wiring as `GET/POST /api/v1/boms/:id/components`, `DELETE /api/v1/boms/:id/components/:compId`
- QualityService already has list inspections — NonConformance as nested resource under inspections
- Equipment: defer to Phase 2 (maintenance routes go with MaintenanceService)
- CostingRecord: add read-only `GET /api/v1/costing-records/:id` endpoint

**Acceptance Criteria:**
- Endpoints return proper HTTP status codes (201 for create, 200 for get/list, 204 for delete)
- Endpoints integrate with existing service methods
- `make build` passes for all 6 affected services

**Files / Areas:**
- `services/auth-service/internal/api/routes/routes.go` + `handlers/rbac_handler.go`
- `services/fm-service/internal/api/routes/routes.go` + `handlers/` (new files per entity group)
- `services/hr-service/internal/api/routes/routes.go` + `handlers/`
- `services/scm-service/internal/api/routes/routes.go` + `handlers/`
- `services/m-service/internal/api/routes/routes.go` + `handlers/`

### Task 4: Wire new components in main.go

**Description:** Connect new repos, services, and handlers in each service's main entry point.

**Acceptance Criteria:**
- All 7 services start without nil pointer dereferences
- `make health` shows all services as healthy

**Files / Areas:**
- `services/*/cmd/main.go` or `services/*/cmd/server/main.go`

---

## Verification

```bash
# Build all affected services
cd services/fm-service && make build
cd services/m-service && make build
cd services/auth-service && make build
cd services/hr-service && make build
cd services/scm-service && make build

# Run tests
make test
```

### Manual

1. Start services: `make run`
2. Hit each new endpoint with `curl` and verify correct response
3. Verify all services report healthy on `/health`

---

## Risks

| Risk | Likelihood | Mitigation |
| ---- | ---------- | ---------- |
| New routes conflict with existing routes | Low | Review existing route files before adding |
| Missing domain structs for some entity fields | Low | Entities are 100% mapped at domain level |
| main.go wiring breaks startup | Low | One service at a time, test after each |

## Open Questions

- Should line-item entities (InvoiceLine, PurchaseOrderLine) get standalone read endpoints or only nested routes?
- What HTTP methods should Session endpoints support (create/delete only, no update)?

## Definition of Done

- [ ] Task 1: All 6 FM entity groups (7 entities, BankStatementLine bundled) have repo interfaces + memory implementations
- [ ] Task 2a: `GetIncomeStatement` + `GetCashFlow` implemented
- [ ] Task 2b: `ListVendorBills` implemented
- [ ] Task 2c: `ConsumeMaterials` implemented
- [ ] Task 2d: `ReceiveFinishedGoods` implemented
- [ ] Task 3a: All 5 Auth entities have HTTP routes
- [ ] Task 3b: All 9 FM entities have HTTP routes
- [ ] Task 3c: All 3 HR entities have HTTP routes
- [ ] Task 3d: All 6 SCM entities have HTTP routes
- [ ] Task 3e: All 4 M entities have HTTP routes (Equipment deferred to Phase 2)
- [ ] Task 4: All services start and pass health check
- [ ] `make build` passes for all services
- [ ] `make test` passes

---

## Handoff Notes

After Phase 0, the codebase will have 100% CDD-entity-to-repository coverage and 99% HTTP route coverage (all primary entities accessible via API). The remaining gaps are architectural (MaintenanceService extraction) and behavioral (event integrity, security).
