package service

import (
	sharedtesting "erp-system/shared/testing"
	"context"
	"testing"
	"time"

	"github.com/erp-system/auth-service/internal/business/domain"
	"github.com/erp-system/auth-service/internal/config"
	"github.com/erp-system/auth-service/internal/data/memory"
)

func newTestConfig() *config.Config {
	return &config.Config{
		JWT: config.JWTConfig{
			Secret:        "test-secret-key-for-unit-tests-only",
			AccessExpiry:  60,
			RefreshExpiry: 24,
		},
	}
}

func newAuthService(t *testing.T) (*AuthService, *UserService, *memory.UserRepository, *memory.SessionRepository) {
	t.Helper()
	userRepo := memory.NewUserRepository()
	sessRepo := memory.NewSessionRepository()
	roleRepo := memory.NewRoleRepository()
	permRepo := memory.NewPermissionRepository()
	urRepo := memory.NewUserRoleRepository()
	rpRepo := memory.NewRolePermissionRepository()
	usRepo := memory.NewUserStoreRepository()
	pub := &sharedtesting.MockPublisher{}
	rbacSvc := NewRBACService(roleRepo, permRepo, urRepo, rpRepo, pub)
	authSvc := NewAuthService(userRepo, sessRepo, rbacSvc, pub, newTestConfig())
	userSvc := NewUserService(userRepo, usRepo, urRepo, pub)
	return authSvc, userSvc, userRepo, sessRepo
}

// TestUser_CreateUser_SetsSecurityStamp verifies that CreateUser assigns an
// initial security_stamp to every new user.
func TestUser_CreateUser_SetsSecurityStamp(t *testing.T) {
	_, userSvc, _, _ := newAuthService(t)
	ctx := context.Background()

	u := &domain.User{
		Username:     "alice",
		Email:        "alice@example.com",
		PasswordHash: "x",
		FirstName:    "Alice",
		LastName:     "A",
	}
	created, err := userSvc.CreateUser(ctx, u, "store_1", nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.SecurityStamp == "" {
		t.Errorf("expected security_stamp to be set on new user, got empty")
	}
}

// TestUser_DeactivateUser_BumpsSecurityStamp is the regression test for
// Phase S4.7: a deactivated user must not be able to use their old JWT.
func TestUser_DeactivateUser_BumpsSecurityStamp(t *testing.T) {
	_, userSvc, userRepo, _ := newAuthService(t)
	ctx := context.Background()

	u := &domain.User{Username: "bob", Email: "b@e.com", PasswordHash: "x", FirstName: "B", LastName: "B"}
	created, _ := userSvc.CreateUser(ctx, u, "", nil)
	originalStamp := created.SecurityStamp

	if err := userSvc.DeactivateUser(ctx, created.ID); err != nil {
		t.Fatalf("deactivate: %v", err)
	}

	fresh, err := userRepo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if fresh.Status == domain.UserStatusACTIVE {
		t.Errorf("expected IsActive=false after DeactivateUser")
	}
	if fresh.SecurityStamp == originalStamp {
		t.Errorf("expected security_stamp to change after deactivation, still %q", originalStamp)
	}
}

// TestUser_UpdateCredentials_BumpsSecurityStamp verifies password change
// invalidates all in-flight JWTs.
func TestUser_UpdateCredentials_BumpsSecurityStamp(t *testing.T) {
	_, userSvc, userRepo, _ := newAuthService(t)
	ctx := context.Background()

	u := &domain.User{Username: "carol", Email: "c@e.com", PasswordHash: "x", FirstName: "C", LastName: "C"}
	created, _ := userSvc.CreateUser(ctx, u, "", nil)
	originalStamp := created.SecurityStamp

	if _, err := userSvc.UpdateCredentials(ctx, created.ID, "new-password-123"); err != nil {
		t.Fatalf("update creds: %v", err)
	}

	fresh, _ := userRepo.GetByID(ctx, created.ID)
	if fresh.SecurityStamp == originalStamp {
		t.Errorf("expected security_stamp to change after password update")
	}
}

// TestAuth_ValidateToken_RejectsStaleSecurityStamp is the integration test:
// issue a JWT, then deactivate the user, then attempt ValidateToken.
// It must fail with a security_stamp mismatch error.
func TestAuth_ValidateToken_RejectsStaleSecurityStamp(t *testing.T) {
	authSvc, userSvc, userRepo, sessRepo := newAuthService(t)
	ctx := context.Background()

	// Pass raw password; CreateUser will bcrypt it.
	u := &domain.User{
		Username:     "dave",
		Email:        "d@e.com",
		PasswordHash: "password-123",
		FirstName:    "D",
		LastName:     "D",
	}
	created, _ := userSvc.CreateUser(ctx, u, "", nil)

	// Manually create a session for the user (bypasses the JWT issuance path
	// to keep the test focused on ValidateToken).
	sess := &domain.Session{
		ID:           "sess_test_1",
		UserID:       created.ID,
		RefreshToken: "rt_test_1",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		CreatedAt:    time.Now(),
	}
	_ = sessRepo.Create(ctx, sess)

	// Issue token.
	token, _, err := authSvc.AuthenticateUser(ctx, "dave", "password-123", "127.0.0.1", "test")
	if err != nil {
		t.Fatalf("auth: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	// Validate: should succeed.
	if _, err := authSvc.ValidateToken(ctx, token); err != nil {
		t.Errorf("expected valid token, got: %v", err)
	}

	// Deactivate the user — this bumps SecurityStamp.
	if err := userSvc.DeactivateUser(ctx, created.ID); err != nil {
		t.Fatalf("deactivate: %v", err)
	}

	// Validate: should now FAIL.
	_, err = authSvc.ValidateToken(ctx, token)
	if err == nil {
		t.Errorf("expected ValidateToken to reject token after deactivation, got nil")
	}

	// Also confirm IsActive is false.
	fresh, _ := userRepo.GetByID(ctx, created.ID)
	if fresh.Status == domain.UserStatusACTIVE {
		t.Errorf("expected user to be inactive")
	}
}

// TestAuth_RevokeToken_SetsIsRevoked verifies that RevokeToken marks the
// session as revoked rather than deleting it (Phase S4.7: enables explicit
// logout before token natural expiration).
func TestAuth_RevokeToken_SetsIsRevoked(t *testing.T) {
	authSvc, userSvc, _, sessRepo := newAuthService(t)
	ctx := context.Background()

	u := &domain.User{Username: "eve", Email: "e@e.com", PasswordHash: "x", FirstName: "E", LastName: "E"}
	created, _ := userSvc.CreateUser(ctx, u, "", nil)

	sess := &domain.Session{
		ID:           "sess_revoke_1",
		UserID:       created.ID,
		RefreshToken: "rt_revoke_1",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		CreatedAt:    time.Now(),
	}
	_ = sessRepo.Create(ctx, sess)

	if err := authSvc.RevokeToken(ctx, "sess_revoke_1"); err != nil {
		t.Fatalf("revoke: %v", err)
	}

	fresh, err := sessRepo.GetByID(ctx, "sess_revoke_1")
	if err != nil {
		t.Fatalf("get after revoke: %v", err)
	}
	if !fresh.IsRevoked {
		t.Errorf("expected IsRevoked=true after RevokeToken, got false")
	}
}
