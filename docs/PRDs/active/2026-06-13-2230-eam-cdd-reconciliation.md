# PRD: EAM Service CDD Reconciliation & Event Ingress/Egress Integration

**PRD ID**: PRD-2026-06-13-2230  
**Date**: 2026-06-13  
**Status**: Ready for Implementation (Approved)  
**Parent Initiative**: Technical Documentation Sync & Domain Parity  
**Target Coverage**: 100% CDD alignment for Equipment Registry, Break-Fix Maintenance, Calendar PMs, and Telemetry Buffers  

---

## 1. Objective & Architectural Context

The Enterprise Asset Management (EAM) service is responsible for physical plant orchestration, tracking equipment uptime, scheduling preventative maintenance, and processing machine telemetry. 

To bring EAM into full modular alignment:
1. **Model Synchronization**: We will align the GORM database structures, SQL repositories, and memory repositories with the 7 CDD entities: `Facility`, `Equipment`, `MaintenanceWorkOrder`, `PreventativeSchedule`, `TelemetryIngestBuffer`, `TransactionalOutbox`, and `KafkaEventInbox`.
2. **Persistence Layer & Indexing**: 
   - Define composite B-Tree indexes on outbox queries (`status`, `created_at`).
   - Implement soft delete support on `Equipment` (`deleted_at`).
   - Track telemetry buffering securely with transactional draining and batch cleanup.
3. **Core Services Implementation**: We will implement the 3 business services (`EquipmentService`, `MaintenanceService`, `TelemetryIngestionService`), the outbox relay worker, and reliable messaging.
4. **API Restructuring**: We will expose clean REST API controllers and register routes under a unified `EamHandler`.
5. **Event Streams (Ingress/Egress)**: We will implement Kafka producers (`eam.machine.offline`, `eam.machine.online`, `eam.workorder.spares_requested`) and consumers (`scm.asset.received`, `fm.asset.capitalized`, `hr.employee.created`) using transactionally safe outbox/inbox processing boundaries.

---

## 2. Technical Scope & Entity Mapping

### A. CDD Entities (7)
* **`Facility`** (`eam_facilities`)
* **`Equipment`** (`eam_equipment`)
  - Must include GORM native soft delete field (`deleted_at`).
  - Stores `technical_specifications` as a flexible `jsonb` column.
* **`MaintenanceWorkOrder`** (`eam_work_orders`)
  - Tracks delta between `reported_at` and `resolved_at` for downtime.
* **`PreventativeSchedule`** (`eam_pm_schedules`)
  - Monitored by scheduler loops to generate preventative work orders.
* **`TelemetryIngestBuffer`** (`eam_telemetry_ingest_buffer`)
  - Temporary high-throughput storage for sensor metrics.
* **`TransactionalOutbox`** (`eam_transactional_outbox`)
  - Composite index on `(status, created_at)`.
* **`KafkaEventInbox`** (`eam_kafka_event_inbox`)

### B. Decoupling Rules ($C_e = 0$)
* All references to external domains (e.g. FM Capital Asset ID, HR Employee ID) must store raw `uuid` strings instead of direct joins.

---

## 3. Kafka Event Interaction Matrix

### Egress (Events Produced)
* **`eam.machine.offline` / `eam.machine.online`**: Published when a work order reports a critical breakdown or a technician resolves the incident. Tells MFG to divert or resume operations.
* **`eam.workorder.spares_requested`**: Published when a work order requests replacement spare parts. Signals SCM to reserve or procure them.

### Ingress (Events Consumed)
* **`scm.asset.received`**: Ingests new asset deliveries from SCM and registers them in the equipment table.
* **`fm.asset.capitalized`**: Updates an equipment record to link it to the fixed-asset depreciation registry.
* **`hr.employee.created`**: Validates technician roles and credentials inside EAM.

---

## 4. Scope & Implementation Checklist

### Phase 1: Clean & Align Repositories
- [x] Refactor `internal/business/domain/repository.go` to define standard CRUD interfaces for the 7 CDD entities.
- [x] Create `internal/data/sql/db.go` and `internal/data/sql/models.go` declaring GORM structures, unique indexes, soft delete fields, and mappers.
- [x] Implement `internal/data/sql/sql_repos.go` wrapping GORM PostgreSQL.
- [x] Implement `internal/data/memory/memory_repos.go` implementing thread-safe mock structures.

### Phase 2: Service Layer & Transactional Telemetry
- [x] Implement `EquipmentService`, `MaintenanceService`, and `TelemetryIngestionService` inside `internal/business/service/service.go`.
- [x] Implement the secure two-phase telemetry drain block (`flushStagedMetricsToTimeSeriesStore`) using GORM transactional locking.
- [x] Implement outbox/inbox relay and reliable messaging interfaces.

### Phase 3: REST Handlers & API Routes
- [x] Create a single unified `EamHandler` inside `internal/api/handlers/`.
- [x] Update `internal/api/routes/routes.go` to configure EAM routing paths.

### Phase 4: Kafka Event Streams & Integration
- [x] Update `internal/data/kafka/consumer.go` to idempotently consume `scm.asset.received`, `fm.asset.capitalized`, and `hr.employee.created`.
- [x] Re-wire `cmd/main.go` to bootstrap the EAM microservice.

### Phase 5: Verification & Tests
- [x] Write GORM SQLite unit tests inside `internal/business/service/service_test.go` verifying equipment creation, work order routing, and telemetry buffer flushing.
- [x] Verify that `go build ./...` and `go test ./...` pass successfully.

---

## 5. Definition of Done
- [x] `eam.cdd` is fully reconciled and implemented in code.
- [x] The EAM service compiles cleanly and passes all test suites.
