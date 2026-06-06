# Installation

Get the ERP system running in 15 minutes with these step-by-step instructions.

## Quick Installation

**Step 1: Clone the Repository**
```bash
git clone https://github.com/your-org/erp-system.git
cd erp-system
```

**Step 2: Start Infrastructure Services**
```bash
# Start PostgreSQL, Redis, Zookeeper, and Kafka
docker compose up -d postgres kafka redis
# Note: PostgreSQL and Redis are not currently connected to any service.
# They run for future migration readiness. Kafka is used for event messaging.
```

**Step 3: Start All Application Services**
```bash
# Build and start all microservices
docker compose up -d
```

**Step 4: Start the API Gateway**

The API Gateway is not in docker-compose. Start it separately:

```bash
cd api-gateway
go run cmd/main.go
```

**Step 5: Verify Installation**

In a separate terminal:

```bash
# Check all services are running
make health
```

Expected output:
```
Checking API Gateway...         http://localhost:8080/health
Checking Auth Service...        http://localhost:8000/health
Checking Financial Service...   http://localhost:8001/health
Checking HR Service...          http://localhost:8002/health
Checking SCM Service...         http://localhost:8003/health
Checking Manufacturing Service... http://localhost:8004/health
Checking CRM Service...         http://localhost:8005/health
Checking Project Service...     http://localhost:8006/health
```

**Step 6: Test API Endpoints**
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

## Manual API Tests

```bash
# Test API Gateway routing (hello endpoints)
curl http://localhost:8080/api/v1/finance/hello
curl http://localhost:8080/api/v1/hr/hello
curl http://localhost:8080/api/v1/scm/hello
curl http://localhost:8080/api/v1/manufacturing/hello
curl http://localhost:8080/api/v1/crm/hello
curl http://localhost:8080/api/v1/projects/hello

# Test with data (no auth required — gateway has no authentication)
curl -X POST http://localhost:8080/api/v1/finance/accounts \
  -H "Content-Type: application/json" \
  -d '{
    "account_code": "1000",
    "account_name": "Cash - Operating",
    "account_type": "ASSET"
  }'
```

## Known Installation Notes

- **API Gateway runs separately**: `docker compose up` does not start the gateway. Run it manually from `api-gateway/`.
- **No authentication**: The deployed gateway has no auth. All endpoints are public. See [Authentication](../operations/authentication.md) for the inactive JWT system.
- **Data is in-memory**: All services lose data on restart. Seed data is recreated for some services.
- **No database connection**: PostgreSQL runs but no service connects to it. SQL migration files exist but are never executed.
- **Kafka is optional**: Event publishing errors are silently ignored. The system works without Kafka.

## Next Steps

With the system successfully installed:

1. **Configure your environment**: [Configuration Guide](configuration.md)
2. **Learn the architecture**: [System Overview](../architecture/README.md)
3. **Explore the features**: [Business Modules](../modules/README.md)
4. **Set up development tools**: [Development Environment](development-environment.md)

## Troubleshooting

If you encounter issues during installation, check the [Common Setup Issues](common-issues.md) guide for solutions.
