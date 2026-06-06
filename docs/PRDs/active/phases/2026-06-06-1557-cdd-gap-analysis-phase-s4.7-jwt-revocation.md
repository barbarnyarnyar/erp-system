# ERP System CDD Gap Analysis — Phase S4.7: JWT Revocation via security_stamp + is_revoked

**Source PRD**: docs/PRDs/active/2026-06-06-1557-cdd-gap-analysis.md
**PRD ID**: PRD-2026-06-06-1557
**Phase**: S4.7 of 6
**Status**: Completed
**Created**: June 06, 2026

---

## Objective

Close the JWT-revocation gap. Today, when a user is deactivated, has their password changed, or has their roles updated, **existing JWTs remain valid until natural expiration** (typically 1 hour for access tokens). A terminated employee can keep using their old token for the remainder of the access window.

The fix embeds the user's current `security_stamp` in the JWT at issuance, and re-validates it on every request against the live user record.

## Rationale

| Issue | Impact |
|-------|--------|
| `User` has no `security_stamp` field | No version counter to invalidate tokens on state change |
| `Session` has no `is_revoked` field | No per-session kill switch for explicit logout |
| `JWTClaims` carries no user-state fingerprint | Tokens are trusted for their full TTL regardless of account changes |
| `deactivateUser()` only sets `IsActive=false` | Real security gap: zombie access until token expires |

## Scope

### In Scope
- Add `security_stamp: string` to `User` (CDD + Go struct)
- Add `is_revoked: boolean` to `Session` (CDD + Go struct)
- Set initial `SecurityStamp` in `UserService.CreateUser`
- Bump `SecurityStamp` in `UserService.DeactivateUser`
- Bump `SecurityStamp` in `UserService.UpdateCredentials`
- Bump `SecurityStamp` in `UserService.UpdateUser` when `isActive` toggles
- Add `SecurityStamp` field to `JWTClaims`
- Embed `user.SecurityStamp` in `AuthenticateUser` claims
- Update `AuthService.ValidateToken` to:
  1. Reload user from repo
  2. Reject if `!user.IsActive`
  3. Reject if `claims.SecurityStamp != user.SecurityStamp`
- Update `AuthService.RevokeToken` to set `session.IsRevoked = true` (was `sessRepo.Delete`)
- Add `Update` method to `SessionRepository` interface + memory impl
- Unit tests covering all scenarios

### Out of Scope
- Refresh-token rotation tied to security_stamp (deferred — out of scope for S4.7)
- Real-time session list per user
- Audit log of who/when bumped a stamp
- A separate `UserRoleService.AssignRole` method (no such method exists yet; roles are set at CreateUser time)

---

## Implementation Tasks

### Task 1: CDD + struct updates
- `services/auth-service/contracts/auth.cdd` — add `security_stamp string @optional` to User, `is_revoked boolean @optional` to Session
- `services/auth-service/internal/business/domain/user.go` — add field
- `services/auth-service/internal/business/domain/session.go` — add field

### Task 2: SessionRepository Update
- `services/auth-service/internal/business/domain/repository.go` — add `Update(ctx, *Session) error`
- `services/auth-service/internal/data/memory/memory_repos.go` — implement

### Task 3: UserService bumps
- `CreateUser`: set `u.SecurityStamp = fmt.Sprintf("ss_%d", time.Now().UnixNano())`
- `DeactivateUser`: bump before save
- `UpdateCredentials`: bump before save
- `UpdateUser`: bump if `isActive` flips (covers reactivation too)

### Task 4: AuthService JWT
- `JWTClaims.SecurityStamp string` field
- `AuthenticateUser`: populate from `user.SecurityStamp`
- `ValidateToken`: reload user, check `IsActive` and `SecurityStamp` match
- `RevokeToken`: `session.IsRevoked = true; sessRepo.Update`

### Task 5: Tests
- File: `services/auth-service/internal/business/service/security_stamp_test.go`
- 5 test functions:
  1. `TestUser_CreateUser_SetsSecurityStamp`
  2. `TestUser_DeactivateUser_BumpsSecurityStamp`
  3. `TestUser_UpdateCredentials_BumpsSecurityStamp`
  4. `TestAuth_ValidateToken_RejectsStaleSecurityStamp` (the key integration test)
  5. `TestAuth_RevokeToken_SetsIsRevoked`

---

## Verification

```bash
cd services/auth-service && go test ./internal/business/service/ -run "TestUser_|TestAuth_" -v
# All 5 pass

for svc in auth fm hr scm m crm pm; do
  (cd services/$svc-service && go build ./...)
done
# All 7 services build cleanly
```

---

## Risks

| Risk | Likelihood | Mitigation |
| ---- | ---------- | ---------- |
| Performance — ValidateToken now does a DB lookup per request | Medium | In-memory map lookup; sub-microsecond. Real DB would add a JOIN/RPC cost. Acceptable for security gain. |
| Existing tests break because User now has a new field | Low | All struct fields are optional + zero-value compatible. Bcrypt hashes are unchanged. |
| Future role-change flow needs to bump stamp | Low | Pattern is established; a future `RBACService.AssignRole` / `RevokeRole` will follow the same one-line bump. |
| RefreshToken path reuses AuthenticateUser which re-checks bcrypt | None | The bcrypt check is on the user's stored hash, not the refresh token. Path is unchanged. |

## Definition of Done

- [x] `User.security_stamp` field added in CDD + Go
- [x] `Session.is_revoked` field added in CDD + Go
- [x] `SessionRepository.Update` added to interface + memory impl
- [x] `JWTClaims.SecurityStamp` populated at issuance
- [x] `ValidateToken` rejects deactivated users
- [x] `ValidateToken` rejects stale security_stamp
- [x] `RevokeToken` sets `IsRevoked` (not delete)
- [x] 5 unit tests pass
- [x] All 7 services build cleanly
- [x] Master PRD 2.17 DoD checkbox marked complete

## Handoff Notes

This closes a real security vulnerability: terminated employees previously retained access for the full access-token TTL (default 1 hour). Now the moment `DeactivateUser` is called, any in-flight JWT becomes invalid because the stamp embedded in the token no longer matches the live user record.

The `ValidateToken` cost is now one user lookup per request. For a production deployment, this could be cached briefly (e.g., 30s) in Redis to reduce DB load — but that's an optimization for S14/S15, not a correctness requirement.
