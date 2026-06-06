# Development Workflow

Daily development practices for working on the ERP system.

## Starting Your Development Session

```bash
# Quick start (if system is already set up)
make run

# Or start infrastructure in Docker and run services locally
docker compose up -d postgres kafka redis
```

## Making Changes

```bash
# Make code changes in services/ directory
# If using air, services will hot-reload automatically

# Run tests for the service you're working on
cd services/fm-service
make test

# Run linter
make lint

# Format code
gofmt -w .
```

## Adding a New API Endpoint

Example: Adding an endpoint to the CRM service.

**1. Create the handler:**
```bash
cd services/crm-service
touch internal/api/handlers/customer_handler.go
```

**2. Add repository methods** in `internal/data/memory/memory_repos.go` (in-memory) or add a new file.

**3. Register the route** in `internal/api/routes/routes.go`.

**4. Test the endpoint:**
```bash
# Direct access
curl -X POST http://localhost:8005/api/v1/crm/customers \
  -H "Content-Type: application/json" \
  -d '{"company_name":"Test Corp","contact_name":"John","email":"john@test.com"}'

# Through API Gateway
curl -X POST http://localhost:8080/api/v1/crm/customers \
  -H "Content-Type: application/json" \
  -d '{"company_name":"Test Corp","contact_name":"John","email":"john@test.com"}'
```

## Git Workflow

### Branch Strategy

```bash
# Create feature branch from main
git checkout main
git pull origin main
git checkout -b feature/your-feature

# Work, commit, push
git add .
git commit -m "feat: add customer management endpoints"
git push origin feature/your-feature

# Create pull request
```

### Commit Message Convention

```bash
git commit -m "feat: add customer management API endpoints"
git commit -m "fix: resolve account lookup timeout"
git commit -m "docs: update API documentation"
git commit -m "refactor: extract common validation logic"
```

## Testing

### Where to Find Tests

There is **one test file** in the entire codebase:
- `services/fm-service/internal/business/service/service_test.go` (102 lines, 2 test cases)

All other services have zero test coverage.

### Running Tests

```bash
# FM service only
cd services/fm-service
make test

# With coverage
make test-coverage
```

### API Verification

```bash
# Through API Gateway
make test

# Direct to services
make test-direct
```

## Debugging

```bash
# Check service logs
docker compose logs fm-service

# Check service health
curl http://localhost:8001/health

# Debug with Delve
cd services/fm-service
dlv debug cmd/server/main.go
```

## Database Work

**Note**: No database is currently connected. All data is in-memory and lost on restart.

When a database backend is implemented:
- Migration files exist at `internal/data/migrations/schema.sql` per service
- FM and M services have `make migrate-up` / `make migrate-down` commands
- Migration tool (`golang-migrate`) must be installed separately

## Common Issues

- **Data lost on restart**: All services use in-memory storage. Seed data is recreated on startup for some services.
- **No auth**: The deployed gateway has no authentication. All endpoints are public.
- **Kafka optional**: Event publishing errors are silently ignored.
- **Port mismatches**: See [Common Issues](common-issues.md) for details.

## Next Steps

- [Testing and Verification](testing-verification.md) — Comprehensive testing guide
- [System Architecture](../architecture/README.md) — Understand the system design
- [Operations Guide](../operations/README.md) — Deployment reference
