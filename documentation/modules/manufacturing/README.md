# Manufacturing Execution (MFG) Core Module

A highly optimized, multi-tenant Shop Floor Execution core built with contract-driven Go, Gin, GORM, and PostgreSQL. It enforces **zero compile-time database coupling ($C_e = 0$)** by relying exclusively on asynchronous event streams and primitive `uuid` tracking identifiers for integration with external modules (such as PLM, SCM, FM, and HR).

---

## 1. Bounded Context & Topology

The Manufacturing module focuses purely on shop-floor orchestration, equipment routing, yield logging, and material consumption logs. All cross-module entities (BOM, Sourcing, Costing, Equipment Master, employee profiles) are kept strictly outside the database boundary and referenced via primitives.

```mermaid
graph TB
    subgraph "MFG Bounded Context"
        WC[WorkCenter] --> RS[RoutingStation]
        WO[WorkOrder] --> WORS[WorkOrderRoutingState]
        WO --> MCL[MaterialConsumptionLog]
        WO --> PYL[ProductionYieldLog]
        
        TXO[TransactionalOutbox]
        KFI[KafkaEventInbox]
    end

    subgraph "External Systems (Event Integration)"
        PLM[PLM Service] -. plm.bom.released .-> KFI
        QMS[QMS Service] -. qms.inspection.passed/failed .-> KFI
        EAM[EAM Service] -. eam.offline .-> KFI
        
        TXO -. mfg.production.started .-> FM[Finance / Ledger]
        TXO -. mfg.material.consumed .-> SCM[Supply Chain]
    end
```

---

## 2. Standardized Quantitative Metrics

| Metric Category | Metric Code | Target Value | Realized Value | Status |
| --- | --- | --- | --- | --- |
| **Efferent Coupling** | $C_e$ | `0` | `0` (Zero direct DB/code coupling to other microservices) | ✅ Compliant |
| **Afferent Coupling** | $C_a$ | `0` | `0` (External services communicate solely via Kafka) | ✅ Compliant |
| **Instability Index** | $I$ | `0.0` | `0.0` (Highly stable, resilient to external system changes) | ✅ Compliant |

---

## 3. Domain Models (The 8 CDD Entities)

These models map directly to the `mfg_` table schemas in PostgreSQL:

| Entity | DB Table | Partitioning / Indexes | Description |
|---|---|---|---|
| `WorkCenter` | `mfg_work_centers` | Unique Composite `(legal_entity_id, work_center_code)` | Core shop floor work area (e.g., machining, assembly line). |
| `RoutingStation` | `mfg_routing_stations` | Unique Composite `(work_center_id, routing_code)` | Individual production step/station inside a work center. |
| `WorkOrder` | `mfg_work_orders` | Unique Composite `(legal_entity_id, work_order_number)` | Production execution task for a specific target quantity. |
| `WorkOrderRoutingState` | `mfg_work_order_routing_states` | Unique Composite `(work_order_id, current_station_id)` | Tracking state machine for routing gates, including rework loops. |
| `MaterialConsumptionLog` | `mfg_material_consumption_logs` | Partitioned Monthly `(consumed_at)`, Unique `(id, work_order_id, material_id, consumed_at)` | Tracks physical material usage at routing stations. |
| `ProductionYieldLog` | `mfg_production_yield_logs` | Partitioned Monthly `(recorded_at)`, Unique `(id, work_order_id, recorded_at)` | Tracks scrap vs good production count at routing stations. |
| `TransactionalOutbox` | `mfg_transactional_outbox` | Index Composite `(status, created_at)` | Guaranteed event delivery broker (Outbox Pattern). |
| `KafkaEventInbox` | `mfg_kafka_event_inbox` | Primary Key `event_id` | Deduplication inbox to guarantee idempotent consumption. |

---

## 4. Bounded Context Interfaces (Go Services)

### `FloorConfigurationService`
Manages work center setups and station assignments.
- `EstablishWorkCenter(ctx, legalEntityId, code, name)`
- `AppendStationToCenter(ctx, workCenterId, routingCode, stationType, equipmentId, setupTime, runTime)`

### `WorkOrderExecutionService`
Coordinates the state transitions and routing of work orders.
- `InstantiateWorkOrder(ctx, legalEntityId, materialId, bomHeaderId, qtyTarget, start, end)`
- `TransitionWorkOrderState(ctx, workOrderId, currentState, targetState)` (Emits `mfg.production.started` / `mfg.work_order.completed`)
- `RerouteWorkOrderStation(ctx, workOrderId, currentStationId, targetStationId, isRework)`

### `ShopFloorTelemetryService`
Captures raw operational logs from the shop floor.
- `RecordBulkMaterialConsumption(ctx, legalEntityId, workOrderId, lines)` (Emits `mfg.material.consumed`)
- `CommitProductionYield(ctx, legalEntityId, workOrderId, stationId, qtyGood, qtyScrap, operatorHrId)` (Emits `mfg.yield.produced`)

### `OutboxRelayWorker`
Dispatches transactional outbox events to the Kafka cluster.
- `GetUnsentMessages(ctx, limit)`
- `LogProcessingAttempt(ctx, outboxId, currentRetries, errorNotes)`
- `UpdateOutboxStatus(ctx, outboxId, status)`

### `ReliableMessagingService`
De-duplicates incoming events to ensure exactly-once processing (idempotency).
- `IsEventProcessed(ctx, eventId)`
- `ExecuteIdempotentTransaction(ctx, eventId, eventType, payload, businessRoutine)`

---

## 5. REST API Endpoints

All actions require JSON payloads and return JSON responses.

### Work Centers & Routing
```http
POST   /api/v1/mfg/work-centers                       # Establish work center
POST   /api/v1/mfg/work-centers/:id/stations          # Append station to work center
```

### Work Order Execution
```http
POST   /api/v1/mfg/work-orders                        # Instantiate work order
POST   /api/v1/mfg/work-orders/:id/transition         # Transition state machine
POST   /api/v1/mfg/work-orders/:id/reroute            # Reroute stations or start rework
```

### Telemetry Logs
```http
POST   /api/v1/mfg/work-orders/:id/consumption        # Record material consumption
POST   /api/v1/mfg/work-orders/:id/yield              # Commit yield (good/scrap counts)
```

---

## 6. Asynchronous Event Streams

### Emitted Events (Emitters)
These events are written to the transactional outbox:

| Event | Event Schema | Description |
|---|---|---|
| `mfg.production.started` | `{event_id, legal_entity_id, work_order_id, material_id, timestamp}` | Fired when a work order transitions to `IN_PROGRESS`. |
| `mfg.material.consumed` | `{event_id, legal_entity_id, work_order_id, items: List<ConsumedItemPayload>, timestamp}` | Fired when material consumption logs are recorded. |
| `mfg.yield.produced` | `{event_id, legal_entity_id, work_order_id, routing_station_id, quantity_good, quantity_scrap, operator_hr_id, timestamp}` | Fired when yield logs are saved, updating total output. |
| `mfg.work_order.completed` | `{event_id, legal_entity_id, work_order_id, material_id, quantity_produced, timestamp}` | Fired when a work order transitions to `COMPLETED`. |

### Consumed Events (Consumers)
Idempotently processed events triggering internal state updates:

| Event | Logic |
|---|---|
| `plm.bom.released` | Logs engineering release version. |
| `qms.inspection.passed` | Transitions the associated `WorkOrder` status from `IN_PROGRESS` to `COMPLETED`. |
| `qms.inspection.failed` | Transitions the associated `WorkOrder` status from `IN_PROGRESS` to `ON_HOLD` for rework. |
| `eam.machine.offline` | Places the work order executing at the offline machine `ON_HOLD`. |
