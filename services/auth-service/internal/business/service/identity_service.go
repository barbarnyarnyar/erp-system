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

type IdentityService struct {
	userRepo   domain.UserRepository
	sessRepo   domain.SessionRepository
	roleRepo   domain.RoleRepository
	permRepo   domain.PermissionRepository
	urRepo     domain.UserRoleRepository
	usRepo     domain.UserStoreRepository
	rpRepo     domain.RolePermissionRepository
	cfg        *config.Config
}

func NewIdentityService(
	userRepo domain.UserRepository,
	sessRepo domain.SessionRepository,
	roleRepo domain.RoleRepository,
	permRepo domain.PermissionRepository,
	urRepo domain.UserRoleRepository,
	usRepo domain.UserStoreRepository,
	rpRepo domain.RolePermissionRepository,
	cfg *config.Config,
) *IdentityService {
	return &IdentityService{
		userRepo:   userRepo,
		sessRepo:   sessRepo,
		roleRepo:   roleRepo,
		permRepo:   permRepo,
		urRepo:     urRepo,
		usRepo:     usRepo,
		rpRepo:     rpRepo,
		cfg:        cfg,
	}
}

func (s *IdentityService) AuthenticateUser(ctx context.Context, username, passwordHash string) (string, string, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return "", "", fmt.Errorf("invalid credentials")
	}

	if !user.IsActive {
		return "", "", fmt.Errorf("user account is deactivated")
	}

	// Simple check (in production, use bcrypt or secure comparison)
	if user.PasswordHash != passwordHash {
		return "", "", fmt.Errorf("invalid credentials")
	}

	// Resolve Roles and Permissions
	roles, permissions, err := s.GetUserRolesAndPermissions(ctx, user.ID)
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
		ExpiresAt:    time.Now().Add(time.Duration(s.cfg.JWT.RefreshExpiry) * time.Hour),
		CreatedAt:    time.Now(),
	}

	err = s.sessRepo.Create(ctx, session)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *IdentityService) GetUserRolesAndPermissions(ctx context.Context, userID string) ([]string, []string, error) {
	urLinks, err := s.urRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	var roles []string
	var permissions []string
	seenPerms := make(map[string]bool)

	for _, ur := range urLinks {
		role, err := s.roleRepo.GetByID(ctx, ur.RoleID)
		if err == nil {
			roles = append(roles, role.Name)

			// Get permissions for this role
			rpLinks, err := s.rpRepo.ListByRoleID(ctx, role.ID)
			if err == nil {
				for _, rp := range rpLinks {
					p, err := s.permRepo.GetByID(ctx, rp.PermissionID)
					if err == nil && !seenPerms[p.Code] {
						seenPerms[p.Code] = true
						permissions = append(permissions, p.Code)
					}
				}
			}
		}
	}

	return roles, permissions, nil
}

func (s *IdentityService) CreateUser(ctx context.Context, u *domain.User, initialStoreID string, roleIDs []string) (*domain.User, error) {
	u.ID = fmt.Sprintf("user_%d", time.Now().UnixNano())
	u.IsActive = true
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()

	err := s.userRepo.Create(ctx, u)
	if err != nil {
		return nil, err
	}

	// Link to Store
	if initialStoreID != "" {
		err = s.AssignUserToStore(ctx, u.ID, initialStoreID)
		if err != nil {
			return nil, err
		}
	}

	// Link to Roles
	for _, roleID := range roleIDs {
		linkID := fmt.Sprintf("ur_%d", time.Now().UnixNano())
		ur := &domain.UserRole{
			ID:     linkID,
			UserID: u.ID,
			RoleID: roleID,
		}
		_ = s.urRepo.Create(ctx, ur)
	}

	return u, nil
}

func (s *IdentityService) UpdateUser(ctx context.Context, userID string, updatedFields map[string]interface{}) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if val, ok := updatedFields["first_name"].(string); ok {
		user.FirstName = val
	}
	if val, ok := updatedFields["last_name"].(string); ok {
		user.LastName = val
	}
	if val, ok := updatedFields["email"].(string); ok {
		user.Email = val
	}
	if val, ok := updatedFields["is_active"].(bool); ok {
		user.IsActive = val
	}
	user.UpdatedAt = time.Now()

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *IdentityService) UpdateCredentials(ctx context.Context, userID string, newPasswordHash string) (bool, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, err
	}

	user.PasswordHash = newPasswordHash
	user.UpdatedAt = time.Now()

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *IdentityService) AssignUserToStore(ctx context.Context, userID, storeID string) error {
	linkID := fmt.Sprintf("us_%d", time.Now().UnixNano())
	us := &domain.UserStore{
		ID:         linkID,
		UserID:     userID,
		StoreID:    storeID,
		AssignedAt: time.Now(),
	}
	return s.usRepo.Create(ctx, us)
}

func (s *IdentityService) ValidatePermissions(ctx context.Context, userID string, requiredPermission string) (bool, error) {
	_, permissions, err := s.GetUserRolesAndPermissions(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, perm := range permissions {
		if perm == requiredPermission {
			return true, nil
		}
	}

	return false, nil
}

func (s *IdentityService) DeactivateUser(ctx context.Context, userID string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	user.IsActive = false
	user.UpdatedAt = time.Now()
	return s.userRepo.Update(ctx, user)
}

func (s *IdentityService) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
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

	// Issue new tokens
	return s.AuthenticateUser(ctx, user.Username, user.PasswordHash)
}

func (s *IdentityService) Logout(ctx context.Context, refreshToken string) error {
	session, err := s.sessRepo.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		return err
	}
	return s.sessRepo.Delete(ctx, session.ID)
}

// Seed Helpers (For initial role/permission catalog setup)
func (s *IdentityService) CreateRole(ctx context.Context, name, description string) (*domain.Role, error) {
	role := &domain.Role{
		ID:          fmt.Sprintf("role_%d", time.Now().UnixNano()),
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err := s.roleRepo.Create(ctx, role)
	return role, err
}

func (s *IdentityService) CreatePermission(ctx context.Context, code, description string) (*domain.Permission, error) {
	perm := &domain.Permission{
		ID:          fmt.Sprintf("perm_%d", time.Now().UnixNano()),
		Code:        code,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err := s.permRepo.Create(ctx, perm)
	return perm, err
}

func (s *IdentityService) LinkRolePermission(ctx context.Context, roleID, permissionID string) error {
	link := &domain.RolePermission{
		ID:           fmt.Sprintf("rp_%d", time.Now().UnixNano()),
		RoleID:       roleID,
		PermissionID: permissionID,
	}
	return s.rpRepo.Create(ctx, link)
}
