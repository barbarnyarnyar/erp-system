package routes

import (
	"github.com/erp-system/plm-service/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, h *handlers.PlmHandler) {
	v1 := r.Group("/api/v1/plm")
	{
		// Materials
		v1.POST("/materials", h.CreateMaterial)
		v1.PUT("/materials/:id/specs", h.UpdateTechnicalSpecs)
		v1.PUT("/materials/:id/status", h.TransitionStatus)

		// BOM
		v1.POST("/boms", h.EstablishBomHeader)
		v1.POST("/boms/:id/release", h.ReleaseBom)
		v1.GET("/boms/:id/explode", h.ExplodeBillOfMaterials)

		// ECO
		v1.POST("/ecos", h.InitiateChangeRequest)
		v1.POST("/ecos/:id/action", h.ProcessApprovalAction)
	}
}
