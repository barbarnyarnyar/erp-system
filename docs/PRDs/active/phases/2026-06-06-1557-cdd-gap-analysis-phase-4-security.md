# ERP System CDD Gap Analysis — Phase 4: Security Hardening

**Source PRD**: docs/PRDs/active/2026-06-06-1557-cdd-gap-analysis.md
**PRD ID**: PRD-2026-06-06-1557
**Phase**: 4 of 6
**Status**: Draft
**Created**: June 06, 2026

---

## Objective

Fix the most critical security issues: plaintext password storage, hardcoded JWT secret, and prepare TLS readiness.

## Scope

### In Scope

- Migrate password storage from plaintext to bcrypt
- Move JWT secret from hardcoded constant to environment variable
- Add TLS configuration stubs in service configs
- Add a default admin user seed on auth service startup

### Out of Scope

- Full TLS implementation (certificate management, mutual TLS)
- OAuth2 or SSO integration
- Audit logging
- Rate limiting activation

---

## Inputs

| Input | Source | Notes |
| ----- | ------ | ----- |
| Auth service | `services/auth-service/` | Contains password comparison, JWT generation |
| Other service configs | `services/*/internal/config/config.go` | TLS fields may need adding |

## Dependencies

| Dependency | Type | Required Before | Notes |
| ---------- | ---- | --------------- | ----- |
| Phase 3 | Code | Phase 4 start | Auth gateway deployed makes JWT fix urgent |

---

## Implementation Tasks

### Task 1: Migrate to bcrypt password hashing

**Description:** Replace plaintext password comparison with bcrypt.

**Current (auth_service.go):**
```go
if user.PasswordHash != password {
```
**Target:**
```go
if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
```

**Registration (user_service.go):** Hash password before storing:
```go
hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
user.PasswordHash = string(hash)
```

**Acceptance Criteria:**
- New user registration stores bcrypt hash (not plaintext)
- Login with correct password succeeds
- Login with incorrect password fails
- `golang.org/x/crypto` added to go.mod if not present

**Files / Areas:**
- `services/auth-service/internal/business/service/auth_service.go`
- `services/auth-service/internal/business/service/user_service.go`
- `services/auth-service/go.mod`

### Task 2: Externalize JWT secret

**Description:** Move hardcoded JWT secret to environment variable with a dev-only fallback.

**Current (auth_service.go):**
```go
var jwtSecret = []byte("super-secret-key-123")
```

**Target:**
```go
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))
```

**Config struct update:**
```go
type Config struct {
    JWTSecret string
    // ...
}
```

If `JWT_SECRET` is empty in dev, log a warning and generate a random one per session (so devs notice if they restart).

**Acceptance Criteria:**
- Setting `JWT_SECRET=my-secret-key` in env uses that value
- Unset `JWT_SECRET` logs a warning and uses a random per-run secret
- Previous hardcoded value removed

**Files / Areas:**
- `services/auth-service/internal/business/service/auth_service.go`
- `services/auth-service/internal/config/config.go`
- `docker-compose.yml` — add `JWT_SECRET` to auth-service environment

### Task 3: Add TLS config stubs (split by service)

**Description:** Add `TLS_ENABLED`, `TLS_CERT_FILE`, `TLS_KEY_FILE` config fields to all services. The fields are parsed but TLS is not yet active.

**Sub-tasks (one per service, ~15 min each):**
- **3a**: Auth service — `services/auth-service/internal/config/config.go`
- **3b**: FM service — `services/fm-service/internal/config/config.go`
- **3c**: HR service — `services/hr-service/internal/config/config.go`
- **3d**: SCM service — `services/scm-service/internal/config/config.go`
- **3e**: M service — `services/m-service/internal/config/config.go`
- **3f**: CRM service — `services/crm-service/internal/config/config.go`
- **3g**: PM service — `services/pm-service/internal/config/config.go`

Each sub-task: add `TLSEnabled bool`, `TLSCertFile string`, `TLSKeyFile string` fields to Config struct, parse from env vars with defaults (disabled, empty).

**Acceptance Criteria:**
- All 7 services have `TLS.Enabled`, `TLS.CertFile`, `TLS.KeyFile` config fields
- Default is disabled
- No behavioral change

**Files / Areas:**
- `services/*/internal/config/config.go` — all 7 services

### Task 4: Add seed admin user

**Description:** On first auth service startup, if no users exist, create a default admin user with known credentials. This is essential because auth gateway (Phase 3) requires login.

**Default credentials:** `admin` / `admin123`

**Acceptance Criteria:**
- Auth service creates admin user on first startup
- Admin user is usable for gateway login testing
- Seed only runs if user table is empty

**Files / Areas:**
- `services/auth-service/cmd/main.go`
- `services/auth-service/internal/business/service/user_service.go`

---

## Verification

```bash
# Test password hashing
curl -X POST http://localhost:8080/api/v1/auth/register \
  -d '{"username":"test","email":"test@test.com","password":"secret123"}'

# Verify stored hash starts with $2a$ (bcrypt prefix)
# Check the user_store.go memory or add a GET endpoint

# Test login with new JWT secret
JWT_SECRET=my-secret-key go run cmd/main.go &
curl -X POST http://localhost:8080/api/v1/auth/login \
  -d '{"username":"admin","password":"admin123"}'
# Should receive valid JWT
```

---

## Risks

| Risk | Likelihood | Mitigation |
| ---- | ---------- | ---------- |
| bcrypt breaks existing stored passwords | High | In-memory storage — no persisted passwords to migrate. Seed only, no migration needed |
| JWT secret change invalidates existing tokens | Low | In-memory tokens — restart clears all sessions |
| Admin seed creates security hole in production | Low | Seed only runs when user store is empty; production should use real user creation |

## Definition of Done

- [ ] Task 1: Passwords hashed with bcrypt
- [ ] Task 2: JWT secret is environment variable
- [ ] Task 3a–3g: TLS config stubs in all 7 services
- [ ] Task 4: Admin user seed on first startup (credentials: admin / admin123 — consistent with Phase 3)
- [ ] Auth service login flow works end-to-end
- [ ] `make build` passes for auth-service
