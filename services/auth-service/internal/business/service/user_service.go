package service

import (
	"context"
	"erp-system/shared/utils"
	"fmt"
	"time"

	"github.com/erp-system/auth-service/internal/business/domain"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo  domain.UserRepository
	usRepo    domain.UserStoreRepository
	urRepo    domain.UserRoleRepository
	publisher domain.EventPublisher
}

func NewUserService(
	userRepo domain.UserRepository,
	usRepo domain.UserStoreRepository,
	urRepo domain.UserRoleRepository,
	publisher domain.EventPublisher,
) *UserService {
	return &UserService{
		userRepo:  userRepo,
		usRepo:    usRepo,
		urRepo:    urRepo,
		publisher: publisher,
	}
}

func (s *UserService) CreateUser(ctx context.Context, u *domain.User, initialStoreID string, roleIDs []string) (*domain.User, error) {
	u.ID = utils.NewID("user")
	u.IsActive = true
	// Initial security stamp. Every subsequent state change bumps this
	// value so that any JWT issued before the change is rejected on
	// ValidateToken (claims.SecurityStamp != user.SecurityStamp).
	u.SecurityStamp = utils.NewID("ss")
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()

	hash, err := bcrypt.GenerateFromPassword([]byte(u.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	u.PasswordHash = string(hash)

	err = s.userRepo.Create(ctx, u)
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
		linkID := utils.NewID("ur")
		ur := &domain.UserRole{
			ID:        linkID,
			UserID:    u.ID,
			RoleID:    roleID,
			CreatedAt: time.Now(),
		}
		_ = s.urRepo.Create(ctx, ur)

		// Publish user role assigned event
		if err := s.publisher.Publish(ctx, domain.TopicAuthUserRoleAssigned, ur.ID, domain.UserRoleEventPayload{
			ID:         ur.ID,
			UserID:     ur.UserID,
			RoleID:     ur.RoleID,
			AssignedBy: "",
			Timestamp:  time.Now(),
		}); err != nil {
			utils.LogPublishErr("auth-service", domain.TopicAuthUserRoleAssigned, err)
		}
	}

	// Publish user created event
	if err := s.publisher.Publish(ctx, domain.TopicAuthUserCreated, u.ID, domain.UserEventPayload{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		IsActive:  u.IsActive,
		Timestamp: time.Now(),
	}); err != nil {
		utils.LogPublishErr("auth-service", domain.TopicAuthUserCreated, err)
	}

	return u, nil
}

func (s *UserService) UpdateUser(ctx context.Context, userID string, firstName, lastName, email *string, isActive *bool) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if firstName != nil {
		user.FirstName = *firstName
	}
	if lastName != nil {
		user.LastName = *lastName
	}
	if email != nil {
		user.Email = *email
	}
	if isActive != nil && user.IsActive != *isActive {
		// Bumping the security_stamp invalidates any in-flight JWTs the
		// moment the activation flag flips. Critical for offboarding flow.
		user.IsActive = *isActive
		user.SecurityStamp = utils.NewID("ss")
	}
	user.UpdatedAt = time.Now()

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) UpdateCredentials(ctx context.Context, userID string, newPassword string) (bool, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return false, fmt.Errorf("failed to hash password: %w", err)
	}
	user.PasswordHash = string(hash)
	// Bump security_stamp: any JWT issued before the password change is now invalid.
	user.SecurityStamp = utils.NewID("ss")
	user.UpdatedAt = time.Now()

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return false, err
	}

	// Publish password changed event
	if err := s.publisher.Publish(ctx, domain.TopicAuthPasswordChanged, user.ID, domain.PasswordChangedEventPayload{
		UserID:    user.ID,
		Timestamp: time.Now(),
	}); err != nil {
		utils.LogPublishErr("auth-service", domain.TopicAuthPasswordChanged, err)
	}

	return true, nil
}

func (s *UserService) DeactivateUser(ctx context.Context, userID string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	user.IsActive = false
	// Critical: bump the security stamp so that any in-flight JWT becomes
	// invalid the moment the user is deactivated. Without this, a terminated
	// employee could keep using their old token until natural expiration.
	user.SecurityStamp = utils.NewID("ss")
	user.UpdatedAt = time.Now()
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return err
	}

	// Publish user deactivated event
	if err := s.publisher.Publish(ctx, domain.TopicAuthUserDeactivated, user.ID, domain.UserEventPayload{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		IsActive:  user.IsActive,
		Timestamp: time.Now(),
	}); err != nil {
		utils.LogPublishErr("auth-service", domain.TopicAuthUserDeactivated, err)
	}

	return nil
}

func (s *UserService) AssignUserToStore(ctx context.Context, userID, storeID string) error {
	linkID := utils.NewID("us")
	us := &domain.UserStore{
		ID:         linkID,
		UserID:     userID,
		StoreID:    storeID,
		AssignedAt: time.Now(),
	}
	err := s.usRepo.Create(ctx, us)
	if err != nil {
		return err
	}

	// Publish user store assigned event
	if err := s.publisher.Publish(ctx, domain.TopicAuthUserStoreAssigned, us.ID, domain.UserStoreEventPayload{
		ID:        us.ID,
		UserID:    us.UserID,
		StoreID:   us.StoreID,
		Timestamp: time.Now(),
	}); err != nil {
		utils.LogPublishErr("auth-service", domain.TopicAuthUserStoreAssigned, err)
	}

	return nil
}

func (s *UserService) RemoveUserFromStore(ctx context.Context, userID, storeID string) error {
	return s.usRepo.Delete(ctx, userID, storeID)
}
