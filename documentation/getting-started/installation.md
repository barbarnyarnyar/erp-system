# Installation

Get the ERP system running in 15 minutes with these step-by-step instructions.

## Quick Installation (5 minutes)

The fastest way to get started:

**Step 1: Clone the Repository**
```bash
git clone https://github.com/your-org/erp-system.git
cd erp-system
```

**Step 2: Start All Services**
```bash
# Start complete system with one command
make run
```

This command will:
- Pull and build all Docker images
- Start PostgreSQL database with multiple schemas
- Start Redis cache server
- Start Kafka message broker
- Launch all 6 microservices (Financial, HR, SCM, CRM, Manufacturing, Project)
- Configure and start the API Gateway

**Step 3: Verify Installation**
```bash
# Check all services are running
make health
```

Expected output:
```
✅ API Gateway (port 8080): OK
✅ Financial Service (port 8001): OK
✅ HR Service (port 8002): OK
✅ SCM Service (port 8003): OK
✅ CRM Service (port 8004): OK
✅ Manufacturing Service (port 8005): OK
✅ Project Service (port 8006): OK
✅ PostgreSQL (port 5432): OK
✅ Redis (port 6379): OK
✅ Kafka (port 9092): OK
```

**Step 4: Test API Endpoints**
```bash
# Test all services through API Gateway
make test
```

Expected output:
```
Testing Financial Service...     ✅ OK
Testing HR Service...            ✅ OK
Testing SCM Service...           ✅ OK
Testing CRM Service...           ✅ OK
Testing Manufacturing Service... ✅ OK
Testing Project Service...       ✅ OK
```

## Detailed Installation Steps

If the quick installation fails, follow these detailed steps:

### Step 1: Environment Variables Setup
```bash
# Copy environment template
cp .env.example .env

# Edit configuration (optional for development)
nano .env
```

Basic `.env` configuration:
```bash
# Environment
ENVIRONMENT=development
LOG_LEVEL=info

# Database Configuration
DB_HOST=postgres
DB_PORT=5432
DB_USERNAME=postgres
DB_PASSWORD=postgres

# Redis Configuration
REDIS_HOST=redis
REDIS_PORT=6379

# Kafka Configuration
KAFKA_BROKERS=kafka:9092

# Service Ports (defaults work for development)
API_GATEWAY_PORT=8080
FM_SERVICE_PORT=8001
HR_SERVICE_PORT=8002
SCM_SERVICE_PORT=8003
CRM_SERVICE_PORT=8004
MFG_SERVICE_PORT=8005
PM_SERVICE_PORT=8006
```

### Step 2: Infrastructure Services First
```bash
# Start infrastructure services first
docker-compose up -d postgres redis kafka

# Wait for services to be ready
sleep 30

# Verify infrastructure is running
docker-compose ps
```

### Step 3: Database Initialization
```bash
# Run database migrations for each service
cd services/fm-service && make migrate-up && cd ../..
cd services/hr-service && make migrate-up && cd ../..
cd services/scm-service && make migrate-up && cd ../..
cd services/crm-service && make migrate-up && cd ../..
```

### Step 4: Start Application Services
```bash
# Start all application services
docker-compose up -d fm-service hr-service scm-service crm-service mfg-service pm-service

# Start API Gateway last
docker-compose up -d api-gateway
```

## Verification Tests

### Manual API Tests
```bash
# Test API Gateway routing
curl http://localhost:8080/api/v1/finance/hello
curl http://localhost:8080/api/v1/hr/hello
curl http://localhost:8080/api/v1/scm/hello

# Test with JSON data
curl -X POST http://localhost:8080/api/v1/finance/accounts \
  -H "Content-Type: application/json" \
  -d '{
    "account_code": "1000",
    "account_name": "Cash - Operating",
    "account_type": "ASSET"
  }'
```

### Health Checks
```bash
# System health check
make health

# Manual health checks
curl http://localhost:8080/health
curl http://localhost:8001/health  # Financial service
curl http://localhost:8002/health  # HR service
```

## Next Steps

With the system successfully installed:

1. **Configure your environment**: [Configuration Guide](configuration.md)
2. **Set up development tools**: [Development Environment](development-environment.md)  
3. **Learn the architecture**: [System Overview](../architecture/README.md)
4. **Explore the features**: [Business Modules](../modules/README.md)

## Troubleshooting

If you encounter issues during installation, check the [Common Setup Issues](common-issues.md) guide for solutions.