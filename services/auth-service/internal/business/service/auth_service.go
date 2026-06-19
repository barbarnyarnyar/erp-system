package service

import (
	"context"
	"erp-system/shared/utils"
	"fmt"
	"time"

	"github.com/erp-system/auth-service/internal/business/domain"
	"github.com/erp-system/auth-service/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// TokenClaims is the canonical typed contract for decoded JWT payloads.
// Mirrors the `struct TokenClaims` declaration in `auth.cdd`:
//
//	{ user_id, tenant_id, roles }
//
// Plus internal-only fields (Username, Email, Permissions, SecurityStamp)
// used by ValidateToken and downstream consumers.
type TokenClaims struct {
	UserID        string   `json:"user_id"`
	TenantID      string   `json:"tenant_id"`
	Roles         []string `json:"roles"`
	Username      string   `json:"username"`
	Email         string   `json:"email"`
	Permissions   []string `json:"permissions"`
	SecurityStamp string   `json:"security_stamp"`
	jwt.RegisteredClaims
}

// JWTClaims is a deprecated alias preserved for backward compatibility with
// any internal callers still referencing the old name. New code should use
// TokenClaims (the typed contract declared in auth.cdd).
//
// Deprecated: use TokenClaims.
type JWTClaims = TokenClaims

type AuthService struct {
	userRepo  domain.UserRepository
	sessRepo  domain.SessionRepository
	rbacSvc   *RBACService
	publisher domain.EventPublisher
	cfg       *config.Config
}

func NewAuthService(
	userRepo domain.UserRepository,
	sessRepo domain.SessionRepository,
	rbacSvc *RBACService,
	publisher domain.EventPublisher,
	cfg *config.Config,
) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		sessRepo:  sessRepo,
		rbacSvc:   rbacSvc,
		publisher: publisher,
		cfg:       cfg,
	}
}

func (s *AuthService) AuthenticateUser(ctx context.Context, username, password, ipAddress, userAgent string) (string, string, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return "", "", fmt.Errorf("invalid credentials")
	}

	if user.Status != domain.UserStatusACTIVE {
		return "", "", fmt.Errorf("user account is deactivated")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", "", fmt.Errorf("invalid credentials")
	}

	return s.generateTokens(ctx, user, ipAddress, userAgent)
}

func (s *AuthService) generateTokens(ctx context.Context, user *domain.User, ipAddress, userAgent string) (string, string, error) {
	// Resolve Roles and Permissions via RBACService
	roles, permissions, err := s.rbacSvc.GetUserRolesAndPermissions(ctx, user.ID)
	if err != nil {
		return "", "", err
	}

	// Generate Access Token (JWT) — embeds the user's current security_stamp
	// so that any subsequent deactivation / password change / role change can
	// be detected by ValidateToken simply by reloading the user.
	claims := TokenClaims{
		UserID:        user.ID,
		TenantID:      "default", // single-tenant MVP; multi-tenant deferred
		Roles:         roles,
		Username:      user.Username,
		Email:         user.Email,
		Permissions:   permissions,
		SecurityStamp: user.SecurityStamp,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.cfg.JWT.AccessExpiry) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate Refresh Token
	refreshToken := fmt.Sprintf("rt_%s_%s", utils.NewID("rt"), user.ID)
	session := &domain.Session{
		ID:           utils.NewID("sess"),
		UserID:       user.ID,
		RefreshToken: refreshToken,
		IpAddress:    &ipAddress,
		UserAgent:    &userAgent,
		ExpiresAt:    time.Now().Add(time.Duration(s.cfg.JWT.RefreshExpiry) * time.Hour),
		CreatedAt:    time.Now(),
	}

	err = s.sessRepo.Create(ctx, session)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	session, err := s.sessRepo.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("session expired or invalid")
	}

	if session.ExpiresAt.Before(time.Now()) {
		_ = s.sessRepo.Delete(ctx, session.ID)
		return "", "", fmt.Errorf("session expired")
	}

	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil || user.Status != domain.UserStatusACTIVE {
		return "", "", fmt.Errorf("user account inactive or invalid")
	}

	// Delete old session
	_ = s.sessRepo.Delete(ctx, session.ID)

	var ip, ua string
	if session.IpAddress != nil {
		ip = *session.IpAddress
	}
	if session.UserAgent != nil {
		ua = *session.UserAgent
	}

	return s.generateTokens(ctx, user, ip, ua)
}

func (s *AuthService) RevokeToken(ctx context.Context, sessionID string) error {
	// Mark the session as revoked rather than deleting it. This preserves the
	// audit trail and lets ValidateToken reject any access token still in
	// flight for this session even before the security_stamp catches up.
	session, err := s.sessRepo.GetByID(ctx, sessionID)
	if err != nil {
		return err
	}
	session.IsRevoked = true
	return s.sessRepo.Update(ctx, session)
}

func (s *AuthService) ValidateToken(ctx context.Context, tokenStr string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWT.Secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("token invalid: %w", err)
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("token invalid")
	}

	// Reject tokens for deactivated users.
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("token invalid: user no longer exists")
	}
	if user.Status != domain.UserStatusACTIVE {
		return nil, fmt.Errorf("token invalid: user account is deactivated")
	}

	// Reject tokens whose security_stamp has been bumped since issuance
	// (deactivation, password change, role change, etc.).
	if claims.SecurityStamp != "" && user.SecurityStamp != "" && claims.SecurityStamp != user.SecurityStamp {
		return nil, fmt.Errorf("token invalid: security stamp mismatch (user state changed)")
	}

	return claims, nil
}

func (s *AuthService) GetSessionByRefreshToken(ctx context.Context, token string) (*domain.Session, error) {
	return s.sessRepo.GetByRefreshToken(ctx, token)
}
