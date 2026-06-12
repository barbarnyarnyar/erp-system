package handlers

import (
	"erp-system/shared/utils"
	"net/http"

	"github.com/erp-system/auth-service/internal/business/domain"
	"github.com/erp-system/auth-service/internal/business/service"
	"github.com/gin-gonic/gin"
)

type IdentityHandler struct {
	authSvc *service.AuthService
	userSvc *service.UserService
	rbacSvc *service.RBACService
	response *utils.ResponseHelper
}

func NewIdentityHandler(authSvc *service.AuthService,
	userSvc *service.UserService,
	rbacSvc *service.RBACService, response *utils.ResponseHelper) *IdentityHandler {
	return &IdentityHandler{
		authSvc: authSvc,
		userSvc: userSvc,
		rbacSvc: rbacSvc,
		response: response,
	}
}

type RegisterReq struct {
	Username       string   `json:"username" binding:"required"`
	Email          string   `json:"email" binding:"required"`
	Password       string   `json:"password" binding:"required"`
	FirstName      string   `json:"first_name" binding:"required"`
	LastName       string   `json:"last_name" binding:"required"`
	InitialStoreID string   `json:"initial_store_id"`
	RoleIDs        []string `json:"role_ids"`
}

func (h *IdentityHandler) Register(c *gin.Context) {
	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	user := &domain.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: req.Password, // password hash is raw password for simplicity of the in-memory example
		FirstName:    req.FirstName,
		LastName:     req.LastName,
	}

	created, err := h.userSvc.CreateUser(c.Request.Context(), user, req.InitialStoreID, req.RoleIDs)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}

	c.JSON(http.StatusCreated, created)
}

type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *IdentityHandler) Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	accessToken, refreshToken, err := h.authSvc.AuthenticateUser(c.Request.Context(), req.Username, req.Password, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
	})
}

type RefreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *IdentityHandler) Refresh(c *gin.Context) {
	var req RefreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	accessToken, refreshToken, err := h.authSvc.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
	})
}

type LogoutReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *IdentityHandler) Logout(c *gin.Context) {
	var req LogoutReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	session, err := h.authSvc.GetSessionByRefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	err = h.authSvc.RevokeToken(c.Request.Context(), session.ID)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

type AssignStoreReq struct {
	StoreID string `json:"store_id" binding:"required"`
}

func (h *IdentityHandler) AssignStore(c *gin.Context) {
	id := c.Param("id")
	var req AssignStoreReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	err := h.userSvc.AssignUserToStore(c.Request.Context(), id, req.StoreID)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User assigned to store successfully"})
}

type ValidatePermissionReq struct {
	Permission string `json:"permission" binding:"required"`
}

func (h *IdentityHandler) ValidatePermission(c *gin.Context) {
	id := c.Param("id")
	var req ValidatePermissionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	valid, err := h.rbacSvc.ValidatePermissions(c.Request.Context(), id, req.Permission)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"valid": valid})
}

type UpdateUserReq struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Email     *string `json:"email"`
	IsActive  *bool   `json:"is_active"`
}

func (h *IdentityHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var req UpdateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	updated, err := h.userSvc.UpdateUser(c.Request.Context(), id, req.FirstName, req.LastName, req.Email, req.IsActive)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *IdentityHandler) Deactivate(c *gin.Context) {
	id := c.Param("id")
	err := h.userSvc.DeactivateUser(c.Request.Context(), id)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deactivated successfully"})
}
