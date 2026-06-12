package routes

import (
	"github.com/erp-system/qms-service/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, h *handlers.QmsHandler) {
	v1 := r.Group("/api/v1/qms")
	{
		// Inspection Plans
		v1.POST("/plans", h.ConfigurePlan)
		v1.POST("/plans/metrics", h.RegisterPlanMetric)

		// Inspection Execution
		v1.POST("/inspections", h.StageInspection)
		v1.PUT("/inspections/:id/assign", h.AssignInspector)
		v1.POST("/inspections/:id/results", h.RecordBulkMeasurements)

		// Non-Conformances
		v1.POST("/non-conformances", h.LogFailureIncident)
		v1.POST("/non-conformances/:id/disposition", h.ExecuteDisposition)

		// SPC Analytics
		v1.GET("/analytics/spc", h.ComputeSpcDistribution)
	}
}
