# Technology Stack

Technologies actually in use across the ERP system, with documented gaps between stated and actual usage.

## Go Programming Language

**Version**: 1.21+ (Dockerfiles specify `golang:1.21-alpine`, go.mod files specify `go 1.23`)

All 8 services (Auth + 6 domain + Gateway) are written in Go.

### Go Dependencies (from all go.mod files)

| Library | Version | Used By | Purpose |
|---------|---------|---------|---------|
| `github.com/gin-gonic/gin` | v1.10.0 | All services | HTTP router + middleware |
| `github.com/segmentio/kafka-go` | v0.4.51 | All 7 backend services | Kafka producer/consumer |
| `github.com/shopspring/decimal` | v1.4.0 | FM, HR, SCM, MFG, PM | Monetary value arithmetic |
| `github.com/golang-jwt/jwt/v5` | v5.2.1 | Auth, Gateway (inactive) | JWT token handling |
| `golang.org/x/crypto` | latest | Auth | bcrypt password hashing |
| `github.com/google/uuid` | v1.6.0 | Gateway (inactive) | UUID generation |
| `github.com/redis/go-redis/v9` | v9.7.1 | Auth | Redis client |

### Libraries Declared but NOT Imported

| Library | Where Declared | Why |
|---------|---------------|-----|
| `gorm.io/gorm` | None | **Not declared anywhere** — aspirational only |
| `github.com/jackc/pgx/v5` | `fm-service/go.mod` | PostgreSQL driver — declared but **never imported** |
| `github.com/lib/pq` | `hr-service/go.mod` | PostgreSQL driver — declared but **never imported** |
| `github.com/go-redis/redis` | None | **Not declared anywhere** — aspirational only |
| `github.com/cespare/xxhash/v2` | `crm-service/go.mod` | Declared but never directly imported (transitive dep of redis) |

## HTTP Framework: Gin

**Version**: v1.10.0 — used by all services.

### Usage Pattern
```go
r := gin.Default()
// or: r := gin.New()
r.GET("/health", handler.Health)
r.GET("/api/v1/accounts", handler.ListAccounts)
r.Run(fmt.Sprintf(":%s", config.ServerPort))
```

### Gin Usage by Service

| Service | Router Init | Middleware |
|---------|------------|------------|
| Auth Service | `gin.New()` | None |
| FM Service | `gin.New()` | None |
| HR Service | `gin.New()` | None |
| SCM Service | `gin.Default()` | Logger + Recovery |
| MFG Service | `gin.Default()` | Logger + Recovery |
| CRM Service | `gin.Default()` | Logger + Recovery |
| PM Service | `gin.New()` | None |
| API Gateway (active) | `gin.Default()` | Logger + Recovery |
| API Gateway (inactive) | `gin.New()` | JWT, RBAC, CORS, Rate Limit |

> **Note**: `gin.Default()` includes Logger and Recovery middleware. `gin.New()` provides a bare router.

## Data Storage

### Current: In-Memory Maps

All services store data in `sync.RWMutex`-protected Go maps:

```go
type MemoryAccountRepo struct {
    mu   sync.RWMutex
    data map[string]*domain.Account
}
```

**Characteristics**:
- No persistence — data lost on restart
- No queries, filtering, or pagination
- Full dataset returned on `List()` calls
- Zero database configuration required to run

### Declared but Disconnected: PostgreSQL

Docker Compose includes `postgres:13` on port 5432, but **no service imports a PostgreSQL driver or connects** to it. SQL migration files exist under `services/*/internal/data/migrations/` (CDD-generated) but are never executed.

### Declared but Disconnected: Redis

Docker Compose includes `redis:6` on port 6379, and `github.com/redis/go-redis/v9` is in Auth's `go.mod`. However, the Auth service stores sessions in-memory, not Redis.

## Message Queue: Kafka

**Version**: Confluent CP 7.0.1 (`confluentinc/cp-kafka:7.0.1`)

### Kafka Client Library

All 7 backend services use `github.com/segmentio/kafka-go` v0.4.51.

**Producer**: `kafka.Writer` with `LeastBytes` balancer, fire-and-forget publishing.

**Consumer**: `kafka.Reader` in a single goroutine, blocking `ReadMessage` loop.

### Kafka Topics

~85 topic constants defined across all services (CDD-generated in `event_topics.go`). Approximately 20+ topics are defined but never published.

## Authentication

### Current State
- **Auth Service** (:8000): Handles login/register with JWT + bcrypt
- **Gateway** (:8080): No auth — all requests pass through unauthenticated
- **Service-level**: No service validates tokens or forwarded headers

### Inactive Auth Infrastructure
A complete JWT + RBAC + rate limiting system exists in `api-gateway/internal/server/server.go` but is **not built into the deployed binary**. The Dockerfile compiles `cmd/main.go` (simple proxy), not `internal/server/server.go`.

## Containerization & Orchestration

### Docker Compose (Current Deployment)

| Container | Image | Exposed Port | Health Check |
|-----------|-------|-------------|--------------|
| postgres | postgres:13 | 5432 | ❌ |
| redis | redis:6 | 6379 | ❌ |
| zookeeper | confluentinc/cp-zookeeper:7.0.1 | 2181 | ❌ |
| kafka | confluentinc/cp-kafka:7.0.1 | 9092 | ❌ |
| auth-service | erp-auth-service (local build) | 8000 | ❌ |
| fm-service | erp-fm-service (local build) | 8001 | ❌ |
| hr-service | erp-hr-service (local build) | 8003 | ❌ |
| scm-service | erp-scm-service (local build) | 8006 | ❌ |
| m-service | erp-m-service (local build) | 8004 | ❌ |
| crm-service | erp-crm-service (local build) | 8002 | ❌ |
| pm-service | erp-pm-service (local build) | 8005 | ❌ |

### Dockerfile Pattern

Most services use a two-stage alpine build:

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/main /main
EXPOSE 8001
CMD ["./main"]
```

**Notable Dockerfile issues**:
- M-service and PM-service both `EXPOSE 8001` regardless of their actual default ports
- Gateway Dockerfile requires building from repo root: `docker build -f api-gateway/Dockerfile .`

### Not Deployed: Kubernetes

No Kubernetes manifests exist in the repository. The architecture docs describe Kubernetes deployment as a future goal.

## Development Tools

### Present in Repository

| Tool | Usage | Defined Where |
|------|-------|---------------|
| Make | Build/test/lint commands | Top-level `Makefile` + FT/MFG service `Makefile`s |
| golangci-lint | Go linting | CLAUDE.md (recommended) |
| air | Hot reload | CLAUDE.md (recommended) |

### Not Present in Repository

| Tool/Framework | Status |
|---------------|--------|
| OpenAPI spec / swagger | ❌ No API documentation files |
| React / TypeScript | ❌ No frontend code |
| React Native | ❌ No mobile app code |
| Terraform | ❌ No IaC files |
| Helm charts | ❌ No K8s packaging |
| GitHub Actions / CI config | ❌ No `.github/workflows/` |
| Prometheus | ❌ No metrics instrumentation |
| Grafana | ❌ No dashboard configs |
| Testify / mockery | ❌ No test frameworks besides `testing` |
| pprof | ❌ No profiling endpoints wired |

## Shared Utils (Declared but Unused)

The `shared/` directory contains:

| File | Provides | Used By |
|------|----------|---------|
| `utils/logger.go` | `Logger` struct with Info/Error/Debug/Warn + request ID | **None** (all services use `log.Printf`) |
| `utils/response.go` | `StandardResponse{Success, Message, Data, Error, Service, RequestID, Timestamp}` | **None** (all services use `gin.H`) |
| `utils/time.go` | Time formatting helpers | **None** |
| `templates/utils/` | Gin middleware templates | **None** |

Each service's `common-utils/` is a symlink to `shared/`, but no service imports from it.

## Dependency Versions (from go.mod)

| Service | go Version | Direct Dependencies |
|---------|-----------|-------------------|
| api-gateway | 1.23 | gin, jwt, uuid |
| auth-service | 1.23 | gin, jwt, crypto, redis |
| fm-service | 1.23 | gin, kafka-go, decimal, pgx (unused) |
| hr-service | 1.23 | gin, kafka-go, decimal, lib/pq (unused) |
| scm-service | 1.23 | gin, kafka-go, decimal |
| m-service | 1.23 | gin, kafka-go, decimal |
| crm-service | 1.23 | gin, kafka-go |
| pm-service | 1.23 | gin, decimal |

## Next Steps

- [System Overview](system-overview.md) — Full C4 architecture and current vs target state
- [Microservices Architecture](microservices-architecture.md) — Service patterns and communication
- [Security Architecture](security-architecture.md) — Auth service and JWT details