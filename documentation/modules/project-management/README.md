# Project Management (PRJ) Core Module

A highly optimized, multi-tenant Professional Services and Project Tracking core built with contract-driven Go, Gin, GORM, and PostgreSQL. It enforces **zero compile-time database coupling ($C_e = 0$)** by relying exclusively on asynchronous event streams and primitive `uuid` tracking identifiers for integration with external modules (such as HR, CRM, and FM).

---

## 1. Bounded Context & Topology

The Project Management module focuses purely on project lifecycle administration, Work Breakdown Structure (WBS) trees (phases, tasks, milestones), and timesheet validation. All cross-module entities (employees, customers, legal entities) are kept strictly outside the database boundary and referenced via primitives.

```mermaid
graph TB
    subgraph "PRJ Bounded Context"
        PROJ[Project] --> WBS[WbsNode]
        WBS --> TL[TimeLog]
        
        TXO[TransactionalOutbox]
        KFI[KafkaEventInbox]
    end

    subgraph "External Systems (Event Integration)"
        HR[HR Service] -. hr.employee.created/terminated .-> KFI
        CRM[CRM Service] -. crm.sales.order.confirmed .-> KFI
        
        TXO -. prj.time.logged .-> FM[Finance / Ledger]
        TXO -. prj.time.logged .-> CRM
        TXO -. prj.milestone.achieved .-> SCM[Supply Chain]
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

## 3. Domain Models (The 5 CDD Entities)

These models map directly to the `prj_` table schemas in PostgreSQL:

| Entity | DB Table | Partitioning / Indexes | Description |
|---|---|---|---|
| `Project` | `prj_projects` | Unique Composite `(legal_entity_id, project_code)` | Core project metadata with multi-tenant isolation. |
| `WbsNode` | `prj_wbs_nodes` | Unique Composite `(project_id, node_code)` | Represents tasks, phases, or milestones in a recursive tree structure. |
| `TimeLog` | `prj_time_logs` | Partitioned Monthly `(work_date)`, Unique `(wbs_node_id, employee_id, work_date)` | Timesheet entries linked to WBS nodes. Prevents duplicate submissions. |
| `TransactionalOutbox` | `prj_transactional_outbox` | Index Composite `(status, created_at)` | Guaranteed event delivery broker (Outbox Pattern). |
| `KafkaEventInbox` | `prj_kafka_event_inbox` | Primary Key `event_id` | Deduplication inbox to guarantee idempotent consumption. |

---

## 4. Bounded Context Interfaces (Go Services)

### `ProjectTrackingService`
Manages project initialization and status transitions.
- `InitializeProject(ctx, legalEntityId, customerId, code, name, billingMethod, start)`
- `TransitionProjectStatus(ctx, projectId, newStatus)`

### `WbsStructureService`
Manages the Work Breakdown Structure recursive trees.
- `AppendWbsNode(ctx, projectId, parentNodeId, code, title, nodeType, hours)`
- `DeclareNodeCompletion(ctx, nodeId, completionHrId)` (Emits `prj.milestone.achieved` if the node is a completed milestone)
- `FetchProjectTree(ctx, projectId)`

### `TimeTrackingService`
Coordinates the submission and approval of time logs.
- `LogOperationalHoursBulk(ctx, legalEntityId, employeeId, logs)`
- `ProcessTimesheetApproval(ctx, timeLogIds, approverHrId)` (Emits `prj.time.logged` grouped by project)

### `OutboxRelayWorker`
Dispatches transactional outbox events to the Kafka cluster.
- `GetUnsentMessages(ctx, limit)`
- `UpdateOutboxStatus(ctx, outboxId, status)`

### `ReliableMessagingService`
De-duplicates incoming events to ensure exactly-once processing (idempotency).
- `IsEventProcessed(ctx, eventId)`
- `ExecuteIdempotentTransaction(ctx, eventId, eventType, payload, businessRoutine)`

---

## 5. REST API Endpoints

All actions require JSON payloads and return JSON responses.

### Projects
```http
POST   /api/v1/projects                         # Initialize project
PUT    /api/v1/projects/:id/status              # Transition project status
GET    /api/v1/projects                         # List projects
GET    /api/v1/projects/:id                     # Get project detail
```

### WBS Structure
```http
POST   /api/v1/projects/:id/wbs                 # Append WBS node to project tree
PUT    /api/v1/wbs/:node_id/complete            # Declare WBS node completion
GET    /api/v1/projects/:id/wbs                 # Fetch full project tree
```

### Time Tracking
```http
POST   /api/v1/time-logs/bulk                   # Log bulk operational hours
POST   /api/v1/time-logs/approve                # Process timesheet approvals
```
