# Performance Architecture

Current performance characteristics, concurrency patterns, and infrastructure considerations.

## Current State

The ERP system uses **in-memory storage** with no database connection at runtime. All performance characteristics described here reflect the current in-memory implementation, with notes on what changes when a database backend is introduced.

## Concurrency Model

### In-Memory Repository Pattern

All 7 services (Auth, FM, HR, SCM, M, CRM, PM) use the same concurrency pattern for data access:

```go
type MemoryXxxRepo struct {
    mu   sync.RWMutex
    data map[string]domain.Xxx
}
```

- **Reads** (`GetByID`, `List`): acquire `RLock()` — multiple concurrent readers allowed
- **Writes** (`Create`, `Update`, `Delete`): acquire `Lock()` — exclusive access

This provides safe concurrent access within a single process. Under the in-memory implementation, all operations are O(1) map lookups with no I/O overhead.

### Read vs Write Throughput

| Operation | Lock | Expected Latency (in-memory) |
|-----------|------|------------------------------|
| Get by ID | `RLock` | < 1μs |
| List all | `RLock` | < 10μs (copy to slice) |
| Create | `Lock` | < 1μs |
| Update | `Lock` | < 1μs |
| Delete | `Lock` | < 1μs |

## Infrastructure Performance

### Kafka Messaging

All services use `segmentio/kafka-go` with `kafka.LeastBytes` balancer for producer writes. Event publishing is:

- **Asynchronous** — the caller does not wait for downstream processing
- **Non-blocking** — publisher errors are silently discarded (`_ = publisher.Publish(...)`)
- **Batchable** — `kafka.Writer` internally batches messages

Consumer performance characteristics:

- Single goroutine per service consuming messages sequentially
- Blocking `ReadMessage()` call — one message at a time
- 2-second sleep on error before retry
- No parallel consumer goroutines within a service

### HTTP Request Handling

All services use **Gin** framework with default settings:

- `gin.Default()` includes Logger and Recovery middleware
- No custom write timeouts or read timeouts configured
- No request body size limits
- No connection pooling (handled by Go's `net/http`)

### ID Generation

All entities use nanosecond-timestamp-based IDs:

```go
id := fmt.Sprintf("prefix_%d", time.Now().UnixNano())
```

Under high concurrency (> 1 million IDs per second on the same machine), this could produce collisions. For current development scale this is not a concern.

## Caching

### Redis Configuration

Redis 6 is defined in `docker-compose.yml` and the FM service has a full Redis configuration block:

```go
type RedisConfig struct {
    Host     string
    Port     string
    Password string
    DB       int
}
```

**However**, Redis is **never connected or used** by any service at runtime. No service imports a Redis client library, and `go.mod` files contain no Redis dependency.

### Application-Level Caching

No application-level caching is implemented anywhere in the codebase. Every request reads directly from in-memory storage.

## Rate Limiting

An in-memory per-IP sliding-window rate limiter is defined in the API Gateway (`internal/middleware/rate_limit.go`) but is **not wired into any route**. When activated, it provides:

- Configurable requests-per-minute threshold
- Per-IP tracking using `sync.RWMutex`-protected map
- Returns HTTP 429 with `retry_after` seconds on throttle

## Bottlenecks

### Current (In-Memory)

| Bottleneck | Impact | Severity |
|------------|--------|----------|
| All data lost on restart | No persistence | Critical |
| No pagination on list endpoints | Full dataset returned every time | Medium |
| Single Kafka consumer goroutine | Sequential message processing | Medium |
| Silent event publish errors | Events lost without recovery | High |
| No connection limits | Unbounded goroutine growth | Low |

### Future (with Database)

| Bottleneck | Impact | Severity |
|------------|--------|----------|
| N+1 queries on aggregate endpoints | Multiple round-trips per request | High |
| Missing indexes | Full table scans on lookups | High |
| No database connection pooling | Connection buildup under load | Medium |
| Sequential consumer processing | Lag under high event volume | Medium |

## Scalability Considerations

### Horizontal Scaling

All services are **stateless** (data is in-memory, not persisted), which means:

- Currently cannot scale horizontally — each instance has its own isolated data
- With a shared database backend, all services become horizontally scalable
- API Gateway is already stateless and horizontally scalable

### Memory Constraints

In-memory storage means total data size is limited by process memory:

| Entity Count (per service) | Approximate Memory |
|---------------------------|-------------------|
| 1,000 records | ~1-5 MB |
| 100,000 records | ~100-500 MB |
| 1,000,000 records | ~1-5 GB |

Each service stores data in Go maps with string keys and struct values. A single entity with ~15 fields (~200 bytes) uses approximately 300-400 bytes total with map overhead.

## Monitoring

### Current State

No monitoring, metrics collection, or observability infrastructure is implemented:

- **No Prometheus metrics** — defined in architecture docs but not present in any service
- **No structured logging** — services use `log.Printf` and Gin's default logger
- **No distributed tracing** — no correlation IDs propagated between services
- **No health check aggregation** — each service has `/health` but no centralized monitoring

### Available Infrastructure

The shared utilities (`common-utils` → `shared/templates/utils`) include:

- **`GinLogger()` middleware** — structured HTTP request logging with request ID
- **`RequestIDMiddleware()`** — adds `X-Request-ID` header

These are **not currently used** by any service — handlers use inline `gin.H` responses and `log.Printf`.

## Optimization Recommendations

### Immediate (Low Effort)

1. **Wire the shared logger middleware** — zero-dependency improvement to request observability
2. **Add pagination to list endpoints** — accept `page` and `page_size` query parameters
3. **Remove the `fmt.Sprintf` ID generation** — use UUIDs to avoid collision risk

### Short-Term

4. **Implement database backend** — replace in-memory repos with PostgreSQL
5. **Add database indexes** — on foreign key columns and frequently queried fields
6. **Configure connection pooling** — set `MaxOpenConns`, `MaxIdleConns`, `ConnMaxLifetime`

### Long-Term

7. **Wire Redis caching** — cache frequently accessed entities (accounts, products, employees)
8. **Parallelize Kafka consumers** — use multiple goroutines per topic partition
9. **Add dead-letter queues** — prevent message loss on processing errors
10. **Implement Prometheus metrics** — request latency, error rates, queue depth
11. **Add distributed tracing** — propagate trace IDs via HTTP headers and Kafka message headers
