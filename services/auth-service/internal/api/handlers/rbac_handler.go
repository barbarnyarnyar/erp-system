package handlers

import (
	"erp-system/shared/utils"
	"net/http"

	"github.com/erp-system/auth-service/internal/business/service"
	"github.com/gin-gonic/gin"
)

type RBACHandler struct {
	svc *service.RBACService
	response *utils.ResponseHelper
}

func NewRBACHandler(svc *service.RBACService, response *utils.ResponseHelper) *RBACHandler {
	return &RBACHandler{
		svc: svc,
		response: response,
	}
}

func (h *RBACHandler) GetRoles(c *gin.Context) {
	roles, err := h.svc.ListRoles(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": roles})
}

func (h *RBACHandler) CreateRole(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	role, err := h.svc.CreateRole(c.Request.Context(), req.Name, req.Description)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}

	c.JSON(http.StatusCreated, role)
}

func (h *RBACHandler) DeleteRole(c *gin.Context) {
	id := c.Param("id")
	err := h.svc.DeleteRole(c.Request.Context(), id)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *RBACHandler) GetPermissions(c *gin.Context) {
	perms, err := h.svc.ListPermissions(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": perms})
}

func (h *RBACHandler) CreatePermission(c *gin.Context) {
	var req struct {
		Code        string `json:"code" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	perm, err := h.svc.CreatePermission(c.Request.Context(), req.Code, req.Description)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}

	c.JSON(http.StatusCreated, perm)
}

func (h *RBACHandler) DeletePermission(c *gin.Context) {
	id := c.Param("id")
	err := h.svc.DeletePermission(c.Request.Context(), id)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *RBACHandler) GetRolePermissions(c *gin.Context) {
	roleID := c.Param("id")
	perms, err := h.svc.GetRolePermissions(c.Request.Context(), roleID)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": perms})
}

func (h *RBACHandler) AssignPermissionToRole(c *gin.Context) {
	roleID := c.Param("id")
	var req struct {
		PermissionID string `json:"permission_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	err := h.svc.AssignPermissionToRole(c.Request.Context(), roleID, req.PermissionID)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permission assigned to role successfully"})
}

func (h *RBACHandler) RemovePermissionFromRole(c *gin.Context) {
	roleID := c.Param("id")
	permissionID := c.Param("permissionId")

	err := h.svc.RemovePermissionFromRole(c.Request.Context(), roleID, permissionID)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
