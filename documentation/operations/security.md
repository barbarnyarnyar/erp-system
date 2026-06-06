# Security Configuration

## Current State

The deployed system has **no security**. All API endpoints are publicly accessible.

### What Exists (Inactive)

A full JWT-based security system is implemented in two locations but **not deployed**:

| Component | Location | Status |
|-----------|----------|--------|
| JWT auth middleware | `api-gateway/internal/middleware/auth.go` | Defined, not wired |
| Downstream auth | `api-gateway/internal/middleware/auth_client.go` | Defined, not used |
| CORS middleware | `api-gateway/internal/server/server.go` | Defined, not deployed |
| Rate limiter | `api-gateway/internal/middleware/rate_limit.go` | Defined, not wired |
| Auth service | `services/auth-service/` | Running on port 8000 |

### Auth Service

The Auth Service issues JWT tokens and manages RBAC:

```bash
# Login as admin
curl -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

### JWT Token

Signed with HMAC-SHA256. Contains user ID, username, email, roles, and permissions.

**Default secret**: `super-secret-key-123` (change via `JWT_SECRET` env var)

### Default Roles

| Role | Permissions |
|------|-------------|
| Admin | Full access (scm:product:create/read, crm:customer:create/read) |
| Manager | Read access (scm:product:read, crm:customer:read) |
| Clerk | Read products (scm:product:read) |

## Enabling Authentication

To deploy the full auth system, switch the API Gateway from `cmd/main.go` to `internal/server/server.go`:

1. Update `api-gateway/Dockerfile` to build `internal/server/server.go` instead of `cmd/main.go`
2. Set `JWT_SECRET` environment variable (change from default)
3. Set `{SERVICE}_SERVICE_URL` env vars for each backend (8081-8086 ports)

## Security Checklist

### Required Before Production

- [ ] Wire the existing auth middleware into the gateway
- [ ] Hash passwords with bcrypt (replace plaintext comparison in auth service)
- [ ] Change the default JWT secret
- [ ] Enable TLS/HTTPS
- [ ] Add CORS middleware

### Recommended

- [ ] Add rate limiting (middleware exists, just needs wiring)
- [ ] Add audit logging for authentication events
- [ ] Replace predictable refresh tokens with cryptographically random values
- [ ] Change default admin credentials (`admin` / `admin123`)
- [ ] Add input validation and sanitization

## Known Vulnerabilities

| Issue | Severity | Details |
|-------|----------|---------|
| No authentication on deployed gateway | Critical | All endpoints are public |
| Plaintext passwords | Critical | Passwords stored and compared without hashing |
| Hardcoded JWT secret | Critical | Default secret visible in source code |
| No TLS | High | All traffic unencrypted |
| Predictable refresh tokens | High | Format: `rt_{timestamp}_{user_id}` |
| No CSRF protection | Medium | No anti-CSRF tokens |
| Verbose error messages | Low | May leak internal details |
