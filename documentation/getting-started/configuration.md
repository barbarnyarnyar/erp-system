# Configuration

Configure the ERP system for different environments and use cases.

## Service Configuration

Each service uses environment variables for configuration. Here are the key settings:

### Database Configuration
```bash
# PostgreSQL settings
DB_HOST=postgres
DB_PORT=5432
DB_USERNAME=postgres
DB_PASSWORD=postgres
DB_DATABASE=financial_db  # Service-specific database name
DB_SSLMODE=disable        # For development only

# Connection pool settings
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=300s
```

### Redis Configuration
```bash
# Redis cache settings
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=           # Empty for development
REDIS_DB=0
REDIS_MAX_IDLE=10
REDIS_MAX_ACTIVE=100
```

### Kafka Configuration
```bash
# Message broker settings
KAFKA_BROKERS=kafka:9092
KAFKA_GROUP_ID=erp-system
KAFKA_AUTO_OFFSET_RESET=earliest
KAFKA_ENABLE_AUTO_COMMIT=true
```

### Service-Specific Configuration
```bash
# JWT Authentication
JWT_SECRET=your-super-secret-jwt-key-here
JWT_EXPIRATION=24h
JWT_REFRESH_EXPIRATION=168h  # 7 days

# API Rate Limiting
RATE_LIMIT_REQUESTS=1000
RATE_LIMIT_WINDOW=3600s      # 1 hour

# File Upload Settings
MAX_UPLOAD_SIZE=10MB
UPLOAD_PATH=/app/uploads

# Email Configuration (optional)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
```

## Multi-Environment Configuration

### Development Environment
Create a `docker-compose.override.yml` for development-specific settings:

```yaml
version: '3.8'
services:
  fm-service:
    environment:
      - LOG_LEVEL=debug
      - DB_LOG_LEVEL=info
    volumes:
      - ./services/fm-service:/app
    command: air  # Hot reload
```

### Staging Environment
Create a `docker-compose.staging.yml` for staging:

```yaml
version: '3.8'
services:
  fm-service:
    environment:
      - LOG_LEVEL=warn
      - DB_HOST=staging-postgres.company.com
      - REDIS_HOST=staging-redis.company.com
    deploy:
      replicas: 2
      resources:
        limits:
          cpus: '1.0'
          memory: 1G
```

### Production Environment
Create a `docker-compose.prod.yml` for production:

```yaml
version: '3.8'
services:
  fm-service:
    environment:
      - LOG_LEVEL=error
      - DB_HOST=prod-postgres.company.com
      - DB_PASSWORD_FILE=/run/secrets/db_password
    secrets:
      - db_password
    deploy:
      replicas: 3
      resources:
        limits:
          cpus: '2.0'
          memory: 2G
```

## Environment File Management

### Creating Environment Files
```bash
# Create environment-specific files
cp .env.example .env.development
cp .env.example .env.staging  
cp .env.example .env.production
```

### Loading Environment Files
```bash
# Development
export $(cat .env.development | grep -v '^#' | xargs)

# Staging
export $(cat .env.staging | grep -v '^#' | xargs)

# Production (use secrets management)
# Don't export production secrets directly
```

## Security Configuration

### JWT Settings
```bash
# Generate a secure JWT secret
JWT_SECRET=$(openssl rand -base64 32)
echo "JWT_SECRET=$JWT_SECRET" >> .env.production
```

### Database Security
```bash
# Production database settings
DB_SSLMODE=require
DB_SSLCERT=/certs/client-cert.pem
DB_SSLKEY=/certs/client-key.pem
DB_SSLROOTCERT=/certs/ca-cert.pem
```

### API Security
```bash
# CORS settings
CORS_ALLOWED_ORIGINS=https://erp.company.com,https://app.company.com
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE
CORS_ALLOWED_HEADERS=Origin,Content-Type,Authorization

# Rate limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=300s  # 5 minutes
```

## Service Ports

### Default Port Configuration
```bash
# API Gateway
API_GATEWAY_PORT=8080

# Core Services
FM_SERVICE_PORT=8001    # Financial Management
HR_SERVICE_PORT=8002    # Human Resources
SCM_SERVICE_PORT=8003   # Supply Chain
CRM_SERVICE_PORT=8004   # Customer Relations
MFG_SERVICE_PORT=8005   # Manufacturing
PM_SERVICE_PORT=8006    # Project Management

# Infrastructure
DB_PORT=5432           # PostgreSQL
REDIS_PORT=6379        # Redis
KAFKA_PORT=9092        # Kafka
```

### Custom Port Configuration
If default ports conflict with existing services, update the configuration:

```yaml
# docker-compose.yml
services:
  api-gateway:
    ports:
      - "8090:8080"  # Change external port
    environment:
      - PORT=8080    # Keep internal port
```

## Logging Configuration

### Log Levels
```bash
# Available log levels
LOG_LEVEL=debug   # Development
LOG_LEVEL=info    # Staging
LOG_LEVEL=warn    # Production
LOG_LEVEL=error   # Critical only
```

### Log Formats
```bash
# JSON format for production
LOG_FORMAT=json

# Text format for development
LOG_FORMAT=text
```

## Configuration Validation

Verify your configuration is correct:

```bash
# Check environment variables
docker-compose config

# Validate service connections
make health

# Test configuration changes
docker-compose up -d --force-recreate fm-service
docker-compose logs fm-service
```

## Next Steps

With configuration complete:
- [Development Environment](development-environment.md) - Set up for coding
- [Testing and Verification](testing-verification.md) - Ensure everything works
- [Architecture Overview](../architecture/README.md) - Understand the system