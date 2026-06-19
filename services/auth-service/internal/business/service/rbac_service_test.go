package service

import (
	"context"
	"errors"
	"testing"

	"github.com/erp-system/auth-service/internal/business/domain"
	"github.com/erp-system/auth-service/internal/data/memory"
)

type mockRoleRepo struct {
	*memory.RoleRepository
	getIDErr error
	listErr  error
}

func (m *mockRoleRepo) GetByID(ctx context.Context, id string) (*domain.Role, error) {
	if m.getIDErr != nil {
		return nil, m.getIDErr
	}
	return m.RoleRepository.GetByID(ctx, id)
}

func (m *mockRoleRepo) List(ctx context.Context) ([]domain.Role, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.RoleRepository.List(ctx)
}

type mockPermissionRepo struct {
	*memory.PermissionRepository
	getIDErr error
	listErr  error
}

func (m *mockPermissionRepo) GetByID(ctx context.Context, id string) (*domain.Permission, error) {
	if m.getIDErr != nil {
		return nil, m.getIDErr
	}
	return m.PermissionRepository.GetByID(ctx, id)
}

func (m *mockPermissionRepo) List(ctx context.Context) ([]domain.Permission, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.PermissionRepository.List(ctx)
}

type mockRolePermissionRepo struct {
	*memory.RolePermissionRepository
	listErr error
}

func (m *mockRolePermissionRepo) ListByRoleID(ctx context.Context, roleID string) ([]domain.RolePermission, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.RolePermissionRepository.ListByRoleID(ctx, roleID)
}

func TestRBACService_GetUserRolesAndPermissions_SuccessAndEdgeCases(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		roleRepo := memory.NewRoleRepository()
		permRepo := memory.NewPermissionRepository()
		urRepo := memory.NewUserRoleRepository()
		rpRepo := memory.NewRolePermissionRepository()
		pub := &dummyPublisher{}

		s := NewRBACService(roleRepo, permRepo, urRepo, rpRepo, pub)

		role, _ := s.CreateRole(ctx, "Admin", "Admin Role")
		perm, _ := s.CreatePermission(ctx, "users.create", "Create users")
		_ = s.AssignPermissionToRole(ctx, role.ID, perm.ID)

		// Link user to role
		_ = urRepo.Create(ctx, &domain.UserRole{
			ID:     "ur_1",
			UserID: "u_1",
			RoleID: role.ID,
		})

		roles, permissions, err := s.GetUserRolesAndPermissions(ctx, "u_1")
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if len(roles) != 1 || roles[0] != "Admin" {
			t.Errorf("expected roles [Admin], got %v", roles)
		}
		if len(permissions) != 1 || permissions[0] != "users.create" {
			t.Errorf("expected permissions [users.create], got %v", permissions)
		}
	})

	t.Run("UserRoleRepoError", func(t *testing.T) {
		roleRepo := memory.NewRoleRepository()
		permRepo := memory.NewPermissionRepository()
		urRepo := &mockUserRoleRepo{
			UserRoleRepository: memory.NewUserRoleRepository(),
			listErr:            errors.New("db error"),
		}
		rpRepo := memory.NewRolePermissionRepository()
		pub := &dummyPublisher{}

		s := NewRBACService(roleRepo, permRepo, urRepo, rpRepo, pub)
		_, _, err := s.GetUserRolesAndPermissions(ctx, "u_1")
		if err == nil || err.Error() != "db error" {
			t.Errorf("expected 'db error', got %v", err)
		}
	})

	t.Run("RoleRepoErrorIsIgnored", func(t *testing.T) {
		roleRepo := &mockRoleRepo{
			RoleRepository: memory.NewRoleRepository(),
			getIDErr:       errors.New("role not found"),
		}
		permRepo := memory.NewPermissionRepository()
		urRepo := memory.NewUserRoleRepository()
		rpRepo := memory.NewRolePermissionRepository()
		pub := &dummyPublisher{}

		s := NewRBACService(roleRepo, permRepo, urRepo, rpRepo, pub)

		// Link user to role
		_ = urRepo.Create(ctx, &domain.UserRole{
			ID:     "ur_1",
			UserID: "u_1",
			RoleID: "r_1",
		})

		roles, permissions, err := s.GetUserRolesAndPermissions(ctx, "u_1")
		if err != nil {
			t.Fatalf("expected nil err, got %v", err)
		}
		if len(roles) != 0 {
			t.Errorf("expected 0 roles, got %v", roles)
		}
		if len(permissions) != 0 {
			t.Errorf("expected 0 permissions, got %v", permissions)
		}
	})

	t.Run("RolePermissionRepoErrorIsIgnored", func(t *testing.T) {
		roleRepo := memory.NewRoleRepository()
		permRepo := memory.NewPermissionRepository()
		urRepo := memory.NewUserRoleRepository()
		rpRepo := &mockRolePermissionRepo{
			RolePermissionRepository: memory.NewRolePermissionRepository(),
			listErr:                  errors.New("db error"),
		}
		pub := &dummyPublisher{}

		s := NewRBACService(roleRepo, permRepo, urRepo, rpRepo, pub)

		role, _ := s.CreateRole(ctx, "Admin", "Admin Role")

		// Link user to role
		_ = urRepo.Create(ctx, &domain.UserRole{
			ID:     "ur_1",
			UserID: "u_1",
			RoleID: role.ID,
		})

		roles, permissions, err := s.GetUserRolesAndPermissions(ctx, "u_1")
		if err != nil {
			t.Fatalf("expected nil err, got %v", err)
		}
		if len(roles) != 1 || roles[0] != "Admin" {
			t.Errorf("expected roles [Admin], got %v", roles)
		}
		if len(permissions) != 0 {
			t.Errorf("expected 0 permissions, got %v", permissions)
		}
	})

	t.Run("PermissionRepoErrorIsIgnored", func(t *testing.T) {
		roleRepo := memory.NewRoleRepository()
		permRepo := &mockPermissionRepo{
			PermissionRepository: memory.NewPermissionRepository(),
			getIDErr:             errors.New("perm not found"),
		}
		urRepo := memory.NewUserRoleRepository()
		rpRepo := memory.NewRolePermissionRepository()
		pub := &dummyPublisher{}

		s := NewRBACService(roleRepo, permRepo, urRepo, rpRepo, pub)

		role, _ := s.CreateRole(ctx, "Admin", "Admin Role")
		_ = s.AssignPermissionToRole(ctx, role.ID, "perm_1")

		// Link user to role
		_ = urRepo.Create(ctx, &domain.UserRole{
			ID:     "ur_1",
			UserID: "u_1",
			RoleID: role.ID,
		})

		roles, permissions, err := s.GetUserRolesAndPermissions(ctx, "u_1")
		if err != nil {
			t.Fatalf("expected nil err, got %v", err)
		}
		if len(roles) != 1 || roles[0] != "Admin" {
			t.Errorf("expected roles [Admin], got %v", roles)
		}
		if len(permissions) != 0 {
			t.Errorf("expected 0 permissions, got %v", permissions)
		}
	})
}

func TestRBACService_ValidatePermissions(t *testing.T) {
	ctx := context.Background()

	t.Run("HasPermission", func(t *testing.T) {
		roleRepo := memory.NewRoleRepository()
		permRepo := memory.NewPermissionRepository()
		urRepo := memory.NewUserRoleRepository()
		rpRepo := memory.NewRolePermissionRepository()
		pub := &dummyPublisher{}

		s := NewRBACService(roleRepo, permRepo, urRepo, rpRepo, pub)

		role, _ := s.CreateRole(ctx, "Admin", "Admin Role")
		perm, _ := s.CreatePermission(ctx, "users.create", "Create users")
		_ = s.AssignPermissionToRole(ctx, role.ID, perm.ID)

		// Link user to role
		_ = urRepo.Create(ctx, &domain.UserRole{
			ID:     "ur_1",
			UserID: "u_1",
			RoleID: role.ID,
		})

		ok, err := s.ValidatePermissions(ctx, "u_1", "users.create")
		if err != nil {
			t.Fatalf("expected nil err, got %v", err)
		}
		if !ok {
			t.Error("expected ValidatePermissions to return true")
		}
	})

	t.Run("DoesNotHavePermission", func(t *testing.T) {
		roleRepo := memory.NewRoleRepository()
		permRepo := memory.NewPermissionRepository()
		urRepo := memory.NewUserRoleRepository()
		rpRepo := memory.NewRolePermissionRepository()
		pub := &dummyPublisher{}

		s := NewRBACService(roleRepo, permRepo, urRepo, rpRepo, pub)

		ok, err := s.ValidatePermissions(ctx, "u_1", "users.create")
		if err != nil {
			t.Fatalf("expected nil err, got %v", err)
		}
		if ok {
			t.Error("expected ValidatePermissions to return false")
		}
	})

	t.Run("ErrorPath", func(t *testing.T) {
		roleRepo := memory.NewRoleRepository()
		permRepo := memory.NewPermissionRepository()
		urRepo := &mockUserRoleRepo{
			UserRoleRepository: memory.NewUserRoleRepository(),
			listErr:            errors.New("db error"),
		}
		rpRepo := memory.NewRolePermissionRepository()
		pub := &dummyPublisher{}

		s := NewRBACService(roleRepo, permRepo, urRepo, rpRepo, pub)

		ok, err := s.ValidatePermissions(ctx, "u_1", "users.create")
		if err == nil || err.Error() != "db error" {
			t.Errorf("expected 'db error', got %v", err)
		}
		if ok {
			t.Error("expected false on error")
		}
	})
}

func TestRBACService_ListsAndDeletes(t *testing.T) {
	ctx := context.Background()
	roleRepo := memory.NewRoleRepository()
	permRepo := memory.NewPermissionRepository()
	urRepo := memory.NewUserRoleRepository()
	rpRepo := memory.NewRolePermissionRepository()
	pub := &dummyPublisher{}

	s := NewRBACService(roleRepo, permRepo, urRepo, rpRepo, pub)

	role, _ := s.CreateRole(ctx, "Role1", "Desc1")
	perm, _ := s.CreatePermission(ctx, "Perm1", "Desc1")

	t.Run("ListRoles", func(t *testing.T) {
		list, err := s.ListRoles(ctx)
		if err != nil {
			t.Fatalf("list roles: %v", err)
		}
		if len(list) != 1 || list[0].Name != "Role1" {
			t.Errorf("expected list with Role1, got %v", list)
		}
	})

	t.Run("ListPermissions", func(t *testing.T) {
		list, err := s.ListPermissions(ctx)
		if err != nil {
			t.Fatalf("list permissions: %v", err)
		}
		if len(list) != 1 || list[0].Code != "Perm1" {
			t.Errorf("expected list with Perm1, got %v", list)
		}
	})

	t.Run("GetRolePermissions_Success", func(t *testing.T) {
		_ = s.AssignPermissionToRole(ctx, role.ID, perm.ID)
		perms, err := s.GetRolePermissions(ctx, role.ID)
		if err != nil {
			t.Fatalf("get role permissions: %v", err)
		}
		if len(perms) != 1 || perms[0].Code != "Perm1" {
			t.Errorf("expected Perm1, got %v", perms)
		}
	})

	t.Run("GetRolePermissions_RepoError", func(t *testing.T) {
		rpRepoMock := &mockRolePermissionRepo{
			RolePermissionRepository: memory.NewRolePermissionRepository(),
			listErr:                  errors.New("db error"),
		}
		sMock := NewRBACService(roleRepo, permRepo, urRepo, rpRepoMock, pub)
		_, err := sMock.GetRolePermissions(ctx, "role_id")
		if err == nil || err.Error() != "db error" {
			t.Errorf("expected 'db error', got %v", err)
		}
	})

	t.Run("GetRolePermissions_GetByIDErrorIsIgnored", func(t *testing.T) {
		permRepoMock := &mockPermissionRepo{
			PermissionRepository: memory.NewPermissionRepository(),
			getIDErr:             errors.New("perm not found"),
		}
		sMock := NewRBACService(roleRepo, permRepoMock, urRepo, rpRepo, pub)
		perms, err := sMock.GetRolePermissions(ctx, role.ID)
		if err != nil {
			t.Fatalf("expected nil err, got %v", err)
		}
		if len(perms) != 0 {
			t.Errorf("expected 0 permissions due to getByID error, got %d", len(perms))
		}
	})

	t.Run("RemovePermissionFromRole", func(t *testing.T) {
		err := s.RemovePermissionFromRole(ctx, role.ID, perm.ID)
		if err != nil {
			t.Fatalf("remove perm from role: %v", err)
		}
		perms, _ := s.GetRolePermissions(ctx, role.ID)
		if len(perms) != 0 {
			t.Errorf("expected 0 permissions, got %d", len(perms))
		}
	})

	t.Run("DeleteRole", func(t *testing.T) {
		err := s.DeleteRole(ctx, role.ID)
		if err != nil {
			t.Fatalf("delete role: %v", err)
		}
		list, _ := s.ListRoles(ctx)
		if len(list) != 0 {
			t.Errorf("expected 0 roles, got %d", len(list))
		}
	})

	t.Run("DeletePermission", func(t *testing.T) {
		err := s.DeletePermission(ctx, perm.ID)
		if err != nil {
			t.Fatalf("delete permission: %v", err)
		}
		list, _ := s.ListPermissions(ctx)
		if len(list) != 0 {
			t.Errorf("expected 0 permissions, got %d", len(list))
		}
	})
}
