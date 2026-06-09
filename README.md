# ERP Microservices System

A comprehensive Enterprise Resource Planning (ERP) system built with Go microservices architecture, featuring event-driven communication through message queues and fully containerized with Docker.

## 📋 Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Services](#services)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [API Documentation](#api-documentation)
- [Development](#development)
- [Testing](#testing)
- [Deployment](#deployment)
- [Monitoring](#monitoring)
- [Contributing](#contributing)
- [License](#license)

## 🎯 Overview

This ERP system provides a complete business management solution with the following capabilities:

- **Financial Management** - Accounting, budgeting, and financial reporting
- **Human Resources** - Employee management, payroll, and performance tracking
- **Supply Chain Management** - Inventory, procurement, and order management
- **Manufacturing** - Production planning, quality control, and shop floor management
- **Customer Relationship Management** - Sales pipeline, customer service, and marketing
- **Project Management** - Planning, resource allocation, and tracking

## 🏗️ Architecture

### Microservices Design

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   API Gateway   │    │   Frontend      │    │   Admin Panel   │
│   (Port 8080)   │    │   (Port 3000)   │    │   (Port 3001)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
    ┌────────────────────────────┼────────────────────────────┐
    │                            │                            │
    │                    Message Queue                        │
    │                      (Kafka)                           │
    │                                                        │
    └────────────────────────────┼────────────────────────────┘
                                 │
         ┌───────────────────────┼───────────────────────┐
         │                       │                       │
┌────────▼────────┐    ┌────────▼────────┐    ┌────────▼────────┐
│  FM Service     │    │  HR Service     │    │  SCM Service    │
│  (Port 8081)    │    │  (Port 8082)    │    │  (Port 8083)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
         ┌───────────────────────┼───────────────────────┐
         │                       │                       │
┌────────▼────────┐    ┌────────▼────────┐    ┌────────▼────────┐
│  M Service      │    │  CRM Service    │    │  PM Service     │
│  (Port 8084)    │    │  (Port 8085)    │    │  (Port 8086)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
    ┌────────────────────────────▼────────────────────────────┐
    │                   Shared Infrastructure                 │
    │  PostgreSQL │ Redis │ Prometheus │ Grafana │ Jaeger    │
    └─────────────────────────────────────────────────────────┘
```

### Communication Patterns

- **Synchronous**: REST API calls for real-time queries
- **Asynchronous**: Event-driven messaging for workflow automation
- **Event Sourcing**: State changes stored as events
- **CQRS**: Separate read/write operations

## 🚀 Services

| Service | Port | Description | Database |
|---------|------|-------------|----------|
| **API Gateway** | 8080 | Request routing, authentication, rate limiting | - |
| **FM Service** | 8081 | Financial management, accounting, budgeting | fm_db |
| **HR Service** | 8082 | Human resources, payroll, employee management | hr_db |
| **SCM Service** | 8083 | Supply chain, inventory, procurement | scm_db |
| **M Service** | 8084 | Manufacturing, production, quality control | m_db |
| **CRM Service** | 8085 | Customer relations, sales, marketing | crm_db |
| **PM Service** | 8086 | Project management, resource allocation | pm_db |

## 📋 Prerequisites

- **Docker** >= 20.10.0
- **Docker Compose** >= 2.0.0
- **Go** >= 1.21 (for development)
- **Make** (optional, for convenience commands)

## 🚀 Quick Start

1. **Clone the repository**
   ```bash
   git clone https://github.com/your-org/erp-microservices.git
   cd erp-microservices
   ```

2. **⚠️ Setup secure credentials**
   ```bash
   # Generate strong credentials for development/production
   ./scripts/setup-secrets.sh --auto
   
   # Or use interactive setup
   ./scripts/setup-secrets.sh
   ```

3. **Start all services**
   ```bash
   docker-compose up -d
   ```

4. **Verify services are running**
   ```bash
   docker-compose ps
   ```

5. **Access the application**
   - API Gateway: http://localhost:8080
   - Check health: `make health`

6. **Initialize sample data**
   ```bash
   make seed-data
   ```

## ⚙️ Configuration

### ⚠️ SECURITY: Environment Variables

**IMPORTANT**: Generate strong credentials before running services!

```bash
# Auto-generate credentials
./scripts/setup-secrets.sh --auto
```

This will create a `.env` file with secure values for:

Create a `.env` file in the project root:

```env
# Database
POSTGRES_USER=admin
POSTGRES_PASSWORD=admin
POSTGRES_DB=erp_db

# Message Queue
KAFKA_BROKERS=localhost:9092

# Redis
REDIS_PASSWORD=admin

# JWT Secret
JWT_SECRET=your-super-secret-key

# Environment
ENVIRONMENT=development
LOG_LEVEL=info
```

### Service Configuration

Each service can be configured via environment variables or config files:

```yaml
# services/fm-service/config.yaml
database:
  host: postgres
  port: 5432
  name: fm_db
  user: admin
  password: admin

kafka:
  brokers:
    - kafka:9092
  
redis:
  host: redis
  port: 6379
  password: admin

server:
  port: 8081
  timeout: 30s
```

## 📚 API Documentation

### Authentication

All API requests require authentication via JWT tokens:

```bash
# Get access token
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin"}'

# Use token in requests
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/fm/accounts
```

### Service Endpoints

| Service | Endpoint | Description |
|---------|----------|-------------|
| **FM** | `/api/v1/fm/*` | Financial operations |
| **HR** | `/api/v1/hr/*` | Human resources |
| **SCM** | `/api/v1/scm/*` | Supply chain |
| **M** | `/api/v1/m/*` | Manufacturing |
| **CRM** | `/api/v1/crm/*` | Customer relations |
| **PM** | `/api/v1/pm/*` | Project management |

Full API documentation is available at: http://localhost:8080/docs

## 🛠️ Development

### Local Development Setup

1. **Install dependencies**
   ```bash
   go mod tidy
   ```

2. **Start infrastructure only**
   ```bash
   docker-compose up -d postgres kafka redis
   ```

3. **Run services locally**
   ```bash
   # Terminal 1 - FM Service
   cd services/fm-service
   go run main.go

   # Terminal 2 - HR Service
   cd services/hr-service
   go run main.go

   # Continue for other services...
   ```

### Project Structure

```
erp-system/
├── docker-compose.yml
├── Makefile
├── README.md
├── .env.example
├── api-gateway/
│   ├── Dockerfile
│   ├── main.go
│   └── internal/
├── services/
│   ├── fm-service/
│   │   ├── Dockerfile
│   │   ├── main.go
│   │   ├── internal/
│   │   │   ├── handlers/
│   │   │   ├── models/
│   │   │   ├── repository/
│   │   │   └── services/
│   │   └── migrations/
│   ├── hr-service/
│   └── ... (other services)
├── shared/
│   ├── events/
│   ├── middleware/
│   ├── models/
│   └── utils/
└── scripts/
    ├── seed-data.sql
    └── setup.sh
```

### Code Standards

- **Go**: Follow official Go style guidelines
- **Git**: Use conventional commits
- **Testing**: Minimum 80% test coverage
- **Documentation**: Update README and API docs

## 🧪 Testing

### Unit Tests

```bash
# Test all services
make test

# Test specific service
cd services/fm-service
go test ./...

# Test with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Integration Tests

```bash
# Start test environment
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
make test-integration

# Cleanup
docker-compose -f docker-compose.test.yml down
```

### Load Testing

```bash
# Install k6
brew install k6

# Run load tests
k6 run tests/load/api-test.js
```

## 🚀 Deployment

### Production Deployment

1. **Build production images**
   ```bash
   make build-prod
   ```

2. **Deploy with Docker Swarm**
   ```bash
   docker stack deploy -c docker-stack.yml erp-stack
   ```

3. **Deploy with Kubernetes**
   ```bash
   kubectl apply -f k8s/
   ```

### Health Checks

All services expose health check endpoints:

```bash
curl http://localhost:8080/health
curl http://localhost:8081/health
```

## 📊 Monitoring

### Metrics and Monitoring

- **Prometheus**: Metrics collection at http://localhost:9090
- **Grafana**: Dashboards at http://localhost:3000
- **Jaeger**: Distributed tracing at http://localhost:16686

### Logging

Centralized logging with ELK stack:

```bash
# View logs
docker-compose logs -f fm-service

# Search logs in Kibana
# http://localhost:5601
```

### Alerts

Configure alerts in `monitoring/alerts.yml`:

```yaml
groups:
  - name: erp-alerts
    rules:
      - alert: ServiceDown
        expr: up == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Service {{ $labels.instance }} is down"
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/new-feature`
3. Commit changes: `git commit -m 'Add new feature'`
4. Push to branch: `git push origin feature/new-feature`
5. Submit a pull request

### Development Workflow

```bash
# Setup pre-commit hooks
make setup-hooks

# Run linting
make lint

# Run tests
make test

# Build and test locally
make build-local
```

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Links

- [Documentation](https://docs.erp-system.com)
- [Issue Tracker](https://github.com/your-org/erp-microservices/issues)
- [Slack Channel](https://your-org.slack.com/channels/erp-dev)

## 📞 Support

For support, email support@erp-system.com or join our Slack channel.

---

**Made with ❤️ by the ERP Development Team**
