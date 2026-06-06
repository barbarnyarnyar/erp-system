package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/erp-system/auth-service/internal/business/domain"
)

type UserRepository struct {
	mu    sync.RWMutex
	users map[string]domain.User
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		users: make(map[string]domain.User),
	}
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[u.ID] = *u
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	if !ok {
		return nil, fmt.Errorf("user not found: %s", id)
	}
	return &u, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if u.Username == username {
			return &u, nil
		}
	}
	return nil, fmt.Errorf("user not found: %s", username)
}

func (r *UserRepository) List(ctx context.Context) ([]domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.User, 0, len(r.users))
	for _, u := range r.users {
		list = append(list, u)
	}
	return list, nil
}

func (r *UserRepository) Update(ctx context.Context, u *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.users[u.ID]; !ok {
		return fmt.Errorf("user not found: %s", u.ID)
	}
	r.users[u.ID] = *u
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.users, id)
	return nil
}

type SessionRepository struct {
	mu       sync.RWMutex
	sessions map[string]domain.Session
}

func NewSessionRepository() *SessionRepository {
	return &SessionRepository{
		sessions: make(map[string]domain.Session),
	}
}

func (r *SessionRepository) Create(ctx context.Context, s *domain.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[s.ID] = *s
	return nil
}

func (r *SessionRepository) GetByID(ctx context.Context, id string) (*domain.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.sessions[id]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	return &s, nil
}

func (r *SessionRepository) GetByRefreshToken(ctx context.Context, token string) (*domain.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, s := range r.sessions {
		if s.RefreshToken == token {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("session not found by token")
}

func (r *SessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, s := range r.sessions {
		if s.UserID == userID {
			delete(r.sessions, id)
		}
	}
	return nil
}

func (r *SessionRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, id)
	return nil
}

type RoleRepository struct {
	mu    sync.RWMutex
	roles map[string]domain.Role
}

func NewRoleRepository() *RoleRepository {
	return &RoleRepository{
		roles: make(map[string]domain.Role),
	}
}

func (r *RoleRepository) Create(ctx context.Context, role *domain.Role) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.roles[role.ID] = *role
	return nil
}

func (r *RoleRepository) GetByID(ctx context.Context, id string) (*domain.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rl, ok := r.roles[id]
	if !ok {
		return nil, fmt.Errorf("role not found: %s", id)
	}
	return &rl, nil
}

func (r *RoleRepository) GetByName(ctx context.Context, name string) (*domain.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, rl := range r.roles {
		if rl.Name == name {
			return &rl, nil
		}
	}
	return nil, fmt.Errorf("role not found: %s", name)
}

func (r *RoleRepository) List(ctx context.Context) ([]domain.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.Role, 0, len(r.roles))
	for _, rl := range r.roles {
		list = append(list, rl)
	}
	return list, nil
}

func (r *RoleRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.roles, id)
	return nil
}

type PermissionRepository struct {
	mu          sync.RWMutex
	permissions map[string]domain.Permission
}

func NewPermissionRepository() *PermissionRepository {
	return &PermissionRepository{
		permissions: make(map[string]domain.Permission),
	}
}

func (r *PermissionRepository) Create(ctx context.Context, p *domain.Permission) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.permissions[p.ID] = *p
	return nil
}

func (r *PermissionRepository) GetByID(ctx context.Context, id string) (*domain.Permission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.permissions[id]
	if !ok {
		return nil, fmt.Errorf("permission not found: %s", id)
	}
	return &p, nil
}

func (r *PermissionRepository) GetByCode(ctx context.Context, code string) (*domain.Permission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.permissions {
		if p.Code == code {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("permission not found: %s", code)
}

func (r *PermissionRepository) List(ctx context.Context) ([]domain.Permission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.Permission, 0, len(r.permissions))
	for _, p := range r.permissions {
		list = append(list, p)
	}
	return list, nil
}

func (r *PermissionRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.permissions, id)
	return nil
}

type UserRoleRepository struct {
	mu    sync.RWMutex
	links map[string]domain.UserRole
}

func NewUserRoleRepository() *UserRoleRepository {
	return &UserRoleRepository{
		links: make(map[string]domain.UserRole),
	}
}

func (r *UserRoleRepository) Create(ctx context.Context, link *domain.UserRole) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.links[link.ID] = *link
	return nil
}

func (r *UserRoleRepository) ListByUserID(ctx context.Context, userID string) ([]domain.UserRole, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.UserRole
	for _, l := range r.links {
		if l.UserID == userID {
			list = append(list, l)
		}
	}
	return list, nil
}

func (r *UserRoleRepository) Delete(ctx context.Context, userID string, roleID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, l := range r.links {
		if l.UserID == userID && l.RoleID == roleID {
			delete(r.links, id)
		}
	}
	return nil
}

type UserStoreRepository struct {
	mu    sync.RWMutex
	links map[string]domain.UserStore
}

func NewUserStoreRepository() *UserStoreRepository {
	return &UserStoreRepository{
		links: make(map[string]domain.UserStore),
	}
}

func (r *UserStoreRepository) Create(ctx context.Context, link *domain.UserStore) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.links[link.ID] = *link
	return nil
}

func (r *UserStoreRepository) ListByUserID(ctx context.Context, userID string) ([]domain.UserStore, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.UserStore
	for _, l := range r.links {
		if l.UserID == userID {
			list = append(list, l)
		}
	}
	return list, nil
}

func (r *UserStoreRepository) Delete(ctx context.Context, userID string, storeID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, l := range r.links {
		if l.UserID == userID && l.StoreID == storeID {
			delete(r.links, id)
		}
	}
	return nil
}

type RolePermissionRepository struct {
	mu    sync.RWMutex
	links map[string]domain.RolePermission
}

func NewRolePermissionRepository() *RolePermissionRepository {
	return &RolePermissionRepository{
		links: make(map[string]domain.RolePermission),
	}
}

func (r *RolePermissionRepository) Create(ctx context.Context, link *domain.RolePermission) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.links[link.ID] = *link
	return nil
}

func (r *RolePermissionRepository) ListByRoleID(ctx context.Context, roleID string) ([]domain.RolePermission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.RolePermission
	for _, l := range r.links {
		if l.RoleID == roleID {
			list = append(list, l)
		}
	}
	return list, nil
}

func (r *RolePermissionRepository) Delete(ctx context.Context, roleID string, permissionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, l := range r.links {
		if l.RoleID == roleID && l.PermissionID == permissionID {
			delete(r.links, id)
		}
	}
	return nil
}
