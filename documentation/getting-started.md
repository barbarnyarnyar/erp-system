# Getting Started

This guide will get you up and running with the ERP system in 15 minutes. You'll have a complete development environment with all services running locally.

## Prerequisites

Before starting, ensure you have these tools installed:

### Required Software
- **Git** (2.30+): Version control
- **Docker** (20.10+): Container runtime
- **Docker Compose** (2.0+): Multi-container orchestration
- **Go** (1.21+): Backend development
- **Node.js** (18.0+): Frontend development
- **Make**: Build automation

### Verify Your Setup

Run these commands to verify your installation:

```bash
# Check versions
git --version
docker --version
docker-compose --version
go version
node --version
npm --version
make --version
```

Expected output format:
```
git version 2.39.0
Docker version 24.0.5
Docker Compose version v2.20.2
go version go1.21.3 linux/amd64
v18.17.0
9.6.7
GNU Make 4.3
```

## Quick Setup (5 minutes)

### 1. Clone the Repository

```bash
git clone https://github.com/your-org/erp-system.git
cd erp-system
```

### 2. Start All Services

The system uses Docker Compose to run all services with a single command:

```bash
# Start all services in detached mode
make run
```

This command:
- Builds all Docker images
- Starts PostgreSQL database
- Starts Redis cache
- Starts Kafka message broker
- Launches all 6 microservices
- Sets up the API Gateway

### 3. Verify Services Are Running

Check that all services are healthy:

```bash
# Check service status
make health
```

You should see output like:
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

### 4. Test the System

Test the API endpoints:

```bash
# Test all Hello World endpoints
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

## Access the System

With all services running, you can access:

### API Gateway
- **URL**: http://localhost:8080
- **Health Check**: http://localhost:8080/health
- **API Docs**: http://localhost:8080/swagger

### Individual Services (for debugging)
- **Financial**: http://localhost:8001
- **HR**: http://localhost:8002
- **SCM**: http://localhost:8003
- **CRM**: http://localhost:8004
- **Manufacturing**: http://localhost:8005
- **Project Management**: http://localhost:8006

### Database Access
- **PostgreSQL**: localhost:5432 (multiple databases)
- **Redis**: localhost:6379
- **Kafka UI**: http://localhost:9021 (if enabled)

## Your First API Call

Test the system with a simple API call:

```bash
# Get financial service status
curl http://localhost:8080/api/v1/finance/hello

# Expected response:
{
  "message": "Financial Management Service is running",
  "version": "v1.0.0",
  "timestamp": "2024-03-15T10:30:00Z"
}
```

## Development Workflow

Now that everything is running, here's your typical development workflow:

### Making Changes to a Service

1. **Edit Code**: Modify files in `services/{service-name}/`
2. **Rebuild Service**: 
   ```bash
   # Rebuild specific service
   docker-compose build fm-service
   
   # Restart just that service
   docker-compose restart fm-service
   ```
3. **Test Changes**:
   ```bash
   # Test specific service
   curl http://localhost:8001/hello
   ```

### Viewing Logs

```bash
# View all service logs
make logs

# View specific service logs
docker-compose logs -f fm-service

# View last 50 lines
docker-compose logs --tail=50 fm-service
```

### Database Operations

```bash
# Access PostgreSQL (Financial DB)
docker-compose exec postgres psql -U postgres -d financial_db

# Access Redis
docker-compose exec redis redis-cli

# View database migrations
cd services/fm-service
ls internal/data/migrations/
```

## Common Development Tasks

### Running Tests

```bash
# Run all tests
make test

# Run tests for specific service
cd services/fm-service
make test

# Run tests with coverage
cd services/fm-service
make test-coverage
```

### Code Linting

```bash
# Lint specific service
cd services/fm-service
make lint

# Lint all Go code (requires golangci-lint)
find . -name "*.go" -path "./services/*" | xargs golangci-lint run
```

### Database Migrations

```bash
# Create new migration
cd services/fm-service
make migrate-create name=add_customer_table

# Run migrations
make migrate-up

# Rollback last migration
make migrate-down
```

## Environment Configuration

The system uses environment variables for configuration. Key variables:

```bash
# Database configuration
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=postgres
DB_PASSWORD=postgres

# Redis configuration  
REDIS_HOST=localhost
REDIS_PORT=6379

# Kafka configuration
KAFKA_BROKERS=localhost:9092

# Service ports
FM_SERVICE_PORT=8001
HR_SERVICE_PORT=8002
# ... etc
```

Configuration files are located in:
- `docker-compose.yml` - Service definitions and environment variables
- `services/{service}/config/config.go` - Service-specific configuration
- `.env.example` - Template environment variables

## Stopping the System

When you're done developing:

```bash
# Stop all services
make stop

# Stop and remove containers, networks, volumes
make clean
```

## Troubleshooting Quick Fixes

### Port Already in Use
```bash
# Find what's using the port
lsof -i :8080

# Kill the process
kill -9 <PID>
```

### Services Won't Start
```bash
# Check Docker daemon is running
docker info

# Restart Docker service
sudo systemctl restart docker
```

### Database Connection Issues
```bash
# Check PostgreSQL is running
docker-compose ps postgres

# Reset database
docker-compose down
docker volume rm erp-system_postgres_data
docker-compose up -d postgres
```

### Out of Disk Space
```bash
# Clean up Docker
docker system prune -a
docker volume prune
```

## Next Steps

Now that you have the system running:

1. **[📚 Architecture Overview](architecture-overview.md)** - Understand the system design
2. **[🔨 Backend Implementation](backend-implementation.md)** - Learn to build services
3. **[📖 API Reference](api-reference.md)** - Explore available endpoints
4. **[🏢 Financial Module](modules/financial-implementation.md)** - Implement business features

## Need Help?

If you encounter issues:

1. **Check [Troubleshooting](troubleshooting.md)** - Common problems and solutions
2. **Review [FAQ](faq.md)** - Frequently asked questions
3. **Examine service logs**: `make logs`
4. **Verify prerequisites**: Ensure all required software is installed

---

**✅ System running successfully?** → Continue to [🏗️ Architecture Overview](architecture-overview.md)

**❌ Having issues?** → Check [🔧 Troubleshooting](troubleshooting.md)