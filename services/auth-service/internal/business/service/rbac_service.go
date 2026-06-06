package service

import (
	"context"
	"fmt"
	"time"

	"github.com/erp-system/auth-service/internal/business/domain"
)

type RBACService struct {
	roleRepo  domain.RoleRepository
	permRepo  domain.PermissionRepository
	urRepo    domain.UserRoleRepository
	rpRepo    domain.RolePermissionRepository
	publisher domain.EventPublisher
}

func NewRBACService(
	roleRepo domain.RoleRepository,
	permRepo domain.PermissionRepository,
	urRepo domain.UserRoleRepository,
	rpRepo domain.RolePermissionRepository,
	publisher domain.EventPublisher,
) *RBACService {
	return &RBACService{
		roleRepo:  roleRepo,
		permRepo:  permRepo,
		urRepo:    urRepo,
		rpRepo:    rpRepo,
		publisher: publisher,
	}
}

func (s *RBACService) GetUserRolesAndPermissions(ctx context.Context, userID string) ([]string, []string, error) {
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

func (s *RBACService) CreateRole(ctx context.Context, name, description string) (*domain.Role, error) {
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

func (s *RBACService) CreatePermission(ctx context.Context, code, description string) (*domain.Permission, error) {
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

func (s *RBACService) AssignPermissionToRole(ctx context.Context, roleID, permissionID string) error {
	link := &domain.RolePermission{
		ID:           fmt.Sprintf("rp_%d", time.Now().UnixNano()),
		RoleID:       roleID,
		PermissionID: permissionID,
		CreatedAt:    time.Now(),
	}
	return s.rpRepo.Create(ctx, link)
}

func (s *RBACService) ValidatePermissions(ctx context.Context, userID string, requiredPermission string) (bool, error) {
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

func (s *RBACService) ListRoles(ctx context.Context) ([]domain.Role, error) {
	return s.roleRepo.List(ctx)
}

func (s *RBACService) ListPermissions(ctx context.Context) ([]domain.Permission, error) {
	return s.permRepo.List(ctx)
}

func (s *RBACService) GetRolePermissions(ctx context.Context, roleID string) ([]domain.Permission, error) {
	rpLinks, err := s.rpRepo.ListByRoleID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	var list []domain.Permission
	for _, link := range rpLinks {
		p, err := s.permRepo.GetByID(ctx, link.PermissionID)
		if err == nil {
			list = append(list, *p)
		}
	}
	return list, nil
}

func (s *RBACService) RemovePermissionFromRole(ctx context.Context, roleID string, permissionID string) error {
	return s.rpRepo.Delete(ctx, roleID, permissionID)
}

func (s *RBACService) DeleteRole(ctx context.Context, id string) error {
	return s.roleRepo.Delete(ctx, id)
}

func (s *RBACService) DeletePermission(ctx context.Context, id string) error {
	return s.permRepo.Delete(ctx, id)
}
