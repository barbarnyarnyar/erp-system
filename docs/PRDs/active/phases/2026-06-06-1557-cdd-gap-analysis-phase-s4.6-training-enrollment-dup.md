# ERP System CDD Gap Analysis — Phase S4.6: TrainingEnrollment Duplicate Protection

**Source PRD**: docs/PRDs/active/2026-06-06-1557-cdd-gap-analysis.md
**PRD ID**: PRD-2026-06-06-1557
**Phase**: S4.6 of 6
**Status**: Completed
**Created**: June 06, 2026

---

## Objective

Fix the duplicate-enrollment bug in `TrainingService.EnrollEmployee()`. The repo's `GetByTrainingAndEmployee(trainingID, employeeID)` method existed and was declared on the `TrainingEnrollmentRepository` interface, but the service **never called it** — same employee could be enrolled in the same training program multiple times.

## Rationale

- `TrainingEnrollment` has no `@unique` composite annotation in `hr.cdd`
- Service called `repo.Create()` directly without existence check
- Result: silent data corruption — duplicate enrollments for the same `(training_id, employee_id)` pair
- The repo method existed but was dead code

## Scope

### In Scope
- Add duplicate check at the start of `EnrollEmployee` using `enrollments.GetByTrainingAndEmployee(...)`
- Policy: reject duplicate only if prior enrollment is `ENROLLED` or `IN_PROGRESS`
- Allow re-enrollment if prior is `CANCELLED` or `COMPLETED` (real-world LMS behavior)
- Add unit tests covering: duplicate active enrollment rejected, different employee allowed, different training allowed, re-enrollment after cancellation allowed
- Use the existing `GetByTrainingAndEmployee` repo method — no schema change needed

### Out of Scope
- Adding a DB-level composite unique constraint (in-memory repo; no DB)
- Withdrawal / cancellation flow (no such method exists yet)
- Bulk enrollment
- Waitlist handling

---

## Implementation Tasks

### Task 1: Read existing code
- `services/hr-service/internal/business/service/training_service.go:93` — `EnrollEmployee` method
- `services/hr-service/internal/data/memory/memory_repos.go:656` — `GetByTrainingAndEmployee` memory impl
- `services/hr-service/internal/business/domain/repository.go:100-106` — repo interface

### Task 2: Add duplicate guard
- Before `Create`, call `enrollments.GetByTrainingAndEmployee(trainingID, employeeID)`
- If exists and status is `ENROLLED` or `IN_PROGRESS`: return descriptive error
- If exists and status is `CANCELLED` or `COMPLETED`: fall through (allow re-enrollment)
- If not found (repo returns error): proceed with create

### Task 3: Tests
- File: `services/hr-service/internal/business/service/training_enrollment_test.go`
- 2 test functions:
  1. `TestEnrollEmployee_PreventsDuplicate` — covers happy path + duplicate rejection + different employee + different training
  2. `TestEnrollEmployee_AllowsReEnrollmentAfterCancellation` — covers re-enrollment policy

---

## Verification

```bash
cd services/hr-service && go test ./internal/business/service/ -run "TestEnrollEmployee" -v
# Both tests pass

for svc in auth fm hr scm m crm pm; do
  (cd services/$svc-service && go build ./...)
done
# All 7 services build cleanly
```

---

## Risks

| Risk | Likelihood | Mitigation |
| ---- | ---------- | ---------- |
| Existing data has duplicates that now block new operations | Low | In-memory store; reset on restart. No migration needed |
| Status string values change in future | Low | Use the CDD-declared values; could be replaced by typed constants in future S9.5 phase |
| Race condition between check and create | Medium | In-memory store with mutex on writes (memory_repos.go uses sync.RWMutex). Real DB would need unique constraint |

## Definition of Done

- [x] `EnrollEmployee` calls `GetByTrainingAndEmployee` before `Create`
- [x] Active duplicate enrollments are rejected
- [x] Re-enrollment after `CANCELLED`/`COMPLETED` is allowed
- [x] 2 unit tests pass
- [x] All 7 services build cleanly
- [x] Master PRD 2.12 DoD checkbox marked complete

## Handoff Notes

This was the **quickest P1 win** — the fix is a 5-line addition, the repo method already existed. The full S4.6 work took < 30 minutes.
