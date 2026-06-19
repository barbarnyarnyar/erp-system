package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/erp-system/auth-service/internal/business/domain"
	"github.com/erp-system/auth-service/internal/data/memory"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Custom mock session repository to force errors
type mockSessionRepo struct {
	*memory.SessionRepository
	createErr         error
	getByIDErr        error
	getByRefreshErr   error
	updateErr         error
	deleteByUIDErr    error
	deleteErr         error
}

func (m *mockSessionRepo) Create(ctx context.Context, s *domain.Session) error {
	if m.createErr != nil {
		return m.createErr
	}
	return m.SessionRepository.Create(ctx, s)
}

func (m *mockSessionRepo) GetByID(ctx context.Context, id string) (*domain.Session, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.SessionRepository.GetByID(ctx, id)
}

func (m *mockSessionRepo) GetByRefreshToken(ctx context.Context, token string) (*domain.Session, error) {
	if m.getByRefreshErr != nil {
		return nil, m.getByRefreshErr
	}
	return m.SessionRepository.GetByRefreshToken(ctx, token)
}

func (m *mockSessionRepo) Update(ctx context.Context, s *domain.Session) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	return m.SessionRepository.Update(ctx, s)
}

type mockUserRoleRepo struct {
	*memory.UserRoleRepository
	listErr error
}

func (m *mockUserRoleRepo) ListByUserID(ctx context.Context, userID string) ([]domain.UserRole, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.UserRoleRepository.ListByUserID(ctx, userID)
}

type mockUserRepo struct {
	*memory.UserRepository
	getByIDErr error
}

func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.UserRepository.GetByID(ctx, id)
}

type dummyPublisher struct{}

func (d *dummyPublisher) Publish(ctx context.Context, topic string, key string, payload interface{}) error {
	return nil
}

func TestAuthService_AuthenticateUser_Success(t *testing.T) {
	userRepo := memory.NewUserRepository()
	sessRepo := memory.NewSessionRepository()
	roleRepo := memory.NewRoleRepository()
	permRepo := memory.NewPermissionRepository()
	urRepo := memory.NewUserRoleRepository()
	rpRepo := memory.NewRolePermissionRepository()
	usRepo := memory.NewUserStoreRepository()

	pub := &dummyPublisher{}
	rbacSvc := NewRBACService(roleRepo, permRepo, urRepo, rpRepo, pub)
	cfg := newTestConfig()
	authSvc := NewAuthService(userRepo, sessRepo, rbacSvc, pub, cfg)
	userSvc := NewUserService(userRepo, usRepo, urRepo, pub)

	ctx := context.Background()

	// Create user
	u := &domain.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "secretpassword",
		FirstName:    "Test",
		LastName:     "User",
	}
	created, err := userSvc.CreateUser(ctx, u, "store_1", nil)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	accessToken, refreshToken, err := authSvc.AuthenticateUser(ctx, "testuser", "secretpassword", "127.0.0.1", "agent")
	if err != nil {
		t.Fatalf("auth user: %v", err)
	}
	if accessToken == "" || refreshToken == "" {
		t.Error("expected non-empty tokens")
	}

	// Verify claims
	claims, err := authSvc.ValidateToken(ctx, accessToken)
	if err != nil {
		t.Fatalf("validate token: %v", err)
	}
	if claims.UserID != created.ID {
		t.Errorf("expected UserID %s, got %s", created.ID, claims.UserID)
	}
	if claims.Username != "testuser" {
		t.Errorf("expected Username testuser, got %s", claims.Username)
	}
}

func TestAuthService_AuthenticateUser_Failures(t *testing.T) {
	ctx := context.Background()

	t.Run("UserNotFound", func(t *testing.T) {
		userRepo := memory.NewUserRepository()
		sessRepo := memory.NewSessionRepository()
		roleRepo := memory.NewRoleRepository()
		permRepo := memory.NewPermissionRepository()
		urRepo := memory.NewUserRoleRepository()
		rpRepo := memory.NewRolePermissionRepository()
		pub := &dummyPublisher{}
		rbacSvc := NewRBACService(roleRepo, permRepo, urRepo, rpRepo, pub)
		authSvc := NewAuthService(userRepo, sessRepo, rbacSvc, pub, newTestConfig())

		_, _, err := authSvc.AuthenticateUser(ctx, "nonexistent", "pw", "ip", "ua")
		if err == nil || err.Error() != "invalid credentials" {
			t.Errorf("expected 'invalid credentials' error, got %v", err)
		}
	})

	t.Run("UserInactive", func(t *testing.T) {
		userRepo := memory.NewUserRepository()
		sessRepo := memory.NewSessionRepository()
		roleRepo := memory.NewRoleRepository()
		permRepo := memory.NewPermissionRepository()
		urRepo := memory.NewUserRoleRepository()
		rpRepo := memory.NewRolePermissionRepository()
		pub := &dummyPublisher{}
		rbacSvc := NewRBACService(roleRepo, permRepo, urRepo, rpRepo, pub)
		authSvc := NewAuthService(userRepo, sessRepo, rbacSvc, pub, newTestConfig())

		u := &domain.User{
			ID:           "u_1",
			Username:     "inactive",
			Status:       domain.UserStatusINACTIVE,
			PasswordHash: "hash",
		}
		_ = userRepo.Create(ctx, u)

		_, _, err := authSvc.AuthenticateUser(ctx, "inactive", "pw", "ip", "ua")
		if err == nil || err.Error() != "user account is deactivated" {
			t.Errorf("expected 'user account is deactivated' error, got %v", err)
		}
	})

	t.Run("InvalidPassword", func(t *testing.T) {
		userRepo := memory.NewUserRepository()
		sessRepo := memory.NewSessionRepository()
		roleRepo := memory.NewRoleRepository()
		permRepo := memory.NewPermissionRepository()
		urRepo := memory.NewUserRoleRepository()
		rpRepo := memory.NewRolePermissionRepository()
		pub := &dummyPublisher{}
		rbacSvc := NewRBACService(roleRepo, permRepo, urRepo, rpRepo, pub)
		authSvc := NewAuthService(userRepo, sessRepo, rbacSvc, pub, newTestConfig())

		pwdBytes, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
		u := &domain.User{
			ID:           "u_1",
			Username:     "user",
			Status:       domain.UserStatusACTIVE,
			PasswordHash: string(pwdBytes),
		}
		_ = userRepo.Create(ctx, u)

		_, _, err := authSvc.AuthenticateUser(ctx, "user", "wrong", "ip", "ua")
		if err == nil || err.Error() != "invalid credentials" {
			t.Errorf("expected 'invalid credentials' error, got %v", err)
		}
	})

	t.Run("RBACServiceError", func(t *testing.T) {
		userRepo := memory.NewUserRepository()
		sessRepo := memory.NewSessionRepository()
		roleRepo := memory.NewRoleRepository()
		permRepo := memory.NewPermissionRepository()
		urRepo := &mockUserRoleRepo{
			UserRoleRepository: memory.NewUserRoleRepository(),
			listErr:            errors.New("db error"),
		}
		rpRepo := memory.NewRolePermissionRepository()
		pub := &dummyPublisher{}
		rbacSvc := NewRBACService(roleRepo, permRepo, urRepo, rpRepo, pub)
		authSvc := NewAuthService(userRepo, sessRepo, rbacSvc, pub, newTestConfig())

		pwdBytes, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
		u := &domain.User{
			ID:           "u_1",
			Username:     "user",
			Status:       domain.UserStatusACTIVE,
			PasswordHash: string(pwdBytes),
		}
		_ = userRepo.Create(ctx, u)

		_, _, err := authSvc.AuthenticateUser(ctx, "user", "correct", "ip", "ua")
		if err == nil || err.Error() != "db error" {
			t.Errorf("expected 'db error', got %v", err)
		}
	})

	t.Run("SessionCreationError", func(t *testing.T) {
		userRepo := memory.NewUserRepository()
		sessRepo := &mockSessionRepo{
			SessionRepository: memory.NewSessionRepository(),
			createErr:         errors.New("failed to save session"),
		}
		roleRepo := memory.NewRoleRepository()
		permRepo := memory.NewPermissionRepository()
		urRepo := memory.NewUserRoleRepository()
		rpRepo := memory.NewRolePermissionRepository()
		pub := &dummyPublisher{}
		rbacSvc := NewRBACService(roleRepo, permRepo, urRepo, rpRepo, pub)
		authSvc := NewAuthService(userRepo, sessRepo, rbacSvc, pub, newTestConfig())

		pwdBytes, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
		u := &domain.User{
			ID:           "u_1",
			Username:     "user",
			Status:       domain.UserStatusACTIVE,
			PasswordHash: string(pwdBytes),
		}
		_ = userRepo.Create(ctx, u)

		_, _, err := authSvc.AuthenticateUser(ctx, "user", "correct", "ip", "ua")
		if err == nil || err.Error() != "failed to save session" {
			t.Errorf("expected 'failed to save session', got %v", err)
		}
	})
}

func TestAuthService_RefreshToken(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		userRepo := memory.NewUserRepository()
		sessRepo := memory.NewSessionRepository()
		roleRepo := memory.NewRoleRepository()
		permRepo := memory.NewPermissionRepository()
		urRepo := memory.NewUserRoleRepository()
		rpRepo := memory.NewRolePermissionRepository()
		pub := &dummyPublisher{}
		rbacSvc := NewRBACService(roleRepo, permRepo, urRepo, rpRepo, pub)
		authSvc := NewAuthService(userRepo, sessRepo, rbacSvc, pub, newTestConfig())

		u := &domain.User{
			ID:     "u_1",
			Status: domain.UserStatusACTIVE,
		}
		_ = userRepo.Create(ctx, u)

		ip := "127.0.0.1"
		ua := "agent"
		sess := &domain.Session{
			ID:           "sess_1",
			UserID:       "u_1",
			RefreshToken: "rt_123",
			IpAddress:    &ip,
			UserAgent:    &ua,
			ExpiresAt:    time.Now().Add(1 * time.Hour),
		}
		_ = sessRepo.Create(ctx, sess)

		accessToken, refreshToken, err := authSvc.RefreshToken(ctx, "rt_123")
		if err != nil {
			t.Fatalf("refresh failed: %v", err)
		}
		if accessToken == "" || refreshToken == "" {
			t.Error("expected non-empty tokens")
		}

		// Verify old session deleted
		_, err = sessRepo.GetByID(ctx, "sess_1")
		if err == nil {
			t.Error("expected old session to be deleted")
		}
	})

	t.Run("SessionNotFound", func(t *testing.T) {
		userRepo := memory.NewUserRepository()
		sessRepo := memory.NewSessionRepository()
		roleRepo := memory.NewRoleRepository()
		permRepo := memory.NewPermissionRepository()
		urRepo := memory.NewUserRoleRepository()
		rpRepo := memory.NewRolePermissionRepository()
		pub := &dummyPublisher{}
		rbacSvc := NewRBACService(roleRepo, permRepo, urRepo, rpRepo, pub)
		authSvc := NewAuthService(userRepo, sessRepo, rbacSvc, pub, newTestConfig())

		_, _, err := authSvc.RefreshToken(ctx, "nonexistent")
		if err == nil || err.Error() != "session expired or invalid" {
			t.Errorf("expected 'session expired or invalid', got %v", err)
		}
	})

	t.Run("SessionExpired", func(t *testing.T) {
		userRepo := memory.NewUserRepository()
		sessRepo := memory.NewSessionRepository()
		roleRepo := memory.NewRoleRepository()
		permRepo := memory.NewPermissionRepository()
		urRepo := memory.NewUserRoleRepository()
		rpRepo := memory.NewRolePermissionRepository()
		pub := &dummyPublisher{}
		rbacSvc := NewRBACService(roleRepo, permRepo, urRepo, rpRepo, pub)
		authSvc := NewAuthService(userRepo, sessRepo, rbacSvc, pub, newTestConfig())

		sess := &domain.Session{
			ID:           "sess_1",
			UserID:       "u_1",
			RefreshToken: "rt_expired",
			ExpiresAt:    time.Now().Add(-1 * time.Hour),
		}
		_ = sessRepo.Create(ctx, sess)

		_, _, err := authSvc.RefreshToken(ctx, "rt_expired")
		if err == nil || err.Error() != "session expired" {
			t.Errorf("expected 'session expired', got %v", err)
		}

		// Verify session was deleted from repo
		_, err = sessRepo.GetByID(ctx, "sess_1")
		if err == nil {
			t.Error("expected expired session to be deleted")
		}
	})

	t.Run("UserInactiveOrInvalid", func(t *testing.T) {
		userRepo := memory.NewUserRepository()
		sessRepo := memory.NewSessionRepository()
		roleRepo := memory.NewRoleRepository()
		permRepo := memory.NewPermissionRepository()
		urRepo := memory.NewUserRoleRepository()
		rpRepo := memory.NewRolePermissionRepository()
		pub := &dummyPublisher{}
		rbacSvc := NewRBACService(roleRepo, permRepo, urRepo, rpRepo, pub)
		authSvc := NewAuthService(userRepo, sessRepo, rbacSvc, pub, newTestConfig())

		// Case 1: User does not exist
		sess1 := &domain.Session{
			ID:           "sess_1",
			UserID:       "u_nonexistent",
			RefreshToken: "rt_1",
			ExpiresAt:    time.Now().Add(1 * time.Hour),
		}
		_ = sessRepo.Create(ctx, sess1)
		_, _, err := authSvc.RefreshToken(ctx, "rt_1")
		if err == nil || err.Error() != "user account inactive or invalid" {
			t.Errorf("expected error for nonexistent user, got %v", err)
		}

		// Case 2: User deactivated
		u := &domain.User{
			ID:     "u_inactive",
			Status: domain.UserStatusINACTIVE,
		}
		_ = userRepo.Create(ctx, u)

		sess2 := &domain.Session{
			ID:           "sess_2",
			UserID:       "u_inactive",
			RefreshToken: "rt_2",
			ExpiresAt:    time.Now().Add(1 * time.Hour),
		}
		_ = sessRepo.Create(ctx, sess2)
		_, _, err = authSvc.RefreshToken(ctx, "rt_2")
		if err == nil || err.Error() != "user account inactive or invalid" {
			t.Errorf("expected error for inactive user, got %v", err)
		}
	})
}

func TestAuthService_RevokeToken(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		userRepo := memory.NewUserRepository()
		sessRepo := memory.NewSessionRepository()
		roleRepo := memory.NewRoleRepository()
		permRepo := memory.NewPermissionRepository()
		urRepo := memory.NewUserRoleRepository()
		rpRepo := memory.NewRolePermissionRepository()
		pub := &dummyPublisher{}
		rbacSvc := NewRBACService(roleRepo, permRepo, urRepo, rpRepo, pub)
		authSvc := NewAuthService(userRepo, sessRepo, rbacSvc, pub, newTestConfig())

		sess := &domain.Session{
			ID:           "sess_1",
			UserID:       "u_1",
			RefreshToken: "rt_1",
			IsRevoked:    false,
		}
		_ = sessRepo.Create(ctx, sess)

		err := authSvc.RevokeToken(ctx, "sess_1")
		if err != nil {
			t.Fatalf("expected nil err, got %v", err)
		}

		s, _ := sessRepo.GetByID(ctx, "sess_1")
		if !s.IsRevoked {
			t.Error("expected session to be marked as revoked")
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		userRepo := memory.NewUserRepository()
		sessRepo := memory.NewSessionRepository()
		roleRepo := memory.NewRoleRepository()
		permRepo := memory.NewPermissionRepository()
		urRepo := memory.NewUserRoleRepository()
		rpRepo := memory.NewRolePermissionRepository()
		pub := &dummyPublisher{}
		rbacSvc := NewRBACService(roleRepo, permRepo, urRepo, rpRepo, pub)
		authSvc := NewAuthService(userRepo, sessRepo, rbacSvc, pub, newTestConfig())

		err := authSvc.RevokeToken(ctx, "nonexistent")
		if err == nil {
			t.Error("expected error for nonexistent session")
		}
	})
}

func TestAuthService_ValidateToken_Failures(t *testing.T) {
	ctx := context.Background()
	cfg := newTestConfig()

	t.Run("InvalidSignatureMethod", func(t *testing.T) {
		userRepo := memory.NewUserRepository()
		sessRepo := memory.NewSessionRepository()
		roleRepo := memory.NewRoleRepository()
		permRepo := memory.NewPermissionRepository()
		urRepo := memory.NewUserRoleRepository()
		rpRepo := memory.NewRolePermissionRepository()
		pub := &dummyPublisher{}
		rbacSvc := NewRBACService(roleRepo, permRepo, urRepo, rpRepo, pub)
		authSvc := NewAuthService(userRepo, sessRepo, rbacSvc, pub, cfg)

		// Create a token with 'none' signing method
		token := jwt.NewWithClaims(jwt.SigningMethodNone, TokenClaims{
			UserID: "u_1",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			},
		})
		tokenStr, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
		if err != nil {
			t.Fatalf("failed to create none signed token: %v", err)
		}

		_, err = authSvc.ValidateToken(ctx, tokenStr)
		if err == nil {
			t.Error("expected error validating token with none signature method, got nil")
		}
	})

	t.Run("MalformedToken", func(t *testing.T) {
		userRepo := memory.NewUserRepository()
		sessRepo := memory.NewSessionRepository()
		roleRepo := memory.NewRoleRepository()
		permRepo := memory.NewPermissionRepository()
		urRepo := memory.NewUserRoleRepository()
		rpRepo := memory.NewRolePermissionRepository()
		pub := &dummyPublisher{}
		rbacSvc := NewRBACService(roleRepo, permRepo, urRepo, rpRepo, pub)
		authSvc := NewAuthService(userRepo, sessRepo, rbacSvc, pub, cfg)

		_, err := authSvc.ValidateToken(ctx, "not-a-token")
		if err == nil {
			t.Error("expected error for malformed token")
		}
	})

	t.Run("UserDeleted", func(t *testing.T) {
		userRepo := &mockUserRepo{
			UserRepository: memory.NewUserRepository(),
			getByIDErr:     errors.New("db error"),
		}
		sessRepo := memory.NewSessionRepository()
		roleRepo := memory.NewRoleRepository()
		permRepo := memory.NewPermissionRepository()
		urRepo := memory.NewUserRoleRepository()
		rpRepo := memory.NewRolePermissionRepository()
		pub := &dummyPublisher{}
		rbacSvc := NewRBACService(roleRepo, permRepo, urRepo, rpRepo, pub)
		authSvc := NewAuthService(userRepo, sessRepo, rbacSvc, pub, cfg)

		claims := TokenClaims{
			UserID: "u_1",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, _ := token.SignedString([]byte(cfg.JWT.Secret))

		_, err := authSvc.ValidateToken(ctx, tokenStr)
		if err == nil || err.Error() != "token invalid: user no longer exists" {
			t.Errorf("expected 'token invalid: user no longer exists', got %v", err)
		}
	})

	t.Run("UserDeactivated", func(t *testing.T) {
		userRepo := memory.NewUserRepository()
		sessRepo := memory.NewSessionRepository()
		roleRepo := memory.NewRoleRepository()
		permRepo := memory.NewPermissionRepository()
		urRepo := memory.NewUserRoleRepository()
		rpRepo := memory.NewRolePermissionRepository()
		pub := &dummyPublisher{}
		rbacSvc := NewRBACService(roleRepo, permRepo, urRepo, rpRepo, pub)
		authSvc := NewAuthService(userRepo, sessRepo, rbacSvc, pub, cfg)

		u := &domain.User{
			ID:     "u_deactivated",
			Status: domain.UserStatusINACTIVE,
		}
		_ = userRepo.Create(ctx, u)

		claims := TokenClaims{
			UserID: "u_deactivated",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, _ := token.SignedString([]byte(cfg.JWT.Secret))

		_, err := authSvc.ValidateToken(ctx, tokenStr)
		if err == nil || err.Error() != "token invalid: user account is deactivated" {
			t.Errorf("expected 'token invalid: user account is deactivated', got %v", err)
		}
	})
}

func TestAuthService_GetSessionByRefreshToken(t *testing.T) {
	userRepo := memory.NewUserRepository()
	sessRepo := memory.NewSessionRepository()
	roleRepo := memory.NewRoleRepository()
	permRepo := memory.NewPermissionRepository()
	urRepo := memory.NewUserRoleRepository()
	rpRepo := memory.NewRolePermissionRepository()
	pub := &dummyPublisher{}
	rbacSvc := NewRBACService(roleRepo, permRepo, urRepo, rpRepo, pub)
	authSvc := NewAuthService(userRepo, sessRepo, rbacSvc, pub, newTestConfig())

	ctx := context.Background()
	sess := &domain.Session{
		ID:           "sess_1",
		UserID:       "u_1",
		RefreshToken: "rt_1",
	}
	_ = sessRepo.Create(ctx, sess)

	got, err := authSvc.GetSessionByRefreshToken(ctx, "rt_1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got.ID != "sess_1" {
		t.Errorf("expected sess_1, got %s", got.ID)
	}
}
