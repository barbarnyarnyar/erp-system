# ERP System CDD Gap Analysis — Phase 0.5: Event Error Logging

**Source PRD**: docs/PRDs/active/2026-06-06-1557-cdd-gap-analysis.md
**PRD ID**: PRD-2026-06-06-1557
**Phase**: 0.5 of 6 (parallel Track B)
**Status**: Completed
**Created**: June 06, 2026

---

## Objective

Replace every `_ = publisher.Publish(...)` with error-logged `if err := publisher.Publish(...)` across all 7 services. This is a mechanical find-replace with zero behavioral change beyond visible error messages.

## Rationale

Currently all 65+ publish calls silently discard errors. Before adding new publishes in Phase 1, we need the error-logging pattern established first. This phase runs in **parallel with Phase 0** (Track A) since it touches different files.

## Scope

### In Scope

- Find all `_ = publisher.Publish(...)` patterns across all services
- Replace with `if err := publisher.Publish(...); err != nil { log.Printf(...) }`
- Use consistent log format: `"ERROR: failed to publish event [topic]: %v"`

### Out of Scope

- Dead-letter queue (Phase 6)
- Retry logic
- Adding new publish calls (Phase 1)

---

## Implementation Tasks

### Task 1: Audit all discard patterns

**Description:** Count all `_ = publisher.Publish(` occurrences.

```bash
rg '_ = .*Publish\(' services/ --type go
```

Expected count: ~65 across 7 services.

### Task 2: Replace by service (7 sub-tasks, parallelizable)

**2a — Auth service:**
- `services/auth-service/internal/business/service/auth_service.go`
- `services/auth-service/internal/business/service/user_service.go`

**2b — FM service:**
- `services/fm-service/internal/business/service/general_ledger_service.go`
- `services/fm-service/internal/business/service/accounts_receivable_service.go`
- `services/fm-service/internal/business/service/cash_management_service.go`
- `services/fm-service/internal/business/service/accounts_payable_service.go`
- `services/fm-service/internal/business/service/budgeting_service.go`

**2c — HR service:**
- `services/hr-service/internal/business/service/employee_management_service.go`
- `services/hr-service/internal/business/service/payroll_service.go`
- `services/hr-service/internal/business/service/time_attendance_service.go`
- `services/hr-service/internal/business/service/leave_management_service.go`
- `services/hr-service/internal/business/service/performance_service.go`
- `services/hr-service/internal/business/service/training_service.go`

**2d — SCM service:**
- `services/scm-service/internal/business/service/purchase_order_service.go`
- `services/scm-service/internal/business/service/inventory_service.go`

**2e — M service:**
- `services/m-service/internal/business/service/production_service.go`
- `services/m-service/internal/business/service/quality_service.go`
- `services/m-service/internal/business/service/costing_service.go`
- `services/m-service/internal/business/service/bom_service.go`

**2f — CRM service:**
- `services/crm-service/internal/business/service/customer_service.go`
- `services/crm-service/internal/business/service/lead_service.go`
- `services/crm-service/internal/business/service/opportunity_service.go`
- `services/crm-service/internal/business/service/order_service.go`
- `services/crm-service/internal/business/service/quote_service.go`
- `services/crm-service/internal/business/service/ticket_service.go`
- `services/crm-service/internal/business/service/campaign_service.go`

**2g — PM service:**
- `services/pm-service/internal/business/service/project_planning_service.go`
- `services/pm-service/internal/business/service/task_management_service.go`
- `services/pm-service/internal/business/service/resource_management_service.go`
- `services/pm-service/internal/business/service/time_expense_service.go`
- `services/pm-service/internal/business/service/collaboration_service.go`

### Replacement Pattern

```go
// Before:
_ = publisher.Publish(ctx, "some.topic", event)

// After:
if err := publisher.Publish(ctx, "some.topic", event); err != nil {
    log.Printf("ERROR: failed to publish event some.topic: %v", err)
}
```

Alternatively, for brevity:
```go
if err := publisher.Publish(ctx, topic, event); err != nil {
    log.Printf("ERROR: publish event failed: %v", err)
}
```

**Acceptance Criteria:**
- `rg '_ = .*Publish\(' services/ --type go` returns zero matches
- Each `Publish` call is wrapped in `if err := ...; err != nil { log.Printf(...) }`

---

## Verification

```bash
# Check zero discarded errors remain
rg '_ = .*Publish\(' services/ --type go
# Expected: empty

# Check consistent pattern
rg 'if err := .*\.Publish\(' services/ --type go | wc -l
# Expected: ~65+

# Build
for svc in auth fm hr scm m crm pm; do
  cd services/$svc-service && go build ./...
done
```

---

## Risks

| Risk | Likelihood | Mitigation |
| ---- | ---------- | ---------- |
| Missed a call site | Low | `rg '_ = .*Publish\('` catches all patterns |
| log import missing in some files | Low | `go build` catches missing imports |
| Inconsistent log format between services | Medium | Use same format string everywhere |

## Definition of Done

- [x] Task 1: Audit complete (baseline count recorded)
- [x] Task 2a–2g: All 7 services converted
- [x] Zero `_ = publisher.Publish(...)` patterns remain
- [x] `go build ./...` passes for all services

---

## Handoff Notes

This is the fastest win in the entire PRD. Estimated 30 minutes for the audit + find-replace across all services. Phase 1 depends on the error-logging pattern being established before adding new publish calls.
