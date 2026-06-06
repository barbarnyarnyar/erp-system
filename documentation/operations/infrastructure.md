# Infrastructure Setup

## Docker Compose Stack

The ERP system runs 11 containers defined in `docker-compose.yml` (Docker Compose v3.8).

### Infrastructure Containers

| Service | Image | Port | Purpose |
|---------|-------|------|---------|
| `postgres` | `postgres:13` | 5432 | Relational database (not currently used by any service) |
| `redis` | `redis:6` | 6379 | Cache and session store (not currently used) |
| `zookeeper` | `confluentinc/cp-zookeeper:7.0.1` | 2181 | Kafka coordination |
| `kafka` | `confluentinc/cp-kafka:7.0.1` | 9092, 29092 | Event messaging |

### Application Containers

| Service | Build Context | Port | Default Config |
|---------|-------------|------|----------------|
| `auth-service` | `./services/auth-service` | 8000 | `PORT=8000` |
| `fm-service` | `./services/fm-service` | 8001 | `PORT=8001` |
| `crm-service` | `./services/crm-service` | 8002 | `PORT=8002` |
| `hr-service` | `./services/hr-service` | 8003 | `PORT=8003` |
| `m-service` | `./services/m-service` | 8004 | `PORT=8004` |
| `pm-service` | `./services/pm-service` | 8005 | `PORT=8005` |
| `scm-service` | `./services/scm-service` | 8006 | `PORT=8006` |

> The API Gateway is **not included** in docker-compose.yml. It must be built and run separately.

## Starting Infrastructure Only

For local Go development without Docker for the services:

```bash
docker compose up -d postgres kafka redis
```

## Starting All Services

```bash
make run
```

This runs `docker compose up -d` which starts all 11 containers.

## Network

All containers communicate over the default Docker Compose bridge network. Services reference each other by container name (e.g., `http://fm-service:8001`).

## Persistent Storage

A single named volume is defined:

```yaml
volumes:
  postgres_data:
```

Mounts to `/var/lib/postgresql/data` in the PostgreSQL container.

## Known Issues

- API Gateway must be started separately — not in docker-compose
- PostgreSQL and Redis are configured but no service connects to them at runtime
- All service data is in-memory and lost on container restart
- No resource limits or health check configurations are set on containers
