package service

import (
	"context"
	"fmt"
	"time"

	"github.com/erp-system/auth-service/internal/business/domain"
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
			ID:        linkID,
			UserID:    u.ID,
			RoleID:    roleID,
			CreatedAt: time.Now(),
		}
		_ = s.urRepo.Create(ctx, ur)

		// Publish user role assigned event
		_ = s.publisher.Publish(ctx, domain.TopicAuthUserRoleAssigned, ur.ID, domain.UserRoleEventPayload{
			ID:         ur.ID,
			UserID:     ur.UserID,
			RoleID:     ur.RoleID,
			AssignedBy: "",
			Timestamp:  time.Now(),
		})
	}

	// Publish user created event
	_ = s.publisher.Publish(ctx, domain.TopicAuthUserCreated, u.ID, domain.UserEventPayload{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		IsActive:  u.IsActive,
		Timestamp: time.Now(),
	})

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
	if isActive != nil {
		user.IsActive = *isActive
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

	user.PasswordHash = newPassword
	user.UpdatedAt = time.Now()

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return false, err
	}

	// Publish password changed event
	_ = s.publisher.Publish(ctx, domain.TopicAuthPasswordChanged, user.ID, domain.PasswordChangedEventPayload{
		UserID:    user.ID,
		Timestamp: time.Now(),
	})

	return true, nil
}

func (s *UserService) DeactivateUser(ctx context.Context, userID string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	user.IsActive = false
	user.UpdatedAt = time.Now()
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return err
	}

	// Publish user deactivated event
	_ = s.publisher.Publish(ctx, domain.TopicAuthUserDeactivated, user.ID, domain.UserEventPayload{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		IsActive:  user.IsActive,
		Timestamp: time.Now(),
	})

	return nil
}

func (s *UserService) AssignUserToStore(ctx context.Context, userID, storeID string) error {
	linkID := fmt.Sprintf("us_%d", time.Now().UnixNano())
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
	_ = s.publisher.Publish(ctx, domain.TopicAuthUserStoreAssigned, us.ID, domain.UserStoreEventPayload{
		ID:        us.ID,
		UserID:    us.UserID,
		StoreID:   us.StoreID,
		Timestamp: time.Now(),
	})

	return nil
}

func (s *UserService) RemoveUserFromStore(ctx context.Context, userID, storeID string) error {
	return s.usRepo.Delete(ctx, userID, storeID)
}
