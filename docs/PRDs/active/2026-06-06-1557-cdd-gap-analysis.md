# PRD: CDD Gap Analysis — Defined vs Implemented

**Date:** 2026-06-06
**Status:** Active
**Supersedes:** None

## 1. Problem Statement

The project's 7 CDD contract files (`auth.cdd`, `fm.cdd`, `hr.cdd`, `scm.cdd`, `m.cdd`, `crm.cdd`, `pm.cdd`) define the intended architecture: 91 entities, 209 service methods, 138 events, and 43 HTTP-aware components. The Go codebase covers 100% of entities and 97.6% of methods, but significant gaps exist in repository wiring, event integrity, HTTP exposure, gateway consistency, and architecture adherence.

## 2. Gap Dimensions

### 2.0 Extra Entities in Code Not in CDD

**FM Service** — has 2 domain structs with full repo+memory implementations but NO CDD definition:

| Struct | File | Has Repo? | Has Memory Impl? | In CDD? |
|--------|------|-----------|-----------------|---------|
| `Transaction` | `transaction.go:18` | ✅ `TransactionRepository` | ✅ `MemoryTransactionRepo` | ❌ |
| `TransactionLine` | `transaction.go:30` | ✅ (bundled in Transaction repo) | ✅ (bundled) | ❌ |

These appear to be a legacy/alternative representation of `JournalEntry`/`JournalEntryLine`. Unlike the 7 missing repos (Section 2.1), these two entities have COMPLETE code but are simply missing from the CDD contract. Fix: add to `fm.cdd`.

### 2.1 Missing Repository Implementations (FM Service)
7 entities have domain structs but NO repository interface or memory implementation:

| Entity | CDD Defined | Domain Struct | Repo Interface | Memory Impl |
|---------|-------------|---------------|----------------|-------------|
| CurrencyRate | ✅ | ✅ | ❌ | ❌ |
| FiscalYear | ✅ | ✅ | ❌ | ❌ |
| CostCenter | ✅ | ✅ | ❌ | ❌ |
| BankAccount | ✅ | ✅ | ❌ | ❌ |
| CustomerCredit | ✅ | ✅ | ❌ | ❌ |
| BankStatement | ✅ | ✅ | ❌ | ❌ |
| BankStatementLine | ✅ | ✅ | ❌ | ❌ |

### 2.2 Missing Service Methods (2 Services, 5 Methods)

**FM Service — GeneralLedgerService**
- `getIncomeStatement()` — standard financial report
- `getCashFlow()` — cash flow statement

**FM Service — AccountsPayableService**
- `listVendorBills()` — basic CRUD gap

**M Service — ProductionService**
- `consumeMaterials()` — inventory integration point
- `receiveFinishedGoods()` — inventory integration point

### 2.3 Missing HTTP CRUD Routes (27 Entities)

| Service | Entities Without Standalone Routes |
|---------|-----------------------------------|
| Auth | Session, Role, Permission, UserRole, RolePermission (5) |
| FM | Budget, CostCenter, BankAccount, CurrencyRate, FiscalYear, TaxRate, VendorBill, BankStatement, CustomerCredit (9) |
| HR | Department, Position, LeaveBalance (3) |
| SCM | Location, InventoryMovement, PurchaseOrderLine, PurchaseRequisitionLine, ReceiptLine, ShipmentLine (6) |
| M | BOMComponent, NonConformance, Equipment, CostingRecord (4) |

### 2.4 Architectural Violations

**ProductionService God Struct (M Service)**
- CDD defines 5 components: BOMService, ProductionService, QualityService, MaintenanceService, CostingService
- Go code has 4 Go structs: `BOMService`, `ProductionService`, `QualityService`, `CostingService`
- **`MaintenanceService` has NO Go struct** — its 7 methods are implemented on `ProductionService` instead
- `ProductionService` has 28 methods vs CDD's 16 — absorbs Quality (3 methods), Maintenance (7 methods), Costing (2 methods)
- The God struct is wired in `main.go` with 13 repos instead of CDD's 5 clean components

### 2.5 Event Integrity Gaps

**Defined Constants Never Published (31 events)**
- SCM: 17 of 22 constants have no publish call (77% gap)
- PM: 8 of 25 constants have no publish call
- Auth: 0 of 5 constants have no publish — all 5 are published (no gap)
- HR: 5 of 22 constants never published (`payroll.failed`, `certification.earned`, `skill.acquired`, `employee.available`, `employee.skills.updated`)
- CRM: 3 of 28 never published (`email.opened`, `email.clicked`, `sales.order.received`)

**Consumed Events With No Publisher (16 events)**
- `scm.invoice.received`, `scm.material.received`, `scm.inventory.updated`, `scm.inventory.available`, `scm.shipment.delivered`, `scm.material.delivered`, `scm.training.required`
- `fin.vendor.payment.processed`, `fin.budget.allocated`, `fin.cost.budget.allocated`, `fin.credit.check.completed`
- `crm.sale.completed`, `crm.sales.order.received`, `crm.customer.demand.forecast`
- `hr.employee.scheduled`, `hr.employee.performance`

**Topic Naming Inconsistency**
- FM consumer subscribes to `crm.sale.completed` — CRM publishes `crm.sales.order.*` (different topic, never matched)
- fm-service defines constants in `event_topics.go` but uses hardcoded strings in publish calls

**Auth Service Has No Kafka Consumer**
- Auth publishes 5 events (`auth.user.created`, `auth.user.deactivated`, `auth.user.role.assigned`, `auth.user.store.assigned`, `auth.password.changed`) but subscribes to zero topics
- No other service subscribes to auth events either — all 5 topics have no consumers

**Fire-and-Forget Pattern (All Services)**
- Every `_ = publisher.Publish(...)` call discards the error
- No retry, no error logging (DLQ handled separately in Phase 6)

### 2.6 Gateway & Infrastructure Mismatches

**Dual Gateway Implementations**
- `api-gateway/cmd/main.go` (deployed): catch-all proxy, NO authentication, uses `finance/manufacturing/projects` URL prefixes
- `api-gateway/internal/server/server.go` (not deployed): explicit routes, JWT+RBAC auth middleware, uses `fm/m/pm` prefixes

**Gateway-to-Service Port Mismatches**
- Gateway routes `hr/*` → `hr-service:8002` but code defaults to port 8003
- Gateway routes `scm/*` → `scm-service:8003` but code defaults to port 8006
- Gateway routes `crm/*` → `crm-service:8005` but code defaults to port 8002

**Dockerfile EXPOSE Mismatches**
| Service | EXPOSE | Code Default |
|---------|--------|-------------|
| m-service | 8001 | 8004 |
| pm-service | 8001 | 8006 |
| crm-service | 8001 | 8002 |

**Security Gaps**
- Auth gateway (`server.go`) with JWT+RBAC not deployed — `main.go` has zero auth
- JWT secret hardcoded as `super-secret-key-123`
- Passwords stored as plaintext (`user.PasswordHash != password`)
- Zero TLS/HTTPS in any `.go` file

## 3. Definition of Done

- [ ] **2.0 resolved**: Transaction + TransactionLine entities added to `fm.cdd` (not removed — they have full repo+memory implementations)
- [ ] **2.1 resolved**: All 7 FM entities have repository interfaces + memory implementations
- [ ] **2.2 resolved**: All 5 missing service methods implemented
- [ ] **2.3 resolved**: All 27 entities have HTTP CRUD routes
- [ ] **2.4 resolved**: `MaintenanceService` extracted from God struct into its own Go struct; `ProductionService` composed with `MaintenanceService` as a dependency for internal cross-calls
- [ ] **2.5 resolved**: Event integrity: 0 missing publishers, 0 dead consumer subscriptions, topic names consistent; auth consumer marked intentional (no changes)
- [ ] **2.6 resolved**: Single gateway implementation with auth deployed; route prefix convention decided: `/finance/`, `/manufacturing/`, `/projects/` (matches `make test`)
- [ ] Gateway port mappings match code defaults
- [ ] Dockerfile EXPOSE ports match code defaults
- [ ] Plaintext passwords migrated to bcrypt
- [ ] JWT secret moved to environment variable
- [ ] Kafka publish errors at least logged (not discarded with `_ =`)
- [ ] All changes verified by `make test` passing

## 3.5 Resolved Design Decisions

| Question | Decision | Rationale |
|----------|----------|-----------|
| Auth Service Kafka consumer? | **No consumer needed** — auth events are fire-and-forget, CDD shows no `consumer_events` | Document as intentional; no code change |
| FM `Transaction` entity? | **Add to `fm.cdd`** — it's a legitimate domain concept with repo+memory | Matches pattern of other entities; no code change needed |
| URL prefix convention? | **`/finance/`, `/manufacturing/`, `/projects/`** — match existing `make test` scripts | Avoids breaking test suite; `server.go` adjusts to `main.go` convention |
| Auth events purpose? | **Fire-and-forget notifications** — no downstream service currently needs auth events | If a future service needs `auth.user.created` / `auth.user.deactivated`, add consumer then |
| ProductionService ↔ MaintenanceService coupling? | **Composition** — `ProductionService` holds a `MaintenanceService` reference; calls internal maintenance methods through it | Avoids circular deps, keeps services independently testable, matches how `QualityService` uses `ProductionService` |
| DLQ as separate feature? | **Yes, new Phase 6** — DLQ is a new architectural feature, not a gap fix | Keeps Phase 1 focused on existing contract compliance; DLQ can be prioritized independently |

## 4. Priority-Ordered Execution Plan

### P0 — Critical (system doesn't work or CDD is wrong, do first)

| Step | Task | Est. Time | Rationale |
|------|------|-----------|-----------|
| 1 | Phase S0: Add Transaction + TransactionLine entities to `fm.cdd` | 0.25d | CDD source of truth is incomplete — 2 entities fully coded but missing from contract |
| 2 | Phase S1: Event error logging (`_ =` → `if err != nil`) | 0.5d | All 65+ publishes silently fail — zero visibility |
| 3 | Phase S2: Fix gateway backend port mismatches (HR 8003, SCM 8006, CRM 8002) | 0.5d | 3 of 6 services unreachable via gateway |
| 4 | Phase S3: Fix Dockerfile EXPOSE mismatches (M 8004, PM 8006, CRM 8002) | 0.5d | Container orchestration reads wrong ports |

### P1 — Security (immediately exploitable)

| Step | Task | Est. Time | Rationale |
|------|------|-----------|-----------|
| 5 | Phase S4: Migrate passwords to bcrypt | 0.5d | Plaintext comparison in auth service |
| 6 | Phase S4: Move JWT secret to env var | 0.5d | Hardcoded `super-secret-key-123` |

### P2 — Functional Completeness (CDD spec, 27% API surface missing)

| Step | Task | Est. Time | Rationale |
|------|------|-----------|-----------|
| 7 | Phase S5: Add 7 missing FM repository implementations | 1d | Entities exist in domain but can't be stored |
| 8 | Phase S6: Implement 5 missing service methods | 1d | CDD-defined business logic absent |
| 9 | Phase S7: Add HTTP routes for 14 entities with existing services | 1.5d | API endpoints for entities with existing repos+methods |
| 10 | Phase S8: Add HTTP routes for remaining 13 entities (need new handlers) | 1.5d | Auth roles/permissions, FM vendor bills, etc. |

### P3 — Event Integrity (event-driven architecture broken)

| Step | Task | Est. Time | Rationale |
|------|------|-----------|-----------|
| 11 | Phase S9: Fix missing event publishes (31 events) + add dead sub comments | 1.5d | Events define service integrations |
| 12 | Phase S9: Fix topic naming inconsistency (FM → `crm.sales.order.confirmed`) | 0.5d | Cross-service integration broken |
| 13 | Phase S9: Migrate fm-service 21 hardcoded topic strings to constants | 0.5d | Code quality: bypasses typed constants |

### P4 — Architecture & Remaining Security (works but messy)

| Step | Task | Est. Time | Rationale |
|------|------|-----------|-----------|
| 14 | Phase S10: Gateway — reconcile route prefixes + deploy new router | 1d | Alignment: Needed before auth can be enabled |
| 15 | Phase S10: Gateway — enable JWT+RBAC middleware | 1d | Depends on bcrypt + JWT env (P1) |
| 16 | Phase S11: Extract MaintenanceService from God struct | 1.5d | ProductionService holds 12 methods from other services |
| 17 | Phase S12: TLS config stubs (all 7 services) | 0.5d | Prep: No behavioral change |
| 18 | Phase S12: Admin seed user on auth startup | 0.5d | Enables gateway login testing |

### P5 — Optional

| Step | Task | Est. Time | Rationale |
|------|------|-----------|-----------|
| 19 | Phase S14: Dead-letter queue for consumer errors | 2-3d | New Feature: Not a gap, not required for correctness |
| 20 | Phase S15: Verification (all DoD items) | 1d | Final check after all above done |
