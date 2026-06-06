# Performance Optimization

## Current Performance Characteristics

All services use in-memory storage with `sync.RWMutex`-protected maps. All CRUD operations are O(1) map lookups with no I/O overhead.

| Operation | Lock | Expected Latency |
|-----------|------|-----------------|
| Get by ID | `RLock` | < 1μs |
| List all | `RLock` | < 10μs |
| Create/Update/Delete | `Lock` | < 1μs |

## Bottlenecks

### Current (In-Memory)

| Bottleneck | Impact |
|------------|--------|
| No pagination on list endpoints | Full dataset returned every request |
| Single Kafka consumer goroutine | Sequential message processing |
| Fire-and-forget event publishing | Silent event loss |
| No connection limits | Unbounded goroutine growth |

### Future (with Database)

| Bottleneck | Impact |
|------------|--------|
| N+1 queries on aggregate endpoints | Multiple round-trips per request |
| Missing indexes | Full table scans |
| No connection pooling | Connection buildup under load |

## Caching

**Redis is not used.** Although Redis 6 is in docker-compose and FM service has Redis config, no service connects to it. No Redis client library exists in any `go.mod`.

## Rate Limiting

An in-memory per-IP rate limiter exists in `api-gateway/internal/middleware/rate_limit.go` but is **not wired** into any route.

## Memory Constraints

In-memory storage limits total data to process memory:

| Records | Approximate Memory |
|---------|-------------------|
| 1,000 | ~1-5 MB |
| 100,000 | ~100-500 MB |
| 1,000,000 | ~1-5 GB |

## Optimization Recommendations

### Immediate

1. **Wire the shared logging middleware** — zero-cost improvement to observability
2. **Add pagination** — accept `page` and `page_size` parameters on list endpoints
3. **Use UUIDs** instead of nanosecond-timestamp IDs to avoid collision risk

### Short-Term

4. **Implement database backend** — replace in-memory repos with PostgreSQL
5. **Add database indexes** — on FK columns and frequently queried fields
6. **Configure connection pooling** — `MaxOpenConns`, `MaxIdleConns`, `ConnMaxLifetime`

### Long-Term

7. **Wire Redis caching** — cache frequently accessed entities
8. **Parallelize Kafka consumers** — multiple goroutines per topic partition
9. **Add dead-letter queues** — prevent message loss
10. **Implement Prometheus metrics** — request latency, error rates, queue depth
