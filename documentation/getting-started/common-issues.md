# Common Issues

Known issues, inconsistencies, and technical debt across the ERP codebase, discovered through code analysis of all 8 services.

## Index by Severity

| Severity | Count |
|----------|-------|
| Critical | 4 |
| High | 12 |
| Medium | 15 |
| Low | 10 |

---

## Critical

### C1. No Authentication Deployed

The API Gateway has two code paths. The deployed path (`cmd/main.go`) is a simple reverse proxy with **no authentication, authorization, or rate limiting**. The full JWT auth system (token validation, permission checks, role checks) is defined in `internal/server/server.go` and `internal/middleware/auth.go` but is **never wired into the running binary**. All API endpoints are publicly accessible.

**Affects**: API Gateway
**Files**: `api-gateway/cmd/main.go`, `api-gateway/internal/server/server.go`

### C2. Plaintext Passwords

The Auth Service stores and compares passwords as plaintext. The code explicitly documents this as a simplification but it remains in the codebase:

```go
// Simple check (in production, use bcrypt or secure comparison)
if user.PasswordHash != password {
```

No hashing (bcrypt, scrypt, or argon2), no salting, no peppering.

**Affects**: Auth Service
**Files**: `services/auth-service/internal/business/service/auth_service.go:57`, `services/auth-service/internal/api/handlers/identity_handler.go:49`

### C3. Hardcoded JWT Secret

The default JWT signing secret is hardcoded as `super-secret-key-123` in `config.go`. Visible in source code and accessible to anyone with repository access.

**Affects**: Auth Service, API Gateway
**Files**: `services/auth-service/internal/config/config.go:42`, `api-gateway/internal/middleware/auth.go`

### C4. All Data Lost on Restart

Every service uses in-memory storage (`map[string]T` with `sync.RWMutex`). Despite PostgreSQL being in docker-compose, SQL schemas existing in every service, and database configuration being loadable, no service actually connects to a database. All data is lost when any service restarts.

**Affects**: All services
**Files**: All `internal/data/memory/memory_repos.go` files

---

## High

### H1. Port Inconsistencies Across Services

Multiple services have code-default ports that differ from both docker-compose configuration and architecture documentation:

| Service | Code Default | Docker Compose | Architecture Doc |
|---------|-------------|----------------|-----------------|
| CRM | 8002 | 8002 | 8005 |
| HR | 8003 | 8003 | 8002 |
| SCM | 8006 | 8006 | 8003 |

Additionally, M-service and PM-service Dockerfiles hardcode `EXPOSE 8001` while the services default to 8004 and 8006 respectively.

**Affects**: CRM, HR, SCM, M, PM services

### H2. Three Different Naming Conventions for the Same Services

Service names vary across the build script, Docker Compose, and the two gateway code paths:

| Logical Service | Build Script | Docker Compose | Gateway (active) | Gateway (inactive) |
|----------------|-------------|----------------|-------------------|-------------------|
| Financial | `finance` | `fm-service` | `finance-service` | `fm-service` |
| Manufacturing | `manufacturing` | `m-service` | `manufacturing-service` | `m-service` |
| Project Mgmt | `projects` | `pm-service` | `projects-service` | `pm-service` |

The Makefile test routes use yet another set: `/api/v1/finance/hello`, `/api/v1/manufacturing/hello`, `/api/v1/projects/hello`.

**Affects**: Build pipeline, gateway routing, Makefile commands

### H3. RabbitMQ Config Defined But Never Used

The FM service has a full `RabbitMQConfig` struct, `.env.example` entries, and `Makefile` references for RabbitMQ. The codebase uses **only Kafka** for messaging and has no RabbitMQ client dependency.

**Affects**: FM Service
**Files**: `services/fm-service/internal/config/config.go`, `services/fm-service/.env.example`

### H4. Dead Code — Unused Service Components

Two sets of service components are fully implemented but never instantiated:

- **FM**: `TaxService` — has repository, service layer, and in-memory implementation but is never wired in `main.go`
- **M**: `QualityService` and `CostingService` — defined as standalone components with full logic but never wired; `ProductionService` handles all their responsibilities

**Affects**: FM Service, M Service

### H5. Most Kafka Events Are Logged Only

Across all services, only 6 out of ~50 consumed event topics have real business logic side-effects. The vast majority log the event payload and return nil. Comments describe intended future behavior that was never implemented:

- HR consumer: 1 real handler out of 5 topics
- CRM consumer: 1 real handler out of 7 topics
- PM consumer: 1 real handler out of 7 topics
- FM consumer: 13 real handlers (the only well-implemented consumer)

**Affects**: HR, CRM, PM, SCM services

### H6. Income Statement and Cash Flow Reports Are Stubs

The FM service's `GET /api/v1/reports/income-statement` and `GET /api/v1/reports/cash-flow` return placeholder strings. Only the balance sheet is actually implemented. The `GeneralLedgerService` has no `GetIncomeStatement()` or `GetCashFlow()` methods.

**Affects**: FM Service
**Files**: `services/fm-service/internal/api/handlers/report_handler.go`

### H7. Rate Limiter Defined But Not Wired

The API Gateway has a complete in-memory per-IP sliding-window rate limiter in `internal/middleware/rate_limit.go` but it is **never registered** on any router group in either entry point.

**Affects**: API Gateway

### H8. Bug: PO Line Update Uses Create Instead of Update

In SCM's `WarehouseService.CreateReceipt`, when updating a purchase order line's received quantity, the code calls `Create` on the line repository instead of `Update`. This creates a **new duplicate entry** rather than modifying the existing one. The `MemoryPurchaseOrderLineRepo` has no `Update` method.

```go
// Creates a new duplicate entry instead of updating
_ = s.poLRepo.Create(ctx, &pol)
```

**Affects**: SCM Service
**Files**: `services/scm-service/internal/business/service/warehouse_service.go:121`

### H9. Fire-and-Forget Event Publishing

Every service silently discards Kafka publish errors:

```go
_ = s.publisher.Publish(ctx, topic, key, payload)
```

If Kafka is down, events are lost with no retry, circuit-breaking, or dead-letter queue. No service checks the error return from `Publish`.

**Affects**: All services

### H10. No Test Coverage

The codebase has only one test file across all 8 services: `services/fm-service/internal/business/service/service_test.go` (102 lines, testing `CreateAccount` and `CreateInvoice` event publishing). No other service has any tests.

**Affects**: All services

### H11. Printf Logging Instead of Structured Logging

All services use `log.Printf` and Gin's default logger for all logging. The shared `common-utils` symlink provides a structured `Logger` with request ID support and log levels (`Info`, `Error`, `Debug`, `Warn`), but **no service imports or uses it**.

**Affects**: All services

### H12. No Pagination on List Endpoints

Every list endpoint (`GET /api/v1/xxx`) returns the full dataset with no pagination support. Under in-memory storage this is acceptable, but it will become a bottleneck when a database backend is added.

**Affects**: All services

---

## Medium

### M1. No Makefiles on Most Services

Only FM and M services have individual `Makefiles`. HR, SCM, CRM, PM, and Auth have no Makefile, despite the project `CLAUDE.md` documenting per-service Makefile commands as standard.

**Affects**: HR, SCM, CRM, PM, Auth services

### M2. `common-utils` Symlink Unused

Every service has a `common-utils/` symlink pointing to `../../shared/templates/utils`, which provides a structured `Logger` and `StandardResponse` helper. No service imports or uses these utilities. All handlers use raw `gin.H` and `log.Printf`.

**Affects**: All services

### M3. ID Generation Uses Nanosecond Timestamps Not UUIDs

All entities generate IDs as:

```go
id := fmt.Sprintf("prefix_%d", time.Now().UnixNano())
```

The CDD contracts and SQL schemas specify `UUID PRIMARY KEY`, but the Go code uses predictable, non-cryptographic string IDs. Under high concurrency (>1M/sec), collisions are possible.

**Affects**: All services

### M4. Unused Kafka Topic Constants

Approximately 20+ event topic constants are defined across services but never published. These represent planned features or contract specifications that were never implemented:

- HR: `hr.certification.earned`, `hr.skill.acquired`, `hr.goal.achieved`, `hr.employee.available`, `hr.employee.skills.updated`, `hr.payroll.failed`
- PM: `prj.project.delayed`, `prj.task.overdue`, `prj.resource.released`, `prj.resource.overallocated`, `prj.time.rejected`, `prj.expense.rejected`, `prj.milestone.achieved`, `prj.milestone.delayed`
- SCM: `scm.product.*`, `scm.shipment.*`, `scm.vendor.*`, `scm.training.required`, `scm.material.delivered`

**Affects**: HR, PM, SCM services

### M5. Go Version Discrepancy

Dockerfiles specify `golang:1.21-alpine` but `go.mod` files specify `go 1.23.0` with toolchain `go1.24.6`. The builder auto-downloads the newer toolchain at build time, adding build overhead.

**Affects**: All services

### M6. Redis Configured But Never Used

Redis 6 is defined in `docker-compose.yml`. The FM service has a full `RedisConfig` struct with host, port, password, and DB settings. However, no service imports a Redis client library or connects to Redis at runtime. No `go.mod` contains a Redis dependency.

**Affects**: FM Service, infrastructure

### M7. Data Layer Mismatch — UUID Schemas vs String IDs

The PostgreSQL schemas (`internal/data/migrations/schema.sql`) define all primary keys as `UUID` type, but the in-memory implementation uses Go strings. There is no database driver or migration runner actually connected at runtime.

**Affects**: All services

### M8. Missing Authentication in CRUD Handlers

Several domain entities have repositories and interfaces defined but are not manageable through any API:
- FM: `MarkInvoiceOverdue` — service method exists but no HTTP route
- HR: `Department` and `Position` — repos/interfaces exist but never instantiated or exposed
- SCM: Several event-publish-only topics have no endpoint triggering them

**Affects**: FM, HR, SCM services

### M9. JSON Decimal Parsing Errors Silently Ignored

In FM's handlers, `decimal.NewFromString` failures are silently ignored, defaulting to `decimal.Zero` instead of returning an error:

```go
amt, _ := decimal.NewFromString(req.Amount)
```

This means invalid numeric input silently becomes zero without the caller knowing.

**Affects**: FM Service

### M10. BOM Components Accepted But Ignored

The M service's `POST /api/v1/boms` handler accepts a `components` field in its request body but the service method `CreateBillOfMaterials` does not accept or process components. The field is silently ignored.

**Affects**: M Service

### M11. `bill_of_materialss` Typo in Schema

The M service's PostgreSQL migration names the BOM table `bill_of_materialss` (double `s`). The `equipment` table is also incorrectly pluralized as `equipments`.

**Affects**: M Service
**Files**: `services/m-service/internal/data/migrations/schema.sql`

### M12. Refresh Tokens Follow Predictable Pattern

Auth service refresh tokens use the format `rt_{unix_nano}_{user_id}`, making them enumerable. An attacker who can observe one refresh token can predict future tokens for the same user.

**Affects**: Auth Service

### M13. User ID Type Mismatch Between Services

The Auth Service uses string user IDs (e.g., `usr_1749267184000000000`). The API Gateway's `auth_client.go` middleware parses `X-User-ID` as `uint`. If the downstream auth middleware were ever activated, it would fail to parse the string IDs.

**Affects**: API Gateway, Auth Service

### M14. Consumer Uses Only Single Service Reference

The HR Kafka consumer receives all 9 services in its constructor but only uses `TrainingService` (for the `scm.training.required` topic). The other 8 services are passed but never called from the consumer.

**Affects**: HR Service

### M15. Seed Data Assumes Current Time

Several services seed mock data with comments like "Started 1 month ago" and "Ends in 5 months" using `time.Now()`. These dates will be stale relative to the comment but are dynamically correct at runtime.

**Affects**: PM Service

---

## Low

### L1. No `.gitignore` in Individual Service Directories

No service directory has its own `.gitignore`. Binary artifacts (`bin/`) from local builds could be accidentally committed.

### L2. Empty Scripts

`scripts/deploy.sh` and `scripts/setup-dev.sh` are empty files with no implementation.

### L3. Dockerfile Build Context Assumptions

The API Gateway Dockerfile expects to be built from the repository root (not from `api-gateway/`) because it `COPY shared/ ./shared/`. This differs from all other service Dockerfiles which build from their own directory.

**File**: `api-gateway/Dockerfile`

### L4. Inventory Delete Returns Success Without Deleting

SCM's `DeleteInventoryItem` handler returns a success message with the item ID without actually calling the service to delete it.

**File**: `services/scm-service/internal/api/handlers/inventory_handler.go`

### L5. `PARTIALLY_DELIVERED` Status Not in Domain Model

SCM's `WarehouseService.CreateReceipt` can set a PO status to `PARTIALLY_DELIVERED`, but this value is not listed in the `PurchaseOrder.Status` domain model's documented enum values.

**File**: `services/scm-service/internal/business/service/warehouse_service.go`

### L6. Journal Entries Always Posted

FM's `CreateJournalEntry` always sets `Status: "POSTED"` even though the `Transaction` domain model defines a `PENDING → POSTED → REVERSED` state machine.

### L7. Recruitment and Document Services Publish No Events

HR's `RecruitmentService` and `EmployeeDocumentService` are the only services that have no event publishing at all, despite managing significant domain operations (hiring pipeline, document uploads).

### L8. PriceListService Has No Event Publisher

CRM's `PriceListService` is the only service in the CRM that does not accept an `EventPublisher` dependency — it publishes no events.

### L9. Async Response for Synchronous Operations

Several service methods publish Kafka events and return HTTP responses in the same handler, mixing synchronous (HTTP response) and asynchronous (Kafka event) concerns. The client gets a 200 response before the event is necessarily consumed.

### L10. No Distributed Tracing

No correlation IDs or trace contexts are propagated between services via HTTP headers or Kafka message headers. There is no way to trace a request across service boundaries.
