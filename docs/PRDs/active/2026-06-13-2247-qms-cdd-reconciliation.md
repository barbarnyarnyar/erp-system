# PRD: QMS Service CDD Reconciliation & Event Ingress/Egress Integration

**PRD ID**: PRD-2026-06-13-2247  
**Date**: 2026-06-13  
**Status**: Implemented & Approved  
**Parent Initiative**: Technical Documentation Sync & Domain Parity  
**Target Coverage**: 100% CDD alignment for Quality Plans, Inspections, Non-Conformance, and SPC analytics  

---

## 1. Objective & Architectural Context

The Quality Management System (QMS) service (`qms-service`) is currently running purely on in-memory maps without database persistence. It lacks a GORM-based PostgreSQL repository layer, transactional event outbox, and idempotent event ingestion boundaries. Furthermore, its statistical process control (SPC) calculations are mock values.

To implement a production-grade QMS service in alignment with ERP modular architecture:
1. **GORM Database & SQL Repositories**: Create the GORM structures and SQL repository layer mapping to the 7 CDD entities: `InspectionPlan`, `InspectionMetricDefinition`, `QualityInspection`, `InspectionResultLine`, `NonConformanceLog`, `TransactionalOutbox`, and `KafkaEventInbox`.
2. **Idempotence & Transaction Boundaries**: Implement the `ReliableMessagingService` and `OutboxRelayWorker` using context-propagated GORM transactions (using key `"gorm_tx"`).
3. **Idempotent Consumer**: Integrate incoming Kafka topics (`scm.receipt.staged`, `mfg.yield.produced`, `hr.employee.created`) using transaction-locked deduplication logs.
4. **Transactional Outbox Egress**: Publish QMS events (`qms.inspection.passed`, `qms.inspection.failed`, `qms.disposition.executed`) transactionally via the outbox table.
5. **Partition-Pruned SPC Analytics**: Implement statistical calculation (`computeSpcDistribution`) in `QualityAnalyticsService` querying metrics dynamically over a strict `TimeRange` boundary.
6. **SQLite Testing**: Verify all services and repositories end-to-end via in-memory SQLite unit tests.

---

## 2. Technical Scope & Entity Mapping

### A. CDD Entities (7)
* **`InspectionPlan`** (`qms_inspection_plans`)
* **`InspectionMetricDefinition`** (`qms_inspection_metric_definitions`)
* **`QualityInspection`** (`qms_quality_inspections`)
* **`InspectionResultLine`** (`qms_inspection_results`)
  - Range-partitioned monthly on `created_at`. Unique composite key on `(inspection_id, metric_definition_id, sample_sequence, created_at)`.
* **`NonConformanceLog`** (`qms_non_conformances`)
* **`TransactionalOutbox`** (`qms_transactional_outbox`)
  - Composite B-Tree index on `(status, created_at)`.
* **`KafkaEventInbox`** (`qms_kafka_event_inbox`)

### B. Decoupling Rules ($C_e = 0$)
* All references to external boundaries (e.g. SCM Receipt ID, MFG WorkOrder ID, HR Employee ID, PLM Material ID) must store raw `uuid` strings instead of database pointers or foreign key joins.

---

## 3. Kafka Event Interaction Matrix

### Egress (Events Produced)
* **`qms.inspection.passed`**: Published when all metric samples comply with plan tolerance limits. Signals MES to complete the work order or SCM to release the inventory.
* **`qms.inspection.failed`**: Published when one or more metric samples violate limits, generating a non-conformance ticket. Tells MES/SCM to quarantine the lot.
* **`qms.disposition.executed`**: Emitted when a quarantine non-conformance log receives structural resolution (Release, Rework, Scrap, Return).

### Ingress (Events Consumed)
* **`scm.receipt.staged`**: Automatically creates a pending inbound receipt quality inspection ticket.
* **`mfg.yield.produced`**: Automatically creates a production yield quality inspection ticket.
* **`hr.employee.created`**: Synchronizes employee roles to validate inspector credentials.

---

## 4. Scope & Implementation Checklist

### Phase 1: SQL Persistence Layer
- [x] Create `internal/data/sql/db.go` initializing connection pool, transaction helpers, and `AutoMigrate` for the 7 entities.
- [x] Create `internal/data/sql/models.go` with GORM structs, unique composite indexes, and domain mappers.
- [x] Create `internal/data/sql/sql_repos.go` implementing GORM SQL repositories.
- [x] Update `internal/data/memory/memory_repos.go` to align memory repositories with the updated CRUD domain signatures.

### Phase 2: Service Layer & Analytics implementation
- [x] Update `internal/business/service/service.go` to implement `InspectionPlanService`, `InspectionExecutionService`, `NonConformanceService`, `QualityAnalyticsService`, `OutboxRelayWorker`, and `ReliableMessagingService`.
- [x] Re-implement `ComputeSpcDistribution` to compute actual Mean and Standard Deviation of numeric metrics matching the `TimeRange` bounding coordinates.
- [x] Integrate GORM transaction context propagation across all service layers.

### Phase 3: Consumer & Outbox Relaying
- [x] Refactor `internal/data/kafka/consumer.go` to process messages idempotently using `ReliableMessagingService.ExecuteIdempotentTransaction`.
- [x] Wire outgoing events to write to the transactional outbox during status updates.

### Phase 4: Route Registration & Main Entrypoint
- [x] Update `internal/api/handlers/handlers.go` and `internal/api/routes/routes.go` to support transaction context propagation.
- [x] Update `cmd/main.go` to bootstrap the SQL persistence layer.

### Phase 5: Verification & Testing
- [x] Write SQLite-based tests in `internal/business/service/service_test.go` testing plan creation, inspection execution, non-conformance log logging, and SPC distribution computation.
- [x] Ensure that `go build ./...` compiles cleanly and `go test ./...` passes.

---

## 5. Definition of Done
- [x] QMS microservice uses SQL PostgreSQL repository in production and SQLite in tests.
- [x] All 7 CDD entities are aligned.
- [x] Code compiles and tests pass successfully.
