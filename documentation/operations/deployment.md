# Production Deployment

Deploy the ERP system to production using Docker Compose. The current codebase has no Kubernetes manifests, Helm charts, Terraform configurations, or cloud deployment scripts.

## Current Deployment

### Docker Compose (Only Deployment Method)

```bash
# Start all services
make run

# Build all services
make build

# Stop all services
make stop

# Full cleanup (removes images, volumes, orphans)
make clean
```

### Infrastructure Dependencies

The system requires these infrastructure services, all defined in `docker-compose.yml`:

| Service | Image | Port | Purpose | Used By |
|---------|-------|------|---------|---------|
| `postgres` | `postgres:13` | 5432 | Relational database | **Not connected** — all services use in-memory |
| `redis` | `redis:6` | 6379 | Cache and session store | **Not connected** — no Redis client in any go.mod |
| `zookeeper` | `confluentinc/cp-zookeeper:7.0.1` | 2181 | Kafka coordination | Kafka |
| `kafka` | `confluentinc/cp-kafka:7.0.1` | 9092, 29092 | Event messaging | All services for fire-and-forget event publishing |

### Application Services

| Service | Build Context | Port | Default Config |
|---------|-------------|------|----------------|
| `auth-service` | `./services/auth-service` | 8000 | `PORT=8000` |
| `fm-service` | `./services/fm-service` | 8001 | `PORT=8001` |
| `crm-service` | `./services/crm-service` | 8002 | `PORT=8002` |
| `hr-service` | `./services/hr-service` | 8003 | `PORT=8003` |
| `m-service` | `./services/m-service` | 8004 | `PORT=8004` |
| `pm-service` | `./services/pm-service` | 8005 | `PORT=8005` |
| `scm-service` | `./services/scm-service` | 8006 | `PORT=8006` |

> **Important**: The API Gateway is **not included** in docker-compose.yml. It must be built and run separately.

### Starting the API Gateway

```bash
# From repository root (required — gateway needs shared/ module)
cd api-gateway
go run cmd/main.go

# Or build and run
cd api-gateway && go build -o bin/main cmd/main.go && ./bin/main
```

> The gateway Dockerfile expects the build context to include both `shared/` and `api-gateway/`:
> ```bash
> # Build from repo root
> docker build -f api-gateway/Dockerfile .
> ```

## Health Checks

```bash
# Check all services via Makefile
make health

# Check individual service directly
curl http://localhost:8000/health   # Auth
curl http://localhost:8001/health   # FM
curl http://localhost:8002/health   # CRM
curl http://localhost:8003/health   # HR
curl http://localhost:8004/health   # M
curl http://localhost:8005/health   # PM
curl http://localhost:8006/health   # SCM

# Via API Gateway
curl http://localhost:8080/health
```

## Known Deployment Issues

1. **API Gateway not in docker-compose.yml**: The gateway must be built and run separately — `docker compose up` does not start it.

2. **Dockerfile port mismatches**: M-service and PM-service Dockerfiles `EXPOSE 8001` but services default to `8004` and `8006` respectively. Set `PORT` at runtime to override.

3. **Build script naming mismatch**: `scripts/build.sh` uses directory names (`finance`, `manufacturing`, `projects`) that differ from docker-compose service names (`fm-service`, `m-service`, `pm-service`).

4. **Go version discrepancies**: Dockerfiles use `golang:1.21-alpine` but `go.mod` files specify `go 1.23.0`. The builder auto-downloads the newer toolchain at build time.

5. **No database persistence**: All services use in-memory storage. No service connects to PostgreSQL at runtime. Data is lost on container restart.

6. **Port discrepancies between code defaults and documented ports**:

   | Service | Code Default | Docker Compose Port |
   |---------|-------------|-------------------|
   | HR | 8003 | 8003 |
   | CRM | 8002 | 8002 |
   | SCM | 8006 | 8006 |

## Security

The deployed API Gateway has **no authentication**. All endpoints are publicly accessible. See [Security Configuration](security.md) for details on the inactive JWT auth system.

## Monitoring

There is **no monitoring infrastructure** deployed. See [Monitoring and Alerting](monitoring.md) for current state and recommendations.

## Next Steps

- [Configuration Management](configuration.md) — Environment-specific settings
- [Troubleshooting](troubleshooting.md) — Common issues and solutions
- [System Maintenance](maintenance.md) — Routine tasks
