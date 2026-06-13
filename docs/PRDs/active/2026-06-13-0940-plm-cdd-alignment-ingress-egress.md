# PRD: PLM Service CDD Alignment & Event Ingress/Egress Integration

**PRD ID**: PRD-2026-06-13-0940  
**Date**: 2026-06-13  
**Status**: Approved (Implemented)  
**Parent Initiative**: Technical Documentation Sync & Domain Parity  
**Target Coverage**: 100% CDD event alignment, zero architectural drift  

---

## 1. Objective & Architectural Context

The Product Lifecycle Management (PLM) module operates as the absolute design authority for materials, Bill of Materials (BOM), and Engineering Change Orders (ECO). To reinforce this boundary:
1. The PLM service must preserve an Efferent Coupling metric of zero ($C_e = 0$) at the database tier.
2. The domain contract (`plm.cdd`) must be aligned with the actual event fabric, adding support for inbound QMS inspection failures (`qms.inspection.failed`) and EAM offline alerts (`eam.machine.offline`).
3. The event pipeline must process inbound ingress events via an idempotent consumer model registering history logs in the `plm_kafka_event_inbox` table.
4. Go domain models must be cleanly generated, and background handlers/services must compile and pass tests cleanly.

---

## 2. Technical Scope & Event Matrix

### A. CDD & Domain Model Alignments
* Reconcile `services/plm-service/contracts/plm.cdd` to include the target consumer events:
  - `qms.inspection.failed`
  - `eam.machine.offline`
* Keep all existing active customer and product structures.

### B. Event Pipeline Ingress
* Update `services/plm-service/internal/data/kafka/consumer.go` to subscribe to and process:
  - `qms.inspection.failed`: Log when quality inspection fails for a material and trigger corresponding logic (simulated logic mapping or logging).
  - `eam.machine.offline`: Log machinery disruptions and record workspace warning flags.

---

## 3. Scope & Implementation Checklist

### Phase 1: Reconcile CDD Contract
- [x] Update `services/plm-service/contracts/plm.cdd` to define `qms.inspection.failed` and `eam.machine.offline` consumer events.
- [x] Add the corresponding topics to `internal/business/domain/event_topics.go`.

### Phase 2: Domain Model & Code Generation
- [x] Regenerate domain models using the `cdd-engine`.
- [x] Verify that model structs (`MaterialMaster`, `BomHeader`, `BomLine`, `EngineeringChangeOrder`) remain fully consistent with database schema definitions.

### Phase 3: Ingress Handler Implementation
- [x] Update `KafkaConsumer` in `consumer.go` to listen to the new QMS and EAM topics.
- [x] Add parsing, logging, and transactional idempotency check handlers for both events.

### Phase 4: Verification & Compilation
- [x] Verify that `go build ./...` compiles cleanly without any errors.
- [x] Run the test suite using `go test ./...` to ensure all tests pass.

### Phase 5: Technical Documentation Sync
- [x] Document the PLM integration topologies, ingress/egress pipelines, and ATAM trade-off matrix in the PLM module README.

---

## 4. Definition of Done
- [x] `plm.cdd` fully declares the ingress and egress event streams.
- [x] PLM consumer processes QMS and EAM events without thread locks or failures.
- [x] Go models compile cleanly and `go test ./...` passes successfully.
