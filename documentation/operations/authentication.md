# Authentication

## Current State

The deployed API Gateway (`cmd/main.go`) has **no authentication**. All endpoints are publicly accessible.

A full JWT-based authentication system exists in `internal/server/server.go` and `internal/middleware/auth.go` but is **not wired into the running binary**.

## Auth Service

The Auth Service runs on port 8000 and provides:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/auth/login` | POST | Authenticate with username + password, receive JWT |
| `/api/v1/auth/register` | POST | Create a new user account |
| `/api/v1/auth/refresh` | POST | Exchange a refresh token for new tokens |
| `/api/v1/auth/logout` | POST | Revoke a refresh token |

### Login

```bash
curl -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

Response:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "rt_1749267184000000000_usr_...",
  "token_type": "Bearer"
}
```

### Using the JWT

Include the token in the `Authorization` header for subsequent requests:

```bash
curl http://localhost:8000/api/v1/auth/validate \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

### Token Refresh

```bash
curl -X POST http://localhost:8000/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"rt_1749267184000000000_usr_..."}'
```

### Logout

```bash
curl -X POST http://localhost:8000/api/v1/auth/logout \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"rt_1749267184000000000_usr_..."}'
```

## Default Credentials

| Username | Password | Role |
|----------|----------|------|
| `admin` | `admin123` | Admin |

## Known Issues

- **No auth on deployed gateway** — the running API gateway does not enforce any authentication despite the code existing
- **Plaintext passwords** — passwords are stored and compared as plaintext (not hashed)
- **Hardcoded JWT secret** — default is `super-secret-key-123`, set via `JWT_SECRET` env var
- **No TLS** — all communication is plain HTTP
