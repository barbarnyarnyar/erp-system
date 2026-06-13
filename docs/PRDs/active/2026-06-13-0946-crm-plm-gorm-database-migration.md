# PRD: CRM & PLM Services PostgreSQL/GORM Database Persistence Migration

**PRD ID**: PRD-2026-06-13-0946  
**Date**: 2026-06-13  
**Status**: Implemented  
**Parent Initiative**: Database Integration & Enterprise Persistence  
**Target Coverage**: 100% GORM persistent database coverage for CRM and PLM repositories  

---

## 1. Objective & Architectural Context

Currently, the CRM service (`crm-service`) and PLM service (`plm-service`) operate on volatile in-memory maps. This causes complete data loss whenever a microservice container restarts. To bring both services to enterprise-grade readiness and align them with the persistent architectures of the Finance (`fm-service`) and Supply Chain (`scm-service`) modules:
1. We will implement database connection and transaction management logic inside a new `internal/data/sql/` directory for both services.
2. We will map all domain models defined in their respective CDDs (`crm.cdd` and `plm.cdd`) to GORM PostgreSQL schema entities.
3. We will write SQL repository implementations that implement the domain repository interfaces.
4. We will wire up GORM database initialization in the main entrypoints (`cmd/main.go`) of both services.

---

## 2. Technical Specification & Scope

### A. CRM Service Database Schema (17 Entities)
We will define GORM mappings and database repository adapters for the following entities:
* `CustomerProfile`, `PriceBookHeader`, `PriceBookEntry`, `PricingStrategy`, `SalesOrder`, `SalesOrderLine`, `BillingTrigger`, `TransactionalOutbox`, `KafkaEventInbox` (Package `erp.crm.core`).
* `Campaign`, `Lead`, `Opportunity`, `CustomerInteraction`, `ServiceTicket`, `Quote`, `QuoteLineItem` (Package `erp.crm.operations`).
* `OpportunityStageHistory` (operational helper structure).

### B. PLM Service Database Schema (6 Entities)
We will define GORM mappings and database repository adapters for the following entities:
* `MaterialMaster`, `BomHeader`, `BomLine`, `EngineeringChangeOrder`, `TransactionalOutbox`, `KafkaEventInbox` (Package `erp.engineering`).

### C. GORM Transaction Manager
We will implement transaction managers that implement the domain's transactional execution contracts, allowing services to perform atomic operations over multiple repositories.

---

## 3. Scope & Implementation Checklist

### Phase 1: CRM GORM Models & Setup
- [x] Create `services/crm-service/internal/data/sql/db.go` (DB initialization and transaction manager).
- [x] Create `services/crm-service/internal/data/sql/models.go` (GORM entity model mappings).

### Phase 2: CRM SQL Repositories & main.go Wiring
- [x] Create `services/crm-service/internal/data/sql/sql_repos.go` (17 repository adapters).
- [x] Wire the GORM connection and repositories into `services/crm-service/cmd/main.go`.
- [x] Update CRM service tests to compile and pass with the database repository changes.

### Phase 3: PLM GORM Models & Setup
- [x] Create `services/plm-service/internal/data/sql/db.go` (DB initialization and transaction manager).
- [x] Create `services/plm-service/internal/data/sql/models.go` (GORM entity model mappings).

### Phase 4: PLM SQL Repositories & main.go Wiring
- [x] Create `services/plm-service/internal/data/sql/sql_repos.go` (6 repository adapters).
- [x] Wire the GORM connection and repositories into `services/plm-service/cmd/main.go`.
- [x] Update PLM service tests to compile and pass with the database repository changes.

### Phase 5: Verification & Compilation
- [x] Verify both `crm-service` and `plm-service` compile cleanly with `go build ./...`.
- [x] Run test suites for both services using `go test ./...` and ensure all pass.

---

## 4. Definition of Done
- [x] All 17 CRM repositories and 6 PLM repositories are implemented in SQL using GORM.
- [x] Both services initialize PostgreSQL databases on startup using GORM.
- [x] All unit and integration tests compile and pass successfully.
