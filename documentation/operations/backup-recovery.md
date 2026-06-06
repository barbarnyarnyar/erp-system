# Backup and Recovery

## Current State

All services use **in-memory storage** with no persistence layer. Data is lost whenever a service restarts. There is no backup mechanism to describe because nothing is persisted.

## What Exists

### PostgreSQL Volume

The `docker-compose.yml` defines a named volume for PostgreSQL:

```yaml
volumes:
  postgres_data:
```

This volume persists database files across container restarts. However, **no service currently connects to PostgreSQL** — all data lives in Go maps.

### Migration Files

Every service has SQL migration files at `internal/data/migrations/schema.sql`. These define the full PostgreSQL schema (tables, foreign keys, indexes). The FM and M services have Makefile targets for running migrations:

```bash
make migrate-up
make migrate-down
make migrate-create name=migration_name
```

These targets are configured but **never executed** since no database is connected.

## Recovery Procedure (When Database Is Implemented)

### Prerequisites

1. Ensure PostgreSQL container is running: `docker compose up -d postgres`
2. Configure database env vars for each service (see `configuration.md`)
3. Run migrations:

```bash
# For each service with a Makefile
cd services/fm-service && make migrate-up
cd services/m-service && make migrate-up
```

### Backup PostgreSQL

```bash
docker exec -t erp-system-postgres-1 pg_dump -U admin erp_db > backup_$(date +%Y%m%d_%H%M%S).sql
```

### Restore PostgreSQL

```bash
docker exec -i erp-system-postgres-1 psql -U admin erp_db < backup_file.sql
```

## Known Issues

- All in-memory data is lost on restart (development only)
- No database is connected at runtime despite full schema definitions
- Migration tool (`golang-migrate`) is not a Go dependency — must be installed separately
