package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/erp-system/auth-service/internal/business/domain"
	"github.com/erp-system/auth-service/internal/data/memory"
	sharedtesting "erp-system/shared/testing"
)

type mockUserStoreRepo struct {
	*memory.UserStoreRepository
	createErr error
}

func (m *mockUserStoreRepo) Create(ctx context.Context, us *domain.UserStore) error {
	if m.createErr != nil {
		return m.createErr
	}
	return m.UserStoreRepository.Create(ctx, us)
}

type fullMockUserRepo struct {
	*memory.UserRepository
	createErr error
	updateErr error
	getIDErr  error
}

func (m *fullMockUserRepo) Create(ctx context.Context, u *domain.User) error {
	if m.createErr != nil {
		return m.createErr
	}
	return m.UserRepository.Create(ctx, u)
}

func (m *fullMockUserRepo) Update(ctx context.Context, u *domain.User) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	return m.UserRepository.Update(ctx, u)
}

func (m *fullMockUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	if m.getIDErr != nil {
		return nil, m.getIDErr
	}
	return m.UserRepository.GetByID(ctx, id)
}

func TestUserService_CreateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		userRepo := memory.NewUserRepository()
		usRepo := memory.NewUserStoreRepository()
		urRepo := memory.NewUserRoleRepository()
		pub := &sharedtesting.MockPublisher{}

		s := NewUserService(userRepo, usRepo, urRepo, pub)

		u := &domain.User{
			Username:     "john",
			Email:        "john@example.com",
			PasswordHash: "password123",
		}

		created, err := s.CreateUser(ctx, u, "store_1", []string{"role_1", "role_2"})
		if err != nil {
			t.Fatalf("create user: %v", err)
		}
		if created.ID == "" {
			t.Error("expected ID to be generated")
		}
		if created.SecurityStamp == "" {
			t.Error("expected SecurityStamp to be generated")
		}

		// Verify event publications
		if len(pub.Events) != 4 {
			t.Errorf("expected 4 events published (1 store assignment, 2 role assignments, 1 user created), got %d", len(pub.Events))
		}
	})

	t.Run("BcryptPasswordTooLongError", func(t *testing.T) {
		userRepo := memory.NewUserRepository()
		usRepo := memory.NewUserStoreRepository()
		urRepo := memory.NewUserRoleRepository()
		pub := &sharedtesting.MockPublisher{}

		s := NewUserService(userRepo, usRepo, urRepo, pub)

		// bcrypt password length limit is 72 bytes
		longPassword := strings.Repeat("a", 100)
		u := &domain.User{
			Username:     "john",
			Email:        "john@example.com",
			PasswordHash: longPassword,
		}

		_, err := s.CreateUser(ctx, u, "", nil)
		if err == nil || !strings.Contains(err.Error(), "password length exceeds 72 bytes") {
			t.Errorf("expected bcrypt password too long error, got %v", err)
		}
	})

	t.Run("UserRepoCreateError", func(t *testing.T) {
		userRepo := &fullMockUserRepo{
			UserRepository: memory.NewUserRepository(),
			createErr:      errors.New("db create error"),
		}
		usRepo := memory.NewUserStoreRepository()
		urRepo := memory.NewUserRoleRepository()
		pub := &sharedtesting.MockPublisher{}

		s := NewUserService(userRepo, usRepo, urRepo, pub)

		u := &domain.User{
			Username:     "john",
			Email:        "john@example.com",
			PasswordHash: "pw",
		}

		_, err := s.CreateUser(ctx, u, "", nil)
		if err == nil || err.Error() != "db create error" {
			t.Errorf("expected db create error, got %v", err)
		}
	})

	t.Run("AssignUserToStoreError", func(t *testing.T) {
		userRepo := memory.NewUserRepository()
		usRepo := &mockUserStoreRepo{
			UserStoreRepository: memory.NewUserStoreRepository(),
			createErr:           errors.New("db store error"),
		}
		urRepo := memory.NewUserRoleRepository()
		pub := &sharedtesting.MockPublisher{}

		s := NewUserService(userRepo, usRepo, urRepo, pub)

		u := &domain.User{
			Username:     "john",
			Email:        "john@example.com",
			PasswordHash: "pw",
		}

		_, err := s.CreateUser(ctx, u, "store_1", nil)
		if err == nil || err.Error() != "db store error" {
			t.Errorf("expected db store error, got %v", err)
		}
	})

	t.Run("PublishErrorsAreLoggedAndIgnored", func(t *testing.T) {
		userRepo := memory.NewUserRepository()
		usRepo := memory.NewUserStoreRepository()
		urRepo := memory.NewUserRoleRepository()
		pub := &sharedtesting.MockPublisher{
			FailPublish: true,
		}

		s := NewUserService(userRepo, usRepo, urRepo, pub)

		u := &domain.User{
			Username:     "john",
			Email:        "john@example.com",
			PasswordHash: "pw",
		}

		_, err := s.CreateUser(ctx, u, "store_1", []string{"role_1"})
		if err != nil {
			t.Fatalf("expected create user to succeed despite publish error, got %v", err)
		}
	})
}

func TestUserService_UpdateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		userRepo := memory.NewUserRepository()
		usRepo := memory.NewUserStoreRepository()
		urRepo := memory.NewUserRoleRepository()
		pub := &sharedtesting.MockPublisher{}

		s := NewUserService(userRepo, usRepo, urRepo, pub)

		u := &domain.User{
			Username:     "john",
			Email:        "john@example.com",
			PasswordHash: "pw",
			FirstName:    "John",
			LastName:     "Doe",
		}
		created, _ := s.CreateUser(ctx, u, "", nil)
		originalStamp := created.SecurityStamp

		firstName := "Johnny"
		lastName := "Smith"
		email := "johnny@example.com"
		isActive := false

		updated, err := s.UpdateUser(ctx, created.ID, &firstName, &lastName, &email, &isActive)
		if err != nil {
			t.Fatalf("update user: %v", err)
		}
		if updated.FirstName != "Johnny" || updated.LastName != "Smith" || updated.Email != "johnny@example.com" {
			t.Error("fields were not updated correctly")
		}
		if updated.Status != domain.UserStatusINACTIVE {
			t.Error("expected status to be inactive")
		}
		if updated.SecurityStamp == originalStamp {
			t.Error("expected security stamp to be bumped after status change")
		}

		// Update to same status - stamp should NOT bump
		secondStamp := updated.SecurityStamp
		updated, _ = s.UpdateUser(ctx, created.ID, nil, nil, nil, &isActive)
		if updated.SecurityStamp != secondStamp {
			t.Error("security stamp bumped when status did not change")
		}
	})

	t.Run("UserNotFound", func(t *testing.T) {
		userRepo := memory.NewUserRepository()
		usRepo := memory.NewUserStoreRepository()
		urRepo := memory.NewUserRoleRepository()
		pub := &sharedtesting.MockPublisher{}

		s := NewUserService(userRepo, usRepo, urRepo, pub)

		_, err := s.UpdateUser(ctx, "nonexistent", nil, nil, nil, nil)
		if err == nil {
			t.Error("expected error for nonexistent user")
		}
	})

	t.Run("UpdateRepoError", func(t *testing.T) {
		userRepoMock := &fullMockUserRepo{
			UserRepository: memory.NewUserRepository(),
		}
		usRepo := memory.NewUserStoreRepository()
		urRepo := memory.NewUserRoleRepository()
		pub := &sharedtesting.MockPublisher{}

		s := NewUserService(userRepoMock, usRepo, urRepo, pub)

		u := &domain.User{
			Username:     "john",
			Email:        "john@example.com",
			PasswordHash: "pw",
		}
		// Create bypassing mock error
		created, _ := s.CreateUser(ctx, u, "", nil)

		userRepoMock.updateErr = errors.New("db update error")
		_, err := s.UpdateUser(ctx, created.ID, nil, nil, nil, nil)
		if err == nil || err.Error() != "db update error" {
			t.Errorf("expected db update error, got %v", err)
		}
	})
}

func TestUserService_UpdateCredentials(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		userRepo := memory.NewUserRepository()
		usRepo := memory.NewUserStoreRepository()
		urRepo := memory.NewUserRoleRepository()
		pub := &sharedtesting.MockPublisher{}

		s := NewUserService(userRepo, usRepo, urRepo, pub)

		u := &domain.User{
			Username:     "john",
			Email:        "john@example.com",
			PasswordHash: "pw",
		}
		created, _ := s.CreateUser(ctx, u, "", nil)
		originalStamp := created.SecurityStamp

		ok, err := s.UpdateCredentials(ctx, created.ID, "new-password")
		if err != nil || !ok {
			t.Fatalf("expected update credentials success, got %v", err)
		}

		fresh, _ := userRepo.GetByID(ctx, created.ID)
		if fresh.SecurityStamp == originalStamp {
			t.Error("expected security stamp to be bumped")
		}
		if len(pub.Events) != 2 { // 1 user-created, 1 password-changed
			t.Errorf("expected 2 events, got %d", len(pub.Events))
		}
	})

	t.Run("UserNotFound", func(t *testing.T) {
		userRepo := memory.NewUserRepository()
		usRepo := memory.NewUserStoreRepository()
		urRepo := memory.NewUserRoleRepository()
		pub := &sharedtesting.MockPublisher{}

		s := NewUserService(userRepo, usRepo, urRepo, pub)

		_, err := s.UpdateCredentials(ctx, "nonexistent", "pw")
		if err == nil {
			t.Error("expected error for nonexistent user")
		}
	})

	t.Run("BcryptPasswordTooLongError", func(t *testing.T) {
		userRepo := memory.NewUserRepository()
		usRepo := memory.NewUserStoreRepository()
		urRepo := memory.NewUserRoleRepository()
		pub := &sharedtesting.MockPublisher{}

		s := NewUserService(userRepo, usRepo, urRepo, pub)

		u := &domain.User{
			Username:     "john",
			Email:        "john@example.com",
			PasswordHash: "pw",
		}
		created, _ := s.CreateUser(ctx, u, "", nil)

		longPassword := strings.Repeat("a", 100)
		_, err := s.UpdateCredentials(ctx, created.ID, longPassword)
		if err == nil || !strings.Contains(err.Error(), "password length exceeds 72 bytes") {
			t.Errorf("expected password length exceeds 72 bytes error, got %v", err)
		}
	})

	t.Run("UpdateRepoError", func(t *testing.T) {
		userRepoMock := &fullMockUserRepo{
			UserRepository: memory.NewUserRepository(),
		}
		usRepo := memory.NewUserStoreRepository()
		urRepo := memory.NewUserRoleRepository()
		pub := &sharedtesting.MockPublisher{}

		s := NewUserService(userRepoMock, usRepo, urRepo, pub)

		u := &domain.User{
			Username:     "john",
			Email:        "john@example.com",
			PasswordHash: "pw",
		}
		created, _ := s.CreateUser(ctx, u, "", nil)

		userRepoMock.updateErr = errors.New("db update error")
		_, err := s.UpdateCredentials(ctx, created.ID, "new-password")
		if err == nil || err.Error() != "db update error" {
			t.Errorf("expected db update error, got %v", err)
		}
	})
}

func TestUserService_DeactivateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("UserNotFound", func(t *testing.T) {
		userRepo := memory.NewUserRepository()
		usRepo := memory.NewUserStoreRepository()
		urRepo := memory.NewUserRoleRepository()
		pub := &sharedtesting.MockPublisher{}

		s := NewUserService(userRepo, usRepo, urRepo, pub)

		err := s.DeactivateUser(ctx, "nonexistent")
		if err == nil {
			t.Error("expected error for nonexistent user")
		}
	})

	t.Run("UpdateRepoError", func(t *testing.T) {
		userRepoMock := &fullMockUserRepo{
			UserRepository: memory.NewUserRepository(),
		}
		usRepo := memory.NewUserStoreRepository()
		urRepo := memory.NewUserRoleRepository()
		pub := &sharedtesting.MockPublisher{}

		s := NewUserService(userRepoMock, usRepo, urRepo, pub)

		u := &domain.User{
			Username:     "john",
			Email:        "john@example.com",
			PasswordHash: "pw",
		}
		created, _ := s.CreateUser(ctx, u, "", nil)

		userRepoMock.updateErr = errors.New("db update error")
		err := s.DeactivateUser(ctx, created.ID)
		if err == nil || err.Error() != "db update error" {
			t.Errorf("expected db update error, got %v", err)
		}
	})
}

func TestUserService_AssignRemoveStore(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		userRepo := memory.NewUserRepository()
		usRepo := memory.NewUserStoreRepository()
		urRepo := memory.NewUserRoleRepository()
		pub := &sharedtesting.MockPublisher{}

		s := NewUserService(userRepo, usRepo, urRepo, pub)

		err := s.AssignUserToStore(ctx, "u_1", "store_1")
		if err != nil {
			t.Fatalf("assign to store failed: %v", err)
		}

		err = s.RemoveUserFromStore(ctx, "u_1", "store_1")
		if err != nil {
			t.Fatalf("remove from store failed: %v", err)
		}
	})

	t.Run("AssignError", func(t *testing.T) {
		userRepo := memory.NewUserRepository()
		usRepo := &mockUserStoreRepo{
			UserStoreRepository: memory.NewUserStoreRepository(),
			createErr:           errors.New("db error"),
		}
		urRepo := memory.NewUserRoleRepository()
		pub := &sharedtesting.MockPublisher{}

		s := NewUserService(userRepo, usRepo, urRepo, pub)

		err := s.AssignUserToStore(ctx, "u_1", "store_1")
		if err == nil || err.Error() != "db error" {
			t.Errorf("expected db error, got %v", err)
		}
	})
}
