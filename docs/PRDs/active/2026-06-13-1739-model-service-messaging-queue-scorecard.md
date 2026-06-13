# PRD & Scorecard: Model Service & Messaging Queue Architectural Alignment

This Product Requirement Document (PRD) defines the standards, audits the coding implementation, and outlines the scorecard for the **Model Service (Database Tier/GORM)** and the **Messaging Queue (Transactional Outbox/Kafka)** across the ERP platform. 

---

## 1. Executive Summary

To achieve enterprise-grade stability and resilience, all microservices must align with the **Contract-Driven Development (CDD)** specifications. This document focuses on two critical pillars:
1. **Model Service (GORM DB/Repository Layer):** Ensuring zero compile-time efferent coupling ($C_e = 0$), strict tenant data isolation, and optimistic concurrency locking.
2. **Messaging Queue (Outbox & Inbox Patterns):** Ensuring atomic event logging (Outbox Pattern) and idempotent event processing (Inbox Pattern) to guarantee "exactly-once" delivery semantics over Kafka.

Following our recent comprehensive refactoring of the Manufacturing (`mfg-service`) service, it serves as the reference architecture (**scoring 10/10**) for other microservices to copy.

---

## 2. Deep-Dive Coding Analysis (Mfg Reference Implementation)

### A. Model Service (GORM & DB Layer)
The database mapping in [models.go](file:///Users/sithuhlaing/Projects/erp-system/services/mfg-service/internal/data/sql/models.go) implements the following advanced design patterns:
* **Decoupled Entities:** All outbound relationships are represented as primitive `string` UUID tokens (e.g. `MaterialID` or `EquipmentID`) rather than hard GORM associations to other services' tables. This guarantees $C_e = 0$.
* **Two-Way Translation:** We enforce strict separation of GORM DB models from Domain models. Translation functions `ToDomain` and `FromDomain` map types bidirectionally, preventing database tags or serialization details from leaking into business logic.
* **Optimistic Locking:** The `WorkOrder` entity maps GORM `version` increments inside transaction updates. If the target row's version has changed in the database, the transaction aborts with a concurrency error, preventing "dirty write" overwrites.

### B. Messaging Queue (Outbox & Inbox)
Our Kafka integration in [consumer.go](file:///Users/sithuhlaing/Projects/erp-system/services/mfg-service/internal/data/kafka/consumer.go) and service layer in [mfg_services.go](file:///Users/sithuhlaing/Projects/erp-system/services/mfg-service/internal/business/service/mfg_services.go) implements:
* **Atomic Transactional Outbox:** Outbox events (such as `MfgProductionStartedEvent`) are written to the `mfg_transactional_outbox` table using the *same GORM transaction* that commits the business state mutation. This guarantees events are never lost if the server crashes right after a commit.
* **Idempotent Kafka Inbox:** Incoming Kafka events are routed through the `ReliableMessagingService` which wraps the handler logic inside `ExecuteIdempotentTransaction`. Before executing, it queries `mfg_kafka_event_inbox` for the `event_id`. If it has been processed successfully, the message is skipped. If it failed, it retries. If it succeeds, it writes a `SUCCESS` inbox log. This guarantees exactly-once processing even if Kafka delivers duplicate messages.

---

## 3. Quantitative Scorecard & Evaluation Criteria

We evaluate both pillars on a **10-point scale** across several dimensions:

### Matrix 1: Model Service (GORM Persistence Layer)

| Score | Database Boundary & Coupling | Model Isolation | Concurrency & Multi-Tenancy |
| :---: | :--- | :--- | :--- |
| **9-10** | **$C_e = 0$:** Direct table joins/foreign keys to external services do not exist. References are primitives. | **100% Segregated:** DB structures and domain models are separate; bidirectional translation functions are implemented. | **Fully Safe:** Multi-tenant shielding is applied. Concurrency is handled via optimistic locks (`version` checks). |
| **6-8** | **Loose Coupling:** References are primitives, but some GORM models still reference external domain models in imports. | **Partial Leakage:** Domain models double as GORM structs directly, but packages are separate. | **Basic Safety:** Multi-tenant columns exist, but optimistic locking is absent. |
| **0-5** | **Hard Coupled ($C_e > 0$):** Joins across microservice schemas; hard foreign keys; direct DB cross-talk. | **Monolithic Models:** Domain logic directly depends on DB schemas with no boundary. | **Unsafe:** No tenant validation or concurrency shields. |

### Matrix 2: Messaging Queue (Reliable Event Fabric)

| Score | Event Atomicity (Outbox) | Idempotent Ingestion (Inbox) | Topic Alignment |
| :---: | :--- | :--- | :--- |
| **9-10** | **Guaranteed Atomicity:** Events are persisted in an Outbox table within the same GORM transaction as business state mutations. | **Deduplication Inbox:** All consumer messages are stored in an Inbox table; processed state is validated before handler execution. | **100% CDD Match:** Emitters and Consumers map exactly to the `.cdd` contract registry. |
| **6-8** | **Best Effort:** Events are published directly to Kafka in the API handler without a transactional outbox table. | **Local Cache:** Deduplication is done in-memory (e.g. Redis/Memcached) but lacks ACID transaction isolation with the database. | **Loose Match:** Some events are missing or named incorrectly compared to the CDD spec. |
| **0-5** | **No Guarantee:** Events are published inline; crashes result in lost events or inconsistent business states. | **No Deduplication:** Duplicate Kafka messages result in double-commit bugs and corrupted balances. | **Mismatched:** Handlers listen to unapproved topics; payloads are unstructured. |

---

## 4. Current Scorecard Ratings (Finished Modules)

Here is the audit rating for modules that have completed their CDD reconciliation:

### A. Manufacturing (`mfg-service`)
* **Model Service Rating:** **10 / 10** (All 8 GORM models decoupled, separated translation layers, optimistic locks on work orders).
* **Messaging Queue Rating:** **10 / 10** (Outbox relay fully active, idempotent inbox deduplication verified, topics match `mfg.cdd` perfectly).
* **Status:** Reference standard.

### B. CRM Operations (`crm-service`)
* **Model Service Rating:** **10 / 10** (Customer profiles, opportunities, and billing triggers decoupled; range partitioned on billing logs).
* **Messaging Queue Rating:** **10 / 10** (Transactional outbox active, deduplication inbox handles O2C events like `crm.sales.order.created`).
* **Status:** Complete.

### C. Financials (`fm-service`)
* **Model Service Rating:** **10 / 10** (Ledger, universal journals, AR/AP, and COA fully decoupled).
* **Messaging Queue Rating:** **10 / 10** (Transactional outbox and idempotency verify FM ledger updates).
* **Status:** Complete.

---

## 5. Architectural Metrics Dashboard

```mermaid
radar-chart
    title "Microservice CDD Completeness Scores"
    labels [Decoupling (Ce=0), Model Translation, Tenant Isolation, Outbox Atomicity, Inbox Idempotency, Topic Alignment]
    "Manufacturing Service": [10, 10, 10, 10, 10, 10]
    "CRM Service": [10, 10, 10, 10, 10, 10]
    "Financial Service": [10, 10, 10, 10, 10, 10]
```

---

## 6. Implementation Plan: standardizing remaining services

To upgrade other services (e.g., SCM, HR, PLM) to a 10/10 Scorecard rating:
1. **CDD Model Cleanup:** Run `generate-all.sh` to produce new domain models. Delete legacy structures.
2. **GORM Separation:** Map GORM structs in `internal/data/sql/models.go` and write `ToDomain` / `FromDomain` converters.
3. **Wired Transactional Outbox:** Ensure all services use the outbox table for publishing events.
4. **Idempotent Ingestion:** Wrap consumer entrypoints in `ReliableMessagingService.ExecuteIdempotentTransaction`.
