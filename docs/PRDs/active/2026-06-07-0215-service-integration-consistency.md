# PRD: Cross-Service Integration & Consistency Roadmap

**Date:** 2026-06-07
**Status:** Draft (Proposed)
**Supersedes:** None
**Linked Initiatives:** `2026-06-06-1557-cdd-gap-analysis.md`

---

## 1. Executive Summary & Goals

The ERP system is constructed as a distributed set of 7 microservices (Auth, CRM, FM, HR, M, PM, SCM) interacting asynchronously via Apache Kafka. A codebase audit reveals that while individual services compile and pass tests, their **asynchronous integration is incomplete**. 

Numerous `TODO` comments in Kafka consumer files indicate where cross-service event handling is stubbed out or left as log-only. Furthermore, there are schema discrepancies between the CDD (Contract-Driven Development) contracts, the markdown documentation, and the actual Go memory/data implementations.

### Strategic Goals
1. **Resolve Event Stubs (TODOs):** Map and implement the remaining event handlers in Go, transitioning stubbed logs into operational cross-service workflows.
2. **Standardize Consistency:** Harmonize the CDD files (`*.cdd`), documentation files (`*.md`), and Go structures to achieve 100% contract compliance.
3. **Establish Cross-Service Entity Map:** Document each service's table/entity structures, identifying duplicate data structures and defining owner services.
4. **Enforce Clean Decoupling:** Measure structural coherence and coupling metrics, outlining how to resolve overlaps without resorting to synchronous API calls.

---

## 2. Event Integration Audit & TODO Resolution Plan

Across the 7 services' Kafka consumer files, there are **20+ stubbed handlers** labeled with `TODO`. The following matrix details each stubbed integration and the planned business logic to activate it:

| Consumer Service | Event Topic Subscribed To | Source Service | Current State | Planned Operational Behavior |
|------------------|---------------------------|----------------|---------------|------------------------------|
| **CRM** | `scm.inventory.available` | SCM | Commented / Log-only | Bumps Lead/Opportunity status when inventory matches demand. |
| **CRM** | `fin.credit.check.completed` | FM | Commented / Log-only | Transition Sales Order from `DRAFT` to `CONFIRMED` if credit passes. |
| **CRM** | `hr.employee.performance` | HR | Commented / Log-only | Logs account manager reviews inside Customer Interactions. |
| **FM** | `scm.invoice.received` | SCM | Commented / Log-only | Automatically generates a `VendorBill` with matching line items. |
| **HR** | `fin.budget.allocated` | FM | Commented / Log-only | Adjusts recruitment target caps based on department budget. |
| **M** | `scm.material.received` | SCM | Commented / Log-only | Triggers production line kickoff for related Work Orders. |
| **M** | `scm.inventory.updated` | SCM | Commented / Log-only | Recalculates material constraints for active BOMs. |
| **M** | `fin.cost.budget.allocated` | FM | Commented / Log-only | Adjusts costing records and maximum work center capacities. |
| **M** | `hr.employee.scheduled` | HR | Commented / Log-only | Assigns operators to Work Centers based on training history. |
| **PM** | `hr.employee.available` | HR | Commented / Log-only | Adds employee to resource pool availability for project tasking. |
| **PM** | `hr.employee.skills.updated` | HR | Commented / Log-only | Updates resource allocation capabilities mapping. |
| **PM** | `fin.budget.approved` | FM | Commented / Log-only | Unlocks project budget ceiling, transitioning status to `ACTIVE`. |
| **PM** | `fin.payment.received` | FM | Commented / Log-only | Bumps project invoice billing summary status. |
| **PM** | `scm.material.delivered` | SCM | Commented / Log-only | Marks dependencies/milestones resolved for procurement tasks. |
| **PM** | `mfg.custom.production.completed` | M | Commented / Log-only | Auto-completes material request task for order delivery. |
| **SCM** | `fin.vendor.payment.processed` | FM | Commented / Log-only | Bumps Purchase Order payment status to `PAID` or `PARTIAL`. |

### Implementation Phases for TODOs
* **Phase I (Integration Test Coverage):** Write Go integration tests using `shared/testing.MockPublisher` that simulate these exact payloads before writing handlers.
* **Phase II (Handler Implementation):** Implement the missing event handler methods inside service layers (e.g., `services/pm-service/internal/business/service/project_planning_service.go`).
* **Phase III (Kafka Wire-up):** Uncomment consumer routing in `consumer.go` and update the topic subscription groups.

---

## 3. Consistency Analysis: Documentation vs. CDD vs. Code

We define the **System Consistency Index (SCI)** as the percentage of architectural concepts aligned across **CDD contracts**, **Go structs**, and **System Docs**.

### Current Alignment Assessment

```mermaid
radar-chart
    title Alignment Score by Service
    "Auth" : 95
    "CRM" : 90
    "FM" : 65
    "HR" : 85
    "M" : 70
    "PM" : 90
    "SCM" : 80
```

1. **Auth Service (95% Consistent):**
   - *CDD vs. Code:* High agreement. All 7 entities (`User`, `Session`, etc.) exist in code.
   - *Gaps:* `Session`, `Role`, `Permission`, `UserRole`, and `RolePermission` lack standalone HTTP routes in API gateway (they are managed internally or via RBAC service).
2. **CRM Service (90% Consistent):**
   - *CDD vs. Code:* Full domain struct alignment. 
   - *Gaps:* `OpportunityStageHistory` and `SalesOrderItem` do not have direct endpoints.
3. **FM Service (65% Consistent):**
   - *CDD vs. Code:* Major structural gaps. 7 contract-defined entities (`CurrencyRate`, `FiscalYear`, `CostCenter`, `BankAccount`, `CustomerCredit`, `BankStatement`, `BankStatementLine`) are implemented only in-memory but have **no SQL schemas** or repository bindings in actual DB initialization files.
   - *Gaps:* Reports like `getIncomeStatement` and `getCashFlow` are in `.cdd` but missing implementations in the Go services.
4. **HR Service (85% Consistent):**
   - *CDD vs. Code:* Structs align well.
   - *Gaps:* Lack of database-backed migration scripts for history entities (`EmployeeCompensationHistory`, `DepartmentHistory`, `PositionHistory`).
5. **M Service (70% Consistent):**
   - *CDD vs. Code:* Inconsistency in structural separation. `MaintenanceService` has no struct or class in Go; its methods are merged into `ProductionService`.
6. **PM Service (90% Consistent):**
   - *CDD vs. Code:* Excellent alignment. All entities match.
7. **SCM Service (80% Consistent):**
   - *CDD vs. Code:* Good struct agreement.
   - *Gaps:* Incomplete API endpoints for inner lines (`PurchaseOrderLine`, `ReceiptLine`, `ShipmentLine`).

### Consistency Action Plan
1. **CDD Sync:** Update `.cdd` files to remove deprecated methods or declare them as `@deprecated` if they are not planned.
2. **Go Stub Generation:** Write database migrations and repository interfaces for the 7 missing FM entities.
3. **API Alignment:** Expose missing gateway routes to enable programmatic integration testing.

---

## 4. Cross-Service Entity & Data Structure Directory

The ERP system contains duplicated entities representing similar concepts in different contexts. The directory below maps key database entities, their primary owners, and how they relate across service boundaries:

| Entity Name | Primary Owner | Description | Cross-Service Overlap & Dependency | Decoupling Strategy |
|-------------|---------------|-------------|------------------------------------|---------------------|
| **User** | Auth | Credentials and security sessions. | Maps 1:1 with `Employee` in HR. | Auth owns passwords/tokens. HR owns salary/employment. Linked solely by string ID (`EmployeeID`). |
| **Employee** | HR | Master employee record, roles, and history. | Referenced by PM (`ResourceAllocation.user_id`), M (`LaborReport.employee_id`), SCM (`PurchaseRequisition.requester_id`). | Consumer services must treat `EmployeeID` as a foreign key without database checks. They cache name/email locally if fast reads are required. |
| **Product** | SCM | Master SKU catalog, warehouse locations, and stock levels. | Referenced by M (`BillOfMaterials.product_id`), CRM (`SalesOrderItem.product_id`). | SCM is the source of truth. Manufacturing replicates changes via `scm.inventory.updated` events. |
| **SalesOrder** | CRM | Customer sale transactions, terms, and quotes. | Triggers Manufacturing custom orders (`mfg.custom.*`) and Finance billing (`fm.invoice.*`). | CRM publishes `crm.sales.order.confirmed`. FM and M consume this event to spawn `Invoice` and `ProductionOrder` respectively. |
| **Invoice** | FM | Financial accounts receivable billing. | Generated from CRM `SalesOrder`. | Decoupled via Kafka. CRM does not know about general ledger accounts; FM maps invoice items to GL categories. |
| **Budget** | FM | Fiscal limits and general ledger accounts. | PM service references `budget_id` to block overallocated project spend. | PM service caches active budget amounts. On budget adjustment, FM publishes `fin.budget.allocated` to update PM. |

---

## 5. Decoupling vs. Coherence Analysis

To prevent a "distributed monolith," the microservices must maintain high **coherence** (functional focus within services) and low **coupling** (interdependence across services).

### Coupling Metrics
* **Afferent Coupling ($C_a$):** Number of external services depending on this service's entities.
  - *Highest:* HR ($C_a = 4$) and SCM ($C_a = 3$).
* **Efferent Coupling ($C_e$):** Number of external services this service depends on.
  - *Highest:* PM ($C_e = 3$) and M ($C_e = 3$).

### Decoupling Rules

```
          [ Auth ]      [ CRM ]      [ FM ]
             │             │           │
             ▼             ▼           ▼
        EmployeeID    SalesOrderID  BudgetID
             │             │           │
             ▼             ▼           ▼
    [ HR Service ] ──► [ PM Service ] ◄── [ SCM Service ]
                            │
                            ▼
                        ProductID
```

1. **No Shared Databases:** No service may query the database tables of another service.
2. **Asynchronous replication:** All cross-service lookups (e.g., displaying employee names on PM tasks) must be resolved by:
   - Replicating critical fields via Kafka events and caching them locally.
   - Or resolving references at the API gateway level via parallel fetches.
3. **Eventual Invariant Validation:** Invariants crossing service boundaries (e.g., verifying a project does not exceed its budget) must be handled asynchronously:
   - PM registers project spending.
   - If cost exceeds budget, PM publishes a warning event.
   - Synchronous blocking is avoided; the system alerts rather than locking up under network partition.

---

## 6. Implementation Roadmap by Phases

### Phase 1: Harmonization (Target: 1 Week)
* **Contract and Schema Alignments:**
  - [x] **Subtask 1.1:** Write DB migration files (`.sql`) for missing FM tables: `currency_rates`, `fiscal_years`, `cost_centers`, `bank_accounts`, `customer_credits`, `bank_statements`, `bank_statement_lines`.
  - [x] **Subtask 1.2:** Declare repository interfaces in `services/fm-service/internal/business/domain/repository.go` for the 7 new entities.
  - [x] **Subtask 1.3:** Create in-memory mock implementations in `services/fm-service/internal/data/memory/` and SQL-backed repository structs in `services/fm-service/internal/data/sql/` implementing the interfaces.
  - [x] **Subtask 1.4:** Register/wire the new repository instances into `GeneralLedgerService` and other dependent services inside `services/fm-service/cmd/server/main.go`.
* **API Gateway and Routing Alignments:**
  - [x] **Subtask 2.1:** Add standalone HTTP endpoints for FM inner line entities (`InvoiceLine`, `VendorBillLine`, `BankStatementLine`) in `fm-service` router.
  - [x] **Subtask 2.2:** Add standalone HTTP endpoints for SCM lines (`PurchaseOrderLine`, `ReceiptLine`, `ShipmentLine`) in `scm-service` router.
  - [x] **Subtask 2.3:** Map newly exposed paths (e.g., `/api/fm/invoice-lines`, `/api/scm/po-lines`) inside the dynamic router `/api-gateway/internal/server/server.go` with JWT and RBAC checks.
* **Manufacturing Extraction:**
  - [x] **Subtask 3.1:** Define `MaintenanceService` domain struct and interfaces in `services/m-service/internal/business/domain/`.
  - [x] **Subtask 3.2:** Move the 7 maintenance-related methods (e.g., `CreateMaintenanceOrder`, `CompleteMaintenanceOrder`) from `ProductionService` to the new `MaintenanceService` class/struct.
  - [x] **Subtask 3.3:** Align `services/m-service/contracts/m.cdd` by splitting `MaintenanceService` as a distinct component.
  - [x] **Subtask 3.4:** Re-wire `services/m-service/cmd/main.go` to inject separate `ProductionService` and `MaintenanceService` instances.

### Phase 2: Event Integration (Target: 2 Weeks)
* **CRM Event Handlers:**
  - [x] **Subtask 4.1:** Implement consumer code for `scm.inventory.available` to check active leads and auto-update candidate statuses.
  - [x] **Subtask 4.2:** Implement consumer code for `fin.credit.check.completed` to auto-transition CRM `SalesOrder` from `DRAFT` to `CONFIRMED` upon credit authorization.
  - [x] **Subtask 4.3:** Implement consumer code for `hr.employee.performance` to sync performance indicators for sales representatives.
* **FM Event Handlers:**
  - [x] **Subtask 5.1:** Implement consumer code for `scm.invoice.received` to construct and log Draft `VendorBill` entities in the general ledger.
* **HR Event Handlers:**
  - [x] **Subtask 6.1:** Implement consumer code for `fin.budget.allocated` to calculate and adjust department recruitment caps.
* **Manufacturing (M) Event Handlers:**
  - [x] **Subtask 7.1:** Implement consumer code for `scm.material.received` to update scheduling constraints on related `WorkOrder` runs.
  - [x] **Subtask 7.2:** Implement consumer code for `scm.inventory.updated` to verify inventory level changes and adjust active BOM allocations.
  - [x] **Subtask 7.3:** Implement consumer code for `fin.cost.budget.allocated` to feed fiscal limits into cost centers and costing records.
  - [x] **Subtask 7.4:** Implement consumer code for `hr.employee.scheduled` to map scheduled crew capabilities and assign operators to active Work Centers.
* **Projects (PM) Event Handlers:**
  - [x] **Subtask 8.1:** Implement consumer code for `hr.employee.available` and `hr.employee.skills.updated` to maintain the resource scheduling pool.
  - [x] **Subtask 8.2:** Implement consumer code for `fin.budget.approved` to automatically unlock project budget ceilings and change project statuses to `ACTIVE`.
  - [x] **Subtask 8.3:** Implement consumer code for `fin.payment.received` to update project invoice/timesheet billing status.
  - [x] **Subtask 8.4:** Implement consumer code for `scm.material.delivered` to mark tasks dependencies/milestones completed.
  - [x] **Subtask 8.5:** Implement consumer code for `mfg.custom.production.completed` to complete tasks fulfilling specific sales order items.
* **SCM Event Handlers:**
  - [x] **Subtask 9.1:** Implement consumer code for `fin.vendor.payment.processed` to advance purchase order lifecycle status to paid/completed.
  - [x] **Subtask 9.2:** Implement picking list creation logic triggered by order allocation events.

### Phase 3: Validation & ADR Publication (Target: 0.5 Week)
* **Testing and Verification:**
  - [x] **Subtask 10.1:** Execute full uncached Go test suites (`go test -count=1 ./...`) across all 7 microservices.
  - [x] **Subtask 10.2:** Run the API Gateway routing verification suite to ensure proper path mapping.
  - [x] **Subtask 10.3:** Validate Docker-compose build and clean boot up of all 7 services.
* **Standards Enforcement:**
  - [x] **Subtask 11.1:** Write and publish `ADR-002-Entity-Decoupling-Patterns` inside `docs/architecture/` outlining decoupling patterns.
