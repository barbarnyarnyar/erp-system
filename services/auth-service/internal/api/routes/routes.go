package routes

import (
	"github.com/erp-system/auth-service/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(
	r *gin.Engine,
	handler *handlers.IdentityHandler,
	rbacHandler *handlers.RBACHandler,
) {
	v1 := r.Group("/api/v1/auth")
	{
		v1.POST("/register", handler.Register)
		v1.POST("/login", handler.Login)
		v1.POST("/refresh", handler.Refresh)
		v1.POST("/logout", handler.Logout)

		v1.PUT("/users/:id", handler.UpdateUser)
		v1.POST("/users/:id/store", handler.AssignStore)
		v1.POST("/users/:id/validate-permission", handler.ValidatePermission)
		v1.POST("/users/:id/deactivate", handler.Deactivate)

		// Roles CRUD
		v1.GET("/roles", rbacHandler.GetRoles)
		v1.POST("/roles", rbacHandler.CreateRole)
		v1.DELETE("/roles/:id", rbacHandler.DeleteRole)

		// Permissions CRUD
		v1.GET("/permissions", rbacHandler.GetPermissions)
		v1.POST("/permissions", rbacHandler.CreatePermission)
		v1.DELETE("/permissions/:id", rbacHandler.DeletePermission)

		// Role Permissions association
		v1.GET("/roles/:id/permissions", rbacHandler.GetRolePermissions)
		v1.POST("/roles/:id/permissions", rbacHandler.AssignPermissionToRole)
		v1.DELETE("/roles/:id/permissions/:permissionId", rbacHandler.RemovePermissionFromRole)
	}
}
