package service

import (
	"context"
	"fmt"
	"time"

	"github.com/erp-system/auth-service/internal/business/domain"
	"github.com/erp-system/auth-service/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

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

	if !user.IsActive {
		return "", "", fmt.Errorf("user account is deactivated")
	}

	// Simple check (in production, use bcrypt or secure comparison)
	if user.PasswordHash != password {
		return "", "", fmt.Errorf("invalid credentials")
	}

	// Resolve Roles and Permissions via RBACService
	roles, permissions, err := s.rbacSvc.GetUserRolesAndPermissions(ctx, user.ID)
	if err != nil {
		return "", "", err
	}

	// Generate Access Token (JWT)
	claims := JWTClaims{
		UserID:      user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Roles:       roles,
		Permissions: permissions,
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
	refreshToken := fmt.Sprintf("rt_%d_%s", time.Now().UnixNano(), user.ID)
	session := &domain.Session{
		ID:           fmt.Sprintf("sess_%d", time.Now().UnixNano()),
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
	if err != nil || !user.IsActive {
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

	// Issue new tokens
	return s.AuthenticateUser(ctx, user.Username, user.PasswordHash, ip, ua)
}

func (s *AuthService) RevokeToken(ctx context.Context, sessionID string) error {
	return s.sessRepo.Delete(ctx, sessionID)
}

func (s *AuthService) ValidateToken(ctx context.Context, tokenStr string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWT.Secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("token invalid: %w", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("token invalid")
}

func (s *AuthService) GetSessionByRefreshToken(ctx context.Context, token string) (*domain.Session, error) {
	return s.sessRepo.GetByRefreshToken(ctx, token)
}
