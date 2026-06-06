# Monitoring and Alerting

## Current State

There is **no monitoring infrastructure** deployed. No Prometheus, Grafana, structured logging, distributed tracing, or centralized alerting is implemented.

## Available Endpoints

### Health Checks

Every service exposes `GET /health`:

```bash
curl http://localhost:8001/health
# {"service":"fm-service","status":"healthy","port":"8001"}
```

The inactive API Gateway server (`internal/server/server.go`) has an admin endpoint that aggregates health across all services:

```
GET /api/v1/admin/services/status
```

This endpoint is **not deployed** in the running gateway.

### Makefile Health Check

```bash
make health
```

This curls each service's `/health` endpoint sequentially.

## Shared Utilities (Available But Not Used)

The `common-utils/` symlink (→ `shared/templates/utils`) provides:

### Logger

```go
type Logger struct {
    ServiceName string
    RequestID   string
}

func (l *Logger) Info(msg string)
func (l *Logger) Error(msg string)
func (l *Logger) Debug(msg string)
func (l *Logger) Warn(msg string)
```

Includes `GinLogger()` middleware for structured HTTP request logging with request ID support, and `RequestIDMiddleware()` to inject `X-Request-ID` headers.

### Response Helper

```go
type StandardResponse struct {
    Success   bool        `json:"success"`
    Message   string      `json:"message"`
    Data      interface{} `json:"data,omitempty"`
    Error     string      `json:"error,omitempty"`
    Service   string      `json:"service,omitempty"`
    RequestID string      `json:"request_id,omitempty"`
    Timestamp time.Time   `json:"timestamp"`
}
```

**Neither the logger nor the response helper is currently used by any service.** All services use `log.Printf` and raw `gin.H` responses.

## What Should Be Implemented

### Metrics

| Metric | Source | Tool |
|--------|--------|------|
| Request latency | Each service | Prometheus histogram |
| Error rate | Each service | Prometheus counter |
| Queue depth | Kafka consumer | Prometheus gauge |
| Event throughput | Kafka producer | Prometheus counter |
| Memory/CPU | Container | cAdvisor + Prometheus |

### Logging

| Requirement | Current State |
|-------------|--------------|
| Structured JSON | Not implemented — uses `log.Printf` |
| Log levels | Not implemented |
| Request IDs | Available in shared utils but not used |
| Centralized collection | Not implemented |

### Tracing

No correlation IDs or trace contexts are propagated between services. There is no way to trace a request across service boundaries.

## Recommended Setup

```bash
# Prometheus + Grafana (add to docker-compose)
prometheus:
  image: prom/prometheus
  ports: ["9090:9090"]

grafana:
  image: grafana/grafana
  ports: ["3000:3000"]
```

Wire `GinLogger()` middleware in each service for structured request logging. Export metrics via Prometheus client library.
