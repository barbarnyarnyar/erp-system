# Configuration Management

All services use environment variables for configuration. There are no config files.

## Global Configuration Pattern

Every service loads config via a `config.Load()` function in `internal/config/config.go` that reads `os.Getenv` with defaults. No external config library (viper, etc.) is used.

## Service Configuration Reference

### Auth Service (port 8000)

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8000` | HTTP server port |
| `ENV` | `development` | Runtime environment |
| `JWT_SECRET` | `super-secret-key-123` | HMAC signing key for JWT |
| `JWT_ACCESS_EXPIRY` | `60` | Access token expiry in minutes |
| `JWT_REFRESH_EXPIRY` | `24` | Refresh token expiry in hours |
| `KAFKA_BROKERS` | `localhost:9092` | Comma-separated Kafka brokers |

### Financial Management (port 8001)

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8001` | HTTP server port |
| `ENV` | `development` | Runtime environment |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USERNAME` | `postgres` | Database user |
| `DB_PASSWORD` | `` | Database password |
| `DB_DATABASE` | `fm_service` | Database name |
| `DB_SSL_MODE` | `disable` | PostgreSQL SSL mode |
| `REDIS_HOST` | `localhost` | Redis host |
| `REDIS_PORT` | `6379` | Redis port |
| `REDIS_PASSWORD` | `` | Redis password |
| `REDIS_DB` | `0` | Redis database number |
| `RABBITMQ_URL` | `amqp://guest:guest@localhost:5672/` | RabbitMQ connection URL (unused) |
| `KAFKA_BROKERS` | `localhost:9092` | Comma-separated Kafka brokers |
| `KAFKA_GROUP_ID` | `fm-service` | Kafka consumer group |

### HR, SCM, M, CRM, PM Services

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | varies (see below) | HTTP server port |
| `ENV` | `development` | Runtime environment |
| `KAFKA_BROKERS` | `localhost:9092` | Comma-separated Kafka brokers |
| `KAFKA_GROUP_ID` | `{service-name}` | Kafka consumer group |

**Default ports by service**: HR=`8003`, SCM=`8006`, M=`8004`, CRM=`8002`, PM=`8006`

### API Gateway

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `{SERVICE}_SERVICE_URL` | varies | Backend service URLs (`FINANCE_SERVICE_URL`, `HR_SERVICE_URL`, etc.) |
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

Use a `.env` file or export variables:

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

## Known Issues

- FM service uses `SERVER_PORT` while all others use `PORT`
- RabbitMQ config is defined but never used (only Kafka is active)
- Database and Redis configs are defined but no service connects to them
- JWT secret default is hardcoded in source code
