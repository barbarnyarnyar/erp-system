# Project Setup and Structure Guide

## Table of Contents
1. [Technology Stack](#technology-stack)
2. [Microservices Architecture](#microservices-architecture)
3. [Port Allocation Guide](#port-allocation-guide)
4. [Project Structure](#project-structure)
5. [Development Setup](#development-setup)
6. [Service Dependencies](#service-dependencies)
7. [Environment Configuration](#environment-configuration)

---

## Technology Stack

### Core Technologies
- **Language**: Go 1.21+
- **Web Framework**: Gin (github.com/gin-gonic/gin)
- **Database**: PostgreSQL
- **Message Queue**: RabbitMQ (primary), Kafka (alternative)
- **Containerization**: Docker & Docker Compose
- **Cache**: Redis

### Architecture Pattern
- **Microservices Architecture**: Domain-driven service separation
- **Clean Architecture**: Layered approach within each service
- **Event-Driven Communication**: Asynchronous inter-service messaging
- **API Gateway Pattern**: Centralized routing and cross-cutting concerns

---

## Microservices Architecture

### Service Overview
```
ERP System Microservices:
├── API Gateway (Port 8080) - Request routing, authentication, rate limiting
├── Financial Management Service (Port 8001)
├── Human Resources Service (Port 8002)
├── Supply Chain Management Service (Port 8003)
├── Manufacturing Service (Port 8004)
├── Customer Relationship Management Service (Port 8005)
└── Project Management Service (Port 8006)
```

### Standard Service Structure
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

---

## Port Allocation Guide

### 🚀 Application Services (Docker External Ports)

| Service | External Port | Internal Port | URL | Description |
|---------|---------------|---------------|-----|-------------|
| **API Gateway** | `8080` | `8080` | `http://localhost:8080` | Main entry point for all client requests |
| **Financial Management** | `8001` | `8080` | `http://localhost:8001` | Direct access to financial service |
| **HR/HCM** | `8002` | `8080` | `http://localhost:8002` | Direct access to HR service |
| **Supply Chain Management** | `8003` | `8080` | `http://localhost:8003` | Direct access to SCM service |
| **Manufacturing** | `8004` | `8080` | `http://localhost:8004` | Direct access to manufacturing service |
| **Sales & CRM** | `8005` | `8080` | `http://localhost:8005` | Direct access to sales service |
| **Project Management** | `8006` | `8080` | `http://localhost:8006` | Direct access to project service |

### 🗄️ Infrastructure Services

| Service | External Port | Internal Port | URL | Description |
|---------|---------------|---------------|-----|-------------|
| **PostgreSQL** | `5432` | `5432` | `postgresql://localhost:5432/erp_db` | Database server |
| **Redis** | `6379` | `6379` | `redis://localhost:6379` | Caching and session storage |
| **RabbitMQ** | `5672` | `5672` | `amqp://localhost:5672` | Message queue (AMQP) |
| **RabbitMQ Management** | `15672` | `15672` | `http://localhost:15672` | RabbitMQ web interface |

### Port Range Strategy
- **8000-8099**: Application microservices
- **5000-5999**: Database and persistent storage
- **6000-6999**: Caching and temporary storage
- **15000-15999**: Management interfaces

---

## Project Structure

### Repository Structure
```
erp-system/
├── api-gateway/              # API Gateway service
├── services/                 # Microservices
│   ├── fm-service/          # Financial Management
│   ├── hr-service/          # Human Resources
│   ├── scm-service/         # Supply Chain Management
│   ├── m-service/           # Manufacturing
│   ├── crm-service/         # Customer Relationship Management
│   └── pm-service/          # Project Management
├── shared/                   # Shared utilities and templates
├── infrastructure/          # Docker, database, message queue configs
├── scripts/                 # Build and deployment scripts
├── documentation/           # Project documentation
├── docker-compose.yml       # Multi-service container orchestration
├── Makefile                # Development commands
└── README.md               # Project overview
```

### Individual Service Structure (Go/Gin Template)
```
service-name/
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/        # HTTP request handlers
│   │   │   ├── health.go    # Health check endpoint
│   │   │   └── service.go   # Service-specific handlers
│   │   ├── middleware/      # HTTP middleware
│   │   └── routes/          # Route definitions
│   ├── business/
│   │   ├── domain/          # Domain models and entities
│   │   ├── repository/      # Data access layer
│   │   └── service/         # Business logic layer
│   ├── config/              # Configuration management
│   └── database/            # Database connection and migrations
├── pkg/                     # Reusable packages (if any)
├── go.mod                   # Go module definition
├── go.sum                   # Go dependencies
├── Dockerfile               # Container build instructions
└── README.md               # Service-specific documentation
```

### Example Service Implementation (Go/Gin)

#### `cmd/server/main.go`
```go
package main

import (
    "log"
    "github.com/gin-gonic/gin"
    "fm-service/internal/api/handlers"
    "fm-service/internal/api/routes"
    "fm-service/internal/config"
)

func main() {
    cfg := config.Load()
    
    r := gin.Default()
    
    // Setup routes
    routes.SetupRoutes(r)
    
    log.Printf("Starting FM Service on port %s", cfg.Port)
    r.Run(":" + cfg.Port)
}
```

#### `internal/api/handlers/health.go`
```go
package handlers

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status": "healthy",
        "service": "fm-service",
        "version": "1.0.0",
    })
}
```

#### `internal/api/routes/routes.go`
```go
package routes

import (
    "github.com/gin-gonic/gin"
    "fm-service/internal/api/handlers"
)

func SetupRoutes(r *gin.Engine) {
    // Health check
    r.GET("/health", handlers.HealthCheck)
    
    // API v1 routes
    v1 := r.Group("/api/v1")
    {
        // Service-specific routes
        v1.GET("/accounts", handlers.GetAccounts)
        v1.POST("/accounts", handlers.CreateAccount)
        // ... more routes
    }
}
```

---

## Development Setup

### Prerequisites
```bash
# Install Go 1.21+
go version

# Install Docker and Docker Compose
docker --version
docker-compose --version

# Install development tools
go install github.com/air-verse/air@latest  # Hot reload
```

### Quick Start Commands
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

### Manual Development Setup
```bash
# Start infrastructure only (for local Go development)
docker-compose up -d postgres rabbitmq redis

# Run individual services locally
cd services/fm-service && go run cmd/server/main.go
cd services/hr-service && go run cmd/main.go
# etc.
```

### Testing Commands
```bash
# Test all Hello World APIs through gateway
make test

# Test services directly (bypass gateway)
make test-direct
```

---

## Service Dependencies

```mermaid
graph TD
    Client[Client Applications] --> Gateway[API Gateway :8080]

    Gateway --> Financial[Financial Service :8001]
    Gateway --> HR[HR Service :8002]
    Gateway --> SCM[SCM Service :8003]
    Gateway --> Manufacturing[Manufacturing Service :8004]
    Gateway --> Sales[Sales Service :8005]
    Gateway --> Project[Project Service :8006]

    Financial --> DB[(PostgreSQL :5432)]
    HR --> DB
    SCM --> DB
    Manufacturing --> DB
    Sales --> DB
    Project --> DB

    Financial --> Queue[RabbitMQ :5672]
    HR --> Queue
    SCM --> Queue
    Manufacturing --> Queue
    Sales --> Queue
    Project --> Queue

    Gateway --> Cache[Redis :6379]
```

### Service Communication Patterns
1. **Synchronous**: Direct HTTP API calls via API Gateway
2. **Asynchronous**: Message queue events for decoupled operations
3. **Data Sharing**: Shared PostgreSQL database with separate schemas

---

## Environment Configuration

### Common Environment Variables
```bash
# Service-specific
PORT=8080                    # Internal port (same for all services)
DB_SCHEMA={service_schema}   # financial, hr, supply_chain, etc.
ENV=development              # development/production

# Shared infrastructure
DB_HOST=postgres
DB_PORT=5432
DB_USERNAME=erp_user
DB_PASSWORD=erp_password
DB_DATABASE=erp_db

RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/
REDIS_HOST=redis            # Only for API Gateway
REDIS_PORT=6379             # Only for API Gateway
```

### Docker Compose Configuration
```yaml
version: '3.8'
services:
  api-gateway:
    build: ./api-gateway
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - REDIS_HOST=redis
    depends_on:
      - redis

  fm-service:
    build: ./services/fm-service
    ports:
      - "8001:8080"
    environment:
      - PORT=8080
      - DB_SCHEMA=financial
    depends_on:
      - postgres
      - rabbitmq
```

---

## Security and Best Practices

### 🔒 Security Recommendations
1. **Production**: Only expose API Gateway (port 8080) externally
2. **Internal Services**: Ports 8001-8006 should be internal-only in production
3. **Infrastructure**: Database and queue ports should never be exposed externally
4. **Management Interfaces**: RabbitMQ management should be secured or internal-only

### 🧪 Health Check Endpoints
Once services are running, test connectivity:
```bash
# API Gateway health check
curl http://localhost:8080/health

# Direct service health checks
curl http://localhost:8001/health  # Financial
curl http://localhost:8002/health  # HR
curl http://localhost:8003/health  # SCM
curl http://localhost:8004/health  # Manufacturing
curl http://localhost:8005/health  # Sales
curl http://localhost:8006/health  # Project
```

### Development Workflow
1. **Infrastructure First**: Start databases and message queues
2. **Service Development**: Develop and test individual services
3. **Integration Testing**: Test service interactions via API Gateway
4. **End-to-End Testing**: Full workflow testing across all services

This structure provides a scalable, maintainable foundation for the ERP microservices architecture while following Go and containerization best practices.