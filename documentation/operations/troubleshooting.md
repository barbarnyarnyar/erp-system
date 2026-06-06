# Troubleshooting

## Service Won't Start

### Check Configuration

Verify environment variables are set correctly:

```bash
# Check if PORT is set (all services need this)
echo $PORT

# The FM service uses SERVER_PORT instead of PORT
echo $SERVER_PORT
```

### Check Port Conflicts

```bash
# See if the port is already in use
lsof -i :8001
```

Default ports: Auth=`8000`, FM=`8001`, CRM=`8002`, HR=`8003`, M=`8004`, PM=`8005`, SCM=`8006`, Gateway=`8080`

### Check Go Version

Services require Go 1.23+:

```bash
go version
```

Dockerfiles use `golang:1.21-alpine` but `go.mod` specifies `go 1.23.0`. The builder auto-downloads the newer toolchain.

## API Gateway Issues

### "Service Unavailable" on Proxy Routes

The gateway cannot reach the backend service. Check the service is running:

```bash
# Test directly (bypass gateway)
curl http://localhost:8001/health

# Verify gateway URL configuration
# Default: http://{service-name}:{port}
# e.g., http://finance-service:8001
```

### Gateway Not Running

The API Gateway is **not included in docker-compose.yml**. Start it separately:

```bash
cd api-gateway
go run cmd/main.go
```

## Kafka Issues

### Events Not Being Processed

All services silently ignore Kafka publish errors. The system works without Kafka:

```bash
# Check Kafka is running
docker compose ps kafka

# Check Kafka logs
docker compose logs kafka
```

### Consumer Not Receiving Messages

Each service uses a consumer group. Verify the group ID matches:

```bash
# Default group IDs follow the pattern: {service-name}
# fm-service, hr-service, scm-service, m-service, crm-service, pm-service
```

## Port Conflicts

Multiple services have code-default ports that differ from expected values:

| Service | Code Default | Architected Port |
|---------|-------------|-----------------|
| HR | 8003 | 8002 |
| CRM | 8002 | 8005 |
| SCM | 8006 | 8003 |

Set the `PORT` env var explicitly to override:

```bash
export PORT=8002
go run cmd/main.go
```

## Docker Build Issues

### Gateway Build Fails

The API Gateway Dockerfile must be built from the **repository root**, not from `api-gateway/`:

```bash
# Correct
docker build -f api-gateway/Dockerfile .

# Wrong — will fail because shared/ is not accessible
cd api-gateway && docker build .
```

### Image Port Mismatch

M-service and PM-service Dockerfiles hardcode `EXPOSE 8001` but the services default to different ports. Set `PORT` at runtime:

```bash
docker run -e PORT=8004 -p 8004:8004 m-service
```

## Data Loss

All services use in-memory storage. Restarting any service **deletes all data**:

```bash
# Data is lost after this
docker compose restart fm-service
```

Seed data is recreated on startup for some services (CRM seeds Acme Corporation + leads, PM seeds a portfolio + project), but any data you created manually will be gone.

## Common Error Messages

| Error | Likely Cause | Solution |
|-------|-------------|----------|
| `invalid credentials` | Wrong username/password (Auth) | Use `admin` / `admin123` |
| `user account is deactivated` | User marked inactive | Create a new user via register |
| `session expired` | Refresh token expired | Re-login |
| `404 Not Found` | Wrong URL or service port | Check port mapping table |
| `connection refused` | Service not running | `docker compose ps` to check |

## Getting Help

- Check individual service logs: `docker compose logs -f {service-name}`
- Verify all containers: `docker compose ps`
- Test health endpoints: `make health`
- API Gateway service discovery: `curl http://localhost:8080/services`
