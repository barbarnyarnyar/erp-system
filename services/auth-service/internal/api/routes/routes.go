package routes

import (
	"github.com/erp-system/auth-service/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(
	r *gin.Engine,
	handler *handlers.IdentityHandler,
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
	}
}
