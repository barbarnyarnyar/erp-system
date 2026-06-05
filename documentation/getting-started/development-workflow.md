# Development Workflow

Daily development practices and workflows for efficient ERP system development.

## Daily Development Routine

### Starting Your Development Session
```bash
# Quick start (if system is already set up)
make run

# Or start infrastructure and run services locally for faster development
make dev-start
```

The `dev-start` command:
1. Starts infrastructure services (PostgreSQL, Redis, Kafka) in Docker
2. Waits for services to be ready
3. Displays instructions for running services locally

### Making Changes and Testing
```bash
# Make code changes in services/
# Hot reload will automatically restart services if using `air`

# Run unit tests for the service you're working on
cd services/fm-service
make test

# Run integration tests to ensure compatibility
make test-integration

# Check code quality
make lint
```

### Committing Changes
```bash
# Run pre-commit checks
make lint-all
make test-all

# Stage and commit changes
git add .
git commit -m "feat: add customer management endpoints"

# Push to feature branch
git push origin feature/customer-management
```

## Adding New Features

### Creating a New API Endpoint
Example: Adding a customer endpoint to the CRM service

**Step 1: Navigate to service directory**
```bash
cd services/crm-service
```

**Step 2: Create the handler**
```bash
# Create handler file
touch internal/api/handlers/customer_handler.go

# Create service logic
touch internal/business/services/customer_service.go

# Create repository
touch internal/data/repositories/customer_repository.go
```

**Step 3: Add database migration**
```bash
# Create migration for customers table
make migrate-create name=create_customers_table
```

**Step 4: Update API Gateway routing**
```bash
# Edit api-gateway configuration to include new routes
# This might be in nginx.conf, routes.yaml, or Go code depending on implementation
```

**Step 5: Test the new feature**
```bash
# Test service directly
curl -X POST http://localhost:8004/customers \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Customer","email":"test@example.com"}'

# Test through API Gateway
curl -X POST http://localhost:8080/api/v1/crm/customers \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Customer","email":"test@example.com"}'
```

## Git Workflow

### Branch Strategy
We follow a feature branch workflow:

```bash
# Create feature branch from main
git checkout main
git pull origin main
git checkout -b feature/customer-management

# Work on feature
# Make commits with descriptive messages

# Push feature branch
git push origin feature/customer-management

# Create pull request when ready
```

### Commit Message Convention
Use conventional commits for better changelog generation:

```bash
# Feature additions
git commit -m "feat: add customer management API endpoints"

# Bug fixes
git commit -m "fix: resolve database connection timeout issue"

# Documentation
git commit -m "docs: update API documentation for customer endpoints"

# Refactoring
git commit -m "refactor: extract common validation logic"

# Tests
git commit -m "test: add integration tests for customer service"
```

## Code Quality Practices

### Before Committing
Run these checks before every commit:

```bash
# Format code
gofmt -w .

# Update dependencies
go mod tidy

# Run linting
make lint

# Run tests
make test

# Check test coverage
make test-coverage
```

### Code Review Checklist
When reviewing pull requests, check:
- [ ] Code follows Go conventions and project patterns
- [ ] All tests pass
- [ ] Test coverage is maintained or improved
- [ ] API changes are documented
- [ ] Database migrations are included if needed
- [ ] No sensitive information is committed
- [ ] Error handling is appropriate
- [ ] Logging is adequate but not excessive

## Testing Strategy

### Test Types
1. **Unit Tests**: Test individual functions and methods
2. **Integration Tests**: Test service interactions
3. **API Tests**: Test HTTP endpoints
4. **End-to-end Tests**: Test complete workflows

### Running Tests
```bash
# Run all tests
make test-all

# Run tests for specific service
cd services/fm-service
make test

# Run tests with coverage report
make test-coverage

# Run integration tests
make test-integration

# Run API tests
make test-api
```

### Writing Tests
Follow these patterns for consistent testing:

**Unit Test Example:**
```go
// services/fm-service/internal/business/services/account_service_test.go
func TestAccountService_CreateAccount(t *testing.T) {
    // Arrange
    mockRepo := &MockAccountRepository{}
    service := NewAccountService(mockRepo)
    
    request := CreateAccountRequest{
        AccountCode: "1000",
        AccountName: "Cash",
        AccountType: "ASSET",
    }
    
    // Act
    result, err := service.CreateAccount(context.Background(), request)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "1000", result.AccountCode)
    assert.Equal(t, "Cash", result.AccountName)
}
```

## Debugging Workflow

### Debugging a Service Issue
1. **Check service logs**:
   ```bash
   docker-compose logs fm-service
   ```

2. **Check service health**:
   ```bash
   curl http://localhost:8001/health
   ```

3. **Check database connectivity**:
   ```bash
   docker-compose exec postgres psql -U postgres -d financial_db -c "SELECT 1;"
   ```

4. **Use debugger if needed**:
   ```bash
   cd services/fm-service
   dlv debug cmd/server/main.go
   ```

### Performance Debugging
```bash
# Check service metrics
curl http://localhost:8001/metrics

# Monitor resource usage
docker stats

# Profile Go application
go tool pprof http://localhost:8001/debug/pprof/profile
```

## Database Workflow

### Working with Migrations
```bash
# Create new migration
cd services/fm-service
make migrate-create name=add_audit_fields

# Apply migrations
make migrate-up

# Rollback if needed
make migrate-down

# Check current version
make migrate-version

# Force to specific version (use carefully)
make migrate-force version=3
```

### Database Development Practices
- Always create migrations for schema changes
- Test migrations both up and down
- Never modify existing migrations in production
- Use descriptive names for migrations
- Include both DDL and DML in migrations if needed

## Environment Management

### Switching Between Environments
```bash
# Development
export NODE_ENV=development
docker-compose up -d

# Testing
export NODE_ENV=test  
docker-compose -f docker-compose.test.yml up -d

# Staging
export NODE_ENV=staging
docker-compose -f docker-compose.staging.yml up -d
```

### Local vs Remote Development
**Local Development** (faster iteration):
```bash
# Infrastructure in Docker, services local
docker-compose up -d postgres redis kafka
cd services/fm-service && air
```

**Full Docker Development** (production-like):
```bash
# Everything in Docker
make run
```

## Troubleshooting Development Issues

### Common Development Problems
1. **Port conflicts**: Use `make check-ports` to identify conflicts
2. **Database connection issues**: Restart PostgreSQL container
3. **Service communication**: Check Docker networks with `docker network ls`
4. **Hot reload not working**: Restart air, check file permissions
5. **Tests failing**: Clear test database, run migrations

### Getting Help
- Check [Common Issues](common-issues.md) for solutions
- Review service logs: `docker-compose logs <service-name>`
- Verify environment: `make health`
- Check network connectivity: `make test-connectivity`

## Performance Tips

### Development Speed Optimization
- Use `air` for hot reloading
- Run infrastructure in Docker, services locally
- Use Go module proxy for faster downloads
- Keep database connections pooled
- Use make targets for common tasks
- Cache dependencies in Docker builds

## Next Steps

Master these workflow practices, then:
- [Testing and Verification](testing-verification.md) - Comprehensive testing guide
- [System Architecture](../architecture/README.md) - Understand the system design
- [Operations Guide](../operations/README.md) - Deploy to production