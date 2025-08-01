# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Architecture Overview

This is a microservices-based ERP system built with Go, featuring:

- **API Gateway** (port 8080) - Request routing, authentication, rate limiting
- **6 Core Services**: Financial Management (FM), Human Resources (HR), Supply Chain Management (SCM), Manufacturing (M), Customer Relationship Management (CRM), Project Management (PM)
- **Event-driven architecture** using RabbitMQ for asynchronous communication
- **Docker containerization** for all services
- **PostgreSQL** for data persistence, **Redis** for caching

### Service Structure
Each service follows Clean Architecture patterns:
```
services/{service-name}/
├── cmd/
│   ├── main.go              # Application entry point
│   └── server/main.go       # Alternative entry point
├── internal/
│   ├── api/
│   │   ├── handlers/        # HTTP request handlers
│   │   └── routes/          # Route definitions
│   ├── business/domain/     # Business logic and domain models
│   └── config/              # Service configuration
├── go.mod
├── go.sum
└── Dockerfile
```

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

### Manual Service Management
```bash
# Start infrastructure only (for local Go development)
docker-compose up -d postgres rabbitmq redis

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
- `RABBITMQ_URL` - Message queue connection string
- `ENV` - Environment (development/production)

## Key Technologies

- **Language**: Go 1.21+
- **Web Framework**: Gin (github.com/gin-gonic/gin)
- **Database**: PostgreSQL 
- **Message Queue**: RabbitMQ
- **Cache**: Redis
- **Containerization**: Docker & Docker Compose
- **Architecture**: Clean Architecture with domain-driven design

## Development Notes

- The `un-nessary/` directory contains deprecated/unused code - ignore when working on active development
- Each service has its own `go.mod` file and can be developed independently
- Services communicate via HTTP APIs and asynchronous events through RabbitMQ
- All services expose `/health` endpoints for monitoring
- The project uses conventional Git commits and maintains comprehensive API documentation