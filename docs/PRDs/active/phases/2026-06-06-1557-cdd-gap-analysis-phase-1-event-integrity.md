# ERP System CDD Gap Analysis — Phase 1: Event Integrity

**Source PRD**: docs/PRDs/active/2026-06-06-1557-cdd-gap-analysis.md
**PRD ID**: PRD-2026-06-06-1557
**Phase**: 1 of 6
**Status**: Completed
**Created**: June 06, 2026

---

## Objective

Fix event integrity: ensure all CDD-defined producer events are actually published, remove dead consumer subscriptions for events that have no publisher, fix topic naming inconsistencies, and add at-minimum error logging to Kafka publishes.

## Scope

### In Scope

- Add publish calls for 31 defined-but-unpublished producer event constants
- Remove or postpone consumer subscriptions for 16 events with no publisher
- Fix topic name mismatch: FM consumer `crm.sale.completed` → match CRM's actual `crm.sales.order.confirmed` topic
- Migrate fm-service from hardcoded topic strings to constants
- Log Kafka publish errors instead of discarding with `_`
- Add dead-letter queue pattern for consumer errors

### Out of Scope

- Retry logic or delivery guarantees (Kafka producer config stays fire-and-forget)
- Event schema validation
- Exactly-once semantics

---

## Inputs

| Input | Source | Notes |
| ----- | ------ | ----- |
| Event topic constants | `services/*/internal/business/domain/event_topics.go` | Defined but not all published |
| CDD event definitions | `services/*/contracts/*.cdd` | Source of truth for which events should exist |
| Kafka consumer impls | `services/*/internal/data/kafka/consumer.go` | Subscriptions with dead publishers |
| Kafka publisher impls | `services/*/internal/data/kafka/publisher.go` | Fire-and-forget pattern |

## Dependencies

| Dependency | Type | Required Before | Notes |
| ---------- | ---- | --------------- | ----- |
| Phase 0.5 | Code | Phase 1 start | Error logging pattern must be established before adding new publish calls |
| Phase 0 | Code | Phase 1 (partial) | Only needed for events that fire from newly-added service methods (e.g., `ConsumeMaterials` → `mfg.material.consumed`). Most events can be added independently. |

---

**Note on task numbering:** Tasks 1-7 here correspond to the original Phase 1 plan. Task 8 (error logging) was moved to Phase 0.5 (parallel track, higher urgency). Task 9 (DLQ) was moved to Phase 6 (new feature, separate PRD). Task numbers have been adjusted accordingly.

## Implementation Tasks

### Task 1: Fix SCM event publishing gap (20 missing publishes, split by domain)

**Description:** SCM has 22 event constants but only publishes 2. Add publish calls for the remaining 20 CDD-defined events. Split into sub-tasks by domain for manageable units.

**1a — Product events (3 events):** `scm.product.created`, `scm.product.updated`, `scm.product.discontinued`
- Publish from ProductManagementService methods (CreateProduct, UpdateProduct)
- Files: `services/scm-service/internal/business/service/product_service.go`

**1b — Inventory events (5 events):** `scm.inventory.received`, `scm.inventory.shipped`, `scm.inventory.adjusted`, `scm.inventory.low.stock`, `scm.inventory.out.of.stock`
- Publish from InventoryService methods (AdjustInventory, ReserveStock, ReleaseReservation)
- Files: `services/scm-service/internal/business/service/inventory_service.go`

**1c — Purchase order events (3 events):** `scm.purchase.order.sent`, `scm.purchase.order.received`, `scm.purchase.order.cancelled`
- Publish from PurchaseOrderService methods (SendPurchaseOrder, UpdatePurchaseOrderStatus)
- Files: `services/scm-service/internal/business/service/purchase_order_service.go`

**1d — Vendor events (3 events):** `scm.vendor.created`, `scm.vendor.updated`, `scm.vendor.performance.evaluated`
- Publish from SupplierManagementService methods (CreateSupplier, UpdateSupplier, evaluate)
- Files: `services/scm-service/internal/business/service/supplier_service.go`

**1e — Shipment events (4 events):** `scm.shipment.created`, `scm.shipment.dispatched`, `scm.shipment.delivered`, `scm.shipment.delayed`
- Publish from WarehouseService methods (CreateShipment, UpdateShipmentStatus)
- Files: `services/scm-service/internal/business/service/warehouse_service.go`

**1f — Other events (2 events):** `scm.training.required`, `scm.material.delivered`
- Publish from wherever the triggering logic lives (receiving goods triggers material.delivered)
- Files: `services/scm-service/internal/business/service/warehouse_service.go` or `inventory_service.go`

**Acceptance Criteria:**
- All 22 SCM events are published from the correct service method
- All publishers wired through `main.go`

### Task 2: Fix PM event publishing gap (8 missing publishes)

**Events to add:** `prj.project.delayed`, `prj.task.overdue`, `prj.resource.released`, `prj.resource.overallocated`, `prj.time.rejected`, `prj.expense.rejected`, `prj.milestone.achieved`, `prj.milestone.delayed`

**Acceptance Criteria:**
- All 25 PM events published

### Task 3: Fix HR event publishing gap (5 missing publishes)

**Events to add:** `hr.payroll.failed`, `hr.certification.earned`, `hr.skill.acquired`, `hr.employee.available`, `hr.employee.skills.updated`

**Acceptance Criteria:**
- All 22 HR events published

### Task 4: Fix CRM event publishing gap (3 missing publishes)

**Events to add:** `crm.email.opened`, `crm.email.clicked`, `crm.sales.order.received`

**Acceptance Criteria:**
- All 28 CRM events published

### Task 5: Remove dead consumer subscriptions (16 topics)

**Description:** Services consume events that no service publishes. Remove these consumer subscriptions or add TODO comments explaining when they'll be connected.

**Dead subscriptions to remove/postpone:**
- FM: `scm.invoice.received` (no publisher)
- SCM: `scm.material.received`, `scm.inventory.updated`, `scm.inventory.available`, `scm.shipment.delivered`, `scm.material.delivered`, `scm.training.required` (no publishers)
- SCM: `fin.vendor.payment.processed` (no publisher — FM publishes `fin.payment.processed`)
- HR: `fin.budget.allocated` (no publisher), ~~`mfg.production.scheduled`~~ → ✅ **M publishes this, keep alive**
- M: `scm.material.received`, `scm.inventory.updated`, `fin.cost.budget.allocated`, `hr.employee.scheduled` (no publishers)
- CRM: `scm.inventory.available`, `scm.shipment.delivered`, `fin.credit.check.completed`, `hr.employee.performance` (no publishers)
- PM: `hr.employee.available`, `hr.employee.skills.updated`, `fin.budget.approved`, `crm.sales.order.received`, `mfg.custom.production.completed` (no publishers)

**Strategy:** Comment out handler registrations with a `// TODO: connect when <service> publishes <topic>` comment. Keep the handler functions for when topics become active.

**Acceptance Criteria:**
- No consumer is subscribed to a topic that has zero publishers
- Commented subscriptions have clear TODO notes with the expected source service

### Task 6: Fix topic naming consistency

**Description:** FM consumer expects `crm.sale.completed` but CRM publishes `crm.sales.order.confirmed`. Fix the consumer to use the correct topic.

**Also check:** FM publishes `fin.payment.processed` and `fin.payment.received` — SCM expects `fin.vendor.payment.processed`. If this is a separate event, add it as a new constant + publish call.

**Acceptance Criteria:**
- All cross-service topic names match publisher ↔ consumer

---

## Verification

```bash
# Check all events are published
rg '\.Publish\(' services/ --type go

# Check no hardcoded strings in fm-service
rg 'fin\.' services/fm-service --type go

# Check no discarded errors
rg '_ = .*Publish\(' services/ --type go | wc -l
# Should be 0
```

---

## Risks

| Risk | Likelihood | Mitigation |
| ---- | ---------- | ---------- |
| Adding publishes changes service coupling | Medium | All events are already in CDD — this is catching up to spec |
| Dead subscriptions removal breaks tests | Low | No test asserts consumer behavior (only 1 test file exists) |
| Topic renames cause silent missed events | Medium | Rename consumer side to match publisher — not publisher side |

## Definition of Done

- [x] Task 1a: SCM product events published (3)
- [x] Task 1b: SCM inventory events published (5)
- [x] Task 1c: SCM purchase order events published (3)
- [x] Task 1d: SCM vendor events published (3)
- [x] Task 1e: SCM shipment events published (4)
- [x] Task 1f: SCM other events published (2)
- [x] Task 2: PM publishes all 25 events
- [x] Task 3: HR publishes all 22 events
- [x] Task 4: CRM publishes all 28 events
- [x] Task 5: Zero dead consumer subscriptions (14 truly dead, 1 kept: `mfg.production.scheduled`)
- [x] Task 6: All topic names consistent across services
- [x] Task 7: fm-service uses constants not strings
- [x] `make build` passes for all services
- [x] Note: error logging handled by Phase 0.5, DLQ handled by Phase 6
