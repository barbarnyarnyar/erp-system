package domain

import "context"

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	List(ctx context.Context) ([]User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
}

type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	GetByID(ctx context.Context, id string) (*Session, error)
	GetByRefreshToken(ctx context.Context, token string) (*Session, error)
	DeleteByUserID(ctx context.Context, userID string) error
	Delete(ctx context.Context, id string) error
}

type RoleRepository interface {
	Create(ctx context.Context, role *Role) error
	GetByID(ctx context.Context, id string) (*Role, error)
	GetByName(ctx context.Context, name string) (*Role, error)
	List(ctx context.Context) ([]Role, error)
}

type PermissionRepository interface {
	Create(ctx context.Context, permission *Permission) error
	GetByID(ctx context.Context, id string) (*Permission, error)
	GetByCode(ctx context.Context, code string) (*Permission, error)
	List(ctx context.Context) ([]Permission, error)
}

type UserRoleRepository interface {
	Create(ctx context.Context, ur *UserRole) error
	ListByUserID(ctx context.Context, userID string) ([]UserRole, error)
	Delete(ctx context.Context, userID string, roleID string) error
}

type UserStoreRepository interface {
	Create(ctx context.Context, us *UserStore) error
	ListByUserID(ctx context.Context, userID string) ([]UserStore, error)
	Delete(ctx context.Context, userID string, storeID string) error
}

type RolePermissionRepository interface {
	Create(ctx context.Context, rp *RolePermission) error
	ListByRoleID(ctx context.Context, roleID string) ([]RolePermission, error)
}
