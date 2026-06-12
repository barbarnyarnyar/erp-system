package routes

import (
	"github.com/erp-system/eam-service/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, h *handlers.EamHandler) {
	v1 := r.Group("/api/v1/eam")
	{
		// Facilities
		v1.POST("/facilities", h.CreateFacility)

		// Equipment
		v1.POST("/equipment", h.RegisterEquipment)
		v1.GET("/equipment", h.FetchTargetTenantAssets)
		v1.PUT("/equipment/:id/status", h.UpdateEquipmentStatus)
		v1.PUT("/equipment/:id/finance-asset", h.AssociateFinancialAsset)

		// Work Orders
		v1.POST("/work-orders", h.FileMachineIncident)
		v1.PUT("/work-orders/:id/route", h.RouteToTechnician)
		v1.POST("/work-orders/:id/start", h.TransitionToActiveState)
		v1.POST("/work-orders/:id/resolve", h.FinalizeResolution)

		// Telemetry
		v1.POST("/telemetry/sensor-metrics", h.QueueSensorMetrics)
		v1.POST("/telemetry/flush", h.FlushStagedMetrics)
	}
}
