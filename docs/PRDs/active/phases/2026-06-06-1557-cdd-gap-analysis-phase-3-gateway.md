# ERP System CDD Gap Analysis — Phase 3: Gateway Consolidation

**Source PRD**: docs/PRDs/active/2026-06-06-1557-cdd-gap-analysis.md
**PRD ID**: PRD-2026-06-06-1557
**Phase**: 3 of 6
**Status**: Draft
**Created**: June 06, 2026

---

## Objective

Consolidate the two gateway implementations into a single deployed gateway with proper authentication. Fix port mismatches between gateway routes and service code defaults. Fix Dockerfile EXPOSE directives to match actual service ports.

## Scope

### In Scope

- Replace `api-gateway/cmd/main.go` (catch-all, no auth) with `api-gateway/internal/server/server.go` (explicit routes, JWT+RBAC auth)
- Fix gateway-to-service backend URLs to match actual Go code defaults:
  - HR: `hr-service:8002` → `hr-service:8003`
  - SCM: `scm-service:8003` → `scm-service:8006`
  - CRM: `crm-service:8005` → `crm-service:8002`
- Fix Dockerfile EXPOSE ports:
  - m-service: 8001 → 8004
  - pm-service: 8001 → 8006
  - crm-service: 8001 → 8002
- Add `/health` passthrough for auth service in server.go gateway
- Update docker-compose port mappings to match fix

### Out of Scope

- HSM or certificate management (Phase 4)
- Adding more services to the gateway
- Performance tuning

---

## Inputs

| Input | Source | Notes |
| ----- | ------ | ----- |
| `api-gateway/cmd/main.go` | Current deployed gateway | Catch-all proxy, no auth |
| `api-gateway/internal/server/server.go` | Undeployed auth gateway | Explicit routes, JWT+RBAC middleware |
| Service config defaults | `services/*/internal/config/config.go` | Actual default ports per service |
| Dockerfiles | `services/*/Dockerfile` | EXPOSE directives |

## Dependencies

| Dependency | Type | Required Before | Notes |
| ---------- | ---- | --------------- | ----- |
| Phase 0 completion | Code | Phase 3 start | Routes must exist before gateway proxies to them |
| Phase 2 completion | Code | Phase 3 start | Services should be architecturally clean before routing to them |

---

## Implementation Tasks

### Task 0: Reconcile route prefix conventions

**Description:** There are 3 different route prefix conventions in use:

| Source | Prefixes | Affected Service |
|--------|----------|-----------------|
| `main.go` (deployed) + `make test` | `/finance/`, `/manufacturing/`, `/projects/` | FM, M, PM |
| `server.go` (undeployed) | `/fm/`, `/m/`, `/pm/` | FM, M, PM |
| Actual service routes | `/api/v1/accounts/`, `/api/v1/boms/`, `/api/v1/projects/` | All |

**Decision:** Align `server.go` prefixes with what `make test` expects (`/finance/`, `/manufacturing/`, `/projects/`) to avoid breaking the test suite. This means deploying the auth gateway with `main.go`'s path conventions.

**Sub-tasks:**
- 0a: Update `server.go` route prefixes from `fm` → `finance`, `m` → `manufacturing`, `pm` → `projects`
- 0b: Verify `scripts/test.sh` URLs match the chosen convention
- 0c: Document the chosen convention in architecture docs

**Acceptance Criteria:**
- `make test` URLs still work after gateway consolidation
- No 404s from path mismatches

**Files / Areas:**
- `api-gateway/internal/server/server.go`
- `scripts/test.sh`

### Task 1: Fix port mismatches in gateway

**Description:** Update gateway backend URL defaults to match actual Go service code defaults.

**Current gateway defaults:**
- HR: `http://hr-service:8002` → needs `:8003`
- SCM: `http://scm-service:8003` → needs `:8006`
- CRM: `http://crm-service:8005` → needs `:8002`

**Acceptance Criteria:**
- Gateway health checks pass for all 6 services
- Requests to `gateway/api/v1/hr/*`, `gateway/api/v1/scm/*`, `gateway/api/v1/crm/*` all reach their services

**Files / Areas:**
- `api-gateway/cmd/main.go`
- `api-gateway/internal/server/server.go`

### Task 2: Fix Dockerfile EXPOSE directives

**Description:** Update Dockerfile `EXPOSE` ports to match service code default ports.

| Service | Current EXPOSE | Code Default | Fix |
|---------|---------------|-------------|-----|
| m-service | 8001 | 8004 | EXPOSE 8004 |
| pm-service | 8001 | 8006 | EXPOSE 8006 |
| crm-service | 8001 | 8002 | EXPOSE 8002 |

**Acceptance Criteria:**
- `make build` produces Docker images with correct ports

**Files / Areas:**
- `services/m-service/Dockerfile`
- `services/pm-service/Dockerfile`
- `services/crm-service/Dockerfile`

### Task 3: Deploy new gateway (graduated, 2 checkpoints)

**Description:** Replace the current simple gateway (`cmd/main.go`) with the auth-aware gateway. Done in 2 checkpoints to minimize risk.

**Checkpoint B — Routes only, no auth (Task 3a + 3b):**

**3a:** Copy `server.go` routing logic to `cmd/main.go`, but keep auth middleware disabled. This validates that the new routing structure works.
- Files: `api-gateway/cmd/main.go`, `api-gateway/internal/server/server.go`

**3b:** Add all 6 service backend URLs to the new router, add Auth Service passthrough
- Files: `api-gateway/internal/server/server.go`

**Acceptance Criteria (Checkpoint B):**
- `make test` passes with new routing
- `make health` shows all 6 services

**Checkpoint C — Enable auth (Task 3c + 3d):**

**3c:** Enable JWT+RBAC middleware in `server.go`
- Files: `api-gateway/internal/server/server.go`

**3d:** Remove duplicate route registration from `main.go`; update `main.go` to call `server.SetupRouter(cfg)`
- Files: `api-gateway/cmd/main.go`

**Acceptance Criteria (Checkpoint C):**
- Public routes (login, register, refresh) work without auth
- Protected routes require valid JWT
- `make test` passes (update test script to obtain token)
- `make test-direct` passes (direct-to-service tests still work)

**Files / Areas:**
- `api-gateway/cmd/main.go`
- `api-gateway/internal/server/server.go`
- `api-gateway/internal/config/config.go`

### Task 4: Update docker-compose port mappings

**Description:** If docker-compose exposes ports 8000-8006 for services, fix any that mismatch with actual code defaults.

**Acceptance Criteria:**
- `docker-compose up -d` starts all services with correct port bindings
- Gateway can reach all services

**Files / Areas:**
- `docker-compose.yml`

---

## Verification

### Checkpoint A (after Tasks 0, 1, 2, 4)
```bash
# Verify ports are correct
rg '800[1-6]' api-gateway/cmd/main.go
rg 'EXPOSE' services/m-service/Dockerfile services/pm-service/Dockerfile services/crm-service/Dockerfile

# Test still works with old gateway
make test
make test-direct
make health
```

### Checkpoint B (after Tasks 3a, 3b — new gateway, no auth)
```bash
# Build new gateway
cd api-gateway && go build -o bin/main cmd/main.go

# Test all routes still work without auth
make test
make test-direct
make health
```

### Checkpoint C (after Tasks 3c, 3d — auth enabled)
```bash
# Test public routes
curl -X POST http://localhost:8080/api/v1/auth/login -d '{"username":"admin","password":"admin123"}'

# Test protected route needs token
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/finance/accounts

# Full test
make test-direct
make test
```

---

## Graduated Approach

Phase 3 is split into 3 sequential checkpoints with testing gates between each:

```
Checkpoint A (Ports):      Task 0 + Task 1 + Task 2 + Task 4
                            ↓ make test, make health pass
Checkpoint B (Routes):     Task 3a + Task 3b (server.go with reconcilied routes, no auth yet)
                            ↓ make test, make health pass
Checkpoint C (Auth):       Task 3c + Task 3d (enable JWT+RBAC middleware)
                            ↓ make test passes (with auth token)
```

If Checkpoint C breaks `make test`, gate stays at Checkpoint B (port-correct gateway without auth) and auth deployment moves to Phase 4.

## Risks

| Risk | Likelihood | Mitigation |
| ---- | ---------- | ---------- |
| Makefile test script uses different URL paths than gateway | High | Task 0 reconciles before any changes — verify `scripts/test.sh` first |
| Auth token required breaks existing `make test` | Medium | Graduated approach keeps Checkpoint B as fallback: port-correct, route-correct, no auth |
| Port changes break running containers | Medium | Update docker-compose and run `make clean && make run` between checkpoints |

## Definition of Done

- [ ] Task 0: Route prefix conventions reconciled (server.go uses `/finance/`, `/manufacturing/`, `/projects/`)
- [ ] Task 1: Gateway backend URLs match code default ports
- [ ] Task 2: Dockerfile EXPOSE matches code default ports
- [ ] Task 4: docker-compose ports consistent
- [ ] **Checkpoint A**: `make test` + `make health` + `make test-direct` all pass after port fixes
- [ ] Task 3a + 3b: New gateway with correct routes, auth disabled
- [ ] **Checkpoint B**: `make test` + `make health` pass after route swap (no auth)
- [ ] Task 3c + 3d: JWT+RBAC auth middleware enabled
- [ ] **Checkpoint C**: `make test` passes with auth token
- [ ] `make health` shows all 6 services healthy at every checkpoint

---

## Handoff Notes

After Phase 3, the gateway will be a single deployed implementation with real authentication. Phase 4 (Security Hardening) should follow immediately to fix the hardcoded JWT secret and plaintext passwords that are now exposed through the auth gateway.
