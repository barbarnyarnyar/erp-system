# Configuration

Configuration reference for the ERP system. All services use environment variables — there are no config files.

## How Configuration Works

Each service loads config via `internal/config/config.go` which reads `os.Getenv()` with defaults. No external config library (viper, etc.) is used.

## Service Configuration Reference

### Auth Service (port 8000)

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8000` | HTTP server port |
| `ENV` | `development` | Runtime environment |
| `JWT_SECRET` | `super-secret-key-123` | HMAC signing key (hardcoded default) |
| `JWT_ACCESS_EXPIRY` | `60` | Access token expiry in minutes |
| `JWT_REFRESH_EXPIRY` | `24` | Refresh token expiry in hours |
| `KAFKA_BROKERS` | `localhost:9092` | Comma-separated Kafka brokers |

### Financial Management (port 8001)

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8001` | HTTP server port |
| `ENV` | `development` | Runtime environment |
| `DB_HOST` | `localhost` | PostgreSQL host (not connected) |
| `DB_PORT` | `5432` | PostgreSQL port (not connected) |
| `DB_USERNAME` | `postgres` | Database user (not connected) |
| `DB_PASSWORD` | `` | Database password (not connected) |
| `DB_DATABASE` | `fm_service` | Database name (not connected) |
| `REDIS_HOST` | `localhost` | Redis host (not connected) |
| `REDIS_PORT` | `6379` | Redis port (not connected) |
| `REDIS_PASSWORD` | `` | Redis password (not connected) |
| `REDIS_DB` | `0` | Redis database number (not connected) |
| `RABBITMQ_URL` | `amqp://guest:guest@localhost:5672/` | RabbitMQ URL (unused — only Kafka is active) |
| `KAFKA_BROKERS` | `localhost:9092` | Comma-separated Kafka brokers |
| `KAFKA_GROUP_ID` | `fm-service` | Kafka consumer group |

### HR, SCM, M, CRM, PM Services

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | varies | HTTP server port |
| `ENV` | `development` | Runtime environment |
| `KAFKA_BROKERS` | `localhost:9092` | Comma-separated Kafka brokers |
| `KAFKA_GROUP_ID` | `{service-name}` | Kafka consumer group |

**Default ports by service**: HR=`8003`, SCM=`8006`, M=`8004`, CRM=`8002`, PM=`8006`

### API Gateway

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `FINANCE_SERVICE_URL` | `http://finance-service:8001` | Backend service URL |
| `HR_SERVICE_URL` | `http://hr-service:8002` | Backend service URL |
| `SCM_SERVICE_URL` | `http://scm-service:8003` | Backend service URL |
| `MANUFACTURING_SERVICE_URL` | `http://manufacturing-service:8004` | Backend service URL |
| `CRM_SERVICE_URL` | `http://crm-service:8005` | Backend service URL |
| `PROJECTS_SERVICE_URL` | `http://projects-service:8006` | Backend service URL |
| `JWT_SECRET` | `your-super-secret-key` | JWT validation key (inactive server only) |

## Setting Configuration

### Docker Compose

Variables are set per-service in `docker-compose.yml`:

```yaml
services:
  fm-service:
    environment:
      - PORT=8001
```

### Local Development

```bash
export PORT=8001
export KAFKA_BROKERS=localhost:9092
go run cmd/main.go
```

### Environment-Specific Overrides

The `ENV` variable (default `development`) can be set to `production` to enable Gin release mode:

```go
if cfg.Server.Env == "production" {
    gin.SetMode(gin.ReleaseMode)
}
```

## Known Configuration Issues

- **FM service uses `SERVER_PORT`** while all other services use `PORT`
- **Database and Redis configs exist but no service connects to them** — these are placeholders for future migration
- **RabbitMQ config is defined but unused** — only Kafka is used for messaging
- **JWT secret default is hardcoded** in source code as `super-secret-key-123`
- **The inactive gateway server** uses port range 8081-8086 for backends, while the deployed gateway uses 8001-8006

## Next Steps

With configuration complete:
- [Development Environment](development-environment.md) — Set up for coding
- [System Architecture](../architecture/README.md) — Understand the system
