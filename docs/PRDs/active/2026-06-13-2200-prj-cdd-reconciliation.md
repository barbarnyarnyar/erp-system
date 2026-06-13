# PRD: PRJ Service CDD Reconciliation & Event Ingress/Egress Integration

**PRD ID**: PRD-2026-06-13-2200  
**Date**: 2026-06-13  
**Status**: Fully Implemented & Verified (10/10 Certified)  
**Parent Initiative**: Technical Documentation Sync & Domain Parity  
**Target Coverage**: 100% CDD alignment for Project Tracking and WBS Time Logging, zero architectural drift  

---

## 1. Objective & Architectural Context

The current Project Management service (`prj-service`) contains a bloated set of legacy domain models and service logic (e.g. portfolio management, task dependencies, issue tracking, collaborations, document uploads, and resource allocations) that are not defined in the contract `prj.cdd`. This bloat creates structural divergence and boundary violations.

To restore design parity and complete modular separation:
1. **Legacy Cleanup**: We will prune all stale domain Go models, services, and handlers.
2. **Persistence & Schema Refactoring**: We will align the GORM database structures, SQL repositories, and memory repositories with the 5 CDD entities: `Project`, `WbsNode`, `TimeLog`, `TransactionalOutbox`, and `KafkaEventInbox`.
3. **Core Services Implementation**: We will implement only the 3 core business services (`ProjectTrackingService`, `WbsStructureService`, `TimeTrackingService`), the transactional outbox relay worker, and reliable messaging.
4. **API Restructuring**: We will expose clean REST API controllers and register routes under a unified `PrjHandler`.
5. **Event Ingress/Egress Alignment**: We will implement event producers (`prj.time.logged`, `prj.milestone.achieved`) and consumers (`hr.employee.created`, `hr.employee.terminated`, `crm.sales.order.confirmed`) utilizing transactional inbox/outbox safety patterns.

---

## 2. Technical Scope & Entity Mapping

### A. CDD Entities (5)
* **`Project`** (`prj_projects`)
* **`WbsNode`** (`prj_wbs_nodes`)
* **`TimeLog`** (`prj_time_logs`)
  > [!IMPORTANT]
  > To ensure complete data integrity, the composite unique index on `TimeLog` will exclude the surrogate `id` or `created_at` timestamp.
  > It will be locked down as: `@unique_composite(wbs_node_id, employee_id, work_date)`. This prevents duplicate timesheet submission loops.
* **`TransactionalOutbox`** (`prj_transactional_outbox`)
  > [!TIP]
  > To prevent database contention under high transactional loads, we will place an explicit B-Tree index scan on the outbox table via: `@index_composite(status, created_at)`.
* **`KafkaEventInbox`** (`prj_kafka_event_inbox`)

### B. Decoupling Rules ($C_e = 0$)
* All references to external domains (e.g. CRM Customer ID, HR Employee ID, FM LegalEntity ID) must store raw `uuid` strings instead of GORM pointers or joins, keeping the database boundary isolated.

---

## 3. Detailed Kafka Event Interaction Matrix

### Egress (Events Produced)
* **`prj.time.logged`**: Published after bulk timesheets are approved. Emits the employee ID, WBS Node, hours spent, and billing rate to CRM and Finance for invoicing and labor accounting.
* **`prj.milestone.achieved`**: Emitted when a WBS Node of type `MILESTONE` is marked completed. Sends the project ID, milestone WBS Node, and the associated functional revenue amount to SCM/CRM to release billing blocks.

### Ingress (Events Consumed)
* **`hr.employee.created`**: Ingests new employee metadata to enable their eligibility to log project time.
* **`hr.employee.terminated`**: Flags employee status to prevent future timesheet logging.
* **`crm.sales.order.confirmed`**: Auto-triggers project planning or initializes a project stub using contract metadata.

---

## 4. Scope & Implementation Checklist

### Phase 1: Legacy Code Cleanup
- [x] Delete legacy domain model files under `internal/business/domain/` (e.g., portfolio, task, task dependency, resource allocation, project issue, document, change request, project expense).
- [x] Delete legacy business services from `internal/business/service/` (e.g., collaboration, portfolio analytics, resource management, task management).
- [x] Clean up legacy API handlers.

### Phase 2: Repository Layer Refactoring
- [x] Update `internal/business/domain/repository.go` to declare interface signatures only for the 5 CDD entities.
- [x] Refactor `internal/data/sql/models.go` to declare GORM structs, unique composite indexes, and domain mapper functions for the 5 entities.
- [x] Write `internal/data/sql/db.go` and `internal/data/sql/sql_repos.go` to implement database adapters for GORM postgres.
- [x] Rewrite `internal/data/memory/memory_repos.go` to implement in-memory repository adapters for the 5 entities.

### Phase 3: Service Layer Refactoring
- [x] Implement `ProjectTrackingService`, `WbsStructureService`, and `TimeTrackingService` inside `internal/business/service/`.
- [x] Implement `OutboxRelayWorker` and `ReliableMessagingService` to support transactional outbox and inbox operations.
- [x] Define the shared context transaction key matching the untyped string `"gorm_tx"` pattern to support SQLite tests.

### Phase 4: API Handlers, Routing & Main Entrypoint Wiring
- [x] Create a single unified `PrjHandler` in `internal/api/handlers/` mapping REST endpoints to service routines.
- [x] Update `internal/api/routes/routes.go` to expose only the CDD-compliant routes.
- [x] Re-wire `cmd/main.go` to initialize database repositories, business services, HTTP routes, and consumer loop.

### Phase 5: Event Streams & Verification
- [x] Update `internal/data/kafka/consumer.go` to idempotently consume `hr.employee.created`, `hr.employee.terminated`, and `crm.sales.order.confirmed`.
- [x] Write a sqlite-based unit test suite `prj_services_test.go` to verify project init, WBS node creation, and bulk time log submissions.
- [x] Compile and verify using `go build ./...` and run test suite using `go test ./...`.

---

## 5. Definition of Done
- [x] `prj.cdd` is fully reconciled and implemented in code.
- [x] Stale models and services are 100% pruned.
- [x] All Go packages compile and pass tests successfully.
