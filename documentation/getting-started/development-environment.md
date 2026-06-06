# Development Environment

Set up your local development environment for efficient coding.

## Go Module Setup

```bash
# For a specific service
cd services/fm-service
go mod tidy
go mod download
```

## Local Development Without Docker

Run services locally while keeping infrastructure in Docker:

### Step 1: Start Infrastructure Only

```bash
docker compose up -d postgres kafka redis
# Note: PostgreSQL and Redis are not connected to services yet — they run for future use.
```

### Step 2: Run Services Locally

```bash
# Terminal 1: Financial Service
cd services/fm-service
export KAFKA_BROKERS=localhost:9092
go run cmd/server/main.go

# Terminal 2: HR Service
cd services/hr-service
export PORT=8002
export KAFKA_BROKERS=localhost:9092
go run cmd/main.go

# Continue for other services on their respective ports...
```

### Step 3: API Gateway

```bash
cd api-gateway
export FINANCE_SERVICE_URL=http://localhost:8001
export HR_SERVICE_URL=http://localhost:8002
export SCM_SERVICE_URL=http://localhost:8003
export MANUFACTURING_SERVICE_URL=http://localhost:8004
export CRM_SERVICE_URL=http://localhost:8005
export PROJECTS_SERVICE_URL=http://localhost:8006
go run cmd/main.go
```

## Hot Reload Development

```bash
# Install air globally
go install github.com/cosmtrek/air@latest

# Run service with hot reload
cd services/fm-service
air
```

## Code Quality Tools

### Go Linting

```bash
# Run linter on specific service (requires golangci-lint)
cd services/fm-service
make lint
```

Services **without** individual Makefiles (HR, SCM, CRM, PM, Auth) have no lint command — run `golangci-lint run` directly.

### Go Formatting

```bash
gofmt -w .
go mod tidy
```

## Debugging

### Go Debugging with Delve

```bash
# Install Delve debugger
go install github.com/go-delve/delve/cmd/dlv@latest

# Start service with debugger
cd services/fm-service
dlv debug cmd/server/main.go
```

## Next Steps

- [Development Workflow](development-workflow.md) — Daily development practices
- [Testing and Verification](testing-verification.md) — How to test the system
- [System Architecture](../architecture/README.md) — Understand the codebase structure
