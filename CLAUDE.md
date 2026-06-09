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
│   └── main.go              # Application entry point
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

**Note**: All services now use the standardized `cmd/main.go` entry point.

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
cd services/fm-service && go run cmd/main.go
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
go build -o bin/main cmd/main.go
```

## Service Ports and URLs

### Production Ports (via API Gateway - port 8080)
- Financial Management: `/api/v1/fm/*` → fm-service:8001
- Human Resources: `/api/v1/hr/*` → hr-service:8003  
- Supply Chain: `/api/v1/scm/*` → scm-service:8006
- Manufacturing: `/api/v1/m/*` → m-service:8004
- CRM: `/api/v1/crm/*` → crm-service:8002
- Project Management: `/api/v1/pm/*` → pm-service:8005

**Note**: The Makefile test routes use different patterns:
- Finance: `/api/v1/finance/hello`
- Manufacturing: `/api/v1/manufacturing/hello`  
- Projects: `/api/v1/projects/hello`

### Direct Service Access (Development)
- fm-service: http://localhost:8001
- hr-service: http://localhost:8003
- scm-service: http://localhost:8006
- m-service: http://localhost:8004
- crm-service: http://localhost:8002
- pm-service: http://localhost:8005

## Configuration

Services use environment variables for configuration:
- `SERVER_PORT` - Service port (defaults per service)
- `DB_HOST`, `DB_PORT`, `DB_USERNAME`, `DB_PASSWORD`, `DB_DATABASE` - PostgreSQL settings
- `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD` - Redis settings  
- `KAFKA_BROKERS` - Message queue connection string
- `JWT_SECRET` - JWT signing secret (MUST be changed in production)
- `ENV` - Environment (development/production)

### ⚠️ SECURITY: Credentials & Secrets Management

**CRITICAL**: Default credentials must be changed before production deployment!

#### Setup Secure Credentials

```bash
# Auto-generate strong credentials
./scripts/setup-secrets.sh --auto

# Or interactive setup
./scripts/setup-secrets.sh --interactive
```

This creates a `.env` file with:
- Strong PostgreSQL credentials (32 char random)
- Strong Redis password (32 char random)
- Secure JWT secret (256-bit)
- Secure admin credentials

#### Environment Variables Required

```env
# These MUST be set (no defaults)
POSTGRES_USER=<strong-username>
POSTGRES_PASSWORD=<strong-password>
REDIS_PASSWORD=<strong-password>
JWT_SECRET=<256-bit-hex-string>

# Optional (have defaults)
POSTGRES_DB=erp_db
KAFKA_BROKERS=kafka:9092
ENVIRONMENT=development
```

#### Production Checklist

Before deploying to production:

- [ ] Generate strong credentials via `./scripts/setup-secrets.sh`
- [ ] Store `.env` in a secrets manager (Vault, AWS Secrets Manager, HashiCorp, etc.)
- [ ] Never commit `.env` to version control (it's in `.gitignore`)
- [ ] Change admin password immediately after first login
- [ ] Implement TLS/HTTPS for all services
- [ ] Set up proper authentication (OAuth2, SAML, etc.)
- [ ] Enable rate limiting and DDoS protection
- [ ] Configure audit logging
- [ ] Set up secrets rotation policy
- [ ] Run security scanning tools (gosec, snyk, etc.)

#### JWT Secret Generation

```bash
# Generate a secure JWT secret
openssl rand -hex 32

# Or use this in a script
JWT_SECRET=$(openssl rand -hex 32)
```

#### Password Generation

```bash
# Generate strong passwords
openssl rand -base64 32  # For PostgreSQL & Redis
```

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
- All services use standardized entry point: `cmd/main.go`
- Each service may have its own database and migration files
- `common-utils` directories in services are symlinked to shared utilities
- Use `golangci-lint` for Go code linting (install separately)
- Use `air` for development hot reloading (install separately)
- **SECURITY**: All default credentials must be externalized via `.env` - see Configuration section

## Prerequisites for Development

- **Go** 1.21+
- **Docker** and **Docker Compose**
- **golangci-lint** (for linting): `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
- **air** (for hot reloading): `go install github.com/cosmtrek/air@latest`
- **migrate** (for database migrations): `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`