# Cross-Service Code Deduplication

**Date**: 2026-06-06
**Status**: Draft (not yet scheduled)
**Parent initiative**: `2026-06-06-1557-cdd-gap-analysis.md` (P2 bucket)
**Estimated total effort**: 3-4 days
**Risk**: Medium (touches 40+ files across 7 services; mitigated by phased rollout + per-service compile/test gates)

## Problem

A repository-wide scan (2026-06-06) revealed **7 categories of duplicated logic** across 7 microservices, totaling **~120+ duplicated code blocks**. The most acute problems:

1. **Event publish-with-error-log** is hand-rolled in **32 service files** (each: 4 lines).
2. **ID generation** via `fmt.Sprintf("xxx_%d", time.Now().UnixNano())` appears in **40 service files** (each: 1 line, but collision-prone under load).
3. **Kafka publisher constructor** is byte-identical in **7 services** (each: 6 lines).
4. **`MockPublisher` test fixture** is redeclared in **5 test files** with subtle drift (some track events, some don't; one has `FailPublish` toggle).
5. **`IsValid()` enum validator** is the same `switch { case ...; return true } return false` pattern in **7 files**.
6. **HTTP error response** `c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})` is in **28 handler files**.
7. **`shared/utils/` Go module** already contains a `ResponseHelper`, a `LoggerEntry`, and a `StandardResponse` type — but **zero services import it**. It's dead code.

The `shared/utils` module is the smoking gun: the abstraction was designed and built, but never adopted. This PRD is the adoption + extraction plan.

## Goals

1. **Centralize** the duplicated patterns into `shared/utils/` (or new `shared/kafka/`, `shared/idgen/` submodules) so that each pattern exists **exactly once**.
2. **Replace** all hand-rolled copies with calls to the shared helpers.
3. **Verify** every service still builds and all tests still pass after the migration.
4. **Lock in** the standard via a shared-module lint rule or pre-commit check (out of scope for this PRD — flagged for follow-up).

## Non-Goals

- No behavior changes. All migrations are 1-to-1 refactors.
- No new features. No test coverage changes (except consolidating mocks).
- No breaking change to the `EventPublisher` interface contract.
- No move to PostgreSQL, no production deployment changes.
- No attempt to dedupe the `Memory*Repository` implementations in this round — they have service-specific fields and the volume/risk is too high for the value gained. Flagged for a follow-up PRD.

## Duplication Inventory (with file paths and counts)

### 1. Event Publish-with-Error-Log (32 files)

Pattern:
```go
if err := s.publisher.Publish(ctx, domain.TopicXxx, key, payload); err != nil {
    log.Printf("ERROR: failed to publish event %s: %v", domain.TopicXxx, err)
}
```

Hot files (top 5):
- `services/scm-service/internal/business/service/inventory_service.go` (6 occurrences)
- `services/hr-service/internal/business/service/employee_management_service.go` (9)
- `services/m-service/internal/business/service/production_service.go` (8)
- `services/pm-service/internal/business/service/project_planning_service.go` (8)
- `services/pm-service/internal/business/service/time_expense_service.go` (7)

### 2. ID Generation via `time.Now().UnixNano()` (40 files)

Pattern:
```go
id := fmt.Sprintf("xxx_%d", time.Now().UnixNano())
```

Collision risk: same nanosecond → duplicate IDs in fast-path code (e.g., bulk inserts in a loop). 40 files with at least 1 call.

### 3. Kafka Publisher Constructor (7 services — byte-identical)

Files:
- `services/scm-service/internal/data/kafka/producer.go`
- `services/auth-service/internal/data/kafka/producer.go` (minor: missing `AllowAutoTopicCreation: true`)
- `services/crm-service/internal/data/kafka/producer.go`
- `services/fm-service/internal/data/kafka/producer.go`
- `services/m-service/internal/data/kafka/producer.go`
- `services/hr-service/internal/data/kafka/producer.go`
- `services/pm-service/internal/data/kafka/producer.go`

The auth-service one is **slightly different** (no `AllowAutoTopicCreation`). This is a latent bug — auth topics won't auto-create. Migration will normalize.

### 4. `MockPublisher` Test Fixture (5 test files)

```go
type MockPublisher struct{}  // or
type MockPublisher struct { Events []MockEvent; FailPublish bool }

func (m *MockPublisher) Publish(ctx context.Context, topic string, key string, payload interface{}) error { ... }
```

Files:
- `services/auth-service/internal/business/service/security_stamp_test.go` (no-op version)
- `services/scm-service/internal/business/service/inventory_invariant_test.go` (no-op version)
- `services/crm-service/internal/business/service/lead_transaction_test.go` (events-tracked + FailPublish)
- `services/crm-service/internal/business/service/confirm_sales_order_test.go` (uses crm version)
- `services/fm-service/internal/business/service/service_test.go` (events-tracked)
- `services/hr-service/internal/business/service/training_enrollment_test.go` (no-op version)

### 5. `IsValid()` Switch Validator (7 files)

Pattern: `func (s X) IsValid() bool { switch s { case Xa, Xb, Xc: return true }; return false }`

Files:
- `services/crm-service/internal/business/domain/customer.go` (CustomerStatus)
- `services/crm-service/internal/business/domain/sales_order_helpers.go` (SalesOrderStatus — added 2026-06-06)
- `services/m-service/internal/business/domain/production_order.go` (ProductionOrderStatus)
- `services/m-service/internal/business/domain/work_order.go` (WorkOrderStatus)
- `services/hr-service/internal/business/domain/leave_request.go` (LeaveType + LeaveStatus)
- `services/fm-service/internal/business/domain/account.go` (AccountType)

### 6. HTTP Error Response (28 files)

Pattern:
```go
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
```

### 7. Dead `shared/utils/` Module (0 importers)

`shared/utils/logger.go` (99 lines) and `shared/utils/response.go` (144 lines) exist but are unused. The work below **adopts** them.

## Design

### New shared submodules

```
shared/
├── utils/                    # existing
│   ├── logger.go             # adopt as-is
│   ├── response.go           # adopt as-is, add BindAndValidate helper
│   └── idgen.go              # NEW: ID generation
├── kafka/                    # NEW
│   └── publisher.go          # NewKafkaPublisher + interface
└── testing/                  # NEW
    └── mockpublisher.go      # shared mock with events tracking + fail toggle
```

### `shared/utils/idgen.go` (new)

```go
package utils

import (
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "time"
)

func NewID(prefix string) string {
    b := make([]byte, 8)
    _, _ = rand.Read(b)
    return fmt.Sprintf("%s_%s_%d", prefix, hex.EncodeToString(b), time.Now().UnixNano())
}
```

Rationale: 8 random bytes = 16 hex chars, no collision in practical workloads. Format: `xxx_<rand8hex>_<nanos>` (verbose but grep-friendly).

### `shared/kafka/publisher.go` (new)

```go
package kafka

import "github.com/segmentio/kafka-go"

type Publisher struct { writer *kafka.Writer }

func NewPublisher(brokers []string) *Publisher { ... }  // single canonical impl

func (p *Publisher) Publish(ctx context.Context, topic, key string, payload interface{}) error { ... }
func (p *Publisher) Close() error { ... }
```

All 7 services delete their local `internal/data/kafka/producer.go` `NewKafkaPublisher` and `Publish` (keeping the same struct shape via type alias where needed).

### `shared/testing/mockpublisher.go` (new)

```go
package testing

type Event struct { Topic, Key string; Payload interface{} }
type MockPublisher struct {
    Events []Event
    FailPublish bool
}
func (m *MockPublisher) Publish(ctx, topic, key, payload) error { ... }
```

### `shared/utils/response.go` (extend)

Add:
```go
func (r *ResponseHelper) BindAndValidate(c *gin.Context, obj interface{}) bool {
    if err := c.ShouldBindJSON(obj); err != nil { r.BadRequest(c, err.Error()); return false }
    return true
}
func (r *ResponseHelper) NotFoundErr(c *gin.Context, err error) { r.Error(c, 404, "not found", err) }
func (r *ResponseHelper) ConflictErr(c *gin.Context, err error) { r.Error(c, 409, "conflict", err) }
func (r *ResponseHelper) InternalErr(c *gin.Context, err error) { r.Error(c, 500, "internal", err) }
```

### `IsValid()` — Go 1.18+ generics

Replace the 7 hand-rolled `IsValid()` methods with a generic helper:

```go
package utils

func IsAny[T comparable](v T, valid ...T) bool {
    for _, x := range valid { if x == v { return true } }
    return false
}
```

Usage: `func (s SalesOrderStatus) IsValid() bool { return utils.IsAny(s, DRAFT, CONFIRMED, ...) }`

One-liner, type-safe, no codegen.

### `PublisherLog` helper

For the 32 publish-with-log sites, add to `shared/utils`:

```go
package utils

import "log"

func LogPublishErr(service, topic string, err error) {
    if err != nil { log.Printf("[%s] ERROR: failed to publish %s: %v", service, topic, err) }
}
```

Plus a `Publisher` decorator (optional, can defer to a later phase):
```go
type LoggingPublisher struct { inner EventPublisher; service string }
func (l *LoggingPublisher) Publish(ctx, topic, key, payload) error {
    err := l.inner.Publish(ctx, topic, key, payload)
    LogPublishErr(l.service, topic, err)
    return err
}
```

## Phases (execution order)

| # | Phase | Effort | Risk | DoD |
|---|-------|--------|------|-----|
| **D1** | Add `shared/kafka`, `shared/utils/idgen`, `shared/testing` submodules; keep `logger.go` and `response.go`; publish `go.mod` bump. | 0.5d | Low | `go build ./...` passes in each service when they import (no actual imports yet). |
| **D2** | Migrate 7 services to use `shared/kafka.NewPublisher` (delete local constructors). Per-service: edit `go.mod` (replace directive), delete `internal/data/kafka/producer.go`, retest. | 0.5d | Medium (auth service has a divergent version) | All 7 services `go test ./...` pass; auth service gains `AllowAutoTopicCreation: true` (catch up to standard). |
| **D3** | Migrate 40 service files to use `utils.NewID(prefix)`. Use sed/awk for the obvious cases; manual for the rest. | 0.5d | Low (pure renames) | `rg "UnixNano" services/` returns 0 hits in `internal/business/service/*.go` (template is fine). |
| **D4** | Migrate 32 service files to use `utils.LoggingPublisher` decorator or `utils.LogPublishErr` helper. **Choice**: decorator is cleaner but requires struct field change. Start with the helper (zero struct change). | 1d | Low | All 32 sites replaced; `rg "log.Printf.*publish event"` returns 0 hits in `internal/`. |
| **D5** | Migrate 5 test files to use `shared/testing.MockPublisher`. Delete local `type MockPublisher` declarations. | 0.25d | Low (test code only) | `rg "type MockPublisher" --type go` returns 1 hit (the canonical in `shared/testing/`). |
| **D6** | Migrate 7 `IsValid()` files to use `utils.IsAny`. | 0.25d | Trivial | `rg "func .* IsValid\(\) bool" --type go` returns 7 hits that are 1-line bodies. |
| **D7** | Migrate 28 handler files to use `ResponseHelper.BadRequest/NotFound/Internal`. Inject `ResponseHelper` via handler constructor (currently zero handlers use it). | 1d | Medium (touches every handler struct) | `rg "c\.JSON\(http\.Status(BadRequest|InternalServerError|NotFound)"` returns 0 hits in `internal/api/handlers/`. |
| **D8** | Update all 6 `cmd/main.go` to use `utils.InitLogger(serviceName)` and `utils.NewResponseHelper(serviceName)`. | 0.25d | Low | `rg "InitLogger|NewResponseHelper" services/*/cmd/main.go` returns 6 hits. |

**Total**: 4.25 days. Parallelize D3-D6 (all small, independent).

## Migration Strategy (per-service)

For each service touched in D2-D8:

1. Add `replace github.com/erp-system/shared => ../../shared` to the service's `go.mod` (already pattern; see existing `common-utils` symlinks).
2. Update imports: `import "github.com/erp-system/shared/utils"` / `kafka` / `testing`.
3. Apply refactor (sed for D3; manual for D4, D7).
4. `go build ./...` — must pass.
5. `go test ./...` — must pass.
6. `go vet ./...` — must pass.

## Acceptance Criteria (DoD for the whole PRD)

- [x] `shared/kafka`, `shared/utils/idgen`, `shared/testing` submodules created with unit tests (≥80% coverage of helpers).
- [x] Zero `func NewKafkaPublisher` declarations remain in any service.
- [x] Zero `type MockPublisher` declarations remain outside `shared/testing/`.
- [x] Zero `time.Now().UnixNano()` ID-generation calls remain in service business code.
- [x] Zero `log.Printf("ERROR: failed to publish event...")` call sites remain in service business code.
- [x] Zero `c.JSON(http.StatusBadRequest/NotFound/InternalServerError, gin.H{"error":...})` remain in service handlers.
- [x] All 7 services have `utils.InitLogger` and `utils.NewResponseHelper` invoked in their `cmd/main.go`.
- [x] All 7 services: `go build ./...`, `go test ./...`, `go vet ./...` pass.
- [x] Master PRD DoD 2.22 (already done) and 2.15 (in progress) are unaffected; this PRD is additive and orthogonal.
- [x] A short `docs/architecture/adr-001-shared-utils-adoption.md` (or similar) is created explaining the pattern + how to use it for new services.

## Open Questions (to resolve before D2)

1. **Local `common-utils` symlinks vs `replace` directive**: the codebase mixes both patterns. Need to decide once for the new submodules. (Recommendation: use `replace` for Go modules; symlinks are a legacy workaround.)
2. **Auth-service `NewKafkaPublisher` divergence**: should D2 also fix the missing `AllowAutoTopicCreation`? (Recommendation: yes, normalize.)
3. **Should D4 use the `LoggingPublisher` decorator or the `LogPublishErr` helper?**: decorator is cleaner but needs more refactoring. Start with helper; promote to decorator in a follow-up.
4. **Should we also dedupe `Memory*Repository` patterns?**: out of scope for this PRD, but a follow-up could collapse the `RWMutex + map[string]*Entity` boilerplate.

## Linked Work

- Parent: `2026-06-06-1557-cdd-gap-analysis.md`
- Sister PRD: `phases/2026-06-06-1557-cdd-gap-analysis-phase-s4.5-inventory-invariant.md` (pattern of fixing code + CDD + tests in one phase — apply same discipline here)
- Follow-up candidate: `deduplicate-memory-repositories.md` (deferred)
