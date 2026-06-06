package handlers

import (
	"net/http"

	"github.com/erp-system/auth-service/internal/business/domain"
	"github.com/erp-system/auth-service/internal/business/service"
	"github.com/gin-gonic/gin"
)

type IdentityHandler struct {
	idSvc *service.IdentityService
}

func NewIdentityHandler(idSvc *service.IdentityService) *IdentityHandler {
	return &IdentityHandler{idSvc: idSvc}
}

type RegisterReq struct {
	Username       string   `json:"username" binding:"required"`
	Email          string   `json:"email" binding:"required"`
	PasswordHash   string   `json:"password_hash" binding:"required"`
	FirstName      string   `json:"first_name" binding:"required"`
	LastName       string   `json:"last_name" binding:"required"`
	InitialStoreID string   `json:"initial_store_id"`
	RoleIDs        []string `json:"role_ids"`
}

func (h *IdentityHandler) Register(c *gin.Context) {
	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := &domain.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: req.PasswordHash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
	}

	created, err := h.idSvc.CreateUser(c.Request.Context(), user, req.InitialStoreID, req.RoleIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

type LoginReq struct {
	Username     string `json:"username" binding:"required"`
	PasswordHash string `json:"password_hash" binding:"required"`
}

func (h *IdentityHandler) Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, refreshToken, err := h.idSvc.AuthenticateUser(c.Request.Context(), req.Username, req.PasswordHash)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, refreshToken, err := h.idSvc.RefreshToken(c.Request.Context(), req.RefreshToken)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.idSvc.Logout(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.idSvc.AssignUserToStore(c.Request.Context(), id, req.StoreID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	valid, err := h.idSvc.ValidatePermissions(c.Request.Context(), id, req.Permission)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"valid": valid})
}

func (h *IdentityHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.idSvc.UpdateUser(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *IdentityHandler) Deactivate(c *gin.Context) {
	id := c.Param("id")
	err := h.idSvc.DeactivateUser(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deactivated successfully"})
}
