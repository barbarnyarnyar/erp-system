# PRD: HR Service CDD Reconciliation & Legacy Code Cleanup

**PRD ID**: PRD-2026-06-13-1845  
**Date**: 2026-06-13  
**Status**: Implemented & Verified  
**Parent Initiative**: Technical Documentation Sync & Domain Parity  
**Target Coverage**: 100% CDD alignment for Workforce Identity and payroll core, zero architectural drift  

---

## 1. Objective & Architectural Context

Currently, the Human Resources service (`hr-service`) contains a large amount of legacy domain models and service logic (e.g. attendance tracking, training courses, job recruitment, and performance reviews) that are not defined in the contract `hr.cdd`. This bloating creates compilation mismatches and breaks clean boundary isolation.

To restore structural consistency:
1. We will prune all legacy domain Go models and interfaces.
2. We will align the GORM database structures and SQL/Memory repositories with the 7 CDD entities: `Department`, `EmployeeMaster`, `PayrollRun`, `ExpenseClaim`, `ExpenseClaimLine`, `TransactionalOutbox`, and `KafkaEventInbox`.
3. We will implement only the 3 core business services (`EmployeeService`, `PayrollService`, `ExpenseService`), the transactional outbox worker, and reliable messaging.
4. We will expose clean REST API controllers and map event producers and consumers to the CDD spec.

---

## 2. Technical Scope & Entity Mapping

### A. CDD Entities (7)
* **`Department`** (`hr_departments`)
* **`EmployeeMaster`** (`hr_employees`)
* **`PayrollRun`** (`hr_payroll_runs`)
* **`ExpenseClaim`** (`hr_expense_claims`)
* **`ExpenseClaimLine`** (`hr_expense_claim_lines`)
* **`TransactionalOutbox`** (`hr_transactional_outbox`)
* **`KafkaEventInbox`** (`hr_kafka_event_inbox`)

### B. Decoupling Rules ($C_e = 0$)
* All cross-domain associations (e.g. references to FM LegalEntity, HR Employee ID, SCM items) must store raw `uuid` strings instead of GORM pointers or joins.

---

## 3. Scope & Implementation Checklist

### Phase 1: Legacy Code Cleanup
- [x] Delete legacy domain model files under `internal/business/domain/` (e.g., job application, leave requests, performance review, attendance entry).
- [x] Delete legacy business services from `internal/business/service/` (e.g., leave, performance, recruitment, report, time attendance, training).
- [x] Clean up legacy API handlers.

### Phase 2: Repository Layer Refactoring
- [x] Update `internal/business/domain/repository.go` to declare interface signatures only for the 7 CDD entities.
- [x] Refactor `internal/data/sql/models.go` to declare the GORM structs and domain mapper functions for the 7 entities.
- [x] Rewrite `internal/data/sql/sql_repos.go` and `internal/data/memory/memory_repos.go` to implement database/memory repository adapters for the 7 entities.
- [x] Update GORM `AutoMigrate` inside `internal/data/sql/db.go` to only migrate the 7 CDD tables.

### Phase 3: Service Layer Refactoring
- [x] Implement `EmployeeService`, `PayrollService`, and `ExpenseService` inside `internal/business/service/`.
- [x] Implement `OutboxRelayWorker` and `ReliableMessagingService` to support transactional messaging.

### Phase 4: API Handlers, Routing & Main Entrypoint Wiring
- [x] Create a single unified `HrHandler` in `internal/api/handlers/` mapping REST endpoints to service routines.
- [x] Update `internal/api/routes/routes.go` to expose the new CDD routes.
- [x] Re-wire `cmd/main.go` to initialize database repositories, business services, HTTP routes, and consumer loop.

### Phase 5: Event Streams & Verification
- [x] Update `internal/data/kafka/consumer.go` to idempotently consume `prj.time.logged` and `fm.vendor.paid`.
- [x] Compile and verify using `go build ./...` and run test suite using `go test ./...`.

---

## 4. Definition of Done
- [x] `hr.cdd` is fully reconciled and implemented in code.
- [x] Stale models and services are 100% pruned.
- [x] All Go packages compile and pass tests successfully.
