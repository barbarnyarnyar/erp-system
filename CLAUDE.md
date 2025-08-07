# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Architecture Overview

This is a microservices-based ERP system built with Go, featuring:

- **API Gateway** (port 8080) - Request routing, authentication, rate limiting
- **6 Core Services**: Financial Management (FM), Human Resources (HR), Supply Chain Management (SCM), Manufacturing (M), Customer Relationship Management (CRM), Project Management (PM)
- **Event-driven architecture** using Kafka for asynchronous communication
- **Docker containerization** for all services
- **PostgreSQL** for data persistence, **Redis** for caching

### Service Structure
Each service follows Clean Architecture patterns:
```
services/{service-name}/
├── cmd/
│   ├── main.go              # Application entry point (some services)
│   └── server/main.go       # Alternative entry point (some services)
├── internal/
│   ├── api/
│   │   ├── handlers/        # HTTP request handlers
│   │   └── routes/          # Route definitions
│   ├── business/domain/     # Business logic and domain models
│   ├── config/              # Service configuration
│   └── data/migrations/     # Database migrations (when applicable)
├── common-utils/            # Shared utilities (symlinked from shared/)
├── go.mod
├── go.sum
├── Makefile                 # Service-specific build commands
└── Dockerfile
```

**Note**: Entry points vary by service:
- `fm-service`: Uses `cmd/server/main.go`
- Other services: Use `cmd/main.go`

### Shared Components
- `shared/` directory contains common utilities, templates, and shared Go modules
- `api-gateway/` handles routing to all microservices
- `infrastructure/` contains database, message queue, and cache configurations

## Development Commands

### Core Development Tasks
```bash
# Start all services (recommended for development)
make run                     # Uses docker-compose up -d

# Build all services
make build                   # Uses docker-compose build

# Stop all services  
make stop                    # Uses docker-compose down

# Check service health
make health                  # Calls health endpoints for all services

# View logs
make logs                    # Shows logs for all services

# Clean up containers and images
make clean                   # Full cleanup with volumes
```

### Testing Commands
```bash
# Test all Hello World APIs through gateway
make test

# Test services directly (bypass gateway)
make test-direct
```

### Individual Service Development
For working on specific services locally, each service has its own Makefile with these commands:
```bash
# Navigate to service directory (example: fm-service)
cd services/fm-service

# Run tests
make test

# Run tests with coverage
make test-coverage

# Run linting (requires golangci-lint)
make lint

# Build service locally
make build

# Run service locally (requires infrastructure)
make run

# Development with hot reload (requires air)
make dev

# Database migrations (when applicable)
make migrate-up
make migrate-down
make migrate-create name=migration_name
```

### Manual Service Management
```bash
# Start infrastructure only (for local Go development)
docker-compose up -d postgres kafka redis

# Run individual services locally
cd services/fm-service && go run cmd/server/main.go
cd services/hr-service && go run cmd/main.go
# etc.
```

### Building from Source
```bash
# Build all services manually
./scripts/build.sh          # Builds Go binaries for all services

# Individual service builds
cd services/fm-service
go mod tidy
go build -o bin/main cmd/server/main.go
```

## Service Ports and URLs

### Production Ports (via API Gateway - port 8080)
- Financial Management: `/api/v1/fm/*` → fm-service:8001
- Human Resources: `/api/v1/hr/*` → hr-service:8002  
- Supply Chain: `/api/v1/scm/*` → scm-service:8003
- Manufacturing: `/api/v1/m/*` → m-service:8004
- CRM: `/api/v1/crm/*` → crm-service:8005
- Project Management: `/api/v1/pm/*` → pm-service:8006

**Note**: The Makefile test routes use different patterns:
- Finance: `/api/v1/finance/hello`
- Manufacturing: `/api/v1/manufacturing/hello`  
- Projects: `/api/v1/projects/hello`

### Direct Service Access (Development)
- fm-service: http://localhost:8001
- hr-service: http://localhost:8002
- scm-service: http://localhost:8003
- m-service: http://localhost:8004
- crm-service: http://localhost:8005
- pm-service: http://localhost:8006

## Configuration

Services use environment variables for configuration:
- `SERVER_PORT` - Service port (defaults per service)
- `DB_HOST`, `DB_PORT`, `DB_USERNAME`, `DB_PASSWORD`, `DB_DATABASE` - PostgreSQL settings
- `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD` - Redis settings  
- `KAFKA_BROKERS` - Message queue connection string
- `ENV` - Environment (development/production)

## Key Technologies

- **Language**: Go 1.21+
- **Web Framework**: Gin (github.com/gin-gonic/gin)
- **Database**: PostgreSQL 
- **Message Queue**: Kafka
- **Cache**: Redis
- **Containerization**: Docker & Docker Compose
- **Architecture**: Clean Architecture with domain-driven design

## Development Notes

- The `un-nessary/` directory contains deprecated/unused code - ignore when working on active development
- Each service has its own `go.mod` file and can be developed independently
- Services communicate via HTTP APIs and asynchronous events through Kafka
- All services expose `/health` endpoints for monitoring
- The project uses conventional Git commits and maintains comprehensive API documentation
- Service entry points are inconsistent: `fm-service` uses `cmd/server/main.go`, others use `cmd/main.go`
- Each service may have its own database and migration files
- `common-utils` directories in services are symlinked to shared utilities
- Use `golangci-lint` for Go code linting (install separately)
- Use `air` for development hot reloading (install separately)

## Prerequisites for Development

- **Go** 1.21+
- **Docker** and **Docker Compose**
- **golangci-lint** (for linting): `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
- **air** (for hot reloading): `go install github.com/cosmtrek/air@latest`
- **migrate** (for database migrations): `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`