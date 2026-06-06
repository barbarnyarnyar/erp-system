# ERP System CDD Gap Analysis — Phase 2: Architecture Refactor

**Source PRD**: docs/PRDs/active/2026-06-06-1557-cdd-gap-analysis.md
**PRD ID**: PRD-2026-06-06-1557
**Phase**: 2 of 6
**Status**: Completed
**Created**: June 06, 2026

---

## Objective

Extract `MaintenanceService` from the ProductionService God struct. Restructure M Service to match CDD's 5 clean components: BOMService, ProductionService, QualityService, MaintenanceService, CostingService.

## Scope

### In Scope

- Create `maintenance_service.go` with dedicated `MaintenanceService` struct
- Move 7 methods from `ProductionService` → `MaintenanceService`: `LogMachineStatus`, `CreateEquipment`, `ScheduleMaintenance`, `CompleteMaintenance`, `ListMaintenanceSchedules`, `GetMaintenanceSchedule`, `UpdateMaintenanceSchedule`
- Move 3 methods from `ProductionService` → `QualityService`: `RecordQualityInspection`, `ListQualityInspections`, `GetQualityInspection`, `UpdateQualityInspection`
- Move 2 methods from `ProductionService` → `CostingService`: `GetCostingRecord`, `RunMRP`
- Reduce `ProductionService` from 28 methods to 16 (matching CDD)
- Wire new services in `cmd/main.go`
- Split handler registrations to use the correct service structs
- Move machine log and equipment routes to maintenance handler

### Out of Scope

- Renaming existing routes or changing URL paths
- Adding missing service methods (Phase 0)
- Event integrity fixes (Phase 1)

---

## Inputs

| Input | Source | Notes |
| ----- | ------ | ----- |
| CDD component definitions | `services/m-service/contracts/m.cdd` | 5 components with assigned methods |
| ProductionService (current) | `services/m-service/internal/business/service/production_service.go` | ~28 methods, God struct |
| QualityService (current) | `services/m-service/internal/business/service/quality_service.go` | Exists but only has CDD-mapped methods |
| CostingService (current) | `services/m-service/internal/business/service/costing_service.go` | Exists but only has CDD-mapped methods |

## Dependencies

| Dependency | Type | Required Before | Notes |
| ---------- | ---- | --------------- | ----- |
| Phase 0 | Code | Phase 2 start | Missing methods needed before restructuring |

---

## Implementation Tasks

### Task 1: Extract MaintenanceService

**Description:** Create a new `MaintenanceService` struct in its own file. Move the 7 maintenance-related methods from `ProductionService`.

**Methods to move (sub-tasks by logical group):**
- **1a — Machine log (1 method):** `LogMachineStatus`
- **1b — Equipment (1 method):** `CreateEquipment`
- **1c — Maintenance schedules (5 methods):** `ScheduleMaintenance`, `CompleteMaintenance`, `ListMaintenanceSchedules`, `GetMaintenanceSchedule`, `UpdateMaintenanceSchedule`

**Repository dependencies:** `MachineLogRepository`, `EquipmentRepository`, `MaintenanceOrderRepository` (already wired in main.go)

**Files to create:**
- `services/m-service/internal/business/service/maintenance_service.go`

**Acceptance Criteria:**
- `MaintenanceService` struct with all 7 methods
- `NewMaintenanceService` constructor
- `ProductionService` no longer imports machine/equipment/maintenance repos

### Task 2: Clean QualityService

**Description:** Move 4 methods from `ProductionService` to `QualityService`: `RecordQualityInspection`, `ListQualityInspections`, `GetQualityInspection`, `UpdateQualityInspection`.

**Sub-tasks:**
- 2a: Move `RecordQualityInspection` + `GetQualityInspection` (2 methods, read/write one entity)
- 2b: Move `ListQualityInspections` + `UpdateQualityInspection` (2 methods, list + update)
- 2c: Clean up ProductionService imports and struct fields

**Acceptance Criteria:**
- `QualityService` has all 4 CDD methods (existing + moved from ProductionService)
- `ProductionService` no longer has quality methods

### Task 3: Clean CostingService

**Description:** Move `GetCostingRecord` and `RunMRP` from `ProductionService` to `CostingService`.

**Acceptance Criteria:**
- `CostingService` has all 2 CDD methods
- `ProductionService` no longer has costing methods

### Task 4: Reduce ProductionService to CDD spec

**Description:** After extracting 12 methods, verify `ProductionService` has exactly the 16 CDD-defined methods:
- createProductionOrder, startWorkOrder, reportLabor, completeWorkOrder, completeProductionOrder, listProductionPlans, getProductionPlan, updateProductionPlan, deleteProductionPlan, listWorkOrders, createWorkOrder, getWorkOrder, updateWorkOrder, deleteWorkOrder, consumeMaterials, receiveFinishedGoods

**Acceptance Criteria:**
- `ProductionService` has exactly 16 methods (matching CDD)
- All method signatures preserved from original

### Task 5: Update main.go wiring

**Description:** Wire the new `MaintenanceService` in `cmd/main.go`. Split handler registrations so each route uses the correct service.

**Acceptance Criteria:**
- `main.go` creates `MaintenanceService` with its repos
- Routes that require maintenance logic call `MaintenanceService` methods
- `main.go` does NOT pass machine/equipment/maint repos to `ProductionService`

### Task 6: Split handler routes

**Description:** Move machine-log and maintenance routes to `maintenance_handler.go` or update existing routes to use `MaintenanceService`.

**Current routes affected:**
- `POST /api/v1/work-centers/:id/machine-log` → use MaintenanceService
- `GET/POST/PUT /api/v1/maintenance-schedules` → use MaintenanceService

**Acceptance Criteria:**
- All routes compile and respond correctly
- Route paths unchanged

---

## Verification

```bash
cd services/m-service

# Build
make build

# Check ProductionService method count
rg '^func \(s \*ProductionService\)' internal/business/service/production_service.go | wc -l
# Should be 16

# Check MaintenanceService exists
rg 'type MaintenanceService struct' internal/business/service/maintenance_service.go

# Full test
make test
```

---

## Risks

| Risk | Likelihood | Mitigation |
| ---- | ---------- | ---------- |
| Method move breaks internal callers | Medium | `ProductionService` may call its own methods internally; change call sites to use new service instances |
| Circular dependencies | Low | Services don't depend on each other; they share repos |
| Route handler changes break API contract | Medium | Keep route paths identical, only change the service struct called |

## Design Decision: ProductionService ↔ MaintenanceService Coupling

**Decision: Composition.** `ProductionService` holds an optional `*MaintenanceService` reference.

After extraction, `ProductionService` may internally call maintenance methods (e.g., completing a work order triggers a machine log). The pattern:

```go
type ProductionService struct {
    maintenanceSvc *MaintenanceService  // optional — nil means no maintenance tracking
    qualitySvc     *QualityService      // optional — nil means no quality checks
    costingSvc     *CostingService      // optional — nil means no cost tracking
}
```

**Sub-task in Task 5:** When wiring in `main.go`, create all 5 services and inject maintenance/quality/costing into `ProductionService`:

```go
ms := service.NewMaintenanceService(machineLogRepo, equipmentRepo, maintRepo, publisher)
qs := service.NewQualityService(qualityRepo, nonConfRepo, publisher)
cs := service.NewCostingService(costRepo)
ps := service.NewProductionService(bomRepo, compRepo, ..., ms, qs, cs)
```

No backward-compatible delegate methods — all call sites updated to use the appropriate service directly.

## Definition of Done

- [x] Task 1: `MaintenanceService` exists with 7 methods
- [x] Task 2: `QualityService` has all 4 CDD methods
- [x] Task 3: `CostingService` has all 2 CDD methods
- [x] Task 4: `ProductionService` has exactly 16 CDD methods
- [x] Task 5: main.go wires 5 components (BOM/Production/Quality/Maintenance/Costing)
- [x] Task 6: All routes use correct service structs
- [x] `make build` passes
- [x] `make test` passes
