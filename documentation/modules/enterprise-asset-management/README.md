# Enterprise Asset Management (EAM) Core Module

A highly scalable, multi-tenant physical plant orchestration and asset maintenance engine built using Go, Gin, GORM, and PostgreSQL. It enforces **zero direct database coupling ($C_e = 0$)** to downstream operational contexts by communicating strictly via asynchronous Kafka event streams and referencing external entities (like Finance Assets or Employees) using primitive `uuid` tracking tokens.

---

## 1. Bounded Context & Topology

The EAM module manages physical equipment uptime, break-fix maintenance ticket routing, preventative maintenance scheduling, and high-throughput machine telemetry logging.

```mermaid
graph TB
    subgraph "EAM Bounded Context"
        FAC[Facility] --> EQ[Equipment]
        EQ --> WO[MaintenanceWorkOrder]
        EQ --> PS[PreventativeSchedule]
        EQ --> TIB[TelemetryIngestBuffer]
        
        TXO[TransactionalOutbox]
        KFI[KafkaEventInbox]
    end

    subgraph "External Systems (Event Integration)"
        SCM[SCM Service] -. scm.asset.received .-> KFI
        FM[Finance Service] -. fm.asset.capitalized .-> KFI
        HR[HR Service] -. hr.employee.created .-> KFI
        
        TXO -. eam.machine.offline .-> MFG[Manufacturing Execution]
        TXO -. eam.machine.online .-> MFG
        TXO -. eam.workorder.spares_requested .-> SCM
    end
```

---

## 2. Standardized Quantitative Metrics

| Metric Category | Metric Code | Target Value | Realized Value | Status |
| --- | --- | --- | --- | --- |
| **Efferent Coupling** | $C_e$ | `0` | `0` (Zero direct DB joins or compile-time dependencies to other services) | ✅ Compliant |
| **Afferent Coupling** | $C_a$ | `2` | `2` (Consume streams from SCM, FM, and HR, producing to MFG and SCM) | ✅ Compliant |
| **Instability Index** | $I$ | `0.0` | `0.0` (Stable core; changes in execution services do not break EAM logic) | ✅ Compliant |

---

## 3. Domain Models (The 7 CDD Entities)

These models map directly to the GORM database structures inside EAM PostgreSQL schema:

| Entity | DB Table | Indexing / Properties | Description |
|---|---|---|---|
| `Facility` | `eam_facilities` | Primary Key `id` | Represents a physical location/factory plant where assets are housed. |
| `Equipment` | `eam_equipment` | Unique Index `asset_tag`, Soft Delete `deleted_at` | Physical plant machinery. Supports GORM native soft deletion. |
| `MaintenanceWorkOrder` | `eam_work_orders` | Unique Index `ticket_number` | Tracks repair work, assignment details, and breakdown duration. |
| `PreventativeSchedule` | `eam_pm_schedules` | Primary Key `id` | Stores interval-based preventative maintenance schedules for equipment. |
| `TelemetryIngestBuffer` | `eam_telemetry_ingest_buffer` | Primary Key `id` | Temporary buffer used to stage telemetry sensor logs before database purging. |
| `TransactionalOutbox` | `eam_transactional_outbox` | Composite Index `(status, created_at)` | Holds messages published transactionally as part of model mutations. |
| `KafkaEventInbox` | `eam_kafka_event_inbox` | Primary Key `event_id` | Inbox table for message deduplication to guarantee idempotent ingestion. |

---

## 4. Go Service Interfaces

### `EquipmentService`
Manages plant infrastructure setup and capital assets registry.
- `CreateFacility(ctx, legalEntityId, name, address)`
- `RegisterEquipment(ctx, legalEntityId, facilityId, assetTag, name, serialNumber)`
- `UpdateEquipmentStatus(ctx, tx, equipmentId, newStatus)` (Transactional status mutations emitting `eam.machine.offline`/`online` outbox logs)
- `AssociateFinancialAsset(ctx, equipmentId, financialAssetId)`
- `FetchTargetTenantAssets(ctx, legalEntityId, status)`

### `MaintenanceService`
Controls break-fix and preventative maintenance lifecycles.
- `FileMachineIncident(ctx, legalEntityId, equipmentId, reportedBy, title, priority)` (Emits `eam.machine.offline` when priority is critical or high)
- `RouteToTechnician(ctx, workOrderId, techHrId)`
- `TransitionToActiveState(ctx, workOrderId)`
- `FinalizeResolution(ctx, workOrderId, resolutionNotes)` (Emits `eam.machine.online` and sets equipment status to `ONLINE` upon completion)
- `RequestSpares(ctx, workOrderId, componentDetails)` (Emits `eam.workorder.spares_requested` downstream to SCM)
- `ProcessCronSchedulerLookups(ctx, targetDate)` (Generates PM work orders for PM schedules that are due)

### `TelemetryIngestionService`
Processes and flushes machine sensor metrics.
- `QueueSensorMetrics(ctx, legalEntityId, equipmentId, sensorKey, value)`
- `FlushStagedMetricsToTimeSeriesStore(ctx, tx, batchLimit)` (Two-phase drain with `UPDATE SKIP LOCKED` database transaction safety)

### `ReliableMessagingService`
De-duplicates incoming events to ensure exactly-once processing (idempotency).
- `IsEventProcessed(ctx, eventId)`
- `CommitInboundEvent(ctx, eventId, eventType, payload)`
- `PushToOutbox(ctx, tx, eventType, aggregateId, payload)`
- `ExecuteIdempotentTransaction(ctx, eventId, eventType, payload, businessRoutine)`

### `OutboxRelayWorker`
Relays messages from the transactional outbox to the Kafka broker.
- `GetUnsentMessages(ctx, limit)`
- `UpdateOutboxStatus(ctx, tx, outboxId, status)`

---

## 5. REST API Endpoints

All endpoints use JSON payload mapping:

### Infrastructure & Registry
```http
POST   /api/v1/eam/facilities                          # Create a new plant facility
POST   /api/v1/eam/equipment                           # Register a new piece of equipment
GET    /api/v1/eam/equipment                           # Query equipment list for a tenant
PUT    /api/v1/eam/equipment/:id/status                # Manually override equipment status
PUT    /api/v1/eam/equipment/:id/finance-asset         # Link financial capitalized asset ID
```

### Break-Fix Maintenance & Work Orders
```http
POST   /api/v1/eam/work-orders                         # File an incident ticket
PUT    /api/v1/eam/work-orders/:id/route               # Assign work order to technician
POST   /api/v1/eam/work-orders/:id/start               # Start work order execution
POST   /api/v1/eam/work-orders/:id/resolve             # Submit resolution and restore machine status
```

### Telemetry Logs
```http
POST   /api/v1/eam/telemetry/sensor-metrics            # Queue high-frequency telemetry logs
POST   /api/v1/eam/telemetry/flush                     # Flush telemetry buffer to time-series store
```

---

## 6. Asynchronous Event Streams

### Ingress Events Consumed (Idempotently)
- `scm.asset.received`: Triggered when inventory receives a new physical asset; auto-registers the equipment record inside EAM.
- `fm.asset.capitalized`: Links the equipment record to the fixed asset registry in Financial Management.
- `hr.employee.created`: Validates and syncs technician profiles in the EAM database.

### Egress Events Produced (via Transactional Outbox)
- `eam.machine.offline`: Published when a critical/high breakdown work order is filed or equipment status is updated to broken. Tells MFG to divert shop floor routes.
- `eam.machine.online`: Published when a technician submits a resolution notes and restores the machine. Tells MFG to resume default scheduling.
- `eam.workorder.spares_requested`: Emitted when repairs need materials or spare parts from Supply Chain Management (SCM).
