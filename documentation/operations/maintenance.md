# System Maintenance

## Routine Maintenance Tasks

### Restarting Services

```bash
# Restart all services
docker compose restart

# Restart a single service
docker compose restart fm-service

# Rebuild and restart a single service
docker compose up -d --build fm-service
```

### Viewing Logs

```bash
# All services
docker compose logs -f

# Single service
docker compose logs -f fm-service

# Last 100 lines
docker compose logs --tail=100 fm-service
```

### Database Migrations

When a database backend is eventually connected, migrations can be run:

```bash
# FM service (has Makefile with migration targets)
cd services/fm-service
make migrate-up

# M service
cd services/m-service
make migrate-up
```

Migration files are located at `internal/data/migrations/schema.sql` for each service.

### Updating Dependencies

```bash
# For a single service
cd services/fm-service
go mod tidy
go mod download

# All services via build script
./scripts/build.sh
```

## Build Maintenance

### Rebuilding All Services

```bash
# Docker
make build

# Local binaries
./scripts/build.sh

# Individual service
cd services/fm-service && go build -o bin/main cmd/server/main.go
```

### Cleaning Build Artifacts

```bash
# Docker (removes all images, volumes, orphans)
make clean

# Individual service
cd services/fm-service && make clean
```

## Known Maintenance Issues

- **No database** — migration commands are configured but no database is connected
- **No health check aggregation** — the `/health` endpoint exists per-service but no centralized health monitoring is deployed
- **No Prometheus/Grafana** — metrics collection and dashboarding are not implemented
- **No structured logging** — all services use `log.Printf` with no log levels or structured output
- **Gateway Dockerfile requires root context** — building from `api-gateway/` directory fails; must build from repo root
