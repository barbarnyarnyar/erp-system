# PRD: Manufacturing Service PostgreSQL/GORM Database Persistence Migration

**PRD ID**: PRD-2026-06-13-1005  
**Date**: 2026-06-13  
**Status**: Implemented  
**Parent Initiative**: Database Integration & Enterprise Persistence  
**Target Coverage**: 100% GORM persistent database coverage for Manufacturing repositories  

---

## 1. Objective & Architectural Context

Currently, the Manufacturing service (`mfg-service` / `m-service`) operates on volatile in-memory repositories. This causes complete data loss whenever the service container restarts. To bring the service to enterprise-grade readiness and align it with the database persistence of the other ERP microservices:
1. We will implement database connection and transaction management logic inside a new `internal/data/sql/` directory for the manufacturing service.
2. We will map all domain models defined in the manufacturing domain to GORM PostgreSQL schema entities.
3. We will write SQL repository implementations that implement the domain repository interfaces.
4. We will wire up GORM database initialization in the main entrypoint (`cmd/main.go`) of the service.

---

## 2. Technical Specification & Scope

### A. Manufacturing Service Database Schema (8 CDD Entities)
We will define GORM mappings and database repository adapters for the following entities:
* `WorkCenter`, `RoutingStation`, `WorkOrder`, `WorkOrderRoutingState`, `MaterialConsumptionLog`, `ProductionYieldLog`, `TransactionalOutbox`, `KafkaEventInbox`.

---

## 3. Scope & Implementation Checklist

### Phase 1: GORM Models & Setup
- [x] Create `services/mfg-service/internal/data/sql/db.go` (DB initialization and transaction helper).
- [x] Create `services/mfg-service/internal/data/sql/models.go` (GORM entity model mappings).

### Phase 2: SQL Repositories & main.go Wiring
- [x] Create `services/mfg-service/internal/data/sql/sql_repos.go` (8 repository adapters).
- [x] Wire the GORM connection and repositories into `services/mfg-service/cmd/main.go`.
- [x] Verify compilation of the manufacturing service with database repositories.

---

## 4. Definition of Done
- [x] All 8 Manufacturing repositories are implemented in SQL using GORM.
- [x] The service initializes PostgreSQL database on startup using GORM.
- [x] The service compiles and runs cleanly.
