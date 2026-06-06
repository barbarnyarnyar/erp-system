# ERP System CDD Gap Analysis — Phase S4.8: Audit Trails & History Entities (HR & CRM)

**Source PRD**: [cdd-gap-analysis.md](file:///Users/sithuhlaing/Projects/erp-system/docs/PRDs/active/2026-06-06-1557-cdd-gap-analysis.md)
**PRD ID**: PRD-2026-06-06-1557
**Phase**: S4.8
**Status**: Ready for Implementation
**Created**: June 06, 2026

---

## Objective

Close audit trail gaps in Human Resources and Sales/CRM pipelines by establishing structured change history entities and event-driven logging:
1. **HR (Position/Department History - 2.13)**: Add `PositionHistory` and `DepartmentHistory` entities to log structural employee changes.
2. **HR (Compensation Auditing - 2.19)**: Remove direct updates to `Employee.salary` via `UpdateEmployee`, make salary read-only/computed from `EmployeeCompensationHistory`, and create a dedicated `UpdateCompensation` method.
3. **CRM (Opportunity Stage History - 2.16)**: Introduce `OpportunityStageHistory` to log all stage transitions for sales opportunities.

## Rationale

* **HR (2.13)**: Employee promotions emit `hr.employee.promoted`, but the system lacks a persistent history table. Department transfers are completely unrecorded (no events, no database tracking).
* **HR (2.19)**: Currently, `UpdateEmployee` allows direct salary updates without writing to `EmployeeCompensationHistory`, creating a severe audit bypass. Salary must be read-only, computed from the compensation history ledger.
* **CRM (2.16)**: Pipeline velocity analysis and sales funnel metrics are impossible without records of intermediate opportunity stages (e.g. Discovery -> Negotiation).

---

## Scope

### In Scope
1. **HR Service**:
   - Define `PositionHistory` and `DepartmentHistory` structs in domain models.
   - Implement repository interfaces and in-memory storage for these entities.
   - Emit `hr.employee.transferred` event upon department change.
   - Modify `UpdateEmployee` to reject/ignore direct `salary` field modifications.
   - Implement `UpdateCompensation` inside `EmployeeService` that inserts into `EmployeeCompensationHistory` and updates the base `Employee.salary` dynamically.
   - Add unit tests verifying audit record creation on updates.

2. **CRM Service**:
   - Define `OpportunityStageHistory` struct in domain models.
   - Implement repository interface and memory storage.
   - Modify `UpdateOpportunity` to automatically insert an `OpportunityStageHistory` record when the stage changes.
   - Add unit tests verifying stage transition tracking.

---

## Implementation Tasks

### Task 1: Position & Department History (HR Service)
* **Files**:
  - `services/hr-service/internal/business/domain/position_history.go` (new)
  - `services/hr-service/internal/business/domain/department_history.go` (new)
* **Logic**:
  - Log history entries when position or department changes during `UpdateEmployee`.
  - Publish `hr.employee.transferred` and `hr.employee.promoted` Kafka events.

### Task 2: Read-Only Salary and Compensation Service (HR Service)
* **Files**:
  - `services/hr-service/internal/business/service/employee_service.go`
* **Logic**:
  - Remove salary from `UpdateEmployee` inputs.
  - Implement `UpdateCompensation(ctx, employeeID, salary, effectiveDate, changedBy)` which creates an audit record, and updates the cached `Employee.salary` field.

### Task 3: Opportunity Stage History (CRM Service)
* **Files**:
  - `services/crm-service/internal/business/domain/opportunity_history.go` (new)
  - `services/crm-service/internal/business/service/opportunity_service.go`
* **Logic**:
  - On `UpdateOpportunity`, if `oldStage != newStage`, insert record into `OpportunityStageHistory` store with timestamps.

---

## Verification

```bash
# Compile and test HR service
cd services/hr-service
go test ./...

# Compile and test CRM service
cd services/crm-service
go test ./...

# Full system build
cd ../..
make build
make run
make test
```

---

## Definition of Done

- [ ] `PositionHistory` and `DepartmentHistory` entities and repos are implemented.
- [ ] Direct salary updates in `UpdateEmployee` are blocked.
- [ ] Dedicated `UpdateCompensation` method updates `EmployeeCompensationHistory`.
- [ ] `OpportunityStageHistory` logs transitions during opportunity updates.
- [ ] All unit and integration tests compile and pass.
- [ ] Gap analysis checklists `2.13`, `2.16`, and `2.19` are marked complete.
