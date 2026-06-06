# ERP System CDD Gap Analysis — Phase S4.8: Auth Consumer for hr.employee.terminated

**Source PRD**: docs/PRDs/active/2026-06-06-1557-cdd-gap-analysis.md
**PRD ID**: PRD-2026-06-06-1557
**Phase**: S4.8 of 6
**Status**: Completed
**Created**: June 06, 2026

---

## Objective

Add a Kafka consumer in auth-service that subscribes to `hr.employee.terminated` and automatically calls `deactivateUser(employeeID)`. This closes the offboarding loop: when HR terminates an employee, the corresponding Auth user is deactivated immediately, which (via S4.7) invalidates any in-flight JWT by bumping the security_stamp.

## Rationale

- Auth was the **only service with zero Kafka consumers** — fire-and-forget was the wrong design intent
- Without this, the `hr.employee.terminated` event was published by HR but no one consumed it
- Real security vulnerability: terminated employees' auth accounts remained active indefinitely
- The original design rationale ("Auth events are fire-and-forget") was revised — see master PRD §3.5

## Scope

### In Scope
- Add `consumer_events` block to `auth.cdd` declaring `hr.employee.terminated: HREmployeeTerminatedEvent`
- Add `HREmployeeTerminatedEvent` payload struct to auth domain
- Add `TopicHrEmployeeTerminated` constant to auth's `event_topics.go`
- Create `internal/data/kafka/consumer.go` for auth-service with:
  - `KafkaConsumer` struct holding `reader`, `publisher`, `userSvc`
  - `Start(ctx)` background loop using `kafka-go` reader
  - `handleMessage` dispatcher
  - `handleMessage` for `TopicHrEmployeeTerminated` calls `userSvc.DeactivateUser`
  - Graceful `Close()` for the reader
- Wire consumer into `cmd/main.go` with `consumerCtx` + `go consumer.Start()` + `defer consumer.Close()`
- 2 unit tests covering happy path + unknown user

### Out of Scope
- DLQ (Dead-Letter Queue) — already noted as P5 / S14 future work
- Backfill: scanning existing active users against currently-terminated HR employees
- Cross-service ID mapping service (the `@reference` linkage between Employee and User is the master PRD 2.10 debt)

---

## Implementation Tasks

### Task 1: CDD + struct updates
- `services/auth-service/contracts/auth.cdd` — add `consumer_events` block
- `services/auth-service/internal/business/domain/events.go` — add `HREmployeeTerminatedEvent`
- `services/auth-service/internal/business/domain/event_topics.go` — add `TopicHrEmployeeTerminated`

### Task 2: Create consumer
- New file: `services/auth-service/internal/data/kafka/consumer.go`
- Follows same pattern as `hr-service/internal/data/kafka/consumer.go`
- Group ID: `auth-service`
- Subscribes to: `TopicHrEmployeeTerminated`
- On message: `json.Unmarshal` → `userSvc.DeactivateUser(employeeID)`
- Returns error if user not found (so caller can DLQ / retry)

### Task 3: Wire into main.go
- `cmd/main.go`: after publisher setup, instantiate consumer, start goroutine
- `defer consumer.Close()` for graceful shutdown

### Task 4: Tests
- File: `services/auth-service/internal/data/kafka/consumer_test.go`
- Helper: `newConsumerWithUser` seeds a user and creates a consumer
- 2 test functions:
  1. `TestConsumer_HREmployeeTerminated_DeactivatesUser` — happy path; verifies `IsActive=false` and stamp bumped
  2. `TestConsumer_HREmployeeTerminated_UnknownUserReturnsError` — graceful error for unknown user

---

## Verification

```bash
cd services/auth-service && go test ./... -v
# All 7 tests pass (5 from S4.7 + 2 from S4.8)

for svc in auth fm hr scm m crm pm; do
  (cd services/$svc-service && go build ./...)
done
# All 7 services build cleanly
```

---

## Risks

| Risk | Likelihood | Mitigation |
| ---- | ---------- | ---------- |
| Duplicate processing — termination event consumed twice | Low | `DeactivateUser` is idempotent: setting `IsActive=false` twice is a no-op. The stamp bumps twice, but a stamp mismatch is a "valid" reason to reject. |
| Cross-service ID mapping — what if EmployeeID ≠ UserID? | Medium | The MVP uses EmployeeID-as-UserID convention (master PRD 2.10 deferred). If a future service needs a different mapping, add a lookup table. |
| Consumer not yet started at startup time | Low | Consumer goroutine is started in main.go with consumerCtx, before HTTP server starts accepting requests. If Kafka is unreachable, reader logs and retries. |
| Auth now has a publisher dependency that's no longer fully fire-and-forget | None | Publisher was always there; just adds a consumer that also uses it (for DLQ in future). |

## Definition of Done

- [x] `auth.cdd` updated with `consumer_events` block
- [x] `HREmployeeTerminatedEvent` payload added
- [x] `TopicHrEmployeeTerminated` constant added
- [x] `KafkaConsumer` struct + handlers in `internal/data/kafka/consumer.go`
- [x] Consumer wired into `cmd/main.go` with graceful shutdown
- [x] 2 unit tests pass
- [x] Full auth-service test suite passes (7 tests)
- [x] All 7 services build cleanly
- [x] Master PRD 2.18 DoD checkbox marked complete
- [x] **P1 is fully complete** (5/5 P1 tasks done)

## Handoff Notes

**P1 is done.** All 5 P1 tasks (S1, S4.5, S4.6, S4.7, S4.8) are complete. The system now has:
- Error logging on all Kafka publishes (no silent failures)
- Enforced inventory invariant
- Duplicate-enrollment protection
- JWT revocation via security_stamp
- Auto-deactivation on HR termination

The offboarding flow is now end-to-end: HR terminates employee → `hr.employee.terminated` event published → auth consumer receives it → `DeactivateUser` called → security_stamp bumped → all in-flight JWTs rejected on next validation.

Next step: **P2 (Functional Completeness)** — the largest bucket. 9 days of work for 7 missing repos, 5 missing methods, 27 missing HTTP routes, GL atomicity, ConvertLead atomicity, HR salary refactor, Milestone entity, confirmSalesOrder function.
