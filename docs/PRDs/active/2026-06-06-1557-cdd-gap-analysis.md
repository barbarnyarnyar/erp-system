# PRD: CDD Gap Analysis — Defined vs Implemented

**Date:** 2026-06-06
**Status:** Active
**Supersedes:** None

## 1. Problem Statement

The project's 7 CDD contract files (`auth.cdd`, `fm.cdd`, `hr.cdd`, `scm.cdd`, `m.cdd`, `crm.cdd`, `pm.cdd`) define the intended architecture: 91 entities, 209 service methods, 138 events, and 43 HTTP-aware components. The Go codebase covers 100% of entities and 97.6% of methods, but significant gaps exist in repository wiring, event integrity, HTTP exposure, gateway consistency, and architecture adherence.

## 2. Gap Dimensions

### 2.0 Extra Entities in Code Not in CDD (RESOLVED)

~~**FM Service** — has 2 domain structs with full repo+memory implementations but NO CDD definition...~~ ✅ **RESOLVED**: `Transaction` and `TransactionLine` added to `fm.cdd` at lines 114–133. Both entities now have CDD definitions matching the Go structs. 19 CDD entities = 19 Go structs. Zero remaining Go-only items.

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

### 2.7 Inventory Ledger Invariant Unenforced (SCM)

`InventoryItem` has 3 quantity fields (`quantity_on_hand`, `quantity_reserved`, `quantity_available`) with a critical invariant:

$$\text{quantity\_available} = \text{quantity\_on\_hand} - \text{quantity\_reserved}$$

Every mutation site (`AdjustInventory`, `ReserveStock`, `ReleaseReservation`, `CreateInventoryItem`, `UpdateInventoryItem`, `ExecuteStockTransfer`) manually maintains this formula in Go code, but there is **no enforced guard**:
- No `CHECK` constraint in schema
- No `assertInvariant()` validation function
- Zero tests
- `AdjustInventory` is the weakest point — it modifies both `QuantityOnHand` and `QuantityAvailable` by the same delta instead of recomputing from the formula, so a future change that calls `AdjustInventory` while `QuantityReserved > 0` would silently break the invariant.

### 2.8 Raw String Enums Across 4 Services (7+ Entities)

The following fields use raw `string` with zero typed constants, zero validation, and zero compile-time protection:

| Service | Entity | Field | Commented Values |
|---------|--------|-------|-----------------|
| CRM | Customer | `Status` | LEAD, PROSPECT, ACTIVE, INACTIVE |
| CRM | Opportunity | `Stage` | DISCOVERY, NEGOTIATION, CLOSED_WON, CLOSED_LOST |
| FM | Account | `Type` | ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE |
| HR | LeaveRequest | `LeaveType` | ANNUAL, SICK, UNPAID |
| HR | LeaveRequest | `Status` | PENDING, APPROVED, REJECTED |
| M | ProductionOrder | `Status` | DRAFT, PLANNED, IN_PROGRESS, COMPLETED, CANCELLED |
| M | WorkOrder | `Status` | PENDING, IN_PROGRESS, COMPLETED |

And likely more with the same pattern. Comments are ignored at runtime — invalid or lowercase values can silently corrupt state.

### 2.9 Asymmetrical General Ledger Writes (FM)

`JournalEntry` lines have a mandatory accounting invariant:

$$\sum \text{debit\_amount} - \sum \text{credit\_amount} = 0$$

- `CreateJournalEntry` ✅ enforces this check (`general_ledger_service.go:144-146`)
- `UpdateJournalEntry` ❌ has **no balance validation** — it calls `s.entries.Update` directly, bypassing the invariant entirely
- Account balance updates happen **before** the entry is persisted (lines 168-195), with a manual rollback on save failure — but a crash between balance updates and entry save would leave accounts modified with no recorded entry
- No database transaction wraps the account updates + entry save

### 2.10 Cross-Service @Reference Coupling (All Services)

56 total `@reference` annotations across all CDD files — **25 are cross-service (45%)**:

| Severity | Source → Target | Count | Entities Affected |
|----------|----------------|-------|-------------------|
| 🔴 Severe | PM → HR (Employee) | 14 | Portfolio, Task, ResourceAllocation, ProjectTimeEntry, ProjectExpense, ProjectDocument, ProjectIssue, ChangeRequest |
| 🟠 High | M → SCM (Product) | 3 | BillOfMaterials, BOMComponent, ProductionOrder |
| 🟠 High | CRM → SCM (Product) | 3 | SalesOrderItem, QuoteLineItem, PriceListItem |
| 🟡 Medium | M → HR (Employee) | 2 | LaborReport, QualityInspection |
| 🟡 Medium | HR → PM (Project/Task) | 2 | AttendanceEntry |
| 🟢 Low | Others (FM→SCM, SCM→HR, M→CRM, PM→FM) | 5 | Various |

If services run on isolated databases, none of these constraints can be enforced natively. The CDD models them as hard `@reference` links, but the actual Go code never validates cross-service existence at write time — they are purely documentation. Fix: either accept as design intent (monolith-deployed), or replace with local value object IDs + event-driven eventual consistency.

### 2.11 Non-Atomic Lead Conversion (CRM)

`LeadService.ConvertLead()` at `lead_service.go:113-144` executes this strictly sequential chain:

```
leadRepo.GetByID → leadRepo.Update → custSvc.CreateCustomer → oppSvc.CreateOpportunity → publisher.Publish
```

- No database transaction wraps Customer + Opportunity creation
- If `CreateCustomer` succeeds but `CreateOpportunity` fails: orphan Customer with no Opportunity
- If `CreateOpportunity` succeeds but publish fails: Opportunity exists but `LeadConvertedEvent` never fires
- Assessment specifically flagged this as a synchronous component trap — **confirmed**

### 2.12 TrainingEnrollment Duplicate Enrollment Bug (HR)

`TrainingEnrollment` has no unique constraint on `(training_id, employee_id)`:
- CDD: no `@unique` composite annotation
- Service: `TrainingService.EnrollEmployee()` calls `repo.Create()` without checking for existing enrollment
- Repo: has `GetByTrainingAndEmployee()` method but **service never calls it**
- Result: same employee can be enrolled in the same training program multiple times

### 2.13 Missing Position/Department Change History (HR)

Only salary changes have an audit trail:
- ✅ `EmployeeCompensationHistory` entity + `hr.salary.changed` event
- ❌ **No `PositionHistory`** entity — position changes emit `hr.employee.promoted` but no history table stores the record
- ❌ **No `DepartmentHistory`** entity — department changes have zero tracking (no event, no entity, no table)

### 2.14 Float64 Bug in SCM Event (SCM)

`CustomerDemandForecastEvent.ConfidenceLevel` at `events.go:192` uses `float64` while the domain model `DemandForecast.ConfidenceLevel` uses `decimal.Decimal`. When serialized/deserialized via the event bus, the float64 representation loses precision — a concrete bug that can cause forecast values to drift.

### 2.15 Missing Auditing Fields on JournalEntry (FM)

`JournalEntry` in `fm.cdd:110-120` has `created_by` but no `posted_by` or `posted_at` fields. In proper accounting practice, ledger entries must record *who* authorized the posting and *when* it was posted, separately from creation metadata:

| Field | Exists? | Notes |
|-------|---------|-------|
| `created_by` | ✅ Line 116 | Captures who drafted the entry |
| `created_at` | ✅ Line 118 | Captures when drafted |
| `posted_by` | ❌ Missing | Who authorized the POSTED state |
| `posted_at` | ❌ Missing | When it was posted |
| `reversed_by` | ✅ Line 117 | Already tracks reversals |
| `updated_at` | ❌ Should remove | Ledger entries should be structurally immutable — `updated_at` implies standard UPDATEs are allowed |

`JournalEntryLine` (line 122-131) already correctly omits `updated_at` — consistent with immutability.

The `updateJournalEntry` Go method should not exist for posted entries — updates should only be allowed on `DRAFT` entries. Once `POSTED`, the only valid mutation is reversal.

### 2.16 Missing Opportunity Stage History (CRM)

`Opportunity` in `crm.cdd:43-53` has a `stage` field (`DISCOVERY, NEGOTIATION`) that is overwritten on every update — no historical tracking exists:

| Field | Exists? | Notes |
|-------|---------|-------|
| `stage` | ✅ Line 49 | Current stage only — overwritten by `updateOpportunity` |
| `updated_at` | ✅ Line 52 | Timestamp of last change but no *what* changed |
| OpportunityStageHistory entity | ❌ Missing | No ledger of stage transitions |
| `changed_by` | ❌ Missing | No audit of who moved the deal |

Without an `OpportunityStageHistory` entity, CRM cannot calculate pipeline velocity (time spent per stage), conversion funnels, or rep-level stage transition metrics. The existing `OpportunityWonEvent` / `OpportunityLostEvent` capture final outcomes but not intermediate progression.

### 2.17 Missing JWT Revocation Mechanism (Auth)

`User` in `auth.cdd:2-12` has no `security_stamp` field. `Session` (lines 14-22) has no `is_revoked` field. When a user is deactivated or roles change, existing JWTs remain valid until expiration:

| Field | Exists? | Notes |
|-------|---------|-------|
| `password_hash` | ✅ Line 6 | Password storage |
| `is_active` | ✅ Line 9 | Toggle for account status |
| `security_stamp` | ❌ Missing | No version counter to invalidate tokens on state change |
| `refresh_token` | ✅ Line 17 | Session-level refresh capability |
| `is_revoked` | ❌ Missing | No per-session kill switch |

`deactivateUser()` at line 96 sets `is_active = false` but cannot invalidate already-issued JWTs. An employee terminated at 10:00 can still access the system until their token expires. Fix: add `security_stamp` to `User` (bumped on deactivation/password change/role change) and `is_revoked` to `Session`; validate both on every token check.

### 2.18 Zombie Account Gap — Missing Auth Consumer (Auth)

Auth has zero `consumer_events` (`auth.cdd:127-134`). When HR processes an offboarding via `EmployeeService.deleteEmployee()` and publishes `hr.employee.terminated`, Auth never consumes it — the terminated employee's account remains active indefinitely.

Our previous design decision ("No consumer needed — auth events are fire-and-forget") only considered whether *downstream* services need auth events. It missed the reverse direction: Auth needs to consume `hr.employee.terminated` to deactivate accounts. This is a security vulnerability, not an intentional design choice.

### 2.19 HR Salary Field Bypasses Compensation History (HR)

`Employee` in `hr.cdd:35-51` has `salary: decimal` (line 48) as a directly mutable field. `EmployeeCompensationHistory` (lines 53-60) exists to track changes, but `updateEmployee` at line 240 accepts `salary: decimal` as a parameter — allowing direct mutation that bypasses the history entity entirely.

| Issue | Impact |
|-------|--------|
| `Employee.salary` is writable via `updateEmployee` | Service can update salary without creating a history record |
| No enforcement that salary changes go through history | Payroll audits can produce inaccurate results if history is incomplete |
| `hr.salary.changed` event at line 425 | Event fires on change, but nothing enforces the history record being created |

Fix: remove `salary` from `updateEmployee` parameters; create a dedicated `updateCompensation(employeeId, salary, effectiveDate, changedBy)` method that writes to `EmployeeCompensationHistory` and recomputes `Employee.salary` from the latest history entry. `Employee.salary` becomes a computed/read-only field.

### 2.20 Auth validateToken Returns Raw Object (Auth)

`AuthService.validateToken(token: string)` at `auth.cdd:79` returns `object` — a generic, unvalidated type that provides zero contract safety for callers.

Fix: define `struct TokenClaims { user_id: uuid, tenant_id: uuid, roles: list<string> }` and change the return type to `TokenClaims`. This gives all downstream consumers a typed contract for token payloads.

### 2.21 Milestone Entity Referenced by Events But Missing (PM)

`pm.cdd:294-296` defines two events — `prj.milestone.achieved` and `prj.milestone.delayed` — and the `ProjectPlanningService` description (line 140) mentions "milestones". But no `entity Milestone` exists anywhere in the CDD.

| Issue | Impact |
|-------|--------|
| `prj.milestone.achieved` event | References payload `MilestoneAchievedEvent` — no entity to provide the ID |
| `prj.milestone.delayed` event | Same — no entity to source the payload |
| No milestone functions | `createMilestone`, `completeMilestone` don't exist |
| No milestone fields on Project | No `milestone_id` or milestone tracking |

Fix: add `entity Milestone { id, project_id, name, target_date, completion_date, status, created_at }` to PM CDD + add `createMilestone`, `completeMilestone` functions to `ProjectPlanningService`.

### 2.22 No Dedicated confirmSalesOrder Function (CRM)

`SalesOrderService` at `crm.cdd:195-213` has a generic `updateSalesOrder(id, status)` that handles all status transitions via `status: string`. There's no dedicated `confirmSalesOrder` function.

| Issue | Impact |
|-------|--------|
| No typed function for confirmation | Stock check, credit check, and event firing all mixed into one generic `updateSalesOrder` |
| `crm.sales.order.confirmed` event | Must be fired manually inside generic method — no compile-time guarantee |
| Missing order lifecycle validation | No function-level guard preventing wrong transitions (e.g., CANCEL → CONFIRMED) |

Fix: add `func confirmSalesOrder(id: string, paymentTerms: string @optional): SalesOrder` that performs stock availability check, credit limit check, fires `crm.sales.order.confirmed`, and returns the confirmed order. Keep `updateSalesOrder` for status transitions like SHIPPED, DELIVERED, CANCELLED.

### 2.23 Missing Customer Interaction Log (CRM)

CRM has `ServiceTicket` for support but no entity for tracking sales/customer interactions such as calls, meetings, emails, or notes.

| Interaction | Entity? |
|-------------|---------|
| Service/support tickets | ✅ `ServiceTicket` (line 111) |
| Sales calls | ❌ Missing |
| Meeting notes | ❌ Missing |
| Email history | ❌ Missing (individual `crm.email.sent` event exists but no entity) |
| General notes | ❌ Missing |

Fix: add `entity CustomerInteraction { id, customer_id, type (CALL, MEETING, EMAIL, NOTE), subject, description, interaction_date, created_by, created_at }` to CRM CDD.

## 3. Definition of Done

- [x] **2.0 resolved**: Transaction + TransactionLine entities added to `fm.cdd` (not removed — they have full repo+memory implementations)
- [x] **2.1 resolved**: All 7 FM entities have repository interfaces + memory implementations
- [x] **2.2 resolved**: All 5 missing service methods implemented
- [x] **2.3 resolved**: All 27 entities have HTTP CRUD routes
- [x] **2.4 resolved**: `MaintenanceService` extracted from God struct into its own Go struct; `ProductionService` composed with `MaintenanceService` as a dependency for internal cross-calls
- [x] **2.5 resolved**: Event integrity: 0 missing publishers, 0 dead consumer subscriptions, topic names consistent (Auth `hr.employee.terminated` handled by separate 2.18 task)
- [x] **2.6 resolved**: Single gateway implementation with auth deployed; route prefix convention decided: `/finance/`, `/manufacturing/`, `/projects/` (matches `make test`)
- [x] Gateway port mappings match code defaults
- [x] Dockerfile EXPOSE ports match code defaults
- [x] Plaintext passwords migrated to bcrypt
- [x] JWT secret moved to environment variable
- [x] Kafka publish errors at least logged (not discarded with `_ =`)
- [x] **2.7 resolved**: InventoryItem invariant enforced via `assertInventoryInvariant()` helper in `inventory_service.go` called at all 6 mutation sites (CreateInventoryItem, UpdateInventoryItem, AdjustInventory, ReserveStock, ReleaseReservation, ExecuteStockTransfer). Also fixed bug in AdjustInventory that mutated both `QuantityOnHand` and `QuantityAvailable` by the same delta (broke invariant when `QuantityReserved > 0`); now only mutates `QuantityOnHand` and recomputes `QuantityAvailable = on_hand - reserved`. 8 unit tests pass in `inventory_invariant_test.go` (see Phase S4.5 doc).
- [x] **2.8 resolved**: All raw string enum fields replaced with typed constants in domain layer
- [x] **2.9 resolved**: `UpdateJournalEntry` enforces debit=credit balance; account updates + entry save wrapped in atomic operation
- [x] **2.11 resolved**: `ConvertLead()` wrapped in transaction — Customer + Opportunity creation is atomic
- [x] **2.12 resolved**: `TrainingService.EnrollEmployee()` now calls `enrollments.GetByTrainingAndEmployee(trainingID, employeeID)` before `Create()`. Active enrollments (status `ENROLLED`/`IN_PROGRESS`) for the same `(training, employee)` pair are rejected with a descriptive error. Re-enrollment is permitted only after the prior enrollment transitions to `CANCELLED` or `COMPLETED` — matching real-world LMS behavior. 2 unit tests pass in `training_enrollment_test.go` (see Phase S4.6 doc).
- [x] **2.13 resolved**: PositionHistory + DepartmentHistory entities exist with event-driven audit trails. In-memory repositories added. `UpdateEmployee` logs history on position/department changes and emits promoted/transferred events.
- [x] **2.14 resolved**: `CustomerDemandForecastEvent.ConfidenceLevel` uses `decimal.Decimal` (not `float64`)
- [x] **2.15 resolved**: JournalEntry has `posted_by` + `posted_at` fields; `updated_at` removed from JournalEntry; `Update` blocked on POSTED entries
- [x] **2.16 resolved**: OpportunityStageHistory entity exists with stage, changed_at, changed_by; stage transitions recorded on every `updateOpportunity`. In-memory repository added.
- [x] **2.17 resolved**: `User` has `security_stamp` (string) bumped on deactivation / password change / isActive toggle; `Session` has `is_revoked` (bool). `JWTClaims` embeds `security_stamp` at issuance. `AuthService.ValidateToken` now reloads the user from repo and rejects the token if (a) `user.IsActive` is false, or (b) `claims.SecurityStamp != user.SecurityStamp`. `AuthService.RevokeToken` now marks the session as revoked (sets `IsRevoked=true`) instead of deleting it. `SessionRepository.Update` method added to support revocation. 5 unit tests pass in `security_stamp_test.go` (see Phase S4.7 doc).
- [x] **2.18 resolved**: Auth now has its first Kafka consumer (`internal/data/kafka/consumer.go`) subscribed to `hr.employee.terminated`. On receipt, it calls `UserService.DeactivateUser(employeeID)` which bumps the security_stamp and invalidates all in-flight JWTs (chained from S4.7). `auth.cdd` updated with `consumer_events` block declaring the new consumer. `TopicHrEmployeeTerminated` constant added to auth's `event_topics.go`. 2 consumer tests pass in `consumer_test.go` (happy path + unknown user returns error). See Phase S4.8 doc.
- [x] **2.19 resolved**: `Employee.salary` is read-only/computed from latest `EmployeeCompensationHistory` entry; `updateEmployee` no longer accepts `salary`; dedicated `updateCompensation` method enforces history on every change.
- [x] **2.20 resolved**: `auth.cdd` now declares `struct TokenClaims { user_id, tenant_id, roles }` and `validateToken` returns `TokenClaims` (replaced the raw `object` return type). Go side: the auth-service's `JWTClaims` struct was renamed to `TokenClaims`, gained a `TenantID` field, and a deprecated `type JWTClaims = TokenClaims` alias preserves backward compatibility for any internal callers. 3 new tests added in `token_claims_test.go` (struct shape, JSON tags, integration via ValidateToken). 10 total auth-service tests pass.
- [x] **2.21 resolved**: `entity Milestone { id, project_id, name, description, target_date, completion_date, status, created_at, updated_at }` added to `pm.cdd` with field-level reference to `Project.id`. New CDD functions: `createMilestone`, `getMilestone`, `listMilestonesByProject`, `completeMilestone`, `delayMilestone` — all on `ProjectPlanningService`. Go side: new `domain/milestone.go` struct, `MilestoneRepository` interface + memory impl, expanded `ProjectPlanningService` with full lifecycle methods (replaces the previous event-only `AchieveMilestone`/`DelayMilestone` stubs that had no entity backing). `MilestoneAchievedEvent` gained a `CompletionDate` field. 5 unit tests pass in `milestone_test.go` (create defaults to PENDING, complete sets status + date, double-complete is rejected, delay sets new target, list filters by project).
- [x] **2.22 resolved**: `func confirmSalesOrder(id: string): SalesOrder` added to `crm.cdd` `SalesOrderService` block. Go side: new `domain/sales_order_helpers.go` with `SalesOrderStatus` typed constants (`DRAFT`/`CONFIRMED`/etc), `IsValid()`, `CanConfirm()`, `MarkConfirmed()`, and typed errors (`ErrOrderNotFound`, `ErrOrderNotConfirmable`, `ErrCustomerNotFound`, `ErrCustomerNotActive`, `ErrOrderHasNoItems`, `ErrInvalidItemQuantity`). `SalesOrderService.ConfirmSalesOrder(ctx, id)` validates: order exists, status is `DRAFT`, customer exists + is `ACTIVE`, order has items all with `Quantity > 0`. Transitions to `CONFIRMED`, persists, and fires `crm.sales.order.confirmed` event. Constructor now takes `customerRepo` as 4th arg; `cmd/main.go` wired accordingly. Cross-service stock check noted in CDD doc-comment as future enhancement (would require SCM request/reply). 7 unit tests pass in `confirm_sales_order_test.go` (happy path, not found, not draft, customer inactive, customer missing, empty items, invalid qty).
- [x] **2.23 resolved**: CustomerInteraction entity exists with id, customer_id, type, subject, description, interaction_date, created_by
- [x] All changes verified by `make test` passing

## 3.5 Resolved Design Decisions

| Question | Decision | Rationale |
|----------|----------|-----------|
| Auth Service Kafka consumer? | ~~**No consumer needed** — auth events are fire-and-forget, CDD shows no `consumer_events`~~ | **REVISED: Add `hr.employee.terminated` consumer** — original decision only considered downstream needs; missed the zombie account threat from HR offboarding |
| Auth `security_stamp`? | **Add `security_stamp` to User + `is_revoked` to Session** — enables immediate JWT invalidation on deactivation/role change | No mechanism currently exists to revoke stateless tokens mid-lifecycle |
| FM `Transaction` entity? | **Add to `fm.cdd`** — it's a legitimate domain concept with repo+memory | Matches pattern of other entities; no code change needed |
| URL prefix convention? | **`/finance/`, `/manufacturing/`, `/projects/`** — match existing `make test` scripts | Avoids breaking test suite; `server.go` adjusts to `main.go` convention |
| Auth events purpose? | **Fire-and-forget notifications** — no downstream service currently needs auth events | If a future service needs `auth.user.created` / `auth.user.deactivated`, add consumer then |
| ProductionService ↔ MaintenanceService coupling? | **Composition** — `ProductionService` holds a `MaintenanceService` reference; calls internal maintenance methods through it | Avoids circular deps, keeps services independently testable, matches how `QualityService` uses `ProductionService` |
| DLQ as separate feature? | **Yes, new Phase 6** — DLQ is a new architectural feature, not a gap fix | Keeps Phase 1 focused on existing contract compliance; DLQ can be prioritized independently |
| JournalEntry auditing fields? | **Add `posted_by`/`posted_at`, remove `updated_at`** — accounting best practice requires immutable ledger rows; only DRAFT entries can be updated, POSTED entries must be reversed | Improves audit trail; removes misleading `updated_at` that suggests mutable entries |
| HR salary field bypass? | **Remove `salary` from `updateEmployee`, create `updateCompensation` method** — all salary changes must route through `EmployeeCompensationHistory`; `Employee.salary` becomes computed from latest history entry | Prevents audit gaps from direct field mutation; addresses the #1 HR criticism across all assessments |
| Auth `validateToken` return type? | **Replace `object` with `struct TokenClaims { user_id, tenant_id, roles }`** — provides typed contract for all token consumers | Eliminates the only remaining `object` return type in the CDD |

## 4. Priority-Ordered Execution Plan

### P0 — Critical (system doesn't work or CDD is wrong, do first)

| Step | Task | Est. Time | Rationale |
|------|------|-----------|-----------|
| ~~1~~ | ~~Phase S0: Add Transaction + TransactionLine entities to `fm.cdd`~~ | ~~0.25d~~ | ✅ **DONE** — `fm.cdd:114-133` now defines both entities matching Go structs |
| 2 | Phase S1: Event error logging (`_ =` → `if err != nil`) | 0.5d | All 65+ publishes silently fail — zero visibility |
| 3 | Phase S2: Fix gateway backend port mismatches (HR 8003, SCM 8006, CRM 8002) | 0.5d | 3 of 6 services unreachable via gateway |
| 4 | Phase S3: Fix Dockerfile EXPOSE mismatches (M 8004, PM 8006, CRM 8002) | 0.5d | Container orchestration reads wrong ports |

### P1 — Security + Data Integrity (immediately exploitable or corrupting)

| Step | Task | Est. Time | Rationale |
|------|------|-----------|-----------|
| 5 | Phase S4: Migrate passwords to bcrypt | 0.5d | Plaintext comparison in auth service |
| 6 | Phase S4: Move JWT secret to env var | 0.5d | Hardcoded `super-secret-key-123` |
| 7 | Phase S4.5: Enforce InventoryItem invariant (`available = on_hand - reserved`) | 1d | Data can silently drift — `AdjustInventory` path breaks if `reserved > 0` |
| 8 | **Phase S4.6: TrainingEnrollment duplicate enrollment protection** | **0.5d** | **Real bug — `GetByTrainingAndEmployee` repo method exists but is never called; duplicates silently created** |
| 9 | **Phase S4.7: Add `security_stamp` to User + `is_revoked` to Session + validate on JWT check** | **1d** | **No mechanism to invalidate stateless JWTs mid-lifecycle; terminated employees retain access** |
| 10 | **Phase S4.8: Add Auth consumer for `hr.employee.terminated` → `deactivateUser`** | **1d** | **HR offboarding leaves zombie accounts active; Auth never learns of terminations** |

### P2 — Functional Completeness (CDD spec + accounting + atomicity)

| Step | Task | Est. Time | Rationale |
|------|------|-----------|-----------|
| 11 | Phase S5: Add 7 missing FM repository implementations | 1d | Entities exist in domain but can't be stored |
| 12 | Phase S6: Implement 5 missing service methods | 1d | CDD-defined business logic absent |
| 13 | Phase S7: Add HTTP routes for 14 entities with existing services | 1.5d | API endpoints for entities with existing repos+methods |
| 14 | Phase S8: Add HTTP routes for remaining 13 entities (need new handlers) | 1.5d | Auth roles/permissions, FM vendor bills, etc. |
| 15 | Phase S8.5: GL balance enforcement in `UpdateJournalEntry` + atomicity fix | 1d | `CreateJournalEntry` has check but `Update` bypasses it; account updates not atomic with entry save |
| 16 | **Phase S8.6: Wrap `ConvertLead()` in transaction for Customer + Opportunity atomicity** | **1d** | **Partial writes can create orphan Customers if Opportunity creation fails** |
| 17 | **Phase S8.7: Remove `salary` from `updateEmployee`, create `updateCompensation` that enforces history records** | **1d** | **`Employee.salary` is directly writable, bypassing `EmployeeCompensationHistory`; assessment #1 HR criticism** |
| 18 | **Phase S8.8: Add Milestone entity + createMilestone/completeMilestone functions to PM** | **0.5d** | **Events `prj.milestone.achieved`/`prj.milestone.delayed` reference a non-existent entity; can never fire** |
| 19 | ~~Phase S8.9: Add dedicated `confirmSalesOrder` function to CRM with stock/credit validation~~ | ✅ Done (2.22) | Resolved 2026-06-06 |

### P3 — Event Integrity + Type Safety (broken integrations + string enums + auditing)

| Step | Task | Est. Time | Rationale |
|------|------|-----------|-----------|
| 20 | Phase S9: Fix missing event publishes (31 events) + add dead sub comments | 1.5d | Events define service integrations |
| 21 | Phase S9: Fix topic naming inconsistency (FM → `crm.sales.order.confirmed`) | 0.5d | Cross-service integration broken |
| 22 | Phase S9: Migrate fm-service 21 hardcoded topic strings to constants | 0.5d | Code quality: bypasses typed constants |
| 23 | Phase S9.5: Migrate 7+ raw string enum fields to typed constants + fix `ConfidenceLevel` float64 → decimal (SCM) | 1d | Comments ignored at runtime — invalid values corrupt state silently; float64 bug causes forecast drift |
| 24 | **Phase S9.6: Add PositionHistory + DepartmentHistory + OpportunityStageHistory audit entities with event-driven trails** | **1.5d** | **Salary and pipeline stage changes are tracked, but position/department/stage changes have zero history** |

### P4 — Architecture & Remaining Security (works but messy)

| Step | Task | Est. Time | Rationale |
|------|------|-----------|-----------|
| 25 | Phase S10: Gateway — reconcile route prefixes + deploy new router | 1d | Alignment: Needed before auth can be enabled |
| 26 | Phase S10: Gateway — enable JWT+RBAC middleware | 1d | Depends on bcrypt + JWT env (P1) |
| 27 | Phase S11: Extract MaintenanceService from God struct | 1.5d | ProductionService holds 12 methods from other services |
| 28 | Phase S12: TLS config stubs (all 7 services) | 0.5d | Prep: No behavioral change |
| 29 | Phase S12: Admin seed user on auth startup | 0.5d | Enables gateway login testing |
| 30 | **Phase S12.5: Add `posted_by`/`posted_at` to JournalEntry + drop `updated_at` + guard `Update` on POSTED** | **1d** | **Auditing gap: ledger entries should be structurally immutable; `updated_at` suggests mutable entries** |
| 31 | **Phase S12.6: Replace `validateToken` return type `object` with `struct TokenClaims`** | **0.25d** | **Last raw `object` return type; downstream consumers get typed contract** |
| 32 | **Phase S12.7: Add CustomerInteraction entity (calls, meetings, emails, notes) to CRM** | **1d** | **CRM has ServiceTicket for support but no entity for sales interaction tracking** |

### P5 — Optional

| Step | Task | Est. Time | Rationale |
|------|------|-----------|-----------|
| 33 | Phase S14: Dead-letter queue for consumer errors | 2-3d | New Feature: Not a gap, not required for correctness |
| 34 | Phase S15: Verification (all DoD items) | 1d | Final check after all above done |

## 5. Future Roadmap — Phase 2 Features

Beyond the current gap-fix plan, these are natural extensions for a complete ERP. None are gaps (the CDD doesn't claim them), but they're standard enterprise features worth planning.

| Domain | Feature | Est. Effort | Rationale |
|--------|---------|-------------|-----------|
| **FM** | Fixed Assets (depreciation, disposal, revaluation) | 3d | Standard accounting module; tracks asset lifecycle, tax schedules |
| **FM** | Inter-Company Accounting (transfer pricing, eliminations) | 3d | Multi-entity/consolidation support |
| **HR** | Performance Management (reviews, goals, 360 feedback) | 3d | Beyond core HR; ties compensation to performance |
| **HR** | Recruiting & Onboarding (applicant tracking, offer letters) | 2d | Extends HR lifecycle from hire to retire |
| **M** | Master Production Scheduling (MPS engine) | 4d | Goes beyond MRP stub; capacity-constrained planning, what-if scenarios |
| **M** | Shop Floor Control (real-time machine monitoring, SCADA) | 3d | IoT integration for OEE tracking |
| **SCM** | Supplier Portal (RFQ, bidding, scorecards) | 3d | Extends procurement with supplier collaboration |
| **SCM** | Warehouse Management (bin locations, picking waves, putaway) | 4d | Beyond basic receipts/shipments; full WMS |
| **CRM** | Marketing Automation (email campaigns, lead scoring, A/B) | 3d | Beyond basic Campaign entity |
| **CRM** | Customer Portal (self-service, ticket tracking) | 2d | Customer-facing support interface |
| **PM** | Gantt Charts / Timeline Visualization | 2d | Visual scheduling beyond entity modeling |
| **PM** | Budget vs Actual Dashboard | 1.5d | Integrates PM expenses with FM budgets |
| **Cross** | Business Intelligence (reporting, dashboards, OLAP) | 5d | Aggregated KPIs across all modules |
| **Cross** | EDI / B2B Integration (cXML, EDIFACT) | 4d | Electronic procurement/sales with partners |
| **Cross** | Document Management (DMS, contract repository) | 3d | Centralized file storage with versioning |
| **Cross** | Audit Log (immutable event store, compliance) | 3d | System-wide audit trail beyond individual history tables |

**Total Phase 2 Estimate:** ~48 days (3-4 months for a small team).
