package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"erp-system/shared/utils"
	"github.com/erp-system/auth-service/internal/api/handlers"
	"github.com/erp-system/auth-service/internal/api/routes"
	"github.com/erp-system/auth-service/internal/business/domain"
	"github.com/erp-system/auth-service/internal/business/service"
	"github.com/erp-system/auth-service/internal/config"
	"github.com/erp-system/auth-service/internal/data/memory"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
	utils.InitLogger("auth-service-test")
}

type testEnv struct {
	router    *gin.Engine
	userRepo  *memory.UserRepository
	sessRepo  *memory.SessionRepository
	roleRepo  *memory.RoleRepository
	permRepo  *memory.PermissionRepository
	urRepo    *memory.UserRoleRepository
	usRepo    *memory.UserStoreRepository
	rpRepo    *memory.RolePermissionRepository
	publisher *mockPublisher
}

type mockPublisher struct {
	Published []publishedEvent
}

type publishedEvent struct {
	Topic   string
	Key     string
	Payload interface{}
}

func (m *mockPublisher) Publish(ctx context.Context, topic string, key string, payload interface{}) error {
	m.Published = append(m.Published, publishedEvent{Topic: topic, Key: key, Payload: payload})
	return nil
}

func setupTestEnv() *testEnv {
	userRepo := memory.NewUserRepository()
	sessRepo := memory.NewSessionRepository()
	roleRepo := memory.NewRoleRepository()
	permRepo := memory.NewPermissionRepository()
	urRepo := memory.NewUserRoleRepository()
	usRepo := memory.NewUserStoreRepository()
	rpRepo := memory.NewRolePermissionRepository()
	publisher := &mockPublisher{}

	cfg := &config.Config{}
	cfg.JWT.AccessExpiry = 15
	cfg.JWT.RefreshExpiry = 10080

	wrappedUserRepo := &errorInjectingUserRepo{delegate: userRepo}
	wrappedSessRepo := &errorInjectingSessionRepo{delegate: sessRepo}
	wrappedRoleRepo := &errorInjectingRoleRepo{delegate: roleRepo}
	wrappedPermRepo := &errorInjectingPermRepo{delegate: permRepo}
	wrappedUrRepo := &errorInjectingUserRoleRepo{delegate: urRepo}
	wrappedUsRepo := &errorInjectingUserStoreRepo{delegate: usRepo}
	wrappedRpRepo := &errorInjectingRolePermissionRepo{delegate: rpRepo}

	rbacSvc := service.NewRBACService(wrappedRoleRepo, wrappedPermRepo, wrappedUrRepo, wrappedRpRepo, publisher)
	userSvc := service.NewUserService(wrappedUserRepo, wrappedUsRepo, wrappedUrRepo, publisher)
	authSvc := service.NewAuthService(wrappedUserRepo, wrappedSessRepo, rbacSvc, publisher, cfg)

	response := utils.NewResponseHelper("auth-service")

	identityHandler := handlers.NewIdentityHandler(authSvc, userSvc, rbacSvc, response)
	rbacHandler := handlers.NewRBACHandler(rbacSvc, response)

	router := gin.New()
	routes.SetupAuthRoutes(router, identityHandler, rbacHandler)

	return &testEnv{
		router:    router,
		userRepo:  userRepo,
		sessRepo:  sessRepo,
		roleRepo:  roleRepo,
		permRepo:  permRepo,
		urRepo:    urRepo,
		usRepo:    usRepo,
		rpRepo:    rpRepo,
		publisher: publisher,
	}
}

func TestAuthEndpoints(t *testing.T) {
	env := setupTestEnv()

	// 1. Register User
	body, _ := json.Marshal(map[string]interface{}{
		"username":   "alice",
		"email":      "alice@example.com",
		"password":   "password123",
		"first_name": "Alice",
		"last_name":  "Smith",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 on register, got %d. Body: %s", w.Code, w.Body.String())
	}

	var regResp domain.User
	_ = json.Unmarshal(w.Body.Bytes(), &regResp)
	userID := regResp.ID

	// 2. Login User
	body, _ = json.Marshal(map[string]interface{}{
		"username": "alice",
		"password": "password123",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on login, got %d. Body: %s", w.Code, w.Body.String())
	}

	var loginResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &loginResp)

	// 3. Refresh Token
	body, _ = json.Marshal(map[string]interface{}{
		"refresh_token": loginResp.RefreshToken,
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on refresh, got %d. Body: %s", w.Code, w.Body.String())
	}

	var refreshResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &refreshResp)
	currentRefreshToken := refreshResp.RefreshToken

	// 4. Update User
	body, _ = json.Marshal(map[string]interface{}{
		"first_name": "AliceUpdated",
		"last_name":  "SmithUpdated",
		"email":      "alice.updated@example.com",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/auth/users/"+userID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 on update user, got %d. Body: %s", w.Code, w.Body.String())
	}

	// 5. Assign Store
	body, _ = json.Marshal(map[string]interface{}{
		"store_id": "store_123",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/auth/users/"+userID+"/store", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 on assign store, got %d. Body: %s", w.Code, w.Body.String())
	}

	// 6. Validate Permission (requires token in header or payload)
	body, _ = json.Marshal(map[string]interface{}{
		"permission": "READ_ORDERS",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/auth/users/"+userID+"/validate-permission", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+loginResp.AccessToken)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 on validate permission, got %d. Body: %s", w.Code, w.Body.String())
	}

	// 7. Deactivate User
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/auth/users/"+userID+"/deactivate", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 on deactivate, got %d. Body: %s", w.Code, w.Body.String())
	}

	// 8. Logout User
	body, _ = json.Marshal(map[string]interface{}{
		"refresh_token": currentRefreshToken,
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/auth/logout", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 on logout, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestRBACEndpoints(t *testing.T) {
	env := setupTestEnv()

	// 1. Create Role
	body, _ := json.Marshal(map[string]interface{}{
		"name":        "ADMIN",
		"description": "Administrator role",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/roles", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 on create role, got %d", w.Code)
	}

	var role domain.Role
	_ = json.Unmarshal(w.Body.Bytes(), &role)

	// 2. Create Permission
	body, _ = json.Marshal(map[string]interface{}{
		"code":        "READ_ORDERS",
		"description": "Read sales orders permission",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/auth/permissions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 on create permission, got %d", w.Code)
	}

	var perm domain.Permission
	_ = json.Unmarshal(w.Body.Bytes(), &perm)

	// 3. Assign Permission to Role
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/auth/roles/"+role.ID+"/permissions", bytes.NewBuffer([]byte(`{"permission_id":"`+perm.ID+`"}`)))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 on assign permission, got %d. Body: %s", w.Code, w.Body.String())
	}

	// 4. Get Role Permissions
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/auth/roles/"+role.ID+"/permissions", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 on get role permissions, got %d", w.Code)
	}

	// 5. Get Roles list
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/auth/roles", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 on get roles, got %d", w.Code)
	}

	// 6. Get Permissions list
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/auth/permissions", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 on get permissions, got %d", w.Code)
	}

	// 7. Remove Permission from Role
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/auth/roles/"+role.ID+"/permissions/"+perm.ID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204 on remove permission, got %d", w.Code)
	}

	// 8. Delete Role
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/auth/roles/"+role.ID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204 on delete role, got %d", w.Code)
	}

	// 9. Delete Permission
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/auth/permissions/"+perm.ID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204 on delete permission, got %d", w.Code)
	}
}

func TestAuthErrorPaths(t *testing.T) {
	env := setupTestEnv()

	// 1. Validation bad requests (bad json / missing binding fields)
	badRequests := []struct {
		url    string
		method string
		body   string
	}{
		{"/api/v1/auth/register", http.MethodPost, `{"username":""}`},
		{"/api/v1/auth/login", http.MethodPost, `{"username":""}`},
		{"/api/v1/auth/refresh", http.MethodPost, `{"refresh_token":""}`},
		{"/api/v1/auth/logout", http.MethodPost, `{"refresh_token":""}`},
		{"/api/v1/auth/users/user-123/store", http.MethodPost, `{"store_id":""}`},
		{"/api/v1/auth/users/user-123/validate-permission", http.MethodPost, `{"permission":""}`},
		{"/api/v1/auth/users/user-123", http.MethodPut, `invalid-json`},
		{"/api/v1/auth/roles", http.MethodPost, `{"name":""}`},
		{"/api/v1/auth/permissions", http.MethodPost, `{"code":""}`},
		{"/api/v1/auth/roles/role-123/permissions", http.MethodPost, `{"permission_id":""}`},
	}

	for _, item := range badRequests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(item.method, item.url, bytes.NewBufferString(item.body))
		req.Header.Set("Content-Type", "application/json")
		env.router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400 for %s, got %d. Body: %s", item.url, w.Code, w.Body.String())
		}
	}

	// 2. Canceled context errors (to test InternalErr propagation)
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	// Create a user first to have one
	body, _ := json.Marshal(map[string]interface{}{
		"username":   "bob",
		"email":      "bob@example.com",
		"password":   "pass123",
		"first_name": "Bob",
		"last_name":  "Jones",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	var createdUser domain.User
	_ = json.Unmarshal(w.Body.Bytes(), &createdUser)
	userID := createdUser.ID

	// Login to get refresh token
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	var loginResp struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &loginResp)

	// Now try various endpoints with canceled context to force repo to return context.Canceled error

	// Register with canceled context
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(canceledCtx)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 on register with canceled ctx, got %d", w.Code)
	}

	// Login with invalid password/unauthorized
	w = httptest.NewRecorder()
	loginBad, _ := json.Marshal(map[string]interface{}{
		"username": "bob",
		"password": "wrong-password",
	})
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(loginBad))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 on bad password, got %d", w.Code)
	}

	// Refresh with invalid token
	w = httptest.NewRecorder()
	refreshBad, _ := json.Marshal(map[string]interface{}{
		"refresh_token": "invalid-token",
	})
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBuffer(refreshBad))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 on bad refresh token, got %d", w.Code)
	}

	// Logout with invalid token
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/auth/logout", bytes.NewBuffer(refreshBad))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 on logout with bad refresh token, got %d", w.Code)
	}

	// Logout with canceled context during RevokeToken
	w = httptest.NewRecorder()
	logoutBody, _ := json.Marshal(map[string]interface{}{
		"refresh_token": loginResp.RefreshToken,
	})
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/auth/logout", bytes.NewBuffer(logoutBody))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(canceledCtx)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 on logout with canceled ctx, got %d", w.Code)
	}

	// Assign store with canceled context
	w = httptest.NewRecorder()
	storeBody, _ := json.Marshal(map[string]interface{}{
		"store_id": "store-123",
	})
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/auth/users/"+userID+"/store", bytes.NewBuffer(storeBody))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(canceledCtx)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 on assign store with canceled ctx, got %d", w.Code)
	}

	// Validate permission with canceled context
	w = httptest.NewRecorder()
	permBody, _ := json.Marshal(map[string]interface{}{
		"permission": "READ_ORDERS",
	})
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/auth/users/"+userID+"/validate-permission", bytes.NewBuffer(permBody))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(canceledCtx)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 on validate permission with canceled ctx, got %d", w.Code)
	}

	// Update user with canceled context
	w = httptest.NewRecorder()
	updateBody, _ := json.Marshal(map[string]interface{}{
		"first_name": "Bob2",
	})
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/auth/users/"+userID, bytes.NewBuffer(updateBody))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(canceledCtx)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 on update user with canceled ctx, got %d", w.Code)
	}

	// Deactivate user with canceled context
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/auth/users/"+userID+"/deactivate", nil)
	req = req.WithContext(canceledCtx)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 on deactivate with canceled ctx, got %d", w.Code)
	}

	// RBAC error paths (canceled contexts)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/auth/roles", nil)
	req = req.WithContext(canceledCtx)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 on GetRoles with canceled ctx, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	roleBody, _ := json.Marshal(map[string]interface{}{
		"name": "TEST-ROLE",
	})
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/auth/roles", bytes.NewBuffer(roleBody))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(canceledCtx)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 on CreateRole with canceled ctx, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/auth/roles/some-role", nil)
	req = req.WithContext(canceledCtx)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 on DeleteRole with canceled ctx, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/auth/permissions", nil)
	req = req.WithContext(canceledCtx)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 on GetPermissions with canceled ctx, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	permBody2, _ := json.Marshal(map[string]interface{}{
		"code": "TEST-PERM",
	})
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/auth/permissions", bytes.NewBuffer(permBody2))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(canceledCtx)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 on CreatePermission with canceled ctx, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/auth/permissions/some-perm", nil)
	req = req.WithContext(canceledCtx)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 on DeletePermission with canceled ctx, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/auth/roles/some-role/permissions", nil)
	req = req.WithContext(canceledCtx)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 on GetRolePermissions with canceled ctx, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/auth/roles/some-role/permissions", bytes.NewBuffer([]byte(`{"permission_id":"some-perm"}`)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(canceledCtx)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 on AssignPermissionToRole with canceled ctx, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/auth/roles/some-role/permissions/some-perm", nil)
	req = req.WithContext(canceledCtx)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 on RemovePermissionFromRole with canceled ctx, got %d", w.Code)
	}
}

func checkCtx(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return nil
}

type errorInjectingUserRepo struct {
	delegate domain.UserRepository
}

func (r *errorInjectingUserRepo) Create(ctx context.Context, user *domain.User) error {
	if err := checkCtx(ctx); err != nil {
		return err
	}
	return r.delegate.Create(ctx, user)
}

func (r *errorInjectingUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	if err := checkCtx(ctx); err != nil {
		return nil, err
	}
	return r.delegate.GetByID(ctx, id)
}

func (r *errorInjectingUserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	if err := checkCtx(ctx); err != nil {
		return nil, err
	}
	return r.delegate.GetByUsername(ctx, username)
}

func (r *errorInjectingUserRepo) List(ctx context.Context) ([]domain.User, error) {
	if err := checkCtx(ctx); err != nil {
		return nil, err
	}
	return r.delegate.List(ctx)
}

func (r *errorInjectingUserRepo) Update(ctx context.Context, user *domain.User) error {
	if err := checkCtx(ctx); err != nil {
		return err
	}
	return r.delegate.Update(ctx, user)
}

func (r *errorInjectingUserRepo) Delete(ctx context.Context, id string) error {
	if err := checkCtx(ctx); err != nil {
		return err
	}
	return r.delegate.Delete(ctx, id)
}

type errorInjectingSessionRepo struct {
	delegate domain.SessionRepository
}

func (r *errorInjectingSessionRepo) Create(ctx context.Context, session *domain.Session) error {
	if err := checkCtx(ctx); err != nil {
		return err
	}
	return r.delegate.Create(ctx, session)
}

func (r *errorInjectingSessionRepo) GetByID(ctx context.Context, id string) (*domain.Session, error) {
	if err := checkCtx(ctx); err != nil {
		return nil, err
	}
	return r.delegate.GetByID(ctx, id)
}

func (r *errorInjectingSessionRepo) GetByRefreshToken(ctx context.Context, token string) (*domain.Session, error) {
	if err := checkCtx(ctx); err != nil {
		return nil, err
	}
	return r.delegate.GetByRefreshToken(ctx, token)
}

func (r *errorInjectingSessionRepo) Update(ctx context.Context, session *domain.Session) error {
	if err := checkCtx(ctx); err != nil {
		return err
	}
	return r.delegate.Update(ctx, session)
}

func (r *errorInjectingSessionRepo) DeleteByUserID(ctx context.Context, userID string) error {
	if err := checkCtx(ctx); err != nil {
		return err
	}
	return r.delegate.DeleteByUserID(ctx, userID)
}

func (r *errorInjectingSessionRepo) Delete(ctx context.Context, id string) error {
	if err := checkCtx(ctx); err != nil {
		return err
	}
	return r.delegate.Delete(ctx, id)
}

type errorInjectingRoleRepo struct {
	delegate domain.RoleRepository
}

func (r *errorInjectingRoleRepo) Create(ctx context.Context, role *domain.Role) error {
	if err := checkCtx(ctx); err != nil {
		return err
	}
	return r.delegate.Create(ctx, role)
}

func (r *errorInjectingRoleRepo) GetByID(ctx context.Context, id string) (*domain.Role, error) {
	if err := checkCtx(ctx); err != nil {
		return nil, err
	}
	return r.delegate.GetByID(ctx, id)
}

func (r *errorInjectingRoleRepo) GetByName(ctx context.Context, name string) (*domain.Role, error) {
	if err := checkCtx(ctx); err != nil {
		return nil, err
	}
	return r.delegate.GetByName(ctx, name)
}

func (r *errorInjectingRoleRepo) List(ctx context.Context) ([]domain.Role, error) {
	if err := checkCtx(ctx); err != nil {
		return nil, err
	}
	return r.delegate.List(ctx)
}

func (r *errorInjectingRoleRepo) Delete(ctx context.Context, id string) error {
	if err := checkCtx(ctx); err != nil {
		return err
	}
	return r.delegate.Delete(ctx, id)
}

type errorInjectingPermRepo struct {
	delegate domain.PermissionRepository
}

func (r *errorInjectingPermRepo) Create(ctx context.Context, permission *domain.Permission) error {
	if err := checkCtx(ctx); err != nil {
		return err
	}
	return r.delegate.Create(ctx, permission)
}

func (r *errorInjectingPermRepo) GetByID(ctx context.Context, id string) (*domain.Permission, error) {
	if err := checkCtx(ctx); err != nil {
		return nil, err
	}
	return r.delegate.GetByID(ctx, id)
}

func (r *errorInjectingPermRepo) GetByCode(ctx context.Context, code string) (*domain.Permission, error) {
	if err := checkCtx(ctx); err != nil {
		return nil, err
	}
	return r.delegate.GetByCode(ctx, code)
}

func (r *errorInjectingPermRepo) List(ctx context.Context) ([]domain.Permission, error) {
	if err := checkCtx(ctx); err != nil {
		return nil, err
	}
	return r.delegate.List(ctx)
}

func (r *errorInjectingPermRepo) Delete(ctx context.Context, id string) error {
	if err := checkCtx(ctx); err != nil {
		return err
	}
	return r.delegate.Delete(ctx, id)
}

type errorInjectingUserRoleRepo struct {
	delegate domain.UserRoleRepository
}

func (r *errorInjectingUserRoleRepo) Create(ctx context.Context, ur *domain.UserRole) error {
	if err := checkCtx(ctx); err != nil {
		return err
	}
	return r.delegate.Create(ctx, ur)
}

func (r *errorInjectingUserRoleRepo) ListByUserID(ctx context.Context, userID string) ([]domain.UserRole, error) {
	if err := checkCtx(ctx); err != nil {
		return nil, err
	}
	return r.delegate.ListByUserID(ctx, userID)
}

func (r *errorInjectingUserRoleRepo) Delete(ctx context.Context, userID string, roleID string) error {
	if err := checkCtx(ctx); err != nil {
		return err
	}
	return r.delegate.Delete(ctx, userID, roleID)
}

type errorInjectingUserStoreRepo struct {
	delegate domain.UserStoreRepository
}

func (r *errorInjectingUserStoreRepo) Create(ctx context.Context, us *domain.UserStore) error {
	if err := checkCtx(ctx); err != nil {
		return err
	}
	return r.delegate.Create(ctx, us)
}

func (r *errorInjectingUserStoreRepo) ListByUserID(ctx context.Context, userID string) ([]domain.UserStore, error) {
	if err := checkCtx(ctx); err != nil {
		return nil, err
	}
	return r.delegate.ListByUserID(ctx, userID)
}

func (r *errorInjectingUserStoreRepo) Delete(ctx context.Context, userID string, storeID string) error {
	if err := checkCtx(ctx); err != nil {
		return err
	}
	return r.delegate.Delete(ctx, userID, storeID)
}

type errorInjectingRolePermissionRepo struct {
	delegate domain.RolePermissionRepository
}

func (r *errorInjectingRolePermissionRepo) Create(ctx context.Context, rp *domain.RolePermission) error {
	if err := checkCtx(ctx); err != nil {
		return err
	}
	return r.delegate.Create(ctx, rp)
}

func (r *errorInjectingRolePermissionRepo) ListByRoleID(ctx context.Context, roleID string) ([]domain.RolePermission, error) {
	if err := checkCtx(ctx); err != nil {
		return nil, err
	}
	return r.delegate.ListByRoleID(ctx, roleID)
}

func (r *errorInjectingRolePermissionRepo) Delete(ctx context.Context, roleID string, permissionID string) error {
	if err := checkCtx(ctx); err != nil {
		return err
	}
	return r.delegate.Delete(ctx, roleID, permissionID)
}
